-- Migration 002: Create memberships table
-- Stores role assignments for users within scopes (groups/ontologies).
--
-- UNIQUE constraint on (scope, user_id) enables upsert semantics.
-- Role CHECK constraint enforces valid role values at the database level.

CREATE TABLE IF NOT EXISTS memberships (
    scope       TEXT NOT NULL REFERENCES scopes(id) ON DELETE CASCADE,
    user_id     TEXT NOT NULL,
    role        TEXT NOT NULL CHECK (role IN ('Viewer', 'Editor', 'Maintainer', 'SupportEngineer', 'SRE', 'SecurityLead', 'ProductOwner', 'Owner')),
    inherited   BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (scope, user_id)
);

-- Index: user membership lookups
CREATE INDEX IF NOT EXISTS idx_memberships_user ON memberships(user_id);

-- Index: scope membership lookups
CREATE INDEX IF NOT EXISTS idx_memberships_scope ON memberships(scope);
