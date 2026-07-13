# Implementation Plan: M1 — Ontology Core Engine

Branch: feature/ontology-core-engine
Created: 2026-07-13

## Settings
- Testing: yes (TDD — E2E first, JIT unit tests per task, integration tests for external interfaces)
- Logging: verbose (DEBUG for flow, INFO for events, ERROR for failures)
- Docs: yes (Antora AsciiDoc under src/docs/antora/)

## Roadmap Linkage
Milestone: "M1: Ontology Core Engine"
Rationale: Core ontology editing engine — class/property/individual CRUD, Git-like versioning, import/export, basic REST API and SPARQL/CYPHER query endpoints.

## Specs References

This plan is derived from specifications in the `specs/` submodule. Every task references its source requirements (REQ), architecture decisions (ADR), user stories (US), and use cases (UC).

### Architecture Decisions (ADRs)

| ADR | Status | Relevance to M1 |
|-----|--------|-----------------|
| `ADR-IMPL.STACK.ontology-rust-strategy` | Accepted | Rust + tokio + async-graphql for ontology-service; cargo workspace with versioning-service |
| `ADR-IMPL.STACK.version-control-rust-strategy` | Accepted | Rust + tokio for versioning-service; semantic deltas, materialized state in Neo4j |
| `ADR-DES.API.protocol-stack-strategy` | Accepted | gRPC for internal calls, REST for external API, GraphQL for frontend, API Gateway as single facade |
| `ADR-DES.DATA.storage-stack-strategy` | Accepted | Neo4j for graph store, PostgreSQL + JSONB for version store (deltas, not full snapshots) |
| `ADR-IMPL.STACK.frontend-vue-strategy` | Accepted | Vue 3 + Apollo Client for frontend GraphQL integration |
| `ADR-DES.API.graphql-sparql-split-strategy` | Accepted | GraphQL for navigation queries, SPARQL for analytical queries |
| `ADR-DES.API.unified-root-endpoint-adoption` | Accepted | REST API at `/api/v1/` prefix |
| `ADR-DES.API.sparql-query-language-strategy` | Accepted | SPARQL 1.1 for RDF/OWL graph queries |
| `ADR-DES.API.cypher-query-language-adoption` | Accepted | CYPHER as second query language alongside SPARQL |
| `ADR-DES.API.write-idempotency-strategy` | Accepted | Idempotent write operations |
| `ADR-DES.SECURITY.gitlab-like-organization-model` | Accepted | Multi-tenant groups, ontology projects, role-based access |
| `ADR-DES.UI.ontology-mental-model-strategy` | Accepted | Tree + graph + property panel UI layout |
| `ADR-DES.UI.version-context-visibility-strategy` | Accepted | Version context (branch, commit, dirty state) shown in toolbar |
| `ADR-IMPL.PROCESS.repository-layout-strategy` | Accepted | Monorepo layout with services under `src/services/` |
| `ADR-IMPL.PROCESS.c4-notation-adoption` | Accepted | C4 diagrams for architecture documentation |
| `ADR-DES.UI.navigation-state-strategy` | Accepted | Navigation state management with browser history |
| `ADR-IMPL.STACK.port-mapping-strategy` | Accepted | Port allocation: gateway 8080, ontology 8082, versioning 8083 |
| `ADR-DES.INFRA.monolith-vs-microservices` | Accepted | Microservices architecture (14+ services) |

### Requirements (REQ)

| Requirement | Priority | M1 Domain |
|-------------|----------|-----------|
| `REQ-FUN.DATA.versioning` | NFR | F3 — Git-like versioning model |
| `REQ-FUN.DATA.storage-stack` | FUN | Database selection (Neo4j, PostgreSQL, Redis) |
| `REQ-FUN.DATA.major-version-policy` | FUN | Versioning strategy |
| `REQ-FUN.API.graphql-sparql` | P0 | F1, F10 — Navigation and analytical API split |
| `REQ-FUN.API.cypher-query-language` | P1 | F10 — CYPHER query support |
| `REQ-FUN.API.protocol-stack` | FUN | Protocol stack (gRPC, REST, GraphQL) |
| `REQ-FUN.API.rest-versioning` | FUN | API versioning at `/api/v1/` |
| `REQ-FUN.API.unified-root-endpoint` | FUN | Unified API root |
| `REQ-FUN.API.write-idempotency` | FUN | Idempotent writes |
| `REQ-FUN.API.integration` | FUN | F6 — REST API for integration |
| `REQ-FUN.DATA.major-version-migration` | FUN | Data migration strategy |
| `REQ-FUN.DATA.migration-rollback` | FUN | Rollback capability |
| `REQ-NFR.PERF.performance` | P0 | p95 < 120ms, p99 < 250ms |
| `REQ-NFR.PERF.canonical-workload-profile` | NFR | Workload profiles for performance testing |
| `REQ-NFR.SECURITY.authorization-regression-gates` | P0 | BOLA/BFLA enforcement |
| `REQ-NFR.SECURITY.bola-bfla-negative-tests` | P0 | Negative authorization tests |
| `REQ-NFR.UI.wcag-mvp-scenarios` | P0 | WCAG accessibility for MVP scenarios |
| `REQ-USR.UI.tbox-editor` | P0 | F2 — TBox editor UI requirements |
| `REQ-USR.UI.abox-editor` | P1 | F4 — ABox editor UI requirements |
| `REQ-USR.UI.graph-navigation` | P0 | F1 — Graph navigation UI |
| `REQ-USR.UI.import-export` | P1 | F8 — Import/export UI |
| `REQ-USR.UI.sparql-gui-search` | P1 | F10 — SPARQL GUI search |
| `REQ-CON.STACK.backend-service-stack` | CON | Backend technology stack |
| `REQ-CON.STACK.frontend-stack` | CON | Frontend technology stack |
| `REQ-CON.INFRA.containerization` | CON | Docker containerization |

### Use Cases (UC)

| Use Case | M1 Domain |
|----------|-----------|
| `UC-editor.classes.manage-class-lifecycle` | F2.1 |
| `UC-editor.properties.manage-property-lifecycle` | F2.2, F2.3 |
| `UC-editor.properties.manage-ontology-annotations` | F2.4 |
| `UC-browse.tree.view-ontology-tree-and-graph` | F1.1, F1.2, F1.3 |
| `UC-browse.search.search-ontology-elements` | F1.6 |
| `UC-browse.graph.view-ontology-graph-with-pagination` | F1.2 |
| `UC-git.commits.manage-commit-history` | F3.1, F3.4, F3.6 |
| `UC-git.branches.manage-branch-workflow` | F3.2, F3.3 |
| `UC-git.commits.compare-ontology-versions` | F3.5 |
| `UC-abox.individuals.manage-individual-lifecycle` | F4.1–F4.4 |
| `UC-abox.individuals.navigate-linked-individuals` | F4.11 |
| `UC-io.import.import-and-export-ontology-data` | F8.1, F8.3 |
| `UC-io.import.import-and-export-ontology-xlsx` | F8.4 (post-MVP) |
| `UC-io.export.export-ontology-to-docx` | F8.2 (post-MVP) |
| `UC-io.publish.publish-ontology-snapshot` | F11 (M2) |
| `UC-editor.classes.validate-ontology-with-shacl` | F2.7 (post-MVP) |
| `UC-api.integration.integrate-through-platform-apis` | F6.1 |
| `UC-api.auth.issue-and-validate-jwt-tokens` | F6.2 |
| `UC-api.webhooks.manage-subscriptions-and-delivery` | F6.3 |
| `UC-api.docs.view-openapi-specification` | F6.4 |
| `UC-browse.search.execute-sparql-query-through-gui` | F10.1 (M2 GUI) |

### Existing E2E Scenarios

| E2E Spec | Scope | M1 Coverage |
|----------|-------|-------------|
| `E2E-editor.workflow.full-cycle` | Full ontology lifecycle: classes → properties → individuals → commit | ✅ Covers F1, F2, F3, F4 |
| `E2E-api.integration.rest` | REST API: create ontology, webhooks, Turtle export | ✅ Covers F6, F8 |

---

## User Story Gap Analysis

### Existing user stories covering M1

