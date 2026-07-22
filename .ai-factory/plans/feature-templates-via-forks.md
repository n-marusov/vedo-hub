# Implementation Plan: Templates via Forks

Branch: feature/templates-via-forks
Created: 2026-07-22

## Settings
- Testing: yes
- Logging: verbose
- Docs: yes

## Roadmap Linkage
Milestone: "M5: MVP Scope Gap Closure"
Rationale: Replaces the MVP template flow with Demos+Forks and moves F13.1 (forks) from post-MVP into MVP scope.

## Research Context
Source: .ai-factory/RESEARCH.md (Active Summary, session 2026-07-22)

Goal: Remove Templates BC, replace with `VEDO Demos` group + fork mechanism. F13.1 Forks moves to MVP.

Decisions:
1. Templates BC removed. Ontology templates = demo projects in `VEDO Demos` group + fork. F13.1 Forks moves from post-MVP to MVP.
2. Fork model: `upstream_project_id` on Project (nullable, Git-style). `forks_count` = count query or Redis cache. Endpoint: `POST /api/v1/projects/{id}/fork`.
3. Fork Saga: Organization create Project + Versioning copy branch + Ontology materialize. Orchestrator = auth-service org.go (not API Gateway — anti-pattern "leaky gateway").
4. Lose: semver for templates (git tags richer), auto-deprecate by usage (premature automation), personal-catalog-UX (via Group "My Templates"). Gain: -1 BC, -1 Saga, -1 Published Language, F13.1 in MVP, community PR via MR flow.

Open questions:
- Fork Saga compensating action details (Task 11)
- Whether Guest role can fork (proposal: yes — read access to source is sufficient, fork creates new Project in user space)

## Current State (verified in code, not specs)

| Component | Status | Action |
|-----------|--------|--------|
| `src/services/api-gateway/handlers/template_handler.go` | Implemented, loads 0 templates | Delete |
| `src/services/api-gateway/internal/templates/ontologies/` | Empty (only README.md) | Delete |
| `routes.go` L188-189: `GET /templates/ontologies`, `POST /ontologies/:id/apply-template` | Working routes to empty handler | Delete |
| `models.TemplateSummary`, `models.ApplyTemplateResponse` | Used only in template_handler | Delete after verification |
| `ontologyv1.ApplySequenceRequest` proto | Used by document-extractor + ai-orchestration | **DO NOT DELETE** — Published Language |
| 9 REQ drafts `templates-*` | Drafts, not implemented | Delete |
| Frontend template GUI | Not implemented | Nothing to delete, add fork UI |
| `auth-service/migrations/007-009` | Done: `scopes.type='project'`, `ontologies` 1:1 | Ready, add migration 010 |
| `org_handler.go` scope `"project/" + id` | Done | Ready, add `HandleForkProject` |
| `specs/vision.md` F14.5, F14.8, F13.1 | Vision-level, F13.1 in post-MVP | Update: F13.1 -> MVP |
| `specs/glossary.md` "Ontology Template" | Describes metadata.json + semver | Replace with "Demo Project" + "Ontology Fork" |
| `ROADMAP.md` M15 "fork/versioning APIs" | Forks in post-1.0 | Move F13.1 fork to M5 |

## Commit Plan

- **Commit 1** (after tasks 1-4): "feat: remove template handler, models, and REQ drafts"
- **Commit 2** (after tasks 5-9): "docs: update specs, glossary, ROADMAP, ADR for templates-via-forks"
- **Commit 3** (after task 10): "docs: add fork US/UC requirements"
- **Commit 4** (after task 11): "test: write fork API contract tests (RED phase)"
- **Commit 5** (after tasks 12-15): "feat: implement fork infrastructure with module tests (GREEN phase)"
- **Commit 6** (after tasks 16-17): "feat: add VEDO Demos seed data and Pencil design artifacts"
- **Commit 7** (after tasks 18-19): "feat: add fork UI with GUI tests (RED before code)"
- **Commit 8** (after tasks 20-21): "docs: add fork traceability and Antora docs"

## Tasks

### Phase 1: Cleanup — Remove Templates

- [x] **Task 1: Remove template_handler and routes**
  Delete the template handler, its test, the empty templates directory, and the two template routes from routes.go.

  - Delete `src/services/api-gateway/handlers/template_handler.go`
  - Delete `src/services/api-gateway/handlers/template_handler_test.go`
  - Delete `src/services/api-gateway/internal/templates/` (entire directory with README.md)
  - In `src/services/api-gateway/routes.go` remove the two template route lines: `api.GET("/templates/ontologies", ...)` and `api.POST("/ontologies/:id/apply-template", ...)`
  - Remove the `templateHandler` instantiation in `RegisterRoutes` if present
  - Verify `aiOrchProxy.HandleListTemplates` / `HandleApplyTemplate` references are removed from routes.go

  LOGGING: No runtime logging needed — pure deletion.

  Files: `src/services/api-gateway/handlers/template_handler.go`, `src/services/api-gateway/handlers/template_handler_test.go`, `src/services/api-gateway/internal/templates/`, `src/services/api-gateway/routes.go`
  Verify: `go build ./...` passes, `go test ./src/services/api-gateway/...` passes (minus 1 test file)

