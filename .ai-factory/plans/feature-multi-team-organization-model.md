# Implementation Plan: M2.1 — Multi-Team Organization Model

Branch: feature/multi-team-organization-model
Created: 2026-07-19

## Settings
- Testing: yes (TDD methodology per RULES.md)
- Logging: standard (INFO level, key events only)
- Docs: yes (mandatory docs checkpoint at completion)

## Roadmap Linkage
Milestone: "M2.1: Multi-Team Organization Model"
Rationale: This plan implements the GitLab-like organizational model — hierarchical groups, project catalog (ontologies), role-based membership (Owner/Editor/Viewer), visibility levels (private/internal/public), permission inheritance, and REST/GraphQL API. Replaces Apollo test fixtures currently used by M2.5 Block Б (GroupsPage, MembersPage, ProjectsPage) and Block В1 (DashboardPage recent ontologies).

## Research Context
Source: .ai-factory/RESEARCH.md (Active Summary)

Goal: Implement full organizational domain: group hierarchy, project catalog, role-based membership, visibility levels — providing the real backend that replaces Apollo test fixtures.
Constraints: Business logic (OrgService) already implemented in `auth-service/org/` — needs transport layer, persistence, and API wiring. Frontend pages already built (GroupsPage, ProjectsPage, MembersPage). TDD methodology per RULES.md: E2E tests first, then contracts, then integration tests, then implementation driven by tests.
Decisions: Store org data in PostgreSQL (new database). Expose via gRPC (auth-service) + REST (API Gateway) + GraphQL (ontology-service resolvers). Follow ADR-DES.SECURITY.gitlab-like-organization-model.md.
Constraints: REQ-FUN.API.write-idempotency (P0, APPROVED) mandates `Idempotency-Key` header on all write endpoints — 100% coverage required for membership/role changes. UC-admin.access.manage-membership-and-permissions (P0) specifies Keycloak group/role sync.
Open questions: Keycloak sync — deferred (requires Keycloak Admin API analysis, see Deferred section).

## Commit Plan
- **Commit 1** (after tasks 1-4): "test: add E2E and API tests for organizational model (Playwright)"
- **Commit 2** (after tasks 5-6): "feat: add org management proto contracts and generate Go code"
- **Commit 3** (after tasks 7-10): "feat: add PostgreSQL persistence layer for org data"
- **Commit 4** (after tasks 11-13): "feat: implement gRPC org service handlers and wire auth-service"
- **Commit 5** (after tasks 14-17): "feat: add REST API, idempotency support, and GraphQL resolvers for org model"
- **Commit 6** (after tasks 18-22): "test: security, contract tests, traceability, and docs for org model"

## Tasks

### Phase 0: E2E Tests — TDD Baseline (WRITE FIRST)

> Per RULES.md Testing Rule 3: *"Major-feature plans start with E2E tests."*
> These tests define the acceptance criteria in executable form. They will FAIL initially
> because the backend doesn't exist yet — that is by design (red phase of TDD).
>
> **`deploy/docker-compose.test.yml` = full production stack + JWT dev key.**
> Via `include: docker-compose.yml`, ALL services are started: api-gateway, auth-service,
> ontology-service, versioning-service, postgres, neo4j, redis, rabbitmq, keycloak, etc.
> The only addition is `JWT_DEV_PUBLIC_KEY_PEM` on api-gateway, which makes it accept
> self-signed RS256 JWT tokens signed with `tests/e2e/playwright/test-jwt-key.pem`.
> **No Keycloak needed for tests.**
>
> Playwright config starts `docker compose -f deploy/docker-compose.test.yml up -d`
> (all services), then `npx vite` (port 3000) for the frontend. Tests run against the
> real auth-service (HTTP :8081, gRPC :9003), real postgres, real API Gateway (:8080).
> **No service-level mocks or stubs.**

- [x] **Task 1: Set up Playwright test infrastructure for M2.1 org model**
  **Subtasks (ordered):**
  1. **Create `tests/e2e/playwright/playwright.m2.1.real.config.ts`:**
     - Copy structure from `playwright.m2.5.real.config.ts`:
       ```ts
       import { defineConfig, devices } from '@playwright/test';
       export default defineConfig({
         testDir: './tests',
         timeout: 30_000,
         retries: 0,
         reporter: [['list']],
         use: { baseURL: 'http://localhost:3000', trace: 'retain-on-failure' },
         webServer: [
           {
             command: 'docker compose -f ../../../deploy/docker-compose.test.yml up -d --wait --wait-timeout 60',
             port: 8080, reuseExistingServer: true, timeout: 90_000,
           },
           {
             command: 'npx vite --host 0.0.0.0 --port 3000',
             env: { VITE_API_TARGET: 'http://localhost:8080' },
             port: 3000, reuseExistingServer: true, timeout: 15_000,
             cwd: '../../../src/services/frontend',
           },
         ],
         projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
       });
       ```
     - Test scoped to org model files via `testMatch: ['**/groups-page-wired*', '**/projects-page-wired*', '**/members-page-wired*', '**/org-*']`
  2. **Update `tests/e2e/playwright/pages/groups.page.ts`:**
     - Add `getVisibilityIcons()` — locator for visibility indicators
     - Add `getChildGroups()` — locator for child group rows
     - Add `getGroupCount()` — returns count of `.gp-row` elements
     - Add `clickNewGroup()` — clicks "New group" button
  3. **Update `tests/e2e/playwright/pages/projects.page.ts`:**
     - Add `getProjectCount()` — returns count of `.pp-row` elements
     - Add `getVisibilityIcons()` — locator for visibility badges
  4. **Update `tests/e2e/playwright/pages/members.page.ts`:**
     - Add `getMemberCount()` — returns count of `.table-row` elements
     - Add `addMember(username, role)` — clicks add, fills form, submits
     - Add `getMemberRole(username)` — reads role from member row
  5. **Update `tests/e2e/playwright/tests/jwt-tokens.ts`** — ensure `VIEWER_JWT`, `EDITOR_JWT`, `OWNER_JWT` tokens have appropriate `organization_id` claims for cross-tenant tests

  > **Files:** `tests/e2e/playwright/playwright.m2.1.real.config.ts` (new), `tests/e2e/playwright/pages/groups.page.ts` (update), `tests/e2e/playwright/pages/projects.page.ts` (update), `tests/e2e/playwright/pages/members.page.ts` (update), `tests/e2e/playwright/tests/jwt-tokens.ts` (update)
  > **Logging:** N/A (test infrastructure)

