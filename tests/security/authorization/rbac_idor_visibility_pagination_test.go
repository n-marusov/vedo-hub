// RBAC Full Test Suite — IDOR, Visibility Levels, Pagination
//
// Validates: TC-011, TC-012, TC-019 through TC-023, TC-045
// Categories: D (IDOR), G (Visibility Levels), U (Pagination Compliance)

package authorization

import (
	"testing"
)

// ============================================================================
// D. IDOR — Guessable IDs (Type D)
// ============================================================================

// TestRBAC_TC011_SequentialNumericID_Returns403 validates that accessing a
// resource with a sequential numeric ID always returns 403, never 404,
// to prevent resource existence disclosure.
//
// TC-011: Sequential numeric ID returns 403, never 404.
func TestRBAC_TC011_SequentialNumericID_Returns403(t *testing.T) {
	// Given: u_viewer_A is authenticated; ontology with id=1001 exists in tenant_B.
	jwt, ok := setupTest(t, userViewerA)
	if !ok {
		return
	}

	// When: GET /api/v1/ontologies/1001 (existing foreign object).
	resp1 := doRequest(t, "GET", apiBase+"/ontologies/1001", nil, jwt, "")
	if resp1 == nil {
		return
	}
	resp1.Body.Close()
	// Then: 403, never 404 or 500.
	assertStatus(t, resp1, 403, "TC-011: existing foreign numeric ID")
	assertNotStatus(t, resp1, 404, "TC-011: must not return 404 for existing")
	assertNotStatus(t, resp1, 500, "TC-011: must not return 500 for existing")

	// When: GET /api/v1/ontologies/1002 (non-existent foreign object).
	resp2 := doRequest(t, "GET", apiBase+"/ontologies/1002", nil, jwt, "")
	if resp2 == nil {
		return
	}
	defer resp2.Body.Close()
	body2 := readBody(t, resp2)

	// Then: 403 with FORBIDDEN_OBJECT_NOT_FOUND_OR_ACCESS_DENIED, never 404.
	assertStatus(t, resp2, 403, "TC-011: non-existent foreign numeric ID")
	assertNotStatus(t, resp2, 404, "TC-011: must not return 404 for non-existent")
	assertNotStatus(t, resp2, 500, "TC-011: must not return 500 for non-existent")
	assertBodyContains(t, body2, errObjectNotFoundAccessDenied, "TC-011: error code")
}

// TestRBAC_TC012_InvalidUUIDVariant_Returns403 validates that accessing a
// resource with an invalid UUID variant returns 403, never 404.
//
// TC-012: Invalid UUID variant returns 403, never 404.
func TestRBAC_TC012_InvalidUUIDVariant_Returns403(t *testing.T) {
	// Given: u_viewer_A is authenticated; ont_B exists in tenant_B.
	jwt, ok := setupTest(t, userViewerA)
	if !ok {
		return
	}

	// When: GET /api/v1/ontologies/<mutated UUID> (invalid, not in DB).
	invalidUUID := mutateLastHexChar(ontB)
	resp := doRequest(t, "GET", apiBase+"/ontologies/"+invalidUUID, nil, jwt, "")
	if resp == nil {
		return
	}
	defer resp.Body.Close()
	body := readBody(t, resp)

	// Then: 403 with FORBIDDEN_OBJECT_NOT_FOUND_OR_ACCESS_DENIED, never 404.
	assertStatus(t, resp, 403, "TC-012: invalid UUID variant")
	assertNotStatus(t, resp, 404, "TC-012: must not return 404")
	assertNotStatus(t, resp, 500, "TC-012: must not return 500")
	assertBodyContains(t, body, errObjectNotFoundAccessDenied, "TC-012: error code")
}

// mutateLastHexChar changes the last hex character of a UUID to produce
// a predictable but different ID. If the input is not UUID-like, appends "x".
func mutateLastHexChar(id string) string {
	if len(id) == 0 {
		return "x"
	}
	last := id[len(id)-1]
	// Flip between letter/digit to produce a different hex char.
	switch {
	case last >= '0' && last < '9':
		return id[:len(id)-1] + string(last+1)
	case last == '9':
		return id[:len(id)-1] + "a"
	case last >= 'a' && last < 'f':
		return id[:len(id)-1] + string(last+1)
	case last == 'f':
		return id[:len(id)-1] + "0"
	default:
		return id[:len(id)-1] + "a"
	}
}

// ============================================================================
// G. Visibility Levels
// ============================================================================

// TestRBAC_TC019_PrivateProject_NonMember_Returns403 validates that an
// authenticated user without membership in a private project gets 403.
//
// TC-019: Private Project — non-member gets 403.
func TestRBAC_TC019_PrivateProject_NonMember_Returns403(t *testing.T) {
	// Given: proj_private_A exists (private); u_outsider_A is NOT a member.
	jwt, ok := setupTest(t, userOutsiderA)
	if !ok {
		return
	}

	// When: Outsider GETs the private project.
	resp := doRequest(t, "GET", apiBase+"/projects/"+projPrivateA, nil, jwt, "")
	if resp == nil {
		return
	}
	defer resp.Body.Close()
	body := readBody(t, resp)

	// Then: HTTP 403 with FORBIDDEN_INSUFFICIENT_ROLE.
	assertStatus(t, resp, 403, "TC-019: private project, non-member")
	assertBodyContains(t, body, errInsufficientRole, "TC-019: error code")
}

