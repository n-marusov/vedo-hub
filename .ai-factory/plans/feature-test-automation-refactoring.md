# Implementation Plan: Test Automation Refactoring

Branch: feature/test-automation-refactoring
Created: 2026-07-29

## Settings
- Testing: no (скрипты автоматизации — без тестов)
- Logging: minimal (только WARN/ERROR)
- Docs: yes (обновить make help, README, antora docs)

## Roadmap Linkage
Milestone: none
Rationale: Skipped by user

## Acceptance Criteria

- [ ] **A1 — All targets build**: `make build` passes without errors on all four languages
- [ ] **A2 — `test-unit-fast` fail-fast**: если test-rust-fast падает, test-go-fast НЕ запускается
- [ ] **A3 — `test-unit-full` статистика**: все 4 языка выполняются, ошибки собираются и выводятся в конце
- [ ] **A4 — `test-integration-go-fast` не глотает ошибки**: убрать `|| true`, цель реально падает при падении тестов
- [ ] **A5 — Все integration типы вызываются**: auth-service-org добавлен, versioning включён в test-integration-fast/full
- [ ] **A6 — `ci-fast` fail-fast**: proto → build → lint → test-fast → typecheck — остановка при первой ошибке
- [ ] **A7 — `ci-full` fail-fast**: vendor-go → ci-fast → quality → integration-fast → e2e-fast → gates-fast — без аккумулятора
- [ ] **A8 — Цвет работает**: `make test-unit-fast` показывает зелёный `[PASS]` / красный `[FAIL]`
- [ ] **A9 — Цвет безопасен в CI**: `TERM=dumb make test-unit-fast` не выводит ANSI-escape-символы
- [ ] **A10 — B5 детекция**: `test-quality-gate.sh` находит тесты без единого assert-вызова
- [ ] **A11 — TQS scoring**: `test-quality-gate.sh --score` выводит per-service TQS с грейдом (gold/silver/bronze/fail)
- [ ] **A12 — Machine-readable output**: `test-quality-gate.sh —score` заканчивается строкой `<!-- test-quality-gate: ... -->`
- [ ] **A13 — JSON output**: `test-quality-gate.sh --json` выдаёт валидный JSON
- [ ] **A14 — traceability-validator.sh работает**: скрипт парсит TTL, сверяет с FS, считает RCS, находит P0 orphans
- [ ] **A15 — Качество в гейтах**: `make test-gates-fast` включает quality + traceability проверки
- [ ] **A16 — CI trending**: `ci-full` сохраняет snapshot в `.quality-trends/YYYY-MM-DD.json`
- [ ] **A17 — `make help` актуален**: все новые цели видны в help, старые удалены
- [ ] **A18 — Документация обновлена**: README и antora docs не ссылаются на удалённые цели
- [ ] **A19 — Ветка не содержит мусора**: `.quality-trends/*.json` в .gitignore, нет случайных файлов

## Commit Plan— `make help` актуален**: все новые цели видны в help
- [ ] **A20 — Документация обновлена**: README и antora docs не ссылаются на удалённые цели

## Commit Plan

- **Commit 1** (after tasks 1-4): `feat: add color infra + rust/go/python test-fast/full targets`
- **Commit 2** (after tasks 5-7): `feat: add typescript-fast/full + unit/fast/full aggregates + integration-rust refactor`
- **Commit 3** (after tasks 8-11): `feat: fix integration-go ||true bug + versioning fail-fast + aggregates + e2e-api maxFailures`
- **Commit 4** (after tasks 12-15): `feat: e2e-gui/aggregate + ci-fast + ci-full + final color pass`
- **Commit 5** (after tasks 16-19): `feat: B5 detection + TQS scoring + traceability-validator.sh`
- **Commit 6** (after tasks 20-22): `feat: quality integration into gates + CI trending + documentation`

## Tasks

### Phase 0: Foundation
- [x] Task 1: Add ANSI color infrastructure to Makefile preamble
  - Add `C_RED`, `C_GREEN`, `C_YELLOW`, `C_CYAN`, `C_BOLD`, `C_RESET` variables
  - Wrap with `ifeq ($(TERM),dumb)` for CI-safe fallback
  - Files: `Makefile` (after L26)

### Phase 1: Unit Tests — per-language fast/full targets
- [x] Task 2: Refactor `tools/build/rust.mk` — add test-rust-fast (fail-fast) and test-rust-full (statistics)
  - `test-rust-fast`: loop over RUST_DIRS, `cd $$dir && cargo test 2>&1` — stops at first failure via `set -e`
  - `test-rust-fast-unit`: same but `cargo test --lib`
  - `test-rust-full`: keeps current accumulator pattern with color
  - `test-rust-full-unit`: keeps accumulator
  - Rename old `test-rust` → keep as alias (will be removed later)
  - Add color: `$(C_CYAN)[Rust]$(C_RESET)` prefix, `$(C_RED)TEST_FAILED:$(C_RESET)` on failure
  - Remove old `test-rust-unit` separate target (merged into test-rust-fast/full)
  - Files: `tools/build/rust.mk`

