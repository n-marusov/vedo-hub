## Test Plan: Full RBAC Testing

**Date:** 2026-07-23
**Branch / Version:** rbac-full (specification-driven scope)
**Environment:** staging / local docker-compose (all-services)

---

### 1. Testing Goal

Verify that the **complete RBAC perimeter** of VEDO Core enforces authorization correctly across all organization endpoints, role hierarchies, visibility levels, ABAC policies, protected-branch workflow, MFA/JIT/PAM controls, break-glass access, destructive-command guardrails, and public browse access — with zero BOLA/BFLA/IDOR regressions and 100% mandatory negative-test coverage per the authorization regression gate (`REQ-NFR.SECURITY.authorization-regression-gates`).

This is a **blocking release gate**: any P0 critical defect blocks production deploy.

---

### 2. Test Scope

**In Scope** — we test:

- Organization model: `Group`/`Subgroup`/`Project`/`Ontology` 1:1 CRUD and membership
- Role hierarchy: `Viewer`(Guest=10) / `Reporter`(20) / `Editor`(Developer=30) / `Maintainer`(40) / `Owner`(50); max-role-wins inheritance
- REST organization endpoints: `/api/v1/groups/*`, `/api/v1/projects/*`, members, visibility, policies, fork, move
- Visibility levels: `Private` / `Internal` / `Public` enforcement on Project (inherited by Ontology)
- ABAC policies: semantic graph patterns (URI prefix, annotation, parent class, relationship); most-specific-wins
- Protected `main` + Ontology Merge Request (OMR) workflow with `write` vs `write_with_approval`
- BOLA/BFLA/IDOR negative tests: 4 test types (A/B/C/D) across all P0 endpoint-classes
- Fork RBAC: read-access-sufficient rule, 403-never-404
- Idempotency: `Idempotency-Key` requirement on write endpoints
- MFA: categories A/B/C/D; service-account restrictions; non-cacheable for A/B
- JIT/PAM: no standing access, TTL ≤ 60 min, session recording, phishing-resistant MFA
- Break-glass Emergency Admin: Shamir M-of-N, `/emergency/login`, TTL, immutable audit
- Destructive command guardrails: G1–G4 levels
- Audit logging: every write endpoint emits structured `audit_events`
- Public Browse API: no-auth read-only, rate limiting, isolated Neo4j

**Out of Scope** — we don't test:

- Functional correctness of ontology CRUD (class/property/individual editing) — covered by M1/M4 QA
- LLM/AI feature quality (NL→OWL, suggestions) — only their RBAC gates are in scope
- Performance/load testing of authorization checks — separate NFR.PERF suite
- UI/UX visual regression — only authorization-relevant UI states (hidden/disabled controls)
- Backup/restore data integrity — only their authorization and MFA gates are in scope

---

### 3. Test Types

| Type              | Priority   | Area                                                                                          |
|-------------------|------------|----------------------------------------------------------------------------------------------|
| Functional        | 🔴 High    | Role permissions, inheritance, visibility, membership management, fork, OMR workflow         |
| Negative          | 🔴 High    | BOLA cross-tenant, BOLA cross-object, BFLA privilege escalation, IDOR guessable IDs          |
| Security          | 🔴 High    | MFA categories, JIT/PAM, break-glass, destructive guardrails, audit completeness            |
| Edge cases        | 🟡 Medium  | Role conflicts (max-wins), ABAC most-specific-wins, empty/null IDs, concurrent member changes|
| Regression        | 🟡 Medium  | GraphQL→REST migration boundary, legacy `"ontology/"` scope alias, idempotency dedup         |
| Integration       | 🟡 Medium  | API Gateway → auth-service gRPC → backend; WebSocket collaboration join; SPARQL endpoint    |

---

### 4. Test Data

| Category                    | Data                                                                 | Purpose                                        |
|-----------------------------|----------------------------------------------------------------------|------------------------------------------------|
| Roles                       | 5 test users: `u_viewer`, `u_editor`, `u_maintainer`, `u_owner`, `u_outsider` | One per access level + non-member        |
| Tenants                     | `tenant_A`, `tenant_B`                                               | Cross-tenant BOLA (type A)                     |
| Projects                    | `proj_private` (Private), `proj_internal` (Internal), `proj_public` (Public) | Visibility enforcement               |
| Ontologies                  | `ont_A` (in `proj_private`), `ont_B` (in `tenant_B`)                | Object-level authorization                     |
| Branches / Commits          | `branch_main` (protected), `branch_feature`, `commit_123`           | Versioning RBAC                                |
| IDs                         | Valid UUID, sequential numeric (`1001`, `1002`), invalid UUID variant | IDOR type D                              |
| ABAC policies               | Policy on URI prefix `http://vedo/example/`; policy on parent class  | most-specific-wins                      |
| Service-account tokens      | `sa_cat_C` (scoped: `ontology:delete`), `sa_cat_D` (scoped: export)  | Service-account restrictions            |
| MFA factors                 | Valid TOTP code, expired TOTP, missing TOTP                          | MFA category enforcement                       |
| Idempotency keys            | Valid `Idempotency-Key`, duplicate key, missing key                  | Idempotency enforcement                        |

