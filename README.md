# VEDO Hub

> *GitHub for ontologies* — a social hub, semantic API layer, and LLM generation in one product. **Built with [AI Factory](https://github.com/ai-factory).**

**VEDO Hub** is an open ontology community and infrastructure layer for knowledge-driven applications. A SaaS platform where ontologies become living assets — forked, starred, reviewed via Pull Requests, cited through DOI, published and consumed through an API. Unlike desktop Protégé, VEDO Hub is accessible to everyone, not just knowledge engineers, through NLP-driven generation and visual SPARQL.

Under the hood: a polyglot microservices architecture in Go, Rust, Python, and Vue 3, with Git-like ontology versioning, Neo4j for graph storage, and Keycloak for authentication.

---

## Quick Start

```bash
git clone --recurse-submodules <repo-url> vedo-hub
cd vedo-hub/src && make docker-up
```

Open `http://localhost:3000` — the frontend is ready. The API Gateway is at `http://localhost:8080`.

**Prerequisites:** Docker 24+, Docker Compose v2+, Make. Other toolchains (Go, Rust, Python, Node.js) are only needed for local development outside containers.

---

## Docker Build

`make docker-up` is all you need — it builds images **and** starts containers in one command. No separate build step required.

`make docker-build` is only needed when you want images **without** starting containers (CI, registry push, pre-building):

```bash
# Build all service images
cd src && make docker-build

# By language
make docker-build-rust       # ontology, versioning, publisher, public-browse
make docker-build-go         # api-gateway, auth, commenting, tickets, AI
make docker-build-python     # metrics, classifier, document-extractor
make docker-build-typescript # frontend, publish-browse-ui
```

Images are tagged `vedo-core/<service>:latest`. The build uses the same context (project root) and same Dockerfiles as `docker compose build` — functionally identical.

> **Go services** require a `vendor/` directory. Run `make vendor-go` to populate it. Services without `vendor/` are skipped with a notice.
> **First run** of `make docker-up` or `make docker-build` builds all images from source — can take 20–40 minutes.

---

## Environment Variables

All variables in `deploy/docker-compose.yml` use `${VAR:-default}` syntax — every setting has a sensible default and can be overridden via shell environment or a `.env` file. **No `.env` file is required for development** — `make docker-up` works out of the box.

### Multi-Environment Setup

The project ships three pre-configured env files with non-overlapping ports, so you can run dev, test, and staging stacks in parallel:

| Environment | Env file | Port offset | COMPOSE_PROJECT_NAME |
|-------------|----------|-------------|---------------------|
| **dev** | `.env.dev` | none (defaults) | `vedo-core-dev` |
| **test** | `.env.test` | +10000 | `vedo-core-test` |
| **staging** | `.env.staging` | +20000 | `vedo-core-staging` |

```bash
# Dev (default)
make docker-up

# Test — runs alongside dev, no port conflicts
make docker-up ENV=test

# Staging — runs alongside dev and test
make docker-up ENV=staging
```

Or equivalently with raw Docker Compose from the project root:

```bash
cp .env.dev .env && docker compose -f deploy/docker-compose.yml up -d
docker compose --env-file .env.test -f deploy/docker-compose.yml up -d
docker compose --env-file .env.staging -f deploy/docker-compose.yml up -d
```

Each env file defines separate host ports while keeping container ports identical — inter-service communication always works through the Docker network. See `deploy/README.md` for the full port reference and advanced usage.

### Ways to configure

| Method | Example | Best for |
|--------|---------|----------|
| **Shell env** | `export API_GATEWAY_PORT=9090 && make docker-up` | One-off overrides |
| **`.env` file** | Create `deploy/.env` with `API_GATEWAY_PORT=9090` | Persistent project config |
| **Inline** | `API_GATEWAY_PORT=9090 make docker-up` | Temporary overrides |

### Essential overrides

| Variable | Default | Description |
|----------|---------|-------------|
| `LLM_PROVIDER` | *(empty — disabled)* | LLM backend: `openai`, `anthropic`, `ollama` |
| `LLM_API_KEY` | *(empty)* | API key for the chosen provider |
| `LLM_MODEL` | *(empty)* | Model name (e.g. `gpt-4o`, `claude-3-opus`, `llama3`) |
| `LLM_BASE_URL` | *(empty)* | Custom OpenAI-compatible endpoint |
| `LLM_PROVIDERS_OLLAMA_BASE_URL` | `http://localhost:11434/v1` | Ollama API endpoint |
| `LLM_PROVIDERS_OLLAMA_MODEL` | `llama3` | Ollama model name |
| `GRPC_TLS_ENABLED` | `false` | Enable mTLS for all gRPC services |
| `GRPC_CA_CERT_FILE` | *(empty)* | CA certificate path (api-gateway) |
| `GRPC_TLS_CERT_FILE` | *(empty)* | TLS certificate path (Rust services) |
| `GRPC_TLS_KEY_FILE` | *(empty)* | TLS key path (Rust services) |

### AI features (LLM)

Document extraction and AI orchestration require an LLM provider. Use the bundled Ollama profile for local development:

```bash
# Start VEDO + local Ollama (pulls ollama/ollama image, ~4 GB)
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.llm.yaml --profile llm up -d

# Pull a model into the running Ollama container
docker compose -f deploy/docker-compose.llm.yaml exec ollama ollama pull llama3

# Tell services to use Ollama
export LLM_PROVIDER=ollama
export LLM_PROVIDERS_OLLAMA_BASE_URL=http://ollama:11434/v1
export LLM_PROVIDERS_OLLAMA_MODEL=llama3
```

Or use a cloud provider:

```bash
export LLM_PROVIDER=openai
export LLM_API_KEY=sk-...
export LLM_MODEL=gpt-4o
make docker-up
```

### Service ports

Every host port can be overridden. Container ports are fixed across all environments — see `deploy/README.md` for the full table. Key host ports (dev defaults):

| Variable | Default |
|----------|---------|
| `FRONTEND_PORT` | `3000` |
| `PUBLISH_BROWSE_UI_PORT` | `3002` |
| `API_GATEWAY_PORT` | `8080` |
| `AUTH_SERVICE_GRPC_PORT` | `9003` |
| `VERSIONING_SERVICE_GRPC_PORT` | `9002` |
| `METRICS_SERVICE_REST_PORT` | `8084` |
| `COMMENTING_SERVICE_GRPC_PORT` | `9004` |
| `PUBLISHER_SERVICE_GRPC_PORT` | `9005` |
| `PUBLIC_BROWSE_GRPC_PORT` | `9011` |
| `TICKET_API_REST_PORT` / `TICKET_API_GRPC_PORT` | `8088` / `9010` |
| `TICKET_CLASSIFIER_PORT` | `8089` |
| `TICKET_TELEMETRY_STUB_PORT` | `8090` |
| `TICKET_NOTIFIER_STUB_PORT` | `8091` |
| `DOCUMENT_EXTRACTOR_GRPC_PORT` | `9013` |
| `NEO4J_HTTP_PORT` / `NEO4J_BOLT_PORT` | `7474` / `7687` |
| `POSTGRES_PORT` | `5432` |
| `REDIS_PORT` | `6379` |
| `RABBITMQ_AMQP_PORT` / `RABBITMQ_MGMT_PORT` | `5672` / `15672` |
| `MINIO_API_PORT` / `MINIO_CONSOLE_PORT` | `9000` / `9001` |
| `KC_HOSTNAME_PORT` | `8180` |

### Local native development

Running a service **outside Docker** (for debugging / hot-reload): start only infrastructure, then point the service at `localhost`:

```bash
# Start infra only
docker compose -f deploy/docker-compose.yml up -d neo4j postgres redis rabbitmq keycloak minio

# Example: run API Gateway natively
export API_GATEWAY_PORT=8080
export ONTOLOGY_SERVICE_URL=http://localhost:8082
export AUTH_SERVICE_URL=http://localhost:8081
export KEYCLOAK_JWKS_URL=http://localhost:8180/realms/vedo-core/protocol/openid-connect/certs
cd src/services/api-gateway && go run .
```

Each service reads its config from env vars — see the `environment:` block for that service in `deploy/docker-compose.yml`, or the service's entry point (`main.go` / `main.rs` / `main.py`).

### Test environment

```bash
# Start full test stack (includes JWT_DEV_PUBLIC_KEY_PEM for test auth)
docker compose -f deploy/docker-compose.test.yml up -d
```

`docker-compose.test.yml` extends the main compose via `include` and adds a dev JWT public key to the API Gateway — so E2E tests can use pre-signed RS256 tokens instead of requiring a running Keycloak for auth.

### Tests

| Test suite | Requirements |
|------------|--------------|
| **Unit tests** (`make test`) | Language toolchains only |
| **Contract tests** (`bash tests/test_contract_gate.sh`) | Go toolchain |
| **E2E — API** (`pnpm exec playwright test --config=playwright.api.config.ts`) | Test Docker stack |
| **E2E — GUI** (`pnpm exec playwright test --config=playwright.gui.config.ts`) | Test Docker stack + frontend |
| **Security tests** (`bash tests/test_bola_bfla_gate.sh`) | Running Docker stack |

---

## Project Structure

```
vedo-hub/
├── src/
│   ├── services/               # 15+ microservices
│   │   ├── api-gateway/             # Go — REST API Gateway
│   │   ├── ontology-service/        # Rust — Graph CRUD (Neo4j)
│   │   ├── versioning-service/      # Rust — Git-like versioning
│   │   ├── auth-service/            # Go — Keycloak integration, RBAC
│   │   ├── frontend/                # Vue 3 — Web UI
│   │   ├── document-extractor/      # Python — Document-to-ontology extraction
│   │   ├── publisher-service/       # Rust — Ontology publishing
│   │   ├── metrics-service/         # Python — Metrics & analytics
│   │   ├── commenting-service/      # Go — Comments & collaboration
│   │   ├── public-browse-api/       # Rust — Public read-only API
│   │   ├── publish-browse-ui/       # Vue 3 — Published ontology viewer
│   │   ├── ai-orchestration-service/# Go — LLM orchestration, policy, audit
│   │   ├── support-service/         # Go — Support operations
│   │   ├── ticket-api/              # Go — Ticket management
│   │   ├── ticket-classifier/       # Python — Auto-classification
│   │   ├── ticket-notifier/         # Go — Ticket notifications
│   │   ├── ticket-sync/             # Go — External system sync
│   │   ├── ticket-telemetry-listener# Go — Telemetry ingestion
│   │   ├── shared/                  # Shared libs: proto, llm, rust
│   │   └── ...                      # Additional Go/Python services
│   ├── build/                  # Makefile includes per language
│   └── docs/antora/            # Documentation (Antora / AsciiDoc)
├── deploy/
│   ├── docker-compose.yml      # Full-stack orchestration (28 containers)
│   ├── docker-compose.test.yml # Test stack overlay (adds JWT_DEV_PUBLIC_KEY_PEM)
│   ├── ci/gitlab-ci.yml        # CI/CD pipeline
│   ├── keycloak/               # Realm configuration
│   └── observability/          # Grafana, Tempo, Prometheus, Loki
├── tests/                      # Integration & E2E tests
└── specs/                      # Specifications (submodule)
```

---

## Development with AI Factory

The entire development lifecycle is driven by **AI Factory** — an agentic development framework:

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

## Testing

### Unit Tests (by language)

No test environment required — tests run natively with language toolchains.

```bash
# All tests
cd src && make test

# By language
make test-rust         # cargo test
make test-go           # go test ./...
make test-python       # pytest
make test-typescript   # pnpm test
```

### API / Contract Tests

Requires **Docker** + **Go** toolchain. Validates service contracts, compose syntax, and runs Go module tests.

```bash
# Full contract test run (Go services + bash scripts)
bash tests/test_contract_gate.sh
```

### E2E Tests (Playwright)

Two-phase execution — **API** runs first (fast feedback), **GUI** runs second and stops on the first failure.

Playwright configs automatically start the backend using `deploy/docker-compose.test.yml`,
which overrides the API Gateway to accept self-signed test JWT tokens instead of requiring Keycloak.

```bash
cd tests/e2e/playwright
pnpm install

# Phase 1: API tests (REST, GraphQL, SPARQL)
pnpm exec playwright test --config=playwright.api.config.ts

# Phase 2: GUI tests (pages, user stories, a11y)
pnpm exec playwright test --config=playwright.gui.config.ts

# Or run both via the CI gate script
bash run-ci.sh
```

Test structure:

```
tests/
├── api/
│   ├── rest/          # REST API integration, SPARQL/Cypher queries, gateway
│   └── graphql/       # GraphQL tests
├── gui/
│   ├── pages/         # Screen rendering, a11y, page wiring
│   └── user-stories/  # Full workflows: ontology lifecycle, AI, document extraction
├── fixtures.ts        # Base test fixtures
├── graphql-fixtures.ts# GUI GraphQL mock data
├── jwt-tokens.ts      # Pre-signed RS256 JWT tokens
└── ontology-test-data.ts
```

> **First run:** building Docker images from source can take 20–40 minutes.
> Build images beforehand: `docker compose -f deploy/docker-compose.test.yml build`

GUI tests use `maxFailures: 1` — the run stops immediately on the first failure for fast feedback. API tests use `retries: 2` for robustness against transient issues.

#### Running a single GUI test

Use the Playwright `--grep` (`-g`) flag to filter by test name, or pass a file path to run a single spec file:

```bash
cd tests/e2e/playwright
pnpm exec playwright test --config=playwright.gui.config.ts -g "Dashboard Navigation"

# Or by a single test name within a suite
pnpm exec playwright test --config=playwright.gui.config.ts -g "should navigate to ontology workspace"

# Run a single file by path
pnpm exec playwright test --config=playwright.gui.config.ts tests/gui/pages/dashboard-navigation.spec.ts

# Run a single user story file
pnpm exec playwright test --config=playwright.gui.config.ts tests/gui/user-stories/org-lifecycle.spec.ts
```

When grepping for a subset of tests (e.g., running all user-stories), override `maxFailures: 1` so the run doesn't stop at the first failure:

```bash
pnpm exec playwright test --config=playwright.gui.config.ts tests/gui/user-stories/ --max-failures=0
```

### Security Tests (BOLA/BFLA)

Requires **Docker** stack to be running. Validates authorization boundaries at the API Gateway level.

```bash
bash tests/test_bola_bfla_gate.sh
```

---

## Documentation (Antora)

Documentation lives in `src/docs/antora/` and includes 5 guides:

| Guide | Local Address |
|-------|---------------|
| User Guide | `http://localhost:5000` |
| Developer Guide | `http://localhost:5001` |
| Admin Guide | `http://localhost:5002` |
| Integrator Guide | `http://localhost:5003` |

### Build Locally

```bash
cd src/docs/antora
npx antora antora-playbook.yml
# Output: src/docs/antora/build/site/index.html
```

Or via Docker Compose:

```bash
docker compose -f deploy/docker-compose.docs.yaml --profile documentation up -d
```

---

## License

MIT — see [LICENSE](LICENSE).
