# Implementation Plan: Fix REST/GraphQL Architecture Split

**Branch:** `feature/fix-rest-graphql-split`  
**Created:** 2026-07-21  
**Mode:** Fast

## Settings

| Setting | Value |
|---------|-------|
| Testing | Yes — include integration tests |
| Logging | Standard (INFO level) |
| Documentation | Yes — clarify specs in `specs/` |
| Roadmap | No linkage |

## Research Context

### Goal
Fix architectural inconsistency between REST and GraphQL in VEDO Core API Gateway. The ADR says GraphQL = navigation only, REST = all writes, but the code has GraphQL mutations for entity CRUD, SPARQL executed through GraphQL (bypassing gateway DoS protection), and OpenAPI spec missing ontology write endpoints.

### Key Findings

**Current split in code:**

| Layer | What exists | ADR says |
|-------|------------|----------|
| **REST** (routes.go) | Ontology CRUD (POST/PUT/DELETE), Org CRUD, Versioning, SPARQL, CYPHER | All writes should be REST |
| **GraphQL** (mutation.rs) | `updateDraft`, `updateMemberRole`, `removeMember` | Navigation read-only only |
| **GraphQL** (query.rs) | All navigation queries + `sparqlQuery` + comments + dashboard + metrics | Navigation only; SPARQL should be REST-only |
| **Frontend** (queries.ts) | `CREATE_CLASS_MUTATION`, `CREATE_PROPERTY_MUTATION`, `CREATE_INDIVIDUAL_MUTATION` — **not implemented in backend schema** | These should use REST |
| **OpenAPI** (openapi.json) | Only ontology reads + org CRUD | Missing ontology writes, versioning, AI, etc. |

**Critical issues:**
1. `CREATE_CLASS_MUTATION`, `CREATE_PROPERTY_MUTATION`, `CREATE_INDIVIDUAL_MUTATION` exist in `queries.ts` and are used by Vue components but **have no resolver in the Rust GraphQL schema** — they fail silently
2. `sparqlQuery` in GraphQL bypasses API Gateway DoS protection (rate limits, circuit breaker, query complexity)
3. OpenAPI spec is incomplete — missing 15+ write endpoints, versioning, AI, document extractor
4. ADR doesn't explicitly forbid mutations or clarify the REST/GraphQL boundary

---

## Tasks

### Phase 1: Update ADRs — Clarify REST/GraphQL Boundary

- [x] **Task 1: Update ADR-DES.API.graphql-sparql-split-strategy.md**
  - **Files:** `specs/adr/ADR-DES.API.graphql-sparql-split-strategy.md`
  - **Deliverable:** Add explicit "Разделение ответственности" section that clearly states:
    - GraphQL: navigation read-only queries only (tree, class hierarchy, graph neighborhood, autocomplete, properties, individuals, commits, branches)
    - REST: all write operations (ontology CRUD, entity CRUD, comments, validation, import/export)
    - REST: SPARQL and CYPHER query execution (with DoS protection)
    - Forbidden in GraphQL: SPARQL execution, ontology entity mutations (createClass, createProperty, createIndividual, update*, delete*)
  - **Logging:** N/A (spec-only)
  - **Dependency notes:** None — independent

- [x] **Task 2: Update ADR-DES.API.protocol-stack-strategy.md**
  - **Files:** `specs/adr/ADR-DES.API.protocol-stack-strategy.md`
  - **Deliverable:** Clarify the "GraphQL для API навигации по графу клиентской части" line to explicitly state:
    - GraphQL = read-only navigation (queries only, no mutations for ontology entities)
    - REST = all write operations + SPARQL/CYPHER execution
    - Add a table mapping responsibility to protocol
  - **Logging:** N/A (spec-only)
  - **Dependency notes:** None — independent

- [x] **Task 3: Create new ADR-DES.API.rest-graphql-mutation-boundary.md**
  - **Files:** `specs/adr/ADR-DES.API.rest-graphql-mutation-boundary.md` (new)
  - **Deliverable:** New ADR documenting the explicit boundary:
    - Why GraphQL mutations are forbidden for ontology entity CRUD
    - Why SPARQL must be REST-only (DoS protection, circuit breaker, rate limiting)
    - Migration path for frontend components that currently use GraphQL mutations
    - Status: Accepted
  - **Logging:** N/A (spec-only)
  - **Dependency notes:** Depends on Tasks 1 and 2 for context

### Phase 2: Close SPARQL Security Hole

