-- Migration 007: Rename scopes.type='ontology' to 'project'
--
-- Aligns the vedo_org schema with the 1:1 Project ↔ Ontology model:
--   * scopes holds the Project container (type='project')
--   * ontologies (migration 008) holds the graph content (TBox/ABox)
--
-- MUST run AFTER migration 008 (ontologies table exists) and migration 009
-- (backfill from type='ontology' rows). The migration runner uses an explicit
-- list, not alphabetical order, so the execution order is: 008 → 009 → 007.
-- See postgres_store.go runMigrations() for the canonical ordering.
--
-- Schema: vedo_org

-- Step 1: rename existing ontology scopes to project
UPDATE scopes SET type = 'project', updated_at = now() WHERE type = 'ontology';

-- Step 2: update the CHECK constraint to accept the new vocabulary
ALTER TABLE scopes DROP CONSTRAINT IF EXISTS scopes_type_check;
ALTER TABLE scopes ADD CONSTRAINT scopes_type_check
    CHECK (type IN ('group', 'project'));

-- Rollback (manual — no down-migration framework in place):
--   ALTER TABLE scopes DROP CONSTRAINT scopes_type_check;
--   ALTER TABLE scopes ADD CONSTRAINT scopes_type_check
--       CHECK (type IN ('group', 'ontology'));
--   UPDATE scopes SET type = 'ontology', updated_at = now() WHERE type = 'project';
-- Caution: rollback after new 'project' scopes have been created would
-- conflate them with the legacy 'ontology' vocabulary. Only roll back if no
-- post-migration project scopes exist.
