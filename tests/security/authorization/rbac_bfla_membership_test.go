//go:build integration

// RBAC Full Test Suite — BFLA, Membership Boundary, Role Inheritance
//
// Validates: TC-007 through TC-010, TC-013 through TC-018
// Validates: REQ-CON.SECURITY.write-path-invariant
// Categories: C (BFLA), E (Membership Boundary), F (Role Inheritance)

package authorization

import (
	"fmt"
	"testing"
)

// ============================================================================
// C. BFLA — Privilege Escalation (Type C)
// ============================================================================

// TestRBAC_BFLA_PrivilegeEscalation_Returns403 is a table-driven test covering
// all BFLA privilege escalation scenarios (TC-007 through TC-010).
// Each row validates that a user with insufficient role cannot perform a
// privileged operation. All must return 403 with FORBIDDEN_INSUFFICIENT_ROLE.
//
// Validates: TC-007, TC-008, TC-009, TC-010
// Category: C (BFLA)
func TestRBAC_BFLA_PrivilegeEscalation_Returns403(t *testing.T) {
	tcs := []struct {
		tcID    string
		name    string
		user    string
		method  string
		path    string
		body    string
		idemKey string
	}{
		{
			tcID:    "TC-007",
			name:    "Viewer deletes ontology",
			user:    userViewerA,
			method:  "DELETE",
			path:    "/ontologies/" + ontA,
			idemKey: "22222222-2222-2222-2222-222222222222",
		},
		{
			tcID:    "TC-008",
			name:    "Editor adds member",
			user:    userEditorA,
			method:  "POST",
			path:    "/projects/" + projPrivateA + "/members",
			body:    `{"userId":"` + userOutsiderA + `","role":"Viewer"}`,
			idemKey: "33333333-3333-3333-3333-333333333333",
		},
		{
			tcID:    "TC-009",
			name:    "Maintainer changes visibility",
			user:    userMaintainerA,
			method:  "PUT",
			path:    "/projects/" + projPrivateA + "/visibility",
			body:    `{"visibility":"internal"}`,
			idemKey: "44444444-4444-4444-4444-444444444444",
		},
		{
			tcID:    "TC-010",
			name:    "Maintainer creates ABAC policy",
			user:    userMaintainerA,
			method:  "POST",
			path:    "/projects/" + projPrivateA + "/policies",
			body:    `{"pattern":{"uriPrefix":"http://vedo/example/"},"effect":"deny","action":"delete"}`,
			idemKey: "55555555-5555-5555-5555-555555555555",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID+"_"+tc.name, func(t *testing.T) {
			// Given: user is authenticated with insufficient role for this operation.
			jwt, ok := setupTest(t, tc.user)
			if !ok {
				return
			}

			// When: user attempts the privileged operation.
			var bodyBytes []byte
			if tc.body != "" {
				bodyBytes = []byte(tc.body)
			}
			resp := doRequest(t, tc.method, apiBase+tc.path, bodyBytes, jwt, tc.idemKey)
			if resp == nil {
				return
			}
			defer resp.Body.Close()
			body := readBody(t, resp)

			// Then: HTTP 403 with FORBIDDEN_INSUFFICIENT_ROLE.
			assertStatus(t, resp, 403, tc.tcID+": "+tc.name)
			assertBodyContains(t, body, errInsufficientRole, tc.tcID+": error code")
		})
	}
}

// ============================================================================
// E. Membership Boundary (Owner-only)
// ============================================================================

// TestRBAC_TC013_OwnerAddsMember_Returns201 validates that an Owner can add
// a new member to a project (positive test).
//
// TC-013: Owner adds a new member.
func TestRBAC_TC013_OwnerAddsMember_Returns201(t *testing.T) {
	// Given: u_owner_A is authenticated, u_outsider_A is NOT a member.
	jwt, ok := setupTest(t, userOwnerA)
	if !ok {
		return
	}

	// When: Owner POSTs a new member to the project.
	payload := fmt.Sprintf(`{"userId":"%s","role":"Viewer"}`, userOutsiderA)
	resp := doRequest(t, "POST", apiBase+"/projects/"+projPrivateA+"/members", []byte(payload), jwt, "66666666-6666-6666-6666-666666666666")
	if resp == nil {
		return
	}
	defer resp.Body.Close()

	// Then: HTTP 201 (or 200) — member added successfully.
	assertStatusOneOf(t, resp, "TC-013: Owner adds member", 200, 201)
}

// TestRBAC_TC014_OwnerChangesMemberRole_Returns200 validates that an Owner
// can change an existing member's role (positive test).
//
// TC-014: Owner changes member role.
func TestRBAC_TC014_OwnerChangesMemberRole_Returns200(t *testing.T) {
	// Given: u_owner_A is authenticated, u_viewer_A has existing membership.
	jwt, ok := setupTest(t, userOwnerA)
	if !ok {
		return
	}

	// When: Owner PUTs a role change for u_viewer_A.
	payload := `{"role":"Editor"}`
	resp := doRequest(t, "PUT", apiBase+"/projects/"+projPrivateA+"/members/"+userViewerA, []byte(payload), jwt, "77777777-7777-7777-7777-777777777777")
	if resp == nil {
		return
	}
	defer resp.Body.Close()

	// Then: HTTP 200 — role updated.
	assertStatus(t, resp, 200, "TC-014: Owner changes member role")
}