- [x] **Task 2: Remove models.TemplateSummary and models.ApplyTemplateResponse**
  Remove template-specific model types after verifying they have no other consumers.

  - Grep for `TemplateSummary` and `ApplyTemplateResponse` across the entire `src/` tree to confirm `template_handler.go` is the only consumer
  - Delete the type definitions from `src/services/api-gateway/models/` (locate exact file by grep)
  - If any other file references these types, update or remove the reference

  LOGGING: No runtime logging — pure deletion.

  Files: `src/services/api-gateway/models/*.go` (exact files by grep)
  Verify: `go build ./...` passes

- [x] **Task 3: Delete 9 REQ draft files for templates**
  Remove all template-related requirement draft files.

  - Delete:
    - `specs/requirements/REQ-FUN.API.templates-definition.md`
    - `specs/requirements/REQ-FUN.API.templates-catalog.md`
    - `specs/requirements/REQ-FUN.API.templates-versioning.md`
    - `specs/requirements/REQ-FUN.API.templates-import.md`
    - `specs/requirements/REQ-FUN.DATA.templates-metadata.md`
    - `specs/requirements/REQ-FUN.OPS.templates-ownership.md`
    - `specs/requirements/REQ-FUN.OPS.templates-review.md`
    - `specs/requirements/REQ-FUN.OPS.templates-removal.md`
    - `specs/requirements/REQ-NFR.OPS.templates-analytics.md`
    - `specs/requirements/REQ-USR.UI.templates-apply.md`
    - `specs/requirements/REQ-USR.UI.templates-preview.md` (if exists)
    - `specs/requirements/REQ-USR.UI.templates-search.md` (if exists)
    - `specs/requirements/REQ-USR.UI.templates-notifications.md` (if exists)

  Files: `specs/requirements/REQ-*.templates-*.md`
  Verify: `grep -r "templates-" specs/requirements/` returns empty (cross-references in other REQs cleaned in Task 5)

- [x] **Task 4: Remove template-related user-stories and use-cases**
  Delete template-specific US/UC files and update the matrix.

  - Delete `specs/user-stories/US-io.forms.configure-templates.md` if it exists
  - Delete `specs/use-cases/UC-io.forms.configure-import-form-templates.md` if it exists
  - Update `specs/user-stories/_matrix.md` — remove rows referencing templates
  - Update `specs/use-cases/README.md` if it lists template use cases

  Files: `specs/user-stories/`, `specs/use-cases/`

<!-- Commit checkpoint: tasks 1-4 -->

### Phase 2: Specs and ADR — Formalize Decision

- [x] **Task 5: Update specs/vision.md**
  Rewrite template-related sections to reflect Demos+Forks model.

  - F14.5 "Ontology Templates" -> rephrase: "Demos Group — 5 demo projects in `VEDO Demos` group (Organization, Product, Process, Glossary, Event) that users fork via `POST /api/v1/projects/{id}/fork`. Semver/catalog/metadata.json are not used — replaced by git-tags and forks_count."
  - F14.8 "Custom LLM prompt" -> keep (this is about document-extractor, not templates)
  - F13.1 "Ontology Forks" -> move from post-MVP section into MVP section (2.5); state that fork = copy Project + Ontology + upstream_project_id link
  - F13 description (L156) — keep "Social Hub" but note F13.1 is in MVP, rest is post-MVP
  - Domain events (L606): `OntologyForked` — keep, mark as MVP event
  - Remove references to `metadata.json`, `template.owl`, `semver`, `usage_count`, `last_reviewed_at`, `status (active/deprecated/archived)` from template descriptions

  Files: `specs/vision.md`

- [x] **Task 6: Update specs/glossary.md**
  Replace "Ontology Template" with "Demo Project" and "Ontology Fork" terms.

  - Entry "Ontology Template" (L824-829) -> replace with:
    - "Demo Project" — public Project in `VEDO Demos` group, used as starting point via fork. Replaces the former `Ontology Template` concept (semver/catalog/metadata.json) with git-tags + forks_count + visibility=public.
    - "Ontology Fork" — copy of a Project (with its paired Ontology) into the user's workspace, linked to the source via `upstream_project_id`. Analogous to `git fork`. The fork is private by default; the user becomes Owner of the new Project.
  - Entry "Project" (L666-675) -> add `upstream_project_id` as optional field: "When non-null, this Project is a fork of the referenced upstream Project. Null for original (non-forked) Projects."
  - Remove any references to `Ontology Template` semver, `metadata.json` 13-field schema, `usage_count`, `last_reviewed_at`, template review/deprecate lifecycle

  Files: `specs/glossary.md`

- [x] **Task 7: Update ROADMAP.md**
  Move F13.1 forks from M15 to M5.

  - M2 "Document-to-Ontology & Template Baseline" -> rename to "Document-to-Ontology & Demos Baseline" or add note: "Template mechanism replaced by Demos+Forks in M5 (see ADR-DES.PROCESS.templates-via-forks)"
  - M5 "MVP Scope Gap Closure" -> add bullet: "F13.1 Forks — `POST /api/v1/projects/{id}/fork`, `upstream_project_id` on Project, 5 demo projects in `VEDO Demos` group as seed data"
  - M15 "fork/versioning APIs, public profiles, stars/forks/issues" -> remove "fork/versioning APIs" (moved to M5), keep "public profiles, stars/issues, sponsoring, partner commissions"

  Files: `.ai-factory/ROADMAP.md`