---

### 5. Preconditions

- [ ] All-services docker-compose deployed and healthy (`/health`, `/ready` on every service)
- [ ] Keycloak realm configured with test users (`u_viewer`, `u_editor`, `u_maintainer`, `u_owner`, `u_outsider`) and OIDC clients
- [ ] Two tenants (`tenant_A`, `tenant_B`) seeded with sample groups, projects, ontologies
- [ ] Test projects with all three visibility levels created
- [ ] ABAC policies seeded on at least one project
- [ ] WebSocket endpoint accessible for collaboration join tests
- [ ] SPARQL endpoint available with test ontology graph
- [ ] Audit log store accessible for verification (PostgreSQL `audit_events` table or equivalent)
- [ ] Service-account tokens generated with documented scopes and TTL
- [ ] Emergency Admin Shamir parts distributed to test holders (drill mode)
- [ ] Network access to API Gateway, auth-service gRPC (test only), and Public Browse API

---

### 6. Acceptance Criteria

- [ ] All 🔴 high-priority test cases pass (BOLA/BFLA/IDOR/membership/visibility/MFA/audit)
- [ ] 100% of P0 endpoint-classes have negative tests returning **403** (never 404/500)
- [ ] Zero unauthorized cross-tenant access confirmed
- [ ] Membership operations restricted to `Owner` only (all other roles → 403)
- [ ] Visibility levels enforced: `Private` → members only; `Internal` → authenticated; `Public` → anonymous
- [ ] Fork returns 403 (never 404) when read access denied
- [ ] MFA categories A/B cannot be bypassed or cached
- [ ] Every write endpoint emits `audit_events` with all mandatory fields
- [ ] Break-glass `/emergency/login` works with Keycloak down; TTL enforced; audit immutable
- [ ] Authorization regression gate report generated (JSON) with 0 critical defects and signed policy bundle
- [ ] Gate duration ≤ 15 minutes

---

### 7. Plan Risks

| Risk                                                  | Impact   | Mitigation                                                                                |
|-------------------------------------------------------|----------|--------------------------------------------------------------------------------------------|
| WebSocket BOLA harder to test (non-HTTP)              | Medium   | Use WebSocket test harness with fixture preparation + audit log verification                |
| SPARQL endpoint RBAC is P1 (non-blocking)             | Low      | Run as warning; ensure ≥ 90% pass; document gaps                                           |
| Test isolation between tenants                        | High     | Use dedicated test tenants; reset fixtures between runs; avoid shared seed data             |
| MFA testing requires interactive TOTP input           | Medium   | Pre-generate valid TOTP codes via test secret; automate TOTP generation in CI               |
| Break-glass test risks locking out admin              | High     | Run in drill mode only; use separate Shamir parts; ensure L2/L3 escalation path available   |
| Flaky negative tests due to timing                    | Medium   | Allow 1 automatic re-run per `REQ-NFR.SECURITY.authorization-regression-gates`; deterministic final status |
| Audit log verification latency                        | Low      | Poll audit store with short backoff; assert within 5s window                                |
| Service-account token expiry during test run          | Low      | Regenerate tokens with TTL ≥ test duration; use short scoped tokens                         |

---

### 8. Checklist

