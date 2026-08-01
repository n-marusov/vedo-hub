package main

// Validates: REQ-FUN.API.rest-gitlab-alignment

// RED-phase tests (TDD): planned stub endpoints MUST return
// `501 Not Implemented` with the `x-vedo-status: planned` header and a JSON
// body `{"message":"Not implemented — planned for M10/M11","x_vedo_status":"planned"}`
// per the GitLab-aligned REST contract (ADR-DES.API.rest-gitlab-alignment,
// REQ-FUN.API.rest-gitlab-alignment).
//
// These tests FAIL against the current code (the routes return 404 — no
// stubs are registered) and pass only after the stubs are added (Task 19 of
// the plan). They are unit-level: the router is built with newTestEnv using
// httptest upstreams, so no external infrastructure is required.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// plannedStubPaths are the contract stubs planned for M10/M11.
var plannedStubPaths = []struct {
	method, path string
}{
	{http.MethodGet, "/api/v1/projects/123/releases"},
	{http.MethodGet, "/api/v1/projects/123/merge_requests"},
	{http.MethodPost, "/api/v1/projects/123/merge_requests"},
	{http.MethodGet, "/api/v1/projects/123/protected_branches"},
	{http.MethodPost, "/api/v1/projects/123/protected_branches"},
}

// issuePlannedStub performs a request to a planned stub path and returns the
// response recorder.
func issuePlannedStub(t *testing.T, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", env.bearer(t, "user-1", []string{"Owner"}))
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	return w
}

// TestPlannedStub_GetReleases_ShouldReturn501 verifies
// GET /api/v1/projects/{pid}/releases returns 501.
func TestPlannedStub_GetReleases_ShouldReturn501(t *testing.T) {
	w := issuePlannedStub(t, http.MethodGet, "/api/v1/projects/123/releases")
	if w.Code != http.StatusNotImplemented {
		t.Errorf("GET releases status = %d, want %d", w.Code, http.StatusNotImplemented)
	}
}

// TestPlannedStub_GetReleases_ShouldReturnXStatusPlanned verifies the
// x-vedo-status: planned header on releases.
func TestPlannedStub_GetReleases_ShouldReturnXStatusPlanned(t *testing.T) {
	w := issuePlannedStub(t, http.MethodGet, "/api/v1/projects/123/releases")
	if got := w.Header().Get("x-vedo-status"); got != "planned" {
		t.Errorf("GET releases x-vedo-status = %q, want %q", got, "planned")
	}
}

// TestPlannedStub_GetMergeRequests_ShouldReturn501 verifies
// GET /api/v1/projects/{pid}/merge_requests returns 501.
func TestPlannedStub_GetMergeRequests_ShouldReturn501(t *testing.T) {
	w := issuePlannedStub(t, http.MethodGet, "/api/v1/projects/123/merge_requests")
	if w.Code != http.StatusNotImplemented {
		t.Errorf("GET merge_requests status = %d, want %d", w.Code, http.StatusNotImplemented)
	}
}

// TestPlannedStub_PostMergeRequests_ShouldReturnXStatusPlanned verifies the
// x-vedo-status header on POST merge_requests.
func TestPlannedStub_PostMergeRequests_ShouldReturnXStatusPlanned(t *testing.T) {
	w := issuePlannedStub(t, http.MethodPost, "/api/v1/projects/123/merge_requests")
	if got := w.Header().Get("x-vedo-status"); got != "planned" {
		t.Errorf("POST merge_requests x-vedo-status = %q, want %q", got, "planned")
	}
}

// TestPlannedStub_GetProtectedBranches_ShouldReturn501 verifies
// GET /api/v1/projects/{pid}/protected_branches returns 501.
func TestPlannedStub_GetProtectedBranches_ShouldReturn501(t *testing.T) {
	w := issuePlannedStub(t, http.MethodGet, "/api/v1/projects/123/protected_branches")
	if w.Code != http.StatusNotImplemented {
		t.Errorf("GET protected_branches status = %d, want %d", w.Code, http.StatusNotImplemented)
	}
}

// TestPlannedStub_PostProtectedBranches_ShouldReturnXStatusPlanned verifies
// the x-vedo-status header on POST protected_branches.
func TestPlannedStub_PostProtectedBranches_ShouldReturnXStatusPlanned(t *testing.T) {
	w := issuePlannedStub(t, http.MethodPost, "/api/v1/projects/123/protected_branches")
	if got := w.Header().Get("x-vedo-status"); got != "planned" {
		t.Errorf("POST protected_branches x-vedo-status = %q, want %q", got, "planned")
	}
}

// TestPlannedStub_AllPaths_ShouldReturnPlannedBody verifies every planned
// stub returns the contract JSON body with x_vedo_status: planned.
func TestPlannedStub_AllPaths_ShouldReturnPlannedBody(t *testing.T) {
	for _, tc := range plannedStubPaths {
		t.Run(tc.method+"_"+strings.TrimPrefix(tc.path, "/api/v1/projects/123/"), func(t *testing.T) {
			w := issuePlannedStub(t, tc.method, tc.path)
			if w.Code != http.StatusNotImplemented {
				t.Fatalf("status = %d, want 501", w.Code)
			}
			var body map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("response body is not valid JSON: %v", err)
			}
			if body["x_vedo_status"] != "planned" {
				t.Errorf("body x_vedo_status = %q, want %q", body["x_vedo_status"], "planned")
			}
			if !strings.Contains(body["message"], "planned") {
				t.Errorf("body message = %q, want mention of 'planned'", body["message"])
			}
		})
	}
}