- [x] **Task 8: Create ADR-DES.PROCESS.templates-via-forks.md**
  New ADR documenting the decision to replace Templates BC with Demos+Forks.

  - Status: Accepted
  - Date: 2026-07-22
  - Context: Templates BC with semver/catalog/metadata.json was premature automation for 5 MVP templates; fork = simpler and unified mechanics for both templates and social hub
  - Decision: Templates BC removed; `VEDO Demos` group + 5 demo Projects + `POST /api/v1/projects/{id}/fork` + `upstream_project_id` on Project
  - Alternatives considered: (a) Templates BC with catalog UX (premature), (b) Templates as extension of Ontology BC (lifecycle pollution), (c) Personal templates namespace (Group "My Templates" is simpler)
  - Consequences: -1 BC, -1 Saga, F13.1 in MVP. Lose: semver (git tags richer), auto-deprecate (premature), personal-catalog-UX (via Group). Gain: unified fork mechanics, community PR via MR flow
  - Related ADRs: `ADR-DES.SECURITY.gitlab-like-organization-model` (fork = Project copy), `ADR-DES.API.organization-rest-endpoints` (fork endpoint extension)

  Files: `specs/adr/ADR-DES.PROCESS.templates-via-forks.md`

- [x] **Task 9: Extend ADR-DES.API.organization-rest-endpoints with fork endpoint**
  Add fork to the canonical REST contract.

  - Add to canonical paths table: `POST /api/v1/projects/{id}/fork` — create fork (copy Project + Ontology + upstream link)
  - Fork semantics: new Project with `upstream_project_id = <source>`, `visibility = private` (fork is private by default), Ontology copied via Versioning BC (copy main branch), initial commit message "Initial fork from <source>"
  - RBAC: any user with read access to source Project can fork (public/internal — any authenticated user; private — only members). Guest role can fork (read access to source is sufficient; fork creates new Project in user space, does not modify source)
  - Idempotency: `Idempotency-Key` header required (same as other write-endpoints)
  - Response: `201 Created` with `{project_id, ontology_id, upstream_project_id}`
  - Audit event: `project.forked`, `object_id = new_project_id`, `reason = "forked from <source_project_id>"`

  Files: `specs/adr/ADR-DES.API.organization-rest-endpoints.md`

<!-- Commit checkpoint: tasks 5-9 -->

### Phase 3: US/UC First (TDD — requirements before tests)

Write User Stories and Use Cases for the fork feature BEFORE any tests or implementation. These define the expected behavior that tests must verify.

- [x] **Task 10: F13.1 User Story and Use Case**
  Create formal US/UC for the fork feature.

  - Create `specs/user-stories/US-projects.fork.md`:
    ```gherkin
    @US-projects.fork @UC-projects.fork @P0 @organization @fork
    Feature: US-projects.fork Fork demo project into user workspace

      Background:
        Given user is authenticated as "Knowledge Engineer"
        And a public project "VEDO Demos/Organization" exists

      Scenario: Knowledge Engineer forks a demo project
        When user opens "VEDO Demos/Organization"
        And clicks "Fork"
        Then a new private project is created in user's personal space
        And the new project contains a copy of the "Organization" ontology
        And the new project has upstream_project_id = "VEDO Demos/Organization"
        And the user becomes Owner of the new project

      Scenario: User forks a private project without access
        When user attempts to fork a private project they don't have access to
        Then the system returns 403 Forbidden
        And no new project is created
    ```
  - Create `specs/use-cases/UC-projects.fork.md` — use case with actors, preconditions, main flow, alternative flows, postconditions
  - Update `specs/user-stories/_matrix.md` — add US-projects.fork row
  - Update `specs/use-cases/README.md` — add UC-projects.fork

  Files: `specs/user-stories/US-projects.fork.md`, `specs/use-cases/UC-projects.fork.md`, `specs/user-stories/_matrix.md`, `specs/use-cases/README.md`

<!-- Commit checkpoint: task 10 -->

### Phase 4: API Contract Tests (TDD — RED phase)

Write ALL fork API contract tests BEFORE any implementation. These tests define the contract and must fail (RED) until Phase 4 implements the infrastructure.

