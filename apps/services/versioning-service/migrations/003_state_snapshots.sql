-- Migration: State snapshots for fast materialization
-- Caches materialized ontology state at specific commit points to avoid
-- replaying the entire delta chain from root on every checkout.

CREATE TABLE IF NOT EXISTS state_snapshots (
    commit_id UUID PRIMARY KEY,
    branch_id UUID NOT NULL,
    triples JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_state_snapshots_branch_id ON state_snapshots(branch_id);
