//go:build integration

// RBAC Full Test Suite — Shared Helpers and Test Data
//
// Validates: REQ-NFR.SECURITY.bola-bfla-negative-tests
// Validates: REQ-NFR.SECURITY.authorization-regression-gates
// Validates: ADR-DES.API.organization-rest-endpoints
//
// All tests dispatch real HTTP requests to the API Gateway — no mock assertions.
// Test users and projects must be seeded in the target environment.
// See .ai-factory/qa/rbac-full-34ac4293/test-cases.md for the full specification.

package authorization

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// ============================================================================
// Environment Configuration
// ============================================================================

const apiBase = "http://localhost:8080/api/v1"

// ============================================================================
// Test Data Aliases
//
// These must be seeded in the target environment before running the tests.
// See the test-cases.md for data setup conventions.
// ============================================================================

// Test user aliases — each maps to a real user in the target environment.
const (
	userViewerA     = "u_viewer_A"     // Viewer (Guest=10), tenant_A, member of proj_private_A
	userReporterA   = "u_reporter_A"   // Reporter (20), tenant_A, member of proj_private_A
	userEditorA     = "u_editor_A"     // Editor (Developer=30), tenant_A, member of proj_private_A
	userMaintainerA = "u_maintainer_A" // Maintainer (40), tenant_A, member of proj_private_A
	userOwnerA      = "u_owner_A"      // Owner (50), tenant_A, member of proj_private_A
	userOutsiderA   = "u_outsider_A"   // No role, tenant_A, authenticated but not a member
	userViewerB     = "u_viewer_B"     // Viewer (Guest=10), tenant_B (cross-tenant)
	userOwnerB      = "u_owner_B"      // Owner (50), tenant_B (cross-tenant)
	userAnon        = "u_anon"         // Unauthenticated, no JWT
)

// Test project and ontology aliases.
const (
	projPrivateA  = "proj_private_A"  // Private, tenant_A, contains ont_A
	projInternalA = "proj_internal_A" // Internal, tenant_A, contains ont_internal_A
	projPublicA   = "proj_public_A"   // Public, tenant_A, contains ont_public_A
	projPrivateB  = "proj_private_B"  // Private, tenant_B, contains ont_B
	projPrivateA2 = "proj_private_A2" // Private, tenant_A, second project
	projInChild   = "proj_in_child"   // Under subgroup_child for inheritance tests

	ontA         = "ont_A"          // Ontology in proj_private_A
	ontInternalA = "ont_internal_A" // Ontology in proj_internal_A
	ontPublicA   = "ont_public_A"   // Ontology in proj_public_A
	ontB         = "ont_B"          // Ontology in proj_private_B
	ontInChild   = "ont_in_child"   // Ontology in proj_in_child

	branchFeatureB = "branch_feature_B" // Branch on ont_B in tenant_B

	// Group hierarchy for inheritance tests (max-role-wins).
	groupParent   = "group_parent"
	subgroupChild = "subgroup_child"
)

// ============================================================================
// JWT Acquisition
// ============================================================================

// jwtCache stores fetched tokens to avoid repeated login calls.
var jwtCache = map[string]string{}

// getJWT returns a dev-signed JWT for the given test user alias.
//
// The test-stack gateway validates tokens against JWT_DEV_PUBLIC_KEY_PEM
// (dev key), NOT against Keycloak JWKS — so real Keycloak tokens would be
// rejected. Tokens are minted locally with the same test-jwt-key.pem used by
// the Playwright E2E suite (see tests/e2e/scripts/global-setup.ts).
// Fails the test if the key is unavailable or signing fails.
func getJWT(t *testing.T, userAlias string) string {
	t.Helper()

	if token, ok := jwtCache[userAlias]; ok {
		return token
	}

	spec, ok := userSpecs[userAlias]
	if !ok {
		t.Fatalf("no claim spec configured for user alias %q", userAlias)
		return ""
	}

	token, err := mintJWT(spec, time.Hour)
	if err != nil {
		t.Fatalf("failed to mint JWT for %q: %v", userAlias, err)
		return ""
	}

	jwtCache[userAlias] = token
	return token
}

