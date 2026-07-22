# Implementation Plan: Tighten GraphQL to Graph-Only Boundary

Branch: feature/graphql-graph-only-boundary
Created: 2026-07-22

## Settings
- Testing: yes
- Logging: verbose
- Docs: yes — mandatory docs checkpoint

## Roadmap Linkage
Milestone: "M5: MVP Scope Gap Closure"
Rationale: This plan corrects the M4 wiring approach that treated GraphQL as a generic BFF (versioning reads, org reads, comments, dashboard, metrics, deployments, merge requests, validation, draft state — all wired through `/api/v1/graphql`). The rule enforced here is: **GraphQL = only ontology graph navigation** (class/property/individual queries, hierarchy, neighborhood, autocomplete). Everything else must be REST. This directly closes M5 gaps: separated graph views, comments wiring to commenting-service, strict CRUD semantics.

## Research Context
Source: `.ai-factory/RESEARCH.md` explore session 2026-07-22

**Decision — from user:** GraphQL is only for fetching information FROM the ontology graph — classes, properties, individuals, and graph navigation (tree, neighborhood, autocomplete). Everything else is REST. Boundary cases decided:
- `ontology(id)` metadata (branch, commit, dirty) → REST
- `ontologyMetrics` → REST
- `compareRevisions` → REST

**Current violation inventory:**
- Backend `QueryRoot` (query.rs): ontology, commits, branch, branches, groups, projects, members — must be removed
- Backend `MutationRoot` (mutation.rs): updateDraft, updateMemberRole, removeMember — must be removed → EmptyMutation
- Frontend `queries.ts`: 25+ non-graph queries/mutations — must be stripped to ~11 graph-only
- Frontend pages: 11 pages import non-graph Apollow queries → migrate to REST axios

**REST infrastructure already exists:**
- Versioning reads: `/versioning/commits`, `/branches` (routes.go:138-150)
- Org reads: `/groups`, `/projects`, `/projects/{id}/members` (routes.go:80-110)
- Ontology metadata: `GET /ontologies/{id}` (routes.go:115)
- Entity writes: `api/ontology.ts` (pattern for new axios clients)
- SPARQL: `api/sparql.ts`
- Metrics: `GET /api/v1/metrics/ontologies` in metrics-service Python (FastAPI) — needs Gateway proxy
- Publisher: `/api/v1/ontologies/{id}/snapshots` in publisher-service Rust (axum) — needs Gateway proxy

**New REST endpoints needed:**
- Comments: NO HTTP in commenting-service (gRPC only, not registered) — new handlers + Gateway proxy
- Validation (SHACL): NO endpoint — add to ontology-service or Gateway
- Draft state: NO endpoint — planned `/ontologies/{id}/draft` per ADR
- Dashboard aggregate: NO endpoint — composite aggregator or Gateway handler
- Merge requests: NO endpoint — needs design (likely versioning-service)
- Deployments: partial (publisher snapshots) — adapt or add `/deployments`
- User preferences: NO endpoint — add to auth-service or Gateway

## Tasks

### Phase 1: ADR & Architecture Documentation

- [x] **Task 1.1: Update ADRs — tighten GraphQL scope to graph-only**
  Update three ADRs to reflect the new rule (GraphQL = only ontology graph navigation; everything else = REST):
  - `specs/adr/ADR-DES.API.protocol-stack-strategy.md`: remove "Чтение версии (commits, branches, tags)" from GraphQL column → REST. Add explicit note: "GraphQL-схема НЕ содержит не-графовых резолверов (версионирование, орг-модель, метаданные онтологии)".
  - `specs/adr/ADR-DES.API.graphql-sparql-split-strategy.md`: remove "Чтение версии" row and "Управление организацией" read row from GraphQL table → REST. Update §"Запрещено в GraphQL" to add: "Любые не-графовые Query-резолверы (groups, projects, members, commits, branches, tags) — migration to REST per ADR-DES.API.rest-graphql-mutation-boundary."
  - `specs/adr/ADR-DES.API.organization-rest-endpoints.md`: add explicit read rows to canonical table (GET groups, GET projects, GET members — already in routes.go, now formally documented as the canonical REST read contract).
  Files:
  - `specs/adr/ADR-DES.API.protocol-stack-strategy.md`
  - `specs/adr/ADR-DES.API.graphql-sparql-split-strategy.md`
  - `specs/adr/ADR-DES.API.organization-rest-endpoints.md`

  LOGGING: INFO on ADR update with changed sections. WARN if an ADR references the old GraphQL column without update.