- [x] **Task 11: Write fork API contract tests (RED phase)**
  Define the fork endpoint contract through tests. All tests must fail until implementation in Phase 4.

  **10a. Security negative tests for fork endpoint (real HTTP, no mocks):**
  Create `tests/security/fork_bola_test.go`:
  - **Test A (cross-tenant BOLA):** user without read access to source Project (private, different tenant) -> fork returns 403, NOT 404 (per `REQ-NFR.SECURITY.bola-bfla-negative-tests` — response is always 403, never 404)
  - **Test B (cross-object BOLA):** user with read access to Project A attempts to fork Project B (private, no access) -> 403
  - **Test C (BFLA):** user with Guest role (read-only) forks a public project -> 201 (fork is allowed for Guest — read access to source is sufficient, fork creates new Project, does not modify source). Verify Guest without read access to private source -> 403.
  - **Test D (IDOR):** fork with predictable source ID (integer, sequential) -> always 403 without access, never 404
  - All tests dispatch real HTTP requests (per skill-context rule: "security tests must dispatch a real HTTP request, not use mock assertions")

  **10b. Acceptance test for HandleForkProject handler:**
  Extend `src/services/api-gateway/handlers/org_handler_test.go`:
  - Assert 201 Created with correct response body shape `{project_id, ontology_id, upstream_project_id}`
  - Assert 403 for unauthorized access
  - Assert 503 when versioning service unavailable
  - Assert Idempotency-Key header is required (400 if missing for write ops)

  **10c. Unit test for ForkProject RPC (auth-service):**
  Create `src/services/auth-service/org/fork_test.go`:
  - Mock membership resolver for read access check
  - Assert new Project scope created with `type='project'`, `upstream_project_id = source`
  - Assert paired Ontology row created (1:1)
  - Assert caller receives Owner role
  - Assert Versioning Service gRPC `CopyBranch` called
  - Assert compensating action on Versioning failure (new Project + Ontology deleted)

  **10d. OpenAPI spec contract test:**
  Add test to verify:
  - `POST /api/v1/projects/{id}/fork` present in OpenAPI
  - Request schema: empty body, `Idempotency-Key` header required
  - Response 201 schema: `{project_id, ontology_id, upstream_project_id}`
  - Response 403 schema present
  - `GET /api/v1/templates/ontologies` and `POST /api/v1/ontologies/{id}/apply-template` removed
  - `ProjectDetail.upstream_project_id` (nullable string) present
  - `ProjectDetail.forks_count` (integer) present

  LOGGING (verbose):
  - Tests log `trace_id` for each real HTTP request
  - Assertions include full response body in failure messages

  Files: `tests/security/fork_bola_test.go`, `src/services/api-gateway/handlers/org_handler_test.go`, `src/services/auth-service/org/fork_test.go`, `src/services/api-gateway/docs/openapi.json` test
  Acceptance criteria:
  - [ ] All tests fail (RED) — contract defined but not implemented
  - [ ] Security tests dispatch real HTTP requests (no mocks)
  - [ ] Test Quality Score (TQS) >= bronze (6.0) for all new test files
  - [ ] No B1-B7 anti-patterns (see `.ai-factory/rules/test-quality.md`)
  - [ ] Traceability annotations present (`// Validates: REQ-NFR.SECURITY.bola-bfla-negative-tests`)
  > BDD naming: `[UnauthorizedUser]_[ForkPrivateProject]_[Returns403]`
  > New tests must meet minimum quality (TQS >= bronze)

<!-- Commit checkpoint: task 11 -->

### Phase 5: Fork Infrastructure (TDD — GREEN phase)

Implement infrastructure that makes the API contract tests pass. Each task: write module test FIRST, then implement.

- [x] **Task 12: upstream_project_id migration — test first, then SQL**
  **Step 11a — module test:** Write test asserting migration 010 creates the `upstream_project_id` column idempotently, with correct FK (`ON DELETE SET NULL`) and partial index.
  **Step 11b — implementation:** Create migration SQL and update `postgres_store.go`.

  - Create `src/services/auth-service/migrations/010_add_upstream_project_id.sql`:
    ```sql
    ALTER TABLE scopes ADD COLUMN IF NOT EXISTS upstream_project_id TEXT REFERENCES scopes(id) ON DELETE SET NULL;
    CREATE INDEX IF NOT EXISTS idx_scopes_upstream ON scopes(upstream_project_id) WHERE upstream_project_id IS NOT NULL;
    ```
  - `ON DELETE SET NULL`: when upstream Project is deleted, fork remains (orphan fork — acceptable per Git model)
  - Update `src/services/auth-service/org/postgres_store.go` `runMigrations()` — add migration 010 to explicit ordering list (after 007)

  LOGGING:
  - `slog.Info("migrations.applied", "migration", "010_add_upstream_project_id")` on success
  - `slog.Error("migrations.failed", "migration", "010", "error", err)` on failure

  Files: `src/services/auth-service/migrations/010_add_upstream_project_id.sql`, `src/services/auth-service/org/postgres_store.go`, `src/services/auth-service/org/postgres_store_test.go`
  Verify: migration applies idempotently, `go test ./src/services/auth-service/org/...` passes
  > BDD naming: `[Migration]_[Apply]_[UpstreamProjectIdAdded]`

