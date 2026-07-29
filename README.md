# VEDO Hub

> *GitHub for ontologies* — a social hub, semantic API layer, and LLM generation in one product. **Built with [AI Factory](https://github.com/ai-factory).**

A SaaS platform where ontologies become living assets — forked, starred, reviewed via Pull Requests, cited through DOI, published and consumed through an API. Under the hood: a polyglot microservices architecture in Go, Rust, Python, and Vue 3, with Git-like ontology versioning, Neo4j for graph storage, and Keycloak for authentication.

## Quick Start

```bash
git clone --recurse-submodules <repo-url> vedo-hub
cd vedo-hub && make docker-up
```

`make docker-up` builds all images and starts containers in one command. Open `http://localhost:3000` — the frontend is ready. The API Gateway is at `http://localhost:8080`.

**Prerequisites:** Docker 24+, Docker Compose v2+, Make.

---

## Common Tasks

```bash
make help                    # Show all available targets
```

### Deploy

```bash
make docker-up               # Dev environment (default)
make docker-up-test          # Test environment (JWT dev keys for E2E)
make docker-up-staging       # Staging environment
make docker-down-dev         # Stop dev
make docker-down-test        # Stop test
make docker-down-staging     # Stop staging
make docker-build            # Build all images without starting containers
make docker-build-api-gateway # Build a single service image
make infra-up                # Start infrastructure only (Neo4j, Postgres, Redis, etc.)
```

### Native Build (no Docker)

```bash
make build                   # Build all services natively
make build-api-gateway       # Build a single service (auto-detects language)
make build-ontology-service  # Rust
make build-frontend          # TypeScript
make vendor-go               # Populate vendor/ for offline Docker builds
```

### Lint & Format

```bash
make lint                    # All linters
make lint-api-gateway        # Single service
make fmt                     # Format all source code
make typecheck               # TypeScript type checking (vue-tsc)
```

### Test

Every test target has two explicit modes:

| Mode | Unit | Integration | E2E API | E2E GUI | Gates | All |
|------|------|-------------|---------|---------|-------|-----|
| ⚡ Fail-fast | `make test-unit-fast` | `make test-integration-fast` | `make test-e2e-api-fast` | `make test-e2e-gui-fast` | `make test-gates-fast` | `make test-fast` |
| 📊 Full statistics | `make test-unit-full` | `make test-integration-full` | `make test-e2e-api-full` | `make test-e2e-gui-full` | `make test-gates-full` | `make test-full` |

Additional targets:

| Target | Description |
|--------|-------------|
| `make test-versioning-fast` / `-full` | Versioning-service tests (auto-starts PostgreSQL) |
| `make coverage` | Test coverage (Go only) |

> ⚡ **Fail-fast** stops at the first failure — ideal for development iteration.
> 📊 **Full statistics** runs everything and collects all failures — use for CI nightly or release.
> Integration and E2E tests require Docker (Neo4j, Postgres, test compose stack).
> Use `ENV=test` for E2E: `make docker-up-test` starts the test stack with self-signed JWT tokens.

### CI

| Pipeline | Command | Description |
|----------|---------|-------------|
| ⚡ Fast CI | `make ci-fast` | proto + build + lint + test-fast + typecheck |
| 📦 Full CI | `make ci-full` | Full pipeline + integration + E2E + gates + quality |

> Both CI targets are fail-fast — they stop at the first failure to save pipeline minutes.

### Utility

```bash
make list-services           # Show all services grouped by language
make status                  # Show health of each service in the compose stack
make dev-frontend            # Frontend dev server with hot reload
make dev-api                 # API gateway with go run
make clean                   # Remove build artifacts
make clean-all               # Deep clean (+ vendor/, .venv, node_modules)
make docker-logs             # Tail service logs
make docker-shell SVC=<name> # Open shell in a service container
make install-hooks           # Install Git hooks via Lefthook
```

> **Environment override:** add `ENV=test` or `ENV=staging` to any `docker-*` target.
> **Port override:** prefix with `API_GATEWAY_PORT=9090 make docker-up` etc.

---

## Documentation

Full documentation is available as an [Antora](https://antora.org/) site. Build locally:

```bash
cd docs/antora && npx antora antora-playbook.yml
# Output: docs/antora/build/site/index.html
```

| Guide | Description |
|-------|-------------|
| [Deployment](docs/antora/developer-guide/modules/ROOT/pages/deployment.adoc) | Docker setup, environments, container management |
| [Development](docs/antora/developer-guide/modules/ROOT/pages/development.adoc) | Native builds, per-service, linting, tools |
| [Configuration](docs/antora/developer-guide/modules/ROOT/pages/configuration.adoc) | Environment variables, ports, LLM, multi-env |
| [Architecture](docs/antora/developer-guide/modules/ROOT/pages/architecture.adoc) | Project structure, microservices |
| [Testing](docs/antora/developer-guide/modules/ROOT/pages/testing.adoc) | Unit, integration, E2E, CI pipeline |

## Key Features

- **Ontology Editor** — visual editor for knowledge graphs with <1s response time (p95) for large ontologies
- **Git-like Versioning** — commits, branches, merges, diffs via a dedicated Versioning Service
- **Real-time Collaboration** — WebSocket-powered node locking, comments, notifications
- **Multi-team Organization** — GitLab-like groups, ontologies, membership inheritance, visibility levels
- **LLM Integration** — document extraction, AI orchestration with OpenAI/Anthropic/Ollama
- **REST API & GraphQL** — external integration and frontend data operations

## Development with AI Factory

The project uses [AI Factory](https://github.com/ai-factory), an agentic development framework:

| Command | Purpose |
|---------|---------|
| `$aif-plan` | Create a feature implementation plan |
| `$aif-implement` | Execute tasks from the plan |
| `$aif-review` | Conduct code review |
| `$aif-verify` | Verify implementation completeness |
| `$aif-fix` | Fix a bug with root-cause analysis |
| `$aif-docs` | Generate documentation |
| `$aif-test-quality` | Assess test quality (TQS / RCS) |

> More on AI Factory: `.ai-factory/` and [ai-factory](https://github.com/ai-factory).

---

## License

MIT — see [LICENSE](LICENSE).