- [x] **Task 1.2: Update GraphQL schema documentation**
  Update `src/services/api-gateway/docs/graphql-schema.md`:
  - Remove §2.1 Ontology metadata row (ontology(id) → REST)
  - Remove §2.5 Versioning (read-only) entirely
  - Remove §2.6 Organization (read-only) entirely
  - Remove §3 Mutation root (deprecated) entirely
  - Update §1 Boundary table: remove "Read commits and branches" and "Read org context" from GraphQL column → add to REST column
  - Update §4 Forbidden operations table: add `commits`, `branch`, `branches`, `groups`, `projects`, `members`, `ontology` commands → now forbidden in GraphQL

  Files:
  - `src/services/api-gateway/docs/graphql-schema.md`

  LOGGING: INFO on sections removed, with specific section numbers.

### Phase 2: Backend — Shrink GraphQL Schema

- [x] **Task 2.1: Remove non-graph resolvers from ontology-service QueryRoot**
  In `src/services/ontology-service/src/graphql/query.rs`:
  - Remove `ontology(id)` resolver (lines ~141-177) — metadata now served by REST `GET /ontologies/:id`
  - Remove `commits(ontologyId, ...)` resolver (lines ~465-508) — versioning reads via REST `/versioning/commits`
  - Remove `branch(ontologyId, branchId)` resolver (lines ~511-533) — REST `/versioning/branches/:id`
  - Remove `branches(ontologyId, ...)` resolver (lines ~536-563) — REST `/versioning/branches`
  - Remove `groups(q)` resolver (lines ~568-586) — REST `/groups`
  - Remove `projects(q, sortBy, sortDir, page, perPage)` resolver (lines ~589-618) — REST `/projects`
  - Remove `members(ontologyId)` resolver (lines ~621-643) — REST `/projects/:id/members`
  - Remove `branch_into_gql` helper (lines ~648-661) if only used by removed resolvers

  **Keep only graph resolvers:** class, classes, classTree, classAncestors, classDescendants, graphNeighborhood, autocompleteClasses, property, properties, individual, individuals (11 total).

  Also remove any associated types only used by removed resolvers.

  Files:
  - `src/services/ontology-service/src/graphql/query.rs`

  LOGGING: WARN on non-graph resolver invocation (grace period before removal). After removal, DEBUG on build_schema with resolver count.

- [x] **Task 2.2: Remove MutationRoot → EmptyMutation in ontology-service**
  In `src/services/ontology-service/src/graphql/schema.rs`:
  - Replace `MutationRoot` with `EmptyMutation` in `Schema::build`
  - In `mutation.rs`: remove all remaining mutation resolvers (updateDraft, updateMemberRole, removeMember) or delete the file entirely

  **Rust compile check:**
  ```bash
  cd src/services/ontology-service && cargo check
  ```

  Files:
  - `src/services/ontology-service/src/graphql/schema.rs`
  - `src/services/ontology-service/src/graphql/mutation.rs` (delete or annotate as deprecated for removal)
  - `src/services/ontology-service/src/graphql/mod.rs` (update module exports)

  LOGGING: INFO on EmptyMutation applied. WARN if any mutation resolver is still referenced at compile time.

### Phase 3: Backend — Gateway REST for Existing Service Endpoints (Org + Metrics)

