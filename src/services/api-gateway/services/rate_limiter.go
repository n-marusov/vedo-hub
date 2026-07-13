package services

import (
	"strings"
	"sync"
	"time"
)

// Tier defines rate limit tiers.
type Tier string

const (
	TierAnonymous  Tier = "anonymous"
	TierFree       Tier = "free"
	TierPro        Tier = "pro"
	TierEnterprise Tier = "enterprise"
)

// TierLimits maps tiers to requests per minute.
var TierLimits = map[Tier]int{
	TierAnonymous:  10,
	TierFree:       100,
	TierPro:        500,
	TierEnterprise: 1000,
}

// RateLimiter implements a sliding window counter per user/API key.
type RateLimiter struct {
	mu       sync.Mutex
	windows  map[string]*slidingWindow
	maxLimit int
}

type slidingWindow struct {
	timestamps []time.Time
	limit      int
}

// NewRateLimiter creates a rate limiter with a max limit cap.
func NewRateLimiter(maxLimit int) *RateLimiter {
	if maxLimit <= 0 {
		maxLimit = 1000
	}
	return &RateLimiter{
		windows:  make(map[string]*slidingWindow),
		maxLimit: maxLimit,
	}
}

// Allow checks if a request from the given key is within the rate limit for the tier.
// Returns true if allowed, false if rate limited, and the retry-after duration.
func (rl *RateLimiter) Allow(key string, tier Tier) (bool, time.Duration) {
	limit, ok := TierLimits[tier]
	if !ok {
		limit = TierLimits[TierFree]
	}
	if limit > rl.maxLimit {
		limit = rl.maxLimit
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	window, exists := rl.windows[key]
	if !exists {
		rl.windows[key] = &slidingWindow{
			timestamps: []time.Time{now},
			limit:      limit,
		}
		return true, 0
	}

	// Prune timestamps older than 1 minute
	cutoff := now.Add(-1 * time.Minute)
	pruned := make([]time.Time, 0, len(window.timestamps))
	for _, t := range window.timestamps {
		if t.After(cutoff) {
			pruned = append(pruned, t)
		}
	}
	window.timestamps = pruned

	if len(window.timestamps) >= limit {
		// Calculate when the oldest timestamp expires
		retryAfter := time.Until(window.timestamps[0].Add(1 * time.Minute))
		if retryAfter < 0 {
			retryAfter = time.Second
		}
		return false, retryAfter
	}

	window.timestamps = append(window.timestamps, now)
	return true, 0
}

// ResolveTier resolves a tier from user roles.
func ResolveTier(roles []string) Tier {
	for _, r := range roles {
		switch strings.ToLower(r) {
		case "enterprise":
			return TierEnterprise
		case "pro":
			return TierPro
		case "free":
			return TierFree
		}
	}
	return TierAnonymous
}
