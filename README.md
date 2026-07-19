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

## Environment Variables

All variables in `deploy/docker-compose.yml` use `${VAR:-default}` syntax — every setting has a sensible default and can be overridden via shell environment or a `.env` file. **No `.env` file is required for development** — `make docker-up` works out of the box.

### Ways to configure

| Method | Example | Best for |
|--------|---------|----------|
| **Shell env** | `export API_GATEWAY_PORT=9090` then `docker compose up` | One-off overrides |
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

Every service port can be overridden with a dedicated variable. Useful when conflicts arise:

| Variable | Default |
|----------|---------|
| `API_GATEWAY_PORT` | `8080` |
| `AUTH_SERVICE_REST_PORT` | `8081` |
| `ONTOLOGY_SERVICE_REST_PORT` | `8082` |
| `VERSIONING_SERVICE_REST_PORT` | `8083` |
| `METRICS_SERVICE_REST_PORT` | `8084` |
| `COMMENTING_SERVICE_REST_PORT` | `8085` |
| `PUBLISHER_SERVICE_REST_PORT` | `8086` |
| `PUBLIC_BROWSE_REST_PORT` | `8087` |
| `TICKET_API_REST_PORT` | `8088` |
| `TICKET_CLASSIFIER_PORT` | `8089` |
| `TICKET_TELEMETRY_STUB_PORT` | `8090` |
| `TICKET_NOTIFIER_STUB_PORT` | `8091` |
| `DOCUMENT_EXTRACTOR_HTTP_PORT` | `8092` |
| `AI_ORCH_HEALTH_PORT` | `8093` |
| `NEO4J_HTTP_PORT` / `NEO4J_BOLT_PORT` | `7474` / `7687` |
| `POSTGRES_PORT` | `5432` |
| `REDIS_PORT` | `6379` |
| `RABBITMQ_AMQP_PORT` / `RABBITMQ_MGMT_PORT` | `5672` / `15672` |
| `KC_HOSTNAME_PORT` | `8180` |
| `MINIO_API_PORT` / `MINIO_CONSOLE_PORT` | `9000` / `9001` |

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

Two-phase execution — **API** runs first (fast feedback), **GUI** runs second and stops on the first failure:

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