- [x] **Task 3.1: Verify and document Organization REST API (groups, projects, members)**
  The org REST endpoints already exist in `routes.go` (lines 80–110) via `orgHandler` backed by auth-service gRPC. Before frontend can migrate from GraphQL to REST, we must verify these endpoints work end-to-end:
  - **Groups read:** `GET /api/v1/groups`, `GET /api/v1/groups/:id`, `GET /api/v1/groups/:id/subgroups`, `GET /api/v1/groups/:id/members`
  - **Groups write:** `POST /api/v1/groups`, `PUT /api/v1/groups/:id`, `DELETE /api/v1/groups/:id`
  - **Projects read:** `GET /api/v1/projects`, `GET /api/v1/projects/:id`, `GET /api/v1/projects/:id/members`, `GET /api/v1/projects/:id/visibility`, `GET /api/v1/projects/:id/policies`
  - **Projects write:** `POST /api/v1/projects`, `PUT /api/v1/projects/:id`, `DELETE /api/v1/projects/:id`, `POST /api/v1/projects/:id/members`, `PUT /api/v1/projects/:id/members/:userId`, `DELETE /api/v1/projects/:id/members/:userId`, `PUT /api/v1/projects/:id/visibility`, `POST /api/v1/projects/:id/policies`, `DELETE /api/v1/projects/:id/policies/:policyId`
  - **Missing endpoint (per ADR-DES.API.organization-rest-endpoints):** add `PUT /api/v1/projects/:id/move` — transfer Project to another Group (GitLab-aligned).
  - **Verify idempotency:** all write endpoints under `/projects/` require `Idempotency-Key` header → 400 without it.
  - **Verify RBAC:** only Owner can manage members/visibility/policies; Guest/Developer get 403.
  - **Verify audit:** each write endpoint emits structured audit log with `event`, `user_id`, `object_type`, `object_id`, `trace_id`, `timestamp`.
  - Run existing `auth_integration_test.go`; add smoke test for each read endpoint returning proper JSON shape matching the GraphQL response schema (so frontend can drop-in replace).

  Files:
  - `src/services/api-gateway/routes.go` (add `PUT /projects/:id/move`)
  - `src/services/api-gateway/handlers/` (orgHandler — verify existing handlers)
  - `src/services/api-gateway/auth_integration_test.go` (extend with read-endpoint smoke tests)

  LOGGING: INFO on org endpoint verification with endpoint path and response status. WARN on any endpoint returning non-200 for valid input. ERROR on RBAC bypass.

- [x] **Task 3.2: Add Gateway proxy routes for metrics-service REST endpoints**
  The metrics-service (Python FastAPI, port 8084) already exposes:
  - `GET /api/v1/metrics/ontologies` — list all ontology metrics
  - `GET /api/v1/metrics/ontologies/{ontology_id}` — specific ontology metrics

  Add proxy routes in `src/services/api-gateway/routes.go`:
  ```go
  metricsProxy := mustNewProxy(
      getEnv("METRICS_SERVICE_URL", "http://localhost:8084"),
      "metrics-service",
  )
  api.GET("/metrics/ontologies", gin.WrapH(metricsProxy))
  api.GET("/metrics/ontologies/:ontology_id", gin.WrapH(metricsProxy))
  ```

  Add `METRICS_SERVICE_URL` to docker-compose environment if not present.

  Files:
  - `src/services/api-gateway/routes.go`

  LOGGING: INFO on proxy initialization with upstream URL. WARN on proxy creation failure.

### Phase 4: Backend — New REST Endpoints

- [x] **Task 4.1: Add commenting-service HTTP handlers + Gateway proxy**
  **Backend (commenting-service):**
  - Register HTTP routes in `src/services/commenting-service/main.go`:
    - `GET /api/v1/ontologies/:id/comments` — list comments for ontology entity
    - `POST /api/v1/ontologies/:id/comments` — create comment
  - Implement handlers in `handlers.go` using existing store (store.go) — or delegate to gRPC handlers if available
  - Add `X-User-ID` auth propagation

  **Gateway (routes.go):**
  ```go
  commentingProxy := mustNewProxy(
      getEnv("COMMENTING_SERVICE_URL", "http://localhost:8084"),
      "commenting-service",
  )
  api.GET("/ontologies/:id/comments", gin.WrapH(commentingProxy))
  api.POST("/ontologies/:id/comments", gin.WrapH(commentingProxy))
  ```

  Files:
  - `src/services/commenting-service/main.go`
  - `src/services/commenting-service/handlers.go`
  - `src/services/commenting-service/store.go` (reference)
  - `src/services/api-gateway/routes.go`

  LOGGING: INFO on comment CRUD with ontology_id, user_id, and result count. ERROR on store failure with trace_id.

