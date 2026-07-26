package llm

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

// DefaultRetryConfig is the default retry configuration for LLM calls.
var DefaultRetryConfig = &RetryConfig{
	MaxRetries: 3,
	BaseDelay:  1 * time.Second,
	MaxDelay:   30 * time.Second,
	Jitter:     0.2,
	IsRetryable: func(err error) bool {
		// Retry on rate limits and provider unavailability
		return errors.Is(err, ErrRateLimited) || errors.Is(err, ErrProviderUnavailable)
	},
}

// RetryConfig configures the exponential backoff retry behavior.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (excluding initial call).
	MaxRetries int

	// BaseDelay is the initial delay before the first retry.
	BaseDelay time.Duration

	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration

	// Jitter is the fraction of jitter to apply (0.0-1.0).
	// For example, 0.2 means ±20% jitter.
	Jitter float64

	// IsRetryable returns true if the error should be retried.
	// If nil, all errors are retryable.
	IsRetryable func(error) bool
}

// Do executes the given function with retry logic.
// It returns the result and the last error if all retries are exhausted.
func (r *RetryConfig) Do(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	if r.MaxRetries <= 0 {
		return fn()
	}

	var lastErr error

	for attempt := 0; attempt <= r.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if this error is retryable
		if r.IsRetryable != nil && !r.IsRetryable(err) {
			return nil, err
		}

		if attempt < r.MaxRetries {
			delay := r.BackoffDelay(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return nil, lastErr
}

// BackoffDelay calculates the delay for the given retry attempt.
// It uses exponential backoff: baseDelay * 2^attempt + jitter.
func (r *RetryConfig) BackoffDelay(attempt int) time.Duration {
	if r.BaseDelay <= 0 {
		return 0
	}

	// Calculate exponential backoff: base * 2^attempt
	delay := float64(r.BaseDelay) * math.Pow(2, float64(attempt))

	// Apply jitter
	if r.Jitter > 0 {
		jitter := delay * r.Jitter
		delay = delay - jitter + rand.Float64()*2*jitter
	}

	// Clamp to max delay
	if r.MaxDelay > 0 && delay > float64(r.MaxDelay) {
		delay = float64(r.MaxDelay)
	}

	return time.Duration(delay)
}