- [x] **Task 4: Remove `sparqlQuery` from GraphQL schema**
  - **Files:**
    - `src/services/ontology-service/src/graphql/query.rs` — remove `sparqlQuery` resolver
    - `src/services/frontend/src/apollo/queries.ts` — remove `SPARQL_EXECUTE_QUERY`
    - `src/services/frontend/src/pages/SPARQLPage.vue` — switch to REST `POST /api/v1/sparql` via `axios`
    - `src/services/frontend/src/apollo/mock-data.ts` — update/remove `MOCK_SPARQL_RESULTS` if needed
  - **Deliverable:** SPARQL execution only through REST endpoint (`POST /api/v1/sparql`), which has DoS protection
  - **Logging:** INFO — log migration events in SPARQLPage.vue (switching from Apollo to axios)
  - **Dependency notes:** Depends on Task 1 (ADR documents the decision)

### Phase 3: Migrate Entity Mutations to REST

- [x] **Task 5: Remove GraphQL entity mutations from frontend, switch to REST**
  - **Files:**
    - `src/services/frontend/src/apollo/queries.ts` — remove `CREATE_CLASS_MUTATION`, `CREATE_PROPERTY_MUTATION`, `CREATE_INDIVIDUAL_MUTATION`
    - `src/services/frontend/src/components/ontology/CreateClassDialog.vue` — replace `useMutation(CREATE_CLASS_MUTATION)` with `axios.post('/api/v1/ontologies/:id/classes', ...)`
    - `src/services/frontend/src/components/ontology/CreatePropertyDialog.vue` — same, switch to REST
    - `src/services/frontend/src/components/ontology/CreateIndividualDialog.vue` — same, switch to REST
  - **Deliverable:** Entity CRUD uses REST endpoints that already exist in routes.go (POST /ontologies/:id/classes, etc.)
  - **Logging:** INFO — log REST calls in dialog components
  - **Dependency notes:** Depends on Task 1 (ADR documents the decision)
  - **Note:** REST write endpoints already exist in `routes.go` (lines 122-133), they just need frontend wiring

- [x] **Task 6: Remove unused GraphQL mutations from Rust schema (cleanup)**
  - **Files:**
    - `src/services/ontology-service/src/graphql/mutation.rs` — verify no orphan mutations exist (createClass, createProperty, createIndividual should NOT be in the Rust schema — they were never implemented, but verify). Mark existing `update_draft`, `update_member_role`, `remove_member` as `#[deprecated]` with REST migration pointers (see ADR-DES.API.rest-graphql-mutation-boundary.md).
    - `src/services/ontology-service/src/graphql/schema.rs` — update comment to reflect final decision: **no GraphQL mutations allowed**, including `updateDraft`/`updateMemberRole`/`removeMember`. They are deprecated placeholders pending REST migration.
  - **Deliverable:** All GraphQL mutations are explicitly deprecated and forbidden (no exceptions). `MutationRoot` will be replaced with `EmptyMutation` in a follow-up REST migration of `update_draft`/`update_member_role`/`remove_member`.
  - **Logging:** N/A (cleanup)
  - **Dependency notes:** Depends on Task 5

### Phase 4: Document REST Write Endpoints in OpenAPI

- [x] **Task 7: Add ontology write endpoints to openapi.json**
  - **Files:** `src/services/api-gateway/docs/openapi.json`
  - **Deliverable:** Add missing endpoints that exist in routes.go but not in openapi.json:
    - POST /ontologies (create ontology)
    - PUT /ontologies/{id} (update ontology)
    - DELETE /ontologies/{id} (delete ontology)
    - POST /ontologies/{id}/classes (create class)
    - PUT /ontologies/{id}/classes/{classId} (update class)
    - DELETE /ontologies/{id}/classes/{classId} (delete class)
    - POST /ontologies/{id}/properties (create property)
    - PUT /ontologies/{id}/properties/{propertyId} (update property)
    - DELETE /ontologies/{id}/properties/{propertyId} (delete property)
    - POST /ontologies/{id}/individuals (create individual)
    - PUT /ontologies/{id}/individuals/{individualId} (update individual)
    - DELETE /ontologies/{id}/individuals/{individualId} (delete individual)
    - GET /ontologies/{id}/export
    - POST /ontologies/{id}/import
    - Versioning endpoints (commits, branches)
  - **Logging:** N/A (spec-only)
  - **Dependency notes:** Independent — can be done in parallel with Phase 2-3

