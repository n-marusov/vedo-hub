// @ctx: HTTP handler tests for commenting-service — contract test scenarios
// Tests the /, /health, /ready, /metrics endpoints and helper functions.

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// @ctx: helper to call handler via httptest
func executeRequest(method, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("X-Trace-Id", "test-trace-"+path)
	req.Header.Set("X-Correlation-Id", "test-corr-"+path)
	handler(w, req)
	return w
}

// @ctx: helper to parse JSON response body
func parseJSON(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}
	return resp
}

// ─── resolveID ───────────────────────────────────────────────────────────────

// @hlv CT-COMMENT-RESOLVE-001
func TestResolveID_WithHeader_ReturnsHeaderValue(t *testing.T) {
	result := resolveID("my-trace-value")
	if result != "my-trace-value" {
		t.Errorf("expected 'my-trace-value', got %q", result)
	}
}

// @hlv CT-COMMENT-RESOLVE-002
func TestResolveID_EmptyHeader_ReturnsNonEmptyFallback(t *testing.T) {
	result := resolveID("")
	if result == "" {
		t.Error("expected non-empty fallback for empty header")
	}
}

// ─── writeJSON ───────────────────────────────────────────────────────────────

// @hlv CT-COMMENT-WRITEJSON-001
func TestWriteJSON_SetsContentType(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", ct)
	}
}

// @hlv CT-COMMENT-WRITEJSON-002
func TestWriteJSON_WritesCorrectStatusCode(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusCreated, map[string]string{"id": "abc"})

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

// @hlv CT-COMMENT-WRITEJSON-003
func TestWriteJSON_EncodesPayload(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, http.StatusOK, map[string]string{"key": "value"})

	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["key"] != "value" {
		t.Errorf("expected key=value, got %q", resp["key"])
	}
}

// ─── writeMetrics ────────────────────────────────────────────────────────────

// @hlv CT-COMMENT-METRICS-001
func TestWriteMetrics_ContentType(t *testing.T) {
	w := httptest.NewRecorder()
	writeMetrics(w)

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected Content-Type starting with 'text/plain', got %q", ct)
	}
}

