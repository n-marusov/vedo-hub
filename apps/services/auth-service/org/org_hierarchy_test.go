package org

// Validates: REQ-NFR.SECURITY.organization-access-model
//
// MVP hierarchy depth and project movement tests.
// Verifies MVP group nesting limit (5 levels) and project move semantics.
// BDD: [Condition]_[Action]_[ExpectedResult]
// Validates: REQ-NFR.SECURITY.organization-access-model

import (
	"testing"
)

// setupHierarchyTest creates a nested group tree for testing.
// Returns the OrgService and the deepest scope ID.
func setupHierarchyTest(t *testing.T, depth int) (*OrgService, string) {
	t.Helper()
	store := NewMemStore()
	svc := NewOrgService(store)

	var parentID string
	var lastID string
	for i := 0; i < depth; i++ {
		id := "group/level-" + string(rune('0'+i))
		scope := ScopeNode{ID: id, Type: ScopeGroup, ParentID: parentID, Visibility: VisibilityPrivate}
		if err := store.UpsertScope(scope); err != nil {
			t.Fatalf("failed to create level %d: %v", i, err)
		}
		parentID = id
		lastID = id
	}
	return svc, lastID
}

// TestScopeDepth_RootScope verifies root scope has depth 0.
func TestScopeDepth_RootScope_ReturnsZero(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	_ = store.UpsertScope(ScopeNode{ID: "group/root", Type: ScopeGroup, Visibility: VisibilityPrivate})
	_ = store.UpsertScope(ScopeNode{ID: "group/child", Type: ScopeGroup, ParentID: "group/root", Visibility: VisibilityPrivate})

	depth, err := svc.scopeDepth("group/root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if depth != 0 {
		t.Fatalf("expected root depth 0, got %d", depth)
	}
}

// TestScopeDepth_ChildScope verifies child has correct depth.
func TestScopeDepth_ChildScope_ReturnsCorrectDepth(t *testing.T) {
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
		t.Fatalf("expected depth 2, got %d", depth)
	}
}

// TestCreateScope_ValidDepth verifies creating at depth 5 (the max allowed) succeeds.
func TestCreateScope_ValidDepth_AcceptsMaxLevel(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	// Build chain of 5 levels (depth 5 = levels 0-5 = 6 total groups, creating a 5th-level child)
	currentID := ""
	for i := 0; i < 5; i++ {
		id := "group/lvl" + string(rune('0'+i))
		scope := ScopeNode{ID: id, Type: ScopeGroup, ParentID: currentID, Visibility: VisibilityPrivate}
		if err := svc.CreateScope("owner", scope); err != nil {
			t.Fatalf("unexpected error at level %d: %v", i, err)
		}
		currentID = id
	}

	// Depth should be 4 (5 levels deep = depth 4 since root is 0)
	depth, err := svc.scopeDepth(currentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if depth != 4 {
		t.Fatalf("expected depth 4 for level 5, got %d", depth)
	}
}

// TestCreateScope_ExceedsDepthLimit_ReturnsError verifies creating a group at depth 6 (beyond 5 levels) is rejected.
func TestCreateScope_ExceedsDepthLimit_ReturnsError(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	// Build chain of 6 groups (depth 5 from root = 6 levels, root at depth 0)
	parentID := ""
	for i := 0; i < 6; i++ {
		id := "group/limit-l" + string(rune('0'+i))
		scope := ScopeNode{ID: id, Type: ScopeGroup, ParentID: parentID, Visibility: VisibilityPrivate}
		if err := svc.CreateScope("owner", scope); err != nil {
			t.Fatalf("unexpected error at level %d: %v", i, err)
		}
		parentID = id
	}

	// Now try to create a 7th-level child under the last one
	overflowScope := ScopeNode{ID: "group/too-deep", Type: ScopeGroup, ParentID: parentID, Visibility: VisibilityPrivate}
	err := svc.CreateScope("owner", overflowScope)
	if err == nil {
		t.Fatal("expected hierarchy depth error, got nil")
	}
	if err != ErrHierarchyDepthExceeded {
		t.Fatalf("expected ErrHierarchyDepthExceeded, got %v", err)
	}
}

// TestMoveScope_ValidMove verifies moving a project to a new group succeeds.
func TestMoveScope_ValidMove_UpdatesParent(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	_ = store.UpsertScope(ScopeNode{ID: "group/group-a", Type: ScopeGroup, Visibility: VisibilityPrivate})
	_ = store.UpsertScope(ScopeNode{ID: "group/group-b", Type: ScopeGroup, Visibility: VisibilityPrivate})
	_ = store.UpsertScope(ScopeNode{ID: "ontology/project-1", Type: ScopeOntology, ParentID: "group/group-a", Visibility: VisibilityPrivate})
	_ = store.UpsertMembership(OrgMembership{Scope: "ontology/project-1", UserID: "owner", Role: "Owner"})

	moved, _, err := svc.MoveScope("owner", "ontology/project-1", "group/group-b")
	if err != nil {
		t.Fatalf("expected successful move, got error: %v", err)
	}
	if moved.ParentID != "group/group-b" {
		t.Fatalf("expected parent to be group-b, got %q", moved.ParentID)
	}
}

// TestMoveScope_DepthExceeded_ReturnsError verifies moving to a deep hierarchy is rejected.
func TestMoveScope_DepthExceeded_ReturnsError(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	// Build chain of 6 groups (depth 5 from root = 6 levels)
	parentID := ""
	for i := 0; i < 6; i++ {
		id := "group/deep-l" + string(rune('0'+i))
		_ = store.UpsertScope(ScopeNode{ID: id, Type: ScopeGroup, ParentID: parentID, Visibility: VisibilityPrivate})
		parentID = id
	}

	_ = store.UpsertScope(ScopeNode{ID: "group/shallow", Type: ScopeGroup, Visibility: VisibilityPrivate})
	_ = store.UpsertScope(ScopeNode{ID: "ontology/project-to-move", Type: ScopeOntology, ParentID: "group/shallow", Visibility: VisibilityPrivate})
	_ = store.UpsertMembership(OrgMembership{Scope: "ontology/project-to-move", UserID: "owner", Role: "Owner"})

	// Try to move project to the deepest group (depth 5 + 1 = 7 levels)
	_, _, err := svc.MoveScope("owner", "ontology/project-to-move", parentID)
	if err == nil {
		t.Fatal("expected hierarchy depth error, got nil")
	}
	if err != ErrHierarchyDepthExceeded {
		t.Fatalf("expected ErrHierarchyDepthExceeded, got %v", err)
	}
}

// TestMoveScope_NonOwner_ReturnsForbidden verifies non-Owner cannot move scopes.
func TestMoveScope_NonOwner_ReturnsForbidden(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	_ = store.UpsertScope(ScopeNode{ID: "group/src", Type: ScopeGroup, Visibility: VisibilityPrivate})
	_ = store.UpsertScope(ScopeNode{ID: "group/dst", Type: ScopeGroup, Visibility: VisibilityPrivate})
	_ = store.UpsertScope(ScopeNode{ID: "ontology/proj", Type: ScopeOntology, ParentID: "group/src", Visibility: VisibilityPrivate})
	_ = store.UpsertMembership(OrgMembership{Scope: "ontology/proj", UserID: "editor", Role: "Editor"})

	_, _, err := svc.MoveScope("editor", "ontology/proj", "group/dst")
	if err == nil {
		t.Fatal("expected forbidden error, got nil")
	}
	if err != ErrForbiddenAdminOnly {
		t.Fatalf("expected ErrForbiddenAdminOnly, got %v", err)
	}
}
