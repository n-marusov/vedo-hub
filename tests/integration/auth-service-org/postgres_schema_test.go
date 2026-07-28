//go:build integration

// Package authserviceorg_test contains integration tests for the auth-service
// organization schema contract (PostgreSQL). These tests validate the
// PostgreSQL schema invariants introduced by migrations 001-009.
//
// Moved from apps/services/auth-service/org/*_test.go — these tests were
// misclassified as unit tests but require a running PostgreSQL instance.
//
// Test setup requires a running PostgreSQL instance with:
//
//	DATABASE_URL=postgres://postgres:password@localhost:5432/vedo_org_test?sslmode=disable
//
// Run with:
//
//	go test -tags=integration -v .
//
// Validates: REQ-NFR.SECURITY.organization-access-model
// Validates: ORG-ACCESS-001
package authserviceorg_test

import (
	"database/sql"
	"log/slog"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

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

// schemaTestDB is an alias of getTestDB for self-documenting schema tests.
func schemaTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return getTestDB(t)
}

// cleanupScope removes a test scope and any cascaded rows.
func cleanupScope(t *testing.T, db *sql.DB, scopeID string) {
	t.Helper()
	_, _ = db.Exec("DELETE FROM scopes WHERE id = $1", scopeID)
}

// ─── PostgreSQL Schema Contract Tests ─────────────────────────────────────────
// (from postgres_store_test.go)

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

// ─── Project ↔ Ontology 1:1 Schema Tests ─────────────────────────────────────
// (from schema_test.go)