- [x] Task 3: Refactor `tools/build/go.mk` — add test-go-fast and test-go-full
  - Same pattern as rust: fast = direct `cd $$dir && go test ./...` (fail-fast via `set -e`)
  - full = current accumulator with color
  - Color: `$(C_CYAN)[Go]$(C_RESET)` prefix
  - Files: `tools/build/go.mk`

- [x] Task 4: Refactor `tools/build/python.mk` — add test-python-fast and test-python-full
  - Same pattern, but preserve `|| test $$? -eq 5` for "no tests collected" in BOTH modes
  - Fast: `cd $$dir && (uv run pytest 2>&1 || test $$? -eq 5)`
  - Full: accumulator with same protection
  - Color: `$(C_CYAN)[Python]$(C_RESET)` prefix
  - Files: `tools/build/python.mk`

- [x] Task 5: Refactor `tools/build/typescript.mk` — add test-typescript-fast and test-typescript-full
  - Same pattern as go/rust
  - Color: `$(C_CYAN)[TypeScript]$(C_RESET)` prefix
  - Files: `tools/build/typescript.mk`

### Phase 2: Unit Test Aggregates
- [x] Task 6: Add aggregate targets to Makefile (+ main test entry points)
  - `test-unit-fast`: calls `test-rust-fast-unit` → `test-go-fast` → `test-python-fast` → `test-typescript-fast`, fail-fast
  - `test-unit-full`: calls all `-full` variants, accumulates errors
  - `test-fast`: calls `test-unit-fast` (integration/e2e/gates will be added in later phases)
  - `test-full`: calls `test-unit-full` (full cascade later)
  - Remove old `test` and `test-all` targets (no aliases — explicit naming)
  - Each has green `[PASS]` / red `[FAIL]` banner with color
  - Files: `Makefile` (L180-212)

### Phase 3: Integration Tests — fail-fast + bug fixes
- [x] Task 7: Refactor `test-integration-rust` — add fast/full versions
  - `test-integration-rust-fast`: already fail-fast (break), keep as-is with color
  - `test-integration-rust-full`: run all test files, collect failures, report at end
  - Add color to Neo4j healthcheck messages
  - Files: `Makefile` (L219-269)

- [x] Task 8: Refactor `test-integration-go` — FIX BUG + add fast/full
  - **FIX:** Remove `|| true` that swallows errors (lines 273-280)
  - `test-integration-go-fast`: fail-fast, first failing test stops execution
  - `test-integration-go-full`: runs all, collects failures
  - **Add missing:** `tests/integration/auth-service-org/` was not being called — add it to both fast and full
  - Files: `Makefile` (L271-280)

- [x] Task 9: Refactor `test-versioning` — add fast/full
  - `test-versioning-fast`: fail-fast — unit fails → integration doesn't run
  - `test-versioning-full`: run all, collect failures, report
  - Add color
  - Files: `Makefile` (L282-323)

- [x] Task 10: Add `test-integration-fast` and `test-integration-full` aggregates
  - `test-integration-fast`: calls rust-fast → go-fast → versioning-fast
  - `test-integration-full`: calls rust-full → go-full → versioning-full, accumulates
  - **Note:** `test-integration` previously did NOT call `test-versioning` — this fixes that gap
  - Files: `Makefile` (L216)

### Phase 4: E2E Tests — maxFailures + fast/full
- [x] Task 11: Refactor `test-e2e-api` — add fast/full
  - `test-e2e-api-fast`: `--max-failures=1`
  - `test-e2e-api-full`: no max-failures (full run with retries)
  - Can be done via Makefile flags (no need to change playwright config — `--max-failures` CLI arg)
  - Files: `Makefile` (L330-337)

- [x] Task 12: Refactor `test-e2e-gui` and `test-e2e` — add fast/full
  - `test-e2e-gui-fast`: already has `maxFailures: 1` in config — keep
  - `test-e2e-gui-full`: override via `--max-failures=0`
  - `test-e2e-fast`: calls API-fast → GUI-fast, fail-fast
  - `test-e2e-full`: calls API-full → GUI-full, accumulates
  - Files: `Makefile` (L327-345)

### Phase 5: CI — always fail-fast
- [x] Task 13: Create `ci-fast` target (rename old `ci`)
  - `ci-fast`: proto-all → build → lint → test-fast → typecheck, fail-fast (stops at first failure)
  - Add color section banners: `=== CI Pipeline Started ===` in cyan, `=== CI pipeline passed ===` in green, `!!! CI Pipeline FAILED !!!` in red+bold
  - Files: `Makefile` (L634-654)

