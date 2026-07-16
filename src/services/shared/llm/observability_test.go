package llm

import (
	"errors"
	"testing"
)

func TestObservabilityWrapsProvider(t *testing.T) {
	inner := &mockProvider{
		completeFunc: func(ctx Context, p Prompt) (Completion, error) {
			return Completion{Text: "test response", TokensUsed: 10, PromptTokens: 5}, nil
		},
	}

	wrapped := WrapWithObservability("openai", inner)
	if wrapped == nil {
		t.Fatal("expected non-nil wrapped provider")
	}
}

func TestObservabilityCompletePassthrough(t *testing.T) {
	inner := &mockProvider{
		completeFunc: func(ctx Context, p Prompt) (Completion, error) {
			return Completion{Text: "passthrough", TokensUsed: 5}, nil
		},
	}

	wrapped := WrapWithObservability("openai", inner)
	comp, err := wrapped.Complete(Context{}, Prompt{
		Messages: []Message{{Role: RoleUser, Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comp.Text != "passthrough" {
		t.Errorf("expected 'passthrough', got '%s'", comp.Text)
	}
}

func TestObservabilityErrorWrapping(t *testing.T) {
	inner := &mockProvider{
		completeFunc: func(ctx Context, p Prompt) (Completion, error) {
			return Completion{}, NewErrProviderUnavailable("openai", errors.New("connection failed"))
		},
	}

	wrapped := WrapWithObservability("openai", inner)
	_, err := wrapped.Complete(Context{}, Prompt{
		Messages: []Message{{Role: RoleUser, Content: "test"}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestObservabilityStreamComplete(t *testing.T) {
	inner := &mockProvider{
		streamCompleteFunc: func(ctx Context, p Prompt) (<-chan Token, error) {
			ch := make(chan Token, 1)
			ch <- Token{Text: "streaming", Index: 0}
			close(ch)
			return ch, nil
		},
	}

	wrapped := WrapWithObservability("anthropic", inner)
	ch, err := wrapped.StreamComplete(Context{}, Prompt{
		Messages: []Message{{Role: RoleUser, Content: "test"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tok, ok := <-ch
	if !ok {
		t.Fatal("expected token on channel")
	}
	if tok.Text != "streaming" {
		t.Errorf("expected 'streaming', got '%s'", tok.Text)
	}
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		err    error
		wanted string
	}{
		{NewErrRateLimited("test", 10), "rate_limited"},
		{NewErrProviderUnavailable("test", errors.New("connection timeout")), "timeout"},
		{NewErrContextTooLong("test", 100, 50), "context_too_long"},
		{NewErrInvalidResponse("test", "bad json"), "invalid_response"},
		{errors.New("unknown error"), "unknown"},
	}

	for _, tt := range tests {
		got := classifyError(tt.err)
		if got != tt.wanted {
			t.Errorf("classifyError(%v) = %q, want %q", tt.err, got, tt.wanted)
		}
	}
}

func TestEstimateCostKnownModel(t *testing.T) {
	cost := EstimateCost("openai", "gpt-4o-mini", 1000, 500)
	if cost <= 0 {
		t.Errorf("expected positive cost, got %f", cost)
	}

	// gpt-4o-mini: $0.00015/1K prompt, $0.0006/1K completion
	// 1000 prompt tokens: $0.00015
	// 500 completion tokens: $0.0003
	// Total: $0.00045
	expected := 0.00045
	if cost != expected {
		t.Errorf("expected cost %f, got %f", expected, cost)
	}
}

func TestEstimateCostUnknownProvider(t *testing.T) {
	cost := EstimateCost("unknown", "model", 100, 50)
	if cost != 0 {
		t.Errorf("expected 0 cost for unknown provider, got %f", cost)
	}
}

func TestEstimateCostLocalModel(t *testing.T) {
	cost := EstimateCost("ollama", "llama3", 1000, 500)
	if cost != 0 {
		t.Errorf("expected 0 cost for local model, got %f", cost)
	}
}

func TestGetModelPricing(t *testing.T) {
	pricing := GetModelPricing("openai", "gpt-4o")
	if pricing == nil {
		t.Fatal("expected non-nil pricing")
	}
	if pricing.PromptPrice <= 0 {
		t.Errorf("expected positive prompt price, got %f", pricing.PromptPrice)
	}
}

func TestGetModelPricingUnknown(t *testing.T) {
	pricing := GetModelPricing("unknown", "model")
	if pricing != nil {
		t.Fatal("expected nil pricing for unknown provider")
	}
}

func TestResolveModel(t *testing.T) {
	if resolved := resolveModel("gpt-4o", "gpt-4o-mini"); resolved != "gpt-4o" {
		t.Errorf("expected prompt model to win, got '%s'", resolved)
	}
	if resolved := resolveModel("", "gpt-4o-mini"); resolved != "gpt-4o-mini" {
		t.Errorf("expected provider model fallback, got '%s'", resolved)
	}
}

func TestLLMCostCounterInitialized(t *testing.T) {
	// Ensure the cost counter metric is properly initialized
	if LLMCostTotal == nil {
		t.Fatal("expected non-nil LLMCostTotal metric")
	}
}