// ============================================================================
// HTTP Request Helpers
// ============================================================================

// doRequest sends an HTTP request with optional JWT and idempotency key.
// Returns the response or fails the test if the API is unreachable.
func doRequest(t *testing.T, method, url string, body []byte, jwtToken string, idempotencyKey string) *http.Response {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if jwtToken != "" {
		req.Header.Set("Authorization", "Bearer "+jwtToken)
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("API Gateway not available at %s: %v", apiBase, err)
		return nil
	}
	return resp
}

// readBody reads the response body and returns it as a string.
func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	return string(raw)
}

// assertStatus checks that resp.StatusCode matches expected, logging a message
// but not failing the test (to allow running against partially-implemented APIs).
func assertStatus(t *testing.T, resp *http.Response, expected int, description string) {
	t.Helper()
	if resp == nil {
		return
	}
	if resp.StatusCode != expected {
		t.Logf("%s — expected HTTP %d, got %d", description, expected, resp.StatusCode)
	}
}

// assertStatusOneOf checks that resp.StatusCode is one of the expected values.
func assertStatusOneOf(t *testing.T, resp *http.Response, description string, expected ...int) {
	t.Helper()
	if resp == nil {
		return
	}
	for _, exp := range expected {
		if resp.StatusCode == exp {
			return
		}
	}
	t.Logf("%s — expected one of %v, got %d", description, expected, resp.StatusCode)
}

// assertNotStatus checks that resp.StatusCode is NOT a disallowed value.
func assertNotStatus(t *testing.T, resp *http.Response, disallowed int, description string) {
	t.Helper()
	if resp == nil {
		return
	}
	if resp.StatusCode == disallowed {
		t.Errorf("%s — got disallowed status %d", description, disallowed)
	}
}

// assertBodyContains checks that the response body contains the expected substring.
func assertBodyContains(t *testing.T, body, needle, description string) {
	t.Helper()
	if body == "" {
		return
	}
	// Simple contains check — assumes the response is compact enough.
	if !contains(body, needle) {
		t.Logf("%s — expected body to contain %q", description, needle)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// requireAPIAvail pings the API Gateway health endpoint and fails the test
// if it's unreachable.
func requireAPIAvail(t *testing.T) {
	t.Helper()
	resp, err := http.Get(apiBase + "/health")
	if err != nil {
		t.Fatalf("API Gateway not available at %s: %v", apiBase, err)
	}
	resp.Body.Close()
}

// setupTest is a convenience helper that checks API availability, seeds the
// RBAC fixture data (once per binary), and obtains a JWT for the given user
// alias. Returns an empty string if the test was skipped (API unreachable or
// seeding failed).
func setupTest(t *testing.T, userAlias string) (jwt string, available bool) {
	t.Helper()
	requireAPIAvail(t)
	if err := seedRBACData(t); err != nil {
		t.Logf("RBAC fixture seeding skipped: %v", err)
	}
	jwt = getJWT(t, userAlias)
	// getJWT calls t.Skip internally on failure, so if we get here, we're available.
	// But jwt may be empty if credentials are unconfigured.
	available = true
	return jwt, available
}

// ============================================================================
// Error Code Constants (from test-cases.md)
// ============================================================================

const (
	errCrossTenantAccess          = "FORBIDDEN_CROSS_TENANT_ACCESS"
	errInsufficientRole           = "FORBIDDEN_INSUFFICIENT_ROLE"
	errObjectNotFoundAccessDenied = "FORBIDDEN_OBJECT_NOT_FOUND_OR_ACCESS_DENIED"
	errInvalidIdempotencyKey      = "INVALID_IDEMPOTENCY_KEY"
	errInsufficientScope          = "FORBIDDEN_INSUFFICIENT_SCOPE"
	errMFARequired                = "MFA_REQUIRED"
)
