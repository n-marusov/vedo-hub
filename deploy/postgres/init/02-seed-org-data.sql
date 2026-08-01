-- Seed demo data for E2E "wired" tests (groups, projects, members).
-- Inserted into vedo_org when the test environment starts.
-- Idempotent: uses ON CONFLICT / NOT EXISTS guards.

\connect vedo_org

-- Demo groups (Engineering with a Data Science subgroup, Research)
INSERT INTO scopes (id, type, parent_id, visibility, tenant_id, name, slug, description)
SELECT '11111111-1111-4111-8111-111111111111', 'group', NULL, 'Private', '', 'Engineering', 'engineering', 'Engineering team'
WHERE NOT EXISTS (SELECT 1 FROM scopes WHERE id = '11111111-1111-4111-8111-111111111111');

INSERT INTO scopes (id, type, parent_id, visibility, tenant_id, name, slug, description)
SELECT '22222222-2222-4222-8222-222222222222', 'group', '11111111-1111-4111-8111-111111111111', 'Private', '', 'Data Science', 'data-science', 'Data science team'
WHERE NOT EXISTS (SELECT 1 FROM scopes WHERE id = '22222222-2222-4222-8222-222222222222');

INSERT INTO scopes (id, type, parent_id, visibility, tenant_id, name, slug, description)
SELECT '33333333-3333-4333-8333-333333333333', 'group', NULL, 'Public', '', 'Research', 'research', 'Research division'
WHERE NOT EXISTS (SELECT 1 FROM scopes WHERE id = '33333333-3333-4333-8333-333333333333');

-- Demo project under Engineering
INSERT INTO scopes (id, type, parent_id, visibility, tenant_id, name, slug, description)
SELECT '44444444-4444-4444-8444-444444444444', 'project', '11111111-1111-4111-8111-111111111111', 'Private', '', 'Demo Project', 'demo-project', 'Demo project for E2E tests'
WHERE NOT EXISTS (SELECT 1 FROM scopes WHERE id = '44444444-4444-4444-8444-444444444444');

-- Memberships: OWNER_JWT user (user-123) is Owner of Engineering and Research;
-- a member exists on the project for members-page wired tests.
INSERT INTO memberships (scope, user_id, role, inherited)
SELECT '11111111-1111-4111-8111-111111111111', 'user-123', 'Owner', false
WHERE NOT EXISTS (SELECT 1 FROM memberships WHERE scope = '11111111-1111-4111-8111-111111111111' AND user_id = 'user-123');

INSERT INTO memberships (scope, user_id, role, inherited)
SELECT '33333333-3333-4333-8333-333333333333', 'user-123', 'Owner', false
WHERE NOT EXISTS (SELECT 1 FROM memberships WHERE scope = '33333333-3333-4333-8333-333333333333' AND user_id = 'user-123');

INSERT INTO memberships (scope, user_id, role, inherited)
SELECT '44444444-4444-4444-8444-444444444444', 'user-123', 'Owner', false
WHERE NOT EXISTS (SELECT 1 FROM memberships WHERE scope = '44444444-4444-4444-8444-444444444444' AND user_id = 'user-123');

INSERT INTO memberships (scope, user_id, role, inherited)
SELECT '44444444-4444-4444-8444-444444444444', 'editor-user', 'Developer', false
WHERE NOT EXISTS (SELECT 1 FROM memberships WHERE scope = '44444444-4444-4444-8444-444444444444' AND user_id = 'editor-user');

-- Paired ontology for the demo project (1:1 Project ↔ Ontology)
INSERT INTO ontologies (project_scope, ontology_id, iri)
SELECT '44444444-4444-4444-8444-444444444444', '00000000-0000-0000-0000-0000000000aa', 'http://vedo-core.local/demo-project'
WHERE NOT EXISTS (SELECT 1 FROM ontologies WHERE project_scope = '44444444-4444-4444-8444-444444444444');
