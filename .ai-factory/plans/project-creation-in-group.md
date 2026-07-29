# Plan: Project Creation in Group — Page-based UI with visibility, i18n, and paired Ontology

**Branch:** `feature/project-creation-in-group`
**Created:** 2026-07-27
**Roadmap:** M5 — MVP Scope Gap Closure

## Settings

| Setting | Value |
|---------|-------|
| Testing | Yes — unit, backend, E2E |
| Logging | Standard — INFO for request/success, ERROR for failures |
| Docs | Mandatory checkpoint at the end |
| Roadmap | M5: MVP Scope Gap Closure |

## Architecture Context

- **Pattern**: Microservices (monorepo) — API Gateway (Go) proxies REST → gRPC → auth-service (Go)
- **1:1 Project ↔ Ontology**: Project is the container, Ontology is the graph content. Atomic create/delete.
- **Canonical field**: `name` for Project (not `label`). `label` accepted as deprecated fallback in API Gateway.
- **Visibility hierarchy**: Private < Internal < Public. A child scope cannot be more visible than its parent.
- **Idempotency**: All write endpoints require `Idempotency-Key` header per ADR.

## Dependencies Graph

```
Phase 1: Specs (US, UC, REQ) ──┐
                                ├──► Phase 3.1: Proto (visibility field) ──► Phase 3.2: API Gateway ──► Phase 3.3: auth-service
Phase 2: Design (.pen) ────────┘                               │                                              │
                                                                └──► Phase 3.4: OpenAPI spec ◄─────────────────┘
                                                                           │
                                    Phase 4.1: i18n keys ◄────────────────┤
                                          │
                                    Phase 4.2: Router routes
                                          │
                                    Phase 4.3: API client (org.ts)
                                          │
                                    Phase 4.4: CreateProjectPage.vue
                                          │
                                    Phase 4.5: Update ProjectsPage.vue
                                          │
                                    Phase 4.6: Tests (unit + E2E)
                                          │
                                    Phase 5: Docs checkpoint ──► Phase 6: Traceability
```

## Tasks

---

- [x] ### Phase 1: Specifications & Requirements

- [x] #### Task 1.1: Create user story US-org.projects.create
- **Files to create**: `specs/user-stories/US-org.projects.create.md`
- **Language**: Russian
- **Scenarios**:
  - Owner creates a private Project in a selected Group
  - Owner selects parent Group from a human-readable list (names, not IDs)
  - Empty project name gives inline validation error
  - Visibility is saved and displayed on Project detail/list
  - Project appears in project list with human-readable `name`
  - On success: Toast shown + redirect to `/project/:id/workspace`
  - Backend returns `ontology_id` (UUID v4) in response
  - Project cannot be created in a non-existent Group
  - Project cannot be more public than its parent Group
- **Logging**: N/A — specification artifact
- **BDD naming reference**: Not applicable (spec doc)

- [x] #### Task 1.2: Create/update use case UC-org.projects.manage-project-lifecycle
- **Files to create**: `specs/use-cases/UC-org.projects.manage-project-lifecycle.md`
- **Language**: Russian
- **Content**:
  - Lifecycle of Project: create, view, update, move, delete, fork
  - Main success flow: create Project via dedicated page
  - Alternative flows:
    - Empty name
    - Duplicate slug/name within group
    - Missing/invalid `group_id`
    - Insufficient role (non-Owner/Maintainer in parent Group)
    - Visibility exceeds parent Group
    - Ontology pairing failure with compensating cleanup
    - Missing/reused `Idempotency-Key`
- **Logging**: N/A — specification artifact

- [x] #### Task 1.3: Update requirements for Project creation, visibility, naming
- **Files to read**:
  - `specs/requirements/REQ-FUN.DATA.ontology-identifier-standard.md`
  - `specs/requirements/REQ-FUN.DATA.ontology-visibility-levels.md`
- **Files to create/update**: Update existing files; create `specs/requirements/REQ-FUN.ORG.project-creation.md` if project creation requirement does not exist
- **Language**: Russian (for new/changed content)
- **Changes**:
  - `REQ-FUN.DATA.ontology-identifier-standard.md`: Confirm UUID v4 generation for `ontology_id` server-side; no changes needed, already correct.
  - `REQ-FUN.DATA.ontology-visibility-levels.md`: Add note that visibility is set at Project creation, inherited/validated against parent Group, and paired Ontology inherits via 1:1 link.
  - `REQ-FUN.ORG.project-creation.md` (new): Canonical field is `name`, `group_id` mandatory, Toast feedback, GUI page instead of dialog, visibility at create.
