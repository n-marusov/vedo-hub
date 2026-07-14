// Package main — HTTP handler tests for commenting-service CRUD endpoints.
package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ─── helpers ───────────────────────────────────────────────────────────────────

// setupTestMux creates a ServeMux with comment routes registered but no database store.
// All CRUD operations will fail gracefully with nil-store handling.
func setupTestMux() *http.ServeMux {
	mux := http.NewServeMux()
	commentHandlers := &commentHandlers{store: nil, bus: nil}

	// Info endpoints
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"name":    serviceName,
			"version": "0.3.0",
			"stub":    false,
		})
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy", "service": serviceName})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready", "service": serviceName})
	})
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		writeMetrics(w)
	})

	registerCommentRoutes(mux, commentHandlers)
	return mux
}

func executeRequest(mux *http.ServeMux, method, path string, body []byte) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Trace-Id", "test-trace")
	req.Header.Set("X-Correlation-Id", "test-corr")
	req.Header.Set("X-User-Id", "test-user")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	mux.ServeHTTP(w, req)
	return w
}

func parseJSONResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON response: %v (body: %s)", err, w.Body.String())
	}
	return resp
}

// ─── writeJSON / writeMetrics ──────────────────────────────────────────────────

func TestWriteJSON_SetsContentType(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}
}

func TestWriteJSON_WritesCorrectStatusCode(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusCreated, map[string]string{"id": "abc"})

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestWriteMetrics_ContentType(t *testing.T) {
	w := httptest.NewRecorder()
	writeMetrics(w)

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected Content-Type starting with 'text/plain', got %q", ct)
	}
}

func TestWriteMetrics_PrometheusFormat(t *testing.T) {
	w := httptest.NewRecorder()
	writeMetrics(w)

	body := w.Body.String()
	if !strings.Contains(body, "# HELP ") {
		t.Error("expected HELP line in Prometheus metrics")
	}
	if !strings.Contains(body, "# TYPE ") {
		t.Error("expected TYPE line in Prometheus metrics")
	}
	if !strings.Contains(body, "vedo_service_requests_total") {
		t.Error("expected counter metric 'vedo_service_requests_total'")
	}
}

// ─── Info Endpoints ───────────────────────────────────────────────────────────

