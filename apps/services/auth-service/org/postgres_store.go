package org

// PostgreSQL-backed OrgStore implementation.
// @hlv store_in_postgres
// @hlv:sec [AUTH_BOUNDARY] — Store operations maintain authorization invariants.

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq" // pq driver registered via blank import for database/sql
)

// PostgresOrgStore implements OrgStore using PostgreSQL.
type PostgresOrgStore struct {
	db *sql.DB
}

// NewPostgresOrgStore creates a new PostgresOrgStore with connection pool.
// The databaseURL is read from env var (default: POSTGRES_DATABASE_URL).
// Automatically runs migrations on startup.
func NewPostgresOrgStore(ctx interface{}, databaseURL string) (*PostgresOrgStore, error) {
	if databaseURL == "" {
		databaseURL = os.Getenv("POSTGRES_DATABASE_URL")
	}
	if databaseURL == "" {
		databaseURL = "postgres://postgres:password@localhost:5432/vedo_org?sslmode=disable"
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	store := &PostgresOrgStore{db: db}

	// Run migrations
	if err := store.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Printf(`{"event":"store.initialized","type":"postgres","pool":"max25_idle5","migration":"applied"}`)
	return store, nil
}

// runMigrations applies the migration SQL files in order.
// The list is explicit (not alphabetical) so the Project ↔ Ontology separation
// migrations can run in the logically correct order: 008 (create ontologies
// table) → 009 (backfill from legacy scopes) → 007 (rename type='ontology'
// to type='project'). Renaming 007 before backfilling would lose the legacy
// type='ontology' rows that 009 reads.
func (p *PostgresOrgStore) runMigrations() error {
	migrations := []string{
		"001_create_scopes.sql",
		"002_create_memberships.sql",
		"003_create_policies.sql",
		"004_create_audit_events.sql",
		"005_add_constraints.sql",
		"006_add_mvp_roles.sql",
		// Project ↔ Ontology separation (Phase 4). Order matters:
		//   008 — create ontologies table
		//   009 — backfill ontologies from legacy type='ontology' scopes
		//   007 — rename scopes.type='ontology' → 'project'
		"008_create_ontologies_table.sql",
		"009_seed_ontologies_from_legacy.sql",
		"007_rename_ontology_scope_to_project.sql",
		// Fork infrastructure — upstream tracking
		"010_add_upstream_project_id.sql",
	}

	for _, m := range migrations {
		path := fmt.Sprintf("migrations/%s", m)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			// Try alternate path relative to module root
			path = fmt.Sprintf("src/services/auth-service/migrations/%s", m)
			sqlBytes, err = os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read migration %s: %w", m, err)
			}
		}

		if _, err := p.db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("migration %s failed: %w", m, err)
		}
		log.Printf(`{"event":"migration.applied","migration":"%s"}`, m)
	}

	return nil
}

// Close closes the database connection pool.
func (p *PostgresOrgStore) Close() error {
	return p.db.Close()
}

// ============================================================================
// Scope Operations
// ============================================================================

func (p *PostgresOrgStore) UpsertScope(s ScopeNode) error {
	_, err := p.db.Exec(`
		INSERT INTO scopes (id, type, parent_id, visibility, tenant_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
		ON CONFLICT (id) DO UPDATE SET
			type = EXCLUDED.type,
			parent_id = EXCLUDED.parent_id,
			visibility = EXCLUDED.visibility,
			tenant_id = EXCLUDED.tenant_id,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			updated_at = now()
	`, s.ID, string(s.Type), nullString(s.ParentID), string(s.Visibility), s.TenantID, s.Name, s.Description)
	if err != nil {
		log.Printf(`{"event":"store.error","operation":"UpsertScope","scope":"%s","error":"%v"}`, s.ID, err)
		return err
	}
	return nil
}

func (p *PostgresOrgStore) GetScope(id string) (*ScopeNode, error) {
	var s ScopeNode
	var parentID, name, desc sql.NullString
	err := p.db.QueryRow(`
		SELECT id, type, parent_id, visibility, tenant_id, name, description
		FROM scopes WHERE id = $1
	`, id).Scan(&s.ID, &s.Type, &parentID, &s.Visibility, &s.TenantID, &name, &desc)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Printf(`{"event":"store.error","operation":"GetScope","scope":"%s","error":"%v"}`, id, err)
		return nil, err
	}
	if parentID.Valid {
		s.ParentID = parentID.String
	}
	if name.Valid {
		s.Name = name.String
	}
	if desc.Valid {
		s.Description = desc.String
	}
	return &s, nil
}