- [x] **Task 13: ForkProject RPC — test first, then implement**
  **Step 12a — module test:** Write unit test for `ForkProject` in `src/services/auth-service/org/fork_test.go` building on the contract test from Task 10c, adding edge cases: duplicate fork, fork of non-existent project, Versioning Service timeout.
  **Step 12b — implementation:** Implement `ForkProject` RPC in auth-service with Saga orchestration.

  - Add to proto (`src/services/shared/proto/auth/v1/`):
    ```protobuf
    rpc ForkProject(ForkProjectRequest) returns (ForkProjectResponse);
    message ForkProjectRequest { string source_project_id = 1; string token = 2; }
    message ForkProjectResponse { string project_id = 1; string ontology_id = 2; string upstream_project_id = 3; }
    ```
  - Implement `ForkProject` in `src/services/auth-service/org/org.go`:
    1. Verify read access to source Project (via membership resolver)
    2. Create new Project scope (`type='project'`, `upstream_project_id = source`, `visibility='private'`)
    3. Create paired Ontology row in `ontologies` (1:1, `ontology_id` = new UUID)
    4. Grant caller Owner role on new Project (via membership)
    5. **Saga step:** call Versioning Service gRPC `CopyBranch(source_ontology_id, "main", new_ontology_id)` — copy main branch
    6. If Versioning succeeds: audit event `project.forked`, return `ForkProjectResponse`
    7. If Versioning fails: **compensating action** — delete new Project scope + Ontology row (cascade via `ON DELETE CASCADE` on `ontologies.project_scope`), audit event `project.fork_failed`, return error
  - Regenerate proto Go stubs

  LOGGING (verbose):
  - `slog.Debug("fork.start", "source", source_project_id, "user", userID, "trace_id", traceID)`
  - `slog.Info("fork.project_created", "new_project", newProjectID, "upstream", source_project_id)`
  - `slog.Debug("fork.versioning_copy", "source_ontology", sourceOntologyID, "target_ontology", newOntologyID)`
  - `slog.Info("fork.completed", "new_project", newProjectID, "ontology", newOntologyID, "duration_ms", elapsed)`
  - `slog.Error("fork.versioning_failed", "source", source_project_id, "new_project", newProjectID, "error", err)` — before compensating action
  - `slog.Warn("fork.compensated", "new_project", newProjectID)` — after rollback
  - `slog.Error("fork.compensation_failed", "new_project", newProjectID, "error", compErr)` — if cleanup also fails (critical, manual intervention needed)

  Files: `src/services/shared/proto/auth/v1/*.proto`, `src/services/auth-service/org/org.go`, `src/services/auth-service/org/postgres_store.go`, `src/services/auth-service/org/types.go`, `src/services/auth-service/org/fork_test.go`
  Verify: `go test ./src/services/auth-service/org/...` — Task 10c tests now PASS (GREEN)
  > BDD naming: `[AuthorizedUser]_[ForkProject]_[NewProjectWithUpstreamCreated]`
  > New tests must meet minimum quality (TQS >= bronze)

- [x] **Task 14: HandleForkProject handler — test first, then implement**
  **Step 13a — module test:** Write unit test for `HandleForkProject` in `src/services/api-gateway/handlers/org_handler_test.go` building on the contract test from Task 10b, adding edge cases: malformed request body, malformed project ID, gRPC deadline exceeded.
  **Step 13b — implementation:** Implement handler and route.

  - In `src/services/api-gateway/routes.go` add inside `orgWrite` group (with Idempotency middleware):
    ```go
    orgWrite.POST("/projects/:id/fork", orgHandler.HandleForkProject)
    ```
  - In `src/services/api-gateway/handlers/org_handler.go` add `HandleForkProject`:
    - Extract `source_project_id = c.Param("id")`, JWT token
    - Call `h.orgClient.ForkProject(ctx, &authv1.ForkProjectRequest{SourceProjectId: sourceID, Token: token})`
    - Return `201 Created` with `{project_id, ontology_id, upstream_project_id}`
    - Error handling: 403 (no read access), 404 (source not found — but per BOLA policy, return 403 not 404 for unauthorized), 503 (versioning service unavailable), 500 (internal error)

  LOGGING:
  - `slog.Info("fork.endpoint", "source", sourceID, "trace_id", traceID, "user", userID, "duration_ms", elapsed)`
  - `slog.Warn("fork.endpoint.forbidden", "source", sourceID, "trace_id", traceID)` — on 403
  - `slog.Error("fork.endpoint.failed", "source", sourceID, "trace_id", traceID, "error", err)` — on 5xx

  Files: `src/services/api-gateway/routes.go`, `src/services/api-gateway/handlers/org_handler.go`, `src/services/api-gateway/handlers/org_handler_test.go`
  Verify: `go test ./src/services/api-gateway/handlers/...` — Task 10b tests now PASS (GREEN)

- [x] **Task 15: OpenAPI spec — test first, then update**
  **Step 14a — contract test:** Extend the OpenAPI test written in Task 10d to also verify field types (string, not integer), required fields, and response code ranges.
  **Step 14b — implementation:** Update `openapi.json`.

  - Remove `GET /api/v1/templates/ontologies` path
  - Remove `POST /api/v1/ontologies/{id}/apply-template` path
  - Add `POST /api/v1/projects/{id}/fork`:
    - Request: empty body (source ID from path), `Idempotency-Key` header required
    - Response 201: `{project_id: string, ontology_id: string, upstream_project_id: string}`
    - Response 403: standard error (no read access to source)
    - Response 503: standard error (versioning service unavailable)
  - Update `ProjectDetail` schema: add `upstream_project_id` (nullable string, "ID of the upstream Project if this is a fork, null otherwise")
  - Update `ProjectSummary` schema: add `upstream_project_id` (nullable string)
  - Add `forks_count` to `ProjectDetail` (integer, "Number of forks created from this Project")

  Files: `src/services/api-gateway/docs/openapi.json`
  Verify: OpenAPI spec test from Task 10d now PASSES (GREEN)

