# VEDO Core

> Virtual Environment for Developing Ontologies — a web platform for multi-user ontology (knowledge graph) editing with Git-like versioning, supporting up to 1 million axioms.

## Overview

VEDO Core replaces the desktop Protégé with a cloud-based ontology editor that enables parallel work through branching, real-time collaboration via WebSocket, and REST/GraphQL API for integration. The system supports ontology engineers, IT architects, developers, data analysts, and DevOps engineers with role-based access control.

## Core Features

- **Ontology Editor** — visual editor for knowledge graphs with <1s response time (p95) for large ontologies
- **Git-like Versioning** — commits, branches, merges, diffs via a dedicated Versioning Service
- **Real-time Collaboration** — WebSocket-powered node locking, comments, notifications
- **REST API & GraphQL** — external integration and frontend data operations
- **Import/Export** — Turtle, RDF/XML, OWL format support
- **Multi-team Organization** — GitLab-like groups, ontologies, membership inheritance, visibility levels
- **Ontology Merge Requests** — review workflow with protected branches and approval rules

## Tech Stack

- **Frontend:** Vue 3 + Composition API, TypeScript, Vite, Apollo Client (GraphQL), Vue Flow (graph visualization)
- **Backend:** Go (API Gateway, Auth, Collaboration, Ticket services), Rust (Ontology, Versioning, Publisher services), Python (Metrics, Classifier services)
- **LLM Integration:** Go shared package (provider abstraction, registry, prompt templates, OTEL observability), OpenAI-compatible and Anthropic adapters
- **Databases:** Neo4j (graph/ontology store), PostgreSQL + JSONB (version store), Redis (cache/locks), RabbitMQ (message queue)
- **Authentication:** Keycloak (OIDC/OAuth2), support for Russian OAuth providers (ESIA/Gosuslugi, Yandex, VK) and Google
- **Observability:** OpenTelemetry, Grafana (Tempo, Prometheus, Loki)
- **Infrastructure:** Docker Compose, Helm charts, GitLab CI
- **Documentation:** Antora (AsciiDoc documentation sites)

## Architecture Notes

- **Microservices architecture** with 14+ services organized by language and domain
- **gRPC** for inter-service communication, REST for external API, WebSocket for real-time
- **Structured logging** with OpenTelemetry trace context propagation across all services
- **Health check endpoints** (`/health`, `/ready`) on every service
- **Multi-stage Docker builds** for all services (Go: golang→alpine, Rust: rust→debian-slim, Python: python-slim, TS: node→nginx)
- **Feature-based CLI tool** (`vedo-cli`) for operator operations: ticket management, secret rotation, guardrail controls
- **BOLA/BFLA security gates** enforced at API Gateway level with negative authorization tests

## Architecture

See `.ai-factory/ARCHITECTURE.md` for detailed architecture guidelines.
**Pattern:** Microservices (Monorepo)

## Non-Functional Requirements

- **Scalability:** Up to 1 million axioms per ontology, 10 concurrent users
- **Performance:** p95 latency <120ms, p99 <250ms, export 100K triples in <10s
- **Security:** OWASP Top 10 compliance, SAST/DAST security gates in CI, dependency scanning
- **Observability:** Distributed tracing (OpenTelemetry), structured JSON logging, Prometheus metrics, Loki log aggregation
- **Deployment:** Containerized via Docker, orchestration-ready for Kubernetes via Helm
