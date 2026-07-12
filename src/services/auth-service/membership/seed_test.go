package membership

import "testing"

func TestSeedMemberships_ReturnsAtLeast3(t *testing.T) {
	// @hlv seed_membership_completeness
	memberships := SeedMemberships()
	if len(memberships) < 3 {
		t.Fatalf("expected at least 3 memberships, got %d", len(memberships))
	}
	for _, m := range memberships {
		if m.UserID == "" {
			t.Error("found membership with empty UserID")
		}
		if m.Scope == "" {
			t.Error("found membership with empty Scope")
		}
		if m.Role == "" {
			t.Error("found membership with empty Role")
		}
	}
}

func TestSeedVisibility_CoversAllTypes(t *testing.T) {
	vis := SeedVisibility()
	types := make(map[Visibility]bool)
	for _, v := range vis {
		types[v] = true
	}
	if !types[VisibilityPrivate] {
		t.Error("missing Private visibility")
	}
	if !types[VisibilityInternal] {
		t.Error("missing Internal visibility")
	}
	if !types[VisibilityPublic] {
		t.Error("missing Public visibility")
	}
}

func TestSeedPolicies_ReturnsAtLeastOne(t *testing.T) {
	// @hlv seed_membership_completeness
	policies := SeedPolicies()
	if len(policies) < 1 {
		t.Fatal("expected at least 1 policy")
	}
	if policies[0].Pattern == nil {
		t.Error("policy must have a pattern")
	}
	if policies[0].Right == "" {
		t.Error("policy must have a right")
	}
	if policies[0].Scope == "" {
		t.Error("policy must have a scope")
	}
}