// TestRBAC_TC020_InternalProject_Unauthenticated_Returns401 validates that
// an unauthenticated user gets 401 when accessing an internal project.
//
// TC-020: Internal Project — unauthenticated user gets 401.
func TestRBAC_TC020_InternalProject_Unauthenticated_Returns401(t *testing.T) {
	// Given: proj_internal_A exists with visibility=internal.
	requireAPIAvail(t)

	// When: Anonymous (no JWT) GETs the internal project.
	resp := doRequest(t, "GET", apiBase+"/projects/"+projInternalA, nil, "", "")
	if resp == nil {
		return
	}
	defer resp.Body.Close()

	// Then: HTTP 401 — unauthenticated.
	assertStatus(t, resp, 401, "TC-020: internal project, unauthenticated")
}

// TestRBAC_TC021_InternalProject_AuthenticatedNoRole_Returns200 validates
// that an authenticated user without a specific role can read an internal project.
//
// TC-021: Internal Project — authenticated user with no role gets 200.
func TestRBAC_TC021_InternalProject_AuthenticatedNoRole_Returns200(t *testing.T) {
	// Given: proj_internal_A exists; u_outsider_A is authenticated but NOT a member.
	jwt, ok := setupTest(t, userOutsiderA)
	if !ok {
		return
	}

	// When: Outsider GETs the internal project.
	resp := doRequest(t, "GET", apiBase+"/projects/"+projInternalA, nil, jwt, "")
	if resp == nil {
		return
	}
	defer resp.Body.Close()

	// Then: HTTP 200 — Internal is readable by any authenticated user.
	assertStatus(t, resp, 200, "TC-021: internal project, authenticated")
}

// TestRBAC_TC022_PublicProject_Anonymous_Returns200 validates that an
// anonymous user can read a public project without authentication.
//
// TC-022: Public Project — anonymous user gets 200.
func TestRBAC_TC022_PublicProject_Anonymous_Returns200(t *testing.T) {
	// Given: proj_public_A exists with visibility=public.
	requireAPIAvail(t)

	// When: Anonymous user GETs the public project.
	resp := doRequest(t, "GET", apiBase+"/projects/"+projPublicA, nil, "", "")
	if resp == nil {
		return
	}
	resp.Body.Close()
	// Then: HTTP 200 — public project is readable by anyone.
	assertStatus(t, resp, 200, "TC-022: public project, anonymous")

	// When: Anonymous user GETs ontology classes.
	resp2 := doRequest(t, "GET", apiBase+"/ontologies/"+ontPublicA+"/classes", nil, "", "")
	if resp2 == nil {
		return
	}
	defer resp2.Body.Close()
	// Then: HTTP 200 — ontology content is publicly readable.
	assertStatus(t, resp2, 200, "TC-022: public ontology classes, anonymous")
}

// TestRBAC_TC023_ChangeVisibility_WithoutOwner_Returns403 validates that a
// non-Owner member cannot change project visibility.
//
// TC-023: Change visibility without Owner role returns 403.
func TestRBAC_TC023_ChangeVisibility_WithoutOwner_Returns403(t *testing.T) {
	// Given: u_editor_A is a member of proj_private_A but not Owner.
	jwt, ok := setupTest(t, userEditorA)
	if !ok {
		return
	}

	// When: Editor tries to change visibility via PUT.
	payload := `{"visibility":"public"}`
	resp := doRequest(t, "PUT", apiBase+"/projects/"+projPrivateA+"/visibility", []byte(payload), jwt, "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	if resp == nil {
		return
	}
	defer resp.Body.Close()
	body := readBody(t, resp)

	// Then: HTTP 403 with FORBIDDEN_INSUFFICIENT_ROLE.
	assertStatus(t, resp, 403, "TC-023: Editor changes visibility")
	assertBodyContains(t, body, errInsufficientRole, "TC-023: error code")
}

// ============================================================================
// U. Pagination Compliance
// ============================================================================

// TestRBAC_TC045_Pagination_PerPageExceedsMax_Returns400 validates that
// requesting per_page > 100 is rejected with 400 or 422.
//
// TC-045: per_page > 100 is rejected.
func TestRBAC_TC045_Pagination_PerPageExceedsMax_Returns400(t *testing.T) {
	// Given: u_viewer_A is authenticated.
	jwt, ok := setupTest(t, userViewerA)
	if !ok {
		return
	}

	// When: GET /api/v1/projects?per_page=200 (exceeds max of 100).
	resp := doRequest(t, "GET", apiBase+"/projects?per_page=200", nil, jwt, "")
	if resp == nil {
		return
	}
	defer resp.Body.Close()

	// Then: HTTP 400 or 422 — invalid pagination parameter.
	assertStatusOneOf(t, resp, "TC-045: per_page > 100 expected 400 or 422", 400, 422)
}