| User Story | Domain | Priority | M1 Area | Status |
|------------|--------|----------|---------|--------|
| `US-editor.classes.create-parents` | F2 | P0 | Class CRUD with multiple parents | ✅ Existing |
| `US-editor.classes.edit` | F2 | P0 | Edit class label, comment | ✅ Existing |
| `US-editor.classes.delete` | F2 | P0 | Delete class with cascade check | ✅ Existing |
| `US-editor.properties.create-object` | F2 | P0 | Create ObjectProperty | ✅ Existing |
| `US-editor.properties.create-datatype` | F2 | P0 | Create DatatypeProperty | ✅ Existing |
| `US-editor.annotations.add-label` | F2 | P1 | Add rdfs:label annotation | ✅ Existing |
| `US-browse.tree.graph-view` | F1 | P0 | Class tree and graph view | ✅ Existing |
| `US-browse.search.fulltext` | F1 | P0 | Fulltext search | ✅ Existing |
| `US-browse.search.parametric` | F1 | P0 | Parametric search | ✅ Existing |
| `US-browse.graph.paginated` | F1 | P0 | Paginated graph view | ✅ Existing |
| `US-git.commits.create` | F3 | P1 | Create commit with message | ✅ Existing |
| `US-git.commits.rollback` | F3 | P1 | Rollback to previous commit | ✅ Existing |
| `US-git.commits.compare` | F3 | P2 | Compare ontology versions | ✅ Existing |
| `US-git.branches.create-merge` | F3 | P2 | Create and merge branches | ✅ Existing |
| `US-abox.individuals.create` | F4 | P1 | Create individual instance | ✅ Existing |
| `US-abox.individuals.navigate-link` | F4 | P1 | Navigate linked individuals | ✅ Existing |
| `US-io.import.turtle` | F8 | P0 | Import Turtle file | ✅ Existing |
| `US-io.export.canonical-turtle` | F8 | P0 | Export canonical Turtle | ✅ Existing |
| `US-api.classes.create-rest` | F6 | P1 | Create class via REST API | ✅ Existing |
| `US-api.auth.jwt` | F6 | P0 | JWT authentication for API | ✅ Existing |
| `US-api.docs.openapi` | F6 | P0 | OpenAPI documentation | ✅ Existing |
| `US-api.webhooks.manage` | F6 | P1 | Manage webhook subscriptions | ✅ Existing |
| `US-admin.access.assign-role` | Auth | P0 | Assign user roles | ✅ Existing |

### Missing user stories for M1 (gaps)

The following M1 functional areas lack corresponding user stories. These must be created before or during implementation:

| ID (proposed) | Domain | Priority | Description | Gap rationale |
|---------------|--------|----------|-------------|---------------|
| **`US-editor.properties.edit-delete`** | F2 | P0 | Edit and delete ObjectProperty and DatatypeProperty | Only `create` US exists; no edit/delete flow defined |
| **`US-editor.classes.hierarchy-drag-drop`** | F2 | P1 | Reorder class hierarchy via drag-n-drop | MVP specifies drag-n-drop tree (F2.5), but no US covers it |
| **`US-browse.individuals.list-by-class`** | F1 | P1 | View paginated list of individuals for a class | F1 MVP has "view individual list for class", but no US |
| **`US-git.branches.create-switch`** | F3 | P1 | Create and switch between branches (without merge) | `US-git.branches.create-merge` is P2 and covers merge; MVP needs basic branch creation/switch |
| **`US-abox.individuals.edit-property-values`** | F4 | P1 | Set and modify property values on individuals | `US-abox.individuals.create` covers creation but not property value editing |
| **`US-api.ontologies.read-rest`** | F6 | P1 | List ontologies and read classes/properties via REST API | Only class-create exists; no US for GET operations |
| **`US-api.sparql.execute`** | F10 | P1 | Execute SPARQL query through API endpoint | No US for SPARQL query execution |
| **`US-api.cypher.execute`** | F10 | P1 | Execute CYPHER query through API endpoint | No US for CYPHER query execution |
| **`US-browse.individuals.filter-by-property`** | F4 | P1 | Filter individual list by property values | F4.6 requires property-based filtering |

### Missing E2E scenarios for M1

| ID (proposed) | Domain | Scope | Gap rationale |
|---------------|--------|-------|---------------|
| **`E2E-versioning.branches.switch-rollback`** | F3 | Create branch → switch → modify → rollback → verify state restored | No E2E spec for versioning operations |
| **`E2E-query.sparql.execute`** | F10 | SPARQL query → verify results → CYPHER query → verify unified format | No E2E spec for query endpoints |

### Out-of-MVP-scope User Stories (excluded from M1)

The following existing user stories are part of full F2/F3/F8 specs but are explicitly **not in MVP scope** (vision.md §2.5). They should be moved to M2 planning:

- `US-editor.classes.validate-shacl` (P1) — SHACL validation, not in MVP F2
- `US-editor.rules.create-visual` (P1) — Data quality rules, not in MVP F2
- `US-editor.rules.deduplicate` (P1) — Deduplication rules, not in MVP F2
- `US-io.xlsx.import-export` (P1) — XLSX import/export, not in MVP F8
- `US-io.export.docx` (P2) — DOCX export, not in MVP F8
- `US-io.publish.snapshot` (P1) — Ontology publishing, M2 domain
- `US-browse.public.view` (P1) — Public browsing, M2 domain
- `US-team.reviews.approve` (P1) — Merge request reviews, M2 domain
- `US-team.comments.*` (P1) — Comments & discussions, M2 domain (F5)

---

## Traceability Impact Analysis

### Affected artifacts (from `.ai-factory/traceability/traceability.ttl`)

These existing artifacts MUST be updated when implementation begins — perform change impact analysis before modifying any file:

| Artifact | Type | File | Impact |
|----------|------|------|--------|
| `base:service/ontology-service` | vdo:Service | src/services/ontology-service/ | Full rewrite from stub — add Neo4j CRUD, RDF, GraphQL |
| `base:service/versioning-service` | vdo:Service | src/services/versioning-service/ | Full rewrite from stub — add commit/branch/delta model |
| `base:service/api-gateway` | vdo:Service | src/services/api-gateway/ | Major expansion — add proxy routing, SPARQL/CYPHER endpoints |
| `base:service/frontend` | vdo:Service | src/services/frontend/ | Wire Apollo to real backend, replace mocks |
| `base:service/auth-service` | vdo:Service | src/services/auth-service/ | Verify auth middleware integration |
| `base:service/publisher-service` | vdo:Service | src/services/publisher-service/ | Dependency — may need CRUD contract alignment |
| `base:context/PLAT-LOCAL-002` | vdo:ContextAnnotation | src/Makefile | Update build targets for new modules |

### New artifacts to declare in traceability.ttl

| Artifact | Type | File |
|----------|------|------|
| `base:service/shared-workspace` | vdo:Service | src/services/shared/ |

## User Story Coverage Matrix (M1)

```
Phase 0: E2E Tests
├── Task 0.2 → E2E-editor.workflow.full-cycle, E2E-api.integration.rest
│              US covered: editor.classes.*, editor.properties.*, abox.individuals.*,
│              git.commits.*, browse.tree.*, io.import.*, io.export.*
├── Task 0.3 → E2E-versioning.branches.switch-rollback [NEW]
│              US covered: git.branches.create-merge, git.commits.rollback
└── Task 0.4 → E2E-query.sparql.execute [NEW]
               US covered: api.sparql.execute [NEW], api.cypher.execute [NEW]

Phase 1-9: Implementation
├── Task 2.1 → US-editor.classes.create-parents, US-editor.classes.edit,
│              US-editor.classes.delete, REQ-USR.UI.tbox-editor
├── Task 2.2 → US-editor.properties.create-object, US-editor.properties.create-datatype,
│              US-editor.properties.edit-delete [NEW], US-editor.annotations.add-label
├── Task 2.3 → US-browse.tree.graph-view, US-browse.graph.paginated
├── Task 3.1 → US-git.commits.create, REQ-FUN.DATA.versioning
├── Task 3.2 → US-git.branches.create-switch [NEW], US-git.branches.create-merge
├── Task 3.3 → US-git.commits.rollback, US-git.commits.compare
├── Task 4.1 → US-abox.individuals.create, US-abox.individuals.navigate-link,
│              US-browse.individuals.list-by-class [NEW],
│              US-abox.individuals.edit-property-values [NEW]
├── Task 5.1 → US-io.export.canonical-turtle
├── Task 5.2 → US-io.import.turtle
├── Task 6.1 → US-api.auth.jwt, ADR-DES.API.protocol-stack-strategy
├── Task 6.2 → US-api.sparql.execute [NEW], US-api.cypher.execute [NEW],
│              REQ-FUN.API.graphql-sparql, REQ-FUN.API.cypher-query-language
├── Task 6.3 → US-api.classes.create-rest, US-api.ontologies.read-rest [NEW],
│              US-api.docs.openapi, US-api.webhooks.manage
└── Task 8.1 → All F1-F3 frontend US (indirectly via Apollo wiring)
```

