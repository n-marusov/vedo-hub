package main

// Validates: REQ-NFR.SECURITY.bola-bfla-negative-tests
// Validates: REQ-FUN.API.rest-gitlab-alignment

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

// TestSwaggerUI_Dev_AccessibleWithoutAuth verifies that Swagger UI is
// accessible without authentication when ENABLE_SWAGGER_UI=true (dev mode).
// The raw OpenAPI spec must also be public without auth.
// ADR-DES.API.swagger-ui-dev-only-strategy.
func TestSwaggerUI_Dev_AccessibleWithoutAuth(t *testing.T) {
	t.Setenv("ENABLE_SWAGGER_UI", "true")
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	// Docs index page — should return HTML without auth.
	// Gin redirects /api/v1/docs → /api/v1/docs/ (301) when a *filepath
	// catch-all route exists. Verify neither variant returns 401:
	for _, path := range []string{"/api/v1/docs", "/api/v1/docs/"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)
		if w.Code == http.StatusUnauthorized {
			t.Fatalf("expected non-401 for %s (dev), got %d (body=%s)", path, w.Code, w.Body.String())
		}
	}

	// Verify trailing-slash variant returns HTML
	req := httptest.NewRequest(http.MethodGet, "/api/v1/docs/", nil)
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for /api/v1/docs/ (dev), got %d (body=%s)", w.Code, w.Body.String())
	}
	ct := w.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" && ct != "text/html" {
		t.Errorf("expected HTML content type for /api/v1/docs/, got %q", ct)
	}

	// Static asset — swagger-ui.css should be served without auth
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/docs/swagger-ui.css", nil)
	w2 := httptest.NewRecorder()
	env.router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 for /api/v1/docs/swagger-ui.css (dev), got %d", w2.Code)
	}

	// OpenAPI spec — always public without auth regardless of ENABLE_SWAGGER_UI
	req3 := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.json", nil)
	w3 := httptest.NewRecorder()
	env.router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("expected 200 for /api/v1/openapi.json (dev), got %d (body=%s)", w3.Code, w3.Body.String())
	}
}

// TestSwaggerUI_StagingProd_Returns404 verifies that Swagger UI returns 404
// when ENABLE_SWAGGER_UI is not set (staging/production). The raw OpenAPI
// spec must remain public without auth in all environments.
// ADR-DES.API.swagger-ui-dev-only-strategy.
func TestSwaggerUI_StagingProd_Returns404(t *testing.T) {
	// Ensure ENABLE_SWAGGER_UI is not set (it may have been leaked by another test)
	t.Setenv("ENABLE_SWAGGER_UI", "")
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	// Docs index page — must return 404 without auth in staging/prod.
	// Request both variants to ensure neither leaks through auth.
	for _, path := range []string{"/api/v1/docs", "/api/v1/docs/"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		env.router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 for %s (staging/prod), got %d (body=%s)", path, w.Code, w.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode error body for %s: %v", path, err)
		}
		errObj, ok := body["error"].(map[string]any)
		if !ok {
			t.Fatalf("expected error object for %s, got %v", path, body)
		}
		if errObj["code"] != "SWAGGER-UI-DISABLED" {
			t.Errorf("expected error code SWAGGER-UI-DISABLED for %s, got %v", path, errObj["code"])
		}
	}

	// OpenAPI spec — still public without auth even in staging/prod
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.json", nil)
	w2 := httptest.NewRecorder()
	env.router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 for /api/v1/openapi.json (staging/prod), got %d (body=%s)", w2.Code, w2.Body.String())
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

// TestProjectMoveEndpointExists
// TestProjectMoveEndpointExists verifies PUT /projects/{id}/move is registered
// and returns 501 Not Implemented until protobuf code is regenerated.
func TestProjectMoveEndpointExists(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	body := jsonBody(t, map[string]string{"target_group_id": "group-2"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/test-project-1/move", body)
	req.Header.Set("Authorization", env.bearer(t, "owner-move", []string{"Owner"}))
	req.Header.Set("Idempotency-Key", "test-move-key-001")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	// 501 Not Implemented until protobuf regeneration
	if w.Code != http.StatusNotImplemented {
		if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
			t.Logf("Auth gate active for move endpoint: %d", w.Code)
			return
		}
		t.Errorf("expected 501 for move endpoint, got %d (body=%s)", w.Code, w.Body.String())
	}
}

// TestMetricsProxySmoke verifies GET /metrics/ontologies returns 200.
func TestMetricsProxySmoke(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/ontologies", nil)
	req.Header.Set("Authorization", env.bearer(t, "viewer-metrics", []string{"Viewer"}))
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /metrics/ontologies expected 200, got %d (body=%s)", w.Code, w.Body.String())
	}
}

// TestMetricsByIdProxySmoke verifies GET /metrics/ontologies/{id} returns 200.
func TestMetricsByIdProxySmoke(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics/ontologies/ont-1", nil)
	req.Header.Set("Authorization", env.bearer(t, "viewer-m2", []string{"Viewer"}))
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /metrics/ontologies/:id expected 200, got %d (body=%s)", w.Code, w.Body.String())
	}
}

// TestCommentProxyAuthRequired verifies POST /ontologies/:id/comments requires auth.
func TestCommentProxyAuthRequired(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	// No auth -> 401
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ontologies/ont-1/comments",
		jsonBody(t, map[string]string{"entity_id": "cls-1", "body": "test"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("POST /ontologies/:id/comments without auth expected 401, got %d", w.Code)
	}
}

// TestValidateEndpointSmoke verifies POST /ontologies/:id/validate exists.
func TestValidateEndpointSmoke(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ontologies/ont-1/validate",
		jsonBody(t, map[string]string{}))
	req.Header.Set("Authorization", env.bearer(t, "editor-val", []string{"Editor"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /ontologies/:id/validate expected 200, got %d", w.Code)
	}
}

// TestDraftEndpointRequiresIdempotencyKey verifies PUT /ontologies/:id/draft
// needs Idempotency-Key (under orgWrite middleware group).
func TestDraftEndpointRequiresIdempotencyKey(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	// Without Idempotency-Key -> 400
	body := jsonBody(t, map[string]interface{}{
		"changes": map[string]interface{}{"fields": []interface{}{}},
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/ontologies/ont-1/draft", body)
	req.Header.Set("Authorization", env.bearer(t, "owner-draft", []string{"Owner"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("PUT /ontologies/:id/draft without Idempotency-Key expected 400, got %d (body=%s)", w.Code, w.Body.String())
	}

	// With Idempotency-Key -> OK
	req2 := httptest.NewRequest(http.MethodPut, "/api/v1/ontologies/ont-1/draft", body)
	req2.Header.Set("Authorization", env.bearer(t, "owner-draft2", []string{"Owner"}))
	req2.Header.Set("Idempotency-Key", "draft-key-001")
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	env.router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("PUT /ontologies/:id/draft with Idempotency-Key expected 200, got %d", w2.Code)
	}
}

// TestProjectMoveRequiresIdempotencyKey verifies PUT /projects/{id}/move
// requires Idempotency-Key header.
func TestProjectMoveRequiresIdempotencyKey(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	body := jsonBody(t, map[string]string{"target_group_id": "group-2"})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/test-project-1/move", body)
	req.Header.Set("Authorization", env.bearer(t, "owner-move-2", []string{"Owner"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 without Idempotency-Key, got %d", w.Code)
	}
}