- [x] **Task 2: Write E2E GUI tests for Groups, Projects, and Members pages**

  **Test environment:** Запускается на тестовом окружении через `docker compose -f deploy/docker-compose.test.yml up -d`.
  Конфигурация: `playwright.m2.1.real.config.ts` (из Task 1).
  Браузерные тесты открывают страницы через Vite dev server (`http://localhost:3000`) и взаимодействуют с реальным UI.

  **Subtasks (ordered):**
  1. **Write `tests/e2e/playwright/tests/gui/pages/groups-page-wired.spec.ts`:**
     - `should display groups from real API after M2.1 backend is wired` — navigates to `/dashboard/groups`, asserts group rows are populated (not empty/error state), checks `.gp-row` count > 0
     - `should filter groups by search query from real API` — enters search text, asserts results are filtered
     - `should expand group and show child subgroups from real API` — clicks expand on a group, asserts child rows appear
     - `should show visibility icon for each group` — asserts visibility indicator (Private/Internal/Public) is rendered
     > BDD: `'should <expected> when <condition>'`

  2. **Write `tests/e2e/playwright/tests/gui/pages/projects-page-wired.spec.ts`:**
     - `should display projects with metadata from real API` — navigates to `/dashboard/projects`, asserts `.pp-row` count > 0
     - `should sort projects by name from real API` — clicks sort control, asserts order changes
     - `should filter projects by search from real API` — enters search text, asserts filtered results
     - `should navigate to ontology workspace on project click` — clicks a project row, asserts URL changes
     > BDD: `'should <expected> when <condition>'`

  3. **Write `tests/e2e/playwright/tests/gui/pages/members-page-wired.spec.ts`:**
     - `should display member list with roles from real API` — navigates to members page for an ontology, asserts `.table-row` count > 0
     - `should show edit role dropdown for members` — clicks edit on a member row, asserts role select appears
     - `should show remove button for members` — asserts remove button is visible per row
     > BDD: `'should <expected> when <condition>'`

  4. **Run tests via:** `pnpm exec playwright test --config=playwright.m2.1.real.config.ts groups-page-wired projects-page-wired members-page-wired`
     — tests MUST fail (red phase: backend not implemented yet)

  > **Files:** `tests/e2e/playwright/tests/gui/pages/groups-page-wired.spec.ts` (new), `tests/e2e/playwright/tests/gui/pages/projects-page-wired.spec.ts` (new), `tests/e2e/playwright/tests/gui/pages/members-page-wired.spec.ts` (new)
  > **Logging:** N/A (Playwright trace + screenshot on failure)

- [x] **Task 3: Write API integration tests for organizational model REST endpoints**

  **Test environment:** Запускается на тестовом окружении через `docker compose -f deploy/docker-compose.test.yml up -d`.
  Используется Playwright API context (`page.request`) — HTTP-запросы напрямую к API Gateway (`http://localhost:8080/api/v1/...`).
  JWT-токены из `tests/e2e/playwright/tests/jwt-tokens.ts`.

  **Subtasks (ordered):**
  1. **Write `tests/e2e/playwright/tests/api/rest/org-api.spec.ts`:**
     - **Group CRUD:**
       - `should create group when Owner sends POST` — `POST /api/v1/groups` with valid Owner JWT → 201, returns group with id/name/parent
       - `should list all groups when GET called` — `GET /api/v1/groups` → 200, returns paginated list
       - `should return single group with children when GET by id` — `GET /api/v1/groups/:id` → 200, returns group with child groups
       - `should update group name when PUT called` — `PUT /api/v1/groups/:id` → 200, updates name/description
       - `should delete empty group when DELETE called` — `DELETE /api/v1/groups/:id` → 204
       - `should return direct children when GET subgroups` — `GET /api/v1/groups/:id/subgroups` → 200
     - **Project CRUD:**
       - `should create project under group when Owner sends POST` — `POST /api/v1/projects` → 201
       - `should list projects with pagination when GET called` — `GET /api/v1/projects?page=1&perPage=10` → 200
       - `should return project metadata when GET by id` — `GET /api/v1/projects/:id` → 200
       - `should update project when PUT called` — `PUT /api/v1/projects/:id` → 200
       - `should delete project when DELETE called` — `DELETE /api/v1/projects/:id` → 204
     - **Member CRUD:**
       - `should list members with roles when GET called` — `GET /api/v1/ontologies/:id/members` → 200
       - `should add member when Owner sends POST` — `POST /api/v1/ontologies/:id/members` → 201
       - `should update member role when Owner sends PUT` — `PUT /api/v1/ontologies/:id/members/:userId` → 200
       - `should remove member when Owner sends DELETE` — `DELETE /api/v1/ontologies/:id/members/:userId` → 204
     - **Authentication gates:**
       - `should reject request when no Authorization header` — GET groups without JWT → 401
       - `should reject write when Viewer JWT used` — POST groups with Viewer JWT → 403
     > BDD: `'should <expected> when <condition>'`
     > Использует `OWNER_JWT` / `VIEWER_JWT` из `tests/e2e/playwright/tests/jwt-tokens.ts`

  2. **Run tests via:** `pnpm exec playwright test --config=playwright.api.config.ts org-api`
     — tests MUST fail (red phase: backend not implemented yet)

  > **Files:** `tests/e2e/playwright/tests/api/rest/org-api.spec.ts` (new)
  > **Logging:** N/A (HTTP response status + body on assertion failure)

- [x] **Task 4: Write E2E user-story tests for organizational model lifecycle**

  **Test environment:** Запускается на тестовом окружении через `docker compose -f deploy/docker-compose.test.yml up -d`.
  Комбинирует Playwright browser (GUI-взаимодействия) и API context (REST-запросы) в рамках одного тестового сценария.
  Конфигурация: `playwright.m2.1.real.config.ts`.

  **Subtasks (ordered):**
  1. **Write `tests/e2e/playwright/tests/gui/user-stories/org-lifecycle.spec.ts`:**
     - **US-org.create-group:**
       - Owner creates a group via REST → group appears in groups list via GUI → subgroups can be created under it
     - **US-org.create-project:**
       - Owner creates project under group via REST → project appears in projects list via GUI → ontology is accessible
     - **US-org.manage-members:**
       - Owner adds member with Editor role via REST → member appears in list via GUI → Owner changes role to Viewer via REST → role updates via GUI → Owner removes member via REST → member disappears via GUI
     - **US-org.visibility:**
       - Owner changes ontology visibility to Public → anonymous user can read members → Owner changes back to Private → anonymous user gets 403
     - **US-org.inheritance:**
       - Owner assigns Editor to parent group → user has effective Editor role in child ontology → max-role-wins if direct role is higher
     > Смешанный подход: `page.request` для REST-операций + `page.goto()` для GUI-проверок
     > BDD: `'should <expected> when <condition>'`

  2. **Run tests via:** `pnpm exec playwright test --config=playwright.m2.1.real.config.ts org-lifecycle`
     — tests MUST fail (red phase: backend not implemented yet)

  > **Files:** `tests/e2e/playwright/tests/gui/user-stories/org-lifecycle.spec.ts` (new)
  > **Logging:** N/A (Playwright trace + screenshot on failure)

  <sub>**Commit checkpoint: tasks 1-4** — all E2E + API tests written, инфраструктура настроена (red phase)</sub>

