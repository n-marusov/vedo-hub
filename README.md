# VEDO Core

> Virtual Environment for Developing Ontologies — a cloud-based platform for multi-user knowledge-graph editing with Git-like versioning, supporting up to 1 million axioms. **Built for everyone, not just knowledge engineers.**

VEDO Core replaces the desktop Protégé with a web-based ontology editor that makes knowledge graph creation accessible to a broad audience — domain experts, analysts, educators, students, and enthusiasts — through visual intuitive tools, AI assistance, and 3D navigation. Unlike standalone desktop tools, VEDO Hub enables **legal reuse of published ontologies** from the community, dramatically accelerating knowledge graph development. For specialists, it provides parallel work through branching, real-time collaboration via WebSocket, and REST/GraphQL APIs for integration with industry-domain systems.

## Prerequisites

| Tool | Version | Required for |
|------|---------|-------------|
| Docker | 24+ | All services (containers) |
| Docker Compose | v2+ | Full-stack orchestration |
| Make | any | Build orchestrator (`src/Makefile`) |
| Go | 1.21+ | Go services, `vedo-cli` |
| Rust | 1.75+ | Rust services (ontology, versioning, publisher) |
| Python | 3.11+ | Python services (metrics, classifier) |
| Node.js | 20+ | Frontend services |
| pnpm | 8+ | Frontend package manager |

> **Docker-only:** If you just want to run the stack, Docker + Make is enough — language toolchains are only needed for local development.

## Quick Start

```bash
git clone --recurse-submodules <repo-url> vedo-hub
cd vedo-hub/src && make docker-up
```

Once all containers are healthy, open the frontend at `http://localhost:3000` and the API Gateway at `http://localhost:8080`.

> **Note:** If you already cloned without submodules, run `git submodule update --init`.

## Key Features

- **Visual Intuitive Editor** — drag-and-drop graph editor with <1s p95 response, accessible to users without a knowledge-engineering background
- **AI Assistant** — RAG-powered onboarding and help system; build queries in natural language (NLP → SPARQL/GraphQL) without writing code
- **3D Ontology Navigator** — immersive 3D viewer for exploring knowledge graphs, discovering hidden connections, and presenting ontologies to non-technical stakeholders
- **Community & Sharing** — publish ontologies, browse public knowledge graphs, discuss with other users, and contribute to shared vocabularies
- **Git-like Versioning** — commits, branches, merges, and diffs via a dedicated Versioning Service
- **Real-time Collaboration** — WebSocket-powered node locking, comments, and notifications
- **REST & GraphQL APIs** — integration with industry-domain systems (ERP, CRM, MDM, BI) and frontend data operations
- **Import / Export** — Turtle, RDF/XML, and OWL format support
- **Multi-team Organization** — GitLab-like groups, membership inheritance, and visibility levels
- **Ontology Merge Requests** — review workflow with protected branches and approval rules

## Example

```bash
# Start the full stack (28 containers: services + databases + Keycloak)
cd src && make docker-up

# Check service health
make docker-status
# Full detail: make docker-status-full

# Create an ontology via the API Gateway
curl -X POST http://localhost:8080/api/v1/ontologies \
  -H "Content-Type: application/json" \
  -d '{"name": "My Ontology", "description": "Example knowledge graph"}'

# Start the frontend dev server with hot module replacement
make dev-frontend
```

## Tech Stack

| Layer | Technologies |
|-------|-------------|
| Frontend | Vue 3 + TypeScript, Vite, Apollo Client, Vue Flow |
| Backend (Go) | API Gateway, Auth, Commenting, Ticket services |
| Backend (Rust) | Ontology, Versioning, Publisher, Public Browse API |
| Backend (Python) | Metrics, Ticket Classifier |
| Databases | Neo4j, PostgreSQL + JSONB, Redis, RabbitMQ |
| Auth | Keycloak (OIDC/OAuth2) — ESIA, Yandex, VK, Google |
| Observability | OpenTelemetry, Grafana (Tempo, Prometheus, Loki) |
| Infra | Docker Compose, Helm, GitLab CI |

---

## Documentation

Documentation is built with [Antora](https://antora.org). Source files are in `src/docs/antora/`.

### Build the docs site

```bash
cd src/docs/antora && antora antora-playbook.yml
# Output: src/docs/antora/build/site/index.html
```

### Guides by audience

| Guide | Audience | Key pages |
|-------|----------|-----------|
| [User Guide](src/docs/antora/user-guide/modules/ROOT/pages/index.adoc) | Ontology editors, analysts | Quick Start, Ontology Editing, Versioning |
| [Developer Guide](src/docs/antora/developer-guide/modules/ROOT/pages/index.adoc) | Contributors, developers | Getting Started, Architecture, Configuration, Testing |
| [Admin Guide](src/docs/antora/admin-guide/modules/ROOT/pages/index.adoc) | Operators, DevOps | Deployment, Observability, Security |
| [Integrator Guide](src/docs/antora/integrator-guide/modules/ROOT/pages/index.adoc) | External teams | API Reference, Authentication, Integration Config |

## CLI

The `vedo-cli` operator tool supports ticket management, secret rotation, and guardrail controls. See `src/cli/` for command reference.

## License

MIT — see [LICENSE](LICENSE).