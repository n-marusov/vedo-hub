package org

// Validates: REQ-NFR.SECURITY.organization-access-model
//
// MVP role vocabulary contract tests.
// Verifies that the API accepts MVP roles (Guest, Reporter, Developer, Maintainer, Owner)
// and maintains backward compatibility with legacy roles (Viewer, Editor).
// BDD: [Condition]_[Action]_[ExpectedResult]
// Validates: REQ-NFR.SECURITY.organization-access-model

import (
	"fmt"
	"testing"
)

// TestMvpRole_Guest_IsAccepted verifies Guest is a valid role.
func TestMvpRole_Guest_IsAccepted(t *testing.T) {
	if _, ok := rolePriority["Guest"]; !ok {
		t.Fatal("expected Guest role to be registered in rolePriority")
	}
}

// TestMvpRole_Reporter_IsAccepted verifies Reporter is a valid role.
func TestMvpRole_Reporter_IsAccepted(t *testing.T) {
	if _, ok := rolePriority["Reporter"]; !ok {
		t.Fatal("expected Reporter role to be registered in rolePriority")
	}
}

// TestMvpRole_Developer_IsAccepted verifies Developer is a valid role.
func TestMvpRole_Developer_IsAccepted(t *testing.T) {
	if _, ok := rolePriority["Developer"]; !ok {
		t.Fatal("expected Developer role to be registered in rolePriority")
	}
}

// TestMvpRole_Maintainer_IsAccepted verifies Maintainer is a valid role.
func TestMvpRole_Maintainer_IsAccepted(t *testing.T) {
	if _, ok := rolePriority["Maintainer"]; !ok {
		t.Fatal("expected Maintainer role to be registered in rolePriority")
	}
}

// TestMvpRole_Owner_IsAccepted verifies Owner is a valid role.
func TestMvpRole_Owner_IsAccepted(t *testing.T) {
	if _, ok := rolePriority["Owner"]; !ok {
		t.Fatal("expected Owner role to be registered in rolePriority")
	}
}

// TestMvpRole_Guest_HasCorrectPriority verifies Guest has the lowest priority.
func TestMvpRole_Guest_HasCorrectPriority(t *testing.T) {
	if rolePriority["Guest"] != 1 {
		t.Fatalf("expected Guest priority 1, got %d", rolePriority["Guest"])
	}
}

// TestMvpRole_Owner_HasHighestPriority verifies Owner has the highest priority.
func TestMvpRole_Owner_HasHighestPriority(t *testing.T) {
	for role, prio := range rolePriority {
		if prio > rolePriority["Owner"] {
			t.Fatalf("expected Owner to have highest priority, but %s has %d > %d", role, prio, rolePriority["Owner"])
		}
	}
}

// TestMvpRole_ResolveMaxRole_WithMvpRoles verifies ResolveMaxRole works with MVP roles.
func TestMvpRole_ResolveMaxRole_WithMvpRoles(t *testing.T) {
	memberships := []OrgMembership{
		{Role: "Guest", Inherited: true},
		{Role: "Developer", Inherited: false},
	}
	maxRole := ResolveMaxRole(memberships)
	if maxRole != "Developer" {
		t.Fatalf("expected Developer as max role, got %q", maxRole)
	}
}

// TestLegacyRole_Viewer_IsAccepted verifies Viewer is still a valid backward-compat role.
func TestLegacyRole_Viewer_IsAccepted(t *testing.T) {
	if _, ok := rolePriority["Viewer"]; !ok {
		t.Fatal("expected legacy Viewer role to be preserved in rolePriority")
	}
}

// TestLegacyRole_Editor_IsAccepted verifies Editor is still a valid backward-compat role.
func TestLegacyRole_Editor_IsAccepted(t *testing.T) {
	if _, ok := rolePriority["Editor"]; !ok {
		t.Fatal("expected legacy Editor role to be preserved in rolePriority")
	}
}

// TestLegacyRole_Viewer_Guest_SamePriority verifies Viewer and Guest have the same priority.
func TestLegacyRole_Viewer_Guest_SamePriority(t *testing.T) {
	if rolePriority["Viewer"] != rolePriority["Guest"] {
		t.Fatalf("expected Viewer(%d) and Guest(%d) to have same priority", rolePriority["Viewer"], rolePriority["Guest"])
	}
}

// TestLegacyRole_Editor_Developer_SamePriority verifies Editor and Developer have same priority.
func TestLegacyRole_Editor_Developer_SamePriority(t *testing.T) {
	if rolePriority["Editor"] != rolePriority["Developer"] {
		t.Fatalf("expected Editor(%d) and Developer(%d) to have same priority", rolePriority["Editor"], rolePriority["Developer"])
	}
}

