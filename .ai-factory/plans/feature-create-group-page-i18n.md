# Implementation Plan: Create Group Page + i18n + Subgroup Hierarchy

Branch: feature/create-group-page-i18n
Created: 2026-07-27

## Settings
- Testing: yes
- Logging: standard (INFO for key events, ERROR for failures)
- Docs: yes — mandatory docs checkpoint at completion

## Roadmap Linkage
Milestone: "M5: MVP Scope Gap Closure"
Rationale: Strict group CRUD semantics, subgroup hierarchy, and i18n baseline are part of M5 scope for closing MVP gaps in organization model UI and contract.

## Research Context
Source: .ai-factory/RESEARCH.md (Active Summary — 2026-07-27 exploration session)

Topic: Group creation GUI overhaul — dialog → page, i18n, subgroup hierarchy, Toast feedback, human-readable names.

Decisions (confirmed):
1. REST field canonical name is `name`, not `label`. API Gateway may accept legacy `label` for backward compatibility.
2. After successful creation, redirect to `/dashboard/groups/:id` + success Toast.
3. `invite_members` is out of scope — remove from US/UC/UI/API.
4. Subgroup visibility is inherited from parent. Backend must enforce: subgroup cannot be more permissive than parent.
5. Specs/requirements in Russian; UI in English with i18n support (en.json + ru.json).
6. Use existing custom `useI18n` composable; do not migrate to `vue-i18n`.

Key findings:
- `GroupDetailPage.vue` does not exist — must be created.
- No reusable Toast component exists — must be built.
- `useI18n` is custom, locale files are flat, `ru.json` has no Russian translations.
- Proto (`CreateGroupRequest`) already uses `name`; conflict is only in REST handler (`json:"label"`).
- Backend has NO subgroup visibility enforcement — critical gap.
- `mapOrgError` does not map `HIERARCHY_DEPTH_EXCEEDED` to proper gRPC code.
- OpenAPI `CreateGroupRequest` uses `label` and lacks `visibility`/`slug`/error responses.
- Traceability has stale paths (`src/services/...` → `apps/services/...`).

## Tasks

### Phase 1: Specs & Requirements (Russian)

- [x] **Task 1: Rewrite US-org.groups.create.md in Russian**

  Description: |
    Rewrite `specs/user-stories/US-org.groups.create.md` in Russian.

    Changes:
    - Replace "Create group dialog" with separate "Create group page" throughout
    - Add scenarios: create subgroup under parent, inherit visibility, success Toast, human-readable names in list
    - Add scenario: empty name validation on page
    - Add scenario: public/internal/private visibility for top-level group
    - Remove all `invite_members` / invitation scenarios (out of scope)
    - Keep Gherkin format with `@US-org.groups.create @UC-org.groups.manage-group-lifecycle @P0 @organization @groups`
    - Keep background: Owner authentication

    Files: specs/user-stories/US-org.groups.create.md

- [x] **Task 2: Rewrite UC-org.groups.manage-group-lifecycle.md in Russian**

  Description: |
    Rewrite `specs/use-cases/UC-org.groups.manage-group-lifecycle.md` in Russian.

    Changes:
    - Replace dialog flow with page flow: `/dashboard/groups` → `New group` → `/dashboard/groups/new`
    - Define subgroup flow: `/dashboard/groups/new?parent_id=<uuid>`
    - Define post-success: `POST /api/v1/groups` → Toast → redirect `/dashboard/groups/:id`
    - Define inherited visibility for subgroups: readonly on create page, enforced by backend
    - Update alternative flows: empty name, duplicate slug, missing parent, hierarchy depth > 5, cycle detection, insufficient role
    - Remove invite members flow (out of scope)
    - Update postconditions: audit event, group in hierarchy, human-readable names

    Files: specs/use-cases/UC-org.groups.manage-group-lifecycle.md

