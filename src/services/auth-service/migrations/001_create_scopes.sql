-- Migration 001: Create scopes table
-- Stores group and ontology hierarchy nodes with visibility and tenant isolation.
--
-- Schema: vedo_org

CREATE TABLE IF NOT EXISTS scopes (
    id          TEXT PRIMARY KEY,
    type        TEXT NOT NULL CHECK (type IN ('group', 'ontology')),
    parent_id   TEXT REFERENCES scopes(id) ON DELETE RESTRICT,
    visibility  TEXT NOT NULL DEFAULT 'Private' CHECK (visibility IN ('Private', 'Internal', 'Public')),
    tenant_id   TEXT NOT NULL DEFAULT '',
    name        TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index: scope type lookups
CREATE INDEX IF NOT EXISTS idx_scopes_type ON scopes(type);

-- Index: parent-child hierarchy traversal
CREATE INDEX IF NOT EXISTS idx_scopes_parent ON scopes(parent_id);

-- Index: tenant isolation queries
CREATE INDEX IF NOT EXISTS idx_scopes_tenant ON scopes(tenant_id);