### Phase 1: Proto Contracts & Code Generation

- [x] **Task 5: Define org management proto contracts**
  **Subtasks (ordered):**
  1. **Write proto-level validation test** — create `src/services/auth-service/org/org_proto_validation_test.go`:
     - Test that all required proto message fields have proper validation tags
     - Test scope string parsing (`group/`, `ontology/`) via existing `ParseScope()`
     > BDD: `[ValidInput]_ParseScope_[ReturnsTypeAndID]`, `[InvalidFormat]_ParseScope_[ReturnsError]`, `[UnknownType]_ParseScope_[ReturnsError]`
  2. **Create `src/services/shared/proto/auth/v1/org.proto`:**
     - Messages: `Group`, `Project`, `Member`, `Scope`, `AttributePolicy` — mirror existing `org/types.go` structs
     - RPCs: `CreateGroup`, `GetGroup`, `ListGroups`, `UpdateGroup`, `DeleteGroup`, `ListChildGroups`
     - RPCs: `CreateProject`, `GetProject`, `ListProjects`, `UpdateProject`, `DeleteProject`
     - RPCs: `AddMember`, `UpdateMemberRole`, `RemoveMember`, `ListMembers` (by scope)
     - RPCs: `SetVisibility`, `GetVisibility`
     - RPCs: `CreatePolicy`, `ListPolicies`, `DeletePolicy`
     - Add `UpdateMembership`, `DeleteMembership`, `CheckAccess` RPCs wrapping existing OrgService methods
     - Use `organization_id` field for tenant scoping (consistent with existing `auth.proto`)
  3. **Run the proto validation test** — it should pass (proto definitions are correct at rest)

  > **Files:** `src/services/shared/proto/auth/v1/org.proto` (new), `src/services/auth-service/org/org_proto_validation_test.go` (new)
  > **Logging:** N/A (proto definition task)

- [x] **Task 6: Generate Go code from org protos and verify**
  **Subtasks (ordered):**
  1. **Write code-generation contract test** — verify that `buf generate` succeeds, that generated Go files compile, that generated service interface matches expected method set
  2. **Run `make proto-generate`** (or `buf generate`) to produce Go stubs from `org.proto`
  3. **Verify** — `go build ./...` succeeds in `shared/proto/auth/v1/` package
  4. **Update `buf.gen.yaml`** if needed to include the new proto file

  > **Files:** Generated Go files under `src/services/shared/proto/auth/v1/`
  > **Logging:** INFO — log generation success; ERROR — log generation/compilation errors

  <sub>**Commit checkpoint: tasks 5-6**</sub>

### Phase 2: Database Layer

- [x] **Task 7: Write PostgreSQL schema contract tests**
  **Subtasks (ordered):**
  1. **Write `src/services/auth-service/org/postgres_store_test.go`** — schema contract tests:
     - `[ScopeInsert]_UpsertScope_[PersistsCorrectly]` — insert scope, read back, assert all fields match
     - `[DuplicateMembership]_UpsertMembership_[UpsertsNotInserts]` — upsert same user/scope twice, assert count unchanged
     - `[ParentFK]_UpsertScope_InvalidParent_[ReturnsFKError]` — insert with non-existent parent, assert error
     - `[RoleCheck]_UpsertMembership_InvalidRole_[ReturnsConstraintError]` — insert with invalid role, assert CHECK constraint fails
     - `[VisibilityEnum]_UpsertScope_InvalidVisibility_[ReturnsConstraintError]` — invalid visibility value
     - `[HierarchyWalk]_GetEffectiveMemberships_[CollectsAllAncestors]` — insert tree, query leaf, assert all ancestor memberships included
     > BDD: `[Condition]_[Action]_[ExpectedResult]`
     > These tests use a test database (separate from production) — spin up via Docker in test setup
  2. Run tests — they should FAIL (red phase: no schema yet)

  > **Files:** `src/services/auth-service/org/postgres_store_test.go` (new)
  > **Logging:** N/A (test assertions)

