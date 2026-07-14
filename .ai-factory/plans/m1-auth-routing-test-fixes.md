# Implementation Plan: M1 Auth, Routing & Test Fixes

Branch: feature/ontology-core-engine
Created: 2026-07-14
Type: Fix

## Settings
- Testing: Yes — unit tests per task, integration tests for gateway/ontology/versioning routes
- Logging: Verbose — DEBUG-level logs for development, INFO for key events, ERROR for failures
- Docs: No — WARN [docs] only, no mandatory checkpoint (these are bugfixes within M1)

## Roadmap Linkage
Milestone: "M1: Ontology Core Engine"
Rationale: Critical fixes for production auth, routing, query endpoints, and test suite reliability that block M1 from working end-to-end.

## Research Context

Source: Code review of M1 — `feature/ontology-core-engine` branch (116 files, 25k+ LOC changes)

### Root Cause Summary

The M1 code review identified **6 critical blockers** and **4 suggestions** across production configuration, API routing, and test infrastructure:

**Critical Issues:**
1. **Production auth disabled** — API Gateway `main.go` sets `KeyFunc: nil`, making real JWT verification impossible; tests use a test-only keyfunc
2. **Read-only query endpoints require Editor role** — SPARQL/CYPHER/GraphQL POST endpoints need role=0 override in production config but only have it in test helpers
3. **Missing upstream SPARQL/CYPHER endpoints** — Gateway proxies to ontology-service which has no matching routes → 404
4. **Axum 0.7 route syntax incompatible** — ontology-service uses `{param}` but axum 0.7 requires `:param` → parameterized routes return 404
5. **Same route syntax issue in versioning-service** — commit/branch detail, checkout, rollback, delete, switch all affected
6. **Integration tests not properly skipped** — `skip_if_no_*` helpers only log but tests proceed into DB-backed code and crash when env vars are unset

**Suggestions:**
7. Test fixtures use different Neo4j property names (`class_id`, `SUBCLASS_OF`) than production code (`id`, `CHILD_OF`, `INSTANCE_OF`)
8. Frontend has no `test` script — planned TDD workflow not executable
9. Query mutation detection uses naive substring matching — false positives possible
10. Gateway proxies body buffering closes `r.Body` before upstream reads it — potential data loss

## Commit Plan

- **Commit 1** (after Tasks 1-2): `fix(gateway): wire production JWT verification and read-query role overrides`
- **Commit 2** (after Tasks 3-4): `fix(ontology): add SPARQL/CYPHER endpoints and fix axum route syntax`
- **Commit 3** (after Task 5): `fix(versioning): fix axum route syntax`
- **Commit 4** (after Tasks 6-7): `fix(tests): proper integration test gating and fixture alignment`
- **Commit 5** (after Tasks 8-10): `fix(frontend): add test script and misc query/fixes`

## Tasks

### Phase 1: Production Auth Configuration (Critical)

- [ ] **Task 1: Wire production JWT verification in API Gateway**

    **Problem:** `src/services/api-gateway/main.go:72` sets `KeyFunc: nil`, making all protected routes unreachable with any JWT in production. Tests in `helpers_test.go` inject a valid keyfunc, masking the issue.

    **Fix:**
    - Add startup-phase JWKS loading: configure `KEYCLOAK_JWKS_URL` env var or `JWT_PUBLIC_KEY_PATH` env var (PEM file)
    - Move RSA public key loading + `jwt.Keyfunc` construction before router setup
    - Fail startup (`panic` / `log.Fatal`) when JWT verification is not configured, preventing deployment with disabled auth
    - Preserve exemption paths for `/health`, `/ready`, `/metrics`, `/api/v1/public/*`
    - Update `main_test.go` or add a startup test that verifies auth middleware panics/rejects when KeyFunc is nil

    **Specs references:**
    - REQ-NFR.SECURITY.authorization-regression-gates
    - US-api.auth.jwt
    - ADR-DES.SECURITY.gitlab-like-organization-model

    **Logging:** INFO for KeyFunc initialization with JWKS URL or key source; ERROR and startup abort if neither configured.

    **Files:**
    - Update: `src/services/api-gateway/main.go`
    - Create: `src/services/api-gateway/main_test.go` (startup wiring test)
    - Update: `.env.example` (add `KEYCLOAK_JWKS_URL` / `JWT_PUBLIC_KEY_PATH`)

    Dependencies: None