- **Logging**: N/A — specification artifact

- [x] #### Task 1.4: Check/update ADR for POST /api/v1/projects
- **Files to read**: `specs/adr/ADR-DES.API.organization-rest-endpoints.md`
- **Changes**: No ADR changes needed — the ADR already specifies GitLab-aligned endpoints, `name` as canonical field, 1:1 Project↔Ontology, visibility, idempotency. Add a note about `visibility` field in `CreateProjectRequest` if not already mentioned.
- **Logging**: N/A — specification artifact

---

- [x] ### Phase 2: Design (.pen)

- [x] #### Task 2.1: Create create-project.pen design page
- **Files to create**: `design/pages/create-project.pen`
- **Design pattern**: Follow GitLab-like page layout (mirror the existing CreateGroupPage design pattern from `design/pages/groups.pen`)
- **Layout elements**:
  - Breadcrumbs: Workspace / Projects / New project
  - Title: "New project" with description
  - **Group namespace selector**: Dropdown/select showing human-readable group names (from `listGroups()` data)
  - **Project name**: Text input with auto-slug preview
  - **Project URL/slug preview**: Read-only display: `vedo-core.local/[group-slug]/[project-slug]`
  - **Description**: Textarea (optional)
  - **Visibility options**: Radio cards with icons (Lock/Shield/Globe) — Private / Internal / Public with explanations
  - **Actions**: Cancel (link to `/dashboard/projects`) + Create project (Primary button)
- **Toast**: Reference `ui-kit.lib.pen` Toast component for success state
- **Logging**: N/A — design artifact

- [x] #### Task 2.2: Remove old Create Project Dialog from dialogs.pen
- **Files**: `design/pages/dialogs.pen`
- **Action**: Search for Create Project Dialog frames (the exploration found no explicit CP1/CP1L frames — the dialog exists only in code). Confirm removal is not needed from design files. If any project-dialog frame exists, remove it. Be careful not to remove unrelated group/sidebar dialogs.
- **Logging**: N/A — design artifact

- [x] #### Task 2.3: Update projects.pen — New project button → page link
- **Files**: `design/pages/projects.pen`
- **Action**: If "New project" button currently references a dialog, update to indicate it navigates to the create page (`/dashboard/projects/new`)
- **Logging**: N/A — design artifact

---

- [x] ### Phase 3: Backend/API

- [x] #### Task 3.1: Add `visibility` field to CreateProjectRequest proto
- **Files**:
  - `apps/shared/proto/auth/v1/org.proto` (edit)
- **Changes**:
  ```protobuf
  message CreateProjectRequest {
    string name = 1;
    string description = 2;
    string group_id = 3;
    string organization_id = 4;
    string visibility = 5;  // "Private", "Internal", "Public" — default "Private" server-side
  }
  ```
- **Dependency**: Run `buf generate` if buf is configured, otherwise note the generated files that need regeneration (`apps/shared/src/proto/` for Rust, frontend proto stubs if any). Check `apps/shared/proto/` for buf config.
- **Logging**: N/A — proto compile
- **Tests**: Compilation is the test

- [x] #### Task 3.2: Update API Gateway HandleCreateProject — accept `name`, `visibility`, proper error mapping, structured logging
- **Files**:
  - `apps/services/api-gateway/handlers/org_handler.go` (edit)
- **Changes**:
  1. **Request struct**: Add `Name`, `Visibility`, keep `Label` as deprecated fallback
     ```go
     var req struct {
       Name        string `json:"name"`
       Label       string `json:"label"` // deprecated
       Description string `json:"description"`
       GroupID     string `json:"group_id"`
       Visibility  string `json:"visibility"`
     }
     ```
  2. **Canonical name resolution**: Mirror `HandleCreateGroup` pattern — use `req.Name`, fallback to `req.Label` with deprecation warning
  3. **Structured logging (slog)**:
     - `slog.Info("http.org.create_project.request", ...)` with name, group_id, visibility
     - `slog.Error("http.org.create_project.grpc_error", ...)` with grpc_code, error on gRPC failure
     - `slog.Info("http.org.create_project.success", ...)` with project_id
  4. **Error mapping**: Use `mapGrpcCodeToHTTP()` (already exists) for proper HTTP status codes instead of always 500
  5. **gRPC call**: Pass `name` (not `label`), `description`, `group_id`, `visibility`, `tenant_id` from token
