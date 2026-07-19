package org

// Additional RemoveMember tests for last-owner protection.
// BDD: [Condition]_[Action]_[ExpectedResult]
// Validates: REQ-NFR.SECURITY.organization-access-model

import (
	"testing"
)

// TestRemoveMember_NonOwner_ReturnsForbidden verifies non-Owner cannot remove members.
func TestRemoveMember_NonOwner_ReturnsForbidden(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "ontology/test", Type: ScopeOntology})
	_ = store.UpsertMembership(OrgMembership{UserID: "owner", Scope: "ontology/test", Role: "Owner"})
	_ = store.UpsertMembership(OrgMembership{UserID: "editor", Scope: "ontology/test", Role: "Editor"})

	_, err := svc.RemoveMember("editor", "ontology/test", "owner")
	if err == nil {
		t.Fatal("expected error for non-Owner removal, got nil")
	}
	if err != ErrForbiddenAdminOnly {
		t.Fatalf("expected ErrForbiddenAdminOnly, got %v", err)
	}
}

// TestRemoveMember_OwnerRemovesNonOwner_ReturnsOK verifies Owner can remove a non-Owner member.
func TestRemoveMember_OwnerRemovesNonOwner_ReturnsOK(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "ontology/test", Type: ScopeOntology})
	_ = store.UpsertMembership(OrgMembership{UserID: "owner", Scope: "ontology/test", Role: "Owner"})
	_ = store.UpsertMembership(OrgMembership{UserID: "guest", Scope: "ontology/test", Role: "Guest"})

	_, err := svc.RemoveMember("owner", "ontology/test", "guest")
	if err != nil {
		t.Fatalf("expected successful removal, got error: %v", err)
	}

	mems, _ := store.GetMemberships("ontology/test")
	if len(mems) != 1 {
		t.Fatalf("expected 1 remaining member, got %d", len(mems))
	}
}

// TestRemoveMember_OwnerRemovesOtherOwnerWithMultiple_ReturnsOK verifies an Owner
// can remove another Owner if at least one other Owner remains.
func TestRemoveMember_OwnerRemovesOtherOwnerWithMultiple_ReturnsOK(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "ontology/test", Type: ScopeOntology})
	_ = store.UpsertMembership(OrgMembership{UserID: "owner1", Scope: "ontology/test", Role: "Owner"})
	_ = store.UpsertMembership(OrgMembership{UserID: "owner2", Scope: "ontology/test", Role: "Owner"})

	_, err := svc.RemoveMember("owner1", "ontology/test", "owner2")
	if err != nil {
		t.Fatalf("expected successful removal, got error: %v", err)
	}

	mems, _ := store.GetMemberships("ontology/test")
	if len(mems) != 1 {
		t.Fatalf("expected 1 remaining member, got %d", len(mems))
	}
}
