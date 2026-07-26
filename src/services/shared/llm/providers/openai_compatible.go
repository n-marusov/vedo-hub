package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	llm "vedo-core/src/services/shared/llm"
)

// openAIChatRequest represents an OpenAI chat completions request.
type openAIChatRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIChatResponse represents an OpenAI chat completions response.
type openAIChatResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Choices []openAIChoice   `json:"choices"`
	Usage   *openAIUsage     `json:"usage,omitempty"`
	Error   *openAIErrorBody `json:"error,omitempty"`
}

type openAIChoice struct {
	Index        int               `json:"index"`
	Message      openAIRespMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

type openAIRespMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type openAIErrorBody struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// OpenAICompatibleProvider implements the Provider interface for any
// OpenAI-compatible API (OpenAI, Ollama, vLLM, local LLMs, etc.).
type OpenAICompatibleProvider struct {
	baseURL    string
	apiKey     string
	model      string
	timeout    time.Duration
	httpClient *http.Client
}

// NewOpenAICompatibleProvider creates a new provider for OpenAI-compatible APIs.
func NewOpenAICompatibleProvider(cfg llm.Config) *OpenAICompatibleProvider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &OpenAICompatibleProvider{
		baseURL: baseURL,
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		timeout: cfg.Timeout,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// init registers the OpenAI-compatible provider with the global registry.
func init() {
	llm.Register("openai", func(cfg llm.Config) (llm.Provider, error) {
		return NewOpenAICompatibleProvider(cfg), nil
	})
}

// Complete sends a prompt to the OpenAI-compatible API and returns the completion.
func (p *OpenAICompatibleProvider) Complete(ctx llm.Context, prompt llm.Prompt) (llm.Completion, error) {
	start := time.Now()
	log.Printf("[DEBUG] llm/openai: sending request model=%q messages=%d", prompt.Model, len(prompt.Messages))

	reqBody := p.buildRequest(prompt)
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return llm.Completion{}, fmt.Errorf("llm/openai: failed to marshal request: %w", err)
	}

	// Fall back to Background() if context is nil (test-safe)
	baseCtx := ctx.BaseCtx
	if baseCtx == nil {
		baseCtx = context.Background()
	}

	req, err := http.NewRequestWithContext(baseCtx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return llm.Completion{}, fmt.Errorf("llm/openai: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] llm/openai: request failed: %v", err)
		llm.LLMErrorsTotal.WithLabelValues("openai", p.model, "connection").Inc()
		return llm.Completion{}, llm.NewErrProviderUnavailable("openai", err)
	}
	defer resp.Body.Close()

	// Track latency
	latency := time.Since(start).Seconds()
	llm.LLMLatencySeconds.WithLabelValues("openai", p.model).Observe(latency)

	// Limit response body size to prevent unbounded memory allocation
	const maxResponseSize = 10 * 1024 * 1024 // 10MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return llm.Completion{}, fmt.Errorf("llm/openai: failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := extractRetryAfter(resp)
		log.Printf("[WARN] llm/openai: rate limited (retry after: %ds)", retryAfter)
		llm.LLMErrorsTotal.WithLabelValues("openai", p.model, "rate_limited").Inc()
		return llm.Completion{}, llm.NewErrRateLimited("openai", retryAfter)
	}

	if resp.StatusCode >= 500 {
		log.Printf("[ERROR] llm/openai: server error status=%d", resp.StatusCode)
		llm.LLMErrorsTotal.WithLabelValues("openai", p.model, "server_error").Inc()
		return llm.Completion{}, llm.NewErrProviderUnavailable("openai",
			fmt.Errorf("server returned HTTP %d", resp.StatusCode))
	}

	var chatResp openAIChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		log.Printf("[ERROR] llm/openai: invalid response JSON: %v", err)
		llm.LLMErrorsTotal.WithLabelValues("openai", p.model, "invalid_response").Inc()
		return llm.Completion{}, llm.NewErrInvalidResponse("openai", err.Error())
	}

	if chatResp.Error != nil {
		if isContextLengthError(chatResp.Error) {
			return llm.Completion{}, llm.NewErrContextTooLong("openai", 0, 0)
		}
		log.Printf("[ERROR] llm/openai: API error: %s (%s)", chatResp.Error.Message, chatResp.Error.Type)
		llm.LLMErrorsTotal.WithLabelValues("openai", p.model, "api_error").Inc()
		return llm.Completion{}, llm.NewErrInvalidResponse("openai", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		log.Printf("[WARN] llm/openai: empty choices in response")
		return llm.Completion{}, nil
	}

	choice := chatResp.Choices[0]
	completion := llm.Completion{
		Text:         choice.Message.Content,
		TokensUsed:   0,
		FinishReason: parseFinishReason(choice.FinishReason),
	}

	if chatResp.Usage != nil {
		completion.PromptTokens = chatResp.Usage.PromptTokens
		completion.TokensUsed = chatResp.Usage.CompletionTokens
		llm.LLMTokensTotal.WithLabelValues("openai", p.model, "prompt").Add(float64(chatResp.Usage.PromptTokens))
		llm.LLMTokensTotal.WithLabelValues("openai", p.model, "completion").Add(float64(chatResp.Usage.CompletionTokens))
	}

	log.Printf("[DEBUG] llm/openai: completed model=%q tokens=%d latency=%.2fs",
		p.model, completion.TokensUsed, latency)

	llm.LLMRequestsTotal.WithLabelValues("openai", p.model).Inc()

	return completion, nil
}

// StreamComplete sends a prompt and returns a channel of tokens via SSE.
func (p *OpenAICompatibleProvider) StreamComplete(ctx llm.Context, prompt llm.Prompt) (<-chan llm.Token, error) {
	// For now, fall back to non-streaming and simulate a single token.
	// Full SSE streaming will be implemented in a follow-up.
	completion, err := p.Complete(ctx, prompt)
	if err != nil {
		return nil, err
	}

	ch := make(chan llm.Token, 1)
	ch <- llm.Token{Text: completion.Text, Index: 0}
	close(ch)
	return ch, nil
}

func (p *OpenAICompatibleProvider) buildRequest(prompt llm.Prompt) openAIChatRequest {
	model := prompt.Model
	if model == "" {
		model = p.model
	}

	messages := make([]openAIMessage, len(prompt.Messages))
	for i, msg := range prompt.Messages {
		messages[i] = openAIMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	return openAIChatRequest{
		Model:       model,
		Messages:    messages,
		Temperature: prompt.Temperature,
		MaxTokens:   prompt.MaxTokens,
	}
}

func extractRetryAfter(resp *http.Response) int {
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter != "" {
		var seconds int
		if _, err := fmt.Sscanf(retryAfter, "%d", &seconds); err == nil {
			return seconds
		}
	}
	return 30
}

func parseFinishReason(reason string) llm.FinishReason {
	switch reason {
	case "stop":
		return llm.FinishReasonStop
	case "length":
		return llm.FinishReasonLength
	case "content_filter":
		return llm.FinishReasonContentFilter
	case "tool_calls":
		return llm.FinishReasonToolCalls
	default:
		return llm.FinishReasonStop
	}
}

func isContextLengthError(err *openAIErrorBody) bool {
	return strings.Contains(err.Type, "context_length_exceeded") ||
		strings.Contains(err.Code, "context_length_exceeded")
}
