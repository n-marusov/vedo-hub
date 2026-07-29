# VEDO Core Base Rules

> Auto-detected conventions from codebase analysis. Edit as needed.

## Naming Conventions

- **Files:** `snake_case.go`, `snake_case.ts`, `snake_case.py`, `PascalCase.rs` (Rust source), `kebab-case.yaml`
- **Variables:** `camelCase` (Go, TypeScript), `snake_case` (Python, Rust), `SCREAMING_SNAKE_CASE` for constants (all languages)
- **Functions:** `PascalCase` exported (Go), `camelCase` (TypeScript), `snake_case` (Python, Rust)
- **Types/Structs:** `PascalCase` (all languages)
- **Packages/Modules:** `snake_case` (Go packages), `kebab-case` (NPM packages)
- **Docker images:** `vedo-core/<service-name>:<tag>` format

## Module Structure

- **Go services:** Each service is a standalone Go module under `src/services/<name>/`, entrypoint at `main.go` or `cmd/<name>/main.go`
- **Rust services:** Each service under `src/services/<name>/`, entrypoint at `src/main.rs`, Cargo workspace approach
- **Python services:** Single-file `main.py` entrypoint under `src/services/<name>/`
- **TypeScript (Vue 3) services:** Vite-based project under `src/services/<name>/`, entrypoint at `src/main.ts`
- **CLI tool:** Monolithic Go module at `src/cli/`, commands in `cmd/<domain>/`
- **Templates:** Multi-language service scaffolds in `src/templates/`
- **Dockerfiles:** Templates by language in `src/docker/`, per-service Dockerfiles alongside service code

## Error Handling

- **Go:** Error codes prefixed with domain (e.g., `CLI-TICKET-NOT-FOUND`, `CLI-TICKET-UNAUTHORIZED`), structured error responses via `CliOutput` struct with `Status` and `Error` fields
- **All services:** `/health` and `/ready` endpoints return HTTP 200/503 for liveness/readiness probes
- **Error paths** documented with HLV error codes (e.g., `COMPOSE_FAILED`, `BUILD_FAILED`, `PORT_CONFLICT`)
- **Structured error responses** in JSON format for all API endpoints

## Logging

- **Go (all services):** `log/slog` (structured JSON logging) with trace context (`trace_id`, `span_id`)
- **TypeScript (frontend):** Structured logging via Apollo Client error link, `slog`-style context objects
- **Python/Rust:** Structured logging with OpenTelemetry trace propagation
- **Log format:** JSON with service name, trace_id, request_id, severity level
- **Health/metrics endpoints** excluded from trace sampling

## Test Patterns

- **Shell tests:** `tests/test_*.sh` scripts with `test_*()` functions collected via `declare -F`
- **Go tests:** Standard `_test.go` alongside source, `go test ./...`
- **Integration tests:** Multi-script test runner at `tests/run_all_tests.sh`
- **Security tests:** BOLA/BFLA negative authorization tests (Python pytest + TypeScript vitest patterns)
- **E2E tests:** Playwright for frontend at `tests/e2e/`
- **Test boundaries:** Contract tests (per-service), integration tests (cross-service), security gates, platform integrity checks

## Integration Tests Must Be Spec-Driven

**Integration tests MUST be derived from the actual handler code, not written independently.**

Hand-written integration tests that are not checked against the handler's actual signatures inevitably drift from the implementation. The following rules prevent this:

- **Status codes:** Test assertions for HTTP status codes MUST match the handler's actual return type, not guessed values. If a handler returns `StatusCode::CREATED` (201), the test MUST expect 201, not 200 OK.
- **Request/response shapes:** Test payloads and expected responses MUST be derived from the handler's request/response types (Rust structs, Go types, etc.), not re-invented.
- **Direct DB seeding:** When tests seed data directly via database queries (bypassing the API), they MUST replicate ALL relationships and constraints the handler would create. Prefer calling the API endpoint itself for setup over raw DB queries, as this guarantees the data matches the handler's expectations.
- **Route patterns:** Test URL paths MUST match the routes registered in the router (`build_app` / `lib.rs`), not be guessed independently.
- **AI agents:** When generating integration tests, read the handler source code first, then derive expectations from it. Never write test assertions without first confirming the handler's actual behavior.