// TestEnterpriseRole_SupportEngineer_OutsideMvpPath verifies SupportEngineer exists but is not in MVP core set.
func TestEnterpriseRole_SupportEngineer_OutsideMvpPath(t *testing.T) {
	if _, ok := rolePriority["SupportEngineer"]; !ok {
		t.Fatal("expected SupportEngineer role to be preserved")
	}
	// Should not be a core MVP role
	mvpRoles := map[string]bool{"Guest": true, "Reporter": true, "Developer": true, "Maintainer": true, "Owner": true}
	if mvpRoles["SupportEngineer"] {
		t.Fatal("expected SupportEngineer NOT to be in core MVP role set")
	}
}

// ============================================================================
// Hierarchy Depth & Project Move Tests (MVP Task 25)
// ============================================================================

// TestScopeDepth_RootScope verifies root scopes have depth 0.
func TestScopeDepth_RootScope(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "group/root", Type: ScopeGroup, Visibility: VisibilityPrivate})

	depth, err := svc.scopeDepth("group/root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if depth != 0 {
		t.Fatalf("expected depth 0 for root, got %d", depth)
	}
}

// TestScopeDepth_NestedScope verifies depth calculation for nested scopes.
func TestScopeDepth_NestedScope(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "group/l1", Type: ScopeGroup, Visibility: VisibilityPrivate})
	_ = store.UpsertScope(ScopeNode{ID: "group/l2", Type: ScopeGroup, ParentID: "group/l1", Visibility: VisibilityPrivate})
	_ = store.UpsertScope(ScopeNode{ID: "group/l3", Type: ScopeGroup, ParentID: "group/l2", Visibility: VisibilityPrivate})

	depth, err := svc.scopeDepth("group/l3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if depth != 2 {
		t.Fatalf("expected depth 2 for l3 under l2 under l1, got %d", depth)
	}
}

// TestCreateScope_ValidDepth verifies creating a group at the maximum depth (5) succeeds.
func TestCreateScope_ValidDepth(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	// Create chain: root → l1 → l2 → l3 → l4 (depth 4)
	_ = store.UpsertScope(ScopeNode{ID: "group/root", Type: ScopeGroup, Visibility: VisibilityPrivate})
	for i := 1; i <= 4; i++ {
		parentID := "group/root"
		if i > 1 {
			parentID = fmt.Sprintf("group/l%d", i-1)
		}
		_ = store.UpsertScope(ScopeNode{ID: fmt.Sprintf("group/l%d", i), Type: ScopeGroup, ParentID: parentID, Visibility: VisibilityPrivate})
	}
	// Create at depth 5 (should succeed — the new node itself is at depth 5, its parent is at depth 4)
	err := svc.CreateScope("", ScopeNode{ID: "group/l5", Type: ScopeGroup, ParentID: "group/l4", Visibility: VisibilityPrivate})
	if err != nil {
		t.Fatalf("expected depth 5 to be allowed, got error: %v", err)
	}
}

// TestCreateScope_ExceedsDepthLimit verifies creating a group beyond 5 levels returns error.
func TestCreateScope_ExceedsDepthLimit(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	// Create chain: root → l1 → l2 → l3 → l4 → l5 (depth 5)
	_ = store.UpsertScope(ScopeNode{ID: "group/root", Type: ScopeGroup, Visibility: VisibilityPrivate})
	for i := 1; i <= 5; i++ {
		parentID := "group/root"
		if i > 1 {
			parentID = fmt.Sprintf("group/l%d", i-1)
		}
		_ = store.UpsertScope(ScopeNode{ID: fmt.Sprintf("group/l%d", i), Type: ScopeGroup, ParentID: parentID, Visibility: VisibilityPrivate})
	}
	// Try creating at depth 6 (should fail — parent is at depth 5)
	err := svc.CreateScope("", ScopeNode{ID: "group/l6", Type: ScopeGroup, ParentID: "group/l5", Visibility: VisibilityPrivate})
	if err != ErrHierarchyDepthExceeded {
		t.Fatalf("expected ErrHierarchyDepthExceeded, got %v", err)
	}
}

// TestMoveScope_ValidMove verifies moving a project to a different group.
func TestMoveScope_ValidMove(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "group/root", Type: ScopeGroup, Visibility: VisibilityPrivate})
	_ = store.UpsertScope(ScopeNode{ID: "group/target", Type: ScopeGroup, ParentID: "group/root", Visibility: VisibilityPrivate})
	_ = store.UpsertScope(ScopeNode{ID: "ontology/proj-1", Type: ScopeOntology, ParentID: "group/root", Visibility: VisibilityPrivate})
	_ = store.UpsertMembership(OrgMembership{Scope: "group/root", UserID: "owner", Role: "Owner"})

	moved, audit, err := svc.MoveScope("owner", "ontology/proj-1", "group/target")
	if err != nil {
		t.Fatalf("expected successful move, got error: %v", err)
	}
	if moved.ParentID != "group/target" {
		t.Fatalf("expected parent_id to be group/target, got %q", moved.ParentID)
	}
	if audit.Event != "scope.moved" {
		t.Fatalf("expected audit event scope.moved, got %q", audit.Event)
	}
}

