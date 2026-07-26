-- Migration 009: Backfill ontologies table from legacy scopes
--
-- Copies existing project/ontology scopes into the ontologies table with
-- project_scope = scopes.id and generates new UUID for ontology_id.
-- The legacy scopes used slug-based IDs; we now generate proper UUIDs for
-- ontology_id as per ADR-DES.DATA.uuid-identifiers-for-groups-projects-mandate.
--
-- MUST run AFTER migration 008 (ontologies table exists) and BEFORE migration
-- 007 (type rename). The filter accepts both 'ontology' and 'project' types so
-- the migration is order-independent with respect to 007 — it backfills
-- correctly whether the type rename has been applied or not.
--
-- Idempotent: ON CONFLICT (project_scope) DO NOTHING — safe to re-run.
--
-- Execution order: 008 → 009 → 007.
-- Schema: vedo_org

INSERT INTO ontologies (project_scope, ontology_id, iri)
SELECT id, gen_random_uuid(), '' FROM scopes
WHERE type IN ('ontology', 'project')
ON CONFLICT (project_scope) DO NOTHING;