// TestSchema_ProjectCreated_HasPairedOntologyRow_OntologyIdReturned validates
// that a project scope with a paired ontologies row can be created and read
// back with the correct ontology_id.
func TestSchema_ProjectCreated_HasPairedOntologyRow_OntologyIdReturned(t *testing.T) {
	db := schemaTestDB(t)
	defer db.Close()

	scopeID := "project/schema-test-paired"
	ontologyID := "ont-schema-test-paired-uuid"
	defer cleanupScope(t, db, scopeID)

	_, err := db.Exec(`
		INSERT INTO scopes (id, type, visibility, tenant_id, name, description)
		VALUES ($1, 'project', 'Private', '', 'Schema Test Paired', '')
	`, scopeID)
	if err != nil {
		t.Fatalf("failed to insert project scope: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO ontologies (project_scope, ontology_id)
		VALUES ($1, $2)
	`, scopeID, ontologyID)
	if err != nil {
		t.Fatalf("failed to insert paired ontologies row: %v", err)
	}

	var gotOntologyID string
	err = db.QueryRow(`
		SELECT ontology_id FROM ontologies WHERE project_scope = $1
	`, scopeID).Scan(&gotOntologyID)
	if err != nil {
		t.Fatalf("failed to read paired ontologies row: %v", err)
	}
	if gotOntologyID != ontologyID {
		t.Fatalf("ontology_id mismatch: got %s, want %s", gotOntologyID, ontologyID)
	}

	slog.Info("schema.test", "case", "ProjectCreated_HasPairedOntologyRow_OntologyIdReturned", "scope", scopeID, "ontology_id", gotOntologyID)
}

// TestSchema_ProjectDeleted_CascadesToOntology_OntologyRowGone validates that
// deleting a project scope cascades to the paired ontologies row.
func TestSchema_ProjectDeleted_CascadesToOntology_OntologyRowGone(t *testing.T) {
	db := schemaTestDB(t)
	defer db.Close()

	scopeID := "project/schema-test-cascade"
	ontologyID := "ont-schema-test-cascade-uuid"
	defer cleanupScope(t, db, scopeID)

	_, err := db.Exec(`
		INSERT INTO scopes (id, type, visibility, tenant_id, name, description)
		VALUES ($1, 'project', 'Private', '', 'Schema Test Cascade', '')
	`, scopeID)
	if err != nil {
		t.Fatalf("failed to insert project scope: %v", err)
	}
	_, err = db.Exec(`
		INSERT INTO ontologies (project_scope, ontology_id)
		VALUES ($1, $2)
	`, scopeID, ontologyID)
	if err != nil {
		t.Fatalf("failed to insert paired ontologies row: %v", err)
	}

	_, err = db.Exec("DELETE FROM scopes WHERE id = $1", scopeID)
	if err != nil {
		t.Fatalf("failed to delete project scope: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM ontologies WHERE project_scope = $1", scopeID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query ontologies after cascade: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 ontologies rows after cascade delete, got %d", count)
	}

	slog.Info("schema.test", "case", "ProjectDeleted_CascadesToOntology_OntologyRowGone", "scope", scopeID)
}

// TestSchema_OntologyCreated_Standalone_Rejected validates the 1:1 invariant:
// an ontologies row cannot exist without a matching project scope.
func TestSchema_OntologyCreated_Standalone_Rejected(t *testing.T) {
	db := schemaTestDB(t)
	defer db.Close()

	nonExistentScope := "project/does-not-exist-schema-test"
	defer cleanupScope(t, db, nonExistentScope)

	_, err := db.Exec(`
		INSERT INTO ontologies (project_scope, ontology_id)
		VALUES ($1, $2)
	`, nonExistentScope, "ont-standalone-rejected-uuid")
	if err == nil {
		_, _ = db.Exec("DELETE FROM ontologies WHERE project_scope = $1", nonExistentScope)
		t.Fatal("expected FK violation when inserting ontologies row without matching scope, got nil")
	}

	slog.Info("schema.test", "case", "OntologyCreated_Standalone_Rejected", "scope", nonExistentScope, "err", err.Error())
}

// TestSchema_LegacyOntologyScope_Migrated_BecomesProjectScopeWithTypeRename
// validates the migration sequence: a legacy type='ontology' scope is
// backfilled into the ontologies table and then renamed to type='project'.
func TestSchema_LegacyOntologyScope_Migrated_BecomesProjectScopeWithTypeRename(t *testing.T) {
	db := schemaTestDB(t)
	defer db.Close()

	scopeID := "ontology/schema-test-legacy-migration"
	defer cleanupScope(t, db, scopeID)

	db.Exec("ALTER TABLE scopes DROP CONSTRAINT IF EXISTS scopes_type_check")
	_, err := db.Exec(`
		INSERT INTO scopes (id, type, visibility, tenant_id, name, description)
		VALUES ($1, 'ontology', 'Private', '', 'Schema Test Legacy', '')
	`, scopeID)
	if err != nil {
		t.Fatalf("failed to insert legacy ontology scope: %v", err)
	}
	db.Exec(`ALTER TABLE scopes ADD CONSTRAINT IF NOT EXISTS scopes_type_check
		CHECK (type IN ('group', 'ontology', 'project'))`)

	_, err = db.Exec(`
		INSERT INTO ontologies (project_scope, ontology_id)
		SELECT id, id FROM scopes
		WHERE id = $1 AND type IN ('ontology', 'project')
		ON CONFLICT (project_scope) DO NOTHING
	`, scopeID)
	if err != nil {
		t.Fatalf("failed to backfill ontologies from legacy scope: %v", err)
	}

	_, err = db.Exec("UPDATE scopes SET type = 'project', updated_at = now() WHERE id = $1", scopeID)
	if err != nil {
		t.Fatalf("failed to rename scope type to project: %v", err)
	}

	var scopeType string
	err = db.QueryRow("SELECT type FROM scopes WHERE id = $1", scopeID).Scan(&scopeType)
	if err != nil {
		t.Fatalf("failed to read scope type after rename: %v", err)
	}
	if scopeType != "project" {
		t.Fatalf("scope type mismatch after migration: got %s, want project", scopeType)
	}

	var gotOntologyID string
	err = db.QueryRow("SELECT ontology_id FROM ontologies WHERE project_scope = $1", scopeID).Scan(&gotOntologyID)
	if err != nil {
		t.Fatalf("failed to read paired ontologies row after migration: %v", err)
	}
	if gotOntologyID != scopeID {
		t.Fatalf("ontology_id mismatch: got %s, want %s (legacy identity)", gotOntologyID, scopeID)
	}

	slog.Info("schema.test", "case", "LegacyOntologyScope_Migrated_BecomesProjectScopeWithTypeRename", "scope", scopeID, "type", scopeType)
}
