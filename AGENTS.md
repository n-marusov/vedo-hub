# AGENTS.md

> Auto-generated project structure map for AI agents. Keep updated when project structure changes significantly.

## Project Overview

VEDO Core is a web platform for multi-user ontology (knowledge graph) editing with Git-like versioning, real-time collaboration, and REST/GraphQL APIs for integration. It replaces desktop Protégé with a cloud-based ontology editor supporting up to 1 million axioms.

See `.ai-factory/DESCRIPTION.md` for detailed project specification.

## Tech Stack

- **Frontend:** Vue 3 + TypeScript, Vite, Apollo Client, Vue Flow
- **Backend:** Go (API Gateway, Auth, Ticket services), Rust (Ontology, Versioning, Publisher), Python (Metrics, Classifier)
- **Databases:** Neo4j, PostgreSQL, Redis, RabbitMQ
- **Authentication:** Keycloak (OIDC/OAuth2)
- **Observability:** OpenTelemetry, Grafana (Tempo, Prometheus, Loki)
- **Infrastructure:** Docker Compose, Helm, GitLab CI

## Project Structure

```
.
├── .ai-factory/          # AI Factory configuration and artifacts
├── .agents/              # AI Factory agent skills (installed via npx skills)
├── deploy/               # Deployment configuration
│   ├── ci/               #   GitLab CI pipeline (gitlab-ci.yml)
│   ├── helm/             #   Helm chart values
│   ├── keycloak/         #   Keycloak realm configuration
│   └── observability/    #   Grafana, Loki, Prometheus, Tempo, OTEL configs
├── design/               # UI/UX design files (Pencil .pen format)
│   ├── pages/            #   Page-level design files
│   ├── frontend.pen      #   Main frontend design
│   └── ui-kit.lib.pen    #   UI component library
├── specs/                # Project specifications (git submodule)
│   ├── adr/              #   Architecture Decision Records
│   ├── c4/               #   C4 model diagrams
│   ├── requirements/     #   Functional and non-functional requirements
│   ├── ui/               #   UI specifications
│   ├── use-cases/        #   Use case specifications
│   └── user-stories/     #   User stories
├── src/                  # Source code root
│   ├── build/            #   Makefile build includes (docker.mk, go.mk, rust.mk, etc.)
│   ├── cli/              #   vedo-cli Go module (operator CLI tool)
│   │   ├── cmd/ticket/   #     Ticket management commands
│   │   ├── commands/     #     Command implementations
│   │   └── internal/     #     Internal packages (audit, auth)
│   ├── docker/           #   Multi-stage Dockerfile templates by language
│   ├── docs/antora/      #   Antora documentation source
│   ├── scripts/          #   Utility scripts
│   ├── services/         #   Microservices (14 services)
│   │   ├── api-gateway/         # Go — REST API Gateway
│   │   ├── auth-service/        # Go — Authentication & authorization
│   │   ├── commenting-service/  # Go — Commenting & collaboration
│   │   ├── frontend/            # Vue 3 — Web UI
│   │   ├── metrics-service/     # Python — Metrics & analytics
│   │   ├── ontology-service/    # Rust — Ontology CRUD & Neo4j
│   │   ├── public-browse-api/   # Rust — Public API
│   │   ├── publish-browse-ui/   # Vue 3 — Published ontology viewer
│   │   ├── publisher-service/   # Rust — Ontology publishing
│   │   ├── support-service/     # Go — Support operations
│   │   ├── ticket-api/          # Go — Ticket management API
│   │   ├── ticket-classifier/   # Python — Ticket auto-classification
│   │   ├── ticket-notifier/     # Go — Ticket notifications
│   │   ├── ticket-sync/         # Go — Ticket synchronization
│   │   ├── ticket-telemetry-listener/ # Go — Telemetry ingestion
│   │   └── versioning-service/  # Rust — Git-like versioning engine
│   └── templates/        #   Service scaffolds by language
├── tests/                # Test suites
│   ├── cli/              #   CLI integration tests
│   ├── e2e/playwright/   #   Playwright end-to-end tests
│   ├── security/         #   Authorization test suites (BOLA/BFLA)
│   └── ticket-api/       #   Ticket API tests
├── .ai-factory.json      # Codex agent configuration (installed skills, MCP)
├── .dockerignore         # Docker build context exclusions
├── .gitignore            # Git ignore rules
├── .gitmodules           # Git submodule (specs repo)
└── LICENSE               # Project license
```

## Key Entry Points

| File | Purpose |
|------|---------|
| `src/services/api-gateway/main.go` | API Gateway entry point — routes all external requests |
| `src/services/frontend/index.html` | Frontend app entry point |
| `src/services/ontology-service/src/main.rs` | Ontology service — core graph operations |
| `src/services/versioning-service/src/main.rs` | Versioning engine — Git-like commits/branches |
| `src/services/metrics-service/main.py` | Metrics & analytics service |
| `src/cli/command.go` | CLI command dispatcher |
| `src/Makefile` | Root build orchestrator |
| `deploy/docker-compose.yml` | All-services Docker Compose (28 services) |
| `deploy/ci/gitlab-ci.yml` | GitLab CI pipeline definition |

## Documentation

| Document | Path | Description |
|----------|------|-------------|
| README | `README.md` | Project readme (not yet created) |
| License | `LICENSE` | Project license |
| Project Specs | `specs/context.md` | Project context and overview |
| Tech Stack | `specs/stack.md` | Technology stack decisions and rationale |
| Glossary | `specs/glossary.md` | Domain-specific terminology |
| Architecture | `.ai-factory/ARCHITECTURE.md` | Architecture documentation (to be generated) |

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
