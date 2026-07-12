package membership

import "testing"

func TestMembershipStruct(t *testing.T) {
	// @hlv seed_membership_completeness
	m := Membership{
		UserID:    "test-user",
		Scope:     "group/test",
		Role:      "Editor",
		Inherited: false,
	}
	if m.UserID != "test-user" {
		t.Errorf("expected UserID 'test-user', got %q", m.UserID)
	}
	if m.Scope != "group/test" {
		t.Errorf("expected Scope 'group/test', got %q", m.Scope)
	}
	if m.Role != "Editor" {
		t.Errorf("expected Role 'Editor', got %q", m.Role)
	}
	if m.Inherited != false {
		t.Errorf("expected Inherited false, got %v", m.Inherited)
	}
}

func TestVisibilityValues(t *testing.T) {
	if VisibilityPrivate != "Private" {
		t.Errorf("expected Private, got %q", VisibilityPrivate)
	}
	if VisibilityInternal != "Internal" {
		t.Errorf("expected Internal, got %q", VisibilityInternal)
	}
	if VisibilityPublic != "Public" {
		t.Errorf("expected Public, got %q", VisibilityPublic)
	}
}