- [ ] **Task 2: Add production RequiredRoleLevel overrides for read-only query endpoints**

    **Problem:** Production `auth.Config` in `main.go` does not set `RequiredRoleLevel`, so `methodRequiredLevel` treats every POST as Editor-level (role weight ≥1). The test environment in `helpers_test.go` adds overrides for `/api/v1/sparql`, `/api/v1/cypher`, and `/api/v1/graphql` — masking the production behavior.

    **Fix:** Add the same RequiredRoleLevel overrides in production `main.go`:
    ```go
    RequiredRoleLevel: map[string]int{
        "POST:/api/v1/sparql":  0,
        "POST:/api/v1/cypher":  0,
        "POST:/api/v1/graphql": 0,
    },
    ```
    Verify with a test that a Viewer-role JWT can POST to these endpoints.

    **Specs references:**
    - REQ-NFR.SECURITY.bola-bfla-negative-tests (BFLA enforcement must allow Viewer reads)
    - ADR-DES.SECURITY.gitlab-like-organization-model

    **Logging:** INFO on auth config initialization with role level overrides summary; WARN if overrides are empty.

    **Files:**
    - Update: `src/services/api-gateway/main.go`
    - Update: `src/services/api-gateway/auth/auth_test.go` (add test for Reader access to query endpoints)

    Dependencies: Task 1 (same auth config block)

### Phase 2: Ontology-Service Routing Fixes (Critical)

- [ ] **Task 3: Implement SPARQL and CYPHER query upstream endpoints**

    **Problem:** Gateway exposes `POST /api/v1/sparql` and `POST /api/v1/cypher` (Task 6.2 in the main plan), which proxy to the ontology-service, but ontology-service has no matching routes. Valid queries reach the proxy and get 404.

    **Fix:**
    - Add SPARQL endpoint handler in ontology-service: parse SPARQL SELECT query, execute via Neo4j Cypher translation proxy or direct Neo4j query adaptation
    - Add CYPHER endpoint handler: execute MATCH queries directly against Neo4j pool with read-only enforcement, configurable LIMIT, timeout
    - Both return `{results, execution_time_ms, triple_count}` format
    - Register routes at `/api/v1/sparql` and `/api/v1/cypher` in `build_app()`
    - Add unit tests: valid query → 200, mutation → 400, missing auth → 401, malformed → 400
    - Add integration test verifying both routes resolve through the real ontology-service router

    **Specs references:**
    - US-api.sparql.execute [NEW]
    - US-api.cypher.execute [NEW]
    - REQ-FUN.API.graphql-sparql
    - REQ-FUN.API.cypher-query-language
    - ADR-DES.API.sparql-query-language-strategy
    - ADR-DES.API.cypher-query-language-adoption

    **Logging:** DEBUG per query with hash/params; INFO for completed queries with result count + duration; WARN for timeouts; ERROR for Neo4j query failures.

    **Files:**
    - Create: `src/services/ontology-service/src/handlers/query_handler.rs`
    - Create: `src/services/ontology-service/tests/query_integration_test.rs`
    - Update: `src/services/ontology-service/src/lib.rs` (add query routes)

    Dependencies: Task 1.2 (Neo4j driver, already done in M1)