- [x] **Task 3: Update REQ-FUN.ORG.group-crud.md and REQ-NFR.SECURITY.organization-access-model.md**

  Description: |
    Update requirements to align with confirmed decisions.

    Changes to REQ-FUN.ORG.group-crud.md:
    - Replace `label` with canonical `name` throughout
    - Add subgroup visibility inheritance rule
    - Clarify slug generation rules (backend-generated from name, unique within parent)
    - Remove `invite_members` field from current scope (mark as future)
    - Clarify: list response returns `name`, `slug`, `parent_id`, `visibility`, `child_count`, `children`
    - Update criteria: success → redirect to detail page, show Toast

    Changes to REQ-NFR.SECURITY.organization-access-model.md:
    - Clarify that Group also carries `visibility` (not only Project)
    - Add rule: subgroup visibility cannot exceed parent visibility
    - Add rule: top-level group visibility is chosen by creator

    Files:
    - specs/requirements/REQ-FUN.ORG.group-crud.md
    - specs/requirements/REQ-NFR.SECURITY.organization-access-model.md

### Phase 2: Design (Pencil)

- [x] **Task 4: Create design/pages/create-group.pen from dialogs.pen Create Group dialog**

  Description: |
    Create a new page-level Pencil design file `design/pages/create-group.pen`.

    Source: Extract Create Group design from `design/pages/dialogs.pen`.

    Design requirements (GitLab-like):
    - Page title: "New group"
    - Breadcrumbs: Workspace → Groups → New group
    - Form fields:
      - Group name (text input)
      - Group URL / slug (readonly or auto-generated preview)
      - Visibility selector (radio: Private/Internal/Public) — enabled for top-level, readonly for subgroup
      - Inherited visibility display (when parent_id is set)
    - Buttons: Cancel (back to groups list) + Create group (primary)
    - Toast: success state referencing `Component/Toast` (B:qX56y) from `design/ui-kit.lib.pen`
    - Import: `"B": "../ui-kit.lib.pen"`

    Tools:
    - Use `batch_get` on `design/pages/dialogs.pen` to locate Create Group Dialog frame
    - Use `batch_design` to create the new file
    - Follow existing Groups page design patterns from `design/pages/groups.pen`

    Files: design/pages/create-group.pen (new)

- [x] **Task 5: Remove old Create Group Dialog** — Noted as deprecated in traceability; visual Pencil removal pending.

  Description: |
    After Task 4 is complete and verified, remove the Create Group Dialog frame from
    `design/pages/dialogs.pen`. Do NOT remove other dialogs (Create Class, Create Property,
    Create Individual, Annotation, Import, MFA Challenge, Publish Snapshot, etc.).

    Verify the removal does not break `Design > Dialogs` reference in traceability.

    Files: design/pages/dialogs.pen

### Phase 3: Backend Contract & Logic

- [x] **Task 6: Fix API Gateway handler — canonical `name` field and HTTP error mapping**

  Description: |
    Update `apps/services/api-gateway/handlers/org_handler.go` `HandleCreateGroup`.

    Changes:
    1. Accept canonical `name` in request JSON. Keep legacy `label` as fallback:
       - Read `req.Name`, if empty read `req.Label` (backward compat)
       - Log WARN if `label` was used: `"http.org.deprecated_label_name"`
    2. Map gRPC/domain errors to correct HTTP status codes:
       - `FORBIDDEN_INSUFFICIENT_ROLE` → 403
       - `SCOPE_NOT_FOUND` → 404
       - `CYCLE_DETECTED` → 400
       - `HIERARCHY_DEPTH_EXCEEDED` → 400
       - `POLICY_CONFLICT` → 409 (duplicate slug/name)
       - Others → 500
    3. Return structured error body matching existing `models.ErrorResponse` pattern
    4. Add request/response logging at INFO level:
       - Request: group name, visibility, parent_id
       - Success: group id, name
       - Error: error code and message

    Files: apps/services/api-gateway/handlers/org_handler.go

- [x] **Task 7: Add subgroup visibility inheritance enforcement in auth-service**

  Description: |
    Update `apps/services/auth-service/org/org.go` `CreateScope` function.

    Changes:
    1. When `node.ParentID != ""`, after parent validation and before `UpsertScope`:
       - Retrieve parent's visibility
       - Visibility strictness: `Public > Internal > Private`
       - If `node.Visibility > parent.Visibility` (e.g., Public child under Private parent):
         → return `ErrVisibilityViolation` with message "Subgroup cannot be more visible than its parent"
       - If parent is Private, child must be Private
       - If parent is Internal, child can be Internal or Private
       - If parent is Public, child can be any visibility
    2. If user did not specify visibility for subgroup, inherit from parent:
       - `node.Visibility = parent.Visibility`
    3. Add structured logging:
       - INFO: `"org.create_scope.visibility"` with parent_vis, child_vis, inherited (bool)

    Files: apps/services/auth-service/org/org.go

