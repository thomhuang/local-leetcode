package assistant

// Message is one turn in a chat conversation. It matches the OpenAI/DeepSeek
// chat format, so it serializes as {"role": ..., "content": ...}.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest is the body of a streaming chat completion request. Thinking
// controls the DeepSeek thinking mode; when it is disabled the request uses
// Temperature and leaves ReasoningEffort empty.
type ChatRequest struct {
	Model           string    `json:"model"`
	Messages        []Message `json:"messages"`
	Stream          bool      `json:"stream"`
	Thinking        *Thinking `json:"thinking,omitempty"`
	ReasoningEffort string    `json:"reasoning_effort,omitempty"`
	MaxTokens       int       `json:"max_tokens,omitempty"`
	Temperature     *float64  `json:"temperature,omitempty"`
}

// Thinking toggles thinking mode on or off.
type Thinking struct {
	Type string `json:"type"`
}

// StreamChunk is one server-sent event from a streaming completion. Only the
// fields this tool reads are modeled.
type StreamChunk struct {
	Choices []StreamChoice `json:"choices"`
}

type StreamChoice struct {
	Delta        StreamDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

type StreamDelta struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
}
