# Implementation Plan: REST API Realignment — Phase A: ADR Formalization + Code Guardrails

Branch: feature/rest-api-realignment-adrs
Created: 2026-08-01

## Progress

### Phase 1: REQ Constraint Files
- [ ] Task 1: REQ-CON.SECURITY.write-path-invariant.md
- [ ] Task 2: REQ-FUN.API.rest-gitlab-alignment.md
- [ ] Task 3: REQ-CON.STACK.publishing-extension.md

### Phase 2: Architecture Decision Records
- [ ] Task 4: ADR-DES.API.write-path-invariant.md
- [ ] Task 5: ADR-DES.API.rest-gitlab-alignment.md
- [ ] Task 6: ADR-DES.INFRA.publishing-extension.md

### Phase 3: Artifact Owner Updates
- [ ] Task 7: Обновить glossary.md (+5 терминов, +4 обновления)
- [ ] Task 8: Обновить vision.md (F3, F6.1, F11, F16.1, MVP 2.5)
- [ ] Task 9: Обновить ROADMAP.md (M5, M10, M11)
- [ ] Task 10: Обновить traceability.ttl (+6 entries, +6 cross-references)

### Phase 4: Validation Tests
- [ ] Task 11: ADR structure validation tests
- [ ] Task 12: traceability.ttl integrity tests

### Phase 5: Bug Fixes
- [ ] Task 13: Fix grpcPool.Close() blocking bug