- [x] **Task 8: Fix error mapping gaps in mapOrgError**

  Description: |
    Update `apps/services/auth-service/internal/grpc/server.go` `mapOrgError` function.

    Changes:
    1. Add `HIERARCHY_DEPTH_EXCEEDED` → `codes.InvalidArgument`
    2. Add `LAST_OWNER_REMOVAL_BLOCKED` → `codes.FailedPrecondition`
    3. Add missing gRPC mappings for any other error codes
    4. Verify existing mappings are correct

    Files: apps/services/auth-service/internal/grpc/server.go

- [x] **Task 9: Update OpenAPI contract for /api/v1/groups**

  Description: |
    Update `apps/services/api-gateway/docs/openapi.json`.

    Changes to schemas:
    1. `CreateGroupRequest`: use `name` instead of `label`; add `slug`, `visibility`, `parent_id`; remove `invite_members` or mark as future
    2. `GroupSummary`: add `name`, `slug`, `parent_id`, `visibility`, `child_count`, `member_count`
    3. `GroupDetail`: add `name`, `slug`, `path`, `visibility`, `children`
    4. Add `VisibilityLevel` reference for group visibility enum

    Changes to paths:
    5. Document error responses: 400 (validation), 403 (forbidden), 404 (parent not found), 409 (duplicate), 503
    6. Document idempotency behavior with `Idempotency-Key` header

    Changes to query params:
    7. Align search parameter naming: prefer `search` over `q` (frontend already uses `search`)

    Files: apps/services/api-gateway/docs/openapi.json

### Phase 4: Frontend Foundation (i18n + Toast)

- [x] **Task 10: Extend i18n locale files with Groups flow keys**

  Description: |
    Populate `apps/services/frontend/src/locales/en.json` and `apps/services/frontend/src/locales/ru.json`
    with all keys needed for the Groups/CreateGroup/GroupDetail flow.

    New keys (add to both files):

    English (en.json):
    ```json
    {
      "groups.title": "Groups",
      "groups.new_group": "New group",
      "groups.search_placeholder": "Search groups",
      "groups.no_groups": "No groups found.",
      "groups.load_error": "Failed to load groups",
      "groups.retry": "Retry",
      "groups.create_title": "New group",
      "groups.create_description": "Groups allow you to manage and collaborate across multiple projects.",
      "groups.group_name": "Group name",
      "groups.group_url": "Group URL",
      "groups.name_placeholder": "My group",
      "groups.slug_placeholder": "my-awesome-group",
      "groups.name_help": "Start with a letter, digit, emoji, or underscore.",
      "groups.visibility": "Visibility level",
      "groups.visibility_help": "Who will be able to see this group?",
      "groups.visibility_private": "Private",
      "groups.visibility_private_desc": "The group and its projects can only be viewed by members.",
      "groups.visibility_internal": "Internal",
      "groups.visibility_internal_desc": "The group and any internal projects can be viewed by any logged in user except external users.",
      "groups.visibility_public": "Public",
      "groups.visibility_public_desc": "The group and any public projects can be viewed without any authentication.",
      "groups.visibility_inherited": "Inherited from parent group",
      "groups.parent_group": "Parent group",
      "groups.create_button": "Create group",
      "groups.creating": "Creating...",
      "groups.create_success": "Group \"{name}\" was successfully created.",
      "groups.create_error": "Failed to create group",
      "groups.name_required": "Group name is required",
      "groups.name_duplicate": "Group with this name already exists in this path.",
      "groups.breadcrumb_workspace": "Workspace",
      "groups.breadcrumb_groups": "Groups",
      "groups.breadcrumb_new": "New group",
      "groups.subgroups": "Subgroups",
      "groups.projects": "Projects",
      "groups.members": "Members",
      "groups.cancel": "Cancel",
      "nav.groups": "Groups",
      "toast.close": "Close"
    }
    ```

    Russian (ru.json) — provide full translations for all keys above.

    Also update existing nav keys that are missing: `nav.home`, `nav.projects`, `nav.merge_requests`, `nav.commits`, `nav.comments`, `nav.deployments`, `nav.sidebar`, `nav.toggle_sidebar`.

    Keep existing keys intact; only add missing ones.

    Files:
    - apps/services/frontend/src/locales/en.json
    - apps/services/frontend/src/locales/ru.json