| Check                                                                                                              | Priority |
|--------------------------------------------------------------------------------------------------------------------|----------|
| Cross-tenant BOLA (type A): `tenant_A` user → `tenant_B` object returns 403 on all P0 endpoints                    | High     |
| Cross-object BOLA (type B): user without role → object in same tenant returns 403                                   | High     |
| BFLA (type C): `Viewer`/`Guest` → destructive/admin endpoint returns 403                                            | High     |
| IDOR (type D): guessable IDs (numeric+1, UUID variant) return 403, never 404/500                                    | High     |
| Only `Owner` can add/remove/change members on Project (`POST`/`PUT`/`DELETE /projects/{id}/members`)                | High     |
| Only `Owner` can change visibility (`PUT /projects/{id}/visibility`)                                                | High     |
| Only `Owner` can manage ABAC policies (`POST`/`DELETE /projects/{id}/policies`)                                     | High     |
| `Maintainer`/`Editor`/`Viewer` → membership/visibility/policies endpoints return 403                               | High     |
| Role inheritance: max-role-wins across Group/Subgroup/Project                                                       | High     |
| `Private` Project: non-member gets 403; `Internal`: unauthenticated gets 401; `Public`: anonymous gets 200          | High     |
| Fork: read access sufficient for public/internal; Guest can fork public; private without access → 403 (not 404)     | High     |
| GraphQL mutations with `ontologyId` enforce BOLA (cross-tenant, cross-object)                                       | High     |
| GraphQL queries with `ontologyId` enforce BOLA                                                                      | High     |
| WebSocket `JoinRoom` enforces `ontologyId` membership                                                               | High     |
| Write endpoints emit `audit_events` with `event`, `user_id`, `object_id`, `trace_id`, `timestamp`                   | High     |
| MFA category A (tenant purge, backup delete, secure erase): non-cacheable, required                                  | High     |
| MFA category B (production restore, migration rollback): non-cacheable, required                                     | High     |
| Service-account cannot perform category A/B operations                                                              | High     |
| Break-glass `/emergency/login`: works with Keycloak down; TTL ≤ 30 min; immutable audit; security-team notification  | High     |
| Idempotency-Key missing on members/visibility/policies → 400 `INVALID_IDEMPOTENCY_KEY`                             | High     |
| ABAC most-specific-wins: specific class/property policy overrides Project-level permission                          | Medium   |
| OMR `write_with_approval`: Editor cannot commit directly to protected `main`; proposal branch + Maintainer review    | Medium   |
| JIT/PAM: no standing access; TTL ≤ 60 min; 100% session recording                                                   | Medium   |
| Destructive guardrails G1–G4: correct confirmation level per command                                                 | Medium   |
| Public Browse API: no auth required; rate limiting active; no arbitrary Cypher/SPARQL; isolated Neo4j                | Medium   |
| Legacy scope alias `"ontology/" + id`: accepted only in migration window; rejected after cleanup                     | Low      |
| GraphQL serves only graph navigation (class/property/individual); non-graph queries rejected/redirected to REST      | Low      |
| Pagination: `per_page` ≤ 100 enforced; cursor-based pagination intact                                                | Low      |

---

### 9. Test Execution Matrix (Endpoint × Test Type)

| Endpoint class                                      | A (cross-tenant) | B (cross-object) | C (BFLA) | D (IDOR) | Priority |
|-----------------------------------------------------|:-:|:-:|:-:|:-:|:--:|
| `/api/v1/groups/{id}`                               | ✓ | ✓ | ✓ | ✓ | P0  |
| `/api/v1/projects/{id}`                             | ✓ | ✓ | ✓ | ✓ | P0  |
| `/api/v1/projects/{id}/members`                     | ✓ | ✓ | ✓ | — | P0  |
| `/api/v1/projects/{id}/visibility`                  | ✓ | ✓ | ✓ | — | P0  |
| `/api/v1/projects/{id}/policies`                    | ✓ | ✓ | ✓ | ✓ | P0  |
| `/api/v1/projects/{id}/fork`                        | ✓ | ✓ | — | ✓ | P0  |
| `/api/v1/ontologies/{ontologyId}`                   | ✓ | ✓ | ✓ | ✓ | P0  |
| `/api/v1/branches/{branchId}`                       | ✓ | ✓ | ✓ | ✓ | P0  |
| `/api/v1/commits/{commitId}`                        | ✓ | ✓ | ✓ | ✓ | P0  |
| GraphQL mutations (with `ontologyId`)               | ✓ | ✓ | ✓ | — | P0  |
| GraphQL queries (with `ontologyId`)                 | ✓ | ✓ | — | — | P0  |
| WebSocket `JoinRoom`                                | ✓ | ✓ | — | — | P0  |
| `/api/v1/sparql` (with `default-graph-uri`)         | ✓ | ✓ | — | — | P1  |
| `/api/v1/backups/*` (admin)                         | ✓ | — | ✓ | ✓ | P0  |

Legend: ✓ = test required; — = not applicable

---

### 10. Role Permission Matrix (Expected Behavior)