- **Logging**:
  - `INFO http.org.create_project.request` — on incoming request
  - `ERROR http.org.create_project.grpc_error` — on gRPC failure
  - `INFO http.org.create_project.success` — on success
- **Tests**: API Gateway handler unit tests for:
  - `name` as canonical field (with `label` as fallback)
  - Missing `group_id` → 400
  - Missing `name` → 400
  - Valid request → 201 with `data.project`

- [x] #### Task 3.3: Update auth-service CreateProject — visibility, role validation, Owner membership, audit event
- **Files**:
  - `apps/services/auth-service/internal/grpc/server.go` (edit — `CreateProject` handler)
  - `apps/services/auth-service/org/org.go` (edit — `CreateScope` or add `CreateProject` logic)
  - `apps/services/auth-service/org/store.go` (read — no changes likely needed)
  - `apps/services/auth-service/org/postgres_store.go` (read — no changes likely needed)
- **Changes in `server.go` `CreateProject`**:
  1. Extract `visibility` from `req.Visibility`, default to `"Private"` if empty
  2. Set `node.Visibility` on ScopeNode before calling `CreateScope`
  3. After `CreateScope` succeeds, add Owner membership:
     ```go
     s.svc.Store().UpsertMembership(org.OrgMembership{
       UserID: requesterID,
       Scope:  projectID,
       Role:   "Owner",
     })
     ```
  4. Add structured audit event: `"event": "project.created"` — already done via `CreateScope` audit log; add project-specific audit with `ontology_id`
  5. Structured logging with `slog`:
     ```go
     slog.Info("project.create", "project_id", projectID, "ontology_id", ontologyID, "visibility", req.Visibility)
     slog.Error("project.create.ontology_pairing_failed", ...)
     ```
- **Changes in `org.go` `CreateScope`**: Already has:
  - Parent existence validation ✅
  - Cycle detection ✅
  - Hierarchy depth check ✅
  - Visibility inheritance from parent ✅
  - Visibility level enforcement (cannot exceed parent) ✅
  - **Missing**: Role validation — caller must have at least Maintainer in parent Group. **ADD this** before creating scope:
    ```go
    if node.ParentID != "" && node.Type == ScopeProject {
      role, err := s.GetEffectiveRole(requesterID, node.ParentID)
      if err != nil || !IsMaintainerOrAbove(role) {
        return ErrForbiddenInsufficientRole
      }
    }
    ```
- **Logging**:
  - `INFO project.create` — on successful creation
  - `ERROR project.create.ontology_pairing_failed` — on Ontology pairing failure (with project_id for cleanup tracking)
  - Audit event: `"event":"audit.scope.created"` (already in CreateScope) + add `"event":"audit.project.created"` with ontology_id
- **Tests**:
  - Parent Group existence validation
  - Visibility enforcement vs parent Group
  - Insufficient role in parent Group → denied
  - Paired Ontology UUID v4 is generated
  - Compensating cleanup on Ontology create failure
  - Owner membership is created
  - Audit event is logged

- [x] #### Task 3.4: Update OpenAPI spec
- **Files**: `apps/services/api-gateway/docs/openapi.json`
- **Changes**:
  - `CreateProjectRequest`: Add canonical fields
    - `name` (string, required)
    - `description` (string, optional)
    - `group_id` (string, required — UUID format)
    - `visibility` (string, optional — enum: "Private", "Internal", "Public"; default: "Private")
    - `label` (string, optional — deprecated)
    - `slug` (string, optional — only if backend slug support is real; otherwise omit)
  - Response `201`:
    ```json
    {
      "data": {
        "id": "uuid",
        "name": "string",
        "type": "project",
        "parent_id": "uuid (group_id)",
        "visibility": "Private|Internal|Public",
        "ontology_id": "uuid-v4"
      }
    }
    ```
  - Error responses:
    - `400` — validation errors (missing name, missing group_id, invalid visibility)
    - `403` — FORBIDDEN (insufficient role in parent group)
    - `404` — NOT_FOUND (group does not exist)
    - `409` — CONFLICT (duplicate name within group, reused Idempotency-Key)
    - `422` — FAILED_PRECONDITION (visibility exceeds parent group)
  - `Idempotency-Key` header requirement for `POST /api/v1/projects`
