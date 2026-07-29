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
- [ ] `cargo test` passes all unit + mocked tests without PostgreSQL
- [ ] `make test-versioning` auto-starts PG, runs migrations, and executes all
  integration tests
- [ ] `cargo test` FAILS (not silently passes) when PG-dependent tests can't
  connect

## Commit Plan
- **Commit 1** (after tasks 1-3): "refactor(versioning): extract repository traits, add mock support"
- **Commit 2** (after tasks 4-5): "test(versioning): move smoke tests to handler-level with mocks, fail hard on missing PG"
- **Commit 3** (after task 6): "build: add test-versioning Makefile target with auto PG"

## Tasks

### Phase 1: Repository Traits & Mock Foundation

- [ ] Task 1: Define `BranchRepositoryTrait` and `CommitRepositoryTrait` in `src/repositories/`
  - Extract trait from existing `BranchRepository` and `CommitRepository` impl blocks
  - Traits include: `create`, `get_by_id`, `list`, `delete`, etc. — matching existing public methods
  - Keep concrete structs implementing the traits via `impl Trait for Struct`
  - Update `mod.rs` to export both traits and structs
  - Files: `src/repositories/branch_repo.rs`, `src/repositories/commit_repo.rs`, `src/repositories/mod.rs`
  - Logging: `INFO` on trait extraction complete

- [ ] Task 2: Add `mockall` dependency and derive `#[automock]` for repository traits
  - Add `mockall = "0.12"` to `[dev-dependencies]` in `Cargo.toml`
  - Add `#[cfg_attr(test, mockall::automock)]` to `BranchRepositoryTrait` and `CommitRepositoryTrait`
  - Verify `cargo build` and `cargo test --lib` pass without errors
  - Files: `Cargo.toml`, `src/repositories/branch_repo.rs`, `src/repositories/commit_repo.rs`
  - Logging: `INFO` on build success

- [ ] Task 3: Add repository constructor to `AppState` for testability
  - In handler `repo_from_state()` helpers, change from concrete type to `Box<dyn Trait>`
  - Add optional `branch_repo: Option<Box<dyn BranchRepositoryTrait + Send + Sync>>` and `commit_repo: Option<Box<dyn CommitRepositoryTrait + Send + Sync>>` fields to `AppState`
  - When set, handlers use the injected repo; when `None`, fall back to creating real repo from `pg` pool
  - This allows tests to inject mock repos while production code uses PG-backed repos
  - Files: `src/lib.rs` (AppState), `src/handlers/branch_handler.rs`, `src/handlers/commit_handler.rs`
  - Logging: `DEBUG` when injected repo is used instead of PG

### Phase 2: Move Smoke Tests to Handler-Level

- [ ] Task 4: Move 12 HTTP-smoke tests from `tests/` to handler `#[cfg(test)]` blocks
  - Relocate from `tests/branch_integration_test.rs`, `tests/checkout_integration_test.rs`,
    `tests/commit_integration_test.rs`, `tests/merge_integration_test.rs`, `tests/delta_integration_test.rs`
  - Each handler test creates `MockBranchRepositoryTrait` / `MockCommitRepositoryTrait`,
    injects into `AppState`, builds `axum::Router`, and asserts HTTP status codes
  - Test naming follows BDD convention: `[Condition]_[Action]_[ExpectedResult]`
  > BDD naming: `[Condition]_[Action]_[ExpectedResult]`
  > Anti-patterns: see `.ai-factory/rules/test-quality.md`
  - Remove the relocated test functions from the `tests/` files
  - Existing handler unit tests (no-PG → `PgNotConfigured`) are preserved
  - Files: `src/handlers/branch_handler.rs`, `src/handlers/commit_handler.rs`
  - Files removed from: `tests/branch_integration_test.rs`, `tests/checkout_integration_test.rs`,
    `tests/commit_integration_test.rs`, `tests/merge_integration_test.rs`, `tests/delta_integration_test.rs`
  - Logging: `INFO` per handler on test relocation complete

- [ ] Task 5: Replace `skip_if_no_pg() → return` with `assert!`/`panic!` in remaining integration tests
  - In `tests/common/mod.rs`, rename `skip_if_no_pg()` to `require_pg()` that panics with clear message:
    "PG_TEST_DATABASE_URL is required. Set it or run 'make test-versioning'."
  - Update all remaining integration test files to use `require_pg()` instead of `skip_if_no_pg()`
  - Remaining tests: `tests/transactional_integration_test.rs` (5 tests) and
    `tests/delta_integration_test.rs` (`test_commit_delta_endpoint`)
  - Verify `cargo test --lib` passes all handler tests without PG
  - Verify `cargo test` (with `--test '*'`) FAILS loudly when PG is not available
  - Files: `tests/common/mod.rs`, `tests/transactional_integration_test.rs`, `tests/delta_integration_test.rs`
  - Logging: `ERROR` eprintln before panic with setup instructions

### Phase 3: Build Automation

- [ ] Task 6: Add `test-versioning` Makefile target with auto-PostgreSQL
  - Script checks if PostgreSQL is reachable via `pg_isready`
  - If not, starts `docker compose up -d postgres` and waits for healthy
  - Runs `sqlx migrate run` to ensure schema is up to date
  - Sets `PG_TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/vedo_versioning`
  - Runs `cargo test -p versioning-service --test '*' -- --test-threads=1 --nocapture`
  - If PostgreSQL was auto-started, stops it via `docker compose stop postgres` (cleanup)
  - Target: `make test-versioning` in root Makefile
  - File: `Makefile`
  - Logging: echo progress messages for each step

## Acceptance Criteria
- [ ] All tests pass: `cargo test --lib` (unit + mocked) without PostgreSQL
- [ ] `cargo test` (full) FAILS with clear message when PG is not available
- [ ] `make test-versioning` runs all tests including PG-dependent ones
- [ ] No silent test skipping — every test either passes or fails
- [ ] Test Quality Score (TQS) ≥ bronze (6.0)
- [ ] No B1–B7 anti-patterns (see .ai-factory/rules/test-quality.md)
