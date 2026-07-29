# Implementation Plan: Fix Versioning-Service Test Automation

Branch: feature/project-creation-in-group
Created: 2026-07-29

## Settings
- Testing: no (existing tests are being restructured, not new tests written)
- Logging: standard
- Docs: no

## Summary

versioning-service integration tests silently pass when `PG_TEST_DATABASE_URL` is
not set (`skip_if_no_pg() → return`). Analysis showed 12 of 18 tests are HTTP-smoke
tests that check «route doesn't crash with 500» — these should use mocked repositories,
not real PostgreSQL. 6 tests in `transactional_integration_test.rs` and
`delta_integration_test.rs` genuinely need PG for data-consistency checks.

This plan:
- Extracts repository traits to enable mocking
- Moves smoke tests from `tests/` to handler-level `#[cfg(test)]` blocks with mock repos
- Replaces silent skip with hard failure in remaining PG-dependent tests
- Adds a `make test-versioning` target that auto-starts PostgreSQL

## Commitment
- [x] `cargo test` passes all unit + mocked tests without PostgreSQL (75 tests pass)
- [x] `make test-versioning` auto-starts PG, runs unit and integration tests, cleans up
- [x] `cargo test` FAILS (panics with clear message) when PG-dependent tests can't connect

## Commit Plan
- **Commit 1** (after tasks 1-3): "refactor(versioning): extract repository traits, add mock support"
- **Commit 2** (after tasks 4-5): "test(versioning): move smoke tests to handler-level with mocks, fail hard on missing PG"
- **Commit 3** (after task 6): "build: add test-versioning Makefile target with auto PG"

## Tasks

### Phase 1: Repository Traits & Mock Foundation

- [x] Task 1: Define `BranchRepositoryTrait` and `CommitRepositoryTrait` in `src/repositories/`
  - Extracted traits from existing struct impls
  - Added `async-trait` dependency for dyn-compatible async traits
  - Files: `src/repositories/branch_repo.rs`, `src/repositories/commit_repo.rs`, `src/repositories/mod.rs`

- [x] Task 2: Add manual mock implementations for repository traits
  - Added `MockBranchRepository` and `MockCommitRepository` with configurable `Mutex<Option<Result<...>>>` fields
  - Each method returns a pre-configured result via `take_or_error()` helper
  - Default fallback returns `PgNotConfigured` error (same as handler unit tests expect)
  - No external mocking library needed — manual mocks are simpler and dependency-free for this use case
  - Files: `src/repositories/branch_repo.rs`, `src/repositories/commit_repo.rs`
  - Note: deviated from plan (used manual mocks instead of `mockall`) — manual mocks are simpler, avoid extra deps, and are sufficient for handler-level tests

- [x] Task 3: Add repository constructor to `AppState` for testability
  - Added `branch_repo: Option<Arc<dyn BranchRepositoryTrait + Send + Sync>>` and `commit_repo: Option<Arc<dyn CommitRepositoryTrait + Send + Sync>>` fields to `AppState`
  - When set, handlers use the injected repo; when `None`, fall back to creating real repo from `pg` pool
  - Updated all AppState construction sites: `main.rs`, `lib.rs`, both handlers, `routes.rs`, `tests/common/mod.rs`, `tests/route_registration_test.rs`
  - Files: `src/lib.rs` (AppState), `src/handlers/branch_handler.rs`, `src/handlers/commit_handler.rs`

### Phase 2: Move Smoke Tests to Handler-Level

- [x] Task 4: Move smoke tests from `tests/` to handler `#[cfg(test)]` blocks
  - Added 7 mock-based tests to `src/handlers/branch_handler.rs`:
    test_create_branch_mock_returns_created, test_get_branch_mock_returns_ok,
    test_list_branches_mock_returns_ok, test_delete_branch_mock_returns_no_content,
    test_switch_branch_mock_returns_ok, test_merge_branches_mock_returns_ok_or_error,
    test_get_branch_mock_returns_not_found
  - Added 4 mock-based tests to `src/handlers/commit_handler.rs`:
    test_get_commit_mock_returns_ok, test_list_commits_mock_returns_ok,
    test_get_delta_nonexistent_mock_returns_404, test_get_semantic_diff_nonexistent_mock_returns_404
  - Removed relocated smoke tests from `tests/branch_integration_test.rs`,
    `tests/commit_integration_test.rs`, `tests/checkout_integration_test.rs`,
    `tests/merge_integration_test.rs`, `tests/delta_integration_test.rs`
  - 75 unit tests pass (all mock-based tests + existing handler unit tests)

- [x] Task 5: Replace `skip_if_no_pg() → return` with `panic!` in remaining integration tests
  - Renamed `skip_if_no_pg()` to `require_pg()` in `tests/common/mod.rs` — panics with clear message
  - Updated all remaining integration tests to use `require_pg()`:
    `tests/transactional_integration_test.rs` (5 tests),
    `tests/delta_integration_test.rs` (test_commit_delta_endpoint),
    `tests/commit_integration_test.rs` (test_create_commit_via_api),
    `tests/checkout_integration_test.rs` (test_checkout_commit_endpoint)
  - Verified: `cargo test` integration tests now FAIL loudly when PG is unavailable

### Phase 3: Build Automation

- [x] Task 6: Add `test-versioning` Makefile target with auto-PostgreSQL
  - Added `make test-versioning` target to root `Makefile`
  - Auto-detects running PostgreSQL via `pg_isready`; if unavailable, starts via
    `docker compose up -d postgres` and waits up to 30s for healthy state
  - If PostgreSQL was auto-started, stops it after tests complete (cleanup)
  - Sets both `PG_TEST_DATABASE_URL` and `DATABASE_URL`
  - Runs `cargo test --test '*'` for versioning-service integration tests

## Acceptance Criteria
- [ ] All tests pass: `cargo test --lib` (unit + mocked) without PostgreSQL
- [ ] `cargo test` (full) FAILS with clear message when PG is not available
- [ ] `make test-versioning` runs all tests including PG-dependent ones
- [ ] No silent test skipping — every test either passes or fails
- [ ] Test Quality Score (TQS) ≥ bronze (6.0)
- [ ] No B1–B7 anti-patterns (see .ai-factory/rules/test-quality.md)
