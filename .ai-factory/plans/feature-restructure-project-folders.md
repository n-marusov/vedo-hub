# Plan: Restructure Project Folder Structure

Branch: feature/restructure-project-folders
Created: 2026-07-25

> Move from `src/`-centric layout to domain-separated `apps/`+`tools/`+`docs/`+`config/` layout. No renaming of services or internal service structures. Minimal changes, maximum clarity.

## Settings

| Setting | Value | Reason |
|---------|-------|--------|
| Testing | Yes — verify nothing broke | Run existing gate scripts after restructure |
| Logging | Minimal | This is `git mv` + path replace — no application code changes |
| Documentation | Yes — mandatory checkpoint | Update AGENTS.md and ARCHITECTURE.md with new paths |
| Roadmap linkage | Skip | Infrastructure work — not tied to a specific milestone |
| Constraints | No service renames, no internal service restructuring | As decided in exploration |

## Research Context

From `.ai-factory/RESEARCH.md` Active Summary (abridged):
- The project currently uses a `src/`-centric layout where source code, build tooling, Docker templates, documentation, and deployment configs are mixed.
- Goal: separate into `apps/` (runnable), `tools/` (build/dev), `docs/`, `config/`, keeping `deploy/`, `tests/`, `specs/`, `design/` at root.
- No domain grouping within services — flat list under `apps/services/`.

## Target Structure

```
vedo-hub/
├── apps/
│   ├── services/          # All 18 microservices (flat, no grouping)
│   ├── vedo-cli/          # Operator CLI
│   └── shared/            # Shared libraries (llm, proto, rust)
├── tools/
│   ├── build/             # Makefile includes (.mk)
│   ├── dockerfiles/       # Dockerfile templates
│   ├── scaffolds/         # Service templates
│   └── scripts/           # Pre-commit + quality gate scripts
├── docs/                  # Antora documentation
├── deploy/                # Docker Compose, Helm, CI, Keycloak (unified)
├── config/                # Environment files (.env.dev, .env.test, .env.staging)
├── tests/
│   ├── cli/
│   ├── e2e/               # Flattened from e2e/playwright/
│   │   ├── config/        # Playwright configs
│   │   ├── data/          # Test data files
│   │   ├── pages/         # Page Object Models
│   │   ├── scripts/       # run-ci.sh, stub-server, sign-jwt, global-setup
│   │   └── specs/         # Test specs (was tests/)
│   ├── security/
│   ├── integration/       # org-api + ticket-api
│   └── gates/             # CI gate scripts (test_*.sh)
├── specs/                 # Git submodule
├── design/
├── .ai-factory/
├── .agents/
├── Makefile               # Moved from src/ to root
├── AGENTS.md
├── README.md
└── LICENSE
```

## Tasks

### Phase 1: Move source files

