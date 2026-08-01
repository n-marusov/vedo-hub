package main

// Validates: REQ-FUN.API.rest-gitlab-alignment

// RED-phase tests (TDD): ALL routes under /api/v1/ontologies/* MUST return
// the `Deprecation: true` and `Sunset: Sat, 01 Aug 2026 00:00:00 GMT`
// response headers per the REST API realignment (ADR-DES.API.rest-gitlab-alignment).
//
// These tests FAIL against the current code (no deprecation middleware is
// registered yet) and pass only after the middleware is added (Task 18 of
// the plan). They are unit-level: the router is built with newTestEnv using
// httptest upstreams, so no external infrastructure is required.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// sunsetHeaderValue is the required Sunset header value for deprecated
// ontology routes. The middleware formats the date with time.RFC1123, which
// renders UTC timestamps as "UTC" (equivalent to GMT per RFC 7231).
const sunsetHeaderValue = "Sat, 01 Aug 2026 00:00:00 UTC"

// issueDeprecatedRoute performs a request to the given path and returns the
// response headers.
func issueDeprecatedRoute(t *testing.T, method, path string) http.Header {
	t.Helper()
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", env.bearer(t, "user-1", []string{"Owner"}))
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	return w.Header()
}

// TestDeprecation_GetOntologyList_ShouldReturnDeprecationHeader verifies
// GET /api/v1/ontologies returns the Deprecation header.
func TestDeprecation_GetOntologyList_ShouldReturnDeprecationHeader(t *testing.T) {
	h := issueDeprecatedRoute(t, http.MethodGet, "/api/v1/ontologies")
	if got := h.Get("Deprecation"); got != "true" {
		t.Errorf("GET /api/v1/ontologies Deprecation header = %q, want %q", got, "true")
	}
}

// TestDeprecation_GetOntologyList_ShouldReturnSunsetHeader verifies
// GET /api/v1/ontologies returns the Sunset header.
func TestDeprecation_GetOntologyList_ShouldReturnSunsetHeader(t *testing.T) {
	h := issueDeprecatedRoute(t, http.MethodGet, "/api/v1/ontologies")
	if got := h.Get("Sunset"); got != sunsetHeaderValue {
		t.Errorf("GET /api/v1/ontologies Sunset header = %q, want %q", got, sunsetHeaderValue)
	}
}

// TestDeprecation_GetOntology_ShouldReturnDeprecationHeader verifies
// GET /api/v1/ontologies/:id returns the Deprecation header.
func TestDeprecation_GetOntology_ShouldReturnDeprecationHeader(t *testing.T) {
	h := issueDeprecatedRoute(t, http.MethodGet, "/api/v1/ontologies/test-id")
	if got := h.Get("Deprecation"); got != "true" {
		t.Errorf("GET /api/v1/ontologies/:id Deprecation header = %q, want %q", got, "true")
	}
}

// TestDeprecation_GetClasses_ShouldReturnSunsetHeader verifies
// GET /api/v1/ontologies/:id/classes returns the Sunset header.
func TestDeprecation_GetClasses_ShouldReturnSunsetHeader(t *testing.T) {
	h := issueDeprecatedRoute(t, http.MethodGet, "/api/v1/ontologies/test-id/classes")
	if got := h.Get("Sunset"); got != sunsetHeaderValue {
		t.Errorf("GET /api/v1/ontologies/:id/classes Sunset header = %q, want %q", got, sunsetHeaderValue)
	}
}

// TestDeprecation_GetProperties_ShouldReturnDeprecationHeader verifies
// GET /api/v1/ontologies/:id/properties returns the Deprecation header.
func TestDeprecation_GetProperties_ShouldReturnDeprecationHeader(t *testing.T) {
	h := issueDeprecatedRoute(t, http.MethodGet, "/api/v1/ontologies/test-id/properties")
	if got := h.Get("Deprecation"); got != "true" {
		t.Errorf("GET /api/v1/ontologies/:id/properties Deprecation header = %q, want %q", got, "true")
	}
}

// TestDeprecation_GetIndividuals_ShouldReturnSunsetHeader verifies
// GET /api/v1/ontologies/:id/individuals returns the Sunset header.
func TestDeprecation_GetIndividuals_ShouldReturnSunsetHeader(t *testing.T) {
	h := issueDeprecatedRoute(t, http.MethodGet, "/api/v1/ontologies/test-id/individuals")
	if got := h.Get("Sunset"); got != sunsetHeaderValue {
		t.Errorf("GET /api/v1/ontologies/:id/individuals Sunset header = %q, want %q", got, sunsetHeaderValue)
	}
}

// TestDeprecation_PostClass_ShouldReturnDeprecationHeader verifies a write
// route POST /api/v1/ontologies/:id/classes returns the Deprecation header.
func TestDeprecation_PostClass_ShouldReturnDeprecationHeader(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/ontologies/test-id/classes", nil)
	req.Header.Set("Authorization", env.bearer(t, "user-1", []string{"Owner"}))
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if got := w.Header().Get("Deprecation"); got != "true" {
		t.Errorf("POST /api/v1/ontologies/:id/classes Deprecation header = %q, want %q", got, "true")
	}
}

// TestDeprecation_NonOntologyRoute_ShouldNotHaveDeprecationHeader verifies
// a non-deprecated route (project-scoped) does NOT get deprecation headers —
// guards against over-broad middleware.
func TestDeprecation_NonOntologyRoute_ShouldNotHaveDeprecationHeader(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	req.Header.Set("Authorization", env.bearer(t, "user-1", []string{"Owner"}))
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)

	if got := w.Header().Get("Deprecation"); got != "" {
		t.Errorf("GET /api/v1/projects Deprecation header = %q, want empty (not deprecated)", got)
	}
}

// TestDeprecation_HeaderValues_ShouldUseSpecifiedSunset verifies the Sunset
// value matches the plan-specified date (2026-08-01). Go's time.RFC1123
// renders the UTC timestamp as "... UTC" (http.ParseTime only accepts GMT,
// so the value is compared structurally here).
func TestDeprecation_HeaderValues_ShouldUseSpecifiedSunset(t *testing.T) {
	h := issueDeprecatedRoute(t, http.MethodGet, "/api/v1/ontologies/test-id")
	sunset := h.Get("Sunset")
	if sunset == "" {
		t.Fatal("Sunset header not present — deprecation middleware not applied")
	}
	if !strings.Contains(sunset, "01 Aug 2026") {
		t.Errorf("Sunset header %q does not match plan date 01 Aug 2026", sunset)
	}
}

// TestDeprecation_Sunset_ShouldMatchPlanValue verifies the Sunset header is
// the exact value specified by the plan (RFC 8594 date in UTC/GMT).
func TestDeprecation_Sunset_ShouldMatchPlanValue(t *testing.T) {
	h := issueDeprecatedRoute(t, http.MethodGet, "/api/v1/ontologies/test-id")
	got := h.Get("Sunset")
	if got != "Sat, 01 Aug 2026 00:00:00 UTC" && got != "Sat, 01 Aug 2026 00:00:00 GMT" {
		t.Errorf("Sunset header = %q, want Sat, 01 Aug 2026 00:00:00 UTC/GMT", got)
	}
}