**Common failure patterns addressed by this rule:**
- Handler returns 201 Created, test expects 200 OK → assertion failure
- Direct Neo4j `CREATE` skips `INSTANCE_OF` relationship → handler can't find the entity → 404
- URL pattern uses `:{param}` in router but `{param}` or different path in test → 404

## TDD Compliance (Override Rule)

**This rule overrides the `aif-implement` skill's general «NEVER write tests» instruction whenever a plan task explicitly includes test files or TDD requirements.**

- When a plan task lists `.spec.ts`, `_test.go`, `_test.py`, `*_test.rs` or similar test files in its **«Files to create»** section, those test files MUST be created during implementation — they are not optional.
- When a plan task has a **«Unit tests (TDD Rule 5)»** section describing what to test, those tests MUST be written as part of the task, not deferred or skipped.
- **Priority:** Task-level requirements (explicit test files and TDD sections in the plan) take precedence over the general «NEVER write tests» rule in the agent skill.
- This applies to ALL phases and ALL languages (Go, Rust, Python, TypeScript/Vue).

## Verification Lint & Format Requirement

**This rule adds a project-specific static-analysis gate for `$aif-verify`.** It is cumulative with the base skill: changed-file checks remain useful for focused feedback, and the project gate below runs before a final `pass` result is emitted.

**Scope rule:** Checks run on ALL files in every affected service, not only changed files.
Changed-files-only checking is useful for fast feedback, but it is insufficient as the final verification gate because pre-existing IDE-visible errors can slip through.
When ANY file in a service is changed, run the full check on the ENTIRE service directory as the project gate.

### Validation Gates (per language)

**TypeScript/Vue:**
- `npx biome check --apply .` — lint + format entire service
- Services: `src/services/frontend/`, `src/services/publish-browse-ui/`

**Go (compilation → lint → deps):**
- `go vet ./...` — compilation + basic static analysis (MANDATORY, runs first)
- `gofmt -l .` — format check (MANDATORY)
- `golangci-lint run ./...` — advanced lint (MANDATORY, requires `.golangci.yml`)
- `go mod tidy && git diff --exit-code go.mod go.sum` — dependency drift check (MANDATORY)
- Services: `api-gateway`, `auth-service`, `commenting-service`, `ticket-api`, `ticket-notifier`,
  `ticket-sync`, `ticket-telemetry-listener`, `support-service`, `ai-orchestration-service`,
  `shared/llm`, `shared/proto`, `cli`

**Rust (compilation → lint → format):**
- `cargo check` — compilation gate (MANDATORY, catches namespacing/type errors that clippy misses)
- `cargo clippy --lib -- -D warnings` — lint gate (MANDATORY)
- `cargo fmt --check` — format gate (MANDATORY)
- Services: `ontology-service`, `versioning-service`, `publisher-service`, `public-browse-api`
- Also: `src/services/shared/src/` (shared Rust library)

**Python (lint → type-check):**
- `ruff check .` — lint + format (MANDATORY)
- `basedpyright .` — strict type checking (MANDATORY, catches type errors, import resolution, implicit imports)
- Services: `document-extractor`, `metrics-service`, `ticket-classifier`

### Failure Handling
- If any check fails, the verification report MUST list specific errors with file:line references.
- The verification gate status is `fail` if any available mandatory validation gate reports errors.
- Do NOT skip remaining gates after a failure — run all of them and report cumulative results.
- Exception: If a language toolchain is not available in the environment, emit `WARN [lint] <tool> not available — skipping` and continue with the remaining checks.

### Service Matrix Maintenance