## Commit Plan

| Commit | Scope | Message |
|--------|-------|---------|
| **Commit 0** | Phase 0 | `test: add E2E tests and missing user stories for M1 (TDD anchor)` |
| **Commit 1** | Phase 1 | `feat: add cargo workspace, axum framework, Neo4j/PostgreSQL drivers, Gin gateway` |
| **Commit 2** | Phase 2 | `feat: implement TBox CRUD (classes, properties, graph queries)` |
| **Commit 3** | Phase 3 | `feat: implement versioning engine (commits, branches, checkout/rollback)` |
| **Commit 4** | Phase 4 | `feat: implement ABox individual management` |
| **Commit 5** | Phase 5 | `feat: add Turtle and RDF/XML import/export` |
| **Commit 6** | Phases 6-7 | `feat: add API proxy routing, SPARQL/CYPHER endpoints, REST API, OpenAPI` |
| **Commit 7** | Phase 8 | `test: add integration tests for Neo4j, PostgreSQL, API Gateway interfaces` |
| **Commit 8** | Phase 9 | `feat: wire frontend Apollo to real backend, remove mocks` |
| **Commit 9** | Phase 10 | `docs: update traceability.ttl, add Antora documentation` |

## Tasks

- [ ] <!-- Progress tracking block — update [ ] to [x] as tasks are completed -->
- [x] **Total: 28 tasks** | **Completed: 14** | **In progress: 0**

---

### Phase 0: Specs, Tests & Gap Closure (TDD Anchor — FIRST)

> Per RULES.md: "Major-feature plans start with E2E tests" + "Plan tests in order: E2E first, then integration, then unit just-in-time per task."

- [x] **Task 0.1: Perform traceability change impact analysis**
    Before writing any code, query `.ai-factory/traceability/traceability.ttl` to identify all artifacts affected by M1 changes. Document the affected artifacts and their vdo:affectedBy relationships. This ensures no downstream dependencies are broken during implementation.
    
    Also read the US coverage: review `specs/user-stories/_matrix.md` for complete mapping of existing US to M1 scope.
    
    Specs references:
    - ADR-IMPL.PROCESS.repository-layout-strategy (monorepo layout)
    - specs/user-stories/_matrix.md (US→UC mapping)
    - .ai-factory/traceability/traceability.ttl (artifact traceability)
    
    Files:
    - Read: `.ai-factory/traceability/traceability.ttl`
    - Read: `specs/user-stories/_matrix.md`
    - Create: no new files — log findings to implementation session
    
    Dependencies: none

- [x] **Task 0.2: Create missing user stories for M1 gaps**
    Write 9 new user story files in `specs/user-stories/` following the existing Gherkin format (`@US-<domain>.<feature> @UC-... @P...`). Place in the `specs/` submodule. Each must link to its use case via the matrix.
    
    New US to create:
    1. **US-editor.properties.edit-delete** (P0) — Edit/delete ObjectProperty and DatatypeProperty (complement to existing create US). Reference: `UC-editor.properties.manage-property-lifecycle`.
    2. **US-editor.classes.hierarchy-drag-drop** (P1) — Reorder class hierarchy via drag-n-drop in class tree. Reference: `UC-editor.classes.manage-class-lifecycle`.
    3. **US-browse.individuals.list-by-class** (P1) — View paginated list of individuals filtered by class. Reference: `UC-abox.individuals.manage-individual-lifecycle`.
    4. **US-git.branches.create-switch** (P1) — Create a new branch from current head, switch between branches. Reference: `UC-git.branches.manage-branch-workflow`.
    5. **US-abox.individuals.edit-property-values** (P1) — Set, modify, remove property values on an individual. Reference: `UC-abox.individuals.manage-individual-lifecycle`.
    6. **US-api.ontologies.read-rest** (P1) — List ontologies and read classes/properties/individuals via REST. Reference: `UC-api.integration.integrate-through-platform-apis`.
    7. **US-api.sparql.execute** (P1) — Execute SPARQL SELECT query via API endpoint `/api/v1/sparql`. Reference: `UC-browse.search.execute-sparql-query-through-gui`.
    8. **US-api.cypher.execute** (P1) — Execute CYPHER MATCH query via API endpoint `/api/v1/cypher`. Reference: `UC-browse.search.execute-sparql-query-through-gui` (shared UC).
    9. **US-browse.individuals.filter-by-property** (P1) — Filter individual list by property values. Reference: `UC-abox.individuals.manage-individual-lifecycle`.
    
    Also update `specs/user-stories/_matrix.md` to include the new entries.
    
    Logging: INFO for each created US, WARN for any conflicts with existing US IDs.
    
    Specs references:
    - specs/user-stories/US-editor.classes.create-parents.md (existing pattern)
    - specs/user-stories/US-abox.individuals.create.md (existing pattern)
    - specs/user-stories/_matrix.md (update)
    - vision.md §2.5 (MVP scope)
    
    Files:
    - Create: `specs/user-stories/US-editor.properties.edit-delete.md`
    - Create: `specs/user-stories/US-editor.classes.hierarchy-drag-drop.md`
    - Create: `specs/user-stories/US-browse.individuals.list-by-class.md`
    - Create: `specs/user-stories/US-git.branches.create-switch.md`
    - Create: `specs/user-stories/US-abox.individuals.edit-property-values.md`
    - Create: `specs/user-stories/US-api.ontologies.read-rest.md`
    - Create: `specs/user-stories/US-api.sparql.execute.md`
    - Create: `specs/user-stories/US-api.cypher.execute.md`
    - Create: `specs/user-stories/US-browse.individuals.filter-by-property.md`
    - Update: `specs/user-stories/_matrix.md`
    
    Dependencies: Task 0.1

- [x] **Task 0.3: Create missing E2E scenario specs for versioning and query endpoints**
    Write 2 new E2E scenario files in `specs/user-stories/` following the existing Gherkin format (`@E2E-<domain>.<feature>`):
    
    1. **E2E-versioning.branches.switch-rollback** (P1) — Covers:
       - Create branch from current head
       - Switch to new branch
       - Modify ontology on new branch
       - View commit history on new branch
       - Switch back to original branch (verify state unchanged)
       - Rollback to previous commit on original branch
       - Verify materialized state in Neo4j matches expected
       
       References: `US-git.branches.create-switch` [NEW], `US-git.commits.rollback`, `US-git.commits.create`
    
    2. **E2E-query.sparql.execute** (P1) — Covers:
       - Execute SPARQL SELECT query against ontology
       - Execute CYPHER MATCH query against ontology
       - Reject mutation SPARQL query (INSERT) with 400
       - Verify unified response format (table for SELECT)
       - Execute with invalid query syntax (verify 400)
       - Execute without authentication (verify 401)
       
       References: `US-api.sparql.execute` [NEW], `US-api.cypher.execute` [NEW]
    
    Logging: INFO for each created E2E spec.
    
    Specs references:
    - specs/user-stories/E2E-editor.workflow.full-cycle.md (existing pattern)
    - specs/user-stories/E2E-api.integration.rest.md (existing pattern)
    
    Files:
    - Create: `specs/user-stories/E2E-versioning.branches.switch-rollback.md`
    - Create: `specs/user-stories/E2E-query.sparql.execute.md`
    
    Dependencies: Task 0.2 (new US must exist before E2E can reference them)

- [x] **Task 0.4: Write Playwright E2E tests based on user stories**
    Implement Playwright E2E tests in `tests/e2e/` that implement the scenarios defined in:
    
    - `E2E-editor.workflow.full-cycle` — full ontology lifecycle (P0). Covers US:
      `US-editor.classes.create-parents`, `US-editor.properties.create-object`,
      `US-editor.properties.create-datatype`, `US-abox.individuals.create`,
      `US-git.commits.create`, `US-browse.tree.graph-view`
    
    - `E2E-api.integration.rest` — REST API integration (P1). Covers US:
      `US-api.classes.create-rest`, `US-api.auth.jwt`, `US-api.webhooks.manage`,
      `US-io.export.canonical-turtle`
    
    - `E2E-versioning.branches.switch-rollback` [NEW] — branch and rollback flow (P1). Covers US:
      `US-git.branches.create-switch` [NEW], `US-git.commits.rollback`
    
    - `E2E-query.sparql.execute` [NEW] — SPARQL/CYPHER query execution (P1). Covers US:
      `US-api.sparql.execute` [NEW], `US-api.cypher.execute` [NEW]
    
    Each test file references the E2E spec ID and the covered US IDs in comments.
    
    These tests serve as the behavioral specification for M1. They will initially fail and pass incrementally.
    
    Logging: DEBUG for test setup/teardown, INFO for test step execution referencing US ID, ERROR for assertion failures with DOM snapshot.
    
    Specs references:
    - specs/user-stories/E2E-editor.workflow.full-cycle.md
    - specs/user-stories/E2E-api.integration.rest.md
    - specs/user-stories/E2E-versioning.branches.switch-rollback.md [NEW]
    - specs/user-stories/E2E-query.sparql.execute.md [NEW]
    - specs/user-stories/_matrix.md (for US→E2E mapping)
    
    Files:
    - Create: `tests/e2e/ontology-lifecycle.spec.ts` (implements E2E-editor.workflow.full-cycle + E2E-versioning)
    - Create: `tests/e2e/api-integration.spec.ts` (implements E2E-api.integration.rest)
    - Create: `tests/e2e/query-execution.spec.ts` (implements E2E-query.sparql.execute)
    - Create: `tests/e2e/fixtures/ontology-test-data.ts`
    - Create: `tests/e2e/pages/ontology-workspace.page.ts` (Page Object Model)
    
    Dependencies: Task 0.3