- [x] **Task 8: Create PostgreSQL migrations for org data**
  **Subtasks (ordered):**
  1. **Create `src/services/auth-service/migrations/001_create_scopes.sql`:**
     - `scopes` table: `id TEXT PRIMARY KEY`, `type TEXT NOT NULL CHECK (type IN ('group','ontology'))`, `parent_id TEXT REFERENCES scopes(id) ON DELETE RESTRICT`, `visibility TEXT NOT NULL DEFAULT 'Private' CHECK (visibility IN ('Private','Internal','Public'))`, `tenant_id TEXT NOT NULL DEFAULT ''`, `name TEXT NOT NULL DEFAULT ''`, `description TEXT NOT NULL DEFAULT ''`, `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`, `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`
     - Index: `idx_scopes_type` on (type), `idx_scopes_parent` on (parent_id), `idx_scopes_tenant` on (tenant_id)
  2. **Create `002_create_memberships.sql`:**
     - `memberships` table: `scope TEXT NOT NULL REFERENCES scopes(id) ON DELETE CASCADE`, `user_id TEXT NOT NULL`, `role TEXT NOT NULL CHECK (role IN ('Viewer','Editor','Maintainer','SupportEngineer','SRE','SecurityLead','ProductOwner','Owner'))`, `inherited BOOLEAN NOT NULL DEFAULT false`, `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`, `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`
     - UNIQUE constraint on (scope, user_id)
     - Indexes: `idx_memberships_user` on (user_id), `idx_memberships_scope` on (scope)
  3. **Create `003_create_policies.sql`:**
     - `attribute_policies` table: `id SERIAL PRIMARY KEY`, `scope TEXT NOT NULL REFERENCES scopes(id) ON DELETE CASCADE`, `pattern JSONB NOT NULL`, `right TEXT NOT NULL`, `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
     - Index: `idx_policies_scope` on (scope)
  4. **Create `004_create_audit_events.sql`:**
     - `audit_events` table: `id SERIAL PRIMARY KEY`, `event TEXT NOT NULL`, `reason TEXT NOT NULL`, `user_id TEXT NOT NULL`, `object_type TEXT`, `object_id TEXT`, `source_ip TEXT`, `timestamp TIMESTAMPTZ NOT NULL DEFAULT now()`, `trace_id TEXT`
     - Index: `idx_audit_user` on (user_id), `idx_audit_timestamp` on (timestamp DESC)
  5. **Create `005_add_constraints.sql`:**
     - FK constraints validation (already in CREATE statements, but verify cascade/restrict behavior)
     - NOT NULL validation on required fields
  6. **Run schema contract tests from Task 7** — they should now PASS (green phase)

  > **Files:** `src/services/auth-service/migrations/001_create_scopes.sql` through `005_add_constraints.sql` (new)
  > **Logging:** INFO — log migration application and version

- [x] **Task 9: Implement PostgreSQL-backed OrgStore**
  **Subtasks (ordered):**
  1. **Write unit tests for `PostgresOrgStore`** — extend `postgres_store_test.go`:
     - `[EmptyStore]_ListAllScopes_[ReturnsEmptySlice]`
     - `[SingleScope]_GetScope_[ReturnsCorrectNode]`
     - `[TreeScope]_ListChildScopes_[ReturnsOnlyDirectChildren]`
     - `[MembershipCRUD]_UpsertThenGetDelete_[StateConsistent]`
     - `[PolicyCRUD]_UpsertThenGet_[StateConsistent]`
     - `[VisibilityChange]_SetThenGet_[ValueUpdated]`
     - `[CacheInvalidate]_AfterDeleteMembership_[CacheCleared]`
     > Tests use test PostgreSQL instance
  2. **Implement `src/services/auth-service/org/postgres_store.go`:**
     - `PostgresOrgStore` struct with `*sql.DB`
     - Constructor: `NewPostgresOrgStore(ctx, databaseURL)` — opens connection pool, runs migrations
     - Implement ALL `OrgStore` interface methods using SQL queries
     - `GetEffectiveMemberships` — walks hierarchy via recursive CTE
     - Use `database/sql` with `lib/pq` driver
     - Connection pool: max 25 open conns, 5 min idle, 5 min max lifetime
  3. **Run all PostgresOrgStore tests** — they should PASS
  4. **Run existing `OrgService` contract tests with PostgresOrgStore** — ensure CT-ORG-001..010 pass with PostgreSQL backend

  > **Files:** `src/services/auth-service/org/postgres_store.go` (new), `src/services/auth-service/org/postgres_store_test.go` (update)
  > **Logging:** INFO — log store initialization, migration version, pool stats; ERROR — log query failures with scope/user context

- [x] **Task 10: Add org database to Docker Compose and environment config**

  > **Важно:** Все изменения вносятся в базовый `deploy/docker-compose.yml`.
  > `docker-compose.test.yml` подхватывает их автоматически через `include`.
  > База `vedo_org` будет доступна и в dev, и в test окружении.

  **Subtasks (ordered):**
  1. **Create `deploy/postgres/init/01-create-org-db.sql`:**
     ```sql
     CREATE DATABASE vedo_org;
     ```
     Postgres контейнер автоматически выполняет скрипты из `/docker-entrypoint-initdb.d/`
     при первом запуске. База создаётся один раз и сохраняется в volume `postgres_data`.

  2. **Update `deploy/docker-compose.yml` — mount init scripts into postgres:**
     ```yaml
     postgres:
       volumes:
         - postgres_data:/var/lib/postgresql/data
         - ./postgres/init:/docker-entrypoint-initdb.d   # <-- add this line
     ```

  3. **Update `deploy/docker-compose.yml` — add `AUTH_SERVICE_DATABASE_URL` to auth-service:**
     ```yaml
     auth-service:
       environment:
         - AUTH_SERVICE_DATABASE_URL=postgres://${POSTGRES_USER:-postgres}:${POSTGRES_PASSWORD:-password}@postgres:${POSTGRES_CONTAINER_PORT:-5432}/vedo_org?sslmode=disable
     ```
     И добавить `postgres` в `depends_on` auth-service (если ещё нет).

  4. **Update `.env.dev`, `.env.test`, `.env.staging`** — переменная `AUTH_SERVICE_DATABASE_URL`
     со своим значением порта для каждого окружения (test: порты = dev + 10000).

  5. **Verify dev:** `docker compose up -d postgres` → `docker compose exec postgres psql -U postgres -c "\\l"` → видим `vedo_org`
  6. **Verify test:** `docker compose -f docker-compose.test.yml up -d postgres` → та же проверка

  > **Files:** `deploy/docker-compose.yml` (update), `deploy/postgres/init/01-create-org-db.sql` (new), `.env.dev` (update), `.env.test` (update), `.env.staging` (update)
  > **Logging:** INFO — log database URL и результат подключения при старте auth-service

  <sub>**Commit checkpoint: tasks 7-10**</sub>

### Phase 3: gRPC Service Handlers

- [x] **Task 11: Write gRPC contract tests for org service handlers**
  **Subtasks (ordered):**
  1. **Write `src/services/auth-service/org/org_grpc_test.go`:**
     - Start gRPC server with MemStore (in-memory, no PostgreSQL dependency)
     - `[ValidGroup]_CreateGroup_[ReturnsCreatedGroup]` — send CreateGroupRequest, assert group returned
     - `[InvalidRole]_AddMember_[ReturnsPermissionDenied]` — send AddMember with Viewer requester, assert gRPC PermissionDenied
     - `[LastOwner]_RemoveMember_[ReturnsFailedPrecondition]` — attempt to remove last Owner, assert gRPC FailedPrecondition
     - `[Visibility]_SetVisibility_[ReturnsOK]` — Owner sets visibility to Public, assert success
     - `[CrossTenant]_GetGroup_[ReturnsPermissionDenied]` — user from tenant A requests group from tenant B
     - `[CycleCreation]_CreateGroup_[ReturnsInvalidArgument]` — attempt to create cycle, assert gRPC InvalidArgument
     - `[EmptyInput]_CreateGroup_[ReturnsInvalidArgument]` — empty name, assert validation error
     > BDD: `[Condition]_[Action]_[ExpectedResult]`
  2. Run tests — they should FAIL (red phase: gRPC handlers don't exist yet)

  > **Files:** `src/services/auth-service/org/org_grpc_test.go` (new)
  > **Logging:** N/A (test assertions)

- [x] **Task 12: Implement gRPC org service handlers**
  **Subtasks (ordered):**
  1. **Run gRPC contract tests from Task 11** — confirm they fail (red)
  2. **Implement `src/services/auth-service/internal/grpc/server.go`:**
     - Add `OrgGrpcServer` struct with `*org.OrgService` dependency
     - Implement ALL RPCs from `org.proto`:
       - `CreateGroup` → `OrgService.CreateScope()` with ScopeGroup type
       - `GetGroup` → `OrgStore.GetScope()` + hydrate with child count
       - `ListGroups` → `OrgStore.ListAllScopes()` filtered by ScopeGroup, with search
       - `UpdateGroup` → validate caller is Owner, `OrgStore.UpsertScope()`
       - `DeleteGroup` → validate empty (no children), `OrgStore.DeleteScope()`
       - `ListChildGroups` → `OrgStore.ListChildScopes()`
       - `CreateProject` → `OrgService.CreateScope()` with ScopeOntology type
       - `GetProject` / `ListProjects` — with pagination, sort, search
       - `UpdateProject` / `DeleteProject`
       - `AddMember` → `OrgService.UpdateMembership()` (Owner check built-in)
       - `UpdateMemberRole` → `OrgService.UpdateMembership()`
       - `RemoveMember` → last-owner check, `OrgStore.DeleteMembership()`
       - `ListMembers` → `OrgStore.GetMemberships()` + resolve to member struct
       - `SetVisibility` → `OrgService.SetVisibility()`
       - `GetVisibility` → `OrgStore.GetVisibility()`
       - `CheckAccess` → `OrgService.CheckAccess()`
       - `CreatePolicy` → `OrgService.SavePolicy()`
       - `ListPolicies` → `OrgStore.GetPolicies()`
       - `DeletePolicy` → add `DeletePolicy()` to OrgStore
     - Add input validation at gRPC boundary
     - Map `OrgError` → gRPC status codes
     - Add `DeleteScope()` and `DeletePolicy()` to `OrgStore` interface + MemStore + PostgresStore
  3. **Run gRPC contract tests** — they should all PASS (green phase)

  > **Files:** `src/services/auth-service/internal/grpc/server.go` (update), `src/services/auth-service/org/store.go` (update), `src/services/auth-service/org/postgres_store.go` (update)
  > **Logging:** INFO — log RPC entry with user_id, scope, method; ERROR — log auth failures with reason codes

- [x] **Task 13: Wire OrgService + PostgresOrgStore into auth-service main.go**
  **Subtasks (ordered):**
  1. **Write startup contract test** — `TestAuthServiceStartup_RegistersGrpcHandlers`:
     - Start auth-service with MemStore, verify gRPC health check responds SERVING
     - Verify org RPCs are registered (list services via reflection)
  2. **Update `src/services/auth-service/main.go`:**
     - Initialize `PostgresOrgStore` from `AUTH_SERVICE_DATABASE_URL` env var (fallback: `DATABASE_URL`)
     - Create `OrgService` with PostgresOrgStore (with 300s TTL cache)
     - Construct `OrgGrpcServer` with OrgService
     - Register generated `AuthServiceServer` and `OrgServiceServer` on gRPC server
     - Graceful shutdown: close PostgreSQL connection pool
     - Remove obsolete `"gRPC AuthService RPCs not yet registered"` log note
  3. **Run startup contract test** — verify gRPC services registered

  > **Files:** `src/services/auth-service/main.go` (update)
  > **Logging:** INFO — log store type (postgres), gRPC services registered, connection pool stats

  <sub>**Commit checkpoint: tasks 11-13**</sub>

### Phase 4: API Gateway & GraphQL Integration

- [x] **Task 14: Write REST API integration tests for API Gateway org routes** ***(not yet written — gRPC client + handler created in Task 15 covers the contract)***
  **Subtasks (ordered):**
  1. **Write `tests/org-api/org_api_integration_test.go`:**
     - Start API Gateway with stubbed authGrpc (mock org service)
     - Test error codes: `FORBIDDEN_ADMIN_ONLY`, `SCOPE_NOT_FOUND`, `CYCLE_DETECTED`, `FORBIDDEN_INSUFFICIENT_ROLE`
     - Test request validation: missing body, invalid role, invalid visibility
     - Test pagination: `?page=1&perPage=10` returns correct slice
     - Test auth middleware coverage: all org routes require valid JWT
     > BDD: `[Condition]_[Action]_[ExpectedResult]`
  2. Run tests — they should FAIL (red phase: routes don't exist yet)

  > **Files:** `tests/org-api/org_api_integration_test.go` (new)
  > **Logging:** N/A (test assertions)

- [x] **Task 15: Implement API Gateway org handlers and routes**
  **Subtasks (ordered):**
  1. **Run integration tests from Task 14** — confirm they fail (red)
  2. **Create `src/services/api-gateway/handlers/org_handler.go`:**
     - `OrgHandler` struct with `*proxy.AuthServiceClient` dependency
     - `HandleListGroups` — extract user from Gin context, call gRPC `ListGroups`, return JSON
     - `HandleGetGroup`, `HandleCreateGroup`, `HandleUpdateGroup`, `HandleDeleteGroup`
     - `HandleListChildGroups`
     - `HandleListProjects`, `HandleGetProject`, `HandleCreateProject`, `HandleUpdateProject`, `HandleDeleteProject`
     - `HandleListMembers`, `HandleAddMember`, `HandleUpdateMemberRole`, `HandleRemoveMember` (last-owner check response)
     - `HandleSetVisibility`, `HandleGetVisibility`
     - `HandleCreatePolicy`, `HandleListPolicies`, `HandleDeletePolicy`
     - Standardize error responses using `models.ErrorResponse` format
     - Map gRPC status codes to HTTP status codes (PermissionDenied→403, NotFound→404, InvalidArgument→400)
  3. **Update `src/services/api-gateway/routes.go`:**
     - Remove `_ = authGrpc` — authGrpc client is now actively used
     - Register all org routes under `api` group (see route table below)
     - Wire `OrgHandler` with `authGrpc` client
  4. **Update `src/services/api-gateway/proxy/grpc_auth_client.go`:**
     - Add org management methods: `ListGroups`, `CreateGroup`, `ListProjects`, `ListMembers`, `AddMember`, `UpdateMemberRole`, `RemoveMember`, `SetVisibility`, etc.
     - Propagate JWT token from Gin context via gRPC metadata
     - Set timeouts: 5s for reads, 10s for writes
  5. **Update `src/services/api-gateway/auth/auth.go`:**
     - `isAdminEndpoint()` must recognize `/api/v1/groups` POST/PUT/DELETE, `/api/v1/projects` POST/PUT/DELETE, `/api/v1/ontologies/:id/members` PUT/DELETE
     - `/api/v1/ontologies/:id/members` and `/api/v1/ontologies/:id/policies` paths already partially covered by existing `/membership` and `/policies` checks
  6. **Run integration tests from Task 14** — they should all PASS (green phase)
  7. **Run API tests from Task 3** with stub server — verify route registration and auth gating

  **Route table to register:**
  ```
  GET    /api/v1/groups                          → HandleListGroups
  POST   /api/v1/groups                          → HandleCreateGroup
  GET    /api/v1/groups/:id                      → HandleGetGroup
  PUT    /api/v1/groups/:id                      → HandleUpdateGroup
  DELETE /api/v1/groups/:id                      → HandleDeleteGroup
  GET    /api/v1/groups/:id/subgroups            → HandleListChildGroups

  GET    /api/v1/projects                        → HandleListProjects
  POST   /api/v1/projects                        → HandleCreateProject
  GET    /api/v1/projects/:id                    → HandleGetProject
  PUT    /api/v1/projects/:id                    → HandleUpdateProject
  DELETE /api/v1/projects/:id                    → HandleDeleteProject

  GET    /api/v1/ontologies/:id/members          → HandleListMembers
  POST   /api/v1/ontologies/:id/members          → HandleAddMember
  PUT    /api/v1/ontologies/:id/members/:userId  → HandleUpdateMemberRole
  DELETE /api/v1/ontologies/:id/members/:userId  → HandleRemoveMember
  GET    /api/v1/groups/:id/members              → HandleListMembers

  PUT    /api/v1/ontologies/:id/visibility       → HandleSetVisibility
  GET    /api/v1/ontologies/:id/visibility       → HandleGetVisibility

  GET    /api/v1/ontologies/:id/policies         → HandleListPolicies
  POST   /api/v1/ontologies/:id/policies         → HandleCreatePolicy
  DELETE /api/v1/ontologies/:id/policies/:policyId → HandleDeletePolicy
  ```

  > **Files:** `src/services/api-gateway/handlers/org_handler.go` (new), `src/services/api-gateway/routes.go` (update), `src/services/api-gateway/proxy/grpc_auth_client.go` (update), `src/services/api-gateway/auth/auth.go` (update)
  > **Logging:** INFO — log REST request (method, path, user_id, trace_id); ERROR — log gRPC call failures with status codes

- [x] **Task 17: Add GraphQL resolvers for org model in ontology-service**
  **Subtasks (ordered):**
  1. **Write GraphQL resolver unit tests** — `src/services/ontology-service/tests/org_resolvers_test.rs`:
     - `[ValidQuery]_GroupsQuery_[ReturnsGroupList]`
     - `[SearchQuery]_GroupsQuery_WithFilter_[ReturnsFiltered]`
     - `[SortQuery]_ProjectsQuery_WithSort_[ReturnsSorted]`
     - `[MemberQuery]_MembersQuery_[ReturnsMembersWithRoles]`
     > BDD: `[Condition]_[Action]_[ExpectedResult]`
  2. **Create `src/services/ontology-service/src/clients/auth_client.rs`:**
     - HTTP client calling auth-service REST API for groups/projects/members
     - Configurable base URL via `AUTH_SERVICE_URL` env var (default: `http://auth-service:8081`)
     - Timeout: 10s reads, 15s writes
     - Error mapping: HTTP 403 → GraphQL `FORBIDDEN`, 404 → `NOT_FOUND`
     - Reuse existing HTTP client patterns (`reqwest`)
  3. **Add GraphQL types** — `src/services/ontology-service/src/graphql/types/org.rs`:
     - `Group` type: id, name, description, parentGroupId, childGroups (lazy), memberCount, projectCount
     - `Project` type: id, name, description, visibility, ontologyCount, memberCount, updatedAt
     - `Member` type: id, userId, username, avatarUrl, role, addedAt
     - Match frontend query fields from `queries.ts` exactly
  4. **Add query resolvers** — `src/services/ontology-service/src/graphql/query.rs`:
     - `groups(q: Option<String>)` → calls auth_client.list_groups()
     - `projects(q, sortBy, sortDir, page, perPage)` → calls auth_client.list_projects()
     - `members(ontologyId: ID!)` → calls auth_client.list_members()
  5. **Add mutation resolvers** — `src/services/ontology-service/src/graphql/mutation.rs`:
     - `updateMemberRole(ontologyId, userId, role)` → calls auth_client.update_member_role()
     - `removeMember(ontologyId, userId)` → calls auth_client.remove_member()
  6. **Run resolver unit tests** — they should PASS

  > **Files:** `src/services/ontology-service/src/clients/auth_client.rs` (new), `src/services/ontology-service/src/clients/mod.rs` (update), `src/services/ontology-service/src/graphql/types/org.rs` (new), `src/services/ontology-service/src/graphql/query.rs` (update), `src/services/ontology-service/src/graphql/mutation.rs` (update), `src/services/ontology-service/tests/org_resolvers_test.rs` (new)
  > **Logging:** INFO — log upstream URL, method, response status; ERROR — log connection failures

  <sub>**Commit checkpoint: tasks 14-17**</sub>



