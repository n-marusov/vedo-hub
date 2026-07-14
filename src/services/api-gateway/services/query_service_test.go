package services

import (
	"testing"
)

func TestValidateAndSanitizeCYPHER_ClassNamedCreateResource_Allowed(t *testing.T) {
	// Regression: the naive strings.Contains check falsely flagged this
	// MATCH query because "CREATE" appears inside the identifier
	// `CreateResource`. Whole-word matching must allow it.
	result := ValidateAndSanitizeCYPHER("MATCH (n:CreateResource) RETURN n", 100)
	if !result.Valid {
		t.Errorf("expected valid, got error %s: %s", result.ErrorCode, result.ErrorMessage)
	}
}

func TestValidateAndSanitizeCYPHER_MatchDeleteMeNode_Allowed(t *testing.T) {
	// Another regression: `DeleteMe` is a label name, not a mutation.
	result := ValidateAndSanitizeCYPHER("MATCH (n:DeleteMe) RETURN n LIMIT 5", 100)
	if !result.Valid {
		t.Errorf("expected valid, got error %s: %s", result.ErrorCode, result.ErrorMessage)
	}
}

func TestValidateAndSanitizeCYPHER_InsertData_Rejected(t *testing.T) {
	result := ValidateAndSanitizeCYPHER("INSERT DATA { value: 1 }", 100)
	if result.Valid {
		t.Fatal("expected INSERT to be rejected")
	}
	if result.ErrorCode != "GATEWAY-QUERY-READONLY" {
		t.Errorf("expected GATEWAY-QUERY-READONLY, got %s", result.ErrorCode)
	}
}

func TestValidateAndSanitizeCYPHER_CreateNode_Rejected(t *testing.T) {
	result := ValidateAndSanitizeCYPHER("CREATE (n:Foo {name: 'bar'})", 100)
	if result.Valid {
		t.Fatal("expected CREATE to be rejected")
	}
}

func TestValidateAndSanitizeCYPHER_DeleteNode_Rejected(t *testing.T) {
	result := ValidateAndSanitizeCYPHER("MATCH (n:Foo) DELETE n", 100)
	if result.Valid {
		t.Fatal("expected DELETE to be rejected")
	}
}

func TestValidateAndSanitizeCYPHER_SetProperty_Rejected(t *testing.T) {
	result := ValidateAndSanitizeCYPHER("MATCH (n:Foo) SET n.name = 'bar'", 100)
	if result.Valid {
		t.Fatal("expected SET to be rejected")
	}
}

func TestValidateAndSanitizeCYPHER_MergeNode_Rejected(t *testing.T) {
	result := ValidateAndSanitizeCYPHER("MERGE (n:Foo {name: 'bar'})", 100)
	if result.Valid {
		t.Fatal("expected MERGE to be rejected")
	}
}

func TestValidateAndSanitizeSPARQL_InsertData_Rejected(t *testing.T) {
	result := ValidateAndSanitizeSPARQL("INSERT DATA { <a> <b> <c> }", 100)
	if result.Valid {
		t.Fatal("expected INSERT to be rejected")
	}
}

func TestValidateAndSanitizeSPARQL_DeleteData_Rejected(t *testing.T) {
	result := ValidateAndSanitizeSPARQL("DELETE WHERE { ?s ?p ?o }", 100)
	if result.Valid {
		t.Fatal("expected DELETE to be rejected")
	}
}

func TestValidateAndSanitizeSPARQL_SelectQuery_Allowed(t *testing.T) {
	result := ValidateAndSanitizeSPARQL("SELECT ?s ?p ?o WHERE { ?s ?p ?o }", 100)
	if !result.Valid {
		t.Errorf("expected valid SELECT, got error %s: %s", result.ErrorCode, result.ErrorMessage)
	}
}

func TestContainsKeyword_WholeWordBoundaries(t *testing.T) {
	tests := []struct {
		s        string
		keyword  string
		expected bool
	}{
		{"INSERT DATA", "INSERT", true},
		{"(CREATE)", "CREATE", true},
		{"DELETE n", "DELETE", true},
		// Identifiers glued to keyword — must NOT match.
		{"CREATERESOURCE", "CREATE", false},
		{"CreateResource", "CREATE", false},
		{"DELETEME", "DELETE", false},
		{"mySet", "SET", false},
		// Case-insensitive.
		{"insert data", "INSERT", true},
		{"create (n)", "CREATE", true},
	}
	for _, tc := range tests {
		got := containsKeyword(tc.s, tc.keyword)
		if got != tc.expected {
			t.Errorf("containsKeyword(%q, %q) = %v, want %v", tc.s, tc.keyword, got, tc.expected)
		}
	}
}