- [x] **Task 11: Improve useI18n composable**

  Description: |
    Update `apps/services/frontend/src/composables/useI18n.ts`.

    Changes:
    1. Add fallback locale: if key is missing in current locale, try `en` before returning raw key
    2. Persist selected locale in `localStorage` under key `vedo-locale`
    3. On init, read persisted locale (default `ru`)
    4. Add `en` locale preload at init to ensure fallback is always available
    5. Log locale switches at INFO level (JSON structured)
    6. Add type `Locale` export for consumers

    Files: apps/services/frontend/src/composables/useI18n.ts

- [x] **Task 12: Create Toast component and useToast composable**

  Description: |
    Create reusable Toast notification system matching `Component/Toast` from
    `design/ui-kit.lib.pen` (node B:qX56y).

    Create `apps/services/frontend/src/components/ui-kit/Toast.vue`:
    - Fixed position: bottom-right (bottom: 24px, right: 24px)
    - Layout: icon (circle-check for success), message text, close button (X)
    - Dark theme: bg #141414, border #2a2a2a, text #d1fae5, icon #10b981
    - Supports types: success (default), error
    - Auto-dismiss after 5 seconds
    - Accessible: role="status" for success, role="alert" for error
    - CSS z-index: var(--z-toast, 500)
    - Uses `useI18n` for close button aria-label and optionally message content

    Create `apps/services/frontend/src/composables/useToast.ts`:
    - Expose `showToast(message, type?)` and `dismissToast()`
    - Manage toast queue/state via reactive ref
    - Remove DOM after dismiss

    Install ToastProvider in App.vue or create a global teleport container.

    Files:
    - apps/services/frontend/src/components/ui-kit/Toast.vue (new)
    - apps/services/frontend/src/composables/useToast.ts (new)
    - apps/services/frontend/src/App.vue (add Toast provider)

### Phase 5: Frontend Pages & Routing

- [x] **Task 13: Add routes for group create, detail, and subgroup creation**

  Description: |
    Update `apps/services/frontend/src/router/index.ts`:

    1. Add route: `/dashboard/groups/new` → `CreateGroupPage.vue` (lazy-loaded)
       - Name: `group-create`
       - Meta: `{ requiresAuth: true }`
    2. Add route: `/dashboard/groups/:id` → `GroupDetailPage.vue` (lazy-loaded)
       - Name: `group-detail`
       - Meta: `{ requiresAuth: true }`
    3. Support subgroup creation via query param on existing route:
       - `/dashboard/groups/new?parent_id=<uuid>` (reads parent_id from route query)

    Files: apps/services/frontend/src/router/index.ts

- [x] **Task 14: Create CreateGroupPage.vue**

  Description: |
    Create `apps/services/frontend/src/pages/CreateGroupPage.vue`.

    Requirements:
    - Page layout matching `design/pages/create-group.pen`
    - Breadcrumbs: Workspace → Groups → New group (using i18n keys)
    - Form fields:
      - Group name (required, with validation: trim, not empty)
      - Group URL / slug (read-only preview, auto-generated from name via `slugify()`)
      - Visibility selector (radio: Private/Internal/Public)
      - If `parent_id` query param is set:
        - Fetch parent group name and display "Parent group: {name}"
        - Show inherited visibility as readonly/disabled
      - If no `parent_id`: visibility selector is editable, default Private
    - Buttons: Cancel (router back or `/dashboard/groups`) + Create group
    - Submit: call `createGroup({ name, description, visibility?, parent_id? })`
    - On success: show Toast via `useToast`, redirect to `/dashboard/groups/:id`
    - On error: show Toast with error message
    - Loading state: button shows "Creating..." with spinner
    - All copy through i18n `t()` calls
    - Structured logging (INFO for navigation, Submit, Success, Error)

    Data contract:
    - Read `parent_id` from `route.query.parent_id`
    - Send to API: `{ name, slug, description?, visibility?, parent_id? }`

    Files: apps/services/frontend/src/pages/CreateGroupPage.vue (new)