<!-- Commit checkpoint: tasks 12-15 -->

### Phase 6: Demos Seed Data

- [x] **Task 16: Create 5 demo projects in VEDO Demos group**
  Seed data for the demos group with 5 canonical domain templates as forkable projects.

  - Create `deploy/seeds/vedo-demos/` directory with:
    - `bootstrap.sh` — idempotent initialization script (checks if `VEDO Demos` group exists before creating)
    - 5 JSON sequence files (reuse `SequenceStep` format from existing proto): `organization.json`, `product.json`, `process.json`, `glossary.json`, `event.json`
    - `README.md` — describes the 5 demos and bootstrap instructions
  - Bootstrap flow:
    1. Create Group `VEDO Demos` (via auth-service org gRPC `CreateGroup`) — idempotent check by slug
    2. For each of 5 sequences:
       a. Create Project in `VEDO Demos` (via `CreateProject`) — idempotent check by name
       b. Apply sequence via `ontology.ApplySequence` gRPC (reuse existing `ontologyv1.ApplySequenceRequest` proto — Published Language, NOT a new mechanism)
       c. Set `visibility=public` (via `SetVisibility`)
    3. Log completion
  - 5 sequence files contain the same classes/properties as described in the former `REQ-FUN.API.templates-catalog`:
    - Organization: Organization, Department, Employee, Position
    - Product: Product, Category, Manufacturer, Review
    - Process: Process, Step, Input, Output, Agent
    - Glossary: Term, Definition, RelatedTerm, Source
    - Event: Event, Participant, Location, DateTime

  LOGGING (verbose):
  - `slog.Info("seed.demos.start")`
  - `slog.Debug("seed.project", "name", name, "steps", len(steps))`
  - `slog.Info("seed.demos.complete", "projects_created", count)`

  Files: `deploy/seeds/vedo-demos/bootstrap.sh`, `deploy/seeds/vedo-demos/{organization,product,process,glossary,event}.json`, `deploy/seeds/vedo-demos/README.md`
  Verify: running bootstrap against dev environment creates 5 Projects in `VEDO Demos`, each with an Ontology and several classes
  > Note: `SequenceStep` proto already exists — reuse `ontologyv1.ApplySequenceRequest`, do NOT create a new proto

### Phase 7: Design Sync (Pencil.dev)

- [x] **Task 17: Sync Pencil.dev design files for fork flow**
  Update `design/` Pencil files to reflect the fork-based template replacement BEFORE any frontend code changes. The frontend implementation (Task 16) will use these designs as source of truth.

  **Step 1: New organism `ForkDemoDialog` in `design/ui-kit.lib.pen`**
  - Create organism `Organism/ForkDemoDialog` in the Navigation group (or a new "Project Lifecycle" group)
  - Set `"reusable": true`
  - Content: modal dialog with:
    - Title: "Fork demo project"
    - Grid/list of 5 demo project cards (Organization, Product, Process, Glossary, Event)
    - Each card shows: project name, description, class count, property count, domain tag
    - "Fork" button on each card
    - "Cancel" button (GhostButton) in footer
    - Search/filter field at top (SearchField molecule)
  - Create light theme variant `Organism/ForkDemoDialog (Light)` with `"theme": { "mode": "light" }`
  - Both dark and light variants must stay synchronized (per `design/README.md` §8)

  **Step 2: New molecule `DemoProjectCard` in `design/ui-kit.lib.pen`**
  - Create molecule `Molecule/DemoProjectCard`
  - Set `"reusable": true`
  - Content: card with project icon, name, description, metadata badges (class count, property count), "Fork" PrimaryButton
  - Create light theme variant `Molecule/DemoProjectCard (Light)`

  **Step 3: Update `design/pages/projects.pen`**
  - Add "Fork demo project" entry point: a PrimaryButton or card in the page header/toolbar area
  - Add `ForkDemoDialog` organism as a modal overlay (ref `B:<ForkDemoDialog-id>`)
  - Remove any "Create from template" UI elements if present
  - Wire the "Fork demo project" button to show the `ForkDemoDialog` modal

  **Step 4: Update `design/pages/dialogs.pen`**
  - Add `ForkDemoDialog` to the dialog collection (alongside CreateClass, CreateProperty, etc.)
  - Add `DemoProjectCard` examples inside the dialog preview

  **Step 5: Update `design/pages/ontology-workspace.pen`**
  - If a "Create from template" button exists in the toolbar → replace with "Fork demo project" button
  - If no template reference exists → add a "Fork" button in the ontology toolbar for forking the current public project

  **Step 6: Update `design/pages/dashboard.pen`**
  - Add a "Quick start: Fork a demo" widget/card in the dashboard greeting area
  - Widget shows 3-5 demo project cards with Fork buttons (using `DemoProjectCard` molecule)
  - Links to `ForkDemoDialog` for full catalog

  **Step 7: Update `design/README.md`**
  - Add `ForkDemoDialog` to the Organisms list (Navigation or Project Lifecycle group)
  - Add `DemoProjectCard` to the Molecules list
  - Add `fork-demo-dialog` to the dependency graph mermaid diagram (under pages that use it: projects.pen, dialogs.pen, dashboard.pen)
  - Document the fork flow in section 3 ("How to create a new page") as a reference for modal-based flows

  **Step 8: Verify reference integrity**
  - All `ref` values in modified `.pen` files must contain `:` (alias prefix `B:`) or be local references
  - No broken references to removed template-related components (if any existed)
  - Run `batch_get` on each modified `.pen` file to verify structure

  LOGGING: N/A — Pencil.dev files are design artifacts, not runtime code

  Files: `design/ui-kit.lib.pen`, `design/pages/projects.pen`, `design/pages/dialogs.pen`, `design/pages/ontology-workspace.pen`, `design/pages/dashboard.pen`, `design/README.md`
  Verify: `batch_get` on each modified `.pen` file returns valid node tree; `get_screenshot` on ForkDemoDialog organism shows correct layout; no broken `ref` references
  > Use `pencil-design` skill and `batch_design` / `batch_get` / `get_screenshot` MCP tools for .pen file manipulation
  > Page size: FullHD (1920x1080 px) per `design/README.md` §3
  > All organisms need dark + light variants per `design/README.md` §8

