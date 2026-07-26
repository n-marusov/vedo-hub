package proxy

import (
	"sync"
	"sync/atomic"
	"time"

	"log/slog"
)

// CircuitBreaker states
const (
	StateClosed   = "closed"
	StateOpen     = "open"
	StateHalfOpen = "half-open"
)

// CircuitBreaker prevents cascading failures by stopping requests to an
// unhealthy upstream. After 5 consecutive 5xx responses it opens the circuit
// for 30 seconds, after which a probe request is allowed through (half-open).
type CircuitBreaker struct {
	mu              sync.Mutex
	state           string
	failureCount    int64
	lastFailureTime time.Time
	lastStateChange time.Time

	failureThreshold int64
	cooldownPeriod   time.Duration
	resetTimeout     time.Duration

	// For atomic fast-path reads
	open int32 // 1 = open, 0 = closed/half-open
}

// NewCircuitBreaker creates a circuit breaker with the given thresholds.
//   - failureThreshold: consecutive failures before opening (default: 5)
//   - cooldownPeriod:  time to wait before transitioning to half-open (default: 30s)
//   - resetTimeout:    time after which the failure count resets (default: 60s)
func NewCircuitBreaker(failureThreshold int64, cooldownPeriod, resetTimeout time.Duration) *CircuitBreaker {
	if failureThreshold <= 0 {
		failureThreshold = 5
	}
	if cooldownPeriod <= 0 {
		cooldownPeriod = 30 * time.Second
	}
	if resetTimeout <= 0 {
		resetTimeout = 60 * time.Second
	}
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		cooldownPeriod:   cooldownPeriod,
		resetTimeout:     resetTimeout,
		lastStateChange:  time.Now(),
	}
}

// IsOpen returns true if the circuit is open (requests should be rejected).
func (cb *CircuitBreaker) IsOpen() bool {
	if atomic.LoadInt32(&cb.open) == 0 {
		return false
	}
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateOpen && time.Since(cb.lastStateChange) >= cb.cooldownPeriod {
		// Transition to half-open
		cb.state = StateHalfOpen
		cb.lastStateChange = time.Now()
		atomic.StoreInt32(&cb.open, 0)
		slog.Info("circuit_breaker.half_open", "state", cb.state)
		return false
	}
	return cb.state == StateOpen
}

// RecordFailure records a failure and potentially opens the circuit.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	// Reset counter if the reset timeout has elapsed since last failure
	if now.After(cb.lastFailureTime.Add(cb.resetTimeout)) {
		atomic.StoreInt64(&cb.failureCount, 0)
	}

	cb.lastFailureTime = now
	count := atomic.AddInt64(&cb.failureCount, 1)

	slog.Debug("circuit_breaker.failure",
		"failure_count", count,
		"threshold", cb.failureThreshold,
		"state", cb.state,
	)

	if count >= cb.failureThreshold {
		cb.state = StateOpen
		cb.lastStateChange = now
		atomic.StoreInt32(&cb.open, 1)
		atomic.StoreInt64(&cb.failureCount, 0)
		slog.Warn("circuit_breaker.opened",
			"failure_threshold", cb.failureThreshold,
			"cooldown", cb.cooldownPeriod.String(),
		)
	}
}

// RecordSuccess records a success and closes the circuit if half-open.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.lastStateChange = time.Now()
		atomic.StoreInt32(&cb.open, 0)
		atomic.StoreInt64(&cb.failureCount, 0)
		slog.Info("circuit_breaker.closed", "state", cb.state)
	} else if cb.state == StateClosed && time.Since(cb.lastFailureTime) > cb.resetTimeout {
		// Reset failure count after a period of no failures
		atomic.StoreInt64(&cb.failureCount, 0)
	}
}

// State returns the current state.
func (cb *CircuitBreaker) State() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
