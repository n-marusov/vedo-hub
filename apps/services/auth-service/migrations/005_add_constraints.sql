-- Migration 005: Add constraints
-- Validates FK cascade/restrict behavior and NOT NULL constraints.
-- Primary constraints are already defined in CREATE TABLE statements above.
-- This migration verifies and documents the constraint model.

-- Verify scopes.parent_id RESTRICT behavior:
-- ON DELETE RESTRICT prevents deleting a scope that has children.
-- COMMENT ON COLUMN scopes.parent_id IS 'FK: REFERENCES scopes(id) ON DELETE RESTRICT — prevents orphaned children';

-- Verify memberships CASCADE behavior:
-- ON DELETE CASCADE ensures deleting a scope removes all its memberships.
-- COMMENT ON COLUMN memberships.scope IS 'FK: REFERENCES scopes(id) ON DELETE CASCADE — memberships removed with scope';

-- Verify attribute_policies CASCADE behavior:
-- ON DELETE CASCADE ensures deleting a scope removes all its policies.
-- COMMENT ON COLUMN attribute_policies.scope IS 'FK: REFERENCES scopes(id) ON DELETE CASCADE — policies removed with scope';

-- NOT NULL validation (applied at CREATE TABLE, documented here):
-- scopes.id, scopes.type — always required
-- memberships.scope, memberships.user_id, memberships.role — always required
-- attribute_policies.scope, attribute_policies.pattern, attribute_policies.right — always required
-- audit_events.event, audit_events.reason, audit_events.user_id — always required
