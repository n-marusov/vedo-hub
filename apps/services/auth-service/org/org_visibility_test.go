package org

// Validates: REQ-NFR.SECURITY.organization-access-model
//
// Visibility inheritance tests for subgroup creation.
// Verifies MVP visibility rules: subgroup cannot be more visible than parent.
// BDD: [Condition]_[Action]_[ExpectedResult]

import (
	"testing"
)

// TestCreateScope_SubgroupVisibilityInherited checks that a subgroup without
// explicit visibility inherits from its parent.
func TestCreateScope_SubgroupVisibilityInherited(t *testing.T) {
	tests := []struct {
		name        string
		parentVis   Visibility
		expectedVis Visibility
	}{
		{"inherit private from parent", VisibilityPrivate, VisibilityPrivate},
		{"inherit internal from parent", VisibilityInternal, VisibilityInternal},
		{"inherit public from parent", VisibilityPublic, VisibilityPublic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemStore()
			svc := NewOrgService(store)

			parent := ScopeNode{ID: "parent", Type: ScopeGroup, Visibility: tt.parentVis}
			if err := store.UpsertScope(parent); err != nil {
				t.Fatalf("failed to create parent: %v", err)
			}

			// Create child WITHOUT explicit visibility
			child := ScopeNode{ID: "child", Type: ScopeGroup, ParentID: "parent", Visibility: ""}
			if err := svc.CreateScope("owner", child); err != nil {
				t.Fatalf("CreateScope failed: %v", err)
			}

			// Verify child inherited parent visibility
			stored, _ := store.GetScope("child")
			if stored == nil {
				t.Fatal("child scope not found")
			}
			if stored.Visibility != tt.expectedVis {
				t.Fatalf("expected visibility %s, got %s", tt.expectedVis, stored.Visibility)
			}
		})
	}
}

// TestCreateScope_SubgroupVisibilityRejected checks that a subgroup with
// more permissive visibility than its parent is rejected.
func TestCreateScope_SubgroupVisibilityRejected(t *testing.T) {
	tests := []struct {
		name       string
		parentVis  Visibility
		childVis   Visibility
		shouldFail bool
	}{
		{"reject public child under private parent", VisibilityPrivate, VisibilityPublic, true},
		{"reject public child under internal parent", VisibilityInternal, VisibilityPublic, true},
		{"reject internal child under private parent", VisibilityPrivate, VisibilityInternal, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemStore()
			svc := NewOrgService(store)

			parent := ScopeNode{ID: "parent", Type: ScopeGroup, Visibility: tt.parentVis}
			if err := store.UpsertScope(parent); err != nil {
				t.Fatalf("failed to create parent: %v", err)
			}

			child := ScopeNode{ID: "child", Type: ScopeGroup, ParentID: "parent", Visibility: tt.childVis}
			err := svc.CreateScope("owner", child)
			if tt.shouldFail && err == nil {
				t.Fatal("expected error but CreateScope succeeded")
			}
			if !tt.shouldFail && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestCreateScope_SubgroupVisibilityAllowed checks valid visibility combinations.
func TestCreateScope_SubgroupVisibilityAllowed(t *testing.T) {
	tests := []struct {
		name      string
		parentVis Visibility
		childVis  Visibility
	}{
		{"allow private child under internal parent", VisibilityInternal, VisibilityPrivate},
		{"allow private child under public parent", VisibilityPublic, VisibilityPrivate},
		{"allow internal child under public parent", VisibilityPublic, VisibilityInternal},
		{"allow internal child under internal parent", VisibilityInternal, VisibilityInternal},
		{"allow public child under public parent", VisibilityPublic, VisibilityPublic},
		{"allow private child under private parent", VisibilityPrivate, VisibilityPrivate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemStore()
			svc := NewOrgService(store)

			parent := ScopeNode{ID: "parent", Type: ScopeGroup, Visibility: tt.parentVis}
			if err := store.UpsertScope(parent); err != nil {
				t.Fatalf("failed to create parent: %v", err)
			}

			child := ScopeNode{ID: "child", Type: ScopeGroup, ParentID: "parent", Visibility: tt.childVis}
			if err := svc.CreateScope("owner", child); err != nil {
				t.Fatalf("CreateScope failed unexpectedly: %v", err)
			}
		})
	}
}

// TestCreateScope_TopLevelAnyVisibility checks that top-level groups can have any visibility.
func TestCreateScope_TopLevelAnyVisibility(t *testing.T) {
	tests := []Visibility{VisibilityPrivate, VisibilityInternal, VisibilityPublic}

	for _, vis := range tests {
		t.Run(string(vis), func(t *testing.T) {
			store := NewMemStore()
			svc := NewOrgService(store)

			child := ScopeNode{ID: "top-level", Type: ScopeGroup, ParentID: "", Visibility: vis}
			if err := svc.CreateScope("owner", child); err != nil {
				t.Fatalf("CreateScope failed for top-level %s: %v", vis, err)
			}

			stored, _ := store.GetScope("top-level")
			if stored.Visibility != vis {
				t.Fatalf("expected visibility %s, got %s", vis, stored.Visibility)
			}
		})
	}
}

// TestVisibilityLevel confirms the numeric mapping for visibility comparison.
func TestVisibilityLevel(t *testing.T) {
	if visibilityLevel(VisibilityPrivate) != 0 {
		t.Fatal("Private should map to 0")
	}
	if visibilityLevel(VisibilityInternal) != 1 {
		t.Fatal("Internal should map to 1")
	}
	if visibilityLevel(VisibilityPublic) != 2 {
		t.Fatal("Public should map to 2")
	}
}