<!-- Commit checkpoint: task 17 -->

### Phase 8: GUI (TDD — GUI tests before frontend code)

- [x] **Task 18: GUI tests for fork UI (RED phase)**
  Write frontend tests that define the fork UI contract, BEFORE implementing any Vue components. Tests must fail (RED) until implementation in Task 18.

  Create `src/services/frontend/src/__tests__/ForkDemoDialog.spec.ts`:
  - Mock REST endpoint `POST /api/v1/projects/:id/fork` (use `axios` mock)
  - Assert `ForkDemoDialog` renders list of 5 demo projects with their names and descriptions
  - Assert "Fork" button exists on each card
  - Assert clicking "Fork" calls `POST /api/v1/projects/:id/fork` with correct project ID
  - Assert on success: redirect to new Project workspace URL
  - Assert on 403 error: shows "Access denied" error message
  - Assert on 503 error: shows "Service unavailable, please retry" with Retry button
  - Assert empty state when no demos available: shows "No demo projects available" message
  - Assert search/filter reduces visible cards
  - Assert "Cancel" button closes dialog without action

  LOGGING: `console.info("[fork.test]", scenario)` for each test scenario in dev mode

  Files: `src/services/frontend/src/__tests__/ForkDemoDialog.spec.ts`
  Verify: `pnpm vitest` — tests fail (RED) because components don't exist yet
  > BDD naming (TS): `'should fork demo project when user clicks Fork'`
  > New tests must meet minimum quality (TQS >= bronze)

<!-- Commit checkpoint: task 18 -->

- [x] **Task 19: Frontend implementation (GREEN phase)**
  Make the GUI tests pass by implementing the fork UI components.

  - In `src/services/frontend/src/`:
    - Remove any template-related queries/mutations from `apollo/queries.ts` (verify by grep — earlier scan showed 0 business-template matches, only Vue SFC `<template>` tags)
    - Add fork API call via REST (NOT GraphQL mutation — fork is a write operation, must use REST per ADR `rest-graphql-mutation-boundary`):
      ```typescript
      // api/fork.ts
      export async function forkProject(sourceProjectId: string): Promise<ForkResponse> {
        const res = await axios.post(`/api/v1/projects/${sourceProjectId}/fork`, {}, {
          headers: { 'Idempotency-Key': crypto.randomUUID() }
        });
        return res.data;
      }
      ```
  - Create `src/services/frontend/src/components/projects/ForkDemoDialog.vue`:
    - Displays list of 5 demo projects from `VEDO Demos` group (fetch via `GET /api/v1/groups/<vedo-demos-id>/projects`)
    - Each demo shows preview (class count, description) and a "Fork" button
    - On fork: call `forkProject()`, redirect to new Project workspace
    - Error states: 403 (access denied), 503 (retry), empty (no demos)
    - Search/filter field to filter demo projects by name
  - Update `ProjectsPage.vue` or `CreateOntologyDialog.vue`:
    - Replace any "Create from template" UI with "Fork demo project" entry point
    - Add "Fork" button on any public/internal Project page (not just demos)

  LOGGING: client-side `console.info("[fork]", sourceProjectId)` in dev mode (controlled by `import.meta.env.DEV`)

  Files: `src/services/frontend/src/api/fork.ts` (new), `src/services/frontend/src/components/projects/ForkDemoDialog.vue` (new), `src/services/frontend/src/components/projects/ProjectsPage.vue`, `src/services/frontend/src/apollo/queries.ts`
  Verify: `pnpm vitest` — Task 17 GUI tests now PASS (GREEN)

<!-- Commit checkpoint: tasks 18-19 -->

### Phase 9: Docs & Traceability