---

### Phase 1: Infrastructure Foundation

> Per RULES.md: "Write unit tests just-in-time per task" — each task below includes writing unit tests immediately before implementation of that specific feature.

- [x] **Task 1.1: Set up Rust cargo workspace and axum framework**
    Write unit tests first: test axum route registration, health check response format, tracing span propagation.
    
    Then implement: Create shared cargo workspace (`src/services/shared/`) with common types, health check middleware, metrics middleware, and tracing utilities. Add `axum` + `tokio` to both ontology-service and versioning-service. Replace raw `TcpListener` with axum `Router`. Preserve existing `/health`, `/ready`, `/metrics` endpoints. Ensure no stub markers remain in production code.
    
    Logging: DEBUG for route registration, INFO for startup with listener address, ERROR for binding failures.
    
    Specs references:
    - ADR-IMPL.STACK.ontology-rust-strategy (Rust + tokio + cargo workspace)
    - ADR-IMPL.STACK.version-control-rust-strategy (shared cargo workspace)
    - REQ-CON.STACK.backend-service-stack (Rust for compute-intensive services)
    
    Files:
    - Update: `src/services/ontology-service/Cargo.toml` (add axum, tokio, tracing, serde)
    - Update: `src/services/versioning-service/Cargo.toml` (add axum, tokio, tracing, serde)
    - Create: `src/services/shared/Cargo.toml` (workspace member)
    - Create: `src/services/shared/src/lib.rs`
    - Create: `src/services/shared/src/health.rs`
    - Create: `src/services/shared/src/tracing.rs`
    - Create: `src/services/shared/src/metrics.rs`
    - Create: `src/services/Cargo.toml` (workspace root)
    - Update: `src/services/ontology-service/src/main.rs` (axum router)
    - Update: `src/services/versioning-service/src/main.rs` (axum router)
    
    Dependencies: Task 0.4 (E2E tests define expected behavior)

- [x] **Task 1.2: Integrate Neo4j driver into ontology-service**
    Write unit tests first: test connection pool initialization, health check query (`RETURN 1`), connection retry logic.
    
    Then implement: Add `neo4rs` crate. Create connection pool manager with config from env (`NEO4J_URI`, `NEO4J_USER`, `NEO4J_PASSWORD`). Add health check that runs `RETURN 1` Cypher query. Inject pool as axum application state. Add connection retry with exponential backoff on startup.
    
    Logging: DEBUG for pool state changes, INFO for successful connection, ERROR for connection failures with retry attempt count.
    
    Specs references:
    - ADR-DES.DATA.storage-stack-strategy (Neo4j for graph store)
    - REQ-CON.INFRA.containerization (Neo4j Docker config)
    
    Files:
    - Update: `src/services/ontology-service/Cargo.toml` (add neo4rs)
    - Create: `src/services/ontology-service/src/neo4j.rs` (connection manager)
    - Create: `src/services/ontology-service/src/neo4j/mod.rs`
    - Update: `src/services/ontology-service/src/main.rs` (init pool, inject state, update health check)
    
    Dependencies: Task 1.1

- [x] **Task 1.3: Integrate PostgreSQL driver into versioning-service**
    Write unit tests first: test connection pool init, migration application, table existence verification.
    
    Then implement: Add `sqlx` with postgres feature + `sqlx-cli`. Create connection pool from `DATABASE_URL`. Add migration runner. Create initial migration for versioning schema (commits, branches, deltas tables). Inject pool as axum state.
    
    ```sql
    -- migrations/001_initial_schema.sql
    CREATE TABLE branches (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        name VARCHAR(255) NOT NULL,
        ontology_id UUID NOT NULL,
        head_commit_id UUID,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        is_protected BOOLEAN NOT NULL DEFAULT FALSE
    );
    CREATE TABLE commits (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        branch_id UUID NOT NULL REFERENCES branches(id),
        parent_commit_id UUID REFERENCES commits(id),
        message TEXT NOT NULL,
        author_id VARCHAR(255) NOT NULL,
        author_name VARCHAR(255) NOT NULL,
        delta JSONB NOT NULL DEFAULT '{}',
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    ```
    
    Logging: DEBUG for pool stats, INFO for migration status, WARN for duplicate migrations, ERROR for connection/migration failures.
    
    Specs references:
    - ADR-DES.DATA.storage-stack-strategy (PostgreSQL + JSONB for version store)
    - REQ-FUN.DATA.versioning (version store specification)
    
    Files:
    - Update: `src/services/versioning-service/Cargo.toml` (add sqlx with postgres, uuid, chrono)
    - Create: `src/services/versioning-service/src/postgres.rs` (connection manager)
    - Create: `src/services/versioning-service/migrations/001_initial_schema.sql`
    - Create: `src/services/versioning-service/migrations/002_add_commit_indexes.sql`
    - Update: `src/services/versioning-service/src/main.rs` (init pool, run migrations)
    
    Dependencies: Task 1.1

- [x] **Task 1.4: Wire API Gateway with Gin and auth middleware**
    Write unit tests first: test Gin route registration, auth middleware invocation, CORS header presence, health endpoint accessibility without auth.
    
    Then implement: Update `main.go` to use Gin instead of raw `net/http`. Wire existing `auth/auth.go` middleware into Gin router with exempt paths (health, ready, metrics, public browse). Add CORS middleware for frontend origin. Define route groups:
    - `GET /health`, `GET /ready`, `GET /metrics` — no auth
    - `/api/v1/ontologies/*` — auth required, proxy to ontology-service
    - `/api/v1/versioning/*` — auth required, proxy to versioning-service
    - `/api/v1/sparql`, `/api/v1/cypher` — auth required
    - `/api/v1/public/*` — no auth
    
    Logging: DEBUG for route registration, INFO for auth decisions (granted/denied), WARN for rejected requests with reason code, ERROR for middleware panics.
    
    Specs references:
    - ADR-DES.API.protocol-stack-strategy (API Gateway as single facade)
    - ADR-DES.API.unified-root-endpoint-adoption (/api/v1/ prefix)
    - ADR-DES.SECURITY.gitlab-like-organization-model (RBAC model)
    - US-api.auth.jwt (JWT authentication)
    
    Files:
    - Update: `src/services/api-gateway/main.go` (Gin router, auth wiring, remove net/http stub)
    - Create: `src/services/api-gateway/routes.go` (route group definitions)
    - Create: `src/services/api-gateway/middleware/cors.go`
    - Update: `src/services/api-gateway/auth/auth.go` (verify exempt paths list)
    
    Dependencies: Task 0.4 (E2E tests include auth flow)

---

### Phase 2: Ontology Service — TBox (Classes & Properties)