func (p *PostgresOrgStore) DeleteScope(id string) error {
	result, err := p.db.Exec(`DELETE FROM scopes WHERE id = $1`, id)
	if err != nil {
		log.Printf(`{"event":"store.error","operation":"DeleteScope","scope":"%s","error":"%v"}`, id, err)
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("scope not found: %s", id)
	}
	return nil
}

func (p *PostgresOrgStore) ListChildScopes(parentID string) ([]ScopeNode, error) {
	rows, err := p.db.Query(`
		SELECT id, type, parent_id, visibility, tenant_id, name, description
		FROM scopes WHERE parent_id = $1
	`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scopes []ScopeNode
	for rows.Next() {
		var s ScopeNode
		var pid, name, desc sql.NullString
		if err := rows.Scan(&s.ID, &s.Type, &pid, &s.Visibility, &s.TenantID, &name, &desc); err != nil {
			return nil, err
		}
		if pid.Valid {
			s.ParentID = pid.String
		}
		if name.Valid {
			s.Name = name.String
		}
		if desc.Valid {
			s.Description = desc.String
		}
		scopes = append(scopes, s)
	}
	return scopes, nil
}

func (p *PostgresOrgStore) ListAllScopes() ([]ScopeNode, error) {
	rows, err := p.db.Query(`
		SELECT id, type, parent_id, visibility, tenant_id, name, description
		FROM scopes ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scopes []ScopeNode
	for rows.Next() {
		var s ScopeNode
		var pid, name, desc sql.NullString
		if err := rows.Scan(&s.ID, &s.Type, &pid, &s.Visibility, &s.TenantID, &name, &desc); err != nil {
			return nil, err
		}
		if pid.Valid {
			s.ParentID = pid.String
		}
		if name.Valid {
			s.Name = name.String
		}
		if desc.Valid {
			s.Description = desc.String
		}
		scopes = append(scopes, s)
	}
	return scopes, nil
}

// ============================================================================
// Membership Operations
// ============================================================================

func (p *PostgresOrgStore) UpsertMembership(m OrgMembership) error {
	_, err := p.db.Exec(`
		INSERT INTO memberships (scope, user_id, role, inherited, created_at, updated_at)
		VALUES ($1, $2, $3, $4, now(), now())
		ON CONFLICT (scope, user_id) DO UPDATE SET
			role = EXCLUDED.role,
			inherited = EXCLUDED.inherited,
			updated_at = now()
	`, m.Scope, m.UserID, m.Role, m.Inherited)
	if err != nil {
		log.Printf(`{"event":"store.error","operation":"UpsertMembership","scope":"%s","user":"%s","error":"%v"}`, m.Scope, m.UserID, err)
		return err
	}
	return nil
}

func (p *PostgresOrgStore) DeleteMembership(scope, userID string) error {
	result, err := p.db.Exec(`DELETE FROM memberships WHERE scope = $1 AND user_id = $2`, scope, userID)
	if err != nil {
		log.Printf(`{"event":"store.error","operation":"DeleteMembership","scope":"%s","user":"%s","error":"%v"}`, scope, userID, err)
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("membership not found: %s/%s", scope, userID)
	}
	return nil
}

func (p *PostgresOrgStore) GetMemberships(scope string) ([]OrgMembership, error) {
	rows, err := p.db.Query(`
		SELECT scope, user_id, role, inherited FROM memberships WHERE scope = $1
	`, scope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mems []OrgMembership
	for rows.Next() {
		var m OrgMembership
		if err := rows.Scan(&m.Scope, &m.UserID, &m.Role, &m.Inherited); err != nil {
			return nil, err
		}
		mems = append(mems, m)
	}
	return mems, nil
}

func (p *PostgresOrgStore) GetUserMemberships(userID string) ([]OrgMembership, error) {
	rows, err := p.db.Query(`
		SELECT scope, user_id, role, inherited FROM memberships WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mems []OrgMembership
	for rows.Next() {
		var m OrgMembership
		if err := rows.Scan(&m.Scope, &m.UserID, &m.Role, &m.Inherited); err != nil {
			return nil, err
		}
		mems = append(mems, m)
	}
	return mems, nil
}

// GetEffectiveMemberships walks the hierarchy via recursive CTE to collect
// all memberships for a user at a given scope, including inherited ones.
func (p *PostgresOrgStore) GetEffectiveMemberships(userID, scope string) ([]OrgMembership, error) {
	rows, err := p.db.Query(`
		WITH RECURSIVE scope_tree AS (
			SELECT id, parent_id, 0 AS depth FROM scopes WHERE id = $1
			UNION ALL
			SELECT s.id, s.parent_id, st.depth + 1
			FROM scopes s
			INNER JOIN scope_tree st ON s.id = st.parent_id
		)
		SELECT m.scope, m.user_id, m.role, (m.scope != $1) AS inherited
		FROM memberships m
		INNER JOIN scope_tree st ON m.scope = st.id
		WHERE m.user_id = $2
	`, scope, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mems []OrgMembership
	for rows.Next() {
		var m OrgMembership
		if err := rows.Scan(&m.Scope, &m.UserID, &m.Role, &m.Inherited); err != nil {
			return nil, err
		}
		mems = append(mems, m)
	}
	return mems, nil
}

// ============================================================================
// Policy Operations
// ============================================================================

func (p *PostgresOrgStore) UpsertPolicy(pol AttributePolicy) error {
	_, err := p.db.Exec(`
		INSERT INTO attribute_policies (scope, pattern, "right", created_at)
		VALUES ($1, $2, $3, now())
	`, pol.Scope, pol.Pattern, pol.Right)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			log.Printf(`{"event":"store.warn","operation":"UpsertPolicy","scope":"%s","reason":"duplicate_pattern"}`, pol.Scope)
		}
		return err
	}
	return nil
}

func (p *PostgresOrgStore) DeletePolicy(scope string, policyID string) error {
	result, err := p.db.Exec(`DELETE FROM attribute_policies WHERE scope = $1 AND id::text = $2`, scope, policyID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("policy not found in scope %s", scope)
	}
	return nil
}

func (p *PostgresOrgStore) GetPolicies(scope string) ([]AttributePolicy, error) {
	rows, err := p.db.Query(`
		SELECT scope, pattern, "right" FROM attribute_policies WHERE scope = $1
	`, scope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pols []AttributePolicy
	for rows.Next() {
		var pol AttributePolicy
		if err := rows.Scan(&pol.Scope, &pol.Pattern, &pol.Right); err != nil {
			return nil, err
		}
		pols = append(pols, pol)
	}
	return pols, nil
}

func (p *PostgresOrgStore) GetAllPolicies() ([]AttributePolicy, error) {
	rows, err := p.db.Query(`SELECT scope, pattern, "right" FROM attribute_policies`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pols []AttributePolicy
	for rows.Next() {
		var pol AttributePolicy
		if err := rows.Scan(&pol.Scope, &pol.Pattern, &pol.Right); err != nil {
			return nil, err
		}
		pols = append(pols, pol)
	}
	return pols, nil
}

// ============================================================================
// Visibility Operations
// ============================================================================

func (p *PostgresOrgStore) SetVisibility(scope string, v Visibility) error {
	_, err := p.db.Exec(`UPDATE scopes SET visibility = $1, updated_at = now() WHERE id = $2`, string(v), scope)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresOrgStore) GetVisibility(scope string) (Visibility, error) {
	var v Visibility
	err := p.db.QueryRow(`SELECT visibility FROM scopes WHERE id = $1`, scope).Scan(&v)
	if err == sql.ErrNoRows {
		return VisibilityPrivate, nil
	}
	if err != nil {
		return VisibilityPrivate, err
	}
	return v, nil
}

// InvalidateCache is a no-op for PostgresOrgStore (cache is managed by OrgService).
func (p *PostgresOrgStore) InvalidateCache(scope string) {
	// Cache is managed by OrgService layer, not the store.
}

// ============================================================================
// Ontology Operations (1:1 with project scopes)
// ============================================================================

// CreateOntology inserts a 1:1 paired ontology row for a project scope.
// The project scope must already exist in scopes (the FK constraint enforces this).
// Returns an error if the scope does not exist or if the ontology_id is already taken.
func (p *PostgresOrgStore) CreateOntology(ont Ontology) error {
	_, err := p.db.Exec(`
		INSERT INTO ontologies (project_scope, ontology_id, iri)
		VALUES ($1, $2, $3)
	`, ont.ProjectScope, ont.OntologyID, ont.IRI)
	if err != nil {
		log.Printf(`{"event":"store.error","operation":"CreateOntology","project_scope":"%s","error":"%v"}`, ont.ProjectScope, err)
		return err
	}
	log.Printf(`{"event":"ontology.created","project_scope":"%s","ontology_id":"%s"}`, ont.ProjectScope, ont.OntologyID)
	return nil
}

// GetOntologyByProjectScope returns the paired ontology for a project scope, or nil if none.
func (p *PostgresOrgStore) GetOntologyByProjectScope(projectScope string) (*Ontology, error) {
	var ont Ontology
	err := p.db.QueryRow(`
		SELECT project_scope, ontology_id, iri, created_at, updated_at
		FROM ontologies WHERE project_scope = $1
	`, projectScope).Scan(&ont.ProjectScope, &ont.OntologyID, &ont.IRI, &ont.CreatedAt, &ont.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ont, nil
}

// ============================================================================
// Helpers
// ============================================================================

// nullString returns a *string for sql.NullString compatibility.
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