- [x] **Task 4.2: Add SHACL validation REST endpoint + Gateway proxy**
  The SHACL validation currently uses a GraphQL mutation (`RUN_VALIDATION_MUTATION`). There is no SHACL validation backend yet — the mutation returns stub data.

  **Backend (ontology-service or new route):**
  - Add HTTP handler: `POST /api/v1/ontologies/:id/validate`
  - Return SHACL validation report (status, violations[], validatedAt)
  - Stub implementation acceptable — returns `{ status: "ok", violations: [], validatedAt: "..." }` per existing mock behavior
  - Log the request with ontology_id and user_id

  **Gateway (routes.go):**
  ```go
  api.POST("/ontologies/:id/validate", gin.WrapH(ontologyProxy))
  ```

  Files:
  - `src/services/api-gateway/routes.go`

  LOGGING: INFO on validation request with ontology_id. WARN on stub response (until real SHACL engine is wired). ERROR on validation engine failure.

- [x] **Task 4.3: Add draft-state REST endpoint + Gateway proxy**
  Per ADR-DES.API.rest-graphql-mutation-boundary.md, draft-state coordination should be `PUT /api/v1/ontologies/{id}/draft`.

  **Backend (ontology-service proxy or gRPC):**
  - Add HTTP handler: `PUT /api/v1/ontologies/:id/draft`
  - Accepts `{ changes: DraftInput }` — same payload as current `updateDraft` mutation
  - Returns `{ success: boolean, timestamp: string }`
  - Requires `Idempotency-Key` header

  **Gateway (routes.go):**
  ```go
  api.PUT("/ontologies/:id/draft", gin.WrapH(ontologyProxy))
  ```
  Add `/ontologies/:id/draft` to idempotency middleware CriticalPaths.

  Files:
  - `src/services/api-gateway/routes.go`

  LOGGING: INFO on draft update with ontology_id and change count. ERROR on conflicting concurrent edits.

### Phase 5: Frontend — Clean queries.ts & Add REST Clients

- [x] **Task 5.1: Strip queries.ts to graph-only queries**
  In `src/services/frontend/src/apollo/queries.ts`:

  **KEEP (graph navigation — 15 exports):**
  - `CLASS_SUMMARY_FRAGMENT`, `CLASS_FRAGMENT`, `PROPERTY_FRAGMENT`, `INDIVIDUAL_FRAGMENT`
  - `GET_CLASS_QUERY`, `LIST_CLASSES_QUERY`, `CLASS_TREE_QUERY`, `CLASS_ANCESTORS_QUERY`, `CLASS_DESCENDANTS_QUERY`
  - `GRAPH_NEIGHBORHOOD_QUERY`, `AUTOCOMPLETE_CLASSES_QUERY`
  - `GET_PROPERTY_QUERY`, `LIST_PROPERTIES_QUERY`
  - `GET_INDIVIDUAL_QUERY`, `LIST_INDIVIDUALS_QUERY`

  **REMOVE (non-graph — ~26 exports):**
  - `ONTOLOGY_QUERY`, `VERSION_CONTEXT_QUERY`, `NAVIGATION_STATE_QUERY`
  - `GET_COMMIT_HISTORY_QUERY`, `COMMIT_SUMMARY_FRAGMENT`
  - `GET_BRANCHES_QUERY`, `BRANCH_FRAGMENT`, `GET_BRANCH_QUERY`
  - `GET_TAGS_QUERY`, `COMPARE_REVISIONS_QUERY`
  - `LIST_PROJECTS_QUERY`, `LIST_GROUPS_QUERY`, `LIST_MEMBERS_QUERY`
  - `DASHBOARD_QUERY`, `ONTOLOGY_METRICS_QUERY`
  - `LIST_DEPLOYMENTS_QUERY`, `LIST_MERGE_REQUESTS_QUERY`
  - `GET_ENTITY_COMMENTS_QUERY`, `COMMENT_FRAGMENT`, `GET_COMMENT_FEED_QUERY`
  - `CREATE_COMMENT_MUTATION`, `UPDATE_DRAFT_MUTATION`
  - `UPDATE_MEMBER_ROLE_MUTATION`, `REMOVE_MEMBER_MUTATION`
  - `RUN_VALIDATION_MUTATION`

  Update `src/services/frontend/biome.json` to remove the `queries.ts` override if no longer needed.

  Files:
  - `src/services/frontend/src/apollo/queries.ts`
  - `src/services/frontend/biome.json`

  LOGGING: No runtime logging — build-time change. Verify with `vitest` that no import breaks compile.