- [x] Task 14: Update `ci-full` — always fail-fast + comprehensive
  - Keep existing name `ci-full` (already has `-full` suffix)
  - Change to fail-fast: remove `|| failed=1` accumulator, use direct chaining
  - Pipeline: vendor-go → ci-fast → test-quality-gate → docs-lint → docker-up-test || true → test-integration-fast → test-e2e-fast → test-gates-fast → npm-audit
  - `docker-up-test || true` exempted (may already be running)
  - Add color banners
  - Files: `Makefile` (L655-674)

### Phase 6: Color — Final Pass
- [x] Task 15: Add color to all remaining PASS/FAIL/INFO messages across all Makefile targets
  - `lint-*` targets: `LINT_FAILED` in red
  - `build-*` targets: `BUILD_FAILED` in red
  - `clean-*` targets: section headers in cyan
  - Coverage: `[Coverage]` header in cyan
  - test-quality-gate: section banners
  - docs-lint: `[Docs]` header
  - npm-audit: `[Security]` header
  - Files: `Makefile` (all remaining), `tools/build/*.mk`

### Phase 7: Test Quality Gate Enhancement
- [x] Task 16: Add B5 zero-assertion test detection to `test-quality-gate.sh`
  - Go heuristic (awk-based function body analysis)
  - B5 count in machine-readable output
  - Rust/Python/TS heuristics can be added as follow-up

- [x] Task 17: Add TQS scoring mode (`--score` flag) to `test-quality-gate.sh`
  - When `--score` is passed, calculate TQS for each service directory
  - Heuristic TQS (grep-based, not AST):
    - Structural (0.30): BDD naming count, file location
    - Dependencies (0.25): mock count, fixture usage
    - Readability (0.20): given/when/then comments, test length
    - Safety (0.15): B1-B7 violations
    - Coverage (0.10): assertions per test, scenario variety
  - Output per-service TQS with grade (gold/silver/bronze/fail/na)
  - Uses `bc` for float arithmetic (fallback: integer approximation)
  - Files: `tools/scripts/test-quality-gate.sh`

- [x] Task 18: Add machine-readable and JSON output modes to `test-quality-gate.sh`
  - Always produce HTML comment line: `<!-- test-quality-gate: PASS|BLOCK | tqs: X.X | grade: silver -->`
  - `--json` flag: output full JSON report to stdout
  - JSON schema: `{ gate, tqs, grade, blocking, warnings, b5_count, files_scanned, service_breakdown: [...] }`
  - Files: `tools/scripts/test-quality-gate.sh`

### Phase 8: Traceability Validator
- [x] Task 19: Create `tools/scripts/traceability-validator.sh`
  - B1: Parse `traceability.ttl` — extract `vdo:validates`, `vdo:verifiedBy`, `vdo:Requirement`, `vdo:TestSuite` (grep-based RDF Turtle parser)
  - B2: Filesystem inventory — count REQ-*.md files, test files, US-*
  - B3: Cross-validation — TTL→FS stale references, FS→TTL untracked artifacts, structural integrity
  - B4: RCS computation — weighted coverage (P0=3, P1=2, P2=1, P3=0)
  - B5: P0 orphan detection — P0 requirements without test coverage → BLOCK in strict mode
  - B6: Machine-readable output: `<!-- traceability: VALID|ISSUES | rcs: X.X | grade: partial | orphan-p0: N -->`
  - Minimal logging per user preference (only WARN/ERROR)
  - Files: `tools/scripts/traceability-validator.sh` (NEW)

### Phase 9: Quality Integration into Gates + CI Trending
- [x] Task 20: Integrate quality checks into `test-gates-fast` and `test-gates-full`
  - `test-gates-fast`: run_all_tests.sh → test-quality-gate.sh --score → traceability-validator.sh (fail-fast chain)
  - `test-gates-full`: same sequence but accumulates errors
  - Keep standalone `test-quality-gate` target for CI-only quality check
  - Files: `Makefile` (L345-359)

- [x] Task 21: Add CI quality trending infrastructure
  - Create `.quality-trends/` directory (gitignored)
  - In `ci-full`: save daily TQS/RCS snapshot via `test-quality-gate.sh --json` → `.quality-trends/YYYY-MM-DD.json`
  - Add `.quality-trends/*.json` to `.gitignore`
  - Optional regression check: compare today vs yesterday, delta < -0.5 → WARN
  - Files: `Makefile` (ci-full), `.gitignore`

### Phase 10: Documentation
- [x] Task 22: Update documentation for new targets
  - `make help` already auto-generates from `##` comments — ensure all new targets have `##` descriptions
  - Clean up old `##@ Test — Unit` sections
  - Update README.md if it documents make targets
  - Update Antora docs if they reference test commands
  - Files: `Makefile` (help comments), `README.md`, `docs/antora/{{developer-guide}}/pages/testing.adoc` (if exists)