### Phase 6: Guardrails & Deprecation
- [ ] Task 14: Deprecation headers на /api/v1/ontologies/*
- [ ] Task 15: 501 stubs для /releases, /merge_requests, /protected_branches
- [ ] Task 16: Maintainer-gate annotation в auth.go

### Phase 7: Test Alignment & Suite Verification
- [ ] Task 17: REQ traceability annotations на существующих тестах
- [ ] Task 18: Полный прогон тестов (go + cargo + pytest + Playwright + security)

### Phase 8: Documentation
- [ ] Task 19: Antora developer guide — архитектурные секции

> **Всего: 0/19 задач выполнено**

## Settings
- Testing: yes
- Logging: standard
- Docs: yes

## Roadmap Linkage
Milestone: "M5: MVP Scope Gap Closure"
Rationale: New ADRs formalize architectural decisions discovered during the publishing/write-invariant/REST exploration (session 2026-08-01). Code guardrails (deprecation warnings, planned stubs, grpcPool bugfix) are applied now to align current code with target architecture without requiring M10/M11 implementation. These decisions are prerequisites for M5 REST & Auth alignment and for M10 (MR workflow enforcement) / M11 (publishing pipeline).

## Research Context
Source: .ai-factory/RESEARCH.md (Active Summary, session 2026-08-01 12:00)

Goal:
- Exploration session 2026-08-01: ontology publishing (snapshot serving + access model), write-path invariant (no direct writes), external AI integration (capability scopes), REST API realignment to GitLab.
- Next: formalize via ADRs (write-invariant, REST refactoring, publishing extension) + $aif-plan; apply glossary/vision/roadmap revisions via owners.

Decisions:
1. Publishing: snapshot = read-only serving layer (CQRS; reads >> writes). Access binary: public (no auth) / restricted (API key). Publish gate = maintainer+ (no separate publisher role). Snapshot visibility decoupled from project visibility. Snapshots exposed as /releases (GitLab naming).
2. Write-path invariant: NO direct writes to ontology — nobody (humans, AI, admins), neither ABox nor TBox. Only branch → commit → MR → review → merge; every mutation = logged commit. REST writes to branch snapshot; ontology changes only via logged commit.
3. External AI (DMZ, not part of VEDO): integrates via platform API (F6) — ADR premise "no external integration" stays valid. Capability scopes for machines (NOT role ladder): read, propose:abox, propose:tbox, mr:create, publish:restricted; never write:direct, merge:self, publish:public without human. Autonomy×Authority: ABox auto, TBox deterministic gate, public = human. Review gate = deterministic code (reasoner/policy), never "AI reviews AI". Security containment: AI writes ONLY to own branches (ownership namespace branches/{client}/*); identity separation (proposer ≠ approver ≠ publisher) even in full automation; TBox/ABox split: different channels, frequencies, gates (ABox streamed, TBox versioned).
4. REST API: GitLab-aligned — nested under /projects/{pid}/ (verified vs official GitLab docs; flat contradicts GitLab; global endpoints = read-aggregations only). /api/v1/ontologies/* removed entirely. Content under /projects/{pid}/repository/* with ?ref= (GitLab `ref_name` pattern). MR lifecycle 1:1 GitLab (iid, PUT merge, state_event, /diffs). Publishing = /projects/{pid}/releases. Branch protection = /projects/{pid}/protected_branches. Feature *_access_level on PUT /projects/{pid}.

Findings (code vs specs):
- ADR-DES.INFRA.ontology-publishing NOT implemented: publisher-service + public-browse-api = in-memory stubs; no separate public Neo4j; no rate limiting; publish-browse-ui = stub.
- Write invariant violated in code: ontology-service direct CRUD (classes/properties/individuals) → Neo4j without branch/commit; import replace/merge direct.
- Versioning and CRUD disconnected: commit writes delta to PostgreSQL, does NOT apply to Neo4j (only checkout re-materializes + internal state push).
- MR workflow not implemented (deferred to M10); webhooks (F6.3) declared but absent in code.
- Roles: no publisher (use maintainer weight 2); Keycloak editor="write" contradicts branch model; reviewer in realm, absent in collaboration spec.

Artifact revision (from session 2026-08-01; owners to apply):
- glossary.md: update restApi (project-scoped paths, branch-scoped writes), publisherService/publicBrowseApi (visibility public/restricted + API key), rbac (maintainer publish gate, capability scopes), serviceAccount/m2mToken (external AI machine identity); add terms: Snapshot/Release, API Key, Capability Scope, write-path invariant.
- vision.md: F6.1/F16.1 (project-scoped REST), F11 (snapshot visibility + API key + maintainer gate), F3 (checkout internal), MVP 2.5 REST section.
- ROADMAP.md: M5 Semantic Diff path change (/projects/{pid}/repository/commits/{sha}/diff), M10 MR workflow (invariant enforcement), M11 Publishing (binary visibility, API key, releases naming), M5 REST & Auth (branch-scoped writes).

Open questions:
- ADR-DES.API.organization-rest-endpoints: remove /api/v1/ontologies/{id} from public surface (ontology_id stays internal).
- GraphQL: args ontology_id → project_id (+branch), cursor pagination, DataLoader (ontology-service graphql/query.rs).
- HITL decision: full AI automation requires revising REQ-NFR.SECURITY.llm-write-human-approval (P0) — keep HITL or consciously replace with deterministic gates.
- Capability model: bot-user vs principal_type in membership (auth-service/membership).
- M10 MR workflow is the enforcement backbone — needed before write-invariant holds.
- Public browse shape: flat /api/v1/public/snapshots/{id} vs GitLab-nested /api/v1/public/projects/{pid}/releases/{id} — to decide.
- Storage alternatives for serving layer were analyzed — decision: separate read-only serving store (CQRS); import mechanism for second Neo4j still open.

## REST Endpoint Mapping: Current → Target

> Source: `apps/services/api-gateway/routes.go` (48+ endpoints), `apps/services/api-gateway/docs/openapi.json` (3121 lines)
> Target: ADR-DES.API.rest-gitlab-alignment
> Legend: ✅ already project-scoped · 🔄 needs migration · 🕳️ hole (planned, not implemented)

### 1. Organization (✅ already project-scoped — no migration needed)

| Current | Method | Target | Status |
|---------|--------|--------|--------|
| `/api/v1/groups` | GET/POST | `/api/v1/groups` | ✅ unchanged |
| `/api/v1/groups/{id}` | GET/PUT/DELETE | `/api/v1/groups/{id}` | ✅ unchanged |
| `/api/v1/groups/{id}/subgroups` | GET | `/api/v1/groups/{id}/subgroups` | ✅ unchanged |
| `/api/v1/groups/{id}/members` | GET | `/api/v1/groups/{id}/members` | ✅ unchanged |
| `/api/v1/projects` | GET/POST | `/api/v1/projects` | ✅ unchanged |
| `/api/v1/projects/{id}` | GET/PUT/DELETE | `/api/v1/projects/{pid}` | ✅ unchanged |
| `/api/v1/projects/{id}/members` | GET/POST | `/api/v1/projects/{pid}/members` | ✅ unchanged |
| `/api/v1/projects/{id}/members/{userId}` | PUT/DELETE | `/api/v1/projects/{pid}/members/{userId}` | ✅ unchanged |
| `/api/v1/projects/{id}/visibility` | GET/PUT | `/api/v1/projects/{pid}/visibility` | ✅ unchanged |
| `/api/v1/projects/{id}/policies` | GET/POST | `/api/v1/projects/{pid}/policies` | ✅ unchanged |
| `/api/v1/projects/{id}/policies/{policyId}` | DELETE | `/api/v1/projects/{pid}/policies/{policyId}` | ✅ unchanged |
| `/api/v1/projects/{id}/fork` | POST | `/api/v1/projects/{pid}/fork` | ✅ unchanged |
| `/api/v1/projects/{id}/move` | PUT | `/api/v1/projects/{pid}/move` | ✅ unchanged |

> GitLab param: `ref_name` (query) / `branch` (body). VEDO: `?ref=` for all branch-scoped reads/writes.

### 2. Ontology CRUD (🔄 needs migration to project-scoped)

| Current (`routes.go`) | Method | Target (`/projects/{pid}/repository/...`) |
|------------------------|--------|-------------------------------------------|
| `/api/v1/ontologies` | GET | `/api/v1/projects` (list already exists; ontology-only filter via `?type=ontology`) |
| `/api/v1/ontologies` | POST | `/api/v1/projects` (already exists; create ontology implicitly via CreateProject) |
| `/api/v1/ontologies/{id}` | GET | `/api/v1/projects/{pid}` (ProjectDetail already includes `ontology_id`) |
| `/api/v1/ontologies/{id}` | PUT | `/api/v1/projects/{pid}` (update absorbed into UpdateProject) |
| `/api/v1/ontologies/{id}` | DELETE | `/api/v1/projects/{pid}` (already exists) |
| `/api/v1/ontologies/{id}/classes` | GET | `/api/v1/projects/{pid}/repository/classes?ref=` |
| `/api/v1/ontologies/{id}/classes` | POST | `/api/v1/projects/{pid}/repository/classes?ref=` |
| `/api/v1/ontologies/{id}/classes/{classId}` | GET | `/api/v1/projects/{pid}/repository/classes/{classId}?ref=` |
| `/api/v1/ontologies/{id}/classes/{classId}` | PUT | `/api/v1/projects/{pid}/repository/classes/{classId}?ref=` |
| `/api/v1/ontologies/{id}/classes/{classId}` | DELETE | `/api/v1/projects/{pid}/repository/classes/{classId}?ref=` |
| `/api/v1/ontologies/{id}/properties` | GET | `/api/v1/projects/{pid}/repository/properties?ref=` |
| `/api/v1/ontologies/{id}/properties` | POST | `/api/v1/projects/{pid}/repository/properties?ref=` |
| `/api/v1/ontologies/{id}/properties/{propertyId}` | GET | `/api/v1/projects/{pid}/repository/properties/{propertyId}?ref=` |
| `/api/v1/ontologies/{id}/properties/{propertyId}` | PUT | `/api/v1/projects/{pid}/repository/properties/{propertyId}?ref=` |
| `/api/v1/ontologies/{id}/properties/{propertyId}` | DELETE | `/api/v1/projects/{pid}/repository/properties/{propertyId}?ref=` |
| `/api/v1/ontologies/{id}/individuals` | GET | `/api/v1/projects/{pid}/repository/individuals?ref=` |
| `/api/v1/ontologies/{id}/individuals` | POST | `/api/v1/projects/{pid}/repository/individuals?ref=` |
| `/api/v1/ontologies/{id}/individuals/{individualId}` | GET | `/api/v1/projects/{pid}/repository/individuals/{individualId}?ref=` |
| `/api/v1/ontologies/{id}/individuals/{individualId}` | PUT | `/api/v1/projects/{pid}/repository/individuals/{individualId}?ref=` |
| `/api/v1/ontologies/{id}/individuals/{individualId}` | DELETE | `/api/v1/projects/{pid}/repository/individuals/{individualId}?ref=` |
| `/api/v1/ontologies/{id}/export` | GET | `/api/v1/projects/{pid}/repository/export?ref=` |
| `/api/v1/ontologies/{id}/import` | POST | `/api/v1/projects/{pid}/repository/import?ref=` |
| `/api/v1/ontologies/{id}/validate` | POST | `/api/v1/projects/{pid}/repository/validate?ref=` |
| `/api/v1/ontologies/{id}/draft` | PUT | 🚫 REMOVE — ветка и есть черновик: пиши в branch snapshot, коммить когда готов |
| `/api/v1/ontologies/{id}/comments` | GET/POST | `/api/v1/projects/{pid}/repository/comments?ref=` |

### 3. AI & Document Extraction (🔄 needs migration)

| Current (`routes.go`) | Method | Target (`/projects/{pid}/repository/...`) |
|------------------------|--------|-------------------------------------------|
| `/api/v1/ontologies/{id}/generate-from-text` | POST | `/api/v1/projects/{pid}/repository/ai/generate-from-text?ref=` |
| `/api/v1/ontologies/{id}/ai/suggest-classes` | POST | `/api/v1/projects/{pid}/repository/ai/suggest-classes?ref=` |
| `/api/v1/ontologies/{id}/ai/suggest-properties` | POST | `/api/v1/projects/{pid}/repository/ai/suggest-properties?ref=` |
| `/api/v1/ontologies/{id}/ai/suggest-relationships` | POST | `/api/v1/projects/{pid}/repository/ai/suggest-relationships?ref=` |
| `/api/v1/ontologies/{id}/ai/complete` | POST | `/api/v1/projects/{pid}/repository/ai/complete?ref=` |
| `/api/v1/ontologies/{id}/ai/refine` | POST | `/api/v1/projects/{pid}/repository/ai/refine?ref=` |
| `/api/v1/ontologies/{id}/documents/extract` | POST | `/api/v1/projects/{pid}/repository/documents/extract?ref=` |
| `/api/v1/ontologies/{id}/documents/extract/batch` | POST | `/api/v1/projects/{pid}/repository/documents/extract/batch?ref=` |

### 4. Versioning (🔄 needs migration — GitLab-aligned + VEDO extensions)

> GitLab: commits/diff/branches through `/repository/`; checkout/switch are local git ops, not REST; merge through MR workflow.
> VEDO: `checkout` internal per F3; `switch` removed from REST; `merge` via MR (M10); `rollback` → `revert` (GitLab endpoint); `delta` → `diff`; `{id}` → `{sha}`/`{branch}`; `?ref=` → `?ref=` (GitLab `ref_name`).

| Current | Method | GitLab-aligned target | GitLab match? |
|---------|--------|-----------------------|---------------|
| `/api/v1/versioning/commits` | GET | `/api/v1/projects/{pid}/repository/commits` | Yes (GET /projects/:id/repository/commits) |
| `/api/v1/versioning/commits` | POST | `/api/v1/projects/{pid}/repository/commits` | VEDO ext (GitLab: only via git push) |
| `/api/v1/versioning/commits/{id}` | GET | `/api/v1/projects/{pid}/repository/commits/{sha}` | Yes (GET /:id/repository/commits/:sha) |
| `/api/v1/versioning/commits/{id}/delta` | GET | `/api/v1/projects/{pid}/repository/commits/{sha}/diff` | Yes (rename delta to diff) |
| `/api/v1/versioning/commits/{id}/checkout` | POST | INTERNAL ONLY — not a REST endpoint | No (checkout = local git op) |
| `/api/v1/versioning/commits/{id}/rollback` | POST | POST .../repository/commits (revert commit) | No (GitLab: revert, not rollback) |
| `/api/v1/versioning/branches` | GET | `/api/v1/projects/{pid}/repository/branches` | Yes (GET /:id/repository/branches) |
| `/api/v1/versioning/branches` | POST | `/api/v1/projects/{pid}/repository/branches` | Yes (POST /:id/repository/branches) |
| `/api/v1/versioning/branches/{id}` | GET | `/api/v1/projects/{pid}/repository/branches/{name}` | Yes (use name, not id) |
| `/api/v1/versioning/branches/{id}` | DELETE | `/api/v1/projects/{pid}/repository/branches/{name}` | Yes (DELETE /:id/repository/branches/:branch) |
| `/api/v1/versioning/branches/{id}/switch` | POST | REMOVE — not a REST operation | No (switch = local git op) |
| `/api/v1/versioning/branches/merge` | POST | VIA MR — /merge_requests/{merge_request_iid}/merge (M10) | No (GitLab: merge via MR workflow) |
| *(no current)* | GET | `/api/v1/projects/{pid}/repository/commits/{sha}/diff` | Yes (semantic diff, M5) |
| *(no current)* | GET | `/api/v1/projects/{pid}/repository/compare` | VEDO ext (GitLab: ?from=...&to=...) |

**Result:** 5 of 12 current endpoints not GitLab-aligned. Plan: `checkout` + `switch` removed from REST; `merge` via MR (M10); `rollback` as revert commit; `delta` renamed to `diff`; branch id changed to name.

### 5. Global / Cross-cutting (✅ stays as-is or minor change)

| Current | Method | Target | Status |
|---------|--------|--------|--------|
| `/api/v1/sparql` | POST | `/api/v1/sparql` | ✅ unchanged |
| `/api/v1/cypher` | POST | `/api/v1/cypher` | ✅ unchanged |
| `/api/v1/graphql` | ANY | `/api/v1/graphql` | ✅ unchanged (args: `ontology_id`→`project_id` in schema) |
| `/api/v1/metrics/ontologies` | GET | `/api/v1/metrics/projects` (`ontology_id`→`project_id` in params) | 🔄 minor |
| `/api/v1/metrics/ontologies/{ontology_id}` | GET | `/api/v1/metrics/projects/{pid}` | 🔄 minor |
| `/api/v1/openapi.json` | GET | `/api/v1/openapi.json` | ✅ unchanged |
| `/api/v1/docs` | GET | `/api/v1/docs` | ✅ unchanged |

### 6. 🕳️ Holes — Planned, NOT implemented (return 501 + `x-vedo-status: planned`)

| Target endpoint | Method | Milestone | Purpose |
|-----------------|--------|-----------|---------|
| `/api/v1/projects/{pid}/merge_requests` | GET | M10 | List MRs for a project |
| `/api/v1/projects/{pid}/merge_requests` | POST | M10 | Create MR (source_branch → target_branch) |
| `/api/v1/projects/{pid}/merge_requests/{merge_request_iid}` | GET | M10 | Get MR detail |
| `/api/v1/projects/{pid}/merge_requests/{merge_request_iid}/merge` | PUT | M10 | Merge MR (maintainer+) |
| `/api/v1/projects/{pid}/merge_requests/{merge_request_iid}/diffs` | GET | M10 | MR semantic diff |
| `/api/v1/projects/{pid}/releases` | GET | M11 | List published snapshots |
| `/api/v1/projects/{pid}/releases/{id}` | GET | M11 | Get specific release/snapshot |
| `/api/v1/projects/{pid}/protected_branches` | GET | M10 | List protected branches |
| `/api/v1/projects/{pid}/protected_branches` | POST | M10 | Protect a branch |
| `/api/v1/projects/{pid}/protected_branches/{name}` | DELETE | M10 | Unprotect a branch |
| `/api/v1/projects/{pid}/repository/commits/{sha}/diff` | GET | M5 | Semantic diff (entity-level, per current M5 implementation) |

### 7. 🕳️ Post-1.0 Holes — NOT even stubbed (deferred beyond M11)

| Target endpoint | Method | Milestone | Purpose |
|-----------------|--------|-----------|---------|
| `/api/v1/projects/{pid}/api_keys` | GET/POST | post-M11 | Generate/revoke API keys for restricted publishing |
| `/api/v1/projects/{pid}/api_keys/{id}` | DELETE | post-M11 | Revoke specific API key |
| `/api/v1/projects/{pid}/repository/import/owl` | POST | post-M11 | OWL/RDF import (M11 Import/Export 1.0) |
| `/api/v1/projects/{pid}/repository/export/owl` | GET | post-M11 | OWL/RDF export (M11) |
| `/api/v1/projects/{pid}/repository/export/xlsx` | GET | post-M11 | XLSX export (M11) |
| `/api/v1/public/releases/{id}` | GET | post-M11 | Public browse (shape TBD) |

### Summary

| Category | Count | Action |
|----------|-------|--------|
| ✅ Already project-scoped (unchanged) | 13 | None |
| 🔄 Ontology CRUD → migration | 23 | Phase B (REST migration) |
| 🚫 Ontology CRUD → REMOVE | 1 | draft (ветка = черновик) |
| 🔄 AI/Document → migration | 8 | Phase B |
| 🔄 Versioning → GitLab-aligned migration | 5 | Phase B (rename delta→diff, id→name) |
| 🚫 Versioning → REMOVE from REST | 3 | checkout, switch, direct merge |
| ⚠️ Versioning → VEDO extension | 2 | POST commits, compare |
| 🔄 Minor rename (metrics) | 2 | Phase B |
| 🕳️ Planned (501 stubs) | 11 | **This plan — Task 15** |
| 🕳️ Post-1.0 (no stubs) | 6 | Deferred |
| **Total surface** | **74** | |

## Code Changes by Service

> This plan = Phase A (ADR formalization + guardrails). Phase B (REST migration) changes listed for reference but **deferred**.

### api-gateway (Go) — `routes.go`, `main.go`, `auth/auth.go` (tasks 13–16)

| File | Change | Task | Commit |
|------|--------|------|--------|
| `routes.go` L45 | Move `defer grpcPool.Close()` → return pool ref to caller | 13 | 5 |
| `main.go` | Add `grpcPool.Close()` to graceful shutdown handler (`srv.Shutdown`) | 13 | 5 |
| `routes.go` L120–213 | Add Gin middleware: `Deprecation: true` + `Sunset: <date>` on all `/api/v1/ontologies/*` routes | 14 | 6 |
| `routes.go` (new group) | Register 501 stubs: `GET /projects/{pid}/releases`, `GET/POST /projects/{pid}/merge_requests`, `GET/POST /projects/{pid}/protected_branches` with `x-vedo-status: planned` header + JSON body `{"message":"...","x_vedo_status":"planned"}` | 15 | 6 |
| `auth/auth.go` | Add `const RoleWeightPublish = 2` + comment `// Publish gate per ADR-DES.INFRA.publishing-extension: Maintainer+` | 16 | 6 |

**Endpoints affected in routes.go:**
- 24 `/api/v1/ontologies/*` routes (L120–213) → deprecation headers only (routes NOT removed)
- 3 new stub routes → 501 with `x-vedo-status: planned`
- 1 lifecycle bug (L45 `RegisterRoutes`) → fix

### Specs (Markdown + TTL) — `specs/` + `.ai-factory/` (tasks 1–10)

| File | Change | Task | Commit |
|------|--------|------|--------|
| `specs/requirements/REQ-CON.SECURITY.write-path-invariant.md` | CREATE | 1 | 1 |
| `specs/requirements/REQ-FUN.API.rest-gitlab-alignment.md` | CREATE | 2 | 1 |
| `specs/requirements/REQ-CON.STACK.publishing-extension.md` | CREATE | 3 | 1 |
| `specs/adr/ADR-DES.API.write-path-invariant.md` | CREATE — includes external AI capability scopes subsection | 4 | 2 |
| `specs/adr/ADR-DES.API.rest-gitlab-alignment.md` | CREATE — includes endpoint-class audit table | 5 | 2 |
| `specs/adr/ADR-DES.INFRA.publishing-extension.md` | CREATE — revision of ontology-publishing ADR | 6 | 2 |
| `specs/glossary.md` L631–700 | UPDATE — +5 terms (Snapshot/Release, API Key, Capability Scope, Write-Path Invariant, Branch Snapshot); update 4: restApi, publisherService, rbac, serviceAccount | 7 | 3 |
| `specs/vision.md` F3/F6.1/F11/F16.1/MVP 2.5 | UPDATE — path references `/api/v1/ontologies/*` → `/projects/{pid}/repository/*`; F11: binary visibility + maintainer gate | 8 | 3 |
| `.ai-factory/ROADMAP.md` M5/M10/M11/Notes | UPDATE — M5 diff path, M10 invariant note, M11 publishing model, Notes entry | 9 | 3 |
| `.ai-factory/traceability/traceability.ttl` | UPDATE — +3 ADR entries, +3 REQ entries, +7 cross-reference triples | 10 | 3 |

### Tests (Go + Rust + TS) — new + annotated (tasks 11–12, 17–18)

| File | Change | Task | Commit |
|------|--------|------|--------|
| `tests/specs/adr_structure_test.go` | CREATE — 15–20 cases: frontmatter, sections, cross-refs | 11 | 4 |
| `tests/specs/traceability_integrity_test.go` | CREATE — 12–15 cases: refs, syntax, duplicates + 1 negative security test | 12 | 4 |
| `apps/services/ontology-service/tests/class_integration.rs` | ANNOTATE — `// Validates: REQ-CON.SECURITY.write-path-invariant` | 17 | 7 |
| `apps/services/ontology-service/tests/property_integration.rs` | ANNOTATE — same | 17 | 7 |
| `apps/services/ontology-service/tests/individual_integration.rs` | ANNOTATE — same | 17 | 7 |
| `apps/services/ontology-service/tests/import_export_integration.rs` | ANNOTATE — `// NOTE: violates ADR-DES.API.write-path-invariant; to be migrated in M10` | 17 | 7 |
| `apps/services/api-gateway/auth_integration_test.go` | ANNOTATE — `// Validates: REQ-FUN.API.rest-gitlab-alignment` | 17 | 7 |
| `apps/services/api-gateway/rest_entity_crud_integration_test.go` | ANNOTATE — same | 17 | 7 |
| `tests/security/authorization/rbac_cross_tenant_bola_test.go` | ANNOTATE — `// Validates: REQ-CON.SECURITY.write-path-invariant` | 17 | 7 |
| `tests/security/authorization/rbac_bfla_membership_test.go` | ANNOTATE — same | 17 | 7 |
| `tests/e2e/specs/api/rest/api-gateway-full.spec.ts` | ANNOTATE — `// Validates: REQ-FUN.API.rest-gitlab-alignment` | 17 | 7 |
| `tests/e2e/specs/api/rest/org-api.spec.ts` | ANNOTATE — same | 17 | 7 |

**Full suite run (Task 18):** `go test ./...` (api-gateway, auth-service, vedo-cli) + `cargo test` (ontology-service, versioning-service) + `pytest` (document-extractor) + Playwright E2E smoke + `tests/security/`. Fork BOLA tests (`tests/security/fork_bola_test.go`) remain skipped per ROADMAP M5.

### Documentation (AsciiDoc) — `docs/antora/` (task 19)

| File | Change | Task | Commit |
|------|--------|------|--------|
| `docs/antora/developer-guide/modules/ROOT/pages/architecture.adoc` | UPDATE — 3 sections: Write-Path Invariant, REST API Structure, Publishing Model | 19 | 8 |

### Deferred to Phase B (NOT in this plan — listed for reference)

| Service | What changes | When |
|---------|-------------|------|
| **api-gateway** `routes.go` | 42 routes rewritten: `/api/v1/ontologies/*` → `/projects/{pid}/repository/*`; `/versioning/*` → project-scoped | Phase B |
| **api-gateway** `openapi.json` (3121 lines) | Full rewrite: all paths, schemas, parameters → GitLab-aligned | Phase B |
| **ontology-service** (Rust) | `lib.rs`, `classes.rs`, `properties.rs`, `individuals.rs`, handlers — project-scoped paths | Phase B |
| **ontology-service** `graphql/query.rs` | `ontology_id` arg → `project_id` + optional `branch`; cursor pagination | Phase B |
| **versioning-service** (Rust) | `routes.rs`, `handlers/*`, `sync_client.rs` — project-scoped paths; `/checkout`→internal; `/switch`→removed; `/merge`→via MR | Phase B |
| **publisher-service** (Rust) | `storage.rs`, `handlers/*` — releases, visibility, API key (currently in-memory stubs) | Phase B (M11) |
| **public-browse-api** (Rust) | `snapshot_reader.rs`, `handlers/*` — binary access model (currently stub) | Phase B (M11) |
| **frontend** (Vue 3) | `api/ontology.ts`, `api/versioning.ts`, `api/merge-requests.ts`, 10+ pages, components | Phase B |
| **vedo-cli** (Go) | Command paths updated | Phase B |
| **shared/proto** | Publisher, public_browse stubs → real gRPC contracts | Phase B |
| **Tests** (84+ files) | E2E mock intercepts (`**/api/v1/ontologies/*`), security BOLA/BFLA path tables, integration tests | Phase B |
| **Antora** (5 pages) | API reference, user guide, integrator guide, admin guide | Phase B |

## Commit Plan
- **Commit 1** (after tasks 1-3): "feat(specs): add write-invariant, REST alignment, and publishing extension REQ constraints"
- **Commit 2** (after tasks 4-6): "feat(specs): formalize ADRs for write-path invariant, REST GitLab alignment, and publishing extension"
- **Commit 3** (after tasks 7-10): "docs(specs): update glossary, vision, roadmap, and traceability for new ADRs"
- **Commit 4** (after tasks 11-12): "test(specs): add ADR structure and traceability integrity validation"
- **Commit 5** (after task 13): "fix(api-gateway): move grpcPool.Close() to graceful shutdown hook"
- **Commit 6** (after tasks 14-16): "feat(api-gateway): add deprecation warnings and planned-endpoint stubs"
- **Commit 7** (after tasks 17-18): "test: add REQ traceability annotations; verify full test suite passes"
- **Commit 8** (after task 19): "docs(antora): update developer guide for new architecture decisions"

## Acceptance Criteria

### Specs Quality
- [ ] All 3 ADRs pass structural validation (frontmatter, required sections, cross-references)
- [ ] All 3 REQ files conform to existing naming conventions and structure
- [ ] glossary.md: 5 new terms added, 4 existing terms updated
- [ ] vision.md: F3, F6.1, F11, F16.1, and MVP 2.5 REST section updated
- [ ] ROADMAP.md: M5, M10, M11 path references updated
- [ ] traceability.ttl: all new ADR/REQ entries added; no broken or dangling references
- [ ] Test Quality Score (TQS) ≥ bronze (6.0) for all new test files (Tasks 11, 12, 18)
- [ ] No B1–B7 anti-patterns (see .ai-factory/rules/test-quality.md)
- [ ] `// Validates: REQ-...` traceability annotations present on all new tests

### Code Quality (MR gate — must pass before merge)
- [ ] `grpcPool.Close()` moved to graceful shutdown hook; `go test ./...` in api-gateway passes
- [ ] `/api/v1/ontologies/*` endpoints return `Deprecation: true` + `Sunset: <ISO8601>` headers; no 500 errors
- [ ] `/api/v1/projects/{pid}/releases`, `.../protected_branches`, `.../merge_requests` return `501 Not Implemented` with `x-vedo-status: planned` header
- [ ] `auth.go` maintainer gate annotated: publish action requires weight ≥ 2; comment links to ADR-DES.INFRA.publishing-extension
- [ ] Full existing test suite passes with zero regressions:
  - `go test ./...` (all Go services)
  - `cargo test` (all Rust services)
  - `pytest` (Python services)
  - Playwright E2E (smoke: dashboard → groups → projects → edit → diff)
  - Security: `tests/security/` BOLA/BFLA tests pass (fork_bola may remain RED/skipped)

### MR Readiness
- [ ] All 19 tasks marked complete with `git diff main` showing only planned changes
- [ ] No commented-out code, no `TODO` without ticket reference
- [ ] Commit history is linear and clean (8 атомарных коммитов)
- [ ] `traceability.ttl` passes integrity test (Task 12)

## Tasks

### Phase 1: New REQ Constraint Files
<!-- These are referenced by the ADRs as source-requirement links. Create them first. -->

- [ ] **Task 1: Create REQ-CON.SECURITY.write-path-invariant.md**
  Files to create:
  - `specs/requirements/REQ-CON.SECURITY.write-path-invariant.md`

  Deliverable:
  - Constraint requirement declaring the absolute prohibition of direct writes to ontology storage
  - Scope: nobody (humans, AI agents, admins) writes directly to Neo4j — neither ABox nor TBox
  - Enforcement mechanism: only branch → commit → MR → review → merge pipeline
  - Every mutation MUST be a logged commit with audit trail (who, what, when, branch)
  - REST endpoints operate on branch-local snapshots, never on the live ontology
  - External AI agents are subject to the same invariant (capability scopes limit, not override)
  - References: ADR-DES.DATA.storage-stack-strategy ("deltas per commit"), REQ-FUN.INTEGRATION.collaboration §3.3

  Structure: Follow existing REQ-CON.SECURITY.* format (header, requirement statement, scope, enforcement, references)

  LOGGING REQUIREMENTS (standard):
  - Document decision context in the REQ file itself (rationale section)
  - No runtime code changes in this task
  > BDD naming: N/A (document-only task, no test code)

- [ ] **Task 2: Create REQ-FUN.API.rest-gitlab-alignment.md**
  Files to create:
  - `specs/requirements/REQ-FUN.API.rest-gitlab-alignment.md`

  Deliverable:
  - Functional requirement declaring the REST API target structure (GitLab-aligned)
  - Canonical paths: all write/read operations nested under `/projects/{pid}/`
  - Content paths: `/projects/{pid}/repository/classes|properties|individuals` with `?ref=` parameter
  - MR paths: `/projects/{pid}/merge_requests` (iid-based, PUT merge, state_event, /diffs)
  - Publishing paths: `/projects/{pid}/releases`
  - Branch protection: `/projects/{pid}/protected_branches`
  - Global endpoints: read-aggregations only (list projects, search across projects)
  - Deprecated paths: `/api/v1/ontologies/*` → migration plan with `Deprecation` + `Sunset` headers
  - Backward compatibility: old paths return `501 Not Implemented` with `x-vedo-status: planned` header
  - References: ADR-DES.API.organization-rest-endpoints, GitLab API documentation

  Structure: Follow existing REQ-FUN.API.* format

  LOGGING REQUIREMENTS (standard):
  - Document migration strategy for deprecated paths
  - No runtime code changes in this task
  > BDD naming: N/A (document-only task)

- [ ] **Task 3: Create REQ-CON.STACK.publishing-extension.md**
  Files to create:
  - `specs/requirements/REQ-CON.STACK.publishing-extension.md`

  Deliverable:
  - Extended publishing requirement extending REQ-CON.STACK.ontology-publishing
  - Snapshot model: CQRS architecture with separate read-only serving store
  - Access model: binary — public (no auth) or restricted (API key authentication)
  - Publish gate: maintainer role+ (weight ≥ 2), no separate publisher role
  - Snapshot visibility: decoupled from Project visibility
  - Revisions to original ADR-DES.INFRA.ontology-publishing:
    - Second Neo4j or MinIO+index as serving store (TBD during implementation)
    - API key mechanism for restricted access
    - Rate limiting requirement on public endpoints
    - Publish action as part of MR merge (post-MVP) or manual (MVP)
  - References: REQ-CON.STACK.ontology-publishing, ADR-DES.INFRA.ontology-publishing

  Structure: Follow existing REQ-CON.STACK.* format with explicit revision marker

  LOGGING REQUIREMENTS (standard):
  - Document the evolution from the original publishing REQ
  - No runtime code changes in this task
  > BDD naming: N/A (document-only task)

<!-- Commit checkpoint: tasks 1-3 -->

### Phase 2: New Architecture Decision Records
<!-- Each ADR formalizes a decision from the research session. -->

- [ ] **Task 4: Create ADR-DES.API.write-path-invariant.md** (depends on Task 1)
  Files to create:
  - `specs/adr/ADR-DES.API.write-path-invariant.md`

  Deliverable:
  - Formal ADR declaring the write-path invariant as a first-class architectural constraint
  - Context: current code violates it — ontology-service writes directly to Neo4j; versioning and CRUD are disconnected (commit → PostgreSQL delta, not applied to Neo4j; only checkout re-materializes)
  - Decision: NO path to mutate ontology state exists outside the versioning pipeline. Every mutation = branch → commit → MR → review → merge. REST writes are against branch-local snapshots. Internal handlers call versioning-service, never write directly.
  - External AI containment (subsection):
    - AI agents integrate via platform API (F6), not as internal actors
    - Capability scopes (not RBAC role ladder): read, propose:abox, propose:tbox, mr:create, publish:restricted
    - Forbidden scopes: write:direct, merge:self, publish:public (requires human)
    - Autonomy×Authority: ABox auto-propose OK, TBox deterministic gate required, public publish = human gate
    - Security containment: AI writes ONLY to `branches/{client}/*` namespace
    - Identity separation: proposer ≠ approver ≠ publisher even in full automation
    - Review gate = deterministic code (reasoner, policy engine), never "AI reviews AI"
  - Alternatives: A) Current model (direct writes + disconnected versioning) — rejected (no audit trail, no invariant, data corruption risk); B) Two-phase write (direct + deferred commit) — rejected (races, complexity); C) Full invariant (chosen) — cleanest model, Git-paradigm correct
  - Consequences: M10 MR workflow becomes mandatory enforcement backbone; ontology-service needs refactoring (all write paths → versioning-service); import/export become commit-producing operations; existing direct-write tests need migration
  - References: REQ-CON.SECURITY.write-path-invariant, ADR-DES.DATA.storage-stack-strategy, ADR-DES.PROCESS.merge-request-strategy

  Structure: Follow existing ADR format (Date, Status, Контекст, Требование-источник, Решение, Рассмотренные альтернативы, Последствия, Связанные ADR)

  LOGGING REQUIREMENTS (standard):
  - Document reasoning for each alternative rejection
  - No runtime code changes in this task
  > BDD naming: N/A (document-only task)

- [ ] **Task 5: Create ADR-DES.API.rest-gitlab-alignment.md** (depends on Task 2)
  Files to create:
  - `specs/adr/ADR-DES.API.rest-gitlab-alignment.md`

  Deliverable:
  - Formal ADR declaring the REST API path structure aligned with GitLab conventions
  - Context: current `/api/v1/ontologies/{id}/*` paths are a legacy from pre-project-separation; GitLab model dictates nested paths under `/projects/{pid}/`; flat ontology paths contradict GitLab and the 1:1 Project↔Ontology model
  - Decision: adopt GitLab-compatible REST structure:
    - `GET/POST /api/v1/projects` — list/create projects
    - `GET/PUT/DELETE /api/v1/projects/{pid}` — project CRUD
    - `GET/POST /api/v1/projects/{pid}/repository/classes` with `?ref=` — class CRUD
    - `GET/POST /api/v1/projects/{pid}/repository/properties` with `?ref=`
    - `GET/POST /api/v1/projects/{pid}/repository/individuals` with `?ref=`
    - `GET /api/v1/projects/{pid}/repository/commits` — commit history
    - `GET /api/v1/projects/{pid}/repository/commits/{sha}/diff` — semantic diff
    - `GET/POST /api/v1/projects/{pid}/repository/branches` — branch management
    - `GET/POST /api/v1/projects/{pid}/merge_requests` — MR CRUD (iid-based)
    - `PUT /api/v1/projects/{pid}/merge_requests/{merge_request_iid}/merge` — merge MR
    - `GET /api/v1/projects/{pid}/releases` — published snapshots
    - `GET/POST /api/v1/projects/{pid}/protected_branches` — branch protection
    - `POST /api/v1/projects/{pid}/fork` — fork project (already exists)
  - Migration plan:
    - Old paths (`/api/v1/ontologies/*`) → `501 Not Implemented` with `x-vedo-status: planned` header
    - `ontology_id` in internal code paths stays as internal identifier (not exposed in REST surface)
    - GraphQL: `ontology_id` arg → `project_id` (+ optional `branch`)
  - Endpoint-class audit: all new paths must be included in BOLA/BFLA authorization tables
  - Alternatives: A) Keep current flat paths — rejected (contradicts GitLab model, confusing API surface); B) Dual paths (old + new) — rejected (inconsistent, doubles test surface); C) Full GitLab alignment (chosen) — consistent, predictable, matches organization model
  - Consequences: openapi.json full rewrite; 7 services affected; 84+ test files; Antora docs; traceability.ttl; frontend API clients
  - References: REQ-FUN.API.rest-gitlab-alignment, ADR-DES.API.organization-rest-endpoints, ADR-DES.API.rest-graphql-mutation-boundary

  Structure: Follow existing ADR format

  LOGGING REQUIREMENTS (standard):
  - Document rationale for each path naming decision
  - Include endpoint-class table for BOLA/BFLA audit
  - No runtime code changes in this task
  > BDD naming: N/A (document-only task)

- [ ] **Task 6: Create ADR-DES.INFRA.publishing-extension.md** (depends on Task 3)
  Files to create:
  - `specs/adr/ADR-DES.INFRA.publishing-extension.md`

  Deliverable:
  - Formal ADR extending ADR-DES.INFRA.ontology-publishing with CQRS, access model, and maintainer gate
  - Context: original ADR-DES.INFRA.ontology-publishing defined publisher-service + public-browse-api + second Neo4j but is unimplemented (in-memory stubs); analysis of alternatives identified gaps in access control, visibility, and gate model
  - Decision: extend the publishing model:
    - CQRS: snapshot = materialized read-only serving layer; reads ≫ writes
    - Serving store: separate read-only store (second Neo4j or MinIO+index; decision deferred to implementation)
    - Access model: binary — public (no auth, no rate limiting beyond basic DoS protection) / restricted (API key via `X-VEDO-API-Key` header)
    - Publish gate: maintainer+ (role weight ≥ 2 in api-gateway auth.go RequiredRoleLevel), no separate publisher role
    - Snapshot visibility: decoupled from Project visibility (a private project can publish a public snapshot)
    - Endpoint naming: `/projects/{pid}/releases` (GitLab naming; `snapshots` → `releases`)
    - Public browse shape: TBD — flat `/api/v1/public/releases/{id}` or nested `/api/v1/public/projects/{pid}/releases/{id}`
    - Rate limiting: per-IP and per-API-key; configurable limits
    - API key management: generate/revoke keys via `/projects/{pid}/api_keys` (post-MVP) or project settings
  - Implementation phasing:
    - Phase 1 (M11): snapshot publishing pipeline, second serving store, basic public browse API/UI
    - Phase 2 (post-M11): API key auth, rate limiting, advanced browse features
  - Alternatives: A) Original ADR model (no CQRS, no API key, no visibility decoupling) — rejected (insufficient for real-world publishing); B) Single Neo4j with read replica (Neo4j Enterprise) — deferred for cost analysis; C) Extended model with CQRS (chosen) — evolves the original decision without discarding it
  - Consequences: publisher-service needs full reimplementation (not just activate stubs); public-browse-api needs auth layer (API key validation); operational overhead of second Neo4j/MinIO store
  - References: REQ-CON.STACK.publishing-extension, ADR-DES.INFRA.ontology-publishing, REQ-CON.STACK.ontology-publishing

  Structure: Follow existing ADR format with explicit "Revision of" marker referencing ADR-DES.INFRA.ontology-publishing

  LOGGING REQUIREMENTS (standard):
  - Document evolution from the original ADR
  - Include storage alternative analysis table
  - No runtime code changes in this task
  > BDD naming: N/A (document-only task)

<!-- Commit checkpoint: tasks 4-6 -->

### Phase 3: Artifact Owner Updates
<!-- Update glossary, vision, roadmap, and traceability per the research artifact revision map. -->

- [ ] **Task 7: Update glossary.md** (depends on Tasks 4, 5, 6)
  Files to modify:
  - `specs/glossary.md`

  Deliverable — add 5 new terms:
  1. **Snapshot (Release)** — published read-only view of an ontology at a specific commit; served via CQRS serving layer; exposed as `/projects/{pid}/releases`; distinct from live editing workspace
  2. **API Key** — machine-authentication credential for restricted snapshot access; passed as `X-VEDO-API-Key` header; scoped to specific release; generated/revoked via project settings
  3. **Capability Scope** — permission model for non-human actors (AI agents, external systems); NOT a role ladder; defines what an external identity can DO (read, propose:abox, propose:tbox, mr:create, publish:restricted); never includes write:direct, merge:self, or publish:public
  4. **Write-Path Invariant** — architectural constraint: NO path to mutate ontology state exists outside the versioning pipeline; every mutation = branch → commit → MR → review → merge; enforced across humans, AI agents, and admins
  5. **Branch Snapshot** — local REST-accessible view of a branch's ontology state at a given commit; REST write endpoints operate on branch snapshots, NOT on the live ontology; branch-local writes are later committed and merged

  Deliverable — update 4 existing terms:
  6. **restApi** (API Gateway): add project-scoped paths, branch-scoped writes; note `/api/v1/ontologies/*` → deprecated; clarify `ontology_id` is internal-only
  7. **publisherService / publicBrowseApi**: add visibility levels (public/restricted), API key authentication, CQRS model; note current implementation = stubs
  8. **rbac**: add maintainer publish gate (weight ≥ 2), capability scopes for non-human identities; clarify "editor" role operates within branch model
  9. **serviceAccount / m2mToken**: add external AI machine identity concept; capability scopes attached to service accounts; `branches/{client}/*` ownership namespace

  Structure: Each new term follows the existing pattern: `#### Term Name`, `` `termSlug` ``, definition paragraph(s)

  LOGGING REQUIREMENTS (standard):
  - No runtime code changes in this task
  > BDD naming: N/A (document-only task)

- [ ] **Task 8: Update vision.md** (depends on Tasks 4, 5, 6)
  Files to modify:
  - `specs/vision.md`

  Deliverable — update 5 sections:
  1. **F6.1 (REST API)** — path references `/api/v1/ontologies/{id}/...` → `/projects/{pid}/repository/...`; add branch-scoped writes note
  2. **F16.1 (API Integration)** — same path updates; add `/merge_requests`, `/releases`, `/protected_branches` endpoints
  3. **F11 (Publishing)** — update from single "publishing pipeline" to "binary visibility model (public/restricted)"; add API key mechanism; add maintainer gate; add snapshot → release naming
  4. **F3 (Versioning)** — add "checkout is internal operation" note (REST writes to branch snapshot, not live ontology); clarify commit → materialize flow
  5. **MVP 2.5 REST section** — update path references; add `x-vedo-status: planned` stub policy note; add GitLab-alignment statement

  References: use `git grep` in `specs/vision.md` for `/api/v1/ontologies` to find all occurrences

  LOGGING REQUIREMENTS (standard):
  - No runtime code changes in this task
  > BDD naming: N/A (document-only task)

- [ ] **Task 9: Update ROADMAP.md** (depends on Tasks 4, 5, 6)
  Files to modify:
  - `.ai-factory/ROADMAP.md`

  Deliverable — update 4 milestone references:
  1. **M5 REST & Auth item** — update Semantic Diff path: `/api/v1/versioning/commits/{id}/semantic-diff` → `/projects/{pid}/repository/commits/{sha}/diff`; add "branch-scoped writes" note to REST API item
  2. **M10 MR workflow item** — add "enforcement backbone for write-path invariant (ADR-DES.API.write-path-invariant)" note; add full MR lifecycle (iid, PUT merge, state_event, /diffs)
  3. **M11 Publishing item** — update to "binary visibility (public/restricted), API key auth, releases naming per ADR-DES.INFRA.publishing-extension"; add F11.1 note
  4. **Roadmap Notes** — add entry: "Write-path invariant (ADR-DES.API.write-path-invariant) and REST GitLab alignment (ADR-DES.API.rest-gitlab-alignment) are foundational architectural decisions formalized pre-M5. Implementation deferred to M10 (invariant enforcement) and M11 (publishing pipeline). Code guardrails (deprecation warnings, planned stubs) applied in M5 to align current code with target architecture."

  LOGGING REQUIREMENTS (standard):
  - No runtime code changes in this task
  > BDD naming: N/A (document-only task)

- [ ] **Task 10: Update traceability.ttl** (depends on Tasks 1-9)
  Files to modify:
  - `.ai-factory/traceability/traceability.ttl`

  Deliverable — add new entries and update existing links:
  1. **New ADR entries** (3):
     ```
     base:adr/ADR-DES.API.write-path-invariant a vdo:ArchitectureDecisionRecord .
     base:adr/ADR-DES.API.rest-gitlab-alignment a vdo:ArchitectureDecisionRecord .
     base:adr/ADR-DES.INFRA.publishing-extension a vdo:ArchitectureDecisionRecord .
     ```
  2. **New REQ entries** (3):
     ```
     base:req/REQ-CON.SECURITY.write-path-invariant a vdo:FunctionalRequirement .
     base:req/REQ-FUN.API.rest-gitlab-alignment a vdo:FunctionalRequirement .
     base:req/REQ-CON.STACK.publishing-extension a vdo:FunctionalRequirement .
     ```
  3. **ADR → REQ traces** (3):
     ```
     base:adr/ADR-DES.API.write-path-invariant vdo:tracesTo base:req/REQ-CON.SECURITY.write-path-invariant .
     base:adr/ADR-DES.API.rest-gitlab-alignment vdo:tracesTo base:req/REQ-FUN.API.rest-gitlab-alignment .
     base:adr/ADR-DES.INFRA.publishing-extension vdo:tracesTo base:req/REQ-CON.STACK.publishing-extension .
     ```
  4. **ADR → ADR cross-references** (4):
     ```
     base:adr/ADR-DES.API.write-path-invariant vdo:references base:adr/ADR-DES.DATA.storage-stack-strategy .
     base:adr/ADR-DES.API.write-path-invariant vdo:references base:adr/ADR-DES.PROCESS.merge-request-strategy .
     base:adr/ADR-DES.API.rest-gitlab-alignment vdo:references base:adr/ADR-DES.API.organization-rest-endpoints .
     base:adr/ADR-DES.INFRA.publishing-extension vdo:references base:adr/ADR-DES.INFRA.ontology-publishing .
     ```
  5. **Updated links** — verify no broken `base:req/` references in existing `vdo:validates` triples

  Placement: Insert new entries in the appropriate sections of the file (ADRs section, REQs section, cross-reference section)

  LOGGING REQUIREMENTS (standard):
  - No runtime code changes in this task
  > BDD naming: N/A (document-only task)

<!-- Commit checkpoint: tasks 7-10 -->

### Step N: Update traceability.ttl (integration verification)
  <!-- Per skill-context rule: explicit traceability.ttl update step -->
  - Verify all new `vdo:TestSuite` entries are in place (Tasks 11, 12, 17 create them)
  - Verify all `vdo:validates` triples for `// Validates: REQ-...` annotations
  - Remove any stale triples if files were renamed during artifact updates
  - Ensure turtle syntax validity (no unclosed prefixes, valid URIs)
  > This step is executed after Task 17 (test traceability annotations) as part of the traceability round-trip

### Phase 4: Validation Tests
<!-- Tests for structural integrity of the new ADR/REQ files and traceability cross-references. -->

- [ ] **Task 11: ADR structure validation tests** (depends on Tasks 4, 5, 6)
  Files to create:
  - `tests/specs/adr_structure_test.go`

  Deliverable:
  - Go test suite validating all ADR files in `specs/adr/` conform to the required structure:
    - Frontmatter: Date, Status fields present and valid
    - Required sections: Контекст, Требование-источник, Решение, Рассмотренные альтернативы, Последствия
    - Cross-references: all referenced ADR files exist on disk
    - Source requirements: all linked REQ files exist on disk
    - No broken internal links
  - Test names follow BDD convention: `TestADR_ValidFrontmatter_ShouldPass`, `TestADR_MissingSection_ShouldFail`, etc.
  - Target: 15-20 test cases covering positive and negative scenarios
  - Tests are future-proof: run against ALL ADRs in the directory, not just the new 3
  - `// Validates: REQ-CON.SECURITY.write-path-invariant` annotations on invariant-related tests

  > BDD naming: `[Condition]_[Action]_[ExpectedResult]`
  > Anti-patterns: see `.ai-factory/rules/test-quality.md`
  > New tests must reach TQS ≥ bronze (6.0)

  LOGGING REQUIREMENTS (standard):
  - Log each ADR file processed (INFO) with validation result
  - Log specific section name and line for each validation failure (WARN)
  - Log summary: total ADRs checked, passed, failed (INFO)
  - Use format: `[adr_validator] <message> {file, section, ...}`

- [ ] **Task 12: traceability.ttl integrity tests** (depends on Tasks 10, 11)
  Files to create:
  - `tests/specs/traceability_integrity_test.go`

  Deliverable:
  - Go test suite validating traceability.ttl integrity:
    - All `vdo:ArchitectureDecisionRecord` entries reference existing ADR files
    - All `vdo:validates` triples with `base:req/` subjects reference existing REQ files
    - All `base:adr/` entries have a corresponding `.md` file in `specs/adr/`
    - No dangling references (vdo:references targets exist)
    - Turtle syntax validity (parseable by a standard RDF library)
    - Duplicate detection: no two entries share the same subject URI
  - Test names follow BDD convention
  - Target: 12-15 test cases
  - Security relevance: broken traceability links can hide missing security requirements

  > BDD naming: `[Condition]_[Action]_[ExpectedResult]`
  > Anti-patterns: see `.ai-factory/rules/test-quality.md`
  > New tests must reach TQS ≥ bronze (6.0)
  > Security test: per `.ai-factory/skill-context/aif-plan/SKILL.md` rule — include at least one negative test proving a broken reference is detected

  LOGGING REQUIREMENTS (standard):
  - Log each triple category processed (INFO) with count
  - Log each invalid reference with full subject and object URIs (WARN)
  - Log any parse error with line number (ERROR)
  - Use format: `[traceability_validator] <message> {subject, object, line, ...}`

<!-- Commit checkpoint: tasks 11-12 -->

### Phase 5: Bug Fixes
<!-- Fix known bugs identified in the research session that are blocking or correctness-critical. -->

- [ ] **Task 13: Fix grpcPool.Close() blocking bug in api-gateway**
  Files to modify:
  - `apps/services/api-gateway/routes.go`
  - `apps/services/api-gateway/main.go`

  Deliverable:
  - **Bug:** `defer grpcPool.Close()` in `RegisterRoutes()` closes the gRPC connection pool before the HTTP server starts. HTTP proxy masks the problem — old CRUD routes continue working through HTTP fallback, but gRPC clients never connect.
  - **Fix:**
    1. Remove `defer grpcPool.Close()` from `RegisterRoutes()` (routes.go)
    2. Pass the pool reference to `main()` return value or shutdown context
    3. Call `grpcPool.Close()` in the graceful shutdown handler in `main.go` (alongside existing `srv.Shutdown(ctx)`)
    4. Add INFO log: `[api-gateway] gRPC pool closed during graceful shutdown`
  - **Verification:** `go test ./...` in api-gateway passes; start server locally, verify gRPC connections are established (not just HTTP fallback)

  LOGGING REQUIREMENTS (standard):
  - Log pool initialization in `RegisterRoutes()`: `[api-gateway] gRPC pool initialized {size, targets}`
  - Log pool closure in shutdown: `[api-gateway] gRPC pool closed {reason: "graceful_shutdown"}`
  - Log any close errors: `[api-gateway] gRPC pool close error {error}`
  - Use INFO level for lifecycle, WARN for errors
  > BDD naming: N/A (bugfix; verified via existing gRPC tests)

<!-- Commit checkpoint: task 13 -->

### Phase 6: Guardrails & Deprecation Warnings
<!-- Plug gaps between current code state and target architecture without implementing M10/M11 features. -->

- [ ] **Task 14: Add deprecation warnings on /api/v1/ontologies/* routes** (depends on Task 5)
  Files to modify:
  - `apps/services/api-gateway/routes.go`

  Deliverable:
  - Add `Deprecation: true` and `Sunset: Sat, 01 Aug 2026 00:00:00 GMT` response headers to ALL routes under `/api/v1/ontologies/*`:
    - `/ontologies` (list/create)
    - `/ontologies/:id` (get/update/delete)
    - `/ontologies/:id/classes` (list/create)
    - `/ontologies/:id/properties` (list/create)
    - `/ontologies/:id/individuals` (list/create)
    - `/ontologies/:id/export`
    - `/ontologies/:id/import`
    - `/ontologies/:id/versioning/*`
  - Implementation: add a Gin middleware that injects deprecation headers for matched path prefix
  - Log deprecation warning once per route registration: `[api-gateway] route deprecated {path, sunset, replacement}`
  - **Do NOT remove or disable routes** — this task only adds headers; actual removal happens in Phase B (REST migration)
  - **Verification:** `curl -I http://localhost:8080/api/v1/ontologies/test-id` returns `Deprecation: true` and `Sunset: ...` headers; existing E2E tests still pass (headers are additive, don't change response body)

  LOGGING REQUIREMENTS (standard):
  - Log each deprecated route at startup (INFO): `[api-gateway] deprecated route registered {path, sunset, replacement}`
  - Log each request to deprecated route (INFO, sampled at 1/100): `[api-gateway] deprecated route accessed {path, client_ip}`
  > BDD naming: N/A (middleware change; verified via curl + existing tests)

- [ ] **Task 15: Add x-vedo-status: planned stubs for future endpoints** (depends on Tasks 5, 6)
  Files to modify:
  - `apps/services/api-gateway/routes.go`
  - `apps/services/api-gateway/main.go` (if new route groups needed)

  Deliverable:
  - Register stub endpoints that return `501 Not Implemented` with `x-vedo-status: planned` header:
    - `GET /api/v1/projects/{pid}/releases` — planned for M11
    - `GET /api/v1/projects/{pid}/merge_requests` — planned for M10
    - `POST /api/v1/projects/{pid}/merge_requests` — planned for M10
    - `GET /api/v1/projects/{pid}/protected_branches` — planned for M10
    - `POST /api/v1/projects/{pid}/protected_branches` — planned for M10
  - Response body: `{"message": "Not implemented — planned for M10/M11", "x_vedo_status": "planned"}`
  - **Do NOT implement actual logic** — this is a contract stub for API consumers and test authors
  - **Verification:** `curl http://localhost:8080/api/v1/projects/123/releases` returns `501` with `x-vedo-status: planned`

  LOGGING REQUIREMENTS (standard):
  - Log each stub route registered at startup (INFO): `[api-gateway] planned stub registered {path, milestone, status_code: 501}`
  - Log each stub access (DEBUG, sampled): `[api-gateway] planned stub accessed {path}`
  > BDD naming: N/A (stub endpoints; verified via curl)

- [ ] **Task 16: Add maintainer-gate annotation in auth.go** (depends on Task 6)
  Files to modify:
  - `apps/services/api-gateway/auth/auth.go`

  Deliverable:
  - In `RequiredRoleLevel` function (or equivalent RBAC gate), add annotation comment:
    ```go
    // Publish gate per ADR-DES.INFRA.publishing-extension:
    //   weight >= 2 (Maintainer+)
    //   No separate publisher role — Maintainer performs publish action.
    //   Owner (weight=3) can also publish.
    ```
  - If publish gate is not yet enforced in code (current code has no publish action), add a const:
    ```go
    const RoleWeightPublish = 2 // Maintainer — per ADR-DES.INFRA.publishing-extension
    ```
  - **Do NOT add new auth middleware** — this task only documents the gate; enforcement comes in M11
  - **Verification:** `go build ./...` compiles; no behavioral change

  LOGGING REQUIREMENTS (standard):
  - No runtime logging changes (annotation-only task)
  > BDD naming: N/A (annotation-only task)

<!-- Commit checkpoint: tasks 14-16 -->

### Phase 7: Test Alignment & Suite Verification
<!-- Ensure existing tests are aligned with new ADRs/REQs and the full test suite passes. -->

- [ ] **Task 17: Add REQ traceability annotations to relevant existing tests** (depends on Tasks 1-6, 13-16)
  Files to scan and annotate:
  - `apps/services/ontology-service/tests/class_integration.rs`, `property_integration.rs`, `individual_integration.rs`
  - `apps/services/ontology-service/tests/import_export_integration.rs`
  - `apps/services/api-gateway/auth_integration_test.go`
  - `apps/services/api-gateway/rest_entity_crud_integration_test.go`
  - `tests/security/authorization/rbac_cross_tenant_bola_test.go`
  - `tests/security/authorization/rbac_bfla_membership_test.go`
  - `tests/e2e/specs/api/rest/api-gateway-full.spec.ts`
  - `tests/e2e/specs/api/rest/org-api.spec.ts`

  Deliverable:
  - Add `// Validates: REQ-CON.SECURITY.write-path-invariant` to direct-write tests in ontology-service (these tests currently validate the BROKEN model — annotation documents the gap)
  - Add `// Validates: REQ-FUN.API.rest-gitlab-alignment` to REST path tests in api-gateway integration tests
  - Add `// Validates: REQ-CON.STACK.publishing-extension` to publisher/public-browse related tests
  - Add `// NOTE: violates ADR-DES.API.write-path-invariant; to be migrated in M10` on tests that exercise direct writes
  - Do NOT modify test logic — annotations only
  - `// Validates: REQ-...` follows existing project convention (see `base:ts/` entries in traceability.ttl)

  > BDD naming: N/A (annotation-only task; existing test names unchanged)

  LOGGING REQUIREMENTS (standard):
  - No runtime logging changes (annotation-only task)

- [ ] **Task 18: Run full test suite, verify zero regressions** (depends on Tasks 1-17)
  Test commands:
  - `go test ./...` in `apps/services/api-gateway/`
  - `go test ./...` in `apps/services/auth-service/`
  - `go test ./...` in `apps/vedo-cli/`
  - `cargo test` in `apps/services/ontology-service/`
  - `cargo test` in `apps/services/versioning-service/`
  - `pytest` in `apps/services/document-extractor/`
  - Playwright E2E smoke (if available): `npx playwright test tests/e2e/specs/gui/smoke/screens-rendering.spec.ts`
  - Security suite: `go test ./...` in `tests/security/`

  Deliverable:
  - All test suites **MUST pass** with zero new failures
  - Document any pre-existing failures (e.g., fork_bola skipped tests) as known exceptions
  - If deprecation headers (Task 15) or planned stubs (Task 16) cause test failures:
    - Fix the affected tests to tolerate new headers/stubs
    - Do NOT revert the guardrails
  - Generate a test run report: `test_run_report.txt` with per-suite pass/fail counts
  - Known exceptions: `tests/security/fork_bola_test.go` — 7 tests `t.Skip("not implemented yet")` per ROADMAP M5; these remain skipped

  LOGGING REQUIREMENTS (standard):
  - Log test suite start/finish with suite name (INFO)
  - Log each failure with test name and error summary (ERROR)
  - Log summary: total suites, passed, failed (INFO)
  - Use format: `[test_runner] <message> {suite, passed, failed, duration}`

<!-- Commit checkpoint: tasks 17-18 -->

### Phase 8: Documentation
- [ ] **Task 19: Update Antora developer guide** (depends on Tasks 4-18)
  Files to modify:
  - `docs/antora/developer-guide/modules/ROOT/pages/architecture.adoc`

  Deliverable:
  - Add section "Write-Path Invariant" explaining the architectural constraint:
    - No direct writes to Neo4j; every mutation through versioning pipeline
    - Enforcement deferred to M10 (MR workflow)
    - Current code annotated with deprecation warnings
    - Cross-reference: ADR-DES.API.write-path-invariant
  - Add section "REST API Structure" with GitLab-aligned path table:
    - Current state: `/api/v1/ontologies/*` (deprecated)
    - Target state: `/projects/{pid}/repository/*`, `/merge_requests`, `/releases`, `/protected_branches`
    - Planned stubs returning `501 Not Implemented`
    - Cross-reference: ADR-DES.API.rest-gitlab-alignment
  - Add section "Publishing Model" with CQRS architecture:
    - Separate serving store, binary visibility, API key mechanism
    - Maintainer gate for publish action
    - Cross-reference: ADR-DES.INFRA.publishing-extension
  - Add note: "Code migration to target architecture is deferred to M10 (MR workflow) and M11 (publishing pipeline). Current code includes deprecation warnings and planned stubs per the ADRs above."

  LOGGING REQUIREMENTS (standard):
  - No runtime code changes in this task
  > BDD naming: N/A (document-only task)

<!-- Commit checkpoint: task 19 -->
