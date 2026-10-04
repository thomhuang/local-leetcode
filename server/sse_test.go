package main

import (
	"strings"
	"testing"
)

func TestSSEReaderStreamsContentAndReasoning(t *testing.T) {
	stream := strings.Join([]string{
		": keep-alive",
		"",
		`data: {"choices":[{"delta":{"reasoning_content":"thinking"},"finish_reason":null}]}`,
		"",
		`data: {"choices":[{"delta":{"content":"Hello"},"finish_reason":null}]}`,
		"",
		`data: {"choices":[{"delta":{"content":" world"},"finish_reason":"stop"}]}`,
		"",
		"data: [DONE]",
		"",
	}, "\n")

	reader := newSSEReader(strings.NewReader(stream))

	var content, reasoning strings.Builder
	for {
		chunk, done, err := reader.Next()
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		if done {
			break
		}
		if len(chunk.Choices) > 0 {
			content.WriteString(chunk.Choices[0].Delta.Content)
			reasoning.WriteString(chunk.Choices[0].Delta.ReasoningContent)
		}
	}

	if content.String() != "Hello world" {
		t.Fatalf("content = %q, want %q", content.String(), "Hello world")
	}
	if reasoning.String() != "thinking" {
		t.Fatalf("reasoning = %q, want %q", reasoning.String(), "thinking")
	}
}

func TestSSEReaderEndsAtEOF(t *testing.T) {
	reader := newSSEReader(strings.NewReader(`data: {"choices":[]}` + "\n"))

	if _, done, err := reader.Next(); err != nil || done {
		t.Fatalf("first event: done=%v err=%v, want a live event", done, err)
	}
	if _, done, err := reader.Next(); err != nil || !done {
		t.Fatalf("second event: done=%v err=%v, want done", done, err)
	}
}

func TestSSEReaderReportsBadJSON(t *testing.T) {
	reader := newSSEReader(strings.NewReader("data: not-json\n"))
	if _, _, err := reader.Next(); err == nil {
		t.Fatalf("expected a parse error")
	}
}
