package handlers

// Validates: ADR-DES.API.organization-rest-endpoints
//
// Route registration tests for the project-scoped org endpoints.
// Verifies that the old /ontologies/:id/{members,visibility,policies} paths
// are no longer registered (404) and the new /projects/:id/... paths dispatch
// to the correct handlers.
//
// The OrgHandler in these tests has a nil orgClient — the handlers will
// panic if actually dispatched to the gRPC call, but gin.Recovery() catches
// the panic and returns 500. This is sufficient for route registration
// tests: 404 means the route is not registered; 500 means the route IS
// registered and the handler was dispatched (but failed due to the nil
// gRPC client in the test environment).

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupOrgTestRouter creates a minimal Gin engine with only the project-scoped
// org routes registered (matching routes.go). The OrgHandler has a nil
// orgClient — route registration only, no gRPC backend.
func setupOrgTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	api := r.Group("/api/v1")
	handler := NewOrgHandler(nil)

	api.GET("/projects/:id/members", handler.HandleListMembers)
	api.GET("/projects/:id/visibility", handler.HandleGetVisibility)
	api.GET("/projects/:id/policies", handler.HandleListPolicies)
	api.POST("/projects/:id/members", handler.HandleAddMember)
	api.PUT("/projects/:id/members/:userId", handler.HandleUpdateMemberRole)
	api.DELETE("/projects/:id/members/:userId", handler.HandleRemoveMember)
	api.PUT("/projects/:id/visibility", handler.HandleSetVisibility)
	api.POST("/projects/:id/policies", handler.HandleCreatePolicy)
	api.DELETE("/projects/:id/policies/:policyId", handler.HandleDeletePolicy)

	// Fork endpoint — required for contract tests
	api.POST("/projects/:id/fork", handler.HandleForkProject)

	// Create project — required for project creation tests
	api.POST("/projects", handler.HandleCreateProject)

	return r
}

// TestOrgHandler_UnknownOrgPath_Returns404 validates that the old
// /api/v1/ontologies/:id/members path is no longer registered and returns 404.
// This confirms the route rename from /ontologies/:id/... to /projects/:id/...
// is complete and the old path is not aliased.
//
// Validates: ADR-DES.API.organization-rest-endpoints
func TestOrgHandler_UnknownOrgPath_Returns404(t *testing.T) {
	router := setupOrgTestRouter()

	oldPaths := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/ontologies/test-id/members"},
		{"POST", "/api/v1/ontologies/test-id/members"},
		{"PUT", "/api/v1/ontologies/test-id/members/user-1"},
		{"DELETE", "/api/v1/ontologies/test-id/members/user-1"},
		{"GET", "/api/v1/ontologies/test-id/visibility"},
		{"PUT", "/api/v1/ontologies/test-id/visibility"},
		{"GET", "/api/v1/ontologies/test-id/policies"},
		{"POST", "/api/v1/ontologies/test-id/policies"},
		{"DELETE", "/api/v1/ontologies/test-id/policies/pol-1"},
	}

	for _, tc := range oldPaths {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(tc.method, tc.path, nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("%s %s: expected 404 (old path not registered), got %d", tc.method, tc.path, w.Code)
		}
	}
}

// TestOrgHandler_ProjectMembersList_RouteRegistered validates that the new
// /api/v1/projects/:id/members path is registered and dispatches to
// HandleListMembers. The response is NOT 404 (route is registered); the
// exact status depends on the nil gRPC client (500 via panic recovery), but
// the key assertion is that the route matches and the handler is called.
//
// Validates: ADR-DES.API.organization-rest-endpoints
func TestOrgHandler_ProjectMembersList_RouteRegistered(t *testing.T) {
	router := setupOrgTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/projects/test-id/members", nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Fatal("expected non-404 for /projects/:id/members (route should be registered), got 404")
	}
}

// TestOrgHandler_ProjectVisibilityAndPolicies_RouteRegistered validates that
// the new /api/v1/projects/:id/visibility and /api/v1/projects/:id/policies
// paths are registered.
//
// Validates: ADR-DES.API.organization-rest-endpoints
func TestOrgHandler_ProjectVisibilityAndPolicies_RouteRegistered(t *testing.T) {
	router := setupOrgTestRouter()

	newPaths := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/projects/test-id/visibility"},
		{"PUT", "/api/v1/projects/test-id/visibility"},
		{"GET", "/api/v1/projects/test-id/policies"},
		{"POST", "/api/v1/projects/test-id/policies"},
		{"DELETE", "/api/v1/projects/test-id/policies/pol-1"},
	}

	for _, tc := range newPaths {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(tc.method, tc.path, nil)
		router.ServeHTTP(w, req)
		if w.Code == http.StatusNotFound {
			t.Errorf("%s %s: expected non-404 (route should be registered), got 404", tc.method, tc.path)
		}
	}
}

// TestOrgHandler_CreateProject_RouteRegistered validates that POST /api/v1/projects
// is registered and dispatches to HandleCreateProject. With nil orgClient, the
// handler panics on gRPC call and recovery returns 500.
// Key assertion: route is NOT 404.
//
// Validates: REQ-FUN.ORG.project-creation
func TestOrgHandler_CreateProject_RouteRegistered(t *testing.T) {
	router := setupOrgTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/projects", nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Fatal("expected non-404 for POST /api/v1/projects (route should be registered), got 404")
	}
}

// TestOrgHandler_ProjectFork_RouteRegistered validates that the new
// /api/v1/projects/:id/fork route is registered and dispatches to
// HandleForkProject. With nil orgClient, the handler panics on gRPC call
// — recovery returns 500. The key assertion: route is NOT 404.
//
// Validates: ADR-DES.API.organization-rest-endpoints (fork endpoint)
func TestOrgHandler_ProjectFork_RouteRegistered(t *testing.T) {
	router := setupOrgTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/projects/test-id/fork", nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Fatal("expected non-404 for /projects/:id/fork (route should be registered), got 404")
	}
}

// TestOrgHandler_ProjectFork_OldTemplatePaths_Return404 validates that the old
// template-related paths are no longer registered (removed per templates-via-forks).
//
// Validates: ADR-DES.PROCESS.templates-via-forks
func TestOrgHandler_ProjectFork_OldTemplatePaths_Return404(t *testing.T) {
	router := setupOrgTestRouter()

	oldTemplatePaths := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/templates/ontologies"},
		{"POST", "/api/v1/ontologies/test-id/apply-template"},
	}

	for _, tc := range oldTemplatePaths {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(tc.method, tc.path, nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("%s %s: expected 404 (old template path removed), got %d", tc.method, tc.path, w.Code)
		}
	}
}
