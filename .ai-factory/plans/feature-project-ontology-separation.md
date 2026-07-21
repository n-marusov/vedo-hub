# Plan: Project ↔ Ontology Separation

- **Branch:** `feature/project-ontology-separation`
- **Created:** 2026-07-21
- **Mode:** full
- **Status:** complete (2026-07-21)

## Summary

Resolve the `Project = Ontology` identity error that contaminated ADRs, REQs,
`context.md`, the Antora docs, `openapi.json`, `routes.go`, `org_handler.go`,
and the `auth-service` PostgreSQL schema. Re-establish the GitLab-aligned model:

- **Project** — platform container for collaborative work (Git-like repo,
  members, roles, visibility, ABAC policies). `VEDO Project = GitLab Project`.
- **Ontology** — content of a Project (TBox/ABox, classes, properties,
  individuals, axioms).
- **Ratio 1:1** — one Project contains exactly one Ontology. Grouping of
  multiple ontologies is done through `Group` hierarchy, not by packing into
  one Project. A Project without an Ontology does not exist; an Ontology
  outside a Project does not exist.
- **Members / visibility / policies live on `Project`**, not on `Ontology`.
  Canonical REST paths become `/projects/{id}/members|visibility|policies`,
  matching GitLab `/projects/:id/...`.

The change is documentation-first (specs → ADR → OpenAPI → Antora), then
code (PostgreSQL migration → API Gateway → frontend), then tests.

## Settings

- **Testing:** yes — update existing security and e2e tests; add new negative
  tests for the renamed endpoints. New tests must meet TQS ≥ bronze (6.0).
- **Logging:** standard (INFO for endpoint renames, WARN for fallback paths,
  ERROR for auth/scope failures). No new features, so no DEBUG noise.
- **Docs:** yes — mandatory docs checkpoint. Antora update is part of the
  plan, and a final `$aif-docs` consistency pass is required.

## Roadmap Linkage

- **Milestone:** none
- **Rationale:** Remediation of the M3 (Multi-Team Organization Model)
  deliverable. M3 is already marked complete (2026-07-19), but shipped with
  the `Project = Ontology` identity error. This plan corrects M3 in place
  rather than spinning up a new milestone. `$aif-verify --strict` may emit
  WARN for missing milestone linkage; that is acceptable.

## Research Context

> Source: `.ai-factory/RESEARCH.md` → Active Summary (2026-07-21 18:45)

**Topic:** Conceptual contradiction `Project vs Ontology` in the VEDO Core
organization model.

**Goal:** Resolve the `Project = Ontology` identity закреплённое in
ADR/REQ/context.md, in favour of «Project — container, Ontology — content».
Synchronize glossary, ADR, REQ, Antora doc, OpenAPI, routes.go, org_handler.go.

**Decisions:**
1. `Project ≠ Ontology`. Project = workspace, Ontology = graph content.
2. `VEDO Project = GitLab Project` (unit of access and versioning).
3. **1:1 (Project ↔ Ontology) is mandatory.** Group multiple ontologies via
   `Group`, not via packing into one Project.
4. `members`/`visibility`/`policies` live on **Project** (GitLab-aligned
   `/projects/:id/...`). Existing `/ontologies/{id}/members|visibility|policies`
   are renamed to `/projects/{id}/...` without aliases.
5. `ProjectSummary.ontology_count` is removed (always 1 under 1:1).

**Constraints:**
- Do not break the migration path in `ADR-DES.API.rest-graphql-mutation-boundary`
  (2026-07-21) — update its references to the renamed endpoints in lockstep.
- Preserve GitLab-like inheritance: `Group → Project → members/visibility/policies`.
- Coordinate with the `scopes` PostgreSQL schema (currently
  `type IN ('group', 'ontology')`).
- Use `.ai-factory/references/gitlab-projects-groups-api.md` as the
  canonical GitLab reference for endpoint shapes, access levels, and
  group/project transfer semantics. The VEDO REST surface must mirror
  GitLab `/groups/:id/...` and `/projects/:id/...` unless an explicit
  deviation is documented in `ADR-DES.API.organization-rest-endpoints.md`.

**References:**
- `.ai-factory/references/gitlab-projects-groups-api.md` — GitLab
  Projects & Groups API reference. Authoritative for endpoint shapes
  (`/groups/:id/members`, `/projects/:id/members`, `/projects/:id/transfer`),
  access levels (Guest/Reporter/Developer/Maintainer/Owner), pagination,
  and best practices. VEDO deviations must be documented in the new ADR.
- `.ai-factory/RESEARCH.md` — full exploration context and decision
  rationale.

---

## Affected Files

### Spec / ADR / REQ (submodule `specs`)

- `specs/context.md` — § «Модель организации работы» (L32)
- `specs/adr/ADR-DES.SECURITY.gitlab-like-organization-model.md`
- `specs/adr/ADR-DES.API.protocol-stack-strategy.md` — table row «Управление
  орг. моделью»
- `specs/adr/ADR-DES.API.graphql-sparql-split-strategy.md` — endpoint table
- `specs/adr/ADR-DES.API.rest-graphql-mutation-boundary.md` — migration table
- `specs/adr/ADR-DES.API.organization-rest-endpoints.md` — **NEW**
- `specs/requirements/REQ-NFR.SECURITY.organization-access-model.md`
- `specs/requirements/REQ-FUN.DATA.ontology-visibility-levels.md`
- `specs/adr/README.md` — register the new ADR

### OpenAPI contract

- `src/services/api-gateway/docs/openapi.json` — paths
  `/ontologies/{id}/members|visibility|policies` → `/projects/{id}/...`;
  remove `ProjectSummary.ontology_count`; revisit `CreateOntologyRequest`.

### Antora documentation

- `src/docs/antora/developer-guide/modules/ROOT/pages/organization-model.adoc`
- `src/docs/antora/integrator-guide/modules/ROOT/pages/api-reference.adoc`
- `src/docs/antora/developer-guide/modules/ROOT/pages/project-ontology-model.adoc`
  — **NEW** (dedicated page explaining the 1:1 model)
- `src/docs/antora/developer-guide/modules/ROOT/nav.adoc` — register new page

### PostgreSQL schema

- `src/services/auth-service/migrations/007_rename_ontology_scope_to_project.sql`
  — **NEW** migration
- `src/services/auth-service/migrations/008_create_ontologies_table.sql` —
  **NEW** migration (introduces the `ontologies` table with FK to project
  scope; preserves the 1:1 invariant)
- `src/services/auth-service/migrations/009_seed_ontologies_from_legacy.sql`
  — **NEW** backfill migration (one `ontologies` row per existing
  `type='ontology'` scope before the type rename)

### Auth service (Go)

- `src/services/auth-service/org/*.go` — `ScopeType` constants, project
  creation flow now creates a paired `ontologies` row, project delete cascades
  to `ontologies`
- `src/services/auth-service/membership/*.go` — scope string handling
  (`"project/" + id` canonical; `"ontology/" + id` accepted as legacy alias
  only during the migration window, then removed in a follow-up)
- `src/services/auth-service/internal/proto/*.proto` — revisit
  `ListProjectsResponse` / `ProjectDetail` (drop `ontology_count`; add
  `ontology_id` for the 1:1 pairing)
- Regenerate protobuf bindings

### API Gateway (Go)

- `src/services/api-gateway/routes.go` — rename org-write endpoints
  `/ontologies/:id/members|visibility|policies` → `/projects/:id/...`;
  update `Idempotency.CriticalPaths` to `/api/v1/projects/`