- [ ] **Task 4: Fix axum 0.7 route syntax in ontology-service**

    **Problem:** All parameterized routes in `ontology-service/src/lib.rs` use `{param}` syntax (`/api/v1/ontologies/{ontology_id}/classes/{class_id}`). The workspace pins `axum = "0.7"`, which requires `:param` syntax (`/api/v1/ontologies/:ontology_id/classes/:class_id`). Parameterized routes return 404, breaking all class CRUD, property CRUD, individual CRUD, graph queries, export, and import.

    **Fix:** Convert all `{param}` to `:param` in route definitions. Verify no axum breaking changes by running:
    ```bash
    cargo test -p ontology-service --lib -- tests::test_root_returns_service_info
    ```
    Then run full unit test suite and integration tests (with NEO4J_TEST_URI set).

    Affected routes (in `src/services/ontology-service/src/lib.rs`):
    - `/api/v1/ontologies/{ontology_id}/classes` → `/:ontology_id/classes`
    - `/api/v1/ontologies/{ontology_id}/classes/{class_id}` → `/:ontology_id/classes/:class_id`
    - `/api/v1/ontologies/{ontology_id}/classes/root` → `/:ontology_id/classes/root`
    - `/api/v1/ontologies/{ontology_id}/classes/search/autocomplete` → `/:ontology_id/classes/search/autocomplete`
    - `/api/v1/ontologies/{ontology_id}/classes/{class_id}/children` → `/:ontology_id/classes/:class_id/children`
    - `/api/v1/ontologies/{ontology_id}/classes/{class_id}/ancestors` → `/:ontology_id/classes/:class_id/ancestors`
    - `/api/v1/ontologies/{ontology_id}/classes/{class_id}/descendants` → `/:ontology_id/classes/:class_id/descendants`
    - `/api/v1/ontologies/{ontology_id}/classes/{class_id}/breadcrumb` → `/:ontology_id/classes/:class_id/breadcrumb`
    - `/api/v1/ontologies/{ontology_id}/classes/{class_id}/neighborhood` → `/:ontology_id/classes/:class_id/neighborhood`
    - `/api/v1/ontologies/{ontology_id}/properties` → `/:ontology_id/properties`
    - `/api/v1/ontologies/{ontology_id}/properties/{property_id}` → `/:ontology_id/properties/:property_id`
    - `/api/v1/ontologies/{ontology_id}/individuals` → `/:ontology_id/individuals`
    - `/api/v1/ontologies/{ontology_id}/individuals/{individual_id}` → `/:ontology_id/individuals/:individual_id`
    - `/api/v1/ontologies/{ontology_id}/export` → `/:ontology_id/export`
    - `/api/v1/ontologies/{ontology_id}/import` → `/:ontology_id/import`
    - `/api/v1/graphql` (no change, no params)

    Also add route-registration integration test that exercises each route and verifies it doesn't return 404, without requiring Neo4j.

    **Specs references:**
    - ADR-IMPL.STACK.ontology-rust-strategy (axum for HTTP)
    - All M1 REST endpoints (F1, F2, F4, F8)

    **Logging:** DEBUG for each registered route prefix on startup; INFO confirming route count.

    **Files:**
    - Update: `src/services/ontology-service/src/lib.rs` (all route strings)
    - Create: `src/services/ontology-service/tests/route_registration_test.rs`
    - Update: `src/services/ontology-service/src/handlers/mod.rs` (verify handler signatures)

    Dependencies: Task 3 (same file, apply both changes together)

### Phase 3: Versioning-Service Routing Fixes (Critical)

- [ ] **Task 5: Fix axum 0.7 route syntax in versioning-service**

    **Problem:** Same issue as Task 4 — parameterized routes in `versioning-service/src/routes.rs` use `{param}` syntax. The following endpoints fail:
    - `GET /api/v1/versioning/commits/{id}`
    - `GET /api/v1/versioning/commits/{id}/delta`
    - `POST /api/v1/versioning/commits/{id}/checkout`
    - `POST /api/v1/versioning/commits/{id}/rollback`
    - `GET /api/v1/versioning/branches/{id}`
    - `DELETE /api/v1/versioning/branches/{id}`
    - `POST /api/v1/versioning/branches/{id}/switch`
    - `POST /api/v1/versioning/branches/merge`

    **Fix:** Convert all `{param}` to `:param`. Add route-registration test verifying each route.

    **Specs references:**
    - ADR-IMPL.STACK.version-control-rust-strategy
    - US-git.commits.create, US-git.branches.create-switch, US-git.commits.rollback, etc.

    **Logging:** DEBUG for route registration.

    **Files:**
    - Update: `src/services/versioning-service/src/routes.rs`
    - Create: `src/services/versioning-service/tests/route_registration_test.rs`

    Dependencies: None (independent of other tasks)

