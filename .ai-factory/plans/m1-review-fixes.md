# Plan: M1 Review Fixes

> Fixes for critical issues found during M1 (Ontology Core Engine) code review.

**Branch:** `feature/ontology-core-engine`  
**Created:** 2026-07-14  
**Type:** Fix

## Settings

| Setting | Value |
|---------|-------|
| Testing | Yes — include tests for each fix |
| Logging | Verbose — DEBUG-level logs for development |
| Documentation | Yes — mandatory docs checkpoint at completion |
| Roadmap Linkage | M1: Ontology Core Engine — fixes for completed milestone |

## Research Context

### Issues from M1 Code Review

The code review identified 6 critical issues and 4 suggestions across observability, security, and performance:

**Critical Issues (6):**
1. OpenTelemetry tracing has no OTLP exporter — traces never reach collector
2. Prometheus scrape targets point to wrong service ports — monitoring broken
3. API Gateway `/metrics` endpoint has hardcoded counter — always returns 0
4. Raw database error messages leaked in HTTP responses — information disclosure
5. Merge conflict detection uses O(n²) `Vec::contains` — performance regression
6. State materialization replays from root commit on every checkout — O(n) per operation
7. Metrics test uses global `AtomicU64` without isolation — flaky tests

**Suggestions (4):**
8. API Gateway proxy missing `traceparent` header propagation
9. Versioning service missing composite index on `(branch_id, created_at)`
10. Empty stub tests in test scripts — no real assertions
11. `Shared/lib.rs` does not re-export `tracing` module (inconsistent with health/metrics)

## Tasks

### Phase 1: Observability Fixes (Critical)

These three issues block production monitoring: tracing, Prometheus scraping, and gateway metrics.

#### ✅ Task 1.1: Add OTLP exporter to shared tracing module

**Files:**
- `src/services/Cargo.toml` — add `opentelemetry`, `opentelemetry-otlp`, `opentelemetry_sdk`, `tracing-opentelemetry` workspace dependencies
- `src/services/shared/Cargo.toml` — add `opentelemetry`, `opentelemetry-otlp`, `opentelemetry_sdk`, `tracing-opentelemetry` from workspace
- `src/services/shared/src/tracing.rs` — add OTLP exporter pipeline alongside JSON formatter

**Deliverable:**
- `init_tracing()` configures both JSON logging output AND OTLP gRPC export to `OTEL_EXPORTER_OTLP_ENDPOINT` (default: `http://otel-collector:4317`)
- On OTLP connection failure, log a WARN and continue with JSON-only logging (graceful degradation)
- `tracing_subscriber::Registry` layered with `tracing_opentelemetry::layer` + JSON fmt layer
- Test: verify `init_tracing()` doesn't panic when OTLP collector is unreachable

**Logging:**
- INFO: "OpenTelemetry initialized, exporting to {endpoint}"
- WARN: "OTLP exporter failed to connect, falling back to JSON logging only" (non-fatal)
- DEBUG: span events during layer registration

**Blocked by:** Nothing

---

#### ✅ Task 1.2: Fix Prometheus scrape target ports

**Files:**
- `deploy/observability/prometheus/prometheus.yml` — correct port numbers

**Deliverable:**
- `ontology-service:8082` (was `:8081`, port defined in `ontology-service/src/lib.rs:27`)
- `versioning-service:8083` (was `:8082`, port defined in `versioning-service/src/lib.rs:21`)
- `auth-service:8081` (was `:8084`)
- `metrics-service:8084` (was `:8000`)

**Verification:**
```bash
docker compose -f deploy/docker-compose.yml config | grep -A5 "ontology-service" | grep ports
grep -n "8082\|8083\|8081\|8084" deploy/observability/prometheus/prometheus.yml
```

**Logging:** N/A — config file change only

**Blocked by:** Nothing

---

#### ✅ Task 1.3: Replace hardcoded API Gateway metrics counter with real promhttp collector

**Files:**
- `src/services/api-gateway/main.go` — replace inline `/metrics` handler with promhttp
- `src/services/api-gateway/go.mod` — add `github.com/prometheus/client_golang` dependency
- `src/services/api-gateway/middleware/` — add metrics middleware that increments request counter