- **Logging**: N/A — API spec

---

- [x] ### Phase 4: Frontend

- [x] #### Task 4.1: Add i18n keys for Projects flow
- **Files**:
  - `apps/services/frontend/src/locales/en.json`
  - `apps/services/frontend/src/locales/ru.json`
- **New keys** (`projects.*`):
  ```
  projects.title                  "Projects"
  projects.new_project            "New project"
  projects.create_title           "New project"
  projects.create_description     "Create a new project in a group"
  projects.search_placeholder     "Search projects..."
  projects.no_projects            "No projects found"
  projects.load_error             "Failed to load projects"
  projects.retry                  "Retry"
  projects.name                   "Project name"
  projects.name_placeholder       "My project"
  projects.name_required          "Project name is required"
  projects.description            "Description"
  projects.description_placeholder "Project description"
  projects.group                  "Group"
  projects.group_placeholder      "Select a group"
  projects.group_required         "Group is required"
  projects.group_not_found        "Selected group not found"
  projects.project_url            "Project URL"
  projects.visibility             "Visibility"
  projects.visibility_help        "Choose visibility level for this project"
  projects.visibility_private     "Private"
  projects.visibility_private_desc "Project access must be granted explicitly to each user."
  projects.visibility_internal    "Internal"
  projects.visibility_internal_desc "The project can be accessed by any logged-in user."
  projects.visibility_public      "Public"
  projects.visibility_public_desc  "The project can be accessed without authentication."
  projects.visibility_inherited   "Inherited from parent group"
  projects.cancel                 "Cancel"
  projects.create_button          "Create project"
  projects.creating               "Creating..."
  projects.create_success         "Project created successfully"
  projects.create_error           "Failed to create project: {error}"
  projects.breadcrumb_workspace   "Workspace"
  projects.breadcrumb_projects    "Projects"
  projects.breadcrumb_new         "New project"
  ```
- **Logging**: N/A — locale data

- [x] #### Task 4.2: Add new route for /dashboard/projects/new
- **Files**: `apps/services/frontend/src/router/index.ts`
- **Changes**:
  ```typescript
  {
    path: '/dashboard/projects/new',
    name: 'project-create',
    component: () => import('@/pages/CreateProjectPage.vue'),
    meta: { requiresAuth: true },
  }
  ```
  The route accepts an optional query param `?group_id=<uuid>` for pre-selecting a group.
- **Logging**: N/A — route config

- [x] #### Task 4.3: Update API client — createProject sends `name`, `group_id`, `visibility`
- **Files**: `apps/services/frontend/src/api/org.ts`
- **Changes to `createProject`**:
  1. Accept `visibility` in params:
     ```typescript
     export async function createProject(params: {
       name: string;
       description?: string;
       groupId?: string | null;
       visibility?: string;
     }): Promise<ProjectInfo>
     ```
  2. Send canonical JSON payload:
     ```typescript
     const payload = {
       name: params.name,
       description: params.description,
       group_id: params.groupId ?? undefined,
       visibility: params.visibility ?? 'Private',
     };
     ```
     (Remove legacy `label` field from payload)
  3. Add `Idempotency-Key` header for this write operation:
     ```typescript
     const idempotencyKey = crypto.randomUUID?.() ?? `${Date.now()}-${Math.random()}`;
     const { data } = await api.post('/projects', payload, {
       headers: { 'Idempotency-Key': idempotencyKey },
     });
     ```
- **Logging**: Keep existing `INFO org.projects.create.request/success/error` pattern
- **Tests**: Vitest for `createProject` sending correct payload

- [x] #### Task 4.4: Create CreateProjectPage.vue
- **Files to create**: `apps/services/frontend/src/pages/CreateProjectPage.vue`
- **Pattern**: Follow `CreateGroupPage.vue` pattern (page-based, breadcrumbs, card layout, i18n, Toast)
- **Components to use**:
  - `useI18n().t` — all strings through i18n
  - `useToast().showToast` — success/error Toast
  - `useRouter()`, `useRoute()` — navigation and query params
  - `@/api/org` — `listGroups()`, `createProject()`