| Operation                                   | Viewer (10) | Reporter (20) | Editor (30) | Maintainer (40) | Owner (50) | Outsider |
|---------------------------------------------|:-----------:|:-------------:|:-----------:|:---------------:|:----------:|:--------:|
| View project/ontology (Private)             |     ✅      |      ✅       |     ✅      |       ✅        |     ✅     |   ❌ 403 |
| View project/ontology (Internal)            |     ✅      |      ✅       |     ✅      |       ✅        |     ✅     |   ❌ 401 |
| View project/ontology (Public)              |     ✅      |      ✅       |     ✅      |       ✅        |     ✅     |   ✅ 200 |
| Edit ontology (write)                       |     ❌ 403  |     ❌ 403    |     ✅      |       ✅        |     ✅     |   ❌ 403 |
| Edit ontology (write_with_approval)         |     ❌ 403  |     ❌ 403    |  proposal   |     proposal    |     ✅     |   ❌ 403 |
| Merge OMR to `main`                         |     ❌ 403  |     ❌ 403    |     ❌ 403  |       ✅        |     ✅     |   ❌ 403 |
| Add/remove member                           |     ❌ 403  |     ❌ 403    |     ❌ 403  |       ❌ 403    |     ✅     |   ❌ 403 |
| Change role                                 |     ❌ 403  |     ❌ 403    |     ❌ 403  |       ❌ 403    |     ✅     |   ❌ 403 |
| Change visibility                           |     ❌ 403  |     ❌ 403    |     ❌ 403  |       ❌ 403    |     ✅     |   ❌ 403 |
| Manage ABAC policies                        |     ❌ 403  |     ❌ 403    |     ❌ 403  |       ❌ 403    |     ✅     |   ❌ 403 |
| Delete project                              |     ❌ 403  |     ❌ 403    |     ❌ 403  |       ❌ 403    |     ✅     |   ❌ 403 |
| Fork public project                         |     ✅      |      ✅       |     ✅      |       ✅        |     ✅     |   ✅ 201 |
| Fork private project (no access)            |     ❌ 403  |     ❌ 403    |     ❌ 403  |       ❌ 403    |     ❌ 403 |   ❌ 403 |

Legend: ✅ = allowed; ❌ = denied with specified status; `proposal` = goes through OMR workflow

---

### 11. References (Source Specifications)

| Spec ID / Path                                                        | What it defines                                              |
|-----------------------------------------------------------------------|--------------------------------------------------------------|
| `specs/adr/ADR-DES.SECURITY.gitlab-like-organization-model.md`        | Org model, 4 roles, inheritance, ABAC, protected main        |
| `specs/adr/ADR-DES.API.organization-rest-endpoints.md`                | Canonical REST paths, RBAC matrix, idempotency, audit        |
| `specs/requirements/REQ-NFR.SECURITY.organization-access-model.md`    | Group/Project/Ontology 1:1, roles, visibility, OMR workflow  |
| `specs/requirements/REQ-FUN.DATA.ontology-visibility-levels.md`       | Private/Internal/Public visibility levels                    |
| `specs/adr/ADR-IMPL.SECURITY.bola-bfla-negative-tests-mandate.md`     | 4 negative test types, P0 endpoint-classes, PASS/FAIL criteria |
| `specs/requirements/REQ-NFR.SECURITY.bola-bfla-negative-tests.md`     | YAML templates, error code matrix, CI automation             |
| `specs/requirements/REQ-NFR.SECURITY.authorization-regression-gates.md` | Blocking release gate, 100% coverage, signed bundles       |
| `specs/adr/ADR-DES.SECURITY.authorization-policy-gates-strategy.md`   | Policy-as-code gate strategy                                 |
| `specs/adr/ADR-DES.SECURITY.mfa-critical-ops-mandate.md`              | MFA categories A/B/C/D, service-account rules                |
| `specs/adr/ADR-DES.SECURITY.privileged-access-jit-pam-strategy.md`    | JIT/PAM, no standing access, TTL, session recording          |
| `specs/requirements/REQ-NFR.SECURITY.privileged-access-control.md`    | JIT params, break-glass double-approval                      |
| `specs/adr/ADR-DES.SECURITY.break-glass-access-strategy.md`           | Shamir M-of-N, `/emergency/login`, TTL, audit                |
| `specs/requirements/REQ-NFR.SECURITY.security-requirements.md`        | Emergency Admin details, Shamir storage                      |
| `specs/adr/ADR-DES.SECURITY.destructive-command-guardrails.md`        | G1–G4 guardrail levels                                       |
| `specs/adr/ADR-DES.SECURITY.public-ontology-access.md`                | Public Browse API, read-only, rate limiting                  |
| `specs/adr/ADR-DES.PROCESS.merge-request-strategy.md`                 | OMR workflow, semantic diff, Maintainer review               |
| `.ai-factory/references/gitlab-projects-groups-api.md`                | GitLab API baseline, access levels 5/10/15/20/25/30/40/50    |
