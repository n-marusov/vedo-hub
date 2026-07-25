package llm

import "context"

// MessageRole represents the role of a message in a conversation.
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// FinishReason indicates why the LLM stopped generating.
type FinishReason string

const (
	FinishReasonStop          FinishReason = "stop"
	FinishReasonLength        FinishReason = "length"
	FinishReasonContentFilter FinishReason = "content_filter"
	FinishReasonToolCalls     FinishReason = "tool_calls"
)

// Message represents a single message in a conversation with the LLM.
type Message struct {
	Role    MessageRole `json:"role"`
	Content string      `json:"content"`
}

// Prompt represents a request to an LLM.
type Prompt struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Completion represents the response from an LLM.
type Completion struct {
	Text         string       `json:"text"`
	TokensUsed   int          `json:"tokens_used"`
	PromptTokens int          `json:"prompt_tokens"`
	FinishReason FinishReason `json:"finish_reason"`
}

// Token represents a single token in a streaming response.
type Token struct {
	Text  string `json:"text"`
	Index int    `json:"index"`
}

// Context carries request-scoped values for LLM calls.
// It embeds context.Context for cancellation and deadlines.
type Context struct {
	BaseCtx      context.Context
	TraceID      string
	MaxRetries   int
	SkipTracking bool
}

// ProviderConfig holds configuration for a named provider instance.
type ProviderConfig struct {
	Name       string
	APIKey     string
	BaseURL    string
	Model      string
	Timeout    int // seconds
	MaxRetries int
}