- [x] **Task 5.2: Create axios REST client modules for migrated domains**
  Create REST client files following the pattern in `src/services/frontend/src/api/ontology.ts` (axios instance + JWT interceptor + `Idempotency-Key` + structured logging):

  - `src/services/frontend/src/api/versioning.ts` — commits list, branch list, branch detail, tags, compareRevisions
  - `src/services/frontend/src/api/org.ts` — groups list, projects list, members list
  - `src/services/frontend/src/api/comments.ts` — get comments, create comment
  - `src/services/frontend/src/api/metrics.ts` — get ontology metrics
  - `src/services/frontend/src/api/dashboard.ts` — dashboard aggregate (composite endpoint or local aggregation)
  - `src/services/frontend/src/api/deployments.ts` — list deployments (adapter over publisher snapshots)
  - `src/services/frontend/src/api/merge-requests.ts` — list merge requests
  - `src/services/frontend/src/api/validation.ts` — run SHACL validation
  - `src/services/frontend/src/api/preferences.ts` — get/update user preferences (or use localStorage fallback)

  Each module:
  - Uses axios with `baseURL: "/api/v1"`
  - Adds JWT token interceptor (copy from `api/ontology.ts`)
  - Adds `Idempotency-Key` header for write operations
  - Exports typed request/response interfaces
  - Logs: INFO on request with params, ERROR on failure with server message

  For endpoints that don't yet exist on backend (merge-requests, dashboard, preferences): implement with mock REST responses (return stub data matching the current GraphQL mock shapes) so pages work unchanged after migration.

  Files:
  - `src/services/frontend/src/api/` (9 new files)

  LOGGING: INFO with `[module].[operation].[phase]` format. ERROR with server error payload.

### Phase 6: Frontend — Migrate Pages to REST

- [x] **Task 6.1: Migrate VersioningPage to REST**
  In `src/services/frontend/src/pages/VersioningPage.vue`:
  - Remove imports: `GET_COMMIT_HISTORY_QUERY`, `GET_BRANCHES_QUERY`, `GET_TAGS_QUERY`, `COMPARE_REVISIONS_QUERY` from `@/apollo/queries`
  - KEEP: `GRAPH_NEIGHBORHOOD_QUERY` (graph navigation — stays in GraphQL)
  - Replace `useQuery(GET_COMMIT_HISTORY_QUERY, ...)` with axios call to `GET /api/v1/versioning/commits`
  - Replace `useQuery(GET_BRANCHES_QUERY, ...)` with axios call to `GET /api/v1/versioning/branches`
  - Replace `useQuery(GET_TAGS_QUERY, ...)` with axios call to `GET /api/v1/versioning/tags` (add route if missing)
  - Replace `useQuery(COMPARE_REVISIONS_QUERY, ...)` with axios call to `GET /api/v1/versioning/commits/:id/delta`
  - Add loading/error/empty states per existing M4 UI pattern
  - Use typed interfaces from `api/versioning.ts`

  Files:
  - `src/services/frontend/src/pages/VersioningPage.vue`

  LOGGING: INFO on page mount with ontology_id and branch. ERROR on API failure with endpoint URL and status.

