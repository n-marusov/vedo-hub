package org

// Validates: ORG-ACCESS-001
//
// Proto definition validation tests.
// Validates that org.proto message fields and scope parsing work correctly.
// These tests validate the proto contracts at rest (before code generation).
// Validates: REQ-NFR.SECURITY.organization-access-model

import (
	"testing"
)

// TestParseScope_ValidInput validates that ParseScope correctly handles
// valid "group/" and "project/" scope strings.
func TestParseScope_ValidInput_ReturnsTypeAndID(t *testing.T) {
	tests := []struct {
		input    string
		wantType ScopeType
		wantID   string
	}{
		{"group/my-group", ScopeGroup, "my-group"},
		{"project/proj-1", ScopeProject, "proj-1"},
		{"group/Root/TeamA", ScopeGroup, "Root/TeamA"},
	}
	for _, tc := range tests {
		st, id, err := ParseScope(tc.input)
		if err != nil {
			t.Errorf("ParseScope(%q): unexpected error: %v", tc.input, err)
			continue
		}
		if st != tc.wantType {
			t.Errorf("ParseScope(%q): type = %q, want %q", tc.input, st, tc.wantType)
		}
		if id != tc.wantID {
			t.Errorf("ParseScope(%q): id = %q, want %q", tc.input, id, tc.wantID)
		}
	}
}

// TestParseScope_InvalidFormat validates that malformed scope strings return errors.
func TestParseScope_InvalidFormat_ReturnsError(t *testing.T) {
	invalid := []string{
		"",               // empty
		"no-slash",       // no separator
		"/leading-slash", // starts with slash
		"group",          // no id part
		"only/",          // valid type format but empty id
	}
	for _, input := range invalid {
		_, _, err := ParseScope(input)
		if err == nil {
			t.Errorf("ParseScope(%q): expected error, got nil", input)
		}
	}
}

// TestParseScope_UnknownType validates that unknown scope types are rejected.
// "ontology/" was a legacy alias that was removed in Task 7.4 — it is now rejected.
func TestParseScope_UnknownType_ReturnsError(t *testing.T) {
	unknown := []string{
		"invalid-type/foo",
		"user/user-123",
		"ontology/ont-123", // legacy alias removed in Task 7.4
		"team/alpha",
	}
	for _, input := range unknown {
		_, _, err := ParseScope(input)
		if err == nil {
			t.Errorf("ParseScope(%q): expected error for unknown type, got nil", input)
		}
	}
}
