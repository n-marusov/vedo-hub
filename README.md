# VEDO Hub
*Virtual Environment for Developing Ontologies*

> *GitHub for ontologies* — create, share, and collaborate on knowledge graphs with AI. **Built with [AI Factory](https://github.com/ai-factory).**

A free platform where knowledge graphs come to life. Describe your domain in plain words and AI builds the ontology. Edit it visually in the browser, collaborate with your team through Pull Requests, version it like Git, and share it with the world through a single link. No installs, no OWL expertise needed.

## Why VEDO Hub?

Ontologies today are **dead files on a desktop**. Engineers wait 45 seconds for a class to expand, teams email `.owl` files and merge them in Notepad++, 90% of newcomers quit Protégé in 10 minutes, and there is **no GitHub, no API, no community** for ontologies.

| | Before | VEDO Hub |
|---|--------|----------|
| 💨 **Speed** | 45 sec for 500 subclasses | <1 sec for 1M axioms |
| 👥 **Teamwork** | Email + manual merge | Git branches, Pull Requests, semantic diff |
| 🚪 **Onboarding** | Months of learning OWL | Describe in words → ontology in 5 min; upload a PDF/DOCX/XLSX → done in 1 min |
| 🔗 **Sharing** | Download `.owl`, open in desktop | Interactive graph in the browser, one link |
| 🔌 **Integration** | 40–80 h of custom code | REST + SPARQL + OpenAPI — connect in minutes |
| 🌍 **Community** | No stars, no forks, no profiles | Stars, forks, PRs, DOI, contributor graphs |

**What sets us apart:** free forever for public ontologies, AI-powered generation from text and documents, visual drag-and-drop SPARQL, Git-like versioning with semantic diff, full API-as-infrastructure, and GitLab-style hierarchical organizations. No other platform — Protégé, BioPortal, TopBraid, PoolParty — covers more than 3 of the 9 core needs. **VEDO Hub covers all 9.**

## Quick Start

```bash
git clone --recurse-submodules <repo-url> vedo-hub
cd vedo-hub && make docker-up
```

`make docker-up` builds all images and starts containers in one command. Open `http://localhost:3000` — the frontend is ready. The API Gateway is at `http://localhost:8080`.

**Prerequisites:** Docker 24+, Docker Compose v2+, Make.

---

## Developer Environment Setup

Preparation for native development (no Docker for the code itself):

### Toolchain Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| Go | 1.22+ | API Gateway, auth, ticket services |
| Rust | 1.77+ | ontology, versioning, publisher, public-browse |
| Node.js | 20+ | frontend toolchain (biome, lefthook, vite) |
| pnpm | 11+ | package manager (`corepack enable` or `npm i -g pnpm@11`) |
| Python | 3.12 | **via [uv](https://github.com/astral-sh/uv)** — see below |
| Make | — | build orchestrator |
| Lefthook | via pnpm | git hooks (`make install-hooks`) |

### 1. Install dependencies

```bash
# Frontend (also provides biome + lefthook binaries)
cd apps/services/frontend && pnpm install
cd ../../..

# Install git hooks (pre-commit + pre-push)
make install-hooks
```

### 2. Python via uv (important on Windows)

The project uses **uv** as the Python toolchain (`uv run python`, `uv sync`). Do **not** rely on a bare `python3` — on Windows it may resolve to the Microsoft Store stub and fail silently.

```bash
# uv manages the interpreter + venv automatically
uv run python --version   # resolves a real Python (3.12+)
```

The pre-push hooks (`check-json`, `check-yaml`) run through `uv run python` — if `uv` is missing, install it first: `curl -LsSf https://astral.sh/uv/install.sh | sh` (or `winget install astral-sh.uv`).

### 3. Environment configuration

Copy the dev environment file and adjust domain-specific overrides:

```bash
cp config/.env.example config/.env.dev   # then edit config/.env.dev
```

Defaults used by the stack (overrides must resolve in your environment):

- `KC_HOSTNAME=localhost` — Keycloak dev host; any override must resolve locally
- `KEYCLOAK_JWKS_URL` — Keycloak JWKS endpoint used by the API Gateway auth middleware

> For **E2E tests** use the test compose file, which sets `JWT_DEV_PUBLIC_KEY_PEM` on the API Gateway (self-signed tokens):

| Scenario | Compose File | Auth |
|----------|-------------|------|
| Development | `deploy/docker-compose.yml` | Keycloak JWKS |
| E2E Testing / CI | `deploy/docker-compose.test.yml` | self-signed JWT (`JWT_DEV_PUBLIC_KEY_PEM`) |

### 4. Verify setup

```bash
make help                 # all targets available
make lint-api-gateway     # Go toolchain works
cd apps/services/frontend && pnpm run lint  # Node/pnpm works
```

---

## Infrastructure

`make docker-up` launches **22 services** in 5 startup phases, gated by healthcheck dependencies. Cold-start: **~5–6 min** (limited by Keycloak).

### Startup Phases

| Phase | ⏱ Est. | Services |
|-------|--------|----------|
| **0 — Infra** | ~210s | `neo4j`, `postgres`, `redis`, `rabbitmq`, `minio`, `keycloak` |
| **1 — Core** | ~60s | `ontology-service`, `versioning-service`, `auth-service`, `metrics-service`, `publisher-service`, `commenting-service`, `ticket-api`, `ticket-notifier`, `ticket-classifier`, `ai-orchestration-service` |
| **2 — Dependent** | ~30s | `document-extractor`, `ticket-telemetry-listener`, `public-browse-api` |
| **3 — Gateway** | ~30s | `api-gateway` |
| **4 — Frontends** | ~25s | `frontend`, `publish-browse-ui` |

### Key Endpoints

| Service | Port | URL |
|---------|------|-----|
| Frontend | `3000` | http://localhost:3000 |
| API Gateway | `8080` | http://localhost:8080 |
| Published Viewer | `3002` | http://localhost:3002 |
| Keycloak | `8180` | http://localhost:8180 (`admin`/`admin`) |
| Neo4j Browser | `7474` | http://localhost:7474 |
| RabbitMQ Mgmt | `15672` | http://localhost:15672 |
| MinIO Console | `9001` | http://localhost:9001 |

### Infrastructure-Only Mode

```bash
make infra-up    # Start 6 infra services without the app stack
make infra-down  # Stop them
```

### Profile Services

Additional services behind `COMPOSE_PROFILE`:

| Profile | Adds |
|---------|------|
| `documentation` | 4 doc servers (ports 5000–5003) |
| `llm` | Ollama for local AI dev (port 11434) |
| `local` / `ci` | `vedo-cli-build` (compile-only check) |

```bash
make docker-up COMPOSE_PROFILE=llm   # Start with local LLM
```

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

**Infrastructure requirements:**

| Test type | Requires | Auto-started | Command |
|-----------|----------|-------------|---------|
| Unit (Rust/Go/Python/TS) | Nothing | — | `make test-unit-fast` |
| Rust integration (Neo4j) | Docker | ✅ `make test-integration-rust-fast` starts Neo4j | `make test-integration-rust-fast` |
| Versioning integration (PostgreSQL) | Docker + `pg_isready` | ✅ `make test-versioning-fast` starts Postgres | `make test-versioning-fast` |
| Go integration (ticket-api, org-api, auth) | Docker | ✅ `make test-integration-go-fast` starts PostgreSQL | `make test-integration-go-fast` |
| E2E API | Docker test stack | ✅ Playwright auto-starts compose | `make test-e2e-api-fast` |
| E2E GUI | Docker test stack + frontend | ✅ Playwright auto-starts compose | `make test-e2e-gui-fast` |
| Gates (contract, BOLA/BFLA unit, quality) | Go | ✅ Fully automated (static + unit) | `make test-gates-fast` |
| Security integration (BOLA/BFLA/RBAC full-stack) | Docker test stack | ✅ `make test-gates-security-fast` auto-starts stack | `make test-gates-security-fast` |

> Use `make infra-up` to start all infrastructure services (Neo4j, Postgres, Redis, RabbitMQ, Keycloak, MinIO) without the app stack.
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