### Phase 4: Test Infrastructure Fixes (Critical)

- [ ] **Task 6: Fix integration test gating to properly skip when DB unavailable**

    **Problem:** `skip_if_no_neo4j()` and `skip_if_no_pg()` only print a message and return — tests continue into DB-backed code, which panics when Neo4j/PostgreSQL is not available. This makes `cargo test --workspace` fail in any developer environment without test databases.

    **Fix:** Replace `skip_if_no_*` with proper test-skipping mechanisms. Two options (choose one):
    
    **Option A (recommended):** Replace `skip_if_no_*` with a `#[ignore]` gate and a helper script:
    ```rust
    // In tests/common/mod.rs
    pub fn require_neo4j() {
        if !is_integration_enabled() {
            eprintln!("Skipping: set NEO4J_TEST_URI to run Neo4j integration tests");
            std::process::exit(0); // exit successfully — test infrastructure counts this
        }
    }
    ```
    And a shell script for running all integration tests:
    ```bash
    # tests/run_integration.sh
    cargo test -p ontology-service --test '*' -- --test-threads=1
    cargo test -p versioning-service --test '*' -- --test-threads=1
    ```
    
    **Option B:** Gate with a feature flag in Cargo.toml:
    ```toml
    [features]
    integration-tests = []
    ```
    And annotate integration test files with `#[cfg(feature = "integration-tests")]`.

    Apply the same fix to both ontology-service and versioning-service integration tests.

    **Test:** `cargo test --workspace` must pass without NEO4J_TEST_URI or PG_TEST_DATABASE_URL.

    **Logging:** DEBUG for test setup; INFO on test enter; WARN when skipping due to missing DB.

    **Files:**
    - Update: `src/services/ontology-service/tests/common/mod.rs`
    - Update: `src/services/versioning-service/tests/common/mod.rs`
    - Create: `tests/run_integration_rust.sh` (optional, convenience)

    Dependencies: None

- [ ] **Task 7: Align integration test fixtures with production Neo4j schema**

    **Problem:** Integration tests use a different Neo4j property naming scheme than production code:
    - Tests use `class_id`, `individual_id`, `SUBCLASS_OF`
    - Production uses `id`, `INSTANCE_OF`, `CHILD_OF`
    
    The tests seed data with wrong property names and verify with wrong query patterns, making them useless as behavioral specs even after route syntax is fixed.

    **Fix:** Rewrite all seed data and assertion queries in integration test files to match the production schema:
    - `class_integration_test.rs` — use `id` instead of `class_id`, `CHILD_OF` instead of `SUBCLASS_OF`
    - `graph_query_integration_test.rs` — same alignment
    - `individual_integration_test.rs` — use `id`, `INSTANCE_OF` as in production
    - `import_export_integration_test.rs` — align seed queries
    - `property_integration_test.rs` — align seed queries

    For each file, update both the `CREATE` seed queries and the `MATCH` verification queries.

    **Test:** Run with `NEO4J_TEST_URI` set — all integration tests must pass:
    ```bash
    cargo test -p ontology-service --test class_integration_test -- --nocapture
    cargo test -p ontology-service --test graph_query_integration_test -- --nocapture
    # etc.
    ```

    **Logging:** N/A (test infrastructure change)

    **Files:**
    - Update: `src/services/ontology-service/tests/class_integration_test.rs`
    - Update: `src/services/ontology-service/tests/graph_query_integration_test.rs`
    - Update: `src/services/ontology-service/tests/individual_integration_test.rs`
    - Update: `src/services/ontology-service/tests/import_export_integration_test.rs`
    - Update: `src/services/ontology-service/tests/property_integration_test.rs`

    Dependencies: Task 4 (routes must work for tests to pass), Task 6 (test gating)

### Phase 5: Quality Improvements (Suggestions)

