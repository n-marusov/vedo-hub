package providers_test

// Validates: REQ-FUN.API.llm-policy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	llm "vedo-core/src/services/shared/llm"
	"vedo-core/src/services/shared/llm/providers"
)

// anthropicResponse is a minimal Anthropic Messages API response.
type anthropicResponse struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Role       string             `json:"role"`
	Content    []anthropicContent `json:"content"`
	Model      string             `json:"model"`
	StopReason string             `json:"stop_reason"`
	Usage      anthropicUsage     `json:"usage"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

func TestNewAnthropicProvider(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "",
		APIKey:  "sk-ant-test-key",
		Model:   "claude-3-haiku-20240307",
	}
	provider := providers.NewAnthropicProvider(cfg)
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestAnthropicProviderComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("x-api-key") != "sk-ant-test-key" {
			t.Errorf("expected x-api-key header, got '%s'", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Error("expected anthropic-version header")
		}

		resp := map[string]interface{}{
			"id":   "msg_123",
			"type": "message",
			"role": "assistant",
			"content": []map[string]interface{}{
				{"type": "text", "text": "Generated ontology class: Person"},
			},
			"model":       "claude-3-opus-20240229",
			"stop_reason": "end_turn",
			"usage": map[string]interface{}{
				"input_tokens":  45,
				"output_tokens": 12,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := providers.NewAnthropicProvider(llm.Config{
		BaseURL: server.URL,
		APIKey:  "sk-ant-test-key",
		Model:   "claude-3-opus-20240229",
	})

	comp, err := provider.Complete(llm.Context{}, llm.Prompt{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "Generate ontology class"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comp.Text != "Generated ontology class: Person" {
		t.Errorf("expected 'Generated ontology class: Person', got '%s'", comp.Text)
	}
	if comp.PromptTokens != 45 {
		t.Errorf("expected 45 prompt tokens, got %d", comp.PromptTokens)
	}
	if comp.TokensUsed != 12 {
		t.Errorf("expected 12 completion tokens, got %d", comp.TokensUsed)
	}
}

func TestAnthropicProviderRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Header().Set("Retry-After", "60")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"type":    "rate_limit_error",
				"message": "Number of request tokens has exceeded your rate limit.",
			},
		})
	}))
	defer server.Close()

	provider := providers.NewAnthropicProvider(llm.Config{
		BaseURL: server.URL,
		APIKey:  "sk-ant-test-key",
	})

	_, err := provider.Complete(llm.Context{}, llm.Prompt{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "test"}},
	})
	if err == nil {
		t.Fatal("expected rate limit error")
	}
}

func TestAnthropicProviderServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := providers.NewAnthropicProvider(llm.Config{
		BaseURL: server.URL,
		APIKey:  "sk-ant-test-key",
	})

	_, err := provider.Complete(llm.Context{}, llm.Prompt{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "test"}},
	})
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestAnthropicProviderEmptyContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "msg_empty",
			"type":        "message",
			"role":        "assistant",
			"content":     []interface{}{},
			"model":       "claude-3-haiku",
			"stop_reason": "end_turn",
			"usage":       map[string]interface{}{},
		})
	}))
	defer server.Close()

	provider := providers.NewAnthropicProvider(llm.Config{
		BaseURL: server.URL,
		APIKey:  "sk-ant-test-key",
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

func TestAnthropicProviderSystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify system prompt handling
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "msg_sys",
			"type": "message",
			"role": "assistant",
			"content": []map[string]interface{}{
				{"type": "text", "text": "Response with system context"},
			},
			"model":       "claude-3-haiku",
			"stop_reason": "end_turn",
			"usage":       map[string]interface{}{},
		})
	}))
	defer server.Close()

	provider := providers.NewAnthropicProvider(llm.Config{
		BaseURL: server.URL,
		APIKey:  "sk-ant-test-key",
	})

	comp, err := provider.Complete(llm.Context{}, llm.Prompt{
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "You are an ontology expert."},
			{Role: llm.RoleUser, Content: "Generate class hierarchy"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comp.Text != "Response with system context" {
		t.Errorf("expected 'Response with system context', got '%s'", comp.Text)
	}
}