- [x] **Task 6.2: Migrate ProjectsPage, GroupsPage, MembersPage to REST**
  **ProjectsPage:**
  - Remove `LIST_PROJECTS_QUERY` import from `@/apollo/queries`
  - Replace `useQuery(LIST_PROJECTS_QUERY, ...)` with `GET /api/v1/projects` via `api/org.ts`
  - Preserve search, sort, pagination parameters
  - Use typed interfaces from `api/org.ts`

  **GroupsPage:**
  - Remove `LIST_GROUPS_QUERY` import from `@/apollo/queries`
  - Replace `useQuery(LIST_GROUPS_QUERY, ...)` with `GET /api/v1/groups` via `api/org.ts`

  **MembersPage:**
  - Remove `LIST_MEMBERS_QUERY` import from `@/apollo/queries`
  - Replace `useQuery(LIST_MEMBERS_QUERY, ...)` with `GET /api/v1/projects/:id/members` via `api/org.ts`
  - Note: route param `id` is the project ID (per ADR-DES.API.organization-rest-endpoints 1:1 Project↔Ontology)
  - Any member edit/remove actions already use REST (routes.go:104-106) — verify they work

  Files:
  - `src/services/frontend/src/pages/ProjectsPage.vue`
  - `src/services/frontend/src/pages/GroupsPage.vue`
  - `src/services/frontend/src/pages/MembersPage.vue`

  LOGGING: INFO on page mount with entity type and query params. ERROR on API failure.

- [x] **Task 6.3: Migrate OntologyWorkspace ontology metadata to REST**
  In `src/services/frontend/src/pages/OntologyWorkspace.vue`:
  - Remove `ONTOLOGY_QUERY` import from `../apollo/queries`
  - KEEP: `CLASS_TREE_QUERY`, `LIST_INDIVIDUALS_QUERY` (graph navigation — stays in GraphQL)
  - Replace `useQuery(ONTOLOGY_QUERY, { id })` with `GET /api/v1/ontologies/{id}` via axios
  - Also update `src/services/frontend/src/composables/useVersionContext.ts`:
    - Remove `VERSION_CONTEXT_QUERY` import and `useQuery` call
    - Replace with `GET /api/v1/ontologies/{id}` returning `{ branch, commit, dirty }`
  - Remove `UPDATE_DRAFT_MUTATION` import from `useDraftState.ts` composable:
    - Replace with `PUT /api/v1/ontologies/{id}/draft` via axios

  Files:
  - `src/services/frontend/src/pages/OntologyWorkspace.vue`
  - `src/services/frontend/src/composables/useVersionContext.ts`
  - `src/services/frontend/src/composables/useDraftState.ts`

  LOGGING: INFO on workspace mount with ontology_id. INFO on draft state changes. ERROR on metadata fetch failure.

- [x] **Task 6.4: Migrate CommentsPage, DashboardPage, MetricsPage, DeploymentsPage, MergeRequestsPage, ShaclPage, ValidationPage to REST**
  **CommentsPage:**
  - Remove `CREATE_COMMENT_MUTATION`, `GET_COMMENT_FEED_QUERY` from `@/apollo/queries`
  - Replace with `GET/POST /api/v1/ontologies/{id}/comments` via `api/comments.ts`
  - Keep `Comments.vue` organism unchanged (it receives props)

  **DashboardPage:**
  - Remove `DASHBOARD_QUERY` from `@/apollo/queries`
  - Replace with `GET /api/v1/dashboard` via `api/dashboard.ts`
  - Dashboard aggregate endpoint: if backend doesn't exist yet, create a Gateway handler that aggregates from multiple services, or return mock data matching current GraphQL shape

  **MetricsPage:**
  - Remove `ONTOLOGY_METRICS_QUERY` from `@/apollo/queries`
  - Replace with `GET /api/v1/metrics/ontologies/{id}` via `api/metrics.ts`

  **DeploymentsPage:**
  - Remove `LIST_DEPLOYMENTS_QUERY` from `@/apollo/queries`
  - Replace with `GET /api/v1/deployments` via `api/deployments.ts`

  **MergeRequestsPage:**
  - Remove `LIST_MERGE_REQUESTS_QUERY` from `@/apollo/queries`
  - Replace with `GET /api/v1/merge-requests` via `api/merge-requests.ts`

  **ShaclPage & ValidationPage:**
  - Remove `RUN_VALIDATION_MUTATION` from `@/apollo/queries`
  - Replace with `POST /api/v1/ontologies/{id}/validate` via `api/validation.ts`

  Files:
  - `src/services/frontend/src/pages/CommentsPage.vue`
  - `src/services/frontend/src/pages/DashboardPage.vue`
  - `src/services/frontend/src/pages/MetricsPage.vue`
  - `src/services/frontend/src/pages/DeploymentsPage.vue`
  - `src/services/frontend/src/pages/MergeRequestsPage.vue`
  - `src/services/frontend/src/pages/ShaclPage.vue`
  - `src/services/frontend/src/pages/ValidationPage.vue`

  LOGGING: INFO on page mount with relevant entity ID. ERROR on API failure with endpoint and status.

  **Note for DashboardPage:** `App.vue` also imports `DASHBOARD_QUERY` from `@/apollo/queries` — update to REST import as well. Check `src/services/frontend/src/App.vue:110`.

