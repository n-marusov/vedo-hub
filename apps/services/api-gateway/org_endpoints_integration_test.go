//go:build integration

// Org endpoint smoke tests — requires auth-service gRPC backend.
//
// These tests were extracted from auth_integration_test.go because they require
// the auth-service gRPC backend which is not available in the httptest
// environment used by other tests in this package.
//
// Run with:
//
//	go test -tags=integration -run TestOrgReadEndpointsSmoke -v .
//
// Validates: REQ-NFR.SECURITY.organization-access-model
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestOrgReadEndpointsSmoke verifies that org read endpoints are registered and
// return valid JSON responses when called with an Owner-role JWT.
//
// Requires auth-service gRPC backend to be available (docker-compose integration
// environment). Gracefully skips when auth-service is unreachable.
func TestOrgReadEndpointsSmoke(t *testing.T) {
	env := newTestEnv(t)
	t.Cleanup(env.cleanup)

	// Quick probe: if the first org endpoint can't reach auth-service gRPC,
	// skip the full suite — the integration environment isn't available.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/groups", nil)
	req.Header.Set("Authorization", env.bearer(t, "owner-smoke", []string{"Owner"}))
	req.Header.Set("X-Trace-Id", "trace-smoke-probe")
	w := httptest.NewRecorder()
	env.router.ServeHTTP(w, req)
	if w.Code == http.StatusInternalServerError {
		t.Skip("auth-service gRPC not available — run with docker-compose integration environment")
	}

	tests := []struct {
		name   string
		method string
		path   string
	}{
		// Group reads
		{"ListGroups", http.MethodGet, "/api/v1/groups"},
		{"GetGroup", http.MethodGet, "/api/v1/groups/test-group-1"},
		{"ListChildGroups", http.MethodGet, "/api/v1/groups/test-group-1/subgroups"},
		{"ListGroupMembers", http.MethodGet, "/api/v1/groups/test-group-1/members"},
		// Project reads
		{"ListProjects", http.MethodGet, "/api/v1/projects"},
		{"GetProject", http.MethodGet, "/api/v1/projects/test-project-1"},
		{"ListProjectMembers", http.MethodGet, "/api/v1/projects/test-project-1/members"},
		{"GetVisibility", http.MethodGet, "/api/v1/projects/test-project-1/visibility"},
		{"ListPolicies", http.MethodGet, "/api/v1/projects/test-project-1/policies"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", env.bearer(t, "owner-smoke", []string{"Owner"}))
			req.Header.Set("X-Trace-Id", "trace-smoke-"+tt.name)
			w := httptest.NewRecorder()
			env.router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("%s %s expected 200, got %d (body=%s)", tt.method, tt.path, w.Code, w.Body.String())
				return
			}
			// Verify response is valid JSON
			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Errorf("%s %s response is not valid JSON: %v", tt.method, tt.path, err)
			}
		})
	}
}
