# Architecture: Microservices (Monorepo)

## Overview

VEDO Core uses a **Microservices architecture** organized as a monorepo. The system is decomposed into 14+ loosely coupled, independently deployable services, each responsible for a distinct business capability aligned with domain boundaries. Each service owns its data, communicates through well-defined APIs (REST/gRPC/events), and can be developed, deployed, and scaled independently.

This architecture was chosen because the project requires polyglot persistence (Neo4j, PostgreSQL, Redis), different scaling profiles per service (ontology queries need more CPU than auth), and independent team ownership across Go, Rust, Python, and TypeScript codebases.

## Decision Rationale

- **Project type:** Multi-user ontology editing platform with Git-like versioning
- **Tech stack:** Go (API Gateway, Auth, Tickets), Rust (Ontology, Versioning, Publisher), Python (Metrics, Classifier), TypeScript/Vue 3 (Frontend)
- **Key factor:** Different performance characteristics (Rust for graph operations, Go for I/O-bound APIs, Python for analytics) require different runtimes and scaling profiles

## Folder Structure

```
vedo-core/
├── src/
│   ├── services/                 # Microservices (14+ services, polyglot)
│   │   ├── api-gateway/          # Go — Entry point, auth, routing
│   │   ├── auth-service/         # Go — Keycloak integration, RBAC
│   │   ├── ontology-service/     # Rust — Neo4j CRUD, graph operations
│   │   ├── versioning-service/   # Rust — Git-like commits, branches, diff
│   │   ├── metrics-service/      # Python — Analytics, aggregation
│   │   ├── publisher-service/    # Rust — Ontology publishing pipeline
│   │   ├── public-browse-api/    # Rust — Public read-only API
│   │   ├── commenting-service/   # Go — Comments & collaboration
│   │   ├── ticket-api/           # Go — Ticket management CRUD
│   │   ├── ticket-classifier/    # Python — Auto-classification ML
│   │   ├── ticket-notifier/      # Go — Email/WebSocket notifications
│   │   ├── ticket-sync/          # Go — External system sync
│   │   ├── ticket-telemetry-listener/ # Go — Telemetry ingestion
│   │   ├── support-service/      # Go — Support operations
│   │   ├── frontend/             # Vue 3 — Main web UI (BFF pattern)
│   │   └── publish-browse-ui/    # Vue 3 — Public ontology viewer
│   ├── cli/                      # Go — Operator CLI (vedo-cli)
│   ├── docker/                   # Multi-stage Dockerfile templates
│   ├── templates/                # Service scaffolds by language
│   └── build/                    # Makefile includes per language
├── deploy/                       # Infrastructure-as-Code
│   ├── docker-compose.yml        # All-services orchestration
│   ├── ci/gitlab-ci.yml          # CI/CD pipeline
│   ├── helm/                     # Kubernetes Helm values
│   ├── keycloak/                 # Realm configuration
│   └── observability/            # OTEL, Grafana, Prometheus, Loki, Tempo
├── tests/                        # Cross-service test suites
└── specs/                        # Specifications (submodule)
```

## Dependency Rules

- ✅ Internal service communicates via defined API contracts (REST/gRPC/messaging)
- ✅ Frontend → API Gateway (single entry point, never direct to backend services)
- ✅ Service owns its data store entirely (no cross-service database access)
- ✅ Shared observability contracts (OpenTelemetry trace context, structured JSON logging)
- ❌ Services never access another service's database directly
- ❌ No shared mutable state between services (use Redis/RabbitMQ for coordination)
- ❌ No circular service dependencies

## Service Communication

- **Synchronous (REST/gRPC):** API Gateway proxies to backend services, ticket-api serves CLI/HTTP
- **Asynchronous (RabbitMQ):** Metrics ingestion, ticket classification, notifications
- **WebSocket:** Real-time collaboration (commenting, node locking)
- **GraphQL (Apollo):** Frontend → API Gateway for complex data queries

## Key Principles

1. **Service Autonomy:** Each service owns its data, runtime, and deployment. Services are independently testable and deployable.
2. **API-First Contracts:** Service interfaces (REST paths, gRPC protos, event schemas) defined as contracts before implementation.
3. **Smart Endpoints, Dumb Pipes:** Business logic lives inside services — API Gateway routes but does not transform or orchestrate.
4. **Design for Failure:** Every inter-service call can fail. Circuit breakers, retries with backoff, timeouts, fallback strategies at gateway level.
5. **Polyglot Persistence:** Each service picks the right database — Neo4j for graphs, PostgreSQL for version history, Redis for caching/locks.
6. **Observability by Default:** OpenTelemetry trace context propagated through all services. Structured JSON logging with trace_id, request_id.

## Code Examples

### Go Service — API Gateway pattern
```go
// api-gateway/main.go — Entry point with health check and proxy routing
func main() {
    r := gin.Default()
    r.GET("/health", healthHandler)

    // Proxy to backend services
    api := r.Group("/api/v1")
    api.Any("/ontologies/*path", proxyTo(ontologyServiceURL))
    api.Any("/versioning/*path", proxyTo(versioningServiceURL))

    r.Run(":8080")
}
```

### Rust Service — Domain logic with Neo4j
```rust
// ontology-service/src/main.rs — Graph CRUD operations
#[tokio::main]
async fn main() {
    let neo4j = connect_neo4j(&env::var("NEO4J_URI").unwrap()).await;
    // Domain: Class CRUD, property manipulation, relationship management
    axum::Server::bind(&"0.0.0.0:8080".parse().unwrap())
        .serve(app(neo4j).into_make_service())
        .await
        .unwrap();
}
```

## Anti-Patterns

- ❌ **Distributed Monolith:** Services coupled via synchronous calls requiring coordinated deploys
- ❌ **Shared Database:** Multiple services accessing the same Neo4j/PostgreSQL tables
- ❌ **Nano-services:** Services too small that operational overhead exceeds business value
- ❌ **No Tracing:** Missing OpenTelemetry context propagation — production debugging impossible
- ❌ **Leaky Gateway:** API Gateway containing business logic instead of just routing
