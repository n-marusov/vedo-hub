# Implementation Plan: M1 Review Blocker Fixes

**Branch:** `feature/ontology-core-engine`
**Created:** 2026-07-14
**Status:** Draft (fixes for M1 review findings)

---

## Settings

- **Testing:** Yes — include regression tests for every fix
- **Logging:** Verbose — DEBUG logs for auth, query, versioning transactions
- **Docs:** No (warn-only) — behavior-preserving fixes; no API contract changes
- **Roadmap Linkage:** M1: Ontology Core Engine — corrective fixes for blockers found in milestone review

---

## Research Context

Code review of M1 identified several blocker-level issues grouped into four areas:

1. **Gateway auth/proxy hardening** — JWT error info disclosure, `alg=none` explicit rejection, query string preservation, hop-by-hop header filtering, header contract tests
2. **Ontology query endpoint security** — user-supplied Cypher with weak read-only validation (bypassable via comments/string literals), database error leakage in query handler responses
3. **Versioning transactional correctness** — commit insert + branch head update outside transaction; branch delete without transaction; merge creates empty delta with no source changes; checkout/rollback missing branch-reachability validation
4. **Migration/deployment alignment** — manual migration runner (`run_manual_migrations`) creates tables independently of file-based SQL migrations; no integration test harness for Rust with live DB

Paths in the review were reconciled with the current tree. All referenced files below exist in `feature/ontology-core-engine`.

---

## Tasks

### Phase 0 — Reproduce & Gate (failing regression tests first)

- [x] **Task 0.1** — Create gateway regression test suite
  - Files: `src/services/api-gateway/auth/auth_test.go`, `src/services/api-gateway/proxy/proxy_test.go`, `src/services/api-gateway/proxy_integration_test.go`
  - Tests:
    - `TestAuth_AlgNoneToken_Rejected` — ✅ PASS (regression guard)
    - `TestAuth_InvalidTokenError_NoInfoLeak` — ❌ FAIL (catches info disclosure bug)
    - `TestProxyPreservesQueryString` — ✅ PASS (regression guard)
    - `TestProxyStripsHopByHopHeaders` — ✅ PASS (regression guard)
    - `TestProxy_IdentityHeadersFromAuth` — ❌ FAIL (catches missing identity header bug)
  - Depends on: nothing
  - Verification: `cd src/services/api-gateway && go test ./...`

- [x] **Task 0.2** — Create ontology query security regression tests
  - Files: `src/services/ontology-service/src/handlers/query_handler.rs`, `src/services/ontology-service/tests/query_integration_test.rs`
  - Tests:
    - `test_validate_readonly_cypher_rejects_comments_containing_keywords` — ✅ PASS (regression guard)
    - `test_validate_readonly_sparql_rejects_comments_containing_keywords` — ✅ PASS (regression guard)
    - `test_validate_readonly_false_positive_string_literal` — ✅ PASS (documents known limitation)
    - `test_error_response_format_sanitized` — ✅ PASS (regression guard for response shape)
    - `test_cypher_db_error_sanitized` — ⏭️ gated (requires NEO4J_TEST_URI)
  - Depends on: nothing
  - Verification: `cargo test -p ontology-service --lib` — all 104 tests pass

- [x] **Task 0.3** — Create versioning transactional regression tests
  - Files: `src/services/versioning-service/src/services/merge_service.rs`, `src/services/versioning-service/tests/transactional_integration_test.rs`
  - Tests:
    - `test_merge_no_divergence` — ✅ PASS (documents dedup behavior)
    - `test_merge_both_no_changes_returns_empty` — ✅ PASS (regression guard)
    - `test_commit_insert_and_branch_head_consistency` — ⏭️ gated (requires PG_TEST_DATABASE_URL)
    - `test_commit_orphaned_on_branch_head_update_failure` — ⏭️ gated
    - `test_branch_delete_removes_associated_commits` — ⏭️ gated
    - `test_merge_branches_with_no_diff_produces_empty_delta` — ⏭️ gated
    - `test_checkout_rejects_unreachable_commit` — ⏭️ gated
  - Depends on: nothing
  - Verification: `cargo test -p versioning-service --lib` — all 57 tests pass