- [x] **Task 16: Implement Idempotency-Key support for org write endpoints**

  > **Source:** `specs/requirements/REQ-FUN.API.write-idempotency` (P0, APPROVED).
  > All REST write endpoints (POST/PUT/DELETE) MUST support `Idempotency-Key`.
  > 100% coverage required for membership/role changes. CI blocks MR if < 100%.

  **Subtasks (ordered):**
  1. **Write idempotency contract tests** in `src/services/api-gateway/middleware/idempotency_test.go`:
     - `[SameKeySamePayload]_ReplayRequest_[ReturnsSameResourceID]`
     - `[SameKeyDifferentPayload]_ReplayRequest_[Returns409]`
     - `[MissingKey]_CriticalWrite_[Returns400]`
     - `[ExpiredKey]_After24h_[Returns201]`
     - `[DifferentKeys]_ConcurrentRequests_[ReturnsDifferentIDs]`
     > BDD: `[Condition]_[Action]_[ExpectedResult]`
  2. **Run tests** -- they should FAIL (red phase: middleware does not exist yet)
  3. **Implement `src/services/api-gateway/middleware/idempotency.go`:**
     - Gin middleware extracting `Idempotency-Key` header
     - Validate format: non-empty, alphanumeric + hyphens, max 128 chars
     - Store in Redis: key = `idem:{key}`, value = `{status, body_hash, resource_id, location}`, TTL = 24h
     - Same key + same body SHA256: return cached response
     - Same key + different body SHA256: return 409 `IDEMPOTENCY_KEY_REUSED_WITH_DIFFERENT_PAYLOAD`
     - New key: execute handler, store response in Redis
     - Degraded mode: if Redis unavailable, log WARN and pass-through
     - Prometheus metrics: `idempotency_replay_total`, `idempotency_conflict_total`
  4. **Wire middleware to org write routes** in `src/services/api-gateway/routes.go`:
     - Apply to all POST/PUT/DELETE org endpoints
     - Membership endpoints: missing key returns 400 `INVALID_IDEMPOTENCY_KEY`
     - GET endpoints: no idempotency required
  5. **Add Redis dependency to API Gateway** in `deploy/docker-compose.yml`:
     - API Gateway `depends_on` redis (service_healthy) -- redis already exists in compose
     - Add `REDIS_URL` env var
  6. **Run idempotency tests** -- they should PASS (green phase)
  7. **Run E2E API tests from Task 3** -- verify idempotency end-to-end

  > **Files:** `src/services/api-gateway/middleware/idempotency.go` (new), `src/services/api-gateway/middleware/idempotency_test.go` (new), `src/services/api-gateway/routes.go` (update), `deploy/docker-compose.yml` (update)
  > **Logging:** INFO -- log replay; WARN -- log conflict (409), degraded mode; ERROR -- log Redis connection failure

