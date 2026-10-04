package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/thomhuang/local-leetcode/internal/assistant"
	"github.com/thomhuang/local-leetcode/internal/question"
)

// deepseekAPIKeyEnv is the environment variable that holds the API key. The key
// is never written to disk.
const deepseekAPIKeyEnv = "DEEPSEEK_API_KEY"

// hasAssistantKey reports whether the DeepSeek API key is configured.
func hasAssistantKey() bool {
	return strings.TrimSpace(os.Getenv(deepseekAPIKeyEnv)) != ""
}

// startChatStream sends a streaming chat request and returns the response body.
// The caller owns the body and must close it; ctx cancels an in-flight request.
func (app *App) startChatStream(ctx context.Context, cfg assistant.Config, messages []assistant.Message) (io.ReadCloser, error) {
	apiKey := strings.TrimSpace(os.Getenv(deepseekAPIKeyEnv))
	if apiKey == "" {
		app.log().Warn("chat.http", "assistant requested without an API key", "env", deepseekAPIKeyEnv)
		return nil, fmt.Errorf("set %s to use the assistant", deepseekAPIKeyEnv)
	}

	request := assistant.ChatRequest{
		Model:     cfg.Model,
		Messages:  messages,
		Stream:    true,
		MaxTokens: cfg.MaxTokens,
	}
	if cfg.Thinking {
		request.Thinking = &assistant.Thinking{Type: "enabled"}
		request.ReasoningEffort = cfg.ReasoningEffort
	} else {
		request.Thinking = &assistant.Thinking{Type: "disabled"}
		temperature := cfg.Temperature
		request.Temperature = &temperature
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("encode chat request: %w", err)
	}

	url := strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions"
	app.log().Debug("chat.http", "posting chat completion",
		"url", url,
		"model", cfg.Model,
		"messages", len(messages),
		"payload_bytes", len(payload))
	app.log().VerboseDebug("chat.http.body", "request payload", "json", string(payload))

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("build chat request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Authorization", "Bearer "+apiKey)

	response, err := app.chatClient.Do(httpRequest)
	if err != nil {
		app.log().Error("chat.http", "request failed", "url", url, "error", err.Error())
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		defer response.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = http.StatusText(response.StatusCode)
		}
		app.log().Error("chat.http", "non-success status",
			"status", response.StatusCode,
			"body", message)
		return nil, fmt.Errorf("deepseek returned status %d: %s", response.StatusCode, message)
	}
	app.log().Debug("chat.http", "stream opened", "status", response.StatusCode)

	return response.Body, nil
}

// buildProblemContext assembles the grounding text injected after the system
// prompt so the assistant knows which problem the user is on.
func buildProblemContext(problem question.ProblemSetProblem, lastResult string) string {
	return buildProblemContextIn(problemsDir, problem, lastResult)
}

func buildProblemContextIn(baseDir string, problem question.ProblemSetProblem, lastResult string) string {
	var b strings.Builder

	b.WriteString("The user is working on this LeetCode problem:\n\n")
	fmt.Fprintf(&b, "Title: %d. %s\n", problem.ID, problem.Title)
	if problem.Difficulty != "" {
		fmt.Fprintf(&b, "Difficulty: %s\n", problem.Difficulty)
	}

	if statement := readProblemFile(baseDir, problem.Slug, problem.Slug+".md"); strings.TrimSpace(statement) != "" {
		b.WriteString("\nProblem statement:\n")
		b.WriteString(strings.TrimSpace(statement))
		b.WriteString("\n")
	}

	if starter := starterFilePath(baseDir, problem); starter != "" {
		if data, err := os.ReadFile(starter); err == nil && strings.TrimSpace(string(data)) != "" {
			b.WriteString("\nThe user's current Go solution:\n```go\n")
			b.WriteString(strings.TrimSpace(string(data)))
			b.WriteString("\n```\n")
		}
	}

	if result := strings.TrimSpace(lastResult); result != "" {
		b.WriteString("\nThe most recent test or submission result:\n")
		b.WriteString(result)
		b.WriteString("\n")
	}

	return b.String()
}

func readProblemFile(baseDir, slug, name string) string {
	data, err := os.ReadFile(filepath.Join(baseDir, slug, name))
	if err != nil {
		return ""
	}
	return string(data)
}
