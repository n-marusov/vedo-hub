package membership

// Validates: REQ-NFR.SECURITY.organization-access-model

import "testing"

func TestResolveEffectiveRole_ReturnsMax(t *testing.T) {
	// @hlv max_role_wins
	memberships := []Membership{
		{UserID: "user1", Scope: "ont-123", Role: "Viewer"},
		{UserID: "user1", Scope: "ont-123", Role: "Editor"},
		{UserID: "user1", Scope: "ont-123", Role: "Owner"},
	}
	got := ResolveEffectiveRole(memberships)
	if got != "Owner" {
		t.Errorf("expected 'Owner', got %q", got)
	}
}

func TestResolveEffectiveRole_Empty(t *testing.T) {
	got := ResolveEffectiveRole([]Membership{})
	if got != "" {
		t.Errorf("expected empty string for empty input, got %q", got)
	}
}

func TestResolveEffectiveRole_Single(t *testing.T) {
	got := ResolveEffectiveRole([]Membership{
		{UserID: "u1", Scope: "s1", Role: "Viewer"},
	})
	if got != "Viewer" {
		t.Errorf("expected 'Viewer', got %q", got)
	}
}

func TestResolveEffectiveRole_UnknownRole(t *testing.T) {
	// @hlv max_role_wins
	memberships := []Membership{
		{UserID: "u1", Scope: "s1", Role: "UnknownRole"},
		{UserID: "u1", Scope: "s1", Role: "Editor"},
	}
	got := ResolveEffectiveRole(memberships)
	if got != "Editor" {
		t.Errorf("expected 'Editor', got %q", got)
	}
}
