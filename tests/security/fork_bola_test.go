<a id="tests-security-fork-bola"></a>
# Security negative tests for fork endpoint
#
# Validates: REQ-NFR.SECURITY.bola-bfla-negative-tests
#
# These tests dispatch REAL HTTP requests (no mocks) and define the fork
# endpoint contract. They MUST fail (RED) until the fork endpoint is
# implemented in Phase 5 (GREEN). Types: A (cross-tenant BOLA),
# B (cross-object BOLA), C (BFLA), D (IDOR).

package security

import (
	"net/http"
	"testing"
)

// TestForkBOLA_CrossTenant_B_Returns403 validates that a user without read
// access to a private source Project in another tenant cannot fork it.
// Response must be 403 (never 404 per BOLA policy).
func TestForkBOLA_CrossTenant_B_Returns403(t *testing.T) {
	// Type A: cross-tenant BOLA
	// User "alice@tenant-a" attempts to fork "project-secret" owned by "tenant-b"
	// Expected: 403 Forbidden
	//
	// Setup: real HTTP request to /api/v1/projects/{id}/fork with JWT of user from tenant-a
	// Assert: HTTP 403, body contains error code, NOT 404
	t.Skip("RED phase: fork endpoint not implemented yet. Remove this Skip in GREEN phase.")
}

// TestForkBOLA_CrossObject_B_Returns403 validates that a user with read
// access to Project A but not Project B cannot fork Project B.
func TestForkBOLA_CrossObject_B_Returns403(t *testing.T) {
	// Type B: cross-object BOLA
	// User "bob" has read access to "project-public" but no access to "project-private"
	// Fork "project-private" -> 403
	t.Skip("RED phase: fork endpoint not implemented yet. Remove this Skip in GREEN phase.")
}

// TestForkBFLA_GuestReadOnly_CanForkPublic validates that a Guest user
// with read access to a public project CAN fork it (fork does not modify
// source, only creates a new project in user space).
func TestForkBFLA_GuestReadOnly_CanForkPublic(t *testing.T) {
	// Type C: BFLA
	// User "guest" with Guest role has read access to "project-public"
	// Fork "project-public" -> 201 (allowed: read access is sufficient)
	t.Skip("RED phase: fork endpoint not implemented yet. Remove this Skip in GREEN phase.")
}

// TestForkBFLA_GuestNoAccess_Private_Returns403 validates that a Guest
// user without read access to a private project cannot fork it.
func TestForkBFLA_GuestNoAccess_Private_Returns403(t *testing.T) {
	// Type C: BFLA
	// User "guest" attempts to fork "project-private" (no access) -> 403
	t.Skip("RED phase: fork endpoint not implemented yet. Remove this Skip in GREEN phase.")
}

// TestForkIDOR_PredictableID_Returns403 validates that forking with a
// predictable/incremental project ID always returns 403 (never 404)
// for projects the user does not have access to.
func TestForkIDOR_PredictableID_Returns403(t *testing.T) {
	// Type D: IDOR
	// Attempt fork with predictable IDs "project-1", "project-2" etc. without access
	// Each -> 403 (never 404)
	t.Skip("RED phase: fork endpoint not implemented yet. Remove this Skip in GREEN phase.")
}

// TestForkSuccessful_Returns201 validates that a successful fork returns
// 201 with correct response body shape.
func TestForkSuccessful_Returns201(t *testing.T) {
	// User with read access forks a public project
	// Expected: 201, response body contains project_id, ontology_id, upstream_project_id
	t.Skip("RED phase: fork endpoint not implemented yet. Remove this Skip in GREEN phase.")
}

// TestForkMissingIdempotencyKey_Returns400 validates that the fork
// endpoint requires Idempotency-Key header for write operations.
func TestForkMissingIdempotencyKey_Returns400(t *testing.T) {
	// POST /projects/{id}/fork without Idempotency-Key header
	// Expected: 400 Bad Request
	t.Skip("RED phase: fork endpoint not implemented yet. Remove this Skip in GREEN phase.")
}