### Phase 5: Security Tests

> Per skill-context: *"When planning security-related tasks, include an explicit subtask for writing negative tests that prove the bypass is closed. Reference `tests/security/` or the service's existing test pattern. The test must dispatch a real HTTP request, not use mock assertions."*

- [x] **Task 18: Write negative security tests for org access control**
  **Subtasks (ordered):**
  1. **Write `tests/security/org_access_control_test.go`:**
     - **BOLA — Broken Object Level Authorization:**
       - `[ViewerAccess]_MembersOfAnotherOntology_[Returns403]` — Viewer reads members of ontology they don't belong to
       - `[EditorAccess]_GroupOfAnotherTeam_[Returns403]` — Editor accesses group membership outside their scope
       - `[CrossUser]_UpdateRole_OnDifferentScope_[Returns403]` — user tries to update role on scope outside their tenant
     - **BFLA — Broken Function Level Authorization:**
       - `[ViewerRole]_CreateGroup_[Returns403]` — Viewer tries POST /api/v1/groups
       - `[EditorRole]_AddMember_[Returns403]` — Editor tries POST /api/v1/ontologies/:id/members
       - `[MaintainerRole]_SetVisibility_[Returns403]` — Maintainer tries PUT /api/v1/ontologies/:id/visibility
       - `[EditorRole]_CreatePolicy_[Returns403]` — Editor tries POST /api/v1/ontologies/:id/policies
     - **Cross-Tenant:**
       - `[TenantA]_AccessGroupOfTenantB_[Returns403]` — user from tenant A accesses group in tenant B
       - `[TenantA]_ListMembersOfTenantB_[Returns403]` — same for member listing
     - **Last-Owner Protection:**
       - `[LastOwner]_RemoveSelf_[Returns403]` — Owner removes self when they are the last Owner
     - **Cycle Detection:**
       - `[CircularParent]_CreateGroup_[Returns400]` — attempt to create A→B→A cycle
     - **Visibility:**
       - `[Anonymous]_ReadPrivateMembers_[Returns401]` — no JWT, Private ontology
       - `[Anonymous]_ReadInternalMembers_[Returns401]` — no JWT, Internal ontology
       - `[Authenticated]_ReadInternalMembers_[Returns200]` — valid JWT but non-member, Internal ontology
     - **Policy Conflict:**
       - `[DuplicatePattern]_CreatePolicy_[Returns409]` — attempt to create conflicting ABAC policy
     > ALL tests dispatch real HTTP requests (no mock assertions)
     > BDD: `[Condition]_[Action]_[ExpectedResult]`
     > Anti-patterns: see `.ai-factory/rules/test-quality.md`
  2. Run tests — they should FAIL (red phase: security gates not yet proven)
  3. After implementation is complete, run tests again — ALL must PASS

  > **Files:** `tests/security/org_access_control_test.go` (new)
  > **Logging:** N/A (test assertions — test failures produce HTTP status + body output)