When a new service is added to the monorepo (a new directory under `src/services/`),
the service lists in this file, in `skill-context/aif-verify/SKILL.md`, and in
`skill-context/aif-implement/SKILL.md` MUST be updated. Run:

```bash
find src/services -name go.mod -o -name Cargo.toml -o -name pyproject.toml -o -name package.json | sort
```

Cross-reference against the service lists and add any missing entries.

### Traceability Integrity

When artifacts are created, modified, or deleted during implementation, the
traceability matrix at `.ai-factory/traceability/traceability.ttl` must reflect
the change.

**Artifact taxonomy (directory → vdo class):**

| Directory / Pattern | vdo Class |
|---------------------|-----------|
| `specs/requirements/REQ-BIZ-*` | `vdo:BusinessRequirement` |
| `specs/requirements/REQ-FUN-*` | `vdo:FunctionalRequirement` |
| `specs/requirements/REQ-NFR-*` | `vdo:NonFunctionalRequirement` |
| `specs/requirements/REQ-CON-*` | `vdo:FunctionalRequirement` |
| `specs/requirements/REQ-USR-*` | `vdo:FunctionalRequirement` |
| `specs/user-stories/US-*` | `vdo:UserStory` |
| `specs/use-cases/UC-*` | `vdo:UseCase` |
| `specs/adr/ADR-*` | `vdo:ArchitectureDecisionRecord` |
| `specs/c4/*` | `vdo:DesignArtifact` |
| `specs/ui/*` | `vdo:DesignArtifact` |
| `specs/vision.md`, `specs/context.md`, `specs/stack.md`, `specs/glossary.md` | `vdo:Specification` |
| `src/services/<name>/` (whole directory) | `vdo:Service` |
| `src/services/<name>/**/*.{go,rs,py,ts,vue}` | `vdo:CodeArtifact` (skip if directory already covered) |
| `src/cli/` | `vdo:CLICommand` |
| `src/services/shared/proto/*.proto` | `vdo:CodeArtifact` |
| `src/templates/` | `vdo:CodeArtifact` |
| `tests/e2e/*`, `tests/cli/*` | `vdo:TestSuite` |
| `tests/security/*` | `vdo:TestContract` |
| `tests/ticket-api/*` | `vdo:TestContract` |
| `src/**/*_test.{go,rs}`, `src/**/test_*.py`, `src/**/*_test.py`, `src/**/*.{spec,test}.ts` | `vdo:TestSuite` / `vdo:TestContract` |
| `src/docs/antora/**/*.adoc` | `vdo:Documentation` |
| `deploy/docker-compose.yml` | `vdo:DeploymentConfig` |
| `deploy/helm/*` | `vdo:DeploymentConfig` |
| `deploy/keycloak/*` | `vdo:DeploymentConfig` |
| `deploy/ci/gitlab-ci.yml` | `vdo:CIPipeline` |
| `deploy/observability/*` | `vdo:ObservabilityConfig` |
| `design/*.pen` | `vdo:DesignArtifact` |
| `specs/adr/ADR-DES.SECURITY-*` | `vdo:SecurityBoundary` |

- **Service added/removed:** add/remove `vdo:Service` instance with matching `vdo:filePath` and `vdo:language`.
- **Source file added/removed:** add/remove `vdo:CodeArtifact` or related instance unless already covered by a directory-level `vdo:filePath`.
- **Renamed/relocated artifact:** update `vdo:filePath`.
- **Reference updated:** verify `vdo:implements`, `vdo:validates`, `vdo:tests`, `vdo:documents` IRIs all resolve to existing instances.
- **Check command:** `grep` for the artifact path in `traceability.ttl` — if not found and the change is structural, a new instance is needed.

This rule applies to `$aif-verify`, `$aif-implement`, and `$aif-plan`.

### Docker Compose & CI Gate

When adding a service, verify it is included in Docker Compose (`deploy/docker-compose.yml`)
and CI pipeline (`deploy/ci/gitlab-ci.yml`). Cross-reference service lists with
`deploy/docker-compose.yml` service entries.
