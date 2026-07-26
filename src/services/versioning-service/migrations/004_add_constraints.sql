-- Migration 004: Add database constraints
-- Unique index on branches(ontology_id, name) to prevent duplicate branch names
-- within the same ontology. Foreign key constraints with CASCADE delete on
-- state_snapshots to maintain referential integrity.

CREATE UNIQUE INDEX IF NOT EXISTS idx_branches_ontology_name ON branches(ontology_id, name);
DROP INDEX IF EXISTS idx_branches_name;

ALTER TABLE state_snapshots ADD CONSTRAINT fk_snapshots_commit FOREIGN KEY (commit_id) REFERENCES commits(id) ON DELETE CASCADE;
ALTER TABLE state_snapshots ADD CONSTRAINT fk_snapshots_branch FOREIGN KEY (branch_id) REFERENCES branches(id) ON DELETE CASCADE;
