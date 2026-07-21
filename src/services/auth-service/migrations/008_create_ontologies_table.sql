-- Migration 008: Create ontologies table (1:1 with project scopes)
--
-- Each row represents the graph content (TBox/ABox, classes, properties,
-- individuals, axioms) of a Project. The 1:1 invariant is enforced at the
-- schema level: project_scope is both the PRIMARY KEY and a foreign key to
-- scopes(id), so a Project cannot have more than one Ontology row.
--
-- MUST run BEFORE migration 009 (backfill needs the table) and migration 007
-- (type rename). Execution order: 008 → 009 → 007.
--
-- Schema: vedo_org

CREATE TABLE IF NOT EXISTS ontologies (
    -- project_scope is both PK and FK: enforces 1:1 (one ontology per project scope)
    project_scope TEXT PRIMARY KEY REFERENCES scopes(id) ON DELETE CASCADE,
    -- ontology_id is the identifier used by ontology-service to address the graph
    ontology_id   TEXT NOT NULL UNIQUE,
    iri           TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index: look up ontologies by ontology_id (used by ontology-service)
CREATE INDEX IF NOT EXISTS idx_ontologies_ontology_id ON ontologies(ontology_id);