- [x] **Task 20: Update traceability.ttl**
  Sync traceability graph with new and deleted test files.

  - Add `vdo:TestSuite` entries for:
    - `tests/security/fork_bola_test.go`
    - `src/services/frontend/src/__tests__/ForkDemoDialog.spec.ts`
    - `src/services/auth-service/org/fork_test.go`
  - Add `vdo:validates` triples for each `// Validates: REQ-...` annotation:
    - `fork_bola_test.go` -> `REQ-NFR.SECURITY.bola-bfla-negative-tests`
    - `fork_test.go` -> `REQ-NFR.SECURITY.organization-access-model`
    - `ForkDemoDialog.spec.ts` -> `REQ-USR.UI.gui-implementation`
  - Remove stale triples for deleted `template_handler_test.go`

  Files: `tests/traceability.ttl` (or equivalent location — verify by grep)
  Sync traceability graph with new and deleted test files.

  - Add `vdo:TestSuite` entries for:
    - `tests/security/fork_bola_test.go`
    - `src/services/frontend/src/__tests__/ForkDemoDialog.spec.ts`
    - `src/services/auth-service/org/fork_test.go`
  - Add `vdo:validates` triples for each `// Validates: REQ-...` annotation:
    - `fork_bola_test.go` -> `REQ-NFR.SECURITY.bola-bfla-negative-tests`
    - `fork_test.go` -> `REQ-NFR.SECURITY.organization-access-model`
    - `ForkDemoDialog.spec.ts` -> `REQ-USR.UI.gui-implementation`
  - Remove stale triples for deleted `template_handler_test.go`

  Files: `tests/traceability.ttl` (or equivalent location — verify by grep)

- [ ] **Task 21: Antora docs — fork flow and demos**
  Update user guide, developer guide, and admin guide for the fork-based flow.

  - `src/docs/antora/user-guide/pages/ontology-creation.adoc` (or equivalent):
    - Replace "Create from template" section with "Fork demo project"
    - Add "Fork any public project" section — fork works for any public/internal Project, not just demos
    - Describe the fork flow: browse demos -> preview -> fork -> new private project in user space
  - `src/docs/antora/developer-guide/pages/organization-model.adoc`:
    - Add fork to "Project lifecycle": create -> fork -> (post-MVP: merge upstream)
    - Document `upstream_project_id` field and `POST /api/v1/projects/{id}/fork` endpoint
  - `src/docs/antora/admin-guide/pages/deployment.adoc`:
    - Add "Seed data: VEDO Demos bootstrap" section in post-deploy steps
    - Document `deploy/seeds/vedo-demos/bootstrap.sh` usage

  Files: `src/docs/antora/user-guide/pages/*.adoc`, `src/docs/antora/developer-guide/pages/organization-model.adoc`, `src/docs/antora/admin-guide/pages/deployment.adoc`

<!-- Commit checkpoint: tasks 19-21 -->

## Dependency Graph

```
Task 1 (delete handler)  ─┐
Task 2 (delete models)   ─┼──> Commit 1
Task 3 (delete REQs)     ─┤
Task 4 (delete US/UC)    ─┘
                            ↓
Task 5 (vision)   ─┐
Task 6 (glossary) ─┼──> Commit 2
Task 7 (ROADMAP)  ─┤
Task 8 (ADR new)  ─┤
Task 9 (ADR ext)  ─┘
                     ↓
Task 10 (US/UC) ────> Commit 3  (requirements define expected behavior)
                     ↓
Task 11 (API tests) ──> Commit 4  (RED — tests verify US/UC scenarios)
                     ↓
Task 12 (mig+test)─┐
Task 13 (gRPC+test)─┼──> Commit 5  (GREEN — tests now PASS)
Task 14 (handler+test)─┤  (13 blocked by 12, 14 blocked by 13)
Task 15 (OpenAPI+tests)─┘
                     ↓
Task 16 (seed data) ─┐
                     ├──> Commit 6  (17 blocked by 16)
Task 17 (design)   ──┘
                     ↓
Task 18 (GUI tests)───> Commit 7  (RED — tests verify fork UI)
                     ↓
Task 19 (frontend) ───> Commit 7  (GREEN — tests PASS; blocked by 17+18)
                     ↓
Task 20 (traceability) ─┐
Task 21 (Antora docs)   ─┼──> Commit 8
```

## Out of Scope

- Does NOT modify `ontologyv1.ApplySequenceRequest` proto — Published Language, used by document-extractor and ai-orchestration
- Does NOT implement merge-upstream-changes (fork -> pull updates from upstream) — post-MVP, F13.2 territory
- Does NOT implement rest of Social Hub (stars, DOI, sponsorship) — post-MVP
- Does NOT add Search BC (cross-ontology search) — post-MVP per DDD analysis
- Does NOT add MCP BC — post-MVP per DDD analysis
- Does NOT fix Audit BC — separate plan (`audit-bc`) per DDD analysis
- Does NOT do full doc sync for `project-ontology-separation` (ADR `gitlab-like`, `context.md`) — pending plan; code is already done (migrations 007-009, routes.go, org_handler.go), only doc sync remains

## Validation Plan

- Build: `go build ./...`, `cargo build`, `pnpm build`
- Unit tests: `go test ./src/services/api-gateway/... ./src/services/auth-service/...`, `pnpm vitest`
- Integration: `tests/security/fork_bola_test.go` (real HTTP, no mocks)
- Seed verify: bootstrap against dev — 5 Projects in `VEDO Demos` with Ontologies
- TQS gate: TQS >= bronze for new tests (per skill-context rule)
- Traceability: `traceability.ttl` in sync with new/deleted test files
- OpenAPI: `/api/v1/projects/{id}/fork` present, `/templates/ontologies` absent