- [x] **Task 1 — Move all source directories (git mv)**

  **Subtask 1.1 — apps/**
  ```
  git mv src/services   apps/services     # All 18 services + Cargo.toml workspace
  git mv src/services/shared apps/shared   # Shared libs out of services
  git mv src/cli         apps/vedo-cli     # CLI
  ```

  **Subtask 1.2 — tools/**
  ```
  git mv src/build       tools/build       # Makefile includes
  git mv src/docker      tools/dockerfiles # Dockerfile templates
  git mv src/templates   tools/scaffolds   # Service scaffolds
  git mv src/scripts     tools/scripts     # Pre-commit + quality gate scripts
  ```

  **Subtask 1.3 — docs/**
  ```
  git mv src/docs        docs              # Antora documentation (src/docs/antora → docs/antora)
  ```
  Note: `src/docs/` only contains `antora/` — move the whole thing.

  **Subtask 1.4 — config/**
  ```
  mkdir config
  git mv .env.dev   config/.env.dev
  git mv .env.test  config/.env.test
  git mv .env.staging config/.env.staging
  ```

  **Subtask 1.5 — Root Makefile**
  ```
  git mv src/Makefile .                    # Move to project root
  ```

  **Subtask 1.6 — Cleanup**
  - Remove `src/keycloak/vedo-core-realm.json` (duplicate of `deploy/keycloak/`)
  - Remove `src/services/scripts/run_integration_rust.sh` (will be recreated at `tests/scripts/`)
  - Remove empty `src/` directory once all moves complete
  - Remove `package-lock.json` from `tests/e2e/` (keep only `pnpm-lock.yaml`)

  **Verification:** `ls -d apps/services/*/` shows 18 services; `src/` does not exist.

  > ⚠️ `git mv` preserves git history. All subsequent tasks update path references in files.

---

### Phase 2: Update build tooling

- [x] **Task 2 — Makefile, .mk includes, and Dockerfiles**

  **Subtask 2.1 — Root Makefile (`Makefile`)**
  - Update `ROOT := $(realpath $(dir $(lastword $(MAKEFILE_LIST))))` — now resolves to project root
  - Update include paths: `include $(ROOT)/build/rust.mk` → `include tools/build/rust.mk` (same for all 6 includes)
  - Update `PROTO_DIR := $(ROOT)/services/shared/proto` → `PROTO_DIR := apps/shared/proto`
  - Update `--python_out=../../../../services/document-extractor/` → `--python_out=../../../../services/document-extractor/` (verify relative path still resolves correctly from `apps/shared/proto/` → `apps/services/document-extractor/grpc_client/`)
  - Update echo message: `src/services/document-extractor/grpc_client/` → `apps/services/document-extractor/grpc_client/`

  **Subtask 2.2 — `tools/build/*.mk`**
  - `rust.mk`: `find $(ROOT) -maxdepth 4` — verify it finds Rust services under `apps/services/`. Update `-not -path "*/services/Cargo.toml"` to `-not -path "*/apps/services/Cargo.toml"` (or test if glob still matches)
  - `go.mk`: `cd $(ROOT) && find . -maxdepth 5 -name go.mod` — verify depth still works; add `-not -path "*/apps/shared/*"` if needed
  - `python.mk`: check `PYTHON_DIRS` and update paths
  - `typescript.mk`: check `TS_DIRS` and update paths
  - `docker.mk`: replace all `SERVICE_DIR=src/services` → `SERVICE_DIR=apps/services`; replace `DOCKER_DIR` reference from `$(ROOT)/docker` → `tools/dockerfiles`; update `case "$$dir" in cli|services/*/*) continue;;` → handle new path structure
  - `yaml.mk`: check for path references

  **Subtask 2.3 — `tools/dockerfiles/Dockerfile.*`**
  - **Dockerfile.rust**: Replace `ARG SERVICE_DIR=src/services` → `apps/services` in comments + ARG default. All `${SERVICE_DIR}` references use the ARG — no inline paths to change. Update doc comments.
  - **Dockerfile.go**: Replace `WORKDIR /vedo-core/src/services/${BINARY_NAME}` → `/vedo-core/apps/services/${BINARY_NAME}`. Replace `COPY ${SERVICE_DIR}/shared /vedo-core/src/services/shared/` → `COPY apps/shared /vedo-core/apps/shared/`. Replace `COPY ${SERVICE_DIR}/ticket-api /vedo-core/src/services/ticket-api/` → `COPY apps/services/ticket-api /vedo-core/apps/services/ticket-api/`. Update doc comments.
  - **Dockerfile.python**: `ARG SERVICE_DIR=src/services/<name>` → `apps/services/<name>` in comments. `${SERVICE_DIR}` usage is through ARG — verify.
  - **Dockerfile.typescript**: `ARG SERVICE_DIR=src/services/<name>` → `apps/services/<name>` in comments. Check all `${SERVICE_DIR}` usage.

  > 💡 **Files to change:** `Makefile`, `tools/build/rust.mk`, `tools/build/go.mk`, `tools/build/python.mk`, `tools/build/typescript.mk`, `tools/build/docker.mk`, `tools/build/yaml.mk`, `tools/dockerfiles/Dockerfile.rust`, `tools/dockerfiles/Dockerfile.go`, `tools/dockerfiles/Dockerfile.python`, `tools/dockerfiles/Dockerfile.typescript`

---

### Phase 3: Update CI/CD and deployment configs