- **Template structure**:
  ```
  ┌─ Header ─────────────────────────────────┐
  │  Breadcrumbs: Workspace > Projects > New  │
  │  Title: New project                       │
  │  Subtitle description                     │
  ├─ Card ────────────────────────────────────┤
  │  ┌─ Group selector ───────────────────┐   │
  │  │  <select> with human-readable names│   │
  │  │  Loading state while groups load   │   │
  │  │  Error state if groups fail        │   │
  │  └────────────────────────────────────┘   │
  │  ┌─ Project name ─────────────────────┐   │
  │  │  <input> with validation           │   │
  │  │  Auto-slug preview below           │   │
  │  └────────────────────────────────────┘   │
  │  ┌─ Project URL preview ──────────────┐   │
  │  │  vedo-hub.ru/[group-slug]/[slug]   │   │
  │  └────────────────────────────────────┘   │
  │  ┌─ Description ──────────────────────┐   │
  │  │  <textarea> (optional)             │   │
  │  └────────────────────────────────────┘   │
  │  ┌─ Visibility ───────────────────────┐   │
  │  │  Radio cards: Private / Internal / │   │
  │  │  Public with descriptions          │   │
  │  │  Info when inherited from parent   │   │
  │  └────────────────────────────────────┘   │
  │  ┌─ Actions ──────────────────────────┐   │
  │  │  Cancel (link) │ Create (button)   │   │
  │  └────────────────────────────────────┘   │
  └────────────────────────────────────────────┘
  ```
- **Validation**:
  - Group required
  - Project name required (inline error)
- **Submit flow**:
  1. Show submitting state (button loading spinner)
  2. Call `createProject({ name, description, groupId, visibility })`
  3. On success: `showToast(t('projects.create_success'), 'success')` + `router.push(`/project/${id}/workspace`)`
  4. On error: `showToast(t('projects.create_error', { error: msg }), 'error')` or inline error
- **Logging**:
  - `INFO ui.create_project.submit` — on form submit
  - `INFO ui.create_project.success` — on success (with project_id)
  - `ERROR ui.create_project.failed` — on failure
- **Tests**: Vitest for CreateProjectPage

- [x] #### Task 4.5: Update ProjectsPage.vue — New project → page route
- **Files**: `apps/services/frontend/src/pages/ProjectsPage.vue`
- **Changes**:
  1. Replace `showCreateProjectDialog` with `router.push({ name: 'project-create' })`
  2. Remove `CreateProjectDialog` import and `v-model` binding
  3. If `CreateProjectDialog.vue` is no longer used anywhere, delete it; otherwise keep for any secondary usage
  4. Add i18n for visible strings (title, search placeholder, sort labels, empty state, error state, retry button)
  5. Add `onProjectCreated` handler removed — replace with page-based flow (user navigates back after create)
- **Note**: The `onProjectCreated` callback in ProjectsPage currently reloads the project list — after the page-based flow, this is handled by navigation back. Keep the `fetchProjects` on `onMounted()`.
- **Logging**: Keep existing structured logging pattern, add i18n keys for log messages where appropriate
- **Tests**: Update ProjectsPage tests — "New project" button should navigate to `/dashboard/projects/new`, not open a dialog

- [x] #### Task 4.6: Write frontend unit tests
- **Files**:
  - `apps/services/frontend/src/__tests__/ProjectsPage.spec.ts` (update)
  - `apps/services/frontend/src/__tests__/CreateProjectPage.spec.ts` (create)
- **ProjectsPage tests** (update test 7 — assert button navigates, not opens dialog):
  - `'should navigate to create page when New project is clicked when group is selected'` — verify `router.push` called with `{ name: 'project-create', query: { group_id } }`
  - `'should render i18n-driven text'` — verify translated strings appear
- **CreateProjectPage tests**:
  - `'should render create project page with breadcrumbs and title'`
  - `'should load and display groups in selector'` — mock `listGroups`, verify human-readable names
  - `'should show validation error when project name is empty'`
  - `'should show validation error when no group is selected'`
  - `'should call createProject with name, group_id, visibility on submit'`
  - `'should show success toast and redirect on success'`
  - `'should show error toast on failure'`
  - `'should be in loading state while submitting'`
  - `'should pre-select group from query param group_id'`
