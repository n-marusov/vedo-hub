package middleware

// Idempotency contract tests.
// BDD: [Condition]_[Action]_[ExpectedResult]
//
// These tests use MemIdempotencyStore (no Redis dependency).

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupIdempotencyTest(cfg *IdempotencyConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/test", Idempotency(cfg), func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"id": "resource-123", "status": "created"})
	})
	r.POST("/critical", Idempotency(cfg), func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"id": "resource-456"})
	})
	return r
}

// TestIdem_SameKeySamePayload verifies replay returns the same response.
func TestIdem_SameKeySamePayload_ReplayRequest_ReturnsSameResourceID(t *testing.T) {
	cfg := &IdempotencyConfig{
		Store: NewMemIdempotencyStore(),
	}
	router := setupIdempotencyTest(cfg)

	body := `{"name":"test"}`
	req1 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req1.Header.Set("Idempotency-Key", "key-001")
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("first request: expected 201, got %d", w1.Code)
	}

	// Replay with same key and body
	req2 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req2.Header.Set("Idempotency-Key", "key-001")
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("replay: expected 201, got %d", w2.Code)
	}

	var resp1, resp2 map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &resp1)
	json.Unmarshal(w2.Body.Bytes(), &resp2)

	if resp1["id"] != resp2["id"] {
		t.Fatalf("expected same resource ID on replay, got %v vs %v", resp1["id"], resp2["id"])
	}
}

// TestIdem_SameKeyDifferentPayload verifies conflict detection.
func TestIdem_SameKeyDifferentPayload_ReplayRequest_Returns409(t *testing.T) {
	cfg := &IdempotencyConfig{
		Store: NewMemIdempotencyStore(),
	}
	router := setupIdempotencyTest(cfg)

	// First request with payload A
	req1 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(`{"name":"A"}`))
	req1.Header.Set("Idempotency-Key", "key-002")
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatalf("first request: expected 201, got %d", w1.Code)
	}

	// Second request with same key but different payload
	req2 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(`{"name":"B"}`))
	req2.Header.Set("Idempotency-Key", "key-002")
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 for different payload with same key, got %d", w2.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp)
	errObj, ok := resp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "IDEMPOTENCY_KEY_REUSED_WITH_DIFFERENT_PAYLOAD" {
		t.Fatalf("expected conflict error code, got %v", errObj["code"])
	}
}

// TestIdem_MissingKeyCriticalWrite verifies critical writes require Idempotency-Key.
func TestIdem_MissingKey_CriticalWrite_Returns400(t *testing.T) {
	cfg := &IdempotencyConfig{
		Store:         NewMemIdempotencyStore(),
		CriticalPaths: []string{"/critical"},
	}
	router := setupIdempotencyTest(cfg)

	req := httptest.NewRequest(http.MethodPost, "/critical", bytes.NewBufferString(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing key on critical write, got %d", w.Code)
	}
}

// TestIdem_DifferentKeys verifies concurrent requests with different keys succeed.
func TestIdem_DifferentKeys_ConcurrentRequests_ReturnsDifferentIDs(t *testing.T) {
	cfg := &IdempotencyConfig{
		Store: NewMemIdempotencyStore(),
	}
	router := setupIdempotencyTest(cfg)

	body := `{"name":"test"}`
	req1 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req1.Header.Set("Idempotency-Key", "key-003")
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
	req2.Header.Set("Idempotency-Key", "key-004")
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w1.Code != http.StatusCreated || w2.Code != http.StatusCreated {
		t.Fatalf("both requests should succeed: got %d and %d", w1.Code, w2.Code)
	}
}

// TestIdem_InvalidKeyFormat verifies key format validation.
func TestIdem_InvalidKey_MissingKey_Returns400(t *testing.T) {
	cfg := &IdempotencyConfig{
		Store: NewMemIdempotencyStore(),
	}
	router := setupIdempotencyTest(cfg)

	tests := []struct {
		name string
		key  string
	}{
		{"empty key", ""},
		{"key with spaces", "key with spaces"},
		{"key with special chars", "key@#$"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(`{}`))
			req.Header.Set("Content-Type", "application/json")
			if tc.key != "" {
				req.Header.Set("Idempotency-Key", tc.key)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tc.key != "" {
				if w.Code != http.StatusBadRequest {
					t.Fatalf("expected 400 for invalid key %q, got %d", tc.key, w.Code)
				}
			}
		})
	}
}
