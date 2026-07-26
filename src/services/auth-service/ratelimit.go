package main

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type slidingWindowCounter struct {
	mu      sync.Mutex
	windows map[string][]time.Time
	limits  map[ClientTier]int
}

var globalRateLimiter = newSlidingWindowCounter()

func newSlidingWindowCounter() *slidingWindowCounter {
	return &slidingWindowCounter{
		windows: make(map[string][]time.Time),
		limits: map[ClientTier]int{
			TierAnonymous:      10,
			TierFree:           100,
			TierPro:            500,
			TierEnterprise:     1000,
			TierServiceAccount: 2000,
		},
	}
}

func (r *slidingWindowCounter) allow(key string, tier ClientTier) (bool, int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	limit, ok := r.limits[tier]
	if !ok {
		limit = r.limits[TierAnonymous]
	}

	now := time.Now()
	windowStart := now.Add(-1 * time.Minute)

	entries := r.windows[key]
	var valid []time.Time
	for _, t := range entries {
		if t.After(windowStart) {
			valid = append(valid, t)
		}
	}

	remaining := limit - len(valid)
	if remaining < 0 {
		remaining = 0
	}

	if len(valid) >= limit {
		r.windows[key] = valid
		return false, remaining
	}

	valid = append(valid, now)
	r.windows[key] = valid
	return true, remaining
}

func resolveClientTier(r *http.Request) ClientTier {
	tier := r.Header.Get("X-Client-Tier")
	switch ClientTier(tier) {
	case TierFree, TierPro, TierEnterprise, TierServiceAccount:
		return ClientTier(tier)
	default:
		return TierAnonymous
	}
}

func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tier := resolveClientTier(r)
		clientKey := r.RemoteAddr + ":" + string(tier)

		allowed, remaining := globalRateLimiter.allow(clientKey, tier)

		w.Header().Set("X-RateLimit-Remaining", itoa(remaining))
		w.Header().Set("X-RateLimit-Tier", string(tier))

		if !allowed {
			w.Header().Set("Retry-After", "60")
			writeJSON(w, http.StatusTooManyRequests, ErrorResponse{
				Error: APIError{
					Code:    "RATE_LIMITED",
					Message: "Rate limit exceeded for tier " + string(tier),
				},
			})
			log.Printf("Rate limited client=%s tier=%s", clientKey, tier)
			return
		}

		next(w, r)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