- **Logging**: N/A — test files
- **BDD naming**: Use TypeScript convention `'should <expected> when <condition>'`
- **Quality gate**: TQS ≥ bronze (6.0), no B1–B7 anti-patterns

- [x] #### Task 4.7: Write E2E tests
- **Files**: `tests/e2e/specs/gui/flows/org-lifecycle.spec.ts`
- **Add scenario**: `US-org.projects.create` GUI flow:
  1. Open Projects page (`/dashboard/projects`)
  2. Click "New project"
  3. Verify navigation to `/dashboard/projects/new`
  4. Select a Group from the selector (by human-readable name)
  5. Enter project name
  6. Select visibility level
  7. Click "Create project"
  8. Wait for success Toast
  9. Verify redirect to `/project/:id/workspace`
  10. Navigate back to Projects, verify project appears with human-readable name
- **Keep existing** REST smoke test for project creation
- **Logging**: N/A — E2E test

---

- [x] ### Phase 5: Backend Tests

- [x] #### Task 5.1: Write API Gateway handler tests
- **Files**: Check existing test location `apps/services/api-gateway/handlers/` for existing test files (likely `org_handler_test.go`)
- **Tests**:
  - `HandleCreateProject with name` → 201 + correct gRPC call
  - `HandleCreateProject with label (deprecated)` → 201 with warn log + gRPC uses name
  - `HandleCreateProject with empty name` → 400
  - `HandleCreateProject with missing group_id` → 400
  - `HandleCreateProject with visibility` → gRPC receives visibility
  - `HandleCreateProject gRPC error → mapped HTTP error` — use `mapGrpcCodeToHTTP`
- **Logging**: N/A — test file

- [x] #### Task 5.2: Write auth-service CreateProject tests
- **Files**: Create/update tests in `apps/services/auth-service/org/` (check existing test pattern — likely `org_test.go`)
- **Tests**:
  - `CreateProject_WithValidParentGroup_CreatesScopeAndOntology`
  - `CreateProject_WithMissingParentGroup_ReturnsScopeNotFound`
  - `CreateProject_WithVisibilityExceedingParent_ReturnsVisibilityViolation`
  - `CreateProject_WithInsufficientRole_ReturnsForbidden`
  - `CreateProject_InheritsVisibilityFromParentWhenEmpty`
  - `CreateProject_GeneratesOntologyUUID`
  - `CreateProject_WithOntologyPairingFailure_CleansUpScope`
  - `CreateProject_AddsOwnerMembership`
- **Logging**: N/A — test file
- **BDD naming**: `[Condition]_[Action]_[ExpectedResult]`

---

- [x] ### Phase 6: Docs & Traceability

- [x] #### Task 6.1: Update traceability.ttl
- **Files**: `.ai-factory/traceability/traceability.ttl`
- **Add triples for**:
  - `US-org.projects.create` — new user story
  - `UC-org.projects.manage-project-lifecycle` — new use case
  - `REQ-FUN.ORG.project-creation.md` — new requirement
  - `design/pages/create-project.pen` — design
  - `CreateProjectPage.vue` — frontend page (`vdo:CodeArtifact`)
  - `ProjectsPage.vue` — updated frontend page (`vdo:CodeArtifact`)
  - `api/org.ts` — API client (`vdo:CodeArtifact`)
  - API Gateway handler (`vdo:CodeArtifact`)
  - auth-service CreateProject (`vdo:CodeArtifact`)
  - OpenAPI spec for projects (`vdo:CodeArtifact`)
  - E2E test suite (`vdo:TestSuite`)
  - Vitest test suite for CreateProjectPage (`vdo:TestSuite`)
  - Vitest test suite for ProjectsPage (updated) (`vdo:TestSuite`)
- **Fix stale paths**: Replace any `src/services/...` references with `apps/services/...` for affected artifacts
- **Logging**: N/A — traceability artifact

- [x] #### Task 6.2: Update Antora documentation
- **Files**:
  - `docs/antora/developer-guide/modules/ROOT/pages/organization-model.adoc` (update)
  - `docs/antora/developer-guide/modules/ROOT/pages/project-ontology-model.adoc` (update if needed)
- **Add to `organization-model.adoc`**:
  - Project creation flow: POST /api/v1/projects, required fields (`name`, `group_id`), optional (`description`, `visibility`)
  - Visibility inheritance rule: project visibility ≤ parent group visibility
  - Owner membership auto-assignment
  - Paired Ontology UUID v4 generation