- [x] **Task 2.1: Implement class CRUD with Neo4j**
    Write unit tests first: test class creation Cypher query, class retrieval with hierarchy, class update with parent rewrite, class deletion with cascade warning.
    
    Then implement: Full class lifecycle with Neo4j Cypher queries:
    - Create class with multiple parents (multiple inheritance)
    - Read class by ID with hierarchy (parents, children), labels, comments
    - Update class: modify label/comment, rewrite parent relationships
    - Delete class: check for dependent classes, properties referencing (domain/range), cascade option
    - List classes: paginated tree with search filter by label
    
    Logging: DEBUG on operation entry with entity ID, INFO for create/update/delete with result summary, WARN for cascade warnings with dependency count, ERROR for Neo4j query failures with full query context (sanitized).
    
    Specs references:
    - **US-editor.classes.create-parents** — creates class with multiple parents
    - **US-editor.classes.edit** — edits class label and comment
    - **US-editor.classes.delete** — deletes class with cascade check
    - US-editor.classes.hierarchy-drag-drop [NEW] — prepares for drag-n-drop
    - REQ-USR.UI.tbox-editor — TBox editor UI requirements
    - UC-editor.classes.manage-class-lifecycle — full class lifecycle
    
    Files:
    - Create: `src/services/ontology-service/src/models/mod.rs`
    - Create: `src/services/ontology-service/src/models/class.rs` (Class struct, NewClass, UpdateClass, ClassSummary)
    - Create: `src/services/ontology-service/src/repositories/mod.rs`
    - Create: `src/services/ontology-service/src/repositories/class_repo.rs` (all Cypher queries)
    - Create: `src/services/ontology-service/src/handlers/mod.rs`
    - Create: `src/services/ontology-service/src/handlers/class_handler.rs` (axum handlers)
    - Create: `src/services/ontology-service/src/routes.rs` (axum Router definition)
    - Create: `src/services/ontology-service/src/error.rs` (domain-prefixed error codes: ONT-CLASS-NOT-FOUND, ONT-CLASS-DUPLICATE, ONT-CLASS-CASCADE-DELETE)
    - Create: `src/services/ontology-service/src/lib.rs` (app builder)
    
    Dependencies: Task 1.2

- [x] **Task 2.2: Implement ObjectProperty and DatatypeProperty CRUD**
    Write unit tests first: test property creation with domain/range, ObjectProperty vs DatatypeProperty validation, property listing by class, property update with characteristic changes.
    
    Then implement:
    - ObjectProperty: create with domain (set of class IDs), range (set of class IDs), characteristics (transitive, symmetric, inverse, functional). Store via `(:Class)-[:DOMAIN]->(:Property)-[:RANGE]->(:Class)`.
    - DatatypeProperty: create with domain (set of class IDs), range (XSD type: string, integer, date, boolean, float). Store as `(:Property {is_datatype: true, xsd_type: "string"})`.
    - Edit/Delete for both property types (edit characteristics, domain, range; cascade check on delete).
    - List properties by class (direct + inherited via class hierarchy traversal).
    - Annotation management CRUD (rdfs:label, rdfs:comment, custom annotation properties).
    - Property search by label with autocomplete.
    
    Logging: DEBUG on operation with property ID, INFO for create/update/delete, WARN for domain/range class not found, ERROR for data integrity violations.
    
    Specs references:
    - **US-editor.properties.create-object** — creates ObjectProperty with domain/range
    - **US-editor.properties.create-datatype** — creates DatatypeProperty with XSD type
    - **US-editor.properties.edit-delete** [NEW] — edit/delete properties
    - US-editor.annotations.add-label — annotation management
    - UC-editor.properties.manage-property-lifecycle — property lifecycle
    - UC-editor.properties.manage-ontology-annotations — annotation lifecycle
    
    Files:
    - Create: `src/services/ontology-service/src/models/property.rs` (Property struct, ObjectProperty, DatatypeProperty, PropertyType enum)
    - Create: `src/services/ontology-service/src/repositories/property_repo.rs`
    - Create: `src/services/ontology-service/src/handlers/property_handler.rs`
    - Update: `src/services/ontology-service/src/routes.rs` (add property routes)
    
    Dependencies: Task 2.1 (depends on class model)

- [x] **Task 2.3: Implement class hierarchy tree and graph neighborhood queries**
    Write unit tests first: test ancestor chain query, descendant tree with depth limit, graph neighborhood with property edges, search with autocomplete, pagination consistency.
    
    Then implement Neo4j traversal queries:
    - Full hierarchy tree: ancestors to root + descendants with configurable max depth
    - Breadcrumb path: ordered list from root ancestor to given class
    - Search/lookup: fulltext search by label/name with autocomplete (LIMIT 20, sorted by relevance)
    - Graph neighborhood: connected classes via ObjectProperties (subject→property→object triples)
    - Lazy-loading: paginated children query (expand node → fetch next page)
    - Response format: hierarchical JSON with `{id, label, children: [...]}` for tree, flat `{nodes: [...], edges: [...]}` for graph
    
    Logging: DEBUG for query parameters and execution time, INFO for result counts, WARN for queries exceeding 500 nodes, ERROR for timeout with query hash.
    
    Specs references:
    - **US-browse.tree.graph-view** — view ontology tree and graph
    - **US-browse.search.fulltext** — fulltext search
    - **US-browse.search.parametric** — parametric search
    - **US-browse.graph.paginated** — paginated graph view
    - UC-browse.tree.view-ontology-tree-and-graph
    - UC-browse.search.search-ontology-elements
    
    Files:
    - Create: `src/services/ontology-service/src/services/mod.rs`
    - Create: `src/services/ontology-service/src/services/graph_service.rs`
    - Update: `src/services/ontology-service/src/repositories/class_repo.rs` (add traversal queries)
    - Update: `src/services/ontology-service/src/handlers/class_handler.rs` (add hierarchy, search, neighborhood endpoints)
    - Update: `src/services/ontology-service/src/routes.rs` (add graph query routes)
    
    Dependencies: Task 2.1

---

### Phase 3: Versioning Service — Commits & Branches

- [x] **Task 3.1: Implement commit model and creation**
    Write unit tests first: test commit creation with delta, commit history pagination, empty delta validation, parent commit linking.
    
    Then implement:
    - Commit struct: `{id, branch_id, parent_commit_id, message, author_id, author_name, delta: JSONB, created_at}`
    - Delta format: `{added_triples: [{s, p, o}], removed_triples: [{s, p, o}], modified_triples: [{s, p, old_o, new_o}]}`
    - Create commit endpoint: accepts `branch_id, message, author`. Computes delta by diffing current ontology state vs previous commit's materialized state. Stores delta as JSONB.
    - Commit history: paginated list with author, message, timestamp, branch name. Filterable by branch_id. Sortable by created_at DESC.
    - GET single commit: full details including delta preview (first 10 triples)
    
    Logging: DEBUG for delta computation with triple counts, INFO for commit creation with commit ID, WARN for empty delta (reject), ERROR for storage failures with SQL error context.
    
    Specs references:
    - **US-git.commits.create** — creates commit with message
    - REQ-FUN.DATA.versioning — Git-like versioning model
    - ADR-DES.DATA.storage-stack-strategy (PostgreSQL + JSONB for deltas)
    - UC-git.commits.manage-commit-history
    
    Files:
    - Create: `src/services/versioning-service/src/models/mod.rs`
    - Create: `src/services/versioning-service/src/models/commit.rs` (Commit, CommitDelta, NewCommit, CommitSummary)
    - Create: `src/services/versioning-service/src/repositories/mod.rs`
    - Create: `src/services/versioning-service/src/repositories/commit_repo.rs` (SQL queries with sqlx)
    - Create: `src/services/versioning-service/src/handlers/mod.rs`
    - Create: `src/services/versioning-service/src/handlers/commit_handler.rs` (axum handlers)
    - Create: `src/services/versioning-service/src/routes.rs`
    - Create: `src/services/versioning-service/src/error.rs` (domain-prefixed codes: VER-COMMIT-NOT-FOUND, VER-COMMIT-EMPTY-DELTA, VER-COMMIT-CONFLICT)
    - Create: `src/services/versioning-service/src/lib.rs` (app builder)
    
    Dependencies: Task 1.3

- [x] **Task 3.2: Implement branch model and management**
    Write unit tests first: test branch creation (fork from existing), branch listing, protected branch enforcement, branch deletion with unmerged warning.
    
    Then implement:
    - Branch struct: `{id, name, ontology_id, head_commit_id, created_at, is_protected}`
    - Create branch: POST accepts `name, ontology_id, source_branch_id`. First commit copies source's head.
    - List branches: GET all branches for an ontology with latest commit info (message, author, timestamp, ahead/behind vs default).
    - Delete branch: DELETE by ID. Protected branches require force flag.
    - Merge branches (basic): POST merge source → target. Combined delta computation, auto-resolve conflicts (source preference on triple-level collision), merge commit with two parents.
    - Switch branch: returns head_commit_id; actual state computation deferred to checkout (Task 3.3).
    
    Logging: DEBUG for branch operations, INFO for create/delete/merge with branch name, WARN for merge conflicts with count and examples, ERROR for merge failures with rollback confirmation.
    
    Specs references:
    - **US-git.branches.create-switch** [NEW] — creates and switches branches
    - **US-git.branches.create-merge** — merges branches with conflict resolution
    - UC-git.branches.manage-branch-workflow
    - ADR-IMPL.STACK.version-control-rust-strategy (branch/MR management)
    
    Files:
    - Create: `src/services/versioning-service/src/models/branch.rs` (Branch, NewBranch, BranchSummary)
    - Create: `src/services/versioning-service/src/repositories/branch_repo.rs`
    - Create: `src/services/versioning-service/src/handlers/branch_handler.rs`
    - Create: `src/services/versioning-service/src/services/mod.rs`
    - Create: `src/services/versioning-service/src/services/merge_service.rs` (diff computation, conflict resolution)
    - Update: `src/services/versioning-service/src/routes.rs`
    - Update: `src/services/versioning-service/migrations/002_add_commit_indexes.sql`
    
    Dependencies: Task 3.1

