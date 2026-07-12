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
