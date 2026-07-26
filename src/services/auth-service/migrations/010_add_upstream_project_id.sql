-- Migration 010: Add upstream_project_id to scopes table
--
-- Tracks fork ancestry: each fork stores a reference to its parent Project.
-- When the upstream project is deleted, the fork survives (orphan fork)
-- with upstream_project_id set to NULL.
--
-- Schema: vedo_org

ALTER TABLE scopes ADD COLUMN IF NOT EXISTS upstream_project_id UUID REFERENCES scopes(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_scopes_upstream ON scopes(upstream_project_id) WHERE upstream_project_id IS NOT NULL;

-- Rollback (manual):
--   DROP INDEX IF EXISTS idx_scopes_upstream;
--   ALTER TABLE scopes DROP COLUMN IF EXISTS upstream_project_id;
