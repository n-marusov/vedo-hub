package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	llm "vedo-core/src/services/shared/llm"
)

// anthropicRequest represents an Anthropic Messages API request.
type anthropicRequest struct {
	Model       string         `json:"model"`
	MaxTokens   int            `json:"max_tokens,omitempty"`
	Messages    []anthropicMsg `json:"messages"`
	System      string         `json:"system,omitempty"`
	Temperature float64        `json:"temperature,omitempty"`
}

type anthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse represents an Anthropic Messages API response.
type anthropicResponse struct {
	ID         string              `json:"id"`
	Type       string              `json:"type"`
	Role       string              `json:"role"`
	Content    []anthropicContent  `json:"content"`
	Model      string              `json:"model"`
	StopReason string              `json:"stop_reason"`
	Usage      *anthropicUsage     `json:"usage,omitempty"`
	Error      *anthropicErrorBody `json:"error,omitempty"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicErrorBody struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// AnthropicProvider implements the Provider interface for Anthropic's Messages API.
type AnthropicProvider struct {
	baseURL    string
	apiKey     string
	model      string
	timeout    time.Duration
	httpClient *http.Client
}

// NewAnthropicProvider creates a new provider for Anthropic's Messages API.
func NewAnthropicProvider(cfg llm.Config) *AnthropicProvider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")

	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &AnthropicProvider{
		baseURL: baseURL,
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		timeout: cfg.Timeout,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// init registers the Anthropic provider with the global registry.
func init() {
	llm.Register("anthropic", func(cfg llm.Config) (llm.Provider, error) {
		return NewAnthropicProvider(cfg), nil
	})
}

// Complete sends a prompt to the Anthropic Messages API and returns the completion.
func (p *AnthropicProvider) Complete(ctx llm.Context, prompt llm.Prompt) (llm.Completion, error) {
	start := time.Now()
	log.Printf("[DEBUG] llm/anthropic: sending request model=%q messages=%d", prompt.Model, len(prompt.Messages))

	reqBody := p.buildRequest(prompt)
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return llm.Completion{}, fmt.Errorf("llm/anthropic: failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", p.baseURL+"/messages", bytes.NewReader(jsonBody))
	if err != nil {
		return llm.Completion{}, fmt.Errorf("llm/anthropic: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] llm/anthropic: request failed: %v", err)
		llm.LLMErrorsTotal.WithLabelValues("anthropic", p.model, "connection").Inc()
		return llm.Completion{}, llm.NewErrProviderUnavailable("anthropic", err)
	}
	defer resp.Body.Close()

	// Track latency
	latency := time.Since(start).Seconds()
	llm.LLMLatencySeconds.WithLabelValues("anthropic", p.model).Observe(latency)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return llm.Completion{}, fmt.Errorf("llm/anthropic: failed to read response: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := extractRetryAfter(resp)
		log.Printf("[WARN] llm/anthropic: rate limited (retry after: %ds)", retryAfter)
		llm.LLMErrorsTotal.WithLabelValues("anthropic", p.model, "rate_limited").Inc()
		return llm.Completion{}, llm.NewErrRateLimited("anthropic", retryAfter)
	}

	if resp.StatusCode >= 500 {
		log.Printf("[ERROR] llm/anthropic: server error status=%d", resp.StatusCode)
		llm.LLMErrorsTotal.WithLabelValues("anthropic", p.model, "server_error").Inc()
		return llm.Completion{}, llm.NewErrProviderUnavailable("anthropic",
			fmt.Errorf("server returned HTTP %d", resp.StatusCode))
	}

	if resp.StatusCode >= 400 {
		var errResp anthropicResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != nil {
			log.Printf("[ERROR] llm/anthropic: API error: %s (%s)", errResp.Error.Message, errResp.Error.Type)
			llm.LLMErrorsTotal.WithLabelValues("anthropic", p.model, "api_error").Inc()
			if strings.Contains(errResp.Error.Type, "context_length") {
				return llm.Completion{}, llm.NewErrContextTooLong("anthropic", 0, 0)
			}
			return llm.Completion{}, llm.NewErrInvalidResponse("anthropic", errResp.Error.Message)
		}
	}

	var msgResp anthropicResponse
	if err := json.Unmarshal(body, &msgResp); err != nil {
		log.Printf("[ERROR] llm/anthropic: invalid response JSON: %v", err)
		llm.LLMErrorsTotal.WithLabelValues("anthropic", p.model, "invalid_response").Inc()
		return llm.Completion{}, llm.NewErrInvalidResponse("anthropic", err.Error())
	}

	if len(msgResp.Content) == 0 {
		log.Printf("[WARN] llm/anthropic: empty content in response")
		return llm.Completion{}, nil
	}

	// Concatenate all text content blocks
	var textBuilder strings.Builder
	for _, block := range msgResp.Content {
		if block.Type == "text" {
			textBuilder.WriteString(block.Text)
		}
	}

	completion := llm.Completion{
		Text:         textBuilder.String(),
		TokensUsed:   0,
		FinishReason: parseAnthropicStopReason(msgResp.StopReason),
	}

	if msgResp.Usage != nil {
		completion.PromptTokens = msgResp.Usage.InputTokens
		completion.TokensUsed = msgResp.Usage.OutputTokens
		llm.LLMTokensTotal.WithLabelValues("anthropic", p.model, "prompt").Add(float64(msgResp.Usage.InputTokens))
		llm.LLMTokensTotal.WithLabelValues("anthropic", p.model, "completion").Add(float64(msgResp.Usage.OutputTokens))
	}

	log.Printf("[DEBUG] llm/anthropic: completed model=%q tokens=%d latency=%.2fs",
		p.model, completion.TokensUsed, latency)

	llm.LLMRequestsTotal.WithLabelValues("anthropic", p.model).Inc()

	return completion, nil
}

// StreamComplete sends a prompt and returns a channel of tokens.
func (p *AnthropicProvider) StreamComplete(ctx llm.Context, prompt llm.Prompt) (<-chan llm.Token, error) {
	// Fall back to non-streaming for now.
	completion, err := p.Complete(ctx, prompt)
	if err != nil {
		return nil, err
	}

	ch := make(chan llm.Token, 1)
	ch <- llm.Token{Text: completion.Text, Index: 0}
	close(ch)
	return ch, nil
}

func (p *AnthropicProvider) buildRequest(prompt llm.Prompt) anthropicRequest {
	model := prompt.Model
	if model == "" {
		model = p.model
	}

	maxTokens := prompt.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	var systemPrompt string
	messages := make([]anthropicMsg, 0)

	for _, msg := range prompt.Messages {
		switch msg.Role {
		case llm.RoleSystem:
			systemPrompt = msg.Content
		default:
			messages = append(messages, anthropicMsg{
				Role:    string(msg.Role),
				Content: msg.Content,
			})
		}
	}

	if len(messages) == 0 {
		messages = append(messages, anthropicMsg{
			Role:    "user",
			Content: "",
		})
	}

	return anthropicRequest{
		Model:       model,
		MaxTokens:   maxTokens,
		Messages:    messages,
		System:      systemPrompt,
		Temperature: prompt.Temperature,
	}
}

func parseAnthropicStopReason(reason string) llm.FinishReason {
	switch reason {
	case "end_turn", "stop_sequence":
		return llm.FinishReasonStop
	case "max_tokens":
		return llm.FinishReasonLength
	default:
		return llm.FinishReasonStop
	}
}