- `src/services/api-gateway/handlers/org_handler.go` —
  `scope = "project/" + c.Param("id")`; remove the `ontology/` scope branch
  for members/visibility/policies handlers
- `src/services/api-gateway/handlers/org_handler_test.go` — update route
  expectations

### Frontend (Vue 3)

- `src/services/frontend/src/apollo/queries.ts` — `LIST_MEMBERS_QUERY`,
  `UPDATE_MEMBER_ROLE`, `REMOVE_MEMBER` — switch from `/ontologies/{id}/...`
  to `/projects/{id}/...`
- `src/services/frontend/src/pages/MembersPage.vue` — use `projectId` from
  route params instead of `ontologyId`
- `src/services/frontend/src/pages/ProjectsPage.vue` — ensure project →
  ontology navigation uses the 1:1 link
- `src/services/frontend/src/apollo/mock-link.ts` — update mock fixtures

### Tests

- `tests/security/` — BOLA/BFLA negative tests for renamed endpoints
- `tests/e2e/playwright/` — update any flow that hits
  `/api/v1/ontologies/{id}/members|visibility|policies`
- `src/services/api-gateway/handlers/org_handler_test.go` — route/scope
  expectations
- `src/services/auth-service/org/*_test.go` — `type='project'` scope and
  `ontologies` table FK behavior
- `.ai-factory/traceability/traceability.ttl` — register new artifacts and
  validation triples

---

## Tasks

### Phase 0 — Spec consistency (submodule `specs`)

> Goal: remove the `Project = Ontology` identity from the canonical spec
> documents that downstream ADRs and code reference.

#### Task 0.1 — Update `specs/context.md` organization-model paragraph

> ✅ **Completed:** 2026-07-21

- **Files:** `specs/context.md` (L32–38)
- **Deliverable:**
  - Replace `**Ontology** — аналог GitLab Project: базовая единица работы,
    доступа и независимого версионирования; одна онтология соответствует
    одному внутреннему Git-like репозиторию.` with:
    `**Project** — аналог GitLab Project: платформенная сущность-контейнер
    для совместной работы над одной онтологией. Базовая единица работы,
    доступа и независимого версионирования; один Project содержит ровно
    одну Ontology (1:1) и имеет собственный Git-like репозиторий.`
  - Add a separate bullet for `**Ontology**` describing it as the graph
    content (TBox/ABox) of a Project.
  - Keep the `Membership`, `Owner`, `Maintainer`, `Visibility` bullets
    aligned with the new Project-centric wording.
- **Logging:** none (documentation-only).
- **Dependencies:** none.
- **Validation:** `grep -n "Ontology.*аналог GitLab Project" specs/context.md`
  returns no matches.

#### Task 0.2 — Update `ADR-DES.SECURITY.gitlab-like-organization-model`

> ✅ **Completed:** 2026-07-21

- **Files:**
  `specs/adr/ADR-DES.SECURITY.gitlab-like-organization-model.md`
- **Deliverable:**
  - In «Решение»: replace `Ontology как единица доступа и версионирования`
    with `Project как единица доступа и версионирования; Ontology —
    содержимое Project (TBox/ABox, классы, свойства, индивиды, аксиомы)`.
  - Keep the 1:1 invariant explicit: `Project и Ontology связаны 1:1 —
    один Project содержит ровно одну Ontology. Группировка нескольких
    онтологий выполняется через Group, а не через упаковку в один Project.`
  - In «Рассмотренные альтернативы»: rewrite the rejected alternative
    `Project содержит несколько онтологий` row. The rejection rationale
    changes: it is no longer «противоречит GitLab-like модели: Ontology
    должна быть единицей репозитория и доступа». The new rationale is:
    «Противоречит GitLab-модели "один репозиторий = один Project" и
    нарушает 1:1 Project ↔ Ontology; группировка нескольких онтологий
    выполняется через Group».
  - Add a new «Связанные ADR» link to
    `ADR-DES.API.organization-rest-endpoints.md` (created in Task 1.1).
- **Logging:** none.
- **Dependencies:** Task 0.1 (terminology alignment).
- **Validation:** ADR body no longer contains `Ontology` and
  `GitLab Project` in the same equivalence sentence; the 1:1 invariant is
  stated.

#### Task 0.3 — Update `REQ-NFR.SECURITY.organization-access-model`

> ✅ **Completed:** 2026-07-21

- **Files:**
  `specs/requirements/REQ-NFR.SECURITY.organization-access-model.md`
- **Deliverable:**
  - In «Назначение»: replace `В VEDO Core аналогом GitLab Project является
    **Ontology**` with `В VEDO Core аналогом GitLab Project является
    **Project**. **Ontology** — содержимое Project (граф понятий: TBox/ABox,
    классы, свойства, индивиды, аксиомы). Соответствие 1:1: один Project
    содержит ровно одну Ontology.`
  - In «Структура → Ontology»: rewrite as «Ontology — содержимое Project.
    Правила: Ontology находится внутри Project (одна на Project). Project
    находится внутри Group или Subgroup. Идея "Project содержит несколько
    онтологий" не используется — для группировки применяйте иерархию
    Group. Одна Ontology соответствует одному внутреннему Git-like
    репозиторию, принадлежащему её Project.»
  - In «Управление членством» / «Visibility»: keep Owner-only membership
    management, but anchor it on `Project` instead of `Ontology`.
- **Logging:** none.
- **Dependencies:** Task 0.2.
- **Validation:** `grep -n "аналогом GitLab Project является .Ontology."`
  returns no matches.

#### Task 0.4 — Update `REQ-FUN.DATA.ontology-visibility-levels`

> ✅ **Completed:** 2026-07-21

- **Files:**
  `specs/requirements/REQ-FUN.DATA.ontology-visibility-levels.md`
- **Deliverable:**
  - Reframe visibility as a **Project** attribute that applies to the
    paired Ontology via the 1:1 link.
  - Update the SQL snippet: `visibility TEXT CHECK (visibility IN
    ('public', 'internal', 'private'))` is stored on the `scopes` row
    for the Project (not for a separate ontology row).
  - Keep the AI-policy implications (Public/Internal/Private LLM routing)
    unchanged — they still apply to the Ontology content via the 1:1 link.
  - Update «Связанные ADR» to include the new
    `ADR-DES.API.organization-rest-endpoints.md`.
- **Logging:** none.
- **Dependencies:** Task 0.3.
- **Validation:** requirement body refers to `Project` as the visibility
  carrier; the 1:1 application to Ontology is explicit.

#### Task 0.5 — Commit Phase 0

> ✅ **Completed:** 2026-07-21 (submodule `1e56cb5`, main `90a7fb8`)

```
docs(specs): separate Project (container) from Ontology (content)

- context.md: Project is the GitLab Project analogue; Ontology is its
  graph content (1:1).
- ADR-DES.SECURITY.gitlab-like-organization-model: Project carries
  access/versioning; Ontology is content; 1:1 mandatory.
- REQ-NFR.SECURITY.organization-access-model: rewritten Project/Ontology
  structure.
- REQ-FUN.DATA.ontology-visibility-levels: visibility is a Project
  attribute applied to the paired Ontology.
```

---

### Phase 1 — New ADR for organization REST endpoints

> Goal: provide a single ADR that documents the canonical REST surface for
> the organization model. This ADR did not exist before — endpoint shapes
> were scattered across `protocol-stack-strategy`,
> `graphql-sparql-split-strategy`, and `rest-graphql-mutation-boundary`.