**Deliverable:**
- Import `github.com/prometheus/client_golang/prometheus` and `promhttp`
- Register a `prometheus.NewCounterVec` for `vedo_service_requests_total` with `service` and `method` labels
- Add middleware that increments the counter on every request
- Replace `r.GET("/metrics", ...)` with `gin.WrapH(promhttp.Handler())`
- Add a similar middleware in the Rust shared metrics module for consistency

**Logging:**
- DEBUG: "Metrics requested" — existing, keep
- DEBUG: "Request counted: {method} {path}" — new in middleware

**Blocked by:** Nothing

---

### Phase 2: Security Fixes (Critical)

#### ✅ Task 2.1: Sanitize database error messages in HTTP responses

**Files:**
- `src/services/ontology-service/src/error.rs` — sanitize `Database(String)` variant
- `src/services/versioning-service/src/error.rs` — sanitize `Database(String)` variant (already uses `Database(err.to_string())`)
- `src/services/ontology-service/src/error.rs` — update `IntoResponse` to log raw error, return generic message

**Deliverable:**
- `OntologyError::Database` stores the raw error internally (e.g., `#[source]`) but uses `tracing::error!()` for logging
- HTTP response for `Database` errors returns: `{"error": "ONT-DATABASE-ERROR", "detail": "Internal database error (trace_id: ...)"}` — no raw message leak
- Same for `VersionError::Database` in versioning-service (`error.rs:25`)
- Tests: verify `IntoResponse` for `Database("connection refused".into())` returns generic message without "connection refused"

**Logging:**
- ERROR: "Database error: {error} [trace_id={trace_id}]" — raw message logged server-side
- Response body: generic message with trace_id reference

**Blocked by:** Nothing

---

### Phase 3: Performance Fixes (Critical)

#### ✅ Task 3.1: Optimize merge conflict detection from O(n²) to O(n)

**Files:**
- `src/services/versioning-service/src/services/merge_service.rs`

**Deliverable:**
- Replace `Vec<String>` for `target_added_keys` and `target_removed_keys` with `HashSet<String>`
- Import `std::collections::HashSet`
- All `.contains()` lookups become O(1) instead of O(n)
- Maintain the same merge logic — this is a data structure change only

**Logging:** No new logging needed — existing WARNs for merge conflicts remain

**Blocked by:** Nothing

---

#### ✅ Task 3.2: Add materialized state snapshot table and caching strategy

**Files:**
- `src/services/versioning-service/src/migrations/003_add_state_snapshots.sql` — new migration
- `src/services/versioning-service/src/services/delta_service.rs` — add snapshot check before full replay
- `src/services/versioning-service/src/services/mod.rs` — register new migration
- `src/services/versioning-service/src/postgres.rs` — update migration counter

**Deliverable:**
- New PostgreSQL table: `state_snapshots(commit_id UUID PK, branch_id UUID, triples JSONB, created_at TIMESTAMPTZ)`
- In `materialize()`, first check if a snapshot exists for the target commit or a recent ancestor
- If a snapshot is found within 10 commits of the target, replay only the trailing deltas instead of the full chain
- Snapshot creation: auto-create every 50 commits during `create_commit` (best-effort, non-blocking)
- Migration `003` runs automatically via existing `run_manual_migrations()` mechanism
- Test: verify materialize uses snapshot when available and falls back to full replay when not

**Logging:**
- DEBUG: "Materialize: snapshot found at commit {id}, replaying {n} trailing deltas"  
- DEBUG: "Materialize: no snapshot found, replaying full chain ({n} commits)"  
- INFO: "State snapshot created at commit {id} ({n} triples)"

**Blocked by:** Nothing (architecture change, no code dependency)

---

### Phase 4: Test Quality & Stability

#### ✅ Task 4.1: Fix metrics test isolation (global AtomicU64)

**Files:**
- `src/services/shared/src/metrics.rs`

**Deliverable:**
- Option A: Add a `reset()` function to `AtomicU64` and call it in test setup via `#[ctor]` / `#[before]`
- Option B: Restructure as a `struct Meter` with an `AtomicU64` field, keep a `static METER: Meter` for production use, and allow tests to instantiate a fresh `Meter`
- Preferred: Option A is simpler — add `pub fn reset_request_count()` used only in tests:
  ```rust
  #[cfg(test)]
  pub fn reset_request_count() {
      REQUEST_TOTAL.store(0, Ordering::Relaxed);
  }
  ```