- [x] **Task 3.3: Implement checkout/rollback and materialized state sync**
    Write unit tests first: test checkout (full state replay from deltas), rollback (inverse delta creation), materialized state cache push, empty state for new branches.
    
    Then implement:
    - Checkout: replay deltas sequentially from root commit through parent chain to target. Produce materialized graph state. Cache result.
    - Rollback: create inverse delta commit between current head and target.
    - Materialized state sync: after checkout, call ontology-service internal endpoint to replace Neo4j state with the complete materialized state (delete + recreate nodes/rels for that ontology).
    - Periodic refresh: for active branches (< 1 hour), auto-refresh every 5 minutes.
    
    Logging: DEBUG for delta replay progress (every 100 deltas), INFO for checkout with triple count and duration, WARN for state sync failures (non-critical — can recompute), ERROR for replay corruption (broken delta chain).
    
    Specs references:
    - **US-git.commits.rollback** — rolls back to previous commit
    - **US-git.commits.compare** — compares ontology versions
    - ADR-IMPL.STACK.version-control-rust-strategy (materialized state in Neo4j)
    - ADR-DES.DATA.storage-stack-strategy (deltas in PostgreSQL, materialized state in Neo4j)
    - UC-git.commits.compare-ontology-versions
    
    Files:
    - Create: `src/services/versioning-service/src/services/delta_service.rs` (delta operations, replay engine)
    - Create: `src/services/versioning-service/src/services/state_service.rs` (checkout, rollback, sync protocol)
    - Create: `src/services/versioning-service/src/services/sync_client.rs` (HTTP client to ontology-service state endpoint)
    - Update: `src/services/versioning-service/src/handlers/commit_handler.rs` (add checkout, rollback handlers)
    - Update: `src/services/versioning-service/src/routes.rs`
    
    Dependencies: Task 3.2

---

### Phase 4: Ontology Service — ABox (Individuals)

- [ ] **Task 4.1: Implement individual (ABox) CRUD**
    Write unit tests first: test individual creation of a class, property value assignment (literal + reference), deletion with reference check, listing with property-based filter, cross-link navigation.
    
    Then implement:
    - Create individual: `CREATE (i:Individual {id, label, class_id, ...})` + `INSTANCE_OF` edge. Auto-generate ID if not provided.
    - Read individual: full card with all literal values (`HAS_VALUE` → LiteralValue node) and reference values (`HAS_REF` → Individual via ObjectProperty).
    - Update individual: add/remove/replace property values. Literals: UPSERT by property_id. References: add/remove edge.
    - Delete individual: check incoming `HAS_REF` edges. Warning with count. Support cascade delete.
    - List by class: paginated table with configurable columns (property values as columns). Dynamic WHERE clause from filter params. Fulltext search across label and ID.
    - Filter/search by property values: `?property_id=X&operator=eq&value=Y` (literals), `?property_id=X&individual_id=Y` (references).
    - Cross-link navigation: from reference value → linked individual's card.
    
    Logging: DEBUG for CRUD with entity ID, INFO for create/delete with class context, WARN for cascade deletes with linked entity count, ERROR for data integrity violations with property/class not found.
    
    Specs references:
    - **US-abox.individuals.create** — creates individual of a class
    - **US-abox.individuals.navigate-link** — navigates to linked individual
    - **US-abox.individuals.edit-property-values** [NEW] — sets property values on individuals
    - **US-browse.individuals.list-by-class** [NEW] — lists individuals filtered by class
    - **US-browse.individuals.filter-by-property** [NEW] — filters individuals by property values
    - UC-abox.individuals.manage-individual-lifecycle
    - UC-abox.individuals.navigate-linked-individuals
    - REQ-USR.UI.abox-editor — ABox editor UI requirements
    
    Files:
    - Create: `src/services/ontology-service/src/models/individual.rs`
    - Create: `src/services/ontology-service/src/repositories/individual_repo.rs`
    - Create: `src/services/ontology-service/src/handlers/individual_handler.rs`
    - Update: `src/services/ontology-service/src/routes.rs`
    - Update: `src/services/ontology-service/src/error.rs` (add: ONT-INDIVIDUAL-NOT-FOUND, ONT-INDIVIDUAL-REFERENCE-EXISTS)
    
    Dependencies: Task 2.1, Task 2.2

---

### Phase 5: Import/Export

- [ ] **Task 5.1: Implement Turtle and RDF/XML export**
    Write unit tests first: test serialization of known ontology to Turtle, RDF/XML, canonical sorting, streaming, empty ontology.
    
    Then implement: Add `oxrdf`/`riot` crates. Full graph traversal → RDF triples. Turtle: `rio_turtle::TurtleWriter` with canonical sorting. RDF/XML: `rio_xml` alternative. Streaming for >100K triples (chunks of 10K).
    Endpoint: `GET /api/v1/ontologies/{id}/export?format=turtle|rdf-xml`. Returns `Content-Type: text/turtle` or `application/rdf+xml`.
    
    Performance target: 100K triples → <10s export.
    
    Logging: DEBUG for serialization progress (every 10K triples), INFO for completion with triple count + duration, WARN for blank nodes, ERROR for serialization failures.
    
    Specs references:
    - **US-io.export.canonical-turtle** — exports canonical Turtle
    - UC-io.import.import-and-export-ontology-data (export path)
    - REQ-USR.UI.import-export
    
    Files:
    - Update: `src/services/ontology-service/Cargo.toml` (add oxrdf, rio_turtle, rio_xml)
    - Create: `src/services/ontology-service/src/services/export_service.rs`
    - Create: `src/services/ontology-service/src/handlers/export_handler.rs`
    - Update: `src/services/ontology-service/src/routes.rs`
    
    Dependencies: Task 2.3 (full graph traversal)

- [ ] **Task 5.2: Implement Turtle and RDF/XML import**
    Write unit tests first: test file parsing (valid/invalid), schema validation (unknown classes), import report, auto-commit.
    
    Then implement:
    - Parse Turtle/RDF/XML via `rio_turtle::TurtleParser` / `rio_xml::RdfXmlParser`
    - Validation pass: verify referenced classes exist for rdf:type, domain, range. Configurable: create stubs vs reject.
    - Apply phase: create/update entities via existing CRUD repository methods.
    - Import report: `{entities_created, updated, skipped, warnings, errors}`
    - Auto-commit after successful import.
    - Endpoint: `POST /api/v1/ontologies/{id}/import` (multipart/form-data).
    - Performance target: <15s for 100K triples with streaming parse.
    
    Logging: DEBUG per parsed triple (max 100, then aggregate), INFO at 10% intervals, WARN for schema validation warnings, ERROR for parse errors with line number + snippet.
    
    Specs references:
    - **US-io.import.turtle** — imports Turtle file
    - UC-io.import.import-and-export-ontology-data (import path)
    - REQ-USR.UI.import-export
    - ADR-DES.UI.import-export-safety-strategy (safety for import/export)
    
    Files:
    - Create: `src/services/ontology-service/src/services/import_service.rs`
    - Create: `src/services/ontology-service/src/services/import_validator.rs`
    - Create: `src/services/ontology-service/src/handlers/import_handler.rs`
    - Update: `src/services/ontology-service/src/routes.rs`
    
    Dependencies: Task 5.1, Task 2.1 (class validation), Task 2.2 (property validation)

---

### Phase 6: API Gateway — Service Proxy & Endpoints

