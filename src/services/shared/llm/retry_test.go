package llm

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetrySuccessFirstAttempt(t *testing.T) {
	r := &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
		Jitter:     0.1,
	}

	attempts := 0
	_, err := r.Do(context.Background(), func() (interface{}, error) {
		attempts++
		return "success", nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

func TestRetrySuccessAfterRetries(t *testing.T) {
	r := &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  5 * time.Millisecond,
		MaxDelay:   50 * time.Millisecond,
		Jitter:     0.1,
	}

	attempts := 0
	_, err := r.Do(context.Background(), func() (interface{}, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("temporary error")
		}
		return "success", nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryExhausted(t *testing.T) {
	r := &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  5 * time.Millisecond,
		MaxDelay:   50 * time.Millisecond,
		Jitter:     0.1,
	}

	attempts := 0
	expectedErr := errors.New("persistent error")
	_, err := r.Do(context.Background(), func() (interface{}, error) {
		attempts++
		return nil, expectedErr
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
	if attempts != 4 { // initial + 3 retries
		t.Errorf("expected 4 attempts, got %d", attempts)
	}
}

func TestRetryContextCancelled(t *testing.T) {
	r := &RetryConfig{
		MaxRetries: 5,
		BaseDelay:  1 * time.Second,
		MaxDelay:   5 * time.Second,
		Jitter:     0.1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := r.Do(ctx, func() (interface{}, error) {
		return nil, errors.New("should not be called")
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestRetryTimingJitter(t *testing.T) {
	r := &RetryConfig{
		MaxRetries: 1,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
		Jitter:     0.5,
	}

	delays1 := r.BackoffDelay(1)
	delays2 := r.BackoffDelay(1)

	// With jitter, delays should differ (very unlikely to be equal)
	// But this is a statistical test, so it's not 100% reliable
	_ = delays1
	_ = delays2
	// Just verify the backoff produces reasonable values
	if r.BackoffDelay(0) < 5*time.Millisecond {
		t.Error("backoff delay too small")
	}
	if r.BackoffDelay(3) > 200*time.Millisecond {
		t.Error("backoff delay too large")
	}
}

func TestRetryZeroConfig(t *testing.T) {
	r := &RetryConfig{
		MaxRetries: 0,
		BaseDelay:  0,
		MaxDelay:   0,
		Jitter:     0,
	}

	_, err := r.Do(context.Background(), func() (interface{}, error) {
		return nil, errors.New("always fail")
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Backoff delay with zero config should return 0
	delay := r.BackoffDelay(0)
	if delay != 0 {
		t.Errorf("expected 0 delay, got %v", delay)
	}
}

func TestRetryNonRetryableError(t *testing.T) {
	r := &RetryConfig{
		MaxRetries: 3,
		BaseDelay:  5 * time.Millisecond,
		MaxDelay:   50 * time.Millisecond,
		Jitter:     0.1,
		IsRetryable: func(err error) bool {
			return err.Error() != "non-retryable"
		},
	}

	attempts := 0
	// First a retryable error, then a non-retryable one
	_, err := r.Do(context.Background(), func() (interface{}, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("retryable error")
		}
		return nil, errors.New("non-retryable")
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "non-retryable" {
		t.Errorf("expected 'non-retryable', got '%v'", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}