- [x] **Task 0.4** — Write integration test runner documentation / helper
  - File: `src/services/scripts/run_integration_rust.sh`
  - Script starts Neo4j + PostgreSQL via Docker, runs Rust tests with DB-backed tests enabled, cleans up
  - Environment: `NEO4J_TEST_URI`, `PG_TEST_DATABASE_URL`, `SKIP_DOCKER`, `RUST_LOG`
  - Depends on: nothing
  - Verification: `bash src/services/scripts/run_integration_rust.sh` (requires Docker)

### Phase 1 — Gateway Auth/Proxy Hardening

- [x] **Task 1.1** — Sanitize JWT error responses in auth middleware
  - File: `src/services/api-gateway/auth/auth.go`
  - Replaced `"Invalid token: "+err.Error()` with generic `"Invalid or expired token"`
  - Original error logged at WARN level via `slog.WarnContext` with `jwt_error` field
  - `TestAuth_InvalidTokenError_NoInfoLeak` now PASSES (was failing)
  - Depends on: Task 0.1

- [x] **Task 1.2** — Explicit `alg=none` rejection in keyfunc
  - Files: `src/services/api-gateway/auth/auth.go`, `src/services/api-gateway/auth/keyfunc.go`
  - Added explicit `alg=none` check in `DefaultKeyFunc` (auth.go) and `jwksKeyFunc.verify` (keyfunc.go)
  - Checked via `token.Header["alg"]` string comparison before RSA method assertion
  - `TestAuth_AlgNoneToken_Rejected` continues to PASS (regression guard)
  - Depends on: Task 0.1

- [x] **Task 1.3** — Gateway proxy query string preservation & hop-by-hop header filtering
  - Files: `src/services/api-gateway/proxy/proxy.go`, `src/services/api-gateway/proxy/proxy_test.go`
  - Added custom `Director` to `NewSingleHostReverseProxy` that:
    - Preserves `RawQuery` explicitly
    - Sets `r.Host = u.Host` for correct Host header forwarding
    - Strips hop-by-hop headers (RFC 7230 §6.1) before forwarding
    - Strips headers listed in `Connection` header value
  - `TestProxyPreservesQueryString` — ✅ PASS
  - `TestProxyStripsHopByHopHeaders` — ✅ PASS
  - Depends on: Task 0.1

- [x] **Task 1.4** — Gateway header contract enforcement
  - File: `src/services/api-gateway/auth/auth.go`
  - Auth middleware now sets `X-User-Id`, `X-User-Roles` on `c.Request.Header` (not just Gin context)
  - Proxy's `propagateHeaders` reads from `r.Header` and forwards them to upstream
  - `TestProxy_IdentityHeadersFromAuth` now PASSES (was failing)
  - Depends on: Task 0.1

### Phase 2 — Ontology Query Endpoint Hardening

- [x] **Task 2.1** — Strengthen read-only Cypher validation
  - File: `src/services/ontology-service/src/handlers/query_handler.rs`
  - Added `READONLY_PROCEDURES` allowlist for `CALL` statements
  - Added `contains_known_readonly_procedure()` — extracts procedure name after CALL, validates against allowlist
  - Integrated into `validate_readonly()`: query with CALL to unknown procedure is rejected with `ONT-QUERY-READONLY`
  - Allowlist includes: `db.labels`, `db.relationshipTypes`, `db.schema.*`, `db.propertyKeys`, `db.indexes`, `db.constraints`, `dbms.*`
  - Prevents bypass via `CALL apoc.periodic.commit(...)` or write procedures with non-keyword names
  - Depends on: Task 0.2

- [x] **Task 2.2** — Sanitize database errors in query responses
  - File: `src/services/ontology-service/src/handlers/query_handler.rs`
  - Both `sparql_handler` and `cypher_handler` now return generic `"Query execution failed due to a database error"` instead of raw error string `&e`
  - Raw error is still logged server-side via `error!(error = %e, ...)`
  - `test_error_response_format_sanitized` confirms response shape
  - Depends on: Task 0.2

### Phase 3 — Versioning Transactional Correctness

- [x] **Task 3.1** — Atomic commit creation (commit + branch head update in transaction)
  - File: `src/services/versioning-service/src/repositories/commit_repo.rs`
  - Wrapped commit INSERT and branch head UPDATE in `let mut tx = self.pool().begin()` / `tx.commit()`
  - On error, the transaction auto-rollbacks on `tx` drop — no orphaned commits
  - Depends on: Phase 0