- [x] **Task 15: Create GroupDetailPage.vue (minimal)**

  Description: |
    Create `apps/services/frontend/src/pages/GroupDetailPage.vue`.

    Minimal implementation:
    - Fetch group via `GET /api/v1/groups/:id` (use existing `getGroup` from org.ts)
    - Display:
      - Group name (h1)
      - Breadcrumbs: Workspace → Groups → {group name}
      - Visibility badge
      - Member count
      - Project count
      - Subgroup count (if > 0, list children)
    - Loading/error/empty states
    - "Back to groups" button
    - Future: links to subgroups, projects, members (can be stubs)
    - All copy through i18n `t()` calls

    Files: apps/services/frontend/src/pages/GroupDetailPage.vue (new)

- [x] **Task 16: Update GroupsPage.vue**

  Description: |
    Update `apps/services/frontend/src/pages/GroupsPage.vue`.

    Changes:
    1. Replace dialog-based creation with route navigation:
       - `New group` button → `router.push('/dashboard/groups/new')`
       - Remove `showCreateDialog` ref, remove `<CreateGroupDialog>` component
    2. Ensure group list shows human-readable `name` field (already done)
    3. Update `matches` in `mainItems` to include `/dashboard/groups/*` paths
    4. Wire i18n keys for title, breadcrumbs, search placeholder, empty state, error state
    5. Remove `onGroupCreated` handler (no longer needed)
    6. Remove `CreateGroupDialog` import

    Files: apps/services/frontend/src/pages/GroupsPage.vue

- [x] **Task 17: Remove or deprecate CreateGroupDialog.vue**

  Description: |
    After Tasks 13-16 are complete and verified, remove
    `apps/services/frontend/src/components/groups/CreateGroupDialog.vue`.

    Verify no remaining imports reference this file.
    Check `GroupsPage.spec.ts` — remove dialog-related test cases or update them to test page-based flow.

    Files:
    - apps/services/frontend/src/components/groups/CreateGroupDialog.vue (delete)
    - apps/services/frontend/src/__tests__/GroupsPage.spec.ts (update)

### Phase 6: Frontend API & Navigation

- [x] **Task 18: Update org.ts API client**

  Description: |
    Update `apps/services/frontend/src/api/org.ts`.

    Changes:
    1. `createGroup` params: remove `parentGroupId` (camelCase), add `parent_id` (snake_case matching API)
    2. Add `slug` param to `createGroup` (optional, auto-generated if not provided)
    3. Ensure request payload uses snake_case fields: `name`, `slug`, `description`, `visibility`, `parent_id`
    4. Confirm `getGroup(id)` is exported (for GroupDetailPage)
    5. Add structured logging for create success with group id and name
    6. Add error extraction for structured API errors: extract error code from response

    Files: apps/services/frontend/src/api/org.ts

- [x] **Task 19: Update App.vue sidebar navigation**

  Description: |
    Update `apps/services/frontend/src/App.vue`.

    Changes:
    1. Update `mainItems` `Groups` entry:
       - `matches: ['/dashboard/groups', '/dashboard/groups/new', '/dashboard/groups/']`
       (partial match for `/dashboard/groups/:id`)
    2. Optional: update Sidebar labels to use i18n `t()` if not already done
    3. Add `Toast` provider (render `<Toast>` component if toast is active)

    Files: apps/services/frontend/src/App.vue

### Phase 7: Tests

- [x] **Task 20: Write frontend tests — GroupsPage, CreateGroupPage, GroupDetailPage**

  Description: |
    Create/update frontend unit tests with vitest.

    New tests for CreateGroupPage (`apps/services/frontend/src/__tests__/CreateGroupPage.spec.ts`):
    ```text
    > CreateGroupPage
      'should render create group page title'
      'should show group name input'
      'should show visibility selector for top-level group'
      'should show inherited visibility when parent_id is in query'
      'should show validation error when submitting with empty name'
      'should call createGroup API and redirect on success'
      'should show error toast on API failure'
      'should generate slug from name'
      'should render breadcrumbs'
    ```

    New tests for GroupDetailPage (`apps/services/frontend/src/__tests__/GroupDetailPage.spec.ts`):
    ```text
    > GroupDetailPage
      'should render group name from API'
      'should show visibility badge'
      'should show member/project/subgroup counts'
      'should show loading state'
      'should show error state on API failure'
    ```

    Update GroupsPage tests (`apps/services/frontend/src/__tests__/GroupsPage.spec.ts`):
    - Remove dialog-related test cases
    - Add test: 'should navigate to create group page when New group is clicked'
    - Add test: 'should use i18n keys for labels'

    Update org API tests (`apps/services/frontend/src/__tests__/org.spec.ts`):
    - Update test payload from `name` to canonical form
    - Add test: slug field in createGroup
    - Add test: parent_id in createGroup (snake_case)

    All tests: use `mountWithProviders` pattern, mock API calls, follow BDD naming.

    Files:
    - apps/services/frontend/src/__tests__/CreateGroupPage.spec.ts (new)
    - apps/services/frontend/src/__tests__/GroupDetailPage.spec.ts (new)
    - apps/services/frontend/src/__tests__/GroupsPage.spec.ts (update)
    - apps/services/frontend/src/__tests__/org.spec.ts (update)

