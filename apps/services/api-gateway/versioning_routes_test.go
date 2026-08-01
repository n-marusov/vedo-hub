package main

// Validates: REQ-FUN.API.rest-gitlab-alignment

// Tests for the GitLab-aligned versioning surface in the API Gateway:
//
//   - project-scoped routes /api/v1/projects/{pid}/repository/commits|branches
//     proxy to versioning-service (Task 26);
//   - legacy /api/v1/versioning/* routes stay active but carry deprecation
//     headers during the migration (removal in M10).
//
// Unit-level: the router is built with newTestEnv using httptest upstreams,
// so no external infrastructure is required.

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// issueVersioningRoute performs a request and returns the response headers.
func issueVersioningRoute(t *testing.T, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", env.bearer(t, "user-1", []string{"Owner"}))
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	return w
}

// TestVersioning_ProjectCommitsList_ShouldResolve verifies
// GET /api/v1/projects/{pid}/repository/commits is proxied (not 404).
func TestVersioning_ProjectCommitsList_ShouldResolve(t *testing.T) {
	w := issueVersioningRoute(t, http.MethodGet, "/api/v1/projects/p1/repository/commits")
	if w.Code == http.StatusNotFound {
		t.Error("GET /projects/{pid}/repository/commits resolved to 404 — route not registered")
	}
}

// TestVersioning_ProjectCommitsCreate_ShouldResolve verifies
// POST /api/v1/projects/{pid}/repository/commits is proxied.
func TestVersioning_ProjectCommitsCreate_ShouldResolve(t *testing.T) {
	w := issueVersioningRoute(t, http.MethodPost, "/api/v1/projects/p1/repository/commits")
	if w.Code == http.StatusNotFound {
		t.Error("POST /projects/{pid}/repository/commits resolved to 404 — route not registered")
	}
}

// TestVersioning_ProjectCommitDiff_ShouldResolve verifies
// GET /api/v1/projects/{pid}/repository/commits/{sha}/diff is proxied.
func TestVersioning_ProjectCommitDiff_ShouldResolve(t *testing.T) {
	w := issueVersioningRoute(t, http.MethodGet, "/api/v1/projects/p1/repository/commits/abc123/diff")
	if w.Code == http.StatusNotFound {
		t.Error("GET /projects/{pid}/repository/commits/{sha}/diff resolved to 404 — route not registered")
	}
}

// TestVersioning_ProjectBranchesList_ShouldResolve verifies
// GET /api/v1/projects/{pid}/repository/branches is proxied.
func TestVersioning_ProjectBranchesList_ShouldResolve(t *testing.T) {
	w := issueVersioningRoute(t, http.MethodGet, "/api/v1/projects/p1/repository/branches")
	if w.Code == http.StatusNotFound {
		t.Error("GET /projects/{pid}/repository/branches resolved to 404 — route not registered")
	}
}

// TestVersioning_ProjectBranchGetByName_ShouldResolve verifies
// GET /api/v1/projects/{pid}/repository/branches/{name} is proxied.
func TestVersioning_ProjectBranchGetByName_ShouldResolve(t *testing.T) {
	w := issueVersioningRoute(t, http.MethodGet, "/api/v1/projects/p1/repository/branches/main")
	if w.Code == http.StatusNotFound {
		t.Error("GET /projects/{pid}/repository/branches/{name} resolved to 404 — route not registered")
	}
}

// TestVersioning_ProjectBranchDelete_ShouldResolve verifies
// DELETE /api/v1/projects/{pid}/repository/branches/{name} is proxied.
func TestVersioning_ProjectBranchDelete_ShouldResolve(t *testing.T) {
	w := issueVersioningRoute(t, http.MethodDelete, "/api/v1/projects/p1/repository/branches/old-feat")
	if w.Code == http.StatusNotFound {
		t.Error("DELETE /projects/{pid}/repository/branches/{name} resolved to 404 — route not registered")
	}
}

// TestVersioning_LegacyCommits_ShouldReturnDeprecation verifies the legacy
// /api/v1/versioning/commits route returns the Deprecation header.
func TestVersioning_LegacyCommits_ShouldReturnDeprecation(t *testing.T) {
	w := issueVersioningRoute(t, http.MethodGet, "/api/v1/versioning/commits")
	if got := w.Header().Get("Deprecation"); got != "true" {
		t.Errorf("legacy /api/v1/versioning/commits Deprecation header = %q, want %q", got, "true")
	}
}

// TestVersioning_LegacyBranches_ShouldReturnSunset verifies the legacy
// /api/v1/versioning/branches route returns the Sunset header.
func TestVersioning_LegacyBranches_ShouldReturnSunset(t *testing.T) {
	w := issueVersioningRoute(t, http.MethodGet, "/api/v1/versioning/branches")
	if got := w.Header().Get("Sunset"); got == "" {
		t.Error("legacy /api/v1/versioning/branches Sunset header missing")
	}
}

// TestVersioning_ProjectMerge_ShouldReturn501 verifies the project-scoped
// merge endpoint is a 501 planned stub (MR workflow in M10).
func TestVersioning_ProjectMerge_ShouldReturn501(t *testing.T) {
	w := issueVersioningRoute(t, http.MethodPost, "/api/v1/projects/p1/repository/branches/merge")
	if w.Code != http.StatusNotImplemented {
		t.Errorf("POST /projects/{pid}/repository/branches/merge status = %d, want 501", w.Code)
	}
	if got := w.Header().Get("x-vedo-status"); got != "planned" {
		t.Errorf("merge x-vedo-status = %q, want %q", got, "planned")
	}
}

// TestVersioning_ProjectCheckout_ShouldNotBeREST verifies checkout is NOT
// exposed on the project-scoped surface (internal-only per F3).
func TestVersioning_ProjectCheckout_ShouldNotBeREST(t *testing.T) {
	w := issueVersioningRoute(t, http.MethodPost, "/api/v1/projects/p1/repository/commits/abc123/checkout")
	if w.Code != http.StatusNotFound {
		t.Errorf("checkout must not be a REST endpoint, got status %d", w.Code)
	}
}

// TestVersioning_ProjectSwitch_ShouldNotBeREST verifies switch is removed
// from the REST surface.
func TestVersioning_ProjectSwitch_ShouldNotBeREST(t *testing.T) {
	w := issueVersioningRoute(t, http.MethodPost, "/api/v1/projects/p1/repository/branches/main/switch")
	if w.Code != http.StatusNotFound {
		t.Errorf("switch must be removed from REST, got status %d", w.Code)
	}
}