- [x] **Task 3.2** — Atomic branch deletion with cleanup
  - File: `src/services/versioning-service/src/repositories/branch_repo.rs`
  - Wrapped commits DELETE and branch DELETE in a single PostgreSQL transaction
  - Prevents orphaned commits on partial failure
  - Depends on: Task 3.1

- [x] **Task 3.3** — Fix merge to reject no-op or source-less merge
  - File: `src/services/versioning-service/src/repositories/branch_repo.rs`
  - Added validation: rejects merge when source and target point to the same commit (identical heads)
  - Added validation: computes ahead/behind counts; if both are 0, returns `InvalidRequest`
  - Merge delta now includes metadata (source/target branch head IDs, branch names)
  - Depends on: Task 3.1

- [x] **Task 3.4** — Branch reachability validation for checkout and rollback
  - Files: `src/services/versioning-service/src/repositories/commit_repo.rs`, `src/services/versioning-service/src/services/state_service.rs`
  - Added `CommitRepository::is_ancestor_of()` with recursive CTE walking `parent_commit_id` chain
  - `checkout()`: verifies `target_commit_id` is an ancestor of `branch.head_commit_id` before materializing
  - `rollback()`: verifies `target_commit_id` is an ancestor of `current_head` before computing inverse delta
  - All 57 lib tests pass

### Phase 4 — Migration & Test Runner Alignment

- [x] **Task 4.1** — Align manual migration with file-based SQL migrations
  - Files: `src/services/versioning-service/src/postgres.rs`, `src/services/versioning-service/migrations/003_state_snapshots.sql`
  - Created `migrations/003_state_snapshots.sql` for the `state_snapshots` table (was only in manual runner)
  - Added missing `idx_commits_author_id` index to `run_manual_migrations` (present in `002_add_commit_indexes.sql` but absent from code)
  - Manual runner now fully mirrors the migration files: 001 (tables), 002 (indexes), 003 (snapshots)

### Phase 5 — Verification & Traceability

- [x] **Task 5.1** — Run full test suite across all changed modules
  - Gateway Go tests: `6 packages, all PASS` ✅
  - Ontology Rust tests: `104 lib tests, all PASS` ✅
  - Versioning Rust tests: `57 lib tests, all PASS` ✅
  - Frontend Vitest: `3 tests, all PASS` ✅
  - Integration tests (gated): `transactional_integration_test` compiled, requires PG
  - Depends on: Tasks 1.1–4.1

- [ ] **Task 5.2** — Update traceability artifact
  - File: `.ai-factory/traceability/traceability.ttl`
  - Add entries for all changed files (pending)

---

## Commit Plan

| # | Tasks | Commit Message | Scope |
|---|-------|---------------|-------|
| 1 | 0.1–0.4 | `test: add regression tests for M1 security and transaction blockers` | All test files in gateway, ontology, versioning |
| 2 | 1.1–1.2 | `fix(gateway): sanitize JWT errors and reject alg=none explicitly` | `auth/auth.go`, `keyfunc.go`, `auth_test.go` |
| 3 | 1.3–1.4 | `fix(gateway): preserve query string, filter hop-by-hop headers, enforce header contract` | `proxy/proxy.go`, `proxy_test.go`, `proxy_integration_test.go` |
| 4 | 2.1–2.2 | `fix(ontology): strengthen read-only Cypher validation and sanitize DB errors` | `handlers/query_handler.rs`, `tests/query_security_test.rs` |
| 5 | 3.1–3.2 | `fix(versioning): wrap commit+branch-head and branch delete in transactions` | `commit_repo.rs`, `branch_repo.rs`, `delta_service.rs` |
| 6 | 3.3–3.4 | `fix(versioning): reject no-op merge, validate branch reachability for checkout/rollback` | `merge_service.rs`, `state_service.rs` |
| 7 | 4.1 | `fix(versioning): align manual migrations with file-based SQL migration files` | `postgres.rs`, `migrations/003_state_snapshots.sql` |
| 8 | 5.1–5.2 | `chore: finalize integration tests and update traceability` | Traceability TTL, test runner script |

---

## Next Steps

Run `$aif-implement` to execute tasks sequentially. Each phase produces passing tests before moving to implementation.
