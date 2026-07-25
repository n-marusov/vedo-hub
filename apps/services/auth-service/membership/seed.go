package membership

// @hlv seed_membership_completeness
// @ctx: contracts US-P0-001, US-P1-001 — seed data for development and test environments

// SeedMemberships returns initial membership fixtures for development and testing.
// @hlv:sec config — These fixtures define default authorization boundaries; must not be used in production without review.
func SeedMemberships() []Membership {
	return []Membership{
		{UserID: "alice", Scope: "group/Root", Role: "Owner", Inherited: false},
		{UserID: "bob", Scope: "group/Root/TeamA", Role: "Editor", Inherited: true},
		{UserID: "alice", Scope: "ontology/ont-123", Role: "Owner", Inherited: false},
		{UserID: "bob", Scope: "ontology/ont-123", Role: "Editor", Inherited: false},
		{UserID: "eve", Scope: "ontology/ont-123", Role: "Viewer", Inherited: false},
	}
}

// SeedVisibility returns visibility fixtures for development and testing.
func SeedVisibility() map[string]Visibility {
	return map[string]Visibility{
		"ont-123": VisibilityPrivate,
		"ont-456": VisibilityInternal,
		"ont-789": VisibilityPublic,
	}
}

// SeedPolicies returns initial attribute policy fixtures for development and testing.
// @hlv:sec config — Attribute policies control attribute-level access.
func SeedPolicies() []AttributePolicy {
	return []AttributePolicy{
		{
			Pattern: map[string]string{"type": "uri_prefix", "value": "ex:confidential/"},
			Right:   "read",
			Scope:   "ont-123",
		},
	}
}
