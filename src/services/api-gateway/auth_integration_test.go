package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// TestAuth_MissingAuthorizationRejected exercises the SEC-AUTHZ-GATES-001
// contract: a request without an Authorization header must be rejected with
// 401 UNAUTHENTICATED for any protected path.
func TestAuth_MissingAuthorizationRejected(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ontologies", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing Authorization, got %d", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	errObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object, got %v", body)
	}
	if errObj["code"] != "UNAUTHENTICATED" {
		t.Errorf("expected error code UNAUTHENTICATED, got %v", errObj["code"])
	}
}

// TestAuth_ExpiredTokenRejected rejects an expired JWT with TOKEN_EXPIRED.
func TestAuth_ExpiredTokenRejected(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	expired := env.signTestJWT(t, "user-1", []string{"Viewer"}, time.Now().Add(-1*time.Hour))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ontologies", nil)
	req.Header.Set("Authorization", "Bearer "+expired)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d", w.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	errObj, _ := body["error"].(map[string]any)
	if errObj["code"] != "TOKEN_EXPIRED" {
		t.Errorf("expected error code TOKEN_EXPIRED, got %v", errObj["code"])
	}
}

// TestAuth_ValidViewerReadAllowed verifies that a Viewer-role token can read
// (BFLA: GET requires level 0, Viewer weight = 0).
func TestAuth_ValidViewerReadAllowed(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ontologies", nil)
	req.Header.Set("Authorization", env.bearer(t, "viewer-1", []string{"Viewer"}))
	req.Header.Set("X-Trace-Id", "trace-auth-1")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for Viewer GET, got %d (body=%s)", w.Code, w.Body.String())
	}
}

// TestAuth_InsufficientRoleForWrite verifies BFLA gating: a Viewer cannot
// POST (level 1) while an Editor (weight 1) can.
func TestAuth_InsufficientRoleForWrite(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	// Viewer attempt → 403 FORBIDDEN_INSUFFICIENT_ROLE
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ontologies",
		jsonBody(t, gin.H{"name": "o"}))
	req.Header.Set("Authorization", env.bearer(t, "viewer-2", []string{"Viewer"}))
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for Viewer POST, got %d (body=%s)", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	errObj, _ := body["error"].(map[string]any)
	if errObj["code"] != "FORBIDDEN_INSUFFICIENT_ROLE" {
		t.Errorf("expected FORBIDDEN_INSUFFICIENT_ROLE, got %v", errObj["code"])
	}

	// Editor attempt → passes auth and reaches the mock upstream (200).
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/ontologies",
		jsonBody(t, gin.H{"name": "o"}))
	req2.Header.Set("Authorization", env.bearer(t, "editor-1", []string{"Editor"}))
	w2 := httptest.NewRecorder()
	env.router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 for Editor POST, got %d (body=%s)", w2.Code, w2.Body.String())
	}
}

// TestAuth_AdminEndpointRequiresOwner verifies the isAdminEndpoint path
// (/api/v1/admin/users) requires an Owner/SecurityLead token (level 3).
func TestAuth_AdminEndpointRequiresOwner(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	// Editor (weight 1) → 403 admin-only.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req.Header.Set("Authorization", env.bearer(t, "editor-2", []string{"Editor"}))
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for Editor on admin endpoint, got %d (body=%s)", w.Code, w.Body.String())
	}

	// Owner (weight 3) → admin endpoint passes auth (no upstream expected to exist).
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	req2.Header.Set("Authorization", env.bearer(t, "owner-1", []string{"Owner"}))
	w2 := httptest.NewRecorder()
	env.router.ServeHTTP(w2, req2)
	// The route is not explicitly registered, so it falls through to the
	// gin 404 handler. The important assertion is that auth itself did not
	// reject it with 403.
	if w2.Code == http.StatusForbidden {
		t.Fatalf("Owner should pass admin auth gate, got 403 (body=%s)", w2.Body.String())
	}
}

// jsonBody builds a small JSON reader for test request bodies. We use
// bytes.NewReader so net/http derives a Content-Length header automatically,
// which lets httputil.ReverseProxy stream the body without surprising the
// keep-alive transport when the upstream handler ignores the body.
func jsonBody(t *testing.T, payload any) *bytes.Reader {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	return bytes.NewReader(data)
}
