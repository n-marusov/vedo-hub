-- Migration 003: Create attribute_policies table
-- Stores ABAC policy definitions scoped to ontologies or groups.

CREATE TABLE IF NOT EXISTS attribute_policies (
    id          SERIAL PRIMARY KEY,
    scope       UUID NOT NULL REFERENCES scopes(id) ON DELETE CASCADE,
    pattern     JSONB NOT NULL,
    right       TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index: scope-based policy lookups
CREATE INDEX IF NOT EXISTS idx_policies_scope ON attribute_policies(scope);