- Tests explicitly reset at the start of each test to ensure deterministic behavior
- Verify both `test_increment_and_read` and `test_concurrent_increment` pass regardless of order

**Logging:** N/A — test code only

**Blocked by:** Nothing

---

#### ✅ Task 4.2: Fill empty stub tests with real assertions

**Files:**
- `tests/test_build_commands.sh` — replace echo-only tests
- `tests/test_docker_build.sh` — replace echo-only tests
- `tests/test_health_metadata.sh` — add assertions for new services (ontology, versioning)
- `tests/test_ci_and_compose.sh` — verify the new service ports

**Deliverable:**
- `test_build_fails_on_error`: actually run `make build-fake` or verify exit code of a non-existent service target
- `test_lint_fails_on_error`: run `make lint-go` on a known-bad file and check exit code
- `test_test_fails_on_error`: run `make test-go --dry-run` or similar
- `test_compose_failed_error_path`: verify `docker compose config` returns non-zero for a deliberately broken compose snippet
- `test_port_conflict_error_path`: verify port allocations in compose yaml
- `test_health_metadata.sh`: add assertions for ontology-service (`:8082`) and versioning-service (`:8083`)

**Logging:** N/A — test scripts

**Blocked by:** Phase 1, Phase 2 (services must be properly instrumented before health tests are meaningful)

---

### Phase 5: Non-Blocking Improvements

#### ✅ Task 5.1: Propagate trace context headers through API Gateway proxy

**Files:**
- `src/services/api-gateway/proxy/proxy.go` — add header propagation in `ServeHTTP`

**Deliverable:**
- In `Proxy.ServeHTTP()`, propagate `X-Trace-Id`, `traceparent`, and `tracestate` headers from incoming request to upstream
- Add header propagation in `propagateHeaders()` or directly in `ServeHTTP` after circuit breaker check
- This enables distributed tracing: gateway → ontology-service / versioning-service share the same trace context

**Logging:**
- DEBUG: "Propagating trace context: trace_id={id}, traceparent={tp}" (when present)

**Blocked by:** Task 1.1 (OTLP exporter makes trace context useful)

---

#### ✅ Task 5.2: Add composite index for commit listing performance

**Files:**
- `src/services/versioning-service/src/migrations/002_add_commit_indexes.sql` — add or extend with composite index
- OR create `003` migration alongside Task 3.2

**Deliverable:**
- Add composite index `idx_commits_branch_created` on `commits(branch_id, created_at DESC)`
- This accelerates `list_commits_handler` which filters by `branch_id` and orders by `created_at`

**Logging:** N/A

**Blocked by:** Nothing

---

#### ✅ Task 5.3: Re-export tracing module in shared lib.rs

**Files:**
- `src/services/shared/src/lib.rs`

**Deliverable:**
- Add `pub use tracing::*;` to `lib.rs` so consumers can write `vedo_shared::init_tracing()` instead of `vedo_shared::tracing::init_tracing()`
- Test: compile check — `cargo check -p vedo-shared`

**Logging:** N/A

**Blocked by:** Nothing

---

## Commit Plan

| # | Commit Message | Tasks | Scope |
|---|---------------|-------|-------|
| 1 | `fix(observability): add OTLP exporter to shared tracing module` | 1.1, 5.1 | Rust shared lib + Gateway |
| 2 | `fix(deploy): correct Prometheus scrape target ports` | 1.2 | YAML config |
| 3 | `fix(gateway): replace hardcoded metrics with real promhttp counter` | 1.3 | Go gateway |
| 4 | `fix(security): sanitize database error messages in HTTP responses` | 2.1 | Rust errors |
| 5 | `fix(performance): optimize merge conflict O(n²) → O(n)` | 3.1 | Rust merge |
| 6 | `fix(performance): add materialized state snapshots for fast checkout` | 3.2 | Rust versioning |
| 7 | `fix(tests): stabilize metrics tests and fill empty stub assertions` | 4.1, 4.2 | Rust + Shell tests |
| 8 | `chore: add composite index and fix tracing re-export` | 5.2, 5.3 | Rust DB + shared |
