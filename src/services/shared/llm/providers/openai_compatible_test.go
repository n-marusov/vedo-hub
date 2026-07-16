package providers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	llm "vedo-core/src/services/shared/llm"
	"vedo-core/src/services/shared/llm/providers"
)

// openAIResponse is a minimal OpenAI chat completion response.
type openAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage"`
}

type openAIChoice struct {
	Index        int           `json:"index"`
	Message      openAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

func TestNewOpenAICompatibleProvider(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost:11434/v1",
		APIKey:  "",
		Model:   "llama3",
	}
	provider := providers.NewOpenAICompatibleProvider(cfg)
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestOpenAIProviderDefaultBaseURL(t *testing.T) {
	// Default base URL is applied when BaseURL is empty
	cfg := llm.Config{Model: "gpt-4o"}
	provider := providers.NewOpenAICompatibleProvider(cfg)
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
	// Verify via Complete call to the default URL
	_, err := provider.Complete(llm.Context{}, llm.Prompt{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "test"}},
	})
	// Expected to fail (no server) but not nil panic
	if err == nil {
		t.Fatal("expected connection error to default https://api.openai.com")
	}
}

func TestOpenAIProviderComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Authorization header, got '%s'", r.Header.Get("Authorization"))
		}

		resp := map[string]interface{}{
			"id":     "chatcmpl-123",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Extracted class: Person",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     50,
				"completion_tokens": 10,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := providers.NewOpenAICompatibleProvider(llm.Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "gpt-4o-mini",
	})

	comp, err := provider.Complete(llm.Context{}, llm.Prompt{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "Extract ontology"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comp.Text != "Extracted class: Person" {
		t.Errorf("expected 'Extracted class: Person', got '%s'", comp.Text)
	}
	if comp.PromptTokens != 50 {
		t.Errorf("expected 50 prompt tokens, got %d", comp.PromptTokens)
	}
	if comp.TokensUsed != 10 {
		t.Errorf("expected 10 completion tokens, got %d", comp.TokensUsed)
	}
}

func TestOpenAIProviderRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Header().Set("Retry-After", "30")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Rate limit exceeded",
				"type":    "rate_limit_error",
			},
		})
	}))
	defer server.Close()

	provider := providers.NewOpenAICompatibleProvider(llm.Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "gpt-4o",
	})

	_, err := provider.Complete(llm.Context{}, llm.Prompt{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "test"}},
	})
	if err == nil {
		t.Fatal("expected rate limit error")
	}

	var rateErr *llm.RateLimitError
	if !errors.As(err, &rateErr) {
		t.Fatalf("expected RateLimitError, got %T", err)
	}
	if rateErr.RetryAfter != 30 {
		t.Errorf("expected RetryAfter 30, got %d", rateErr.RetryAfter)
	}
}

func TestOpenAIProviderServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := providers.NewOpenAICompatibleProvider(llm.Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "gpt-4o",
	})

	_, err := provider.Complete(llm.Context{}, llm.Prompt{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "test"}},
	})
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestOpenAIProviderEmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "chatcmpl-empty",
			"object":  "chat.completion",
			"choices": []interface{}{},
			"usage":   map[string]interface{}{},
		})
	}))
	defer server.Close()

	provider := providers.NewOpenAICompatibleProvider(llm.Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "gpt-4o",
	})

	comp, err := provider.Complete(llm.Context{}, llm.Prompt{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "test"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comp.Text != "" {
		t.Errorf("expected empty text, got '%s'", comp.Text)
	}
}

func TestOpenAIProviderStreamComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "chatcmpl-s",
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "streamed response",
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{},
		})
	}))
	defer server.Close()

	provider := providers.NewOpenAICompatibleProvider(llm.Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
	})

	ch, err := provider.StreamComplete(llm.Context{}, llm.Prompt{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "test"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tok, ok := <-ch
	if !ok {
		t.Fatal("expected token on channel")
	}
	if tok.Text != "streamed response" {
		t.Errorf("expected 'streamed response', got '%s'", tok.Text)
	}
}
