package llm

import (
	"errors"
	"fmt"
)

// Sentinel errors for LLM provider interactions.
var (
	// ErrProviderUnavailable indicates the LLM provider is unreachable or down.
	ErrProviderUnavailable = errors.New("llm: provider unavailable")

	// ErrRateLimited indicates the provider returned a rate limit error.
	ErrRateLimited = errors.New("llm: rate limited")

	// ErrContextTooLong indicates the prompt exceeds the model's context window.
	ErrContextTooLong = errors.New("llm: context too long")

	// ErrInvalidResponse indicates the provider returned an unexpected or malformed response.
	ErrInvalidResponse = errors.New("llm: invalid response")
)

// LLMError is a generic LLM error with provider context.
type LLMError struct {
	Provider string
	Err      error
}

func (e *LLMError) Error() string {
	return fmt.Sprintf("llm: %s: %v", e.Provider, e.Err)
}

func (e *LLMError) Unwrap() error {
	return e.Err
}

// RateLimitError indicates the provider returned a rate limit response.
type RateLimitError struct {
	Provider   string
	RetryAfter int // seconds
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("llm: %s: rate limited, retry after %d seconds", e.Provider, e.RetryAfter)
}

func (e *RateLimitError) Unwrap() error {
	return ErrRateLimited
}

// ContextTooLongError indicates the prompt exceeds the model's context window.
type ContextTooLongError struct {
	Provider  string
	Requested int
	Limit     int
}

func (e *ContextTooLongError) Error() string {
	return fmt.Sprintf("llm: %s: context too long (%d tokens requested, limit %d)", e.Provider, e.Requested, e.Limit)
}

func (e *ContextTooLongError) Unwrap() error {
	return ErrContextTooLong
}

// InvalidResponseError indicates an unexpected or malformed response from the provider.
type InvalidResponseError struct {
	Provider string
	Details  string
}

func (e *InvalidResponseError) Error() string {
	return fmt.Sprintf("llm: %s: invalid response: %s", e.Provider, e.Details)
}

func (e *InvalidResponseError) Unwrap() error {
	return ErrInvalidResponse
}

// NewErrProviderUnavailable creates a new ErrProviderUnavailable wrapped error.
func NewErrProviderUnavailable(provider string, err error) error {
	return &LLMError{
		Provider: provider,
		Err:      fmt.Errorf("%w: %v", ErrProviderUnavailable, err),
	}
}

// NewErrRateLimited creates a new ErrRateLimited wrapped error.
func NewErrRateLimited(provider string, retryAfter int) error {
	return &RateLimitError{
		Provider:   provider,
		RetryAfter: retryAfter,
	}
}

// NewErrContextTooLong creates a new ErrContextTooLong wrapped error.
func NewErrContextTooLong(provider string, requested, limit int) error {
	return &ContextTooLongError{
		Provider:  provider,
		Requested: requested,
		Limit:     limit,
	}
}

// NewErrInvalidResponse creates a new ErrInvalidResponse wrapped error.
func NewErrInvalidResponse(provider, details string) error {
	return &InvalidResponseError{
		Provider: provider,
		Details:  details,
	}
}