// TestMoveScope_DepthExceeded verifies moving to a deep hierarchy is rejected.
func TestMoveScope_DepthExceeded(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	// Create chain at depth 5
	_ = store.UpsertScope(ScopeNode{ID: "group/root", Type: ScopeGroup, Visibility: VisibilityPrivate})
	for i := 1; i <= 5; i++ {
		parentID := "group/root"
		if i > 1 {
			parentID = fmt.Sprintf("group/l%d", i-1)
		}
		_ = store.UpsertScope(ScopeNode{ID: fmt.Sprintf("group/l%d", i), Type: ScopeGroup, ParentID: parentID, Visibility: VisibilityPrivate})
	}
	// Project at root
	_ = store.UpsertScope(ScopeNode{ID: "ontology/proj-1", Type: ScopeOntology, ParentID: "group/root", Visibility: VisibilityPrivate})
	_ = store.UpsertMembership(OrgMembership{Scope: "ontology/proj-1", UserID: "owner", Role: "Owner"})

	// Try moving to group/l5 (depth 5 parent → would make project depth 6)
	_, _, err := svc.MoveScope("owner", "ontology/proj-1", "group/l5")
	if err != ErrHierarchyDepthExceeded {
		t.Fatalf("expected ErrHierarchyDepthExceeded, got %v", err)
	}
}

// TestMoveScope_NonOwner verifies non-Owner cannot move scopes.
func TestMoveScope_NonOwner(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "group/root", Type: ScopeGroup, Visibility: VisibilityPrivate})
	_ = store.UpsertScope(ScopeNode{ID: "group/target", Type: ScopeGroup, ParentID: "group/root", Visibility: VisibilityPrivate})
	_ = store.UpsertScope(ScopeNode{ID: "ontology/proj-1", Type: ScopeOntology, ParentID: "group/root", Visibility: VisibilityPrivate})
	_ = store.UpsertMembership(OrgMembership{Scope: "group/root", UserID: "dev", Role: "Developer"})

	_, _, err := svc.MoveScope("dev", "ontology/proj-1", "group/target")
	if err != ErrForbiddenAdminOnly {
		t.Fatalf("expected ErrForbiddenAdminOnly, got %v", err)
	}
}

// TestMoveScope_InvalidScope verifies moving non-existent scope returns error.
func TestMoveScope_InvalidScope(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "group/root", Type: ScopeGroup, Visibility: VisibilityPrivate})
	_ = store.UpsertMembership(OrgMembership{Scope: "group/root", UserID: "owner", Role: "Owner"})

	_, _, err := svc.MoveScope("owner", "ontology/nonexistent", "group/root")
	if err != ErrScopeNotFound {
		t.Fatalf("expected ErrScopeNotFound, got %v", err)
	}
}

// TestMoveScope_InvalidParent verifies moving to non-existent parent returns error.
func TestMoveScope_InvalidParent(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "group/root", Type: ScopeGroup, Visibility: VisibilityPrivate})
	_ = store.UpsertScope(ScopeNode{ID: "ontology/proj-1", Type: ScopeOntology, ParentID: "group/root", Visibility: VisibilityPrivate})
	_ = store.UpsertMembership(OrgMembership{Scope: "ontology/proj-1", UserID: "owner", Role: "Owner"})

	_, _, err := svc.MoveScope("owner", "ontology/proj-1", "group/nonexistent")
	if err != ErrScopeNotFound {
		t.Fatalf("expected ErrScopeNotFound, got %v", err)
	}
}

// TestUpdateMembership_AcceptsMvpRole_Guest verifies UpdateMembership accepts Guest role.
func TestUpdateMembership_AcceptsMvpRole_Guest(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "group/test", Type: ScopeGroup, Visibility: VisibilityPrivate})
	_ = store.UpsertMembership(OrgMembership{Scope: "group/test", UserID: "owner", Role: "Owner"})

	_, _, err := svc.UpdateMembership("owner", "group/test", "user1", "Guest")
	if err != nil {
		t.Fatalf("expected Guest role to be accepted, got error: %v", err)
	}
}

// TestUpdateMembership_AcceptsMvpRole_Developer verifies UpdateMembership accepts Developer role.
func TestUpdateMembership_AcceptsMvpRole_Developer(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "group/test", Type: ScopeGroup, Visibility: VisibilityPrivate})
	_ = store.UpsertMembership(OrgMembership{Scope: "group/test", UserID: "owner", Role: "Owner"})

	_, _, err := svc.UpdateMembership("owner", "group/test", "user1", "Developer")
	if err != nil {
		t.Fatalf("expected Developer role to be accepted, got error: %v", err)
	}
}
