-- Migration 004: Create audit_events table
-- Stores authorization audit trail for membership changes, visibility changes, policy changes.

CREATE TABLE IF NOT EXISTS audit_events (
    id          SERIAL PRIMARY KEY,
    event       TEXT NOT NULL,
    reason      TEXT NOT NULL,
    user_id     TEXT NOT NULL,
    object_type TEXT,
    object_id   TEXT,
    source_ip   TEXT,
    timestamp   TIMESTAMPTZ NOT NULL DEFAULT now(),
    trace_id    TEXT
);

-- Index: user audit trail lookups
CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_events(user_id);

-- Index: time-ordered audit queries
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_events(timestamp DESC);