- [x] **Task 3 — docker-compose, gitlab-ci, pre-commit**

  **Subtask 3.1 — `deploy/docker-compose.yml` + `deploy/docker-compose.*.yml`**
  - Replace `dockerfile: src/docker/Dockerfile.*` → `dockerfile: tools/dockerfiles/Dockerfile.*` in all compose files
  - Replace `SERVICE_DIR: src/services/frontend` → `SERVICE_DIR: apps/services/frontend`
  - Replace `SERVICE_DIR: src/services` → `SERVICE_DIR: apps/services`
  - Check all other compose files: `docker-compose.docs.yaml`, `docker-compose.llm.yaml`, `docker-compose.observability.yml`, `docker-compose.test.yml`
  - Add `env_file: config/.env.dev` directives or document that docker-compose needs `--env-file config/.env.dev` flag

  **Subtask 3.2 — `deploy/ci/gitlab-ci.yml`**
  - Global find-and-replace: `src/services/` → `apps/services/`, `src/cli/` → `apps/vedo-cli/`, `src/docs/` → `docs/`, `src/scripts/` → `tools/scripts/`
  - Update `cd src/services/frontend` → `cd apps/services/frontend`
  - Update `cd src/services/ontology-service` → `cd apps/services/ontology-service`
  - Update `cd src/services/shared/proto` → `cd apps/shared/proto`
  - Update `src/scripts/test-quality-gate.sh` → `tools/scripts/test-quality-gate.sh`
  - Update `changes:` rules patterns: `src/services/**/*.go` → `apps/services/**/*.go`
  - Update artifact paths: `src/services/frontend/junit.xml` → `apps/services/frontend/junit.xml`

  **Subtask 3.3 — `.pre-commit-config.yaml`**
  - Replace `files: ^src/services/(metrics-service|...)` → `files: ^apps/services/(metrics-service|...)`
  - Replace `files: ^src/services/(frontend|publish-browse-ui)/` → `files: ^apps/services/(frontend|publish-browse-ui)/`
  - Replace `entry: sh src/scripts/pre-commit-*.sh` → `entry: sh tools/scripts/pre-commit-*.sh`
  - Replace `exclude: ^src/services/frontend/.*\.css$` → `exclude: ^apps/services/frontend/.*\.css$`

  **Subtask 3.4 — `tools/scripts/pre-commit-*.sh` (4 files)**
  - Update grep patterns: `grep '^src/services/'` → `grep '^apps/services/'`
  - Update sed expressions: `s|^\(src/services/[^/]*\)/.*|\1|` → `s|^\(apps/services/[^/]*\)/.*|\1|`
  - Files: `pre-commit-gofmt.sh`, `pre-commit-golangci-lint.sh`, `pre-commit-cargo-fmt.sh`, `pre-commit-cargo-clippy.sh`

  **Subtask 3.5 — `tools/scripts/test-quality-gate.sh`**
  - Update `TARGET_DIR="${1:-src/services}"` → `TARGET_DIR="${1:-apps/services}"`

  > 💡 **Files to change:** 5 docker-compose files, `deploy/ci/gitlab-ci.yml`, `.pre-commit-config.yaml`, 4 pre-commit scripts, `test-quality-gate.sh`

---

### Phase 4: Update test infrastructure

