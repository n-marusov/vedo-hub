package org

// PostgreSQL schema contract tests.
// These tests validate the expected behavior of the PostgreSQL-backed OrgStore.
// They will FAIL (red phase) until the migration SQL and PostgresOrgStore are implemented.
//
// Test setup requires a running PostgreSQL instance with:
//   DATABASE_URL=postgres://postgres:password@localhost:5432/vedo_org_test?sslmode=disable

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// getTestDB returns a connection to the test PostgreSQL database.
// Skips the test if database is not available.
func getTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/vedo_org_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("failed to open test database connection: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("test database not available: %v", err)
	}
	return db
}

// TestScopeInsert_UpsertScope_PersistsCorrectly validates that a scope
// can be inserted and read back with all fields matching.
func TestScopeInsert_UpsertScope_PersistsCorrectly(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	_, err := db.Exec(`INSERT INTO scopes (id, name, description, type, visibility, tenant_id)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		"group/test-group", "Test Group", "A test group", "group", "Private", "org-001")
	if err != nil {
		t.Fatalf("failed to insert scope: %v", err)
	}
	defer db.Exec("DELETE FROM scopes WHERE id = $1", "group/test-group")

	var id, name, desc, stype, vis, tenant string
	err = db.QueryRow(`SELECT id, name, description, type, visibility, tenant_id FROM scopes WHERE id = $1`,
		"group/test-group").Scan(&id, &name, &desc, &stype, &vis, &tenant)
	if err != nil {
		t.Fatalf("failed to read scope: %v", err)
	}
	if id != "group/test-group" || name != "Test Group" || stype != "group" || vis != "Private" {
		t.Fatalf("scope fields mismatch: got %s/%s/%s/%s", id, name, stype, vis)
	}
}

// TestDuplicateMembership_UpsertMembership_UpsertsNotInserts validates
// that inserting the same user/scope pair twice acts as an upsert.
func TestDuplicateMembership_UpsertMembership_UpsertsNotInserts(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	db.Exec(`INSERT INTO scopes (id, type) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		"ontology/dup-test", "ontology")
	defer db.Exec("DELETE FROM scopes WHERE id = $1", "ontology/dup-test")

	_, err := db.Exec(`INSERT INTO memberships (scope, user_id, role) VALUES ($1, $2, $3)`,
		"ontology/dup-test", "dup-user", "Editor")
	if err != nil {
		t.Fatalf("first insert failed: %v", err)
	}
	defer db.Exec("DELETE FROM memberships WHERE scope = $1 AND user_id = $2", "ontology/dup-test", "dup-user")

	_, err = db.Exec(`INSERT INTO memberships (scope, user_id, role) VALUES ($1, $2, $3) ON CONFLICT (scope, user_id) DO UPDATE SET role = $3`,
		"ontology/dup-test", "dup-user", "Owner")
	if err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM memberships WHERE scope = $1 AND user_id = $2`,
		"ontology/dup-test", "dup-user").Scan(&count)
	if err != nil || count != 1 {
		t.Fatalf("expected 1 row, got %d (err: %v)", count, err)
	}

	var role string
	err = db.QueryRow(`SELECT role FROM memberships WHERE scope = $1 AND user_id = $2`,
		"ontology/dup-test", "dup-user").Scan(&role)
	if err != nil || role != "Owner" {
		t.Fatalf("expected role 'Owner', got %q", role)
	}
}

// TestParentFK_UpsertScope_InvalidParent_ReturnsFKError validates that
// inserting a scope with a non-existent parent_id returns a foreign key error.
func TestParentFK_UpsertScope_InvalidParent_ReturnsFKError(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	_, err := db.Exec(`INSERT INTO scopes (id, type, parent_id) VALUES ($1, $2, $3)`,
		"group/orphan", "group", "group/nonexistent")
	defer db.Exec("DELETE FROM scopes WHERE id = $1", "group/orphan")

	if err == nil {
		t.Fatal("expected foreign key error for invalid parent_id, got nil")
	}
}

// TestRoleCheck_UpsertMembership_InvalidRole_ReturnsConstraintError validates
// that inserting a membership with an invalid role value is rejected.
func TestRoleCheck_UpsertMembership_InvalidRole_ReturnsConstraintError(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	db.Exec(`INSERT INTO scopes (id, type) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		"ontology/role-test", "ontology")
	defer db.Exec("DELETE FROM scopes WHERE id = $1", "ontology/role-test")

	_, err := db.Exec(`INSERT INTO memberships (scope, user_id, role) VALUES ($1, $2, $3)`,
		"ontology/role-test", "role-user", "SuperAdmin")
	if err == nil {
		t.Fatal("expected CHECK constraint error for invalid role, got nil")
	}
}

// TestVisibilityEnum_UpsertScope_InvalidVisibility_ReturnsConstraintError validates
// that inserting a scope with an invalid visibility value is rejected.
func TestVisibilityEnum_UpsertScope_InvalidVisibility_ReturnsConstraintError(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	_, err := db.Exec(`INSERT INTO scopes (id, type, visibility) VALUES ($1, $2, $3)`,
		"ontology/vis-test", "ontology", "Secret")
	defer db.Exec("DELETE FROM scopes WHERE id = $1", "ontology/vis-test")

	if err == nil {
		t.Fatal("expected CHECK constraint error for invalid visibility, got nil")
	}
}

// TestHierarchyWalk_GetEffectiveMemberships_CollectsAllAncestors validates
// that querying memberships for a leaf scope returns memberships from all ancestors.
func TestHierarchyWalk_GetEffectiveMemberships_CollectsAllAncestors(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	scopes := []struct {
		id       string
		typ      string
		parentID string
	}{
		{"group/root", "group", ""},
		{"group/root/team", "group", "group/root"},
		{"group/root/team/sub", "group", "group/root/team"},
	}
	for _, s := range scopes {
		var parent *string
		if s.parentID != "" {
			parent = &s.parentID
		}
		db.Exec(`INSERT INTO scopes (id, type, parent_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
			s.id, s.typ, parent)
	}
	defer func() {
		for _, s := range scopes {
			db.Exec("DELETE FROM memberships WHERE scope = $1", s.id)
			db.Exec("DELETE FROM scopes WHERE id = $1", s.id)
		}
	}()

	_, err := db.Exec(`INSERT INTO memberships (scope, user_id, role) VALUES ($1, $2, $3)`,
		"group/root", "hierarchy-user", "Editor")
	if err != nil {
		t.Fatalf("failed to insert membership: %v", err)
	}
	defer db.Exec("DELETE FROM memberships WHERE scope = $1 AND user_id = $2",
		"group/root", "hierarchy-user")

	rows, err := db.Query(`
		WITH RECURSIVE scope_tree AS (
			SELECT id, parent_id FROM scopes WHERE id = $1
			UNION ALL
			SELECT s.id, s.parent_id FROM scopes s
			INNER JOIN scope_tree st ON s.id = st.parent_id
		)
		SELECT m.scope, m.user_id, m.role FROM memberships m
		INNER JOIN scope_tree st ON m.scope = st.id
		WHERE m.user_id = $2`, "group/root/team/sub", "hierarchy-user")
	if err != nil {
		t.Fatalf("recursive CTE query failed: %v", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		count++
	}
	if count < 1 {
		t.Fatal("expected at least 1 inherited membership via hierarchy walk")
	}
}
