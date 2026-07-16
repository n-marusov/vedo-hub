package llm

import (
	"errors"
	"testing"
)

func TestErrProviderUnavailable(t *testing.T) {
	err := NewErrProviderUnavailable("openai", errors.New("connection refused"))
	if !errors.Is(err, ErrProviderUnavailable) {
		t.Error("expected ErrProviderUnavailable in error chain")
	}

	var llmErr *LLMError
	if !errors.As(err, &llmErr) {
		t.Fatal("expected error to be of type *LLMError")
	}
	if llmErr.Provider != "openai" {
		t.Errorf("expected provider 'openai', got '%s'", llmErr.Provider)
	}
	if llmErr.Err == nil {
		t.Error("expected wrapped error")
	}
}

func TestErrRateLimited(t *testing.T) {
	err := NewErrRateLimited("anthropic", 30)
	if !errors.Is(err, ErrRateLimited) {
		t.Error("expected ErrRateLimited in error chain")
	}

	var rateErr *RateLimitError
	if !errors.As(err, &rateErr) {
		t.Fatal("expected error to be of type *RateLimitError")
	}
	if rateErr.Provider != "anthropic" {
		t.Errorf("expected provider 'anthropic', got '%s'", rateErr.Provider)
	}
	if rateErr.RetryAfter != 30 {
		t.Errorf("expected retry after 30s, got %d", rateErr.RetryAfter)
	}
}

func TestErrContextTooLong(t *testing.T) {
	err := NewErrContextTooLong("gpt-4o", 100000, 8192)
	if !errors.Is(err, ErrContextTooLong) {
		t.Error("expected ErrContextTooLong in error chain")
	}

	var ctxErr *ContextTooLongError
	if !errors.As(err, &ctxErr) {
		t.Fatal("expected error to be of type *ContextTooLongError")
	}
	if ctxErr.Provider != "gpt-4o" {
		t.Errorf("expected provider 'gpt-4o', got '%s'", ctxErr.Provider)
	}
	if ctxErr.Requested != 100000 {
		t.Errorf("expected requested 100000 tokens, got %d", ctxErr.Requested)
	}
	if ctxErr.Limit != 8192 {
		t.Errorf("expected limit 8192 tokens, got %d", ctxErr.Limit)
	}
}

func TestErrInvalidResponse(t *testing.T) {
	err := NewErrInvalidResponse("ollama", "unexpected EOF")
	if !errors.Is(err, ErrInvalidResponse) {
		t.Error("expected ErrInvalidResponse in error chain")
	}

	var respErr *InvalidResponseError
	if !errors.As(err, &respErr) {
		t.Fatal("expected error to be of type *InvalidResponseError")
	}
	if respErr.Provider != "ollama" {
		t.Errorf("expected provider 'ollama', got '%s'", respErr.Provider)
	}
	if respErr.Details != "unexpected EOF" {
		t.Errorf("expected details 'unexpected EOF', got '%s'", respErr.Details)
	}
}

func TestLLMErrorSentinelValues(t *testing.T) {
	// Verify sentinel errors are not nil
	if ErrProviderUnavailable == nil {
		t.Error("ErrProviderUnavailable should not be nil")
	}
	if ErrRateLimited == nil {
		t.Error("ErrRateLimited should not be nil")
	}
	if ErrContextTooLong == nil {
		t.Error("ErrContextTooLong should not be nil")
	}
	if ErrInvalidResponse == nil {
		t.Error("ErrInvalidResponse should not be nil")
	}
}

func TestErrorTypeAssertions(t *testing.T) {
	llmErr := NewErrProviderUnavailable("test", errors.New("cause")).(*LLMError)
	if llmErr.Error() == "" {
		t.Error("error message should not be empty")
	}

	rateErr := NewErrRateLimited("test", 10).(*RateLimitError)
	if rateErr.Error() == "" {
		t.Error("error message should not be empty")
	}

	ctxErr := NewErrContextTooLong("test", 100, 50).(*ContextTooLongError)
	if ctxErr.Error() == "" {
		t.Error("error message should not be empty")
	}

	respErr := NewErrInvalidResponse("test", "bad data").(*InvalidResponseError)
	if respErr.Error() == "" {
		t.Error("error message should not be empty")
	}
}