#### Task 1.1 — Create `ADR-DES.API.organization-rest-endpoints`

> ✅ **Completed:** 2026-07-21

- **Files:**
  - `specs/adr/ADR-DES.API.organization-rest-endpoints.md` — **NEW**
  - `specs/adr/README.md` — register the new ADR
- **Reference:** `.ai-factory/references/gitlab-projects-groups-api.md`
  — GitLab Projects & Groups API. Use as the authoritative source for
  endpoint shapes, access levels (Guest/Reporter/Developer/Maintainer/Owner
  per § «Access Levels Reference»), pagination (§ «Pagination»), and
  group/project transfer semantics (§ «Transfer a group», § «Transfer a
  project"). The ADR must explicitly cite this reference and document any
  VEDO-specific deviation (e.g., the 1:1 Project ↔ Ontology pairing has
  no GitLab analogue and must be called out as a VEDO extension).
- **Deliverable:**
  - ADR body covers:
    - **Context:** organization model needs a public REST surface aligned
      with GitLab `/groups/:id/...` and `/projects/:id/...`.
    - **Decision:** canonical paths
      - `GET/POST /api/v1/groups`
      - `GET/PUT/DELETE /api/v1/groups/{id}`
      - `GET /api/v1/groups/{id}/subgroups`
      - `GET /api/v1/groups/{id}/members`
      - `GET/POST /api/v1/projects`
      - `GET/PUT/DELETE /api/v1/projects/{id}`
      - `PUT /api/v1/projects/{id}/move`
      - `GET/POST /api/v1/projects/{id}/members`
      - `PUT/DELETE /api/v1/projects/{id}/members/{userId}`
      - `GET/PUT /api/v1/projects/{id}/visibility`
      - `GET/POST /api/v1/projects/{id}/policies`
      - `DELETE /api/v1/projects/{id}/policies/{policyId}`
    - **Project ↔ Ontology pairing (1:1):** creating a Project creates a
      paired Ontology; the Ontology is reached via
      `GET /api/v1/ontologies/{id}` where `{id}` is the ontology id (not
      the project id). The mapping Project → Ontology is exposed via
      `ProjectDetail.ontology_id` and `OntologyDetail.project_id`.
      **VEDO extension:** GitLab has no analogue for this 1:1 pairing
      (a GitLab Project *is* the repository; in VEDO the Project is the
      workspace and the Ontology is the content). See
      `.ai-factory/references/gitlab-projects-groups-api.md` § «Create a
      project» for the GitLab baseline.
    - **Scope strings** in the auth-service gRPC contract: `"group/" + id`
      and `"project/" + id`. Legacy `"ontology/" + id` is accepted only
      during the migration window of this plan and is removed in the
      follow-up cleanup task.
    - **Idempotency:** all write endpoints accept `Idempotency-Key`
      (REQ-FUN.API.write-idempotency). Membership, visibility, and policy
      endpoints **require** it (400 `INVALID_IDEMPOTENCY_KEY` if missing).
    - **RBAC:** only `Owner` can manage members, visibility, and policies.
      `Maintainer` is scoped to the ontology workflow (branches, merge
      requests) and cannot touch org endpoints. Access levels mirror
      GitLab per `.ai-factory/references/gitlab-projects-groups-api.md`
      § «Access Levels Reference» (Guest=10, Reporter=20, Developer=30,
      Maintainer=40, Owner=50). VEDO legacy aliases `Viewer`/`Editor`
      map to `Guest`/`Developer` respectively (per Antora
      `organization-model.adoc` § «Role Hierarchy»).
    - **Audit:** every write emits a structured `audit_events` row
      (event, reason, user_id, object_type=`project|group`, object_id,
      source_ip, trace_id, timestamp).
    - **Rejected alternatives:**
      - `Ontology` as the REST carrier for members/visibility/policies
        (the status quo before this ADR) — rejected because it conflates
        the platform container with the graph content and breaks the
        GitLab alignment (see
        `.ai-factory/references/gitlab-projects-groups-api.md` § «Project
        members» — members live on `/projects/:id/members`, not on a
        repository-content path).
      - Aliases `/ontologies/{id}/members` ↔ `/projects/{id}/members` —
        rejected because under 1:1 they only mask the model and complicate
        audit/idempotency routing.
      - Nesting members under `/groups/{id}/projects/{id}/members` —
        rejected as over-nesting; GitLab keeps `/projects/:id/members`
        flat (§ «Project members»).
      - Deviating from GitLab access level integers — rejected to keep
        the role mapping intuitive for users coming from GitLab; the
        numeric ladder (10/20/30/40/50) is preserved as the internal
        authorization scale even though REST payloads use string role
        names.
    - **Связанные ADR:** `gitlab-like-organization-model`,
      `protocol-stack-strategy`, `graphql-sparql-split-strategy`,
      `rest-graphql-mutation-boundary`, `write-idempotency-strategy`.
  - Update `specs/adr/README.md` index with the new ADR entry.
- **Logging:** none.
- **Dependencies:** Task 0.5.
- **Validation:** ADR file exists, is listed in `README.md`, and its
  endpoint table matches `openapi.json` after Phase 2. ADR body cites
  `.ai-factory/references/gitlab-projects-groups-api.md` at least once
  in the RBAC and rejected-alternatives sections.

#### Task 1.2 — Update `ADR-DES.API.protocol-stack-strategy` table row

> ✅ **Completed:** 2026-07-21

- **Files:** `specs/adr/ADR-DES.API.protocol-stack-strategy.md`
- **Deliverable:**
  - In the «Разделение ответственности протоколов» table, replace the row
    `Управление орг. моделью (groups, projects, members, policies)` with:
    `Управление орг. моделью (groups, projects, members, visibility,
    policies)` and reference `ADR-DES.API.organization-rest-endpoints.md`
    for the canonical endpoint list.
- **Logging:** none.
- **Dependencies:** Task 1.1.
- **Validation:** row references the new ADR.

#### Task 1.3 — Update `ADR-DES.API.graphql-sparql-split-strategy` table

> ✅ **Completed:** 2026-07-21

- **Files:** `specs/adr/ADR-DES.API.graphql-sparql-split-strategy.md`
- **Deliverable:**
  - In the «Разделение ответственности (явная граница протоколов)» table,
    replace the row mentioning `/api/v1/ontologies/{id}/members` with
    `/api/v1/projects/{id}/members` (and analogous paths for visibility
    and policies).
  - Keep the REST/GraphQL split semantics unchanged.
- **Logging:** none.
- **Dependencies:** Task 1.1.
- **Validation:** `grep -n "/ontologies/{id}/members"
  specs/adr/ADR-DES.API.graphql-sparql-split-strategy.md` returns no
  matches.

#### Task 1.4 — Update `ADR-DES.API.rest-graphql-mutation-boundary` table

> ✅ **Completed:** 2026-07-21

- **Files:** `specs/adr/ADR-DES.API.rest-graphql-mutation-boundary.md`
- **Deliverable:**
  - In the «Миграция оставшихся GraphQL mutations в REST» table, update:
    - `updateMemberRole(ontologyId, userId, role)` → target
      `PUT /api/v1/projects/{id}/members/{userId}`
    - `removeMember(ontologyId, userId)` → target
      `DELETE /api/v1/projects/{id}/members/{userId}`
  - The rationale column changes from «REST эндпоинт уже существует в
    routes.go» to «REST эндпоинт канонизирован в
    ADR-DES.API.organization-rest-endpoints.md».
  - Update the checklist item about OpenAPI containing all write-endpoints
    to reference the new project-scoped paths.
- **Logging:** none.
- **Dependencies:** Task 1.1.
- **Validation:** migration table no longer references
  `/ontologies/{id}/members`.

#### Task 1.5 — Commit Phase 1

> ✅ **Completed:** 2026-07-21 (submodule `6a73a73`, main `daea0fa`)

```
docs(adr): add ADR-DES.API.organization-rest-endpoints

Add a dedicated ADR for the organization REST surface, aligned with
GitLab /projects/:id and /groups/:id. Members, visibility, and policies
are scoped to Project (not Ontology), preserving the 1:1 Project ↔
Ontology invariant.

Update protocol-stack-strategy, graphql-sparql-split-strategy, and
rest-graphql-mutation-boundary to reference the new ADR and the renamed
project-scoped endpoints.
```

---

### Phase 2 — OpenAPI contract (`openapi.json`)

> Goal: make the OpenAPI spec match the new ADR. All org-write endpoints
> move from `/ontologies/{id}/...` to `/projects/{id}/...`, and
> `ProjectSummary.ontology_count` is removed.

#### Task 2.1 — Remove `ProjectSummary.ontology_count` and align Project schemas

> ✅ **Completed:** 2026-07-21

- **Files:** `src/services/api-gateway/docs/openapi.json`
- **Deliverable:**
  - In `ProjectSummary` (L2752-2762): remove the `ontology_count` field.
  - In `ProjectDetail` (L2763-2774): add `ontology_id` (type string,
    description «ID of the 1:1 paired Ontology»).
  - In `CreateProjectRequest` (L2827-2835): keep `label`, `description`,
    `group_id` unchanged.
- **Logging:** none.
- **Dependencies:** Task 1.5.
- **Validation:** `jq '.components.schemas.ProjectSummary |
  has("ontology_count")' openapi.json` returns `false`; `jq
  '.components.schemas.ProjectDetail.properties | has("ontology_id")'
  openapi.json` returns `true`.

#### Task 2.2 — Move `/ontologies/{id}/members` → `/projects/{id}/members`

> ✅ **Completed:** 2026-07-21

- **Files:** `src/services/api-gateway/docs/openapi.json` (L1325-1410)
- **Deliverable:**
  - Replace path key `/ontologies/{id}/members` with
    `/projects/{id}/members`.
  - Update `operationId`: `listOntologyMembers` → `listProjectMembers`;
    `addOntologyMember` → `addProjectMember`.
  - Update `summary`/`description` to reference Project.
  - Keep parameters (`id`, pagination, idempotency) and
    `AddMemberRequest` / `MemberSummary` schemas unchanged.
- **Logging:** none.
- **Dependencies:** Task 2.1.
- **Validation:** `jq '.paths | has("/projects/{id}/members")'
  openapi.json` returns `true`; `jq '.paths | has("/ontologies/{id}/members")'
  openapi.json` returns `false`.

#### Task 2.3 — Move `/ontologies/{id}/members/{userId}` → `/projects/{id}/members/{userId}`

> ✅ **Completed:** 2026-07-21

- **Files:** `src/services/api-gateway/docs/openapi.json` (L1411-1497)
- **Deliverable:**
  - Replace path key. Update `operationId`s
    (`updateOntologyMemberRole` → `updateProjectMemberRole`,
    `removeOntologyMember` → `removeProjectMember`).
  - Keep `UpdateMemberRoleRequest` schema unchanged.
- **Logging:** none.
- **Dependencies:** Task 2.2.
- **Validation:** old path absent; new path present with PUT and DELETE
  operations.

#### Task 2.4 — Move `/ontologies/{id}/visibility` → `/projects/{id}/visibility`

> ✅ **Completed:** 2026-07-21

- **Files:** `src/services/api-gateway/docs/openapi.json` (L1498-1576)
- **Deliverable:**
  - Replace path key. Update `operationId`s
    (`getOntologyVisibility` → `getProjectVisibility`,
    `setOntologyVisibility` → `setProjectVisibility`).
  - Keep `SetVisibilityRequest` / `VisibilityLevel` schemas unchanged.
- **Logging:** none.
- **Dependencies:** Task 2.2.
- **Validation:** old path absent; new path present with GET and PUT.

#### Task 2.5 — Move `/ontologies/{id}/policies` → `/projects/{id}/policies`

> ✅ **Completed:** 2026-07-21

- **Files:** `src/services/api-gateway/docs/openapi.json` (L1577-1659)
- **Deliverable:**
  - Replace path key. Update `operationId`s
    (`listOntologyPolicies` → `listProjectPolicies`,
    `createOntologyPolicy` → `createProjectPolicy`).
  - Keep `CreatePolicyRequest` / `PolicySummary` schemas unchanged.
- **Logging:** none.
- **Dependencies:** Task 2.2.
- **Validation:** old path absent; new path present with GET and POST.

#### Task 2.6 — Move `/ontologies/{id}/policies/{policyId}` → `/projects/{id}/policies/{policyId}`

> ✅ **Completed:** 2026-07-21

- **Files:** `src/services/api-gateway/docs/openapi.json` (L1660-1693)
- **Deliverable:**
  - Replace path key. Update `operationId`
    (`deleteOntologyPolicy` → `deleteProjectPolicy`).
- **Logging:** none.
- **Dependencies:** Task 2.5.
- **Validation:** old path absent; new path present with DELETE.

#### Task 2.7 — Revisit `CreateOntologyRequest` under 1:1

> ✅ **Completed:** 2026-07-21

- **Files:** `src/services/api-gateway/docs/openapi.json` (L2880-2889)
- **Deliverable:**
  - Under 1:1, an Ontology is created implicitly when its Project is
    created. Remove `project_id` from `CreateOntologyRequest` (the
    pairing is automatic).
  - Keep the explicit `/ontologies` POST endpoint for the legacy
    «import standalone ontology» flow, but mark it as deprecated in
    `description` and add `x-vedo-status: deprecated` (per the stub
    policy in ROADMAP.md § Stub policy).
  - Add `x-vedo-status: canonical` to `POST /projects` to signal that
    project creation is the preferred path.
- **Logging:** none.
- **Dependencies:** Task 2.1.
- **Validation:** `CreateOntologyRequest` no longer lists `project_id`;
  `POST /projects` carries `x-vedo-status: canonical`.

#### Task 2.8 — Validate OpenAPI JSON

> ✅ **Completed:** 2026-07-21 (JSON parses, 0 orphaned $refs; redocly not available — WARN skipped)

- **Files:** `src/services/api-gateway/docs/openapi.json`
- **Deliverable:**
  - Run `python -c "import json; json.load(open('src/services/api-gateway/docs/openapi.json'))"`
    to confirm the file parses.
  - Run `npx redocly lint src/services/api-gateway/docs/openapi.json` (if
    available) and fix any errors. If `redocly` is not available, run
    `npx @redocly/cli@latest lint` as a one-off.
  - Manually verify no `$ref` targets were orphaned by the path moves.
- **Logging:** none.
- **Dependencies:** Tasks 2.1–2.7.
- **Validation:** JSON parses; redocly reports 0 errors.

#### Task 2.9 — Commit Phase 2

> ✅ **Completed:** 2026-07-21 (main `3b93710`)

```
docs(openapi): move org endpoints to /projects/{id}/...

Move members, visibility, and policies endpoints from
/ontologies/{id}/... to /projects/{id}/... to match the new
organization REST ADR. Drop ProjectSummary.ontology_count (always 1
under the 1:1 invariant) and add ProjectDetail.ontology_id for the
1:1 pairing. Mark POST /ontologies as deprecated in favour of
POST /projects.
```

---

### Phase 3 — Antora documentation

> Goal: make the public docs consistent with the new model and the new ADR.

#### Task 3.1 — Update `organization-model.adoc`

> ✅ **Completed:** 2026-07-21

- **Files:**
  `src/docs/antora/developer-guide/modules/ROOT/pages/organization-model.adoc`
- **Deliverable:**
  - In «Key Concepts»: split `Project (Ontology): scope node (type=ontology)
    under a group` into two bullets:
    - `Project — scope node (type=project) under a group; the unit of
      access, versioning, and membership.`
    - `Ontology — graph content of a Project (TBox/ABox, classes,
      properties, individuals). Ratio 1:1 with Project.`
  - In «PostgreSQL Schema»: document the new `scopes.type='project'` and
    the new `ontologies` table introduced in Phase 4. Mention that
    `type='ontology'` is gone after migration 007.
  - In «REST API Endpoints»: rename every `:id` from `ontologies` to
    `projects` for members, visibility, and policies. Add
    `GET /api/v1/projects/:id` returning `ontology_id`.
  - Add a note pointing to the new
    `xref:project-ontology-model.adoc[Project ↔ Ontology Model]` page.
- **Logging:** none.
- **Dependencies:** Task 2.9.
- **Validation:** `grep -n "ontologies/:id/members"
  src/docs/antora/developer-guide/modules/ROOT/pages/organization-model.adoc`
  returns no matches.

#### Task 3.2 — Create `project-ontology-model.adoc`

> ✅ **Completed:** 2026-07-21

- **Files:**
  - `src/docs/antora/developer-guide/modules/ROOT/pages/project-ontology-model.adoc`
    — **NEW**
  - `src/docs/antora/developer-guide/modules/ROOT/nav.adoc` — register
- **Deliverable:**
  - New Antora page that explains:
    - The 1:1 invariant and why it matches GitLab «one repo = one
      Project».
    - The responsibility split: Project = workspace, Ontology = graph
      content.
    - The REST surface (linking to
      `xref:integrator-guide:api-reference.adoc[]`).
    - The PostgreSQL schema (scopes + ontologies table).
    - The migration story (type rename + ontologies backfill).
  - Add the page to `nav.adoc` after `organization-model.adoc`.
- **Logging:** none.
- **Dependencies:** Task 3.1.
- **Validation:** `antora antora-playbook.yml` succeeds (if run) and the
  new page appears in the nav.

#### Task 3.3 — Update `api-reference.adoc`

> ✅ **Completed:** 2026-07-21

- **Files:**
  `src/docs/antora/integrator-guide/modules/ROOT/pages/api-reference.adoc`
- **Deliverable:**
  - In «Organization & Access Control → Members / Visibility / Access
    Policies»: replace every `/api/v1/ontologies/{id}/...` with
    `/api/v1/projects/{id}/...`.
  - In «Projects» section: add `GET /api/v1/projects/{id}` returning
    `ontology_id`, and document the 1:1 pairing.
  - Update the introductory paragraph to reference the new
    `ADR-DES.API.organization-rest-endpoints.md`.
- **Logging:** none.
- **Dependencies:** Task 3.2.
- **Validation:** `grep -n "/api/v1/ontologies/{id}/members"
  src/docs/antora/integrator-guide/modules/ROOT/pages/api-reference.adoc`
  returns no matches.

#### Task 3.4 — Commit Phase 3

> ✅ **Completed:** 2026-07-21 (main `0cb8952`)

```
docs(antora): document Project ↔ Ontology 1:1 model

- organization-model.adoc: split Project and Ontology concepts; document
  the new scopes.type='project' and ontologies table; rename REST
  endpoints to /projects/{id}/...
- project-ontology-model.adoc: new dedicated page explaining the 1:1
  invariant and GitLab alignment.
- api-reference.adoc: integrator-facing endpoint list updated.
```

---

### Phase 4 — PostgreSQL schema migration

> Goal: bring the `vedo_org` schema in line with the 1:1 model. Introduce
> the `ontologies` table and rename the legacy `type='ontology'` to
> `type='project'`.

#### Task 4.1 — Write unit tests for the new schema (TDD)

> ✅ **Completed:** 2026-07-21 (tests skip gracefully without PostgreSQL)

> BDD naming: `[Condition]_[Action]_[ExpectedResult]`
> Anti-patterns: see `.ai-factory/rules/test-quality.md`

- **Files:**
  - `src/services/auth-service/org/schema_test.go` — **NEW**
- **Deliverable:**
  - Tests cover:
    - `ProjectCreated_HasPairedOntologyRow_OntologyIdReturned`
    - `ProjectDeleted_CascadesToOntology_OntologyRowGone`
    - `OntologyCreated_Standalone_Rejected` (1:1 invariant: an Ontology
      cannot exist without a Project)
    - `LegacyOntologyScope_Migrated_BecomesProjectScopeWithTypeRename`
  - Tests use a real PostgreSQL test database (or `pgtest` if already
    wired in the service) — no mock assertions for the schema layer.
  - Each test carries a `// Validates: REQ-NFR.SECURITY.organization-access-model`
    annotation.
- **Logging:** standard — `slog.Info("schema.test", "case", ...)`.
- **Dependencies:** Task 3.4.
- **Validation:** `go test ./org/ -run Schema` passes; TQS ≥ bronze.

#### Task 4.2 — Create migration 007 (rename `type='ontology'` → `type='project'`)

> ✅ **Completed:** 2026-07-21

- **Files:**
  `src/services/auth-service/migrations/007_rename_ontology_scope_to_project.sql`
  — **NEW**
- **Deliverable:**
  - Migration is idempotent (`UPDATE ... WHERE type='ontology'`).
  - Updates the `scopes.type` CHECK constraint to
    `type IN ('group', 'project')`.
  - Includes a rollback comment block (no down-migration framework in
    place — the comment documents the inverse UPDATE for manual
    rollback).
- **Logging:** none (SQL migration).
- **Dependencies:** Task 4.1 (tests must fail first, then pass after the
  migration is applied in the test setup).
- **Validation:** migration applies cleanly against a `vedo_org`
  database seeded by migrations 001-006; the
  `ProjectCreated_HasPairedOntologyRow_OntologyIdReturned` test passes
  after migration 008 is also applied.

#### Task 4.3 — Create migration 008 (`ontologies` table)

> ✅ **Completed:** 2026-07-21

- **Files:**
  `src/services/auth-service/migrations/008_create_ontologies_table.sql`
  — **NEW**
- **Deliverable:**
  - Table `ontologies`:
    - `project_scope TEXT PRIMARY KEY REFERENCES scopes(id) ON DELETE
      CASCADE` (1:1 enforced by PK = FK)
    - `ontology_id TEXT NOT NULL UNIQUE` (the id used by
      `ontology-service` to identify the graph)
    - `iri TEXT NOT NULL DEFAULT ''`
    - `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`
    - `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`
  - The `project_scope` PK enforces 1:1 (one ontology row per project
    scope).
- **Logging:** none.
- **Dependencies:** Task 4.2.
- **Validation:** schema test `ProjectCreated_HasPairedOntologyRow`
  passes after migrations 007 + 008 are applied.

#### Task 4.4 — Create migration 009 (backfill `ontologies` from legacy scopes)

> ✅ **Completed:** 2026-07-21

- **Files:**
  `src/services/auth-service/migrations/009_seed_ontologies_from_legacy.sql`
  — **NEW**
- **Deliverable:**
  - Before the type rename in 007 is applied, this migration copies the
    existing `type='ontology'` scopes into the `ontologies` table with
    `project_scope = scopes.id` and `ontology_id = scopes.id` (legacy
    identity).
  - The migration is ordered **before** 007 in the migration runner. If
    the runner applies files alphabetically, rename to
    `006_5_seed_ontologies.sql` or renumber — confirm the runner's
    ordering contract and apply the correct filename.
  - Idempotent: `INSERT ... ON CONFLICT (project_scope) DO NOTHING`.
- **Logging:** none.
- **Dependencies:** Task 4.3.
- **Validation:** after all three migrations, every existing
  `type='ontology'` scope has a paired row in `ontologies` with
  `type='project'` in `scopes`.

#### Task 4.5 — Update `auth-service` org code for the new schema

> ✅ **Completed:** 2026-07-21 (go vet + go test + go build all pass)

- **Files:**
  - `src/services/auth-service/org/*.go` (all files in `org/`)
  - `src/services/auth-service/membership/*.go`
- **Deliverable:**
  - Replace `ScopeTypeOntology = "ontology"` constant with
    `ScopeTypeProject = "project"`.
  - Project creation flow: insert a `scopes` row with
    `type='project'` AND a paired `ontologies` row in a single
    transaction. If either insert fails, roll back.
  - Project delete flow: cascades to `ontologies` via the FK
    `ON DELETE CASCADE`.
  - `ListProjects` / `GetProject` gRPC handlers return the paired
    `ontology_id` from the `ontologies` table.
  - Membership/visibility/policy handlers now accept scope strings of
    the form `"project/" + id`. A temporary compatibility shim accepts
    `"ontology/" + id` and rewrites it to `"project/" + id` with a
    `WARN` log (`scope.legacy_alias`, `alias=ontology`,
    `target=project`). The shim is removed in Task 7.6.
- **Logging:**
  - `slog.Info("project.create", "project_id", ..., "ontology_id", ...)`
  - `slog.Warn("scope.legacy_alias", "alias", "ontology", "target",
    "project", "scope", scope)`
  - `slog.Error("project.create.ontology_pairing_failed", "project_id",
    ..., "err", ...)`
- **Dependencies:** Task 4.4.
- **Validation:** `go test ./org/...` passes; the legacy alias shim is
  covered by `LegacyOntologyScope_Migrated_BecomesProjectScope`.

#### Task 4.6 — Update gRPC proto contract

> ✅ **Completed:** 2026-07-21 (proto regenerated with protoc; vendored copies synced)

- **Files:**
  - `src/services/auth-service/internal/proto/org.proto` (or
    `src/services/shared/proto/` — confirm canonical location)
  - regenerated `*.pb.go` files
- **Deliverable:**
  - `ProjectDetail` message: drop `ontology_count`; add `ontology_id
    string`.
  - `CreateProjectResponse` carries both `project_id` and
    `ontology_id`.
  - `ListProjectsResponse` items use the updated `ProjectDetail`.
  - Run `make proto` (or `buf generate`) to regenerate bindings.
- **Logging:** none (proto contract).
- **Dependencies:** Task 4.5.
- **Validation:** `go build ./...` succeeds; `org_proto_generate_test.go`
  passes.

#### Task 4.7 — Commit Phase 4

> ✅ **Completed:** 2026-07-21 (main `307a9d4`)

```
feat(auth-service): rename scopes.type='ontology' to 'project' and add ontologies table

- migrations 007-009: rename type, add ontologies table (1:1 with
  project scope), backfill from legacy rows.
- org/membership Go code: project create/delete pairs with ontologies;
  scope strings switch to "project/" prefix with a temporary legacy
  "ontology/" alias shim.
- gRPC proto: ProjectDetail drops ontology_count, adds ontology_id.
```

---

### Phase 5 — API Gateway code

> Goal: align `routes.go` and `org_handler.go` with the renamed endpoints.

#### Task 5.1 — Update `routes.go` endpoint registrations

> ✅ **Completed:** 2026-07-21

- **Files:** `src/services/api-gateway/routes.go` (L80-110)
- **Deliverable:**
  - Replace:
    - `api.GET("/ontologies/:id/members", ...)` →
      `api.GET("/projects/:id/members", ...)`
    - `api.GET("/ontologies/:id/visibility", ...)` →
      `api.GET("/projects/:id/visibility", ...)`
    - `api.GET("/ontologies/:id/policies", ...)` →
      `api.GET("/projects/:id/policies", ...)`
    - `orgWrite.POST("/ontologies/:id/members", ...)` →
      `orgWrite.POST("/projects/:id/members", ...)`
    - `orgWrite.PUT("/ontologies/:id/members/:userId", ...)` →
      `orgWrite.PUT("/projects/:id/members/:userId", ...)`
    - `orgWrite.DELETE("/ontologies/:id/members/:userId", ...)` →
      `orgWrite.DELETE("/projects/:id/members/:userId", ...)`
    - `orgWrite.PUT("/ontologies/:id/visibility", ...)` →
      `orgWrite.PUT("/projects/:id/visibility", ...)`
    - `orgWrite.POST("/ontologies/:id/policies", ...)` →
      `orgWrite.POST("/projects/:id/policies", ...)`
    - `orgWrite.DELETE("/ontologies/:id/policies/:policyId", ...)` →
      `orgWrite.DELETE("/projects/:id/policies/:policyId", ...)`
  - Update `Idempotency.CriticalPaths` from `[]string{"/api/v1/ontologies/"}`
    to `[]string{"/api/v1/projects/"}`.
  - Keep the `/api/v1/ontologies/:id/classes|properties|individuals`
    routes unchanged — they are the graph-content endpoints and stay
    under `/ontologies/{id}`.
- **Logging:** `slog.Info("route.register", "method", "GET",
  "path", "/api/v1/projects/:id/members")` at startup is optional but
  useful for debugging the rename.
- **Dependencies:** Task 4.7.
- **Validation:** `go build ./...` succeeds; `go test ./handlers/...`
  passes after Task 5.2.

#### Task 5.2 — Update `org_handler.go` scope strings

> ✅ **Completed:** 2026-07-21

- **Files:**
 `src/services/api-gateway/handlers/org_handler.go`
- **Deliverable:**
  - In `HandleListMembers`, `HandleAddMember`, `HandleUpdateMemberRole`,
    `HandleRemoveMember`, `HandleSetVisibility`, `HandleGetVisibility`,
    `HandleListPolicies`, `HandleCreatePolicy`, `HandleDeletePolicy`:
    replace `scope := "ontology/" + c.Param("id")` with `scope :=
    "project/" + c.Param("id")`.
  - Update the route-table comment block at the top of the file to
    reflect the new paths.
  - Keep the `groupId` branch in `HandleListMembers` unchanged
    (`"group/" + gid`).
- **Logging:** standard — `slog.Warn("scope.legacy_alias",
  "alias", "ontology", "target", "project")` is emitted by the
  auth-service shim (Task 4.5), not by the gateway.
- **Dependencies:** Task 5.1.
- **Validation:** `grep -n '"ontology/" +' src/services/api-gateway/handlers/org_handler.go`
  returns no matches outside of the legacy alias documentation comment.

#### Task 5.3 — Update `org_handler_test.go` expectations

> ✅ **Completed:** 2026-07-21 (new test file; go test ./handlers/ -run OrgHandler passes)

- **Files:**
  `src/services/api-gateway/handlers/org_handler_test.go` (if present),
  or add a new test file
- **Deliverable:**
  - Update route expectations to the new paths.
  - Add a negative test: `UnknownOrgPath_Returns404` — confirms that
    `/api/v1/ontologies/:id/members` is no longer registered and
    returns `404 GATEWAY-NOT-FOUND`.
  - Add a positive test: `ProjectMembersList_Returns200` — confirms
    the new path dispatches to `HandleListMembers`.
  - Tests carry `// Validates: ADR-DES.API.organization-rest-endpoints`
    annotations.
- **Logging:** test-only.
- **Dependencies:** Task 5.2.
- **Validation:** `go test ./handlers/... -run OrgHandler` passes; TQS
  ≥ bronze.

#### Task 5.4 — Commit Phase 5

> ✅ **Completed:** 2026-07-21 (main `683cc26`)

```
refactor(api-gateway): route org endpoints under /projects/{id}/...

- routes.go: rename /ontologies/:id/{members,visibility,policies} to
  /projects/:id/...; update Idempotency.CriticalPaths.
- org_handler.go: scope strings switch to "project/" prefix.
- org_handler_test.go: update expectations and add negative test for
  the removed /ontologies/:id/members path.
```

---

### Phase 6 — Frontend updates

> Goal: switch the Vue UI from `/ontologies/{id}/members|visibility|policies`
> to `/projects/{id}/...`.

#### Task 6.1 — Update Apollo queries and mocks

> ✅ **Completed:** 2026-07-21 (mock-link.ts has no member fixtures — no-op)

- **Files:**
  - `src/services/frontend/src/apollo/queries.ts`
  - `src/services/frontend/src/apollo/mock-link.ts`
- **Deliverable:**
  - In `LIST_MEMBERS_QUERY`, `UPDATE_MEMBER_ROLE_MUTATION`,
    `REMOVE_MEMBER_MUTATION` (or their REST-call equivalents): switch
    from `/api/v1/ontologies/${ontologyId}/members` to
    `/api/v1/projects/${projectId}/members`.
  - Update `mock-link.ts` fixtures to match the new paths.
  - Add a typed `projectId` variable to each query/mutation.
- **Logging:** standard — surface network errors via `useErrorPresentation`.
- **Dependencies:** Task 5.4.
- **Validation:** `biome check src/services/frontend/src/apollo/` passes;
  existing frontend tests still pass.

#### Task 6.2 — Update `MembersPage.vue` and related pages

> ✅ **Completed:** 2026-07-21 (biome + vitest 166/166 pass; ProjectsPage.vue and OntologyWorkspace.vue have no member links)

- **Files:**
  - `src/services/frontend/src/pages/MembersPage.vue`
  - `src/services/frontend/src/pages/ProjectsPage.vue`
  - `src/services/frontend/src/pages/OntologyWorkspace.vue` (if it links
    to members)
- **Deliverable:**
  - `MembersPage.vue` reads `projectId` from the route (or resolves it
    from `ontologyId` via the 1:1 link exposed by
    `GET /api/v1/ontologies/{id}` — the response now includes
    `project_id`).
  - `ProjectsPage.vue` project rows link to
    `/projects/{id}/members` (not `/ontologies/{id}/members`).
  - Add an empty-state CTA: «This project has no members yet. Add one.»
- **Logging:** standard — `console.error` is replaced by
  `useErrorPresentation` composable (per existing convention).
- **Dependencies:** Task 6.1.
- **Validation:** `biome check` passes; `vitest run` passes for the
  affected components.

#### Task 6.3 — Commit Phase 6

> ✅ **Completed:** 2026-07-21 (main `63a69ec`)

```
refactor(frontend): use /projects/{id}/members|visibility|policies

- queries.ts / mock-link.ts: switch member/visibility/policy Apollo
  operations to project-scoped paths.
- MembersPage.vue / ProjectsPage.vue: route by projectId; add empty
  state CTA.
```

---

### Phase 7 — Tests, security, and traceability

> Goal: lock down the renamed surface with negative security tests and
> keep `traceability.ttl` in sync per RULES.md § Traceability.

#### Task 7.1 — Update `tests/security/` BOLA/BFLA negative tests

> ✅ **Completed:** 2026-07-21 (6 new tests added; paths updated to /projects/:id/...)

> BDD naming: `[Condition]_[Action]_[ExpectedResult]`
> Security tests must dispatch a real HTTP request, not use mock
> assertions (per skill-context «Security Test Tasks in Plan»).

- **Files:**
  - `tests/security/` — update existing BOLA/BFLA suites for the renamed
    endpoints; add new ones if no coverage existed
- **Deliverable:**
  - Add tests:
    - `NonMember_GetProjectMembers_Returns403`
    - `Reporter_PostProjectMember_Returns403` (only Owner can add)
    - `Maintainer_PutProjectVisibility_Returns403` (only Owner can
      change visibility)
    - `Anonymous_GetPrivateProjectMembers_Returns401`
    - `Owner_PostProjectMember_WithoutIdempotencyKey_Returns400`
    - `Owner_PostProjectMember_WithReusedKeyAndDifferentPayload_Returns409`
  - Each test carries `// Validates: REQ-NFR.SECURITY.authorization-regression-gates`
    and `// Validates: ADR-DES.API.organization-rest-endpoints`.
  - Tests hit a real `api-gateway` instance via HTTP (no mocking).
- **Logging:** test-only.
- **Dependencies:** Task 6.3.
- **Validation:** `go test ./tests/security/...` passes; TQS ≥ bronze.

#### Task 7.2 — Update `tests/e2e/playwright/` flows

> ✅ **Completed:** 2026-07-21 (org-api.spec.ts paths updated; members.page.ts uses projectId; new 1:1 pairing test added)

- **Files:** `tests/e2e/playwright/` (any spec that references the old
  endpoints)
- **Deliverable:**
  - Update flow `group → project → members → add member` to use the
    project-scoped URL.
  - Add an e2e check: after creating a Project, the paired Ontology is
    reachable via `GET /api/v1/ontologies/{ontologyId}` where
    `ontologyId` comes from `ProjectDetail.ontology_id`.
- **Logging:** standard Playwright tracing.
- **Dependencies:** Task 7.1.
- **Validation:** `npx playwright test` passes for the updated specs.

#### Task 7.3 — Update `traceability.ttl`

> ✅ **Completed:** 2026-07-21 (ADR, code artifacts, and test triples registered)

- **Files:** `.ai-factory/traceability/traceability.ttl`
- **Deliverable:**
  - Add `vdo:ADR` entry for
    `ADR-DES.API.organization-rest-endpoints`.
  - Add `vdo:implements` triples linking the new ADR to:
    - `vdo:Endpoint` nodes for each renamed path
      (`/projects/{id}/members`, `/projects/{id}/visibility`,
      `/projects/{id}/policies`)
    - `vdo:Service` node `auth-service` (scope handling)
    - `vdo:Service` node `api-gateway` (routing)
  - Add `vdo:validates` triples for each new test file from Tasks 7.1
    and 7.2, linking them to
    `REQ-NFR.SECURITY.authorization-regression-gates` and
    `ADR-DES.SECURITY.gitlab-like-organization-model`.
  - Remove stale `vdo:Endpoint` nodes for
    `/ontologies/{id}/members|visibility|policies` (mark
    `vdo:status "deprecated"` if the ontology POST is still exposed as
    deprecated; otherwise remove).
- **Logging:** none.
- **Dependencies:** Task 7.2.
- **Validation:** `rapper -i turtle -o ntriples traceability.ttl`
  parses without errors (or equivalent `python rdflib` check).

#### Task 7.4 — Remove the legacy `"ontology/" + id` scope shim

> ✅ **Completed:** 2026-07-21 (shim removed from ParseScope; no legacy_alias references remain)

- **Files:**
  - `src/services/auth-service/membership/*.go`
  - `src/services/auth-service/org/*.go`
- **Deliverable:**
  - Remove the temporary compatibility shim added in Task 4.5.
  - The auth-service now accepts only `"project/" + id` and
    `"group/" + id` scope strings.
  - Update any test that exercised the shim.
- **Logging:** `slog.Warn("scope.legacy_alias.removed")` is NOT emitted
  anymore — the alias is gone.
- **Dependencies:** Task 7.3 (traceability first, so the shim removal is
  traceable).
- **Validation:** `grep -n 'legacy_alias' src/services/auth-service/`
  returns no matches.

#### Task 7.5 — Final docs checkpoint (`$aif-docs`)

> ✅ **Completed:** 2026-07-21 (manual verification: ADR in README, Antora nav updated, no stale refs in docs/README)

- **Files:** docs (Antora + README + ADR index)
- **Deliverable:**
  - Run `$aif-docs` to verify consistency:
    - All ADRs referenced in the plan are listed in `specs/adr/README.md`.
    - The Antora nav includes the new `project-ontology-model.adoc`.
    - The OpenAPI `info.description` still matches the new surface.
    - No stale references to `/ontologies/{id}/members|visibility|policies`
      remain in `src/docs/antora/` or `README.md`.
  - Fix any drift `$aif-docs` reports.
- **Logging:** none.
- **Dependencies:** Task 7.4.
- **Validation:** `$aif-docs` exits clean.

#### Task 7.6 — Commit Phase 7

> ✅ **Completed:** 2026-07-21 (main `a0cb8db`)

```
test(security): cover renamed /projects/{id}/... org endpoints

- tests/security: BOLA/BFLA negative tests for project-scoped members,
  visibility, and policies.
- tests/e2e/playwright: update group → project → members flow; verify
  paired ontology reachability.
- traceability.ttl: register ADR-DES.API.organization-rest-endpoints and
  new test triples.
- auth-service: remove the temporary "ontology/" + id scope alias shim.
```

---

## Commit Plan

Group logically related tasks into one commit every 3-5 tasks:

| Commit | After tasks | Message (subject only) |
|---|---|---|
| 1 | 0.1–0.5 | `docs(specs): separate Project (container) from Ontology (content)` |
| 2 | 1.1–1.5 | `docs(adr): add ADR-DES.API.organization-rest-endpoints` |
| 3 | 2.1–2.9 | `docs(openapi): move org endpoints to /projects/{id}/...` |
| 4 | 3.1–3.4 | `docs(antora): document Project ↔ Ontology 1:1 model` |
| 5 | 4.1–4.7 | `feat(auth-service): rename scopes.type to 'project' and add ontologies table` |
| 6 | 5.1–5.4 | `refactor(api-gateway): route org endpoints under /projects/{id}/...` |
| 7 | 6.1–6.3 | `refactor(frontend): use /projects/{id}/members\|visibility\|policies` |
| 8 | 7.1–7.6 | `test(security): cover renamed /projects/{id}/... org endpoints` |

All commit subjects follow the conventional-commits style and stay under
50 characters. Bodies are wrapped at 72 columns.

## Acceptance Criteria

- [ ] All tests pass: `go test ./...` / `cargo test` (n/a) / `pytest` (n/a) /
      `vitest` / `npx playwright test`
- [ ] Test Quality Score (TQS) ≥ bronze (6.0) for new test files
      (Tasks 4.1, 5.3, 7.1, 7.2)
- [ ] No B1–B7 anti-patterns (see `.ai-factory/rules/test-quality.md`)
- [ ] Traceability annotations present (`// Validates: REQ-...`,
      `// Validates: ADR-...`)
- [ ] `traceability.ttl` updated with new ADR, endpoints, services, and
      test triples (Task 7.3)
- [ ] No stale references to `/ontologies/{id}/members|visibility|policies`
      in `specs/`, `src/docs/antora/`, `src/services/api-gateway/`, or
      `src/services/frontend/src/apollo/`
- [ ] `openapi.json` parses and `redocly lint` reports 0 errors
- [ ] `ADR-DES.API.organization-rest-endpoints.md` exists and is listed in
      `specs/adr/README.md`
- [ ] PostgreSQL migrations 007-009 apply cleanly against a database
      seeded by 001-006
- [ ] The 1:1 invariant holds: creating a Project creates a paired
      Ontology row; deleting a Project cascades to the Ontology
- [ ] Security negative tests prove the bypass is closed for the renamed
      endpoints (real HTTP requests, no mocks)
- [ ] Antora build (`antora antora-playbook.yml`) succeeds and includes
      the new `project-ontology-model.adoc` page

## Risks and Mitigations

| Risk | Mitigation |
|---|---|
| **PostgreSQL migration ordering** — migrations 007-009 must run in the right order, and the runner's contract is not confirmed. | Confirm the runner's ordering (alphabetical vs explicit) before writing 009. If alphabetical, renumber to keep 007 → 008 → 009 ordering. |
| **`type='ontology'` data in production** — the rename could break running auth-service instances during rollout. | Migration 009 backfills the `ontologies` table BEFORE the type rename in 007. Roll out in order: 009 → 007 → 008. The legacy `"ontology/" + id` scope shim (Task 4.5) bridges the window. |
| **Frontend cache of `ontologyId`** — existing pages pass `ontologyId` to member endpoints. Switching to `projectId` may break deep links. | `GET /api/v1/ontologies/{id}` returns `project_id` via the 1:1 pairing; the frontend resolves `projectId` from `ontologyId` transparently in Task 6.2. |
| **gRPC proto regeneration** — `org_proto_generate_test.go` exists and may fail if proto changes are not regenerated. | Task 4.6 runs `make proto` / `buf generate` explicitly. |
| **Deprecation of `POST /ontologies`** — existing integrations may call it directly. | Marked `x-vedo-status: deprecated` in OpenAPI (Task 2.7) but NOT removed. Removal is a follow-up after a deprecation period. |
| **Skill-context security test rule** — requires real HTTP requests, no mocks. | Task 7.1 tests hit a real `api-gateway` via HTTP; test setup spins up the service in the test harness. |

## Out of Scope

- Removing `POST /ontologies` (only deprecation marker added; removal is
  a follow-up after a deprecation window).
- Migrating `tests/ticket-api/` (no coverage of org endpoints today).
- Updating `publish-browse-ui` (public viewer does not touch org
  endpoints).
- Revisiting `ADR-DES.SECURITY.public-ontology-access` (unchanged —
  public visibility still applies via the 1:1 link).

## Next Steps

After implementation:

1. Run `$aif-verify` to confirm all acceptance criteria are met.
2. Run `$aif-review` for a fresh-context review of the renamed surface.
3. Open a PR `feature/project-ontology-separation` → `main` with the
   eight commits from the Commit Plan.
4. After merge, schedule a follow-up plan to remove `POST /ontologies`
   once the deprecation window closes.
