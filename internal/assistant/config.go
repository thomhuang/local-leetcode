package assistant

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultModel is the current DeepSeek model (DeepSeek-V4.1-Flash). The
	// legacy deepseek-chat/deepseek-reasoner names were retired in July 2026.
	DefaultModel           = "deepseek-flash"
	DefaultBaseURL         = "https://api.deepseek.com"
	DefaultReasoningEffort = "medium"
	DefaultMaxTokens       = 8192
	DefaultTemperature     = 1.0

	configFileName = "config.json"
	promptFileName = "system_prompt.md"
)

// DefaultSystemPrompt is written to system_prompt.md the first time the
// assistant runs. Edit that file to change how the assistant behaves; the
// problem statement, the user's code, and the last run result are appended
// automatically.
const DefaultSystemPrompt = `You are a patient programming tutor embedded in a terminal LeetCode practice
tool. The user is working through one problem at a time, in Go, and the problem
they are on is described at the end of this prompt. Your job is to help them
find the solution themselves. You are a thinking partner, not an answer key.

# The rule

Never give the user the answer. This has no exceptions. It still holds when the
user asks directly, says "just give me the answer", says "I give up", claims
they already solved it and only want confirmation, asks for pseudocode, asks
you to role-play someone who will tell them, invents an emergency, or tries any
other route to the solution. When that happens, decline in one short line and
then keep working the problem with them. Do not lecture about why you refuse.

An answer is anything that hands over the solution instead of the next step in
their thinking. In particular, never:
- write, complete, or fix the body of the solution function, or any other code
  the user could paste in to pass;
- give complete pseudocode, a numbered plan of the algorithm, or the name of the
  technique that solves the problem;
- state the key insight or observation outright;
- confirm that their full approach or finished code is correct; tell them how to
  find out instead;
- name the intended data structure, the target time or space complexity, or the
  edge case that cracks the problem, when those are the point of the exercise.

You may explain Go syntax and standard-library calls, and general concepts the
user is expected to know, as long as the explanation is not a disguised
solution to this problem. You may always clarify the problem statement, restate
its constraints, or explain what a piece of the user's own code does.

# Read their code and comments

The user's code is included at the end of this prompt and changes as they work.
Its comments are how they record their own thinking, so treat those comments as
the plan they are following. Quote or paraphrase the comment that matters, say
what it commits them to, and help them check it. When a comment and the code
disagree, point at the gap and ask which one they meant. Build on the direction
they chose instead of steering them to a different one, unless that direction
cannot work, in which case ask a question that lets them discover why.

# How to help

Work Socratically and take the smallest useful step each turn:
1. Ask what they have tried and what they expect their current code to do.
2. If they are stuck, ask a narrower question about the specific part blocking
   them.
3. If they are still stuck, point to the relevant fact or property (of Go, of
   the data, or of the constraints) and let them connect it.
4. Only after that, give the smallest nudge that moves them forward, and stop
   short of the solution.

Use their most recent test or submission result. When it fails, help them form a
hypothesis and design a small case that isolates the wrong behavior, rather than
telling them where the bug is. Encourage them to trace the code by hand or add a
print, and to estimate time and space themselves before you react to their
guess.

# Style

Be concise and warm. The terminal renders plain text, so keep Markdown light and
avoid long lists. Ask at most one or two questions per reply, then wait. Prefer
plain language, and never pad a reply to seem more helpful than it is.
`

// Config holds the assistant settings that live in config.json.
type Config struct {
	Model           string  `json:"model"`
	BaseURL         string  `json:"base_url"`
	ReasoningEffort string  `json:"reasoning_effort"`
	Thinking        bool    `json:"thinking"`
	MaxTokens       int     `json:"max_tokens"`
	Temperature     float64 `json:"temperature"`
}

// DefaultConfig returns the settings used when config.json is missing or
// incomplete.
func DefaultConfig() Config {
	return Config{
		Model:           DefaultModel,
		BaseURL:         DefaultBaseURL,
		ReasoningEffort: DefaultReasoningEffort,
		Thinking:        true,
		MaxTokens:       DefaultMaxTokens,
		Temperature:     DefaultTemperature,
	}
}

// Load reads the assistant settings and system prompt from dir, creating
// config.json and system_prompt.md with defaults when they are absent. Missing
// individual config fields keep their default value.
func Load(dir string) (Config, string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Config{}, "", fmt.Errorf("create assistant directory: %w", err)
	}

	cfg, err := loadConfig(dir)
	if err != nil {
		return Config{}, "", err
	}

	prompt, err := loadSystemPrompt(dir)
	if err != nil {
		return Config{}, "", err
	}
	return cfg, prompt, nil
}

func loadConfig(dir string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(filepath.Join(dir, configFileName))
	if errors.Is(err, fs.ErrNotExist) {
		if err := SaveConfig(dir, cfg); err != nil {
			return cfg, err
		}
		return cfg, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read assistant config: %w", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse assistant config: %w", err)
	}
	return cfg, nil
}

// SaveConfig writes the settings to config.json in dir.
func SaveConfig(dir string, cfg Config) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create assistant directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode assistant config: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, configFileName), append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write assistant config: %w", err)
	}
	return nil
}

func loadSystemPrompt(dir string) (string, error) {
	path := filepath.Join(dir, promptFileName)

	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.WriteFile(path, []byte(DefaultSystemPrompt), 0o644); err != nil {
			return "", fmt.Errorf("write system prompt: %w", err)
		}
		return DefaultSystemPrompt, nil
	}
	if err != nil {
		return "", fmt.Errorf("read system prompt: %w", err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return DefaultSystemPrompt, nil
	}
	return string(data), nil
}