// TestRBAC_TC015_OwnerRemovesMember_Returns200 validates that an Owner can
// remove a member from a project (positive test).
//
// TC-015: Owner removes a member.
func TestRBAC_TC015_OwnerRemovesMember_Returns200(t *testing.T) {
	// Given: u_owner_A is authenticated, u_editor_A is an existing member.
	jwt, ok := setupTest(t, userOwnerA)
	if !ok {
		return
	}

	// When: Owner DELETEs the member.
	resp := doRequest(t, "DELETE", apiBase+"/projects/"+projPrivateA+"/members/"+userEditorA, nil, jwt, "88888888-8888-8888-8888-888888888888")
	if resp == nil {
		return
	}
	defer resp.Body.Close()

	// Then: HTTP 200 or 204 — member removed.
	assertStatusOneOf(t, resp, "TC-015: Owner removes member", 200, 204)
}

// TestRBAC_TC016_ViewerRemovesMember_Returns403 validates that a Viewer
// cannot remove a member. Only Owner can manage members.
//
// TC-016: Viewer attempts to remove a member (negative).
func TestRBAC_TC016_ViewerRemovesMember_Returns403(t *testing.T) {
	// Given: u_viewer_A is authenticated, u_editor_A has membership.
	jwt, ok := setupTest(t, userViewerA)
	if !ok {
		return
	}

	// When: Viewer DELETEs u_editor_A from the project.
	resp := doRequest(t, "DELETE", apiBase+"/projects/"+projPrivateA+"/members/"+userEditorA, nil, jwt, "99999999-9999-9999-9999-999999999999")
	if resp == nil {
		return
	}
	defer resp.Body.Close()
	body := readBody(t, resp)

	// Then: HTTP 403 — Viewer cannot manage members.
	assertStatus(t, resp, 403, "TC-016: Viewer removes member")
	assertBodyContains(t, body, errInsufficientRole, "TC-016: error code")
}

// ============================================================================
// F. Role Inheritance (Max-Role-Wins)
// ============================================================================

// TestRBAC_TC017_MaxRoleWins_OwnerInParentViewerInSubgroup_OwnerEffective
// validates that a user with Owner on parent Group and Viewer on Subgroup
// gets effective Owner via max-role-wins, enabling Owner-only operations.
//
// TC-017: User with Owner in parent Group and Viewer in Subgroup gets Owner.
func TestRBAC_TC017_MaxRoleWins_OwnerInParentViewerInSubgroup_OwnerEffective(t *testing.T) {
	// Given: u_user has Owner on group_parent and Viewer on subgroup_child.
	jwt, ok := setupTest(t, userViewerA)
	if !ok {
		return
	}

	// When: User reads a project in the subgroup.
	resp := doRequest(t, "GET", apiBase+"/projects/"+projInChild, nil, jwt, "")
	if resp == nil {
		return
	}
	resp.Body.Close()
	// Then: Read succeeds via inherited role.
	assertStatus(t, resp, 200, "TC-017: read via inherited role")

	// When: User attempts Owner-only operation (add member).
	payload := fmt.Sprintf(`{"userId":"new_member","role":"Viewer"}`)
	resp2 := doRequest(t, "POST", apiBase+"/projects/"+projInChild+"/members", []byte(payload), jwt, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	if resp2 == nil {
		return
	}
	defer resp2.Body.Close()

	// Then: Max-role-wins → effective Owner, operation succeeds.
	assertStatusOneOf(t, resp2, "TC-017: Owner-only member-add via max-role-wins", 200, 201)
}

// TestRBAC_TC018_ViewerInheritance_FromParentGroup_AppliesToSubProject
// validates that a Viewer role inherited from a parent Group applies to a
// project in a subgroup (read allowed, write denied).
//
// TC-018: Viewer role inheritance from parent Group applies to Subgroup Project.
func TestRBAC_TC018_ViewerInheritance_FromParentGroup_AppliesToSubProject(t *testing.T) {
	// Given: u_viewer_A has Viewer on group_parent (no explicit role on subgroup_child).
	jwt, ok := setupTest(t, userViewerA)
	if !ok {
		return
	}

	// When: User reads a project in the subgroup.
	resp := doRequest(t, "GET", apiBase+"/projects/"+projInChild, nil, jwt, "")
	if resp == nil {
		return
	}
	resp.Body.Close()
	// Then: Read succeeds via inherited Viewer role.
	assertStatus(t, resp, 200, "TC-018: read via inherited Viewer role")

	// When: User attempts a write operation (create class).
	payload := `{"name":"TestClass"}`
	resp2 := doRequest(t, "POST", apiBase+"/ontologies/"+ontInChild+"/classes", []byte(payload), jwt, "")
	if resp2 == nil {
		return
	}
	defer resp2.Body.Close()

	// Then: HTTP 403 — inherited Viewer does NOT grant Editor permission.
	assertStatus(t, resp2, 403, "TC-018: write via inherited Viewer role (must be denied)")
}