- [ ] **Task 8: Add frontend test script and verify type/lint pass**

    **Problem:** Frontend has no `test` script, so the planned TDD workflow for frontend is not executable. Available scripts are `typecheck`, `lint`, `lint:ci`.

    **Fix:** Add a `test` script using Vitest (already compatible with Vite used in this project):
    ```json
    // package.json scripts
    "test": "vitest run",
    "test:watch": "vitest"
    ```
    Create a basic Vitest config (`vitest.config.ts`) and placeholder test file to verify the setup works. Add documentation comment explaining the test setup.

    **Logging:** N/A

    **Files:**
    - Update: `src/services/frontend/package.json`
    - Create: `src/services/frontend/vitest.config.ts`

    Dependencies: None

- [ ] **Task 9: Harden query mutation detection in API Gateway**

    **Problem:** The SPARQL/CYPHER query validator uses naive `strings.Contains()` checks for mutation keywords like `INSERT`, `DELETE`, `SET`, `CREATE`, `MERGE`. This produces false positives when keywords appear inside string literals, comments, or identifiers (e.g., a class named `CreateResource`).

    **Fix:** Enhance both `ValidateAndSanitizeSPARQL` and `ValidateAndSanitizeCYPHER` to use word-boundary regex matching instead of substring `Contains`. Add a helper function:
    ```go
    // containsKeyword(s, keyword) checks for keyword as a whole word (not substring).
    func containsKeyword(s, keyword string) bool {
        re := regexp.MustCompile(`\b` + regexp.QuoteMeta(keyword) + `\b`)
        return re.MatchString(s)
    }
    ```
    Update the mutation keyword loops to use `containsKeyword`. Add test cases:
    - A class named `CreateResource` in a MATCH query → allowed
    - `INSERT DATA { ... }` → rejected
    - `MATCH (n:DeleteMe)` → allowed

    **Logging:** DEBUG for match details; WARN for rejected queries with `keyword_match` field in structured log.

    **Files:**
    - Update: `src/services/api-gateway/services/query_service.go`
    - Update: `src/services/api-gateway/services/query_service_test.go`

    Dependencies: None

- [ ] **Task 10: Fix proxy body buffering to avoid early body close**

    **Problem:** In `proxy.go`, the `ServeHTTP` method reads the request body, buffers it, then calls `r.Body.Close()`. The `httputil.ReverseProxy` may still try to read from `r.Body` after this point, resulting in data loss for write endpoints (POST/PUT).

    **Fix:** Replace the manual `Close()` with a buffer-swap pattern that leaves the new reader open:
    ```go
    if r.Body != nil && r.ContentLength != 0 {
        buf, err := io.ReadAll(r.Body)
        if err == nil {
            r.Body = io.NopCloser(bytes.NewReader(buf))
        }
        // Do NOT close r.Body here — the ReverseProxy owns the lifecycle
    }
    ```
    Remove the `_ = r.Body.Close()` line.

    Add a test that verifies body content is forwarded correctly for POST requests with bodies.

    **Logging:** DEBUG for body buffering with content length; WARN if body read fails.

    **Files:**
    - Update: `src/services/api-gateway/proxy/proxy.go`
    - Update: `src/services/api-gateway/proxy/proxy_test.go`

    Dependencies: None

## Progress Tracking

```
Total: 10 tasks
├── Phase 1: Production Auth         [ ] 0/2 — JWT verification, role overrides
├── Phase 2: Ontology Routing        [ ] 0/2 — SPARQL/CYPHER upstream, route syntax fix
├── Phase 3: Versioning Routing      [ ] 0/1 — axum route syntax fix
├── Phase 4: Test Infrastructure     [ ] 0/2 — test gating, fixture alignment
└── Phase 5: Quality Improvements    [ ] 0/3 — frontend test, query detection, proxy body
```

## Next Steps

Plan created with **10 tasks** across **5 phases**.

Plan file: `.ai-factory/plans/m1-auth-routing-test-fixes.md`

To start implementation, run:
```
$aif-implement
```

To view tasks:
```
/tasks (or use TaskList)
```