- [x] **Task 8: Create GraphQL schema documentation**
  - **Files:** New file `src/services/api-gateway/docs/graphql-schema.md` (or `specs/api/graphql-schema.md`)
  - **Deliverable:** Document the GraphQL navigation schema:
    - All query types (Class, Property, Individual, Version, etc.)
    - Explicitly state: no mutations for ontology entities
    - List forbidden operations (SPARQL, entity CRUD)
    - Reference to OpenAPI for write operations
  - **Logging:** N/A (spec-only)
  - **Dependency notes:** Depends on Tasks 1-6 (schema is finalized after cleanup)

### Phase 5: Tests

- [x] **Task 9: Add integration test — REST entity CRUD endpoints exist and work**
  - **Files:** `tests/ticket-api/` or new `tests/integration/rest-entity-crud_test.go` (or equivalent)
  - **Deliverable:** Test that verifies:
    - POST /api/v1/ontologies/:id/classes returns 200/201
    - POST /api/v1/ontologies/:id/properties returns 200/201
    - POST /api/v1/ontologies/:id/individuals returns 200/201
    - These endpoints require authentication (401 without token)
  - **Logging:** INFO — test results
  - **Dependency notes:** Depends on Task 5 (frontend wiring confirms endpoints work)

- [x] **Task 10: Add integration test — GraphQL SPARQL is removed**
  - **Files:** New test file or extend existing GraphQL tests
  - **Deliverable:** Test that verifies:
    - `sparqlQuery` is no longer available in GraphQL schema (introspection query)
    - GraphQL navigation queries still work
    - Entity mutations are not in GraphQL schema
  - **Logging:** INFO — test results
  - **Dependency notes:** Depends on Tasks 4 and 6

---

## Commit Plan

- **Commit 1** (after Tasks 1-3): "docs: clarify REST/GraphQL boundary in ADRs and create mutation boundary ADR"
- **Commit 2** (after Tasks 4-6): "fix: remove SPARQL from GraphQL, migrate entity mutations to REST"
- **Commit 3** (after Tasks 7-8): "docs: complete OpenAPI spec and create GraphQL schema documentation"
- **Commit 4** (after Tasks 9-10): "test: add integration tests for REST CRUD and GraphQL schema validation"

---

## Acceptance Criteria

1. ✅ ADR-DES.API.graphql-sparql-split-strategy.md explicitly forbids mutations and SPARQL through GraphQL
2. ✅ ADR-DES.API.protocol-stack-strategy.md has a clear responsibility table (GraphQL=read, REST=write)
3. ✅ New ADR documents the mutation boundary decision
4. ✅ `sparqlQuery` removed from GraphQL schema and frontend
5. ✅ Frontend entity CRUD uses REST (axios/fetch, not Apollo mutations)
6. ✅ OpenAPI spec documents all REST write endpoints
7. ✅ GraphQL schema documentation exists as a separate file
8. ✅ Integration tests verify REST CRUD and GraphQL schema shape
9. ✅ SPARQLPage.vue uses REST endpoint instead of GraphQL

## Files Changed (Summary)

| File | Action | Phase |
|------|--------|-------|
| `specs/adr/ADR-DES.API.graphql-sparql-split-strategy.md` | Update | 1 |
| `specs/adr/ADR-DES.API.protocol-stack-strategy.md` | Update | 1 |
| `specs/adr/ADR-DES.API.rest-graphql-mutation-boundary.md` | Create | 1 |
| `src/services/ontology-service/src/graphql/query.rs` | Remove sparqlQuery | 2 |
| `src/services/frontend/src/apollo/queries.ts` | Remove SPARQL_QUERY + entity mutations | 2, 3 |
| `src/services/frontend/src/pages/SPARQLPage.vue` | Switch to REST | 2 |
| `src/services/frontend/src/components/ontology/CreateClassDialog.vue` | Switch to REST | 3 |
| `src/services/frontend/src/components/ontology/CreatePropertyDialog.vue` | Switch to REST | 3 |
| `src/services/frontend/src/components/ontology/CreateIndividualDialog.vue` | Switch to REST | 3 |
| `src/services/ontology-service/src/graphql/schema.rs` | Update comment | 3 |
| `src/services/api-gateway/docs/openapi.json` | Add write endpoints | 4 |
| `src/services/api-gateway/docs/graphql-schema.md` | Create | 4 |
| Tests (new) | Create | 5 |