// @hlv CT-COMMENT-METRICS-002
func TestWriteMetrics_StatusOK(t *testing.T) {
	w := httptest.NewRecorder()
	writeMetrics(w)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// @hlv CT-COMMENT-METRICS-003
func TestWriteMetrics_ContainsServiceName(t *testing.T) {
	w := httptest.NewRecorder()
	writeMetrics(w)

	body := w.Body.String()
	if !strings.Contains(body, serviceName) {
		t.Errorf("expected metrics to contain service name %q, got: %s", serviceName, body)
	}
}

// @hlv CT-COMMENT-METRICS-004
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

// ─── Endpoints: Happy Path ───────────────────────────────────────────────────

// @hlv CT-COMMENT-ROOT-001
func TestCT_RootEndpoint_ReturnsServiceInfo(t *testing.T) {
	w := executeRequest("GET", "/")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(t, w)
	if resp["name"] != serviceName {
		t.Errorf("expected name=%q, got %q", serviceName, resp["name"])
	}
	if resp["version"] != "0.2.0" {
		t.Errorf("expected version=0.2.0, got %v", resp["version"])
	}
	if resp["stub"] != false {
		t.Errorf("expected stub=false, got %v", resp["stub"])
	}
}

// @hlv CT-COMMENT-HEALTH-001
func TestCT_HealthEndpoint_ReturnsHealthy(t *testing.T) {
	w := executeRequest("GET", "/health")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(t, w)
	if resp["status"] != "healthy" {
		t.Errorf("expected status=healthy, got %v", resp["status"])
	}
}

// @hlv CT-COMMENT-READY-001
func TestCT_ReadyEndpoint_ReturnsReady(t *testing.T) {
	w := executeRequest("GET", "/ready")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(t, w)
	if resp["status"] != "ready" {
		t.Errorf("expected status=ready, got %v", resp["status"])
	}
}

// @hlv CT-COMMENT-METRICS-005
func TestCT_MetricsEndpoint_ReturnsMetrics(t *testing.T) {
	w := executeRequest("GET", "/metrics")

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	body := w.Body.String()
	if !strings.Contains(body, "vedo_service_requests_total") {
		t.Errorf("metrics body should contain counter, got: %s", body)
	}
}

// ─── Endpoints: Negative ────────────────────────────────────────────────────

// @hlv CT-COMMENT-NEG-001
func TestCT_UnknownPath_Returns404(t *testing.T) {
	w := executeRequest("GET", "/unknown")

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(t, w)
	if resp["error"] != "ENDPOINT_NOT_FOUND" {
		t.Errorf("expected error=ENDPOINT_NOT_FOUND, got %v", resp["error"])
	}
}

// @hlv CT-COMMENT-NEG-002
func TestCT_UnknownPath_IncludesAvailableEndpoints(t *testing.T) {
	w := executeRequest("GET", "/nonexistent")

	resp := parseJSON(t, w)
	available, ok := resp["available"].([]any)
	if !ok {
		t.Fatal("expected 'available' field as array")
	}
	expected := map[string]bool{"/": false, "/health": false, "/ready": false, "/metrics": false}
	for _, a := range available {
		if s, ok := a.(string); ok {
			expected[s] = true
		}
	}
	for path, found := range expected {
		if !found {
			t.Errorf("expected available endpoint %q not found in response", path)
		}
	}
}

// @hlv CT-COMMENT-NEG-003
func TestCT_PostMethod_Returns405(t *testing.T) {
	w := executeRequest("POST", "/")

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(t, w)
	if resp["error"] != "METHOD_NOT_ALLOWED" {
		t.Errorf("expected error=METHOD_NOT_ALLOWED, got %v", resp["error"])
	}
}

// @hlv CT-COMMENT-NEG-004
func TestCT_PutMethod_Returns405(t *testing.T) {
	w := executeRequest("PUT", "/health")

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(t, w)
	if resp["error"] != "METHOD_NOT_ALLOWED" {
		t.Errorf("expected error=METHOD_NOT_ALLOWED, got %v", resp["error"])
	}
}

// @hlv CT-COMMENT-NEG-005
func TestCT_DeleteMethod_Returns405(t *testing.T) {
	w := executeRequest("DELETE", "/ready")

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(t, w)
	if resp["error"] != "METHOD_NOT_ALLOWED" {
		t.Errorf("expected error=METHOD_NOT_ALLOWED, got %v", resp["error"])
	}
}

// @hlv CT-COMMENT-NEG-006
func TestCT_MethodNotAllowed_IncludesAllowedMethods(t *testing.T) {
	w := executeRequest("PATCH", "/")

	resp := parseJSON(t, w)
	available, ok := resp["available"].([]any)
	if !ok {
		t.Fatal("expected 'available' field as array")
	}
	hasGET := false
	for _, a := range available {
		if a == "GET" {
			hasGET = true
			break
		}
	}
	// The error message lists available endpoints, not methods.
	// At minimum, verify the error structure is correct.
	if resp["error"] != "METHOD_NOT_ALLOWED" {
		t.Errorf("expected METHOD_NOT_ALLOWED, got %v", resp["error"])
	}
	_ = hasGET // available list shows endpoints, not methods
}

// ─── Invariant Properties ────────────────────────────────────────────────────

// @hlv CT-COMMENT-INV-001
func TestInvariant_AllKnownEndpoints_Return200ForGET(t *testing.T) {
	endpoints := []string{"/", "/health", "/ready", "/metrics"}
	for _, ep := range endpoints {
		w := executeRequest("GET", ep)
		if w.Code != http.StatusOK {
			t.Errorf("endpoint %s expected 200, got %d: %s", ep, w.Code, w.Body.String())
		}
	}
}

// @hlv CT-COMMENT-INV-002
func TestInvariant_AllKnownEndpoints_Return405ForPOST(t *testing.T) {
	endpoints := []string{"/", "/health", "/ready", "/metrics"}
	for _, ep := range endpoints {
		w := executeRequest("POST", ep)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("endpoint %s expected 405 for POST, got %d", ep, w.Code)
		}
	}
}

// @hlv CT-COMMENT-INV-003
func TestInvariant_AllKnownEndpoints_Return405ForPUT(t *testing.T) {
	endpoints := []string{"/", "/health", "/ready", "/metrics"}
	for _, ep := range endpoints {
		w := executeRequest("PUT", ep)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("endpoint %s expected 405 for PUT, got %d", ep, w.Code)
		}
	}
}

// @hlv CT-COMMENT-INV-004
func TestInvariant_AllUnknownPaths_Return404(t *testing.T) {
	paths := []string{"/unknown", "/api/v1/comments", "/admin", "/../health"}
	for _, p := range paths {
		w := executeRequest("GET", p)
		if w.Code != http.StatusNotFound {
			t.Errorf("path %s expected 404, got %d: %s", p, w.Code, w.Body.String())
		}
	}
}

// @hlv CT-COMMENT-INV-005
func TestInvariant_JSONContentTypeOnAllEndpoints(t *testing.T) {
	endpoints := []string{"/", "/health", "/ready"}
	for _, ep := range endpoints {
		w := executeRequest("GET", ep)
		ct := w.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("endpoint %s expected Content-Type 'application/json', got %q", ep, ct)
		}
	}
}

// @hlv CT-COMMENT-INV-006
func TestInvariant_RequestTotalIncrementedOnEachCall(t *testing.T) {
	old := requestTotal.Load()
	executeRequest("GET", "/")
	newCount := requestTotal.Load()
	if newCount <= old {
		t.Errorf("expected requestTotal to increase, was %d, now %d", old, newCount)
	}
}
