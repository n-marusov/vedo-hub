-- Migration: Add indexes for commit history queries
-- Improves performance for commit listing, filtering, and branch navigation

CREATE INDEX IF NOT EXISTS idx_commits_branch_id ON commits(branch_id);
CREATE INDEX IF NOT EXISTS idx_commits_created_at ON commits(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_commits_author_id ON commits(author_id);
CREATE INDEX IF NOT EXISTS idx_branches_ontology_id ON branches(ontology_id);
CREATE INDEX IF NOT EXISTS idx_branches_name ON branches(name);

-- Composite index for common query pattern: list commits by branch ordered by time
CREATE INDEX IF NOT EXISTS idx_commits_branch_created ON commits(branch_id, created_at DESC);