- **Russian sections**: Document new/changed sections in Russian per project policy
- **Logging**: N/A — documentation

---

- [x] ### Phase 7: Final Verification & Cleanup

- [x] #### Task 7.1: Delete CreateProjectDialog.vue if unused
- **Files**: `apps/services/frontend/src/components/projects/CreateProjectDialog.vue`
- **Action**: After ensuring no other component imports it, delete the file
- **Logging**: N/A — cleanup

- [x] #### Task 7.2: Run full test suite
- **Commands**:
  ```bash
  cd apps/services/frontend && npx vitest run
  cd apps/services/api-gateway && go test ./...
  cd apps/services/auth-service && go test ./...
  cd tests/e2e && npx playwright test --grep "org-lifecycle"
  ```
- **Acceptance**:
  - All tests pass
  - TQS ≥ bronze (6.0) for new test files
  - No B1–B7 anti-patterns
  - Traceability annotations present

---

## Commit Plan

Given 15+ tasks, commits will be grouped logically every 3–5 tasks:

| # | Commit Message | Task Group |
|---|----------------|------------|
| 1 | `feat(specs): add user story, use case, and requirements for project creation in group` | Tasks 1.1–1.4 |
| 2 | `feat(design): add create-project page design, update projects.pen` | Tasks 2.1–2.3 |
| 3 | `feat(proto): add visibility field to CreateProjectRequest` | Task 3.1 |
| 4 | `feat(api): update API Gateway HandleCreateProject with name, visibility, and structured logging` | Task 3.2 |
| 5 | `feat(auth): add visibility, role validation, Owner membership, and audit to CreateProject` | Task 3.3 |
| 6 | `feat(api): update OpenAPI spec for project creation contract` | Task 3.4 |
| 7 | `feat(i18n): add projects flow locale keys for en and ru` | Task 4.1 |
| 8 | `feat(ui): add CreateProjectPage with group selector, validation, i18n, and toast` | Tasks 4.2–4.4 |
| 9 | `feat(ui): update ProjectsPage to route to create page and add i18n` | Task 4.5 |
| 10 | `test(ui): add frontend unit tests for ProjectsPage and CreateProjectPage` | Task 4.6 |
| 11 | `test(e2e): add project creation GUI flow to org-lifecycle spec` | Task 4.7 |
| 12 | `test(api): add API Gateway and auth-service tests for project creation` | Tasks 5.1–5.2 |
| 13 | `docs: update traceability.ttl, Antora docs, and cleanup legacy dialog` | Tasks 6.1–6.2, 7.1 |

## Acceptance Criteria

- [x] `New project` on ProjectsPage opens `/dashboard/projects/new` (page), not a dialog
- [x] Create Project page requires selected Group and project name
- [x] Group selector shows human-readable group names (loaded from `listGroups()`)
- [x] Frontend sends canonical JSON: `name`, `description`, `group_id`, `visibility`
- [x] API Gateway accepts `name` as canonical; `label` as deprecated fallback
- [x] API Gateway uses proper error mapping (not always 500)
- [x] Backend creates Project under Group with paired Ontology UUID v4 (`ontology_id`)
- [x] Visibility is validated: project cannot be more public than parent Group
- [x] Backend validates caller has at least Maintainer role in parent Group
- [x] User becomes Owner of created Project (Owner membership auto-assigned)
- [x] On success: Toast shown (`"Project created successfully"`) + redirect to `/project/:id/workspace`
- [x] Project appears in list with human-readable `name`
- [x] Specs updated: `US-org.projects.create.md`, `UC-org.projects.manage-project-lifecycle.md`, requirements
- [x] Design updated: `design/pages/create-project.pen` created
- [x] OpenAPI spec updated with new contract
- [x] i18n keys added for Projects flow (en + ru)
- [x] Antora docs updated (organization-model.adoc)
- [x] Traceability.ttl updated with new artifact relationships
- [x] All tests pass
- [x] Test Quality Score (TQS) ≥ bronze (6.0) for new tests
- [x] No B1–B7 anti-patterns
- [x] Traceability annotations present (`// Validates: REQ-...`)
- [x] Legacy `CreateProjectDialog.vue` deleted if unused
- [x] Structured logging in all modified backend handlers (INFO request/success, ERROR failure)