- [ ] **Task 4 — Gate scripts + test directory restructuring**

  **Subtask 4.1 — Move and update gate scripts**
  - Create `tests/gates/` directory
  - `git mv tests/test_*.sh tests/gates/`
  - `git mv tests/scripts/terminology_integrity_test.sh tests/gates/`
  - Keep `tests/run_all_tests.sh` and `tests/run_integration_rust.sh` at `tests/` root (orchestrators, not gates)

  **Subtask 4.2 — Update gate script paths**
  Each gate script uses `ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"` — after moving to `tests/gates/`, this now needs `../..` to reach project root.
  - Update: `ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"` → `ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"`
  - Replace `SERVICES_DIR="$ROOT_DIR/src/services"` → `SERVICES_DIR="$ROOT_DIR/apps/services"`
  - Replace `cd "$ROOT/src/services/..."` → `cd "$ROOT/apps/services/..."`
  - Replace `go -C src/cli test` → `go -C apps/vedo-cli test`
  - Update all service-specific paths in:
    - `test_bola_bfla_gate.sh`: `src/services/api-gateway` → `apps/services/api-gateway`
    - `test_ci_and_compose.sh`: `src/services` → `apps/services`
    - `test_cli_contracts.sh`: `src/cli` → `apps/vedo-cli`
    - `test_contract_gate.sh`: all service paths
    - `test_health_metadata.sh`: `src/services` → `apps/services`
    - `test_milestone_docker_gate.sh`: `src/services` → `apps/services`
    - `test_native_stubs.sh`: `src/services` → `apps/services`
    - `test_antora_setup.sh`: `src/docs/antora` → `docs/antora`
    - All other `test_*.sh` files (25 total)

  **Subtask 4.3 — Restructure `tests/e2e/` (flatten from `playwright/`)**
  - Move `tests/e2e/playwright/tests/` → `tests/e2e/specs/`
  - Move `tests/e2e/playwright/fixtures/` → `tests/e2e/data/`
  - Move `tests/e2e/playwright/pages/` → `tests/e2e/pages/` (Page Object Models — keep)
  - Move `tests/e2e/playwright/playwright.*.config.ts` (×6) → `tests/e2e/config/`
  - Move `tests/e2e/playwright/run-ci.sh`, `stub-server.mjs`, `sign-jwt.js`, `global-setup.ts` → `tests/e2e/scripts/`
  - Move `tests/e2e/playwright/package.json`, `pnpm-lock.yaml` → `tests/e2e/` root
  - Move `tests/e2e/playwright/tests/fixtures.ts`, `graphql-fixtures.ts`, `jwt-tokens.ts`, `ontology-test-data.ts` → `tests/e2e/specs/helpers/`
  - Rename `tests/e2e/specs/gui/pages/` → `tests/e2e/specs/gui/smoke/`
  - Rename `tests/e2e/specs/gui/user-stories/` → `tests/e2e/specs/gui/flows/`
  - Delete `tests/e2e/playwright/node_modules/`, `playwright-report/`, `test-results/` (artifacts, not in repo)
  - Delete `tests/e2e/playwright/` after all moves
  - Delete `tests/e2e/package-lock.json` (keep only `pnpm-lock.yaml`)

  **Subtask 4.4 — Update e2e config paths**
  - Update all `playwright.*.config.ts` files: adjust `testDir`, `outputDir`, paths relative to new `config/` location
  - Update `package.json` scripts: adjust paths to configs

  **Subtask 4.5 — Move integration tests**
  - Create `tests/integration/`
  - `git mv tests/org-api tests/integration/org-api`
  - `git mv tests/ticket-api tests/integration/ticket-api`

  **Subtask 4.6 — Update `tests/run_integration_rust.sh`**
  - Update canonical script path reference from `src/services/scripts/` to new location
  - Move `src/services/scripts/run_integration_rust.sh` to `tests/scripts/run_integration_rust.sh` (or appropriate location)

  > 💡 **Files to change:** ~25 gate scripts, e2e configs, `run_integration_rust.sh`

---

### Phase 5: Update service internals

- [ ] **Task 5 — Go modules, Cargo workspace, intra-service references**

  **Subtask 5.1 — Cargo workspace**
  - `apps/services/Cargo.toml`: workspace members use relative directory names (`ontology-service`, `shared`, etc.) — **no change needed** since services stay in same relative positions
  - `apps/shared/Cargo.toml`: verify it's still referenced correctly from workspace
  - Delete `apps/services/target/` if it exists (build artifact)

  **Subtask 5.2 — Go module replace directives**
  - Find all `go.mod` files with `replace` directives pointing to `../shared`: update relative paths
    - `apps/services/api-gateway/go.mod`: `../shared` → `../../shared` (or verify new relative path)
    - `apps/services/auth-service/go.mod`: same
    - All other Go services that import shared
  - Find `replace` directives for `../ticket-api`, `../ticket-notifier`, etc.: update relative paths
  - Verify `go mod tidy` passes for each service

  **Subtask 5.3 — Update `apps/vedo-cli/`**
  - Check for hardcoded paths referencing `src/` or `src/services/` in Go source files
  - Update `go.mod` if it has replace directives

  > 💡 **Files to change:** All `go.mod` files with `replace` directives, `apps/services/Cargo.toml` (verify only)

---

### Phase 6: Update documentation and config files

