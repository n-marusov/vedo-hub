-- Migration 011: Add slug to scopes table
--
-- GitLab-style human-readable identifiers for GUI navigation
-- (e.g. /group/subgroup/project). Derived from the name: lowercase,
-- spaces to hyphens, strip non-alphanumeric. Unique within the same
-- parent scope (NULL parent = top-level namespace).

ALTER TABLE scopes ADD COLUMN IF NOT EXISTS slug TEXT;

-- Unique slug within the same parent scope. Postgres treats NULLs as
-- distinct, so top-level scopes (parent_id NULL) never collide.
CREATE UNIQUE INDEX IF NOT EXISTS idx_scopes_slug_parent
    ON scopes (parent_id, slug)
    WHERE slug IS NOT NULL;

-- Backfill slugs for existing scopes from their names.
UPDATE scopes
SET slug = lower(regexp_replace(regexp_replace(name, '[^a-zA-Z0-9]+', '-', 'g'), '^-+|-+$', '', 'g'))
WHERE slug IS NULL AND name <> '';

-- Rollback (manual):
--   DROP INDEX IF EXISTS idx_scopes_slug_parent;
--   ALTER TABLE scopes DROP COLUMN IF EXISTS slug;