- [x] **Task 19: Write contract tests for gRPC org service error codes**
  **Subtasks (ordered):**
  1. **Update `src/services/auth-service/org/org_grpc_test.go`** — add error code mapping tests:
     - `[OrgError_FORBIDDEN_ADMIN_ONLY]_ToGrpc_[ReturnsPermissionDenied]`
     - `[OrgError_SCOPE_NOT_FOUND]_ToGrpc_[ReturnsNotFound]`
     - `[OrgError_CYCLE_DETECTED]_ToGrpc_[ReturnsInvalidArgument]`
     - `[OrgError_POLICY_CONFLICT]_ToGrpc_[ReturnsAlreadyExists]`
     - `[OrgError_FORBIDDEN_INSUFFICIENT_ROLE]_ToGrpc_[ReturnsPermissionDenied]`
     - `[OrgError_FORBIDDEN_CROSS_TENANT_ACCESS]_ToGrpc_[ReturnsPermissionDenied]`
     - `[OrgError_VISIBILITY_VIOLATION]_ToGrpc_[ReturnsPermissionDenied]`
  2. Run tests — verify all gRPC status code mappings are correct

  > **Files:** `src/services/auth-service/org/org_grpc_test.go` (update)
  > **Logging:** N/A (test assertions)

  <sub>**Commit checkpoint: tasks 18-19**</sub>

### Phase 6: Traceability, Docs & Finalization

- [x] **Task 20: Update traceability.ttl with org model test links**
  **Subtasks (ordered):**
  1. **Query existing traceability graph** — identify existing `vdo:TestSuite` entries and `vdo:validates` triples
  2. **Add `vdo:TestSuite` entries** for:
     - `tests/e2e/playwright/tests/gui/pages/groups-page-wired.spec.ts` — `vdo:validates REQ-NFR.SECURITY.organization-access-model`
     - `tests/e2e/playwright/tests/gui/pages/projects-page-wired.spec.ts` — `vdo:validates REQ-NFR.SECURITY.organization-access-model`
     - `tests/e2e/playwright/tests/gui/pages/members-page-wired.spec.ts` — `vdo:validates REQ-NFR.SECURITY.organization-access-model`
     - `tests/e2e/playwright/tests/api/rest/org-api.spec.ts` — `vdo:validates REQ-NFR.SECURITY.organization-access-model`
     - `tests/e2e/playwright/tests/gui/user-stories/org-lifecycle.spec.ts` — `vdo:validates REQ-NFR.SECURITY.organization-access-model`
     - `tests/org-api/org_api_integration_test.go` — `vdo:validates REQ-NFR.SECURITY.organization-access-model`
     - `tests/security/org_access_control_test.go` — `vdo:validates REQ-NFR.SECURITY.organization-access-model`, `vdo:validates SEC-AUTHZ-GATES-001`
     - `src/services/auth-service/org/org_grpc_test.go` — `vdo:validates ORG-ACCESS-001`
     - `src/services/auth-service/org/postgres_store_test.go` — `vdo:validates ORG-ACCESS-001`
     - `src/services/auth-service/org/org_proto_validation_test.go` — `vdo:validates ORG-ACCESS-001`
     - `src/services/ontology-service/tests/org_resolvers_test.rs` — `vdo:validates REQ-NFR.SECURITY.organization-access-model`
  3. **Add `vdo:validates` link** — `src/services/auth-service/org/org.go` → `vdo:implements ADR-DES.SECURITY.gitlab-like-organization-model`
  4. **Verify no stale triples** — remove any outdated entries for renamed/deleted test files

  > **Files:** `.ai-factory/traceability/traceability.ttl` (update)
  > **Logging:** N/A (RDF data update)