- [ ] **Task 6 — AGENTS.md, ARCHITECTURE.md, .gitignore, .dockerignore**

  **Subtask 6.1 — `AGENTS.md`**
  - Replace all `src/services/` → `apps/services/` in Key Entry Points table
  - Replace `src/cli/command.go` → `apps/vedo-cli/command.go`
  - Replace `src/Makefile` → `Makefile`
  - Replace `src/docs/antora/` → `docs/antora/` in Documentation table
  - Replace `src/services/frontend/index.html` → `apps/services/frontend/index.html`
  - Update Project Structure ASCII diagram to match new layout

  **Subtask 6.2 — `.ai-factory/ARCHITECTURE.md`**
  - Update Folder Structure ASCII diagram
  - Update any path references in prose

  **Subtask 6.3 — `.gitignore`**
  - Replace `src/services/frontend/dist/` → `apps/services/frontend/dist/`
  - Replace `src/services/publish-browse-ui/dist/` → `apps/services/publish-browse-ui/dist/`
  - Replace `tests/e2e/playwright/test-results/` → `tests/e2e/test-results/`
  - Replace `tests/e2e/playwright/playwright-report/` → `tests/e2e/playwright-report/`

  **Subtask 6.4 — `.dockerignore`**
  - Check for any `src/` path references and update

  > 💡 **Files to change:** `AGENTS.md`, `.ai-factory/ARCHITECTURE.md`, `.gitignore`, `.dockerignore`

---

### Phase 7: Verification

- [ ] **Task 7 — Verify everything works**

  **Subtask 7.1 — Structural checks**
  - Verify no `src/` directory remains
  - Verify `apps/services/` contains 18 service directories
  - Verify `tools/build/`, `tools/dockerfiles/`, `tools/scaffolds/`, `tools/scripts/` exist
  - Verify `docs/antora/antora-playbook.yml` exists
  - Verify `config/.env.dev`, `config/.env.test`, `config/.env.staging` exist
  - Verify `tests/gates/` contains all `test_*.sh` scripts
  - Verify `tests/integration/org-api/` and `tests/integration/ticket-api/` exist
  - Verify `tests/e2e/` has `config/`, `data/`, `pages/`, `scripts/`, `specs/`, `package.json`

  **Subtask 7.2 — No stray references**
  ```bash
  grep -r "src/services" --include="*.go" --include="*.rs" --include="*.ts" \
       --include="*.yml" --include="*.yaml" --include="*.sh" --include="*.mk" \
       --include="*.json" --include="*.md" . | grep -v node_modules | grep -v .git | grep -v .ai-factory
  ```
  Expected: zero results (except possibly comments referencing historical paths).

  **Subtask 7.3 — Run gate scripts**
  ```bash
  cd tests/gates
  bash test_native_stubs.sh          # Verifies new service layout
  bash test_ci_and_compose.sh        # Verifies CI + compose config
  bash test_antora_setup.sh          # Verifies docs paths
  bash test_cli_contracts.sh         # Verifies CLI still works
  bash test_contract_gate.sh         # Verifies Go contract tests
  bash test_build_commands.sh        # Verifies build targets
  ```

  **Subtask 7.4 — Build verification**
  ```bash
  make build-go      # All Go services compile
  make build-rust    # All Rust services compile
  make lint-go       # Go lint passes
  make lint-rust     # Rust clippy passes
  ```

  **Subtask 7.5 — Docker build verification**
  ```bash
  make docker-build  # All Docker images build successfully
  ```

  > 💡 **Success criteria:** All gate scripts pass, all services build, all Docker images build.

## Acceptance Criteria

- [ ] `src/` directory no longer exists
- [ ] All 18 services accessible at `apps/services/<name>/`
- [ ] All build targets work: `make build-go`, `make build-rust`, `make lint-go`, `make lint-rust`
- [ ] All Docker images build: `make docker-build`
- [ ] All gate scripts pass from `tests/gates/`
- [ ] No stale `src/` path references in any source, build, or config file
- [ ] AGENTS.md and ARCHITECTURE.md reflect new structure
- [ ] `git log --follow` preserves history for moved files

## Commit Plan

| Checkpoint | Tasks | Commit message |
|-----------|-------|----------------|
| 1 | Task 1 | `refactor: move source files to apps/, tools/, docs/, config/` |
| 2 | Task 2 | `refactor: update build tooling for new folder structure` |
| 3 | Tasks 3 + 4 | `refactor: update CI/CD and test infrastructure paths` |
| 4 | Tasks 5 + 6 | `refactor: update service internals and documentation` |
| 5 | Task 7 | `refactor: verify restructured project — all gates pass` |