- [ ] **Task 6.1: Implement proxy routing to ontology and versioning services**
    Write unit tests first: test proxy forwarding with correct headers, response passthrough, timeout, auth middleware invocation before proxy, upstream error handling.
    
    Then implement reverse proxy:
    - `/api/v1/ontologies/*` → reverse proxy to `ontology-service:8082`
    - `/api/v1/versioning/*` → reverse proxy to `versioning-service:8083`
    - Header propagation: `X-Trace-Id`, `X-Correlation-Id`, `X-User-Id`, `X-User-Roles`
    - Response streaming: upstream body, headers, status code
    - Timeout middleware: 30s default, configurable via `UPSTREAM_TIMEOUT`. Return 504 on expiry.
    - Circuit breaker: 5 consecutive 5xx → stop routing 30s (503), then retry
    - Error sanitization: no internal details leaked to client
    
    Logging: DEBUG per proxied request (method + path + trace_id), INFO for response (status + duration_ms), WARN for upstream 5xx, ERROR for proxy failures with circuit breaker state.
    
    Specs references:
    - ADR-DES.API.protocol-stack-strategy (API Gateway facade, gRPC internal proxy)
    - **US-api.auth.jwt** — requires auth for all API routes
    - REQ-FUN.API.protocol-stack
    
    Files:
    - Create: `src/services/api-gateway/proxy/proxy.go` (generic reverse proxy)
    - Create: `src/services/api-gateway/proxy/proxy_test.go`
    - Create: `src/services/api-gateway/proxy/circuit_breaker.go`
    - Create: `src/services/api-gateway/middleware/timeout.go`
    - Update: `src/services/api-gateway/routes.go` (add proxy routes)
    - Update: `src/services/api-gateway/main.go` (wire proxy handlers)
    
    Dependencies: Task 1.4

- [ ] **Task 6.2: Implement SPARQL and CYPHER query endpoints**
    Write unit tests first: test query forwarding, read-only enforcement (reject INSERT/DELETE), mandatory LIMIT injection, rate limiting, audit log format.
    
    Then implement:
    - `POST /api/v1/sparql` — accepts `{query: "SELECT ..."}`, proxies to ontology-service SPARQL endpoint. Returns `{results, execution_time_ms, triple_count}`.
    - `POST /api/v1/cypher` — accepts `{query: "MATCH ..."}`, proxies to Neo4j CYPHER endpoint. Same response format.
    - Read-only enforcement: parse and reject mutation queries. Return 400 with `GATEWAY-QUERY-READONLY`.
    - Mandatory LIMIT: inject `LIMIT 1000` if absent. Configurable via `QUERY_MAX_LIMIT`.
    - Rate limiting: per-tier (Anonymous: 10/min, Free: 100/min, Pro: 500/min, Enterprise: 1000/min). 429 + `Retry-After`.
    - Audit logging: `{user_id, query_hash, execution_time_ms, result_count, source_ip}`.
    
    Logging: DEBUG per query execution with hash, INFO for completed (duration + count), WARN for rate limit exceeded or limits triggered, ERROR for failures (sanitized, no raw query).
    
    Specs references:
    - **US-api.sparql.execute** [NEW] — SPARQL query execution
    - **US-api.cypher.execute** [NEW] — CYPHER query execution
    - REQ-FUN.API.graphql-sparql (SPARQL as analytical API)
    - REQ-FUN.API.cypher-query-language (CYPHER support)
    - ADR-DES.API.sparql-query-language-strategy
    - ADR-DES.API.cypher-query-language-adoption
    
    Files:
    - Create: `src/services/api-gateway/handlers/query_handler.go`
    - Create: `src/services/api-gateway/services/query_service.go` (read-only validator, LIMIT injector)
    - Create: `src/services/api-gateway/services/rate_limiter.go` (sliding window per tier)
    - Update: `src/services/api-gateway/routes.go`
    - Update: `src/services/api-gateway/main.go`
    
    Dependencies: Task 6.1

- [ ] **Task 6.3: Add REST read API endpoints and OpenAPI documentation**
    Write unit tests first: test GET endpoint response format, pagination, error codes, OpenAPI spec generation.
    
    Then implement REST read endpoints:
    - `GET /api/v1/ontologies` — list ontologies with `?page=1&per_page=20`, `?q=text`
    - `GET /api/v1/ontologies/{id}` — ontology details with counts
    - `GET /api/v1/ontologies/{id}/classes` — list classes with `?q=`, `?parent_id=X`, pagination
    - `GET /api/v1/ontologies/{id}/classes/{classId}` — class details with hierarchy breadcrumb
    - `GET /api/v1/ontologies/{id}/properties` — list properties with `?type=object|datatype`, pagination
    - `GET /api/v1/ontologies/{id}/individuals` — list individuals with `?class_id=X`, property filter, pagination
    
    Response: `{data: ..., meta: {page, per_page, total, total_pages}}`. Errors: `{error: {code, message}}` with domain-prefixed codes.
    
    Add Swagger/OpenAPI 3.1 auto-generation. Expose at `GET /api/v1/openapi.json`.
    
    Logging: DEBUG per API call with params, INFO for response status + duration, WARN for validation failures.
    
    Specs references:
    - **US-api.classes.create-rest** — creates class via REST (write, but paired with read)
    - **US-api.ontologies.read-rest** [NEW] — reads ontologies via REST
    - **US-api.docs.openapi** — OpenAPI documentation
    - US-api.webhooks.manage — webhook management (paired REST endpoint)
    - UC-api.integration.integrate-through-platform-apis
    - ADR-DES.API.unified-root-endpoint-adoption
    
    Files:
    - Create: `src/services/api-gateway/handlers/ontology_handler.go` (REST handlers)
    - Create: `src/services/api-gateway/models/response.go` (standard response types)
    - Create: `src/services/api-gateway/docs/openapi.json` (OpenAPI 3.1 spec)
    - Update: `src/services/api-gateway/routes.go`
    
    Dependencies: Task 6.1

---

### Phase 7: Integration Tests (External Interfaces)

> Per RULES.md: "Plan integration tests for external interfaces — if a feature touches external interfaces, the plan must include integration tests."

- [ ] **Task 7.1: Integration tests for ontology service — Neo4j and RDF interfaces**
    Write integration tests (real Neo4j via Docker Compose test profile):
    - Class CRUD with Cypher verification (read directly from Neo4j after API operations)
    - Property CRUD with domain/range verification
    - Hierarchy tree query correctness (ancestor chain verification)
    - Graph neighborhood query completeness
    - Individual CRUD with property value storage
    - Import/Export roundtrip: export → Turtle file → import → verify triples match
    - Concurrent operations: two simultaneous creates, test isolation
    - Error scenarios: duplicate class ID, missing parent class, invalid XSD type
    
    Specs references:
    - ADR-DES.DATA.storage-stack-strategy (Neo4j verification)
    - REQ-NFR.PERF.performance (latency targets for CRUD)
    
    Files:
    - Create: `src/services/ontology-service/tests/mod.rs`
    - Create: `src/services/ontology-service/tests/common/mod.rs` (test Neo4j container setup)
    - Create: `src/services/ontology-service/tests/class_integration_test.rs`
    - Create: `src/services/ontology-service/tests/property_integration_test.rs`
    - Create: `src/services/ontology-service/tests/individual_integration_test.rs`
    - Create: `src/services/ontology-service/tests/import_export_integration_test.rs`
    - Create: `src/services/ontology-service/tests/graph_query_integration_test.rs`
    
    Dependencies: Phase 2, Phase 4, Phase 5

- [ ] **Task 7.2: Integration tests for versioning service — PostgreSQL and sync interfaces**
    Write integration tests (real PostgreSQL via Docker Compose):
    - Commit creation and delta storage verification (read from PG)
    - Branch creation, listing, deletion
    - Branch merge with conflict resolution
    - Checkout (full state replay from deltas)
    - Rollback (inverse delta verification)
    - Concurrent commits on different branches (isolation)
    - Materialized state sync correctness (verify Neo4j state after checkout)
    
    Specs references:
    - ADR-DES.DATA.storage-stack-strategy (PostgreSQL delta verification)
    - REQ-FUN.DATA.versioning (version store contract)
    
    Files:
    - Create: `src/services/versioning-service/tests/mod.rs`
    - Create: `src/services/versioning-service/tests/common/mod.rs` (test PostgreSQL container setup)
    - Create: `src/services/versioning-service/tests/commit_integration_test.rs`
    - Create: `src/services/versioning-service/tests/branch_integration_test.rs`
    - Create: `src/services/versioning-service/tests/checkout_integration_test.rs`
    - Create: `src/services/versioning-service/tests/merge_integration_test.rs`
    
    Dependencies: Phase 3