### Phase 7: Tests & Traceability

- [x] **Task 7.1: Update frontend test fixtures and mock providers**
  Update test infrastructure to match REST migration:
  - `src/services/frontend/src/__tests__/setup/mock-providers.ts`:
    - Add mock axios instances for new REST client modules
    - Remove Apollo mock resolvers for non-graph queries that are being deleted
    - Keep Apollo mock resolvers for graph-only queries (class, property, individual)
  - `src/services/frontend/src/apollo/mock-link.ts`, `src/services/frontend/src/apollo/mock-data.ts`:
    - Remove mock resolvers for deleted queries (commits, branches, tags, compareRevisions, groups, projects, members, dashboard, metrics, deployments, merge-requests, comments, validation, draft)
    - Keep mock resolvers for graph navigation queries
  - `src/services/frontend/__tests__/CreateClassDialog.spec.ts`, `CreatePropertyDialog.spec.ts`, `CreateIndividualDialog.spec.ts`:
    - Already migrated to REST (`vi.mock("@/api/ontology", ...)`) ✅ — no change needed
  - Add new spec files for axios REST clients: `src/api/__tests__/` or inline in `__tests__/` per existing pattern

  Files:
  - `src/services/frontend/src/__tests__/setup/mock-providers.ts`
  - `src/services/frontend/src/apollo/mock-link.ts`
  - `src/services/frontend/src/apollo/mock-data.ts`
  - New test files as needed

  > BDD naming: `'should <expected> when <condition>'`
  > Anti-patterns: see `.ai-factory/rules/test-quality.md`

  LOGGING: No runtime logging (test infrastructure). Verify with `vitest run`.

- [x] **Task 7.2: Add backend integration tests for new REST endpoints**
  Add or update integration tests in `src/services/api-gateway/`:
  - Metrics proxy: test `GET /api/v1/metrics/ontologies` returns 200 + JSON
  - Comments: test `POST /api/v1/ontologies/:id/comments` requires auth, returns 201
  - Validation: test `POST /api/v1/ontologies/:id/validate` returns validation report
  - Draft: test `PUT /api/v1/ontologies/:id/draft` with and without `Idempotency-Key`

  Also add negative tests proving:
  - GraphQL introspection no longer exposes non-graph query fields (`commits`, `groups`, `projects`, `members`, `ontology`)
  - GraphQL introspection no longer exposes `Mutation` type
  - GraphQL `ontology(id)` query returns error or schema field not found

  Follow existing test pattern from `rest_entity_crud_integration_test.go`, `proxy_integration_test.go`, `auth_integration_test.go`.

  Files:
  - New/modified test files in `src/services/api-gateway/`

  > BDD naming: `[Condition]_[Action]_[ExpectedResult]`
  > Anti-patterns: see `.ai-factory/rules/test-quality.md`

  LOGGING: INFO on test setup with service URLs. ERROR on unexpected responses with full request/response.

- [x] **Task 7.3: Update traceability.ttl**
  - Add `vdo:TestSuite` entry for each new test file created in Tasks 7.1-7.2
  - Add `vdo:validates` triples for each `// Validates: REQ-...` annotation in new test code
  - Remove stale triples for deleted test files (if any)
  - Verify with `aif-test-quality` check

  Files:
  - `.ai-factory/traceability/traceability.ttl`

  LOGGING: INFO on each added/removed triple with reason.

