package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestRateLimit_AnonymousTierReturns429 verifies the sliding-window rate limit
// for the anonymous tier: the gateway returns 429 once the configured window
// (TierAnonymous = 10/min in services.TierLimits) is exceeded.
//
// Because the auth middleware resolves roles from the JWT, anonymous callers
// are simulated by issuing requests with no `X-User-Id` header. They all share
// the same source IP (`httptest` reverse) so the same rate-limit key is used.
func TestRateLimit_AnonymousTierReturns429(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	// Anonymous callers have no roles → ResolveTier returns anonymous (10/min).
	const burst = 12 // 11th should trip the rear-window cut-off

	env.ontologyServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(gin.H{"ok": true})
	})

	jwt := env.bearer(t, "anon-user", nil) // empty roles → anonymous tier

	var statuses []int
	for i := 0; i < burst; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sparql",
			jsonBody(t, gin.H{"query": "SELECT ?s WHERE { ?s ?p ?o }"}))
		req.Header.Set("Authorization", jwt)
		req.Header.Set("Content-Type", "application/json")
		// Force a distinct user_id header per call so rateLimitKey = source IP,
		// which is shared across all calls (httptest.NewRecorder == same addr).
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)
		statuses = append(statuses, w.Code)
	}

	saw200 := 0
	saw429 := false
	for _, s := range statuses {
		switch s {
		case http.StatusOK:
			saw200++
		case http.StatusTooManyRequests:
			saw429 = true
		}
	}

	if !saw429 {
		t.Fatalf("expected at least one 429 after burst of %d anonymous calls, got statuses=%v",
			burst, statuses)
	}
	if saw200 == 0 {
		t.Fatalf("expected some 200 responses before the rate limit fired, got statuses=%v",
			statuses)
	}
}

// TestRateLimit_RetryAfterHeaderPresent ensures the 429 response advertises a
// `Retry-After` header so clients know when to back off.
func TestRateLimit_RetryAfterHeaderPresent(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	env.ontologyServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(gin.H{"ok": true})
	})

	jwt := env.bearer(t, "anon-user-2", nil)
	const burst = 12

	var retryAfter string
	for i := 0; i < burst; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sparql",
			jsonBody(t, gin.H{"query": "SELECT ?s WHERE { ?s ?p ?o }"}))
		req.Header.Set("Authorization", jwt)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)
		if w.Code == http.StatusTooManyRequests {
			retryAfter = w.Header().Get("Retry-After")
			break
		}
	}
	if retryAfter == "" {
		t.Fatalf("expected Retry-After header on 429 response")
	}
}