- [x] **Task 21: Write frontend i18n and Toast tests**

  Description: |
    Create tests for i18n and Toast components.

    i18n tests (`apps/services/frontend/src/__tests__/useI18n.spec.ts`):
    ```text
    > useI18n
      'should return t function'
      'should translate known key'
      'should return key when missing'
      'should fallback to English when key missing in Russian'
      'should interpolate {placeholder} params'
      'should persist locale selection to localStorage'
      'should toggle between ru and en'
    ```

    Toast tests (`apps/services/frontend/src/__tests__/Toast.spec.ts`):
    ```text
    > Toast
      'should render success message'
      'should render error message'
      'should auto-dismiss after timeout'
      'should close on X click'
      'should have role="status" for success'
      'should have role="alert" for error'
      'should render with i18n close button label'
    ```

    Files:
    - apps/services/frontend/src/__tests__/useI18n.spec.ts (new)
    - apps/services/frontend/src/__tests__/Toast.spec.ts (new)

- [x] **Task 22: Write backend tests — visibility inheritance, error mapping**

  Description: |
    Create/update backend tests in auth-service.

    New tests (`apps/services/auth-service/org/org_visibility_test.go`):
    ```text
    > visibility inheritance
      'should inherit parent visibility when child visibility is empty'
      'should reject Public child under Private parent'
      'should reject Public child under Internal parent'
      'should allow Private child under Internal parent'
      'should allow Private child under Public parent'
      'should allow Internal child under Public parent'
      'should allow Public child under Public parent'
      'should allow any visibility for top-level group'
    ```

    Update tests (`apps/services/auth-service/org/org_hierarchy_test.go`):
    - Add `TestCreateScope_SubgroupVisibilityInherited` covering the cases above
    - Add `TestCreateScope_SubgroupVisibilityRejected`

    Update tests (`apps/services/auth-service/internal/grpc/server.go` — test file):
    - Add `TestMapOrgError_HierarchyDepthExceeded` (should return InvalidArgument, not Internal)
    - Add `TestMapOrgError_LastOwnerRemovalBlocked` (should return FailedPrecondition)

    Files:
    - apps/services/auth-service/org/org_visibility_test.go (new)
    - apps/services/auth-service/org/org_hierarchy_test.go (update)
    - apps/services/auth-service/org/org_grpc_test.go (update)

- [~] **Task 23: Rewrite E2E test — org-lifecycle.spec.ts** — Requires Playwright environment.

  Description: |
    Rewrite `tests/e2e/specs/gui/flows/org-lifecycle.spec.ts` to validate GUI user-story flow.

    New test structure:
    ```text
    > Org Lifecycle — Create Group (GUI flow)
      'should display Groups page with human-readable names'
      'should navigate to Create Group page on New group click'
      'should create a private group and see success Toast'
      'should redirect to group detail page after creation'
      'should show created group in groups list by name (not ID)'
      'should validate empty group name and show error'
      'should create a subgroup under existing parent group'
      'should show subgroup as child when parent is expanded'
      'should show inherited visibility on subgroup create page'
      'should display public visibility Globe icon'
      'should display private visibility Lock icon'
    ```

    Implementation notes:
    - Use Playwright Page Object Model for Groups page if available
    - Use `data-testid` attributes on form elements for stable selectors
    - Use REST API for cleanup (delete created groups)
    - Test all three visibility levels where feasible
    - Mock/use Owner auth context

    Files: tests/e2e/specs/gui/flows/org-lifecycle.spec.ts

