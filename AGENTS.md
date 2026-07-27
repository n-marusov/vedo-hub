# AGENTS.md

> Auto-generated project structure map for AI agents. Keep updated when project structure changes significantly.

## Project Overview

VEDO Core is a web platform for multi-user ontology (knowledge graph) editing with Git-like versioning, real-time collaboration, and REST/GraphQL APIs for integration. It replaces desktop Protégé with a cloud-based ontology editor supporting up to 1 million axioms.

See `.ai-factory/DESCRIPTION.md` for detailed project specification.

## Tech Stack

- **Frontend:** Vue 3 + TypeScript, Vite, Apollo Client, Vue Flow
- **Backend:** Go (API Gateway, Auth, Ticket services), Rust (Ontology, Versioning, Publisher), Python (Metrics, Classifier, Document Extractor)
- **Databases:** Neo4j, PostgreSQL, Redis, RabbitMQ
- **Authentication:** Keycloak (OIDC/OAuth2)
- **Observability:** OpenTelemetry, Grafana (Tempo, Prometheus, Loki)
- **Infrastructure:** Docker Compose, Helm, GitLab CI

## Project Structure

```
.
├── .ai-factory/          # AI Factory configuration and artifacts
├── .agents/              # AI Factory agent skills (installed via npx skills)
├── apps/                 # Application source code
│   ├── services/         #   Microservices (15 services)
│   │   ├── api-gateway/         # Go — REST API Gateway
│   │   ├── auth-service/        # Go — Authentication & authorization
│   │   ├── commenting-service/  # Go — Commenting & collaboration
│   │   ├── document-extractor/   # Python — AI document extraction & NL→OWL
│   │   ├── ai-orchestration-service/ # Go — LLM orchestration, policy, audit, templates
│   │   ├── frontend/            # Vue 3 — Web UI
│   │   ├── metrics-service/     # Python — Metrics & analytics
│   │   ├── ontology-service/    # Rust — Ontology CRUD & Neo4j
│   │   ├── public-browse-api/   # Rust — Public API
│   │   ├── publish-browse-ui/   # Vue 3 — Published ontology viewer
│   │   ├── publisher-service/   # Rust — Ontology publishing
│   │   ├── shared/              # Shared libraries (polyglot)
│   │   │   ├── llm/             #   Go — LLM provider abstraction, templates, observability
│   │   │   │   ├── providers/   #     OpenAI-compatible & Anthropic adapters
│   │   │   │   └── templates/  #     Prompt templates (.tmpl)
│   │   │   ├── proto/           #   gRPC protobuf definitions (buf)
│   │   │   └── src/             #   Rust shared library (health, metrics, tracing, protos)
│   │   ├── support-service/     # Go — Support operations
│   │   ├── ticket-api/          # Go — Ticket management API
│   │   ├── ticket-classifier/   # Python — Ticket auto-classification
│   │   ├── ticket-notifier/     # Go — Ticket notifications
│   │   ├── ticket-sync/         # Go — Ticket synchronization
│   │   ├── ticket-telemetry-listener/ # Go — Telemetry ingestion
│   │   └── versioning-service/  # Rust — Git-like versioning engine
│   └── vedo-cli/         #   vedo-cli Go module (operator CLI tool)
│       ├── cmd/ticket/   #     Ticket management commands
│       ├── commands/     #     Command implementations
│       └── internal/     #     Internal packages (audit, auth)
├── config/               # Environment configuration
│   ├── .env.dev          #   Dev environment — default ports
│   ├── .env.test         #   Test environment — ports +10000
│   └── .env.staging      #   Staging environment — ports +20000
├── deploy/               # Deployment configuration
│   ├── ci/               #   GitLab CI pipeline (gitlab-ci.yml)
│   ├── helm/             #   Helm chart values
│   ├── keycloak/         #   Keycloak realm configuration
│   ├── observability/    #   Grafana, Loki, Prometheus, Tempo, OTEL configs
│   └── postgres/         #   PostgreSQL init scripts (org database)
├── design/               # UI/UX design files (Pencil .pen format)
│   ├── pages/            #   Page-level design files
│   ├── frontend.pen      #   Main frontend design
│   └── ui-kit.lib.pen    #   UI component library
├── docs/                 # Documentation
│   └── antora/           #   Antora documentation source
├── specs/                # Project specifications (git submodule)
│   ├── adr/              #   Architecture Decision Records
│   ├── c4/               #   C4 model diagrams
│   ├── requirements/     #   Functional and non-functional requirements
│   ├── ui/               #   UI specifications
│   ├── use-cases/        #   Use case specifications
│   └── user-stories/     #   User stories
├── tests/                # Test suites
│   ├── gates/            #   CI quality gates
│   ├── suites/           #   Integration test suites
│   ├── e2e/playwright/   #   Playwright end-to-end tests
│   └── security/         #   Authorization test suites (BOLA/BFLA)
├── tools/                # Build and development tools
│   ├── build/            #   Makefile build includes (docker.mk, go.mk, rust.mk, etc.)
│   ├── dockerfiles/      #   Multi-stage Dockerfile templates by language
│   └── scaffolds/        #   Service scaffolds by language
├── .ai-factory.json      # Codex agent configuration (installed skills, MCP)
├── .dockerignore         # Docker build context exclusions
├── .gitignore            # Git ignore rules
├── .gitmodules           # Git submodule (specs repo)
└── LICENSE               # Project license
```