### Phase 8: Cleanup & Documentation

- [x] **Task 8.1: Remove dead code, update Antora docs**
  **Dead code removal:**
  - Remove `UPDATE_MEMBER_ROLE_MUTATION` and `REMOVE_MEMBER_MUTATION` from `queries.ts` if they survive Task 5.1 (these are already covered in that task)
  - Verify no remaining imports reference deleted exports — run `vitest run` and `vue-tsc --noEmit` to catch compile errors
  - Check `src/services/frontend/src/composables/useDraftState.ts` no longer imports from `@/apollo/queries`

  **Antora documentation:**
  - Update `src/docs/antora/developer-guide/modules/ROOT/pages/` with revised architecture boundary:
    - GraphQL = ontology graph navigation only (class/property/individual + tree/neighborhood/autocomplete)
    - REST = everything else (versioning, org, comments, metrics, dashboard, deployments, MR, validation, draft, preferences)
  - Update any integration guide references to the GraphQL endpoint

  **Commit check:**
  - Run full `go test ./...` in `src/services/api-gateway/`
  - Run `cargo check` in `src/services/ontology-service/`
  - Run `vitest run` in `src/services/frontend/`
  - Verify `docker-compose up` smoke test passes

  Files:
  - Various (dead code identification)
  - `src/docs/antora/developer-guide/` (Antora AsciiDoc pages)

  LOGGING: INFO on dead code removal count. WARN on any remaining references to deleted exports.

## Commit Plan

- **Commit 1** (after Phase 1-2): "docs: tighten GraphQL scope to ontology graph navigation only in ADR and schema docs"
- **Commit 2** (after Phase 3-4): "feat: verify org REST API, add metrics proxy, and implement comments/validation/draft endpoints"
- **Commit 3** (after Phase 5): "refactor: strip queries.ts to graph-only and add REST client modules"
- **Commit 4** (after Phase 6): "refactor: migrate frontend pages from Apollo GraphQL to axios REST"
- **Commit 5** (after Phase 7-8): "test: add REST integration tests, update traceability, and clean up dead code"

<!-- Commit message candidates (conventional commits):
  docs: restrict GraphQL schema to ontology navigation only
  feat: add REST proxy routes for metrics, comments, and validation services
  refactor: replace non-graph GraphQL queries with typed REST clients
  refactor: migrate 11 frontend pages from Apollo to axios REST calls
  test: add REST endpoint integration tests and verify GraphQL schema cleanup
-->

## Acceptance Criteria
- [ ] All tests pass: `go test ./...` (api-gateway), `cargo test` (ontology-service), `vitest` (frontend)
- [ ] Test Quality Score (TQS) ≥ bronze (6.0)
- [ ] No B1–B7 anti-patterns (see .ai-factory/rules/test-quality.md)
- [ ] Traceability annotations present (`// Validates: REQ-...`)
- [ ] GraphQL introspection returns only 11 query resolvers (class, classes, classTree, classAncestors, classDescendants, graphNeighborhood, autocompleteClasses, property, properties, individual, individuals) and no Mutation type
- [ ] All org REST read/write endpoints return correct status codes and JSON shapes matching GraphQL response schemas (groups, projects, members — all CRUD operations, visibility, policies, idempotency)
- [ ] `PUT /api/v1/projects/{id}/move` transfers project between groups and emits audit event
- [ ] Versioning reads return 200 from REST: `/versioning/commits`, `/versioning/branches`, `/versioning/commits/:id/delta`
- [ ] Metrics reads return 200 from REST: `/metrics/ontologies`, `/metrics/ontologies/{id}`
- [ ] Comments POST returns 201 with valid auth
- [ ] Validation POST returns 200 with report structure
- [ ] Draft PUT requires `Idempotency-Key` (400 without it)
- [ ] No `useMutation` or non-graph `useQuery` from `@vue/apollo-composable` in any frontend page outside ontology workspace graph operations
- [ ] Documentation updated (graphql-schema.md, Antora developer guide)