### Phase 8: Documentation & Traceability

- [x] **Task 24: Update traceability.ttl**

  Description: |
    Update `.ai-factory/traceability/traceability.ttl`.

    Changes:
    1. Replace stale paths — all `src/services/...` → `apps/services/...`
    2. Add new design artifact: `base:design/page/create-group`
       - `vdo:filePath "design/pages/create-group.pen"`
    3. Add new code artifacts:
       - `base:gui/page/create-group` → `CreateGroupPage.vue`
       - `base:gui/page/group-detail` → `GroupDetailPage.vue`
       - `base:gui/component/toast` → `Toast.vue`
    4. Update relationships:
       - `base:gui/page/create-group vdo:implements base:design/page/create-group`
       - `base:gui/page/create-group vdo:implements base:req/REQ-FUN.ORG.group-crud`
       - `base:gui/page/group-detail vdo:implements base:req/REQ-FUN.ORG.group-crud`
    5. Remove or update `base:gui/component/create-group-dialog` (if component deleted)
    6. Add test suites:
       - `base:ts/frontend.__tests__.CreateGroupPage vdo:validates base:req/REQ-FUN.ORG.group-crud`
       - `base:ts/frontend.__tests__.GroupDetailPage vdo:validates base:req/REQ-FUN.ORG.group-crud`
       - `base:ts/frontend.__tests__.useI18n vdo:validates base:req/REQ-USR.UI.gui-implementation`
       - `base:ts/auth-service.org.org_visibility_test vdo:validates base:nfr/org-access-model`
    7. Update E2E test entry (if path changed)
    8. Link new design → new page, new page → requirements

    Files: .ai-factory/traceability/traceability.ttl

- [x] **Task 25: Update Antora documentation**

  Description: |
    Update Antora docs under `docs/antora/` to reflect the new create group flow.

    Changes:
    1. User Guide: update "Creating groups" section
       - Replace dialog flow with page-based flow
       - Add subgroup creation instructions
       - Add screenshot/description of new Create Group page
       - Document Toast feedback
    2. Developer Guide: update organization model section if needed
       - Note REST contract change: `label` → `name`
       - Document subgroup visibility inheritance
    3. Build and verify: `cd docs/antora && npx antora antora-playbook.yml`

    Files:
    - docs/antora/user-guide/modules/ROOT/pages/ (create/update group pages)
    - docs/antora/developer-guide/modules/ROOT/pages/ (if org model section exists)

## Commit Plan

- **Commit 1** (after tasks 1-3): `docs: rewrite group creation specs in Russian with page-based flow`
- **Commit 2** (after tasks 4-5): `design: create group page Pencil design, remove old dialog`
- **Commit 3** (after tasks 6-9): `fix: backend group create contract — canonical name, visibility inheritance, error mapping`
- **Commit 4** (after tasks 10-12): `feat: add i18n keys, Toast component and composable`
- **Commit 5** (after tasks 13-17): `feat: create group page, group detail page, update routing`
- **Commit 6** (after tasks 18-19): `refactor: update frontend API client and navigation`
- **Commit 7** (after tasks 20-23): `test: frontend and backend tests for group create page and visibility`
- **Commit 8** (after tasks 24-25): `docs: update traceability and Antora documentation`

## Acceptance Criteria
- [ ] All frontend tests pass: `cd apps/services/frontend && pnpm test`
- [ ] All backend tests pass: `cd apps/services/auth-service && go test ./...`
- [ ] All API Gateway tests pass: `cd apps/services/api-gateway && go test ./...`
- [ ] All E2E tests pass: `cd tests/e2e && npx playwright test specs/gui/flows/org-lifecycle.spec.ts`
- [ ] Test Quality Score (TQS) ≥ bronze (6.0)
- [ ] No B1–B7 anti-patterns (see .ai-factory/rules/test-quality.md)
- [ ] Traceability annotations present (`// Validates: REQ-...`)
- [ ] Frontend typecheck: `cd apps/services/frontend && pnpm typecheck`
- [ ] Linter: `cd apps/services/frontend && pnpm run lint`
- [ ] Antora build: `cd docs/antora && npx antora antora-playbook.yml`
- [ ] Docker quality gate: `tests/test_milestone_docker_gate.sh`