## Key Entry Points

| File | Purpose |
|------|---------|
| `apps/services/api-gateway/main.go` | API Gateway entry point — routes all external requests |
| `apps/services/frontend/index.html` | Frontend app entry point |
| `apps/services/ontology-service/src/main.rs` | Ontology service — core graph operations |
| `apps/services/versioning-service/src/main.rs` | Versioning engine — Git-like commits/branches |
| `apps/services/document-extractor/main.py` | Document extractor — AI-assisted ontology extraction from documents |
| `apps/services/metrics-service/main.py` | Metrics & analytics service |
| `apps/vedo-cli/command.go` | CLI command dispatcher |
| `Makefile` | Root build orchestrator |
| `deploy/docker-compose.yml` | All-services Docker Compose (28 services) |
| `deploy/ci/gitlab-ci.yml` | GitLab CI pipeline definition |

## Documentation

Antora documentation source in `docs/antora/`. Build with `cd docs/antora && npx antora antora-playbook.yml`.

| Document | Path | Description |
|----------|------|-------------|
| README | `README.md` | Project landing page |
| Deployment | `docs/antora/developer-guide/modules/ROOT/pages/deployment.adoc` | Docker setup, environments, compose |
| Development | `docs/antora/developer-guide/modules/ROOT/pages/development.adoc` | Native builds, linting, dev tools |
| Configuration | `docs/antora/developer-guide/modules/ROOT/pages/configuration.adoc` | Environment variables, ports, LLM |
| Architecture | `docs/antora/developer-guide/modules/ROOT/pages/architecture.adoc` | Project structure, microservices |
| Testing | `docs/antora/developer-guide/modules/ROOT/pages/testing.adoc` | Unit, integration, E2E, CI |
| User Guide | `docs/antora/user-guide/` | Quick start, ontology editing, versioning |
| Developer Guide (home) | `docs/antora/developer-guide/modules/ROOT/pages/index.adoc` | Developer documentation hub |
| Org Model | `docs/antora/developer-guide/modules/ROOT/pages/organization-model.adoc` | Multi-team organization model |
| Admin Guide | `docs/antora/admin-guide/` | Deployment, port reference, observability, security |
| Integrator Guide | `docs/antora/integrator-guide/` | API reference, authentication, integration config |
| Antora Playbook | `docs/antora/antora-playbook.yml` | Site build configuration |
| Env: dev | `config/.env.dev` | Dev environment — default ports |
| Env: test | `config/.env.test` | Test environment — ports +10000 |
| Env: staging | `config/.env.staging` | Staging environment — ports +20000 |
| License | `LICENSE` | Project license |

## AI Context Files

| File | Purpose |
|------|---------|
| AGENTS.md | This file — structural map for AI agents |
| .ai-factory/DESCRIPTION.md | Project specification and tech stack |
| .ai-factory/ARCHITECTURE.md | Architecture patterns and guidelines |
| .ai-factory/config.yaml | AI Factory runtime configuration |
| .ai-factory/rules/base.md | Codebase conventions and rules |
| .ai-factory.json | Codex agent configuration (installed skills, MCP servers) |

## Agent Rules

- Decompose shell commands into separate steps for reliability: do not chain `cd` and `git` operations with `&&`.
  - ❌ Incorrect: `git checkout main && git pull`
  - ✅ Correct: First `git checkout main`, then `git pull origin main`
- Read relevant files before making changes — understand the full context before editing.
- Run tests associated with changed code after making modifications.
- Use `$aif-plan` for new features and `$aif-fix` for bug fixes before implementation.