func TestInfoEndpoint_ReturnsServiceInfo(t *testing.T) {
	mux := setupTestMux()
	w := executeRequest(mux, "GET", "/", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSONResponse(t, w)
	if resp["name"] != serviceName {
		t.Errorf("expected name=%q, got %q", serviceName, resp["name"])
	}
	if resp["version"] != "0.3.0" {
		t.Errorf("expected version=0.3.0, got %v", resp["version"])
	}
	if resp["stub"] != false {
		t.Errorf("expected stub=false, got %v", resp["stub"])
	}
}

func TestHealthEndpoint_Returns200(t *testing.T) {
	mux := setupTestMux()
	w := executeRequest(mux, "GET", "/health", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSONResponse(t, w)
	if resp["status"] != "healthy" {
		t.Errorf("expected status=healthy, got %v", resp["status"])
	}
}

func TestReadyEndpoint_Returns200(t *testing.T) {
	mux := setupTestMux()
	w := executeRequest(mux, "GET", "/ready", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSONResponse(t, w)
	if resp["status"] != "ready" {
		t.Errorf("expected status=ready, got %v", resp["status"])
	}
}

func TestMetricsEndpoint_ReturnsPrometheus(t *testing.T) {
	mux := setupTestMux()
	w := executeRequest(mux, "GET", "/metrics", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, "vedo_service_requests_total") {
		t.Errorf("expected counter in metrics body, got: %s", body)
	}
}

// ─── Comment CRUD: create ──────────────────────────────────────────────────────

func TestCreateComment_NoStore_Returns503(t *testing.T) {
	mux := setupTestMux()
	body := `{"entity_id":"ent-1","entity_type":"class","body":"test comment"}`
	w := executeRequest(mux, "POST", "/api/v1/ontologies/ont-1/comments", []byte(body))

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 with no store, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSONResponse(t, w)
	if resp["error"] != "STORE_UNAVAILABLE" {
		t.Errorf("expected STORE_UNAVAILABLE, got %v", resp["error"])
	}
}

func TestCreateComment_MissingEntityAndBody_Returns400(t *testing.T) {
	mux := setupTestMux()
	body := `{"entity_id":"","entity_type":"class","body":""}`
	w := executeRequest(mux, "POST", "/api/v1/ontologies/ont-1/comments", []byte(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSONResponse(t, w)
	if resp["error"] != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %v", resp["error"])
	}
}

func TestCreateComment_InvalidEntityType_Returns400(t *testing.T) {
	mux := setupTestMux()
	body := `{"entity_id":"ent-1","entity_type":"invalid","body":"test comment"}`
	w := executeRequest(mux, "POST", "/api/v1/ontologies/ont-1/comments", []byte(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSONResponse(t, w)
	if resp["error"] != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %v", resp["error"])
	}
}

func TestCreateComment_InvalidJSON_Returns400(t *testing.T) {
	mux := setupTestMux()
	w := executeRequest(mux, "POST", "/api/v1/ontologies/ont-1/comments", []byte(`not json`))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSONResponse(t, w)
	if resp["error"] != "INVALID_REQUEST_BODY" {
		t.Errorf("expected INVALID_REQUEST_BODY, got %v", resp["error"])
	}
}

// ─── Comment CRUD: get ────────────────────────────────────────────────────────

func TestGetComment_NoStore_Returns503(t *testing.T) {
	mux := setupTestMux()
	w := executeRequest(mux, "GET", "/api/v1/comments/cmt-1", nil)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 with no store, got %d: %s", w.Code, w.Body.String())
	}
}

// ─── Comment CRUD: update ─────────────────────────────────────────────────────

func TestUpdateComment_NoStore_Returns503(t *testing.T) {
	mux := setupTestMux()
	body := `{"body":"updated body"}`
	w := executeRequest(mux, "PUT", "/api/v1/comments/cmt-1", []byte(body))

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 with no store, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUpdateComment_EmptyBody_Returns400(t *testing.T) {
	mux := setupTestMux()
	body := `{"body":""}`
	w := executeRequest(mux, "PUT", "/api/v1/comments/cmt-1", []byte(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSONResponse(t, w)
	if resp["error"] != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %v", resp["error"])
	}
}

// ─── Comment CRUD: delete ─────────────────────────────────────────────────────

func TestDeleteComment_NoStore_Returns503(t *testing.T) {
	mux := setupTestMux()
	w := executeRequest(mux, "DELETE", "/api/v1/comments/cmt-1", nil)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 with no store, got %d: %s", w.Code, w.Body.String())
	}
}

// ─── Entity-scoped listing ────────────────────────────────────────────────────

func TestListComments_MissingEntityID_Returns400(t *testing.T) {
	mux := setupTestMux()
	w := executeRequest(mux, "GET", "/api/v1/ontologies/ont-1/comments", nil)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSONResponse(t, w)
	if resp["error"] != "MISSING_ENTITY_ID" {
		t.Errorf("expected MISSING_ENTITY_ID, got %v", resp["error"])
	}
}

func TestListComments_NoStore_Returns503(t *testing.T) {
	mux := setupTestMux()
	w := executeRequest(mux, "GET", "/api/v1/ontologies/ont-1/comments?entity_id=ent-1", nil)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 with no store, got %d: %s", w.Code, w.Body.String())
	}
}

// ─── Comment Feed ─────────────────────────────────────────────────────────────

func TestCommentFeed_InvalidSince_Returns400(t *testing.T) {
	mux := setupTestMux()
	w := executeRequest(mux, "GET", "/api/v1/ontologies/ont-1/comment-feed?since=not-a-date", nil)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSONResponse(t, w)
	if resp["error"] != "INVALID_SINCE_PARAM" {
		t.Errorf("expected INVALID_SINCE_PARAM, got %v", resp["error"])
	}
}

func TestCommentFeed_NoStore_Returns503(t *testing.T) {
	mux := setupTestMux()
	w := executeRequest(mux, "GET", "/api/v1/ontologies/ont-1/comment-feed?since=2024-01-01T00:00:00Z", nil)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 with no store, got %d: %s", w.Code, w.Body.String())
	}
}

// ─── parseIntParam ────────────────────────────────────────────────────────────

func TestParseIntParam_Empty_ReturnsDefault(t *testing.T) {
	result := parseIntParam("", 42)
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestParseIntParam_Valid_ReturnsValue(t *testing.T) {
	result := parseIntParam("7", 42)
	if result != 7 {
		t.Errorf("expected 7, got %d", result)
	}
}

func TestParseIntParam_Invalid_ReturnsDefault(t *testing.T) {
	result := parseIntParam("abc", 10)
	if result != 10 {
		t.Errorf("expected 10, got %d", result)
	}
}

func TestParseIntParam_Negative_ReturnsDefault(t *testing.T) {
	result := parseIntParam("-5", 10)
	if result != 10 {
		t.Errorf("expected 10, got %d", result)
	}
}

// ─── Negative: unknown routes ─────────────────────────────────────────────────

func TestUnknownRoute_ReturnsServiceInfo(t *testing.T) {
	// Go 1.22 ServeMux: `GET /` pattern acts as catch-all for unmatched paths.
	// This is expected backward-compatible behavior.
	mux := setupTestMux()
	w := executeRequest(mux, "GET", "/unknown/path", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (root catch-all), got %d", w.Code)
	}

	resp := parseJSONResponse(t, w)
	if resp["name"] != serviceName {
		t.Errorf("expected service name from root handler, got %v", resp["name"])
	}
}