- [ ] **Task 7.3: Integration tests for API Gateway — proxy and auth middleware integration**
    Write integration tests (API Gateway + ontology-service + versioning-service):
    - Auth middleware: valid JWT → 200, expired → 401, missing → 401, insufficient role → 403
    - Ontology proxy: GET classes through gateway vs direct (results match)
    - Versioning proxy: GET commits through gateway
    - SPARQL endpoint: valid query → 200, mutation → 400, missing auth → 401
    - CYPHER endpoint: valid query → 200, mutation → 400
    - Rate limiting: rapid requests → 429
    - Timeout: slow upstream → 504
    - OpenAPI spec: `GET /api/v1/openapi.json` → 200 with valid spec
    
    Specs references:
    - REQ-NFR.SECURITY.bola-bfla-negative-tests (auth rejection scenarios)
    - ADR-DES.API.protocol-stack-strategy (gateway proxy contract)
    - ADR-DES.SECURITY.gitlab-like-organization-model (RBAC enforcement)
    
    Files:
    - Create: `src/services/api-gateway/tests/auth_integration_test.go`
    - Create: `src/services/api-gateway/tests/proxy_integration_test.go`
    - Create: `src/services/api-gateway/tests/query_integration_test.go`
    - Create: `src/services/api-gateway/tests/rate_limit_integration_test.go`
    - Create: `deploy/docker-compose.test.yml` (test profile with all services)
    
    Dependencies: Phase 6

---

### Phase 8: Frontend Wiring

- [ ] **Task 8.1: Connect frontend Apollo queries to real backend data**
    Write unit tests first: test Apollo client config, GraphQL query against mock server, response parsing, error handling with retry, loading state rendering.
    
    Then implement:
    - Update Apollo Client `client.ts`: point `uri` to API Gateway GraphQL endpoint. Keep auth link + add retry link (3 retries, exponential backoff).
    - Implement real GraphQL schema on ontology-service (async-graphql): ClassType, PropertyType, IndividualType, CommitType, BranchType with repository-backed resolvers.
    - Update `queries.ts`: replace mock queries with real operations:
      - `GET_CLASS_TREE`, `GET_CLASS_DETAIL`, `GET_PROPERTIES`, `GET_INDIVIDUALS`
      - `GET_GRAPH_NEIGHBORHOOD`, `GET_COMMIT_HISTORY`, `GET_BRANCHES`
    - Wire ClassTree.vue, PropertyPanel.vue, GraphVisualization.vue, individuals components, CommitHistory.vue, BranchList.vue to real data.
    - Remove all hardcoded mock data from ontology components.
    
    Logging: DEBUG per GraphQL query with variables, INFO for successful results with timing, WARN for partial results, ERROR for failures (sanitized).
    
    Specs references:
    - ADR-IMPL.STACK.frontend-vue-strategy (Vue 3 + Apollo Client)
    - ADR-DES.API.graphql-sparql-split-strategy (GraphQL for navigation queries)
    - ADR-DES.UI.ontology-mental-model-strategy (tree + graph + property panel)
    - ADR-DES.UI.version-context-visibility-strategy (version toolbar)
    - ADR-DES.UI.navigation-state-strategy (navigation state management)
    - US-browse.tree.graph-view (class tree and graph)
    - US-editor.classes.create-parents (class CRUD via UI)
    - US-git.commits.create (commit via UI)
    - REQ-USR.UI.tbox-editor, REQ-USR.UI.abox-editor, REQ-USR.UI.graph-navigation
    
    Files:
    - Update: `src/services/frontend/src/apollo/client.ts` (endpoint URL, retry link)
    - Update: `src/services/frontend/src/apollo/queries.ts` (real queries)
    - Update: `src/services/frontend/src/components/organisms/ClassTree.vue`
    - Update: `src/services/frontend/src/components/organisms/PropertyPanel.vue`
    - Update: `src/services/frontend/src/components/organisms/GraphVisualization.vue`
    - Update: `src/services/frontend/src/components/organisms/OntologyToolbar.vue`
    - Update various versioning components
    - Update: `src/services/ontology-service/Cargo.toml` (add async-graphql)
    - Create: `src/services/ontology-service/src/graphql/mod.rs`
    - Create: `src/services/ontology-service/src/graphql/schema.rs`
    - Create: `src/services/ontology-service/src/graphql/query.rs`
    - Create: `src/services/ontology-service/src/graphql/mutation.rs`
    - Create: `src/services/ontology-service/src/graphql/types.rs`
    - Update: `src/services/ontology-service/src/routes.rs` (add `/api/v1/graphql` route)
    - Update: `src/services/api-gateway/routes.go` (add GraphQL proxy route)
    
    Dependencies: Phase 2, Phase 3, Phase 4, Task 6.1

---

### Phase 9: Traceability & Documentation

- [ ] **Task 9.1: Update traceability.ttl with new and modified M1 artifacts**
    Update `.ai-factory/traceability/traceability.ttl` to reflect:
    - New service declarations for shared cargo workspace (`vdo:Service`)
    - Updated ontology-service, versioning-service, api-gateway, frontend declarations (new coverage levels: `"partial"` or `"full"`)
    - New test suite declarations for integration tests (ontology, versioning, gateway)
    - New E2E test suite declarations
    - New US declarations for the 9 created user stories
    - Traceability relationships: `vdo:implements` (code → US/REQ), `vdo:validates` (tests → code), `vdo:deploys` (Docker Compose → services), `vdo:affectedBy` (inter-service dependencies)
    - Update coverage levels from `"stub"` to appropriate level
    
    Specs references:
    - RULES.md Traceability.1 (keep traceability.ttl up to date)
    - RULES.md Traceability.3 (use vdo: prefix)
    
    Files:
    - Update: `.ai-factory/traceability/traceability.ttl`
    
    Dependencies: All previous phases

- [ ] **Task 9.2: Create Antora user documentation for M1 features**
    Write AsciiDoc documentation under `src/docs/antora/` using Antora component structure. Create component `ontology-editor` with:
    - Component descriptor (`antora.yml`)
    - Quick start guide: create and edit an ontology
    - Class editor: creating, editing, deleting classes with hierarchy
    - Property editor: working with Object and Datatype properties
    - Individual editor: managing instances and property values
    - Versioning: commits, branches, switching, rollback basics
    - Import/Export: supported formats, how to import/export
    - API reference: REST endpoints with examples (narrative + OpenAPI reference)
    
    Specs references:
    - ADR-IMPL.STACK.antora-docs-adoption (Antora for documentation)
    - RULES.md Documentation.1 (user-facing docs in Antora AsciiDoc)
    
    Files:
    - Create: `src/docs/antora/ontology-editor/antora.yml`
    - Create: `src/docs/antora/ontology-editor/modules/ROOT/pages/index.adoc`
    - Create: `src/docs/antora/ontology-editor/modules/ROOT/pages/quickstart.adoc`
    - Create: `src/docs/antora/ontology-editor/modules/ROOT/pages/classes.adoc`
    - Create: `src/docs/antora/ontology-editor/modules/ROOT/pages/properties.adoc`
    - Create: `src/docs/antora/ontology-editor/modules/ROOT/pages/individuals.adoc`
    - Create: `src/docs/antora/ontology-editor/modules/ROOT/pages/versioning.adoc`
    - Create: `src/docs/antora/ontology-editor/modules/ROOT/pages/import-export.adoc`
    - Create: `src/docs/antora/ontology-editor/modules/ROOT/pages/api-reference.adoc`
    - Create: `src/docs/antora/ontology-editor/modules/ROOT/attachments/`
    
    Dependencies: All previous phases

---

## Progress Tracking

```
Total: 28 tasks
├── Phase 0: Specs & E2E Tests   [x] 4/4 — US gap closure, E2E scenarios, Playwright tests
├── Phase 1: Infrastructure      [x] 4/4 — workspace, drivers, gateway
├── Phase 2: TBox CRUD           [x] 3/3 — classes, properties, graph queries
├── Phase 3: Versioning          [x] 3/3 — commits, branches, checkout
├── Phase 4: ABox                [ ] 0/1 — individuals
├── Phase 5: Import/Export       [ ] 0/2 — Turtle, RDF/XML
├── Phase 6: API Gateway         [ ] 0/3 — proxy, SPARQL, REST API
├── Phase 7: Integration Tests   [ ] 0/3 — Neo4j, PG, Gateway interfaces
├── Phase 8: Frontend Wiring     [ ] 0/1 — Apollo → real backend
└── Phase 9: Docs & Trace        [ ] 0/2 — traceability.ttl, Antora docs
```

## Next Steps

Plan created with **32 tasks** across **10 phases**.

Plan file: `.ai-factory/plans/ontology-core-engine.md`

To start implementation, run:
```
$aif-implement
```

To view tasks:
```
/tasks (or use TaskList)
```