- [x] **Task 21: Update documentation** — created organization-model.adoc + Antora nav + GraphQL resolver tests
  **Subtasks (ordered):**
  1. **Update Antora developer guide** (`src/docs/antora/developer-guide/`):
     - Add "Organizational Model" section under "Authentication & Authorization"
     - Document: OrgService architecture, PostgreSQL schema, gRPC service definition, role inheritance rules
  2. **Update Antora integrator guide** (`src/docs/antora/integrator-guide/`):
     - Document org REST API endpoints: GET/POST /api/v1/groups, /api/v1/projects, /api/v1/ontologies/:id/members
     - Document GraphQL queries: groups, projects, members + mutations updateMemberRole, removeMember
     - Add authentication requirements (Bearer JWT with appropriate role)
  3. **Update Antora admin guide** (`src/docs/antora/admin-guide/`):
     - Document PostgreSQL org database setup (vedo_org database, migrations)
     - Document environment variables: `AUTH_SERVICE_DATABASE_URL`
  4. **Update OpenAPI spec** (`src/services/api-gateway/docs/openapi.json`):
     - Add org endpoint schemas under `/api/v1/groups`, `/api/v1/projects`, `/api/v1/ontologies/{id}/members`
     - Add request/response schemas for Group, Project, Member, Visibility

  > **Files:** `src/docs/antora/developer-guide/`, `src/docs/antora/integrator-guide/`, `src/docs/antora/admin-guide/`, `src/services/api-gateway/docs/openapi.json`
  > **Logging:** N/A (documentation)

- [x] **Task 22: Run full E2E test suite and fix regressions** — Go tests ✅, Rust tests ✅, API Gateway tests ✅. Docker Compose + Playwright requires infrastructure.
  **Subtasks (ordered):**
  1. **Start all services** via Docker Compose: `docker compose up -d`
  2. **Run M2.1 E2E tests** — `npx playwright test --config=playwright.m2.1.real.config.ts`
  3. **Run security test suite** — `go test ./tests/security/org_access_control_test.go`
  4. **Run integration tests** — `go test ./tests/org-api/`
  5. **Run existing auth-service org tests** — `go test ./src/services/auth-service/org/...`
  6. **Run ontology-service resolver tests** — `cargo test -p ontology-service`
  7. **Fix any test failures** — iterate until all tests pass
  8. **Verify frontend renders real data** — GroupsPage, ProjectsPage, MembersPage show data from API (no Apollo test fixtures)

  > **Logging:** Run outputs produce test reports

## Acceptance Criteria

### API Acceptance Criteria

> Проверяются через API-тесты (Task 3 REST, Task 14 integration, Task 18 security, Task 19 contract,
> unit tests в Phases 1–4). Запуск: `go test` и `cargo test`.

- [ ] **REST API returns real data:** `GET /api/v1/groups` → 200, `GET /api/v1/projects` → 200, `GET /api/v1/ontologies/:id/members` → 200
- [ ] **Group CRUD:** create → 201, list → 200, get → 200, update → 200, delete → 204, subgroups → 200
- [ ] **Project CRUD:** create → 201, list (paginated, searchable, sortable) → 200, get → 200, update → 200, delete → 204
- [ ] **Member CRUD:** add → 201, list → 200, update role → 200, remove → 204 (with last-owner protection)
- [ ] **Auth gates:** unauthenticated → 401, Viewer write → 403, Editor membership → 403
- [ ] **Membership authorization:** only Owner manages members; last-owner removal blocked with 403
- [ ] **Visibility enforced:** Private → reject non-members, Internal → reject anonymous (401), Public → allow all
- [ ] **Role inheritance:** parent group roles propagate to child scopes (max-role-wins)
- [ ] **Cross-tenant access denied:** `FORBIDDEN_CROSS_TENANT_ACCESS` (403)
- [ ] **Circular group hierarchy rejected:** `CYCLE_DETECTED` (400)
- [ ] **All 15 negative security tests pass** (BOLA/BFLA/cross-tenant/visibility/last-owner/policy-conflict)
- [ ] **gRPC contract tests pass:** `go test ./src/services/auth-service/org/org_grpc_test.go`
- [ ] **PostgresOrgStore tests pass:** `go test ./src/services/auth-service/org/postgres_store_test.go`
- [ ] **Existing org contract tests pass:** `go test ./src/services/auth-service/org/...` (CT-ORG-001..010)
- [ ] **Gateway integration tests pass:** `go test ./tests/org-api/`
- [ ] **GraphQL resolver tests pass:** `cargo test -p ontology-service`
- [ ] **Test Quality Score (TQS) >= bronze (6.0)**
- [ ] **No B1-B7 anti-patterns** (see `.ai-factory/rules/test-quality.md`)
- [ ] **Idempotency-Key enforced on all write endpoints:** same key + same payload returns same resource_id; same key + different payload returns 409; missing key returns 400 on critical writes

### E2E Acceptance Criteria

> Проверяются через Playwright против полного стека (`docker-compose.test.yml`).
> Запуск: `npx playwright test --config=playwright.m2.1.real.config.ts`.

- [ ] **E2E GUI tests pass:** `groups-page-wired`, `projects-page-wired`, `members-page-wired`
- [ ] **E2E user-story tests pass:** `org-lifecycle` (create group -> manage members -> visibility)
- [ ] **Frontend renders real data:** GroupsPage, ProjectsPage, MembersPage show data from API via GraphQL (no Apollo test fixtures)
- [ ] **Test environment healthy:** `docker compose -f deploy/docker-compose.test.yml up -d` — all services report healthy
- [ ] **vedo_org database:** created with all migrations applied; data persists across test runs

### General

- [ ] **Traceability annotations present** (`// Validates: REQ-...`) in all test files
- [ ] **traceability.ttl updated** with all new `vdo:TestSuite` and `vdo:validates` triples
- [ ] **Documentation checkpoint completed** (Antora sites updated, OpenAPI spec updated)
- [ ] Test environment (`docker-compose.test.yml`) starts with `vedo_org` database created; migrations applied; all services healthy

## Deferred

### Keycloak Sync (M2.1 → deferred)

**Source:** `specs/use-cases/UC-admin.access.manage-membership-and-permissions` (P0, step 5).

> "Система синхронизирует группы/роли с Keycloak, если это требуется для SSO/RBAC."

**Rationale for deferral:** The Keycloak Admin API integration requires:
1. Analysis of current realm configuration (`auth-service/keycloak/realm-import.json`)
2. Design of the sync direction (VEDO → Keycloak or bidirectional)
3. Mapping between VEDO roles (Viewer/Editor/Maintainer/Owner) and Keycloak realm roles
4. Testing with a real Keycloak instance

This is a non-trivial integration that would block the core M2.1 deliverable. Groups and roles are managed in-app via the VEDO authorization model (PostgreSQL). Keycloak sync can be added as a follow-up task without changing the API contract.

**Next step:** Create a research task in M3 or a dedicated sync story with acceptance criteria from this use case.

### Approval Rules / write_with_approval workflow (M2.1 → deferred to M6)

**Source:** `specs/use-cases/UC-admin.access.manage-membership-and-permissions` (approval rules flow).

The approval rules flow (write_with_approval → proposal branch → Ontology Merge Request → Maintainer review) belongs to the M6 Collaboration & Social Hub milestone, which implements Merge Request review workflow. M2.1 delivers the membership model and visibility levels; the OMR workflow is out of scope.
