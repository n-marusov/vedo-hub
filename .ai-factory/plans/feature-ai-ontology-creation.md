# Plan: AI-Assisted Ontology Creation & Document-to-Ontology Extraction (M2)

**Branch:** `feature/ai-ontology-creation`
**Created:** 2026-07-16
**Mode:** Full

---

## Settings

| Setting | Value |
|---------|-------|
| Methodology | TDD — E2E tests drive the feature; unit tests are written JIT before each implementation task per RULES.md §Testing |
| Logging | Verbose — DEBUG/INFO with trace_id propagation, structured JSON in all services |
| Docs | Yes — mandatory Antora checkpoint at completion per RULES.md §Documentation |
| Traceability | Yes — `.ai-factory/traceability/traceability.ttl` updated for all new/modified artifacts per RULES.md §Traceability |

---

## Roadmap Linkage

**Milestone:** M2 — AI-Assisted Ontology Creation & Document-to-Ontology Extraction

**Rationale:** Full implementation of the low-barrier ontology authoring layer: NL→OWL generation with iterative refinement, AI-assisted class/property completion, relationship hints, ontology templates for common domains, and document-to-ontology extraction from 8 file formats. This milestone is the first user-facing AI feature and builds directly on the M1 core engine.

---

## Research Context

**Source:** `specs/vision.md` F14.1–F14.8, `specs/use-cases/UC-io.import.*`, `specs/user-stories/US-io.document.*`

**Key decisions from ADRs:**
- **Internal service-to-service protocol MUST be gRPC** (`ADR-DES.API.protocol-stack-strategy`, `ADR-DES.INFRA.monolith-vs-microservices`)
- **Port mapping:** internal gRPC ports 9001–9012 (`ADR-IMPL.STACK.port-mapping-strategy`)
- **API Gateway as single facade** — converts external REST/GraphQL to internal gRPC calls
- **Currently violated:** All services use HTTP/REST internally — this MUST be fixed before M2 implementation (Phase 0)

**Core functional scope:**
- NL→OWL generation with iterative refinement
- AI-assisted class/property completion and relationship hints
- Ontology templates for common domains (Person, Organization, Event, Product, etc.)
- Document-to-ontology extraction from 8 formats: MD, TXT, PDF, DOCX, JSON, XML, CSV, XLSX
- User-editable preview with include/exclude, inline editing, duplicate detection
- Atomic application via gRPC `ApplySequence` with automatic commit
- Batch upload of multiple documents with deduplication
- Custom LLM prompt configuration
- Original file attachment as commit artifact

**Related user stories:** US-io.document.extract-md-txt, US-io.document.extract-pdf-docx, US-io.document.extract-structured, US-io.document.batch-extract, US-io.document.preview-sequence

**Related use cases:** UC-io.import.extract-ontology-from-document, UC-io.import.batch-extract-ontology

**TDD strategy (per RULES.md §Testing):**
- **E2E tests first** — this is a major feature (M2); implementation must begin with E2E test specifications that define the expected behavior for all user-facing flows (Rule 3)
- **Unit tests JIT per task** — each implementation task produces unit tests immediately before adding new functionality (Rule 5)
- **Integration tests for external interfaces** — LLM providers are external interfaces requiring dedicated integration tests (Rule 4)
- **Traceability updates** — every new test suite, service, and API contract must be reflected in `.ai-factory/traceability/traceability.ttl` (Rule 6)

---

## Related Artifacts

### Architecture Decision Records (ADRs)

| ADR | Relevance | Phase |
|-----|-----------|-------|
| `specs/adr/ADR-DES.API.protocol-stack-strategy.md` | Mandates gRPC for internal calls, REST for external — currently violated | 0 |
| `specs/adr/ADR-DES.INFRA.monolith-vs-microservices.md` | Microservices decomposition with gRPC/protobuf contracts | 0 |
| `specs/adr/ADR-IMPL.STACK.port-mapping-strategy.md` | Port mapping 9001–9012 for internal gRPC endpoints | 0, 6 |
| `specs/adr/ADR-DES.API.llm-policy-router-strategy.md` | **LLM Policy Router** in API Gateway — central routing by ontology visibility (Public/Internal/Private) and deployment type (SaaS/on-premise). SaaS blocks external LLM for Internal/Private by default. | 2, 3 |
| `specs/adr/ADR-DES.INFRA.doc-extractor-service-strategy.md` | **Document Extractor** design — Python service, two-phase protocol (preview→confirm), Sequence JSON schema, gRPC ApplySequence, 8 format parsers | 3 |
| `specs/adr/ADR-DES.SECURITY.prompt-injection-defense.md` | Two-level prompt injection defense: pre-filtering (VEDO-side), system prompt hardening, post-processing. Sandbox mode, audit. | 2, 3, 4 |
| `specs/adr/ADR-DES.SECURITY.nl-query-opt-in-mandate.md` | NL query mode requires explicit user opt-in; MCP server exempt | 4 |
| `specs/adr/ADR-DES.UI.ai-suggestion-ux-strategy.md` | AI suggestion UX — ranked list with confidence scores, caching (60s TTL), rate limiting, configurable triggers | 4, 5 |
| `specs/adr/ADR-IMPL.STACK.microservice-language-stack-strategy.md` | Language strategy: Rust for CPU-bound (ontology, versioning), Python for ML/NLP, Go for I/O-bound | 2, 3 |
| `specs/adr/ADR-IMPL.STACK.ontology-rust-strategy.md` | Ontology service on Rust (axum → tonic for gRPC) | 0, 3 |
| `specs/adr/ADR-IMPL.STACK.metrics-python-strategy.md` | Python service pattern (FastAPI, uvicorn, structured logging) — used as template for document-extractor | 3 |
| `specs/adr/ADR-IMPL.STACK.api-gateway-go-strategy.md` | API Gateway on Go (gin) — hosts LLM Policy Router | 0, 2 |
| `specs/adr/ADR-IMPL.STACK.frontend-vue-strategy.md` | Frontend on Vue 3 + Apollo Client + Vite | 5 |
| `specs/adr/ADR-DES.INTEGRATION.saga-pattern-strategy.md` | Saga pattern for distributed transactions across services | 0 |
| `specs/adr/ADR-DES.SECURITY.gitlab-like-organization-model.md` | Organization model — Owner/Admin roles manage external LLM override settings | 2 |
| `specs/adr/ADR-DES.INFRA.otel-observability-strategy.md` | OpenTelemetry observability — spans for all LLM calls and gRPC calls | All |
| `specs/adr/ADR-DES.INFRA.airgap-offline-deployment-strategy.md` | Air-gapped deployment — local LLM required, no external providers | 2 |
| `specs/adr/ADR-IMPL.PROCESS.c4-notation-adoption.md` | C4 model notation for architecture diagrams | Reference |

### C4 Architecture Diagrams

| Diagram | Relevance | Phase |
|---------|-----------|-------|
| `specs/c4/context.md` | System Context — VEDO Hub and its external actors (users, LLM providers, integrators) | All |
| `specs/c4/container.md` | Container diagram — all 14+ microservices, data stores, and their interconnections | All |
| `specs/c4/document-extractor-components.md` | **Document Extractor** internal components — Extractors, LLM Client, Sequence Validator, gRPC Client | 3 |
| `specs/c4/ontology-service-components.md` | Ontology Service components — gRPC handlers, Neo4j operations, ApplySequence | 0, 3 |
| `specs/c4/api-gateway-components.md` | API Gateway components — routing, auth, **LLM Policy Router** | 0, 2 |
| `specs/c4/frontend-components.md` | Frontend components — pages, organisms, molecules for AI creation UI | 5 |
| `specs/c4/deployment.md` | Deployment diagram — Docker Compose, CI/CD, observability stack | 0, 6 |

---

## Tasks

### Progress Tracking

> **Legend:** `[ ]` — pending, `[x]` — done, `[~]` — in progress, `[!]` — blocked

| # | Phase | Task | Status | Depends On | Commit |
|---|-------|------|--------|------------|--------|
| 0.1 | 0 | Create shared protobuf definitions | `[x]` | — | 1 |
| 0.2 | 0 | Migrate Rust services to gRPC | `[x]` | 0.1 | 2 |
| 0.3 | 0 | Migrate Go services to gRPC | `[x]` | 0.1 | 2 |
| 0.4 | 0 | Update API Gateway to gRPC proxy (PARTIAL — HTTP proxy still used) | `[x]` | 0.2, 0.3, 0.6 | 3 |
| 0.5 | 0 | Update Docker Compose, CI, configuration | `[x]` | 0.2, 0.3, 0.4 | 3 |
| 0.6 | 0 | Generate proto stubs and create Go gRPC service clients | `[x]` | 0.5 | 3 |
| 1.1 | 1 | E2E tests — document extraction flows | `[x]` | 0.5 | 4 |
| 1.2 | 1 | E2E tests — NL→OWL and AI flows | `[x]` | 0.5 | 4 |
| 2.1 | 2 | LLM abstraction package (Go) | `[x]` | — | 5 |
| 2.2 | 2 | LLM provider registry (runtime-configured) | `[x]` | 2.1 | 5 |
| 2.3 | 2 | Prompt template system | `[x]` | 2.1 | 6 |
| 2.4 | 2 | LLM observability + integration tests | `[x]` | 2.2 | 6 |
| 2.5 | 2 | Implement LLM Policy Router in API Gateway | `[x]` | 2.2, 0.4 | 6 |
| 3.1 | 3 | Scaffold document-extractor service | `[x]` | 0.2, 2.2 | 7 |
| 3.2 | 3 | File parsers — text documents (MD, TXT, PDF, DOCX) | `[x]` | 3.1 | 7 |
| 3.3 | 3 | File parsers — structured data (JSON, XML, CSV, XLSX) | `[x]` | 3.1, 3.2 | 8 |
| 3.4 | 3 | LLM integration + gRPC ApplySequence client | `[x]` | 3.2, 3.3 | 8 |
| 3.5 | 3 | Implement ApplySequence domain logic in ontology-service | `[x]` | 3.4, 2.2 | 8 |
| 4.1 | 4 | NL→OWL conversion endpoint | `[x]` | 2.2, 2.3, 0.6 | 9 |
| 4.2 | 4 | AI-assisted class/property completion | `[x]` | 2.2, 2.3, 4.1 | 9 |
| 4.3 | 4 | Iterative refinement workflow | `[x]` | 4.1, 2.3 | 10 |
| 4.4 | 4 | Ontology domain templates | `[x]` | 0.2, 3.5 | 10 |
| 4.5 | 4 | Implement prompt injection defense | `[x]` | 4.1, 2.3 | 9 |
| 5.1 | 5 | File upload UI component | `[ ]` | 3.4, 0.4 | 11 |
| 5.2 | 5 | Sequence preview table component | `[ ]` | 5.1, 3.4 | 11 |
| 5.3 | 5 | Apply workflow with progress | `[ ]` | 5.2, 3.4 | 12 |
| 5.4 | 5 | Batch upload + conflict resolution | `[ ]` | 5.1, 5.2, 3.4 | 12 |
| 6.1 | 6 | Docker Compose, CI, quality gate | `[ ]` | 3.1–3.5, 0.5 | 13 |
| 6.2 | 6 | Integration test validation (LLM) | `[ ]` | All impl. phases | 14 |
| 6.3 | 6 | Traceability + docs validation | `[ ]` | 6.2 | 15 |
| 6.4 | 6 | Antora documentation | `[ ]` | 6.2 | 15 |

**Progress:** 23 / 31 tasks complete

---

### Phase 0: gRPC Protocol Stack Migration (PREREQUISITE) ⚠️

> ⚠ **BLOCKER:** Current internal service communication uses HTTP/REST, violating [`ADR-DES.API.protocol-stack-strategy`](specs/adr/ADR-DES.API.protocol-stack-strategy.md), [`ADR-DES.INFRA.monolith-vs-microservices`](specs/adr/ADR-DES.INFRA.monolith-vs-microservices.md), and [`ADR-IMPL.STACK.port-mapping-strategy`](specs/adr/ADR-IMPL.STACK.port-mapping-strategy.md). This phase MUST be completed before any M2 functionality implementation.

**Covered ADRs:**
- [`specs/adr/ADR-DES.API.protocol-stack-strategy.md`](specs/adr/ADR-DES.API.protocol-stack-strategy.md) — mandates gRPC for internal service-to-service communication
- [`specs/adr/ADR-DES.INFRA.monolith-vs-microservices.md`](specs/adr/ADR-DES.INFRA.monolith-vs-microservices.md) — microservices decomposition with gRPC/protobuf contracts
- [`specs/adr/ADR-IMPL.STACK.port-mapping-strategy.md`](specs/adr/ADR-IMPL.STACK.port-mapping-strategy.md) — defines gRPC port range 9001–9012
- [`specs/adr/ADR-IMPL.STACK.ontology-rust-strategy.md`](specs/adr/ADR-IMPL.STACK.ontology-rust-strategy.md) — Rust for ontology-service, gRPC migration (axum → tonic)
- [`specs/adr/ADR-IMPL.STACK.api-gateway-go-strategy.md`](specs/adr/ADR-IMPL.STACK.api-gateway-go-strategy.md) — Go for API Gateway, hosts gRPC client pool
- [`specs/adr/ADR-DES.INTEGRATION.saga-pattern-strategy.md`](specs/adr/ADR-DES.INTEGRATION.saga-pattern-strategy.md) — saga pattern for distributed transactions across services

---

#### Task 0.1: Create shared protobuf definitions for all service contracts

**Deliverable:** Proto files defining gRPC contracts for all internal services.

**Files to create:**
- `src/services/shared/proto/ontology/v1/ontology.proto` — Class, Property, Individual CRUD; SPARQL; GraphQL; ApplySequence
- `src/services/shared/proto/versioning/v1/versioning.proto` — Commit, Branch, Diff, Merge, Rollback, Checkout
- `src/services/shared/proto/auth/v1/auth.proto` — TokenIntrospect, CheckPermission, GetUserRoles
- `src/services/shared/proto/commenting/v1/commenting.proto` — CreateComment, ListComments, Mention
- `src/services/shared/proto/publisher/v1/publisher.proto` — PublishSnapshot, GetPublished
- `src/services/shared/proto/public_browse/v1/public_browse.proto` — ReadOnlyBrowse, Search
- `src/services/shared/proto/common/v1/common.proto` — Shared types (Pagination, Error, Timestamp)

**ApplySequence message design (ontology/v1/ontology.proto):**
```protobuf
message ApplySequenceRequest {
  string ontology_id = 1;
  string branch_id = 2;
  repeated SequenceStep steps = 3;
  string commit_message = 4;
  bytes source_file = 5;       // optional — original file as artifact
  string source_filename = 6;  // optional — filename for the artifact
}

message SequenceStep {
  enum Operation {
    OPERATION_UNSPECIFIED = 0;
    OPERATION_CREATE_CLASS = 1;
    OPERATION_CREATE_OBJECT_PROPERTY = 2;
    OPERATION_CREATE_DATATYPE_PROPERTY = 3;
    OPERATION_CREATE_INDIVIDUAL = 4;
    OPERATION_ADD_ANNOTATION = 5;
    OPERATION_SET_PARENT = 6;
    OPERATION_SET_DOMAIN = 7;
    OPERATION_SET_RANGE = 8;
  }
  Operation operation = 1;
  string entity_id = 2;        // IRI or generated ID
  string label = 3;            // rdfs:label
  string parent_id = 4;        // for classes — parent class ID
  string domain_id = 5;        // for properties
  string range_id = 6;         // for properties
  repeated string annotations = 7; // key=value pairs
  string source_file = 8;      // which source file provided this step (batch)
  bool skip_if_exists = 9;     // skip if entity already exists in ontology
}
```

**Logging:** INFO on proto generation, DEBUG on validation.

**Dependencies:** None (first task).

---

#### Task 0.2: Migrate Rust services to gRPC (ontology, versioning, publisher, public-browse)

**Deliverable:** All Rust services expose gRPC servers on ADR-mandated ports while retaining HTTP `/health` and `/ready` endpoints.

**Files to change:**
- `src/services/Cargo.toml` — add workspace deps: `tonic`, `prost`
- `src/services/ontology-service/Cargo.toml` — add `tonic`, `prost`; build.rs for proto compilation
- `src/services/ontology-service/build.rs` — `tonic_build` configuration
- `src/services/ontology-service/src/main.rs` — spawn gRPC server on port 9001, keep HTTP health on current port
- `src/services/ontology-service/src/grpc/ontology_server.rs` — implement OntologyService gRPC trait
- `src/services/ontology-service/src/grpc/mod.rs`
- `src/services/versioning-service/Cargo.toml` — add `tonic`, `prost`
- `src/services/versioning-service/build.rs`
- `src/services/versioning-service/src/main.rs` — gRPC on port 9002
- `src/services/publisher-service/Cargo.toml` — add tonic, prost
- `src/services/publisher-service/src/main.rs` — gRPC on port 9005
- `src/services/public-browse-api/Cargo.toml` — add tonic, prost
- `src/services/public-browse-api/src/main.rs` — gRPC on port 9011

**Key implementation details:**
- Each service MUST start both gRPC (tonic) and HTTP (axum) servers
- HTTP server continues to serve `/health`, `/ready`, `/metrics` only
- gRPC server handles all functional RPCs
- Use `tokio::spawn` to run both servers concurrently
- Health check endpoint MUST NOT be removed — Docker Compose depends on it
- ApplySequence implementation must be atomic: all steps in one Neo4j transaction, rollback on any failure

**Logging:** INFO on server start (both gRPC and HTTP), DEBUG on each RPC call with trace_id, ERROR on failures.

**Dependencies:** Task 0.1 (proto definitions).

---

#### Task 0.3: Migrate Go services to gRPC (auth, commenting, ticket-api)

**Deliverable:** Go services expose gRPC servers; API-relevant Go services use gRPC stubs.

**Files to change:**
- `src/services/auth-service/go.mod` — add `google.golang.org/grpc`, `google.golang.org/protobuf`
- `src/services/auth-service/main.go` — gRPC server on port 9003, keep HTTP health
- `src/services/auth-service/internal/grpc/auth_server.go` — implement AuthService
- `src/services/commenting-service/main.go` — gRPC server on port 9004
- `src/services/ticket-api/main.go` — gRPC server on port 9010 (keep HTTP for CLI compatibility)
- Proto generation: use `protoc` with Go plugins in a `Makefile` target or `buf generate`

**Key implementation details:**
- Go gRPC servers coexist with HTTP health endpoints (separate listeners via goroutines)
- `/health` and `/ready` remain HTTP on the current port for Docker health checks
- Auth service must implement TokenIntrospect and CheckPermission RPCs
- Use `google.golang.org/grpc/health/grpc_health_v1` for gRPC health checking

**Logging:** INFO on server start, DEBUG with trace_id on each RPC, WARN on auth failures.

**Dependencies:** Task 0.1 (proto definitions).

---

#### Task 0.4: Update API Gateway to gRPC proxy (PARTIAL — HTTP proxy still used)

**Deliverable:** API Gateway converts external REST/GraphQL/SPARQL requests to internal gRPC calls. Currently only the gRPC connection pool and address constants exist in `proxy/grpc_client.go` — the proto-specific service clients and route wiring are NOT yet implemented.

**⚠ STATUS:** Infrastructure (gRPC pool, health checking) is in place. Proto-specific gRPC clients need to be created (Task 0.6) and route handlers need migrating from HTTP reverse proxy to gRPC calls.

**Files to change:**
- `src/services/api-gateway/go.mod` — add `google.golang.org/grpc`, gRPC client stubs
- `src/services/api-gateway/routes.go` — replace `mustNewProxy()` with gRPC client calls (depends on Task 0.6)
- `src/services/api-gateway/handlers/ontology_handler.go` — use gRPC client instead of HTTP proxy
- `src/services/api-gateway/handlers/query_handler.go` — use gRPC client for SPARQL/CYPHER
- `src/services/api-gateway/proxy/grpc_client.go` — gRPC connection pool with health checking (DONE)
- `src/services/api-gateway/proxy/grpc_ontology_client.go` — **NEW:** OntologyService gRPC client
- `src/services/api-gateway/proxy/grpc_versioning_client.go` — **NEW:** VersioningService gRPC client
- Remove or deprecate `proxy/proxy.go` HTTP reverse proxy

**Key implementation details:**
- Gateway creates gRPC connections to: ontology-service:9001, versioning-service:9002, auth-service:9003, commenting-service:9004, publisher-service:9005, public-browse-api:9011
- gRPC connections use connection pooling with `grpc.DialContext`
- Health checking via `grpc_health_v1.Health/Check`
- Read endpoints (GET) map to gRPC unary calls
- Write endpoints (POST/PUT/DELETE) map to gRPC unary calls
- SPARQL/CYPHER/GraphQL endpoints map to corresponding gRPC calls
- Circuit breaker: if a gRPC backend is unhealthy, return 503
- Request timeout from `UPSTREAM_TIMEOUT` env var applied to gRPC context deadline

**Logging:** INFO on gRPC connection establishment, DEBUG with full request/response metadata, WARN on connection failures, ERROR on unrecoverable gRPC errors.

**Dependencies:** Tasks 0.2, 0.3 (gRPC servers must be running).

---

#### Task 0.5: Update Docker Compose, CI, and configuration for gRPC

**Deliverable:** Docker Compose uses ADR-mandated internal gRPC ports; CI validates proto generation; all configurations updated.

**Files to change:**
- `deploy/docker-compose.yml` — change internal service ports from HTTP (8081-8091) to gRPC (9001-9012):
  - ontology-service: `expose: 9001` (remove `ports: 8082:8082`)
  - versioning-service: `expose: 9002`
  - auth-service: `expose: 9003`
  - commenting-service: `expose: 9004`
  - publisher-service: `expose: 9005`
  - public-browse-api: `expose: 9011`
  - ticket-api: `expose: 9010`
  - metrics-service: `expose: 9012`
  - Keep HTTP health endpoints published internally only (`expose <http_port>`) — NOT on host
  - api-gateway: update env vars from `*_SERVICE_URL=http://...` to `*_SERVICE_URL=grpc://...` with gRPC connection strings
  - Add health checks for gRPC ports using `grpc_health_probe`
- `deploy/docker-compose.yml` — add `grpc_health_probe` to relevant Dockerfiles or use bash-based gRPC health check
- `deploy/ci/gitlab-ci.yml` — add job: `proto-lint` (buf lint), `proto-breaking` (buf breaking), `proto-generate` (verify generated code matches source)
- `src/Makefile` — add targets: `proto-lint`, `proto-generate`, `proto-breaking`
- `src/build/rust.mk` — update to include proto compilation step
- `src/build/go.mk` — update to include proto generation step
- `deploy/README.md` — update port table to reflect new gRPC ports
- Environment variable documentation update

**Key implementation details:**
- Install `grpc_health_probe` binary in Go/Rust/Python Dockerfiles for health checking
- Port mapping follows `ADR-IMPL.STACK.port-mapping-strategy` exactly
- Metrics service stays on HTTP (9012) — no gRPC needed for Prometheus scrape
- Ticket API stays dual-stack (HTTP for CLI, gRPC for internal) on port 9010
- `buf` tool for proto linting and breaking change detection

**Logging:** Not applicable (infrastructure change).

**Dependencies:** Tasks 0.2, 0.3, 0.4 (services must be gRPC-capable before config update).

**Traceability:** Add proto artifacts, gRPC service endpoints, Docker Compose config changes to `.ai-factory/traceability/traceability.ttl`.

---

#### Task 0.6: Generate proto stubs and create Go gRPC service clients

**Deliverable:** Proto stubs generated for all languages (Go, Python, Rust) and Go gRPC service clients created in the API Gateway wrapping the connection pool.

**Files to change:**
- Run `make proto-generate` — generates Go stubs in `src/services/shared/proto/`
- Run `buf generate` — generates Rust stubs in `src/services/shared/src/protos/`
- Run `grpc_tools.protoc` for Python stubs — output to `src/services/document-extractor/grpc_client/proto/`

**Files to create:**
- `src/services/api-gateway/proxy/grpc_ontology_client.go` — `OntologyServiceClient` wrapper with methods: `CreateClass`, `GetClass`, `ListClasses`, `ApplySequence`
- `src/services/api-gateway/proxy/grpc_versioning_client.go` — `VersioningServiceClient` wrapper with methods: `CreateCommit`, `ListCommits`, `CreateBranch`, `DiffCommits`
- `src/services/api-gateway/proxy/grpc_auth_client.go` — `AuthServiceClient` wrapper: `TokenIntrospect`, `CheckPermission`

**Key implementation details:**
- Each client wraps `GrpcClientPool.GetConn()` and calls the generated proto stub
- Context deadlines come from `pool.GetUpstreamTimeout()`
- All clients log: DEBUG on every call, WARN on connection issues, ERROR on RPC failures
- Clients are lazy-initialized on first use (not at startup for faster boot)
- Python stubs output to `grpc_client/proto/` matching the import path in `grpc_client/ontology.py`
- Add `make proto-generate-python` target to `src/Makefile`

**Logging:** INFO on proto generation, DEBUG on client instantiation and RPC calls.

**Dependencies:** Task 0.5 (build tooling in place), proto definitions from Task 0.1.

**Traceability:** Add proto-generated stubs and gRPC service clients to `.ai-factory/traceability/traceability.ttl`.

---

### Phase 1: E2E Test Specifications (TDD — FIRST)

> 📐 **TDD Rule 3:** A plan for a major feature must begin by implementing the E2E tests for that feature. These tests define the expected behavior for all M2 user-facing flows and serve as the acceptance gate. They will initially fail until implementation makes them pass.

**Covered ADRs:**
- [`specs/adr/ADR-IMPL.PROCESS.c4-notation-adoption.md`](specs/adr/ADR-IMPL.PROCESS.c4-notation-adoption.md) — C4 model for architecture diagrams (test scenario coverage)
- [`specs/adr/ADR-IMPL.SECURITY.bola-bfla-negative-tests-mandate.md`](specs/adr/ADR-IMPL.SECURITY.bola-bfla-negative-tests-mandate.md) — negative authorization tests for E2E security coverage

---

#### Task 1.1: Write E2E tests for document extraction flows

**Deliverable:** Playwright E2E tests covering all document extraction user stories.

**Covered user stories:**
- [`specs/user-stories/US-io.document.extract-md-txt.md`](specs/user-stories/US-io.document.extract-md-txt.md) — извлечение из MD/TXT
- [`specs/user-stories/US-io.document.extract-pdf-docx.md`](specs/user-stories/US-io.document.extract-pdf-docx.md) — извлечение из PDF/DOCX
- [`specs/user-stories/US-io.document.extract-structured.md`](specs/user-stories/US-io.document.extract-structured.md) — извлечение из структурированных форматов (JSON, XML, CSV, XLSX)
- [`specs/user-stories/US-io.document.batch-extract.md`](specs/user-stories/US-io.document.batch-extract.md) — пакетное извлечение
- [`specs/user-stories/US-io.document.preview-sequence.md`](specs/user-stories/US-io.document.preview-sequence.md) — предпросмотр и редактирование sequence

**Files to create:**
- `tests/e2e/playwright/tests/m2/document-extraction-md-txt.spec.ts` — US-io.document.extract-md-txt: upload `specification.md` → verify preview with 8 classes, 5 properties → edit label → apply → verify commit
- `tests/e2e/playwright/tests/m2/document-extraction-pdf-docx.spec.ts` — US-io.document.extract-pdf-docx: upload PDF/DOCX → verify structured extraction → password-protected PDF rejection → scanned PDF detection → large document progress
- `tests/e2e/playwright/tests/m2/document-extraction-structured.spec.ts` — US-io.document.extract-structured: upload XLSX with Class/Parent/Property columns → verify column mapping; JSON with nested objects → verify hierarchy; XML with tags/attributes → verify class mapping; CSV with delimiter detection → verify parsing
- `tests/e2e/playwright/tests/m2/document-batch-extract.spec.ts` — US-io.document.batch-extract: upload 3 files of different formats → verify combined preview with source attribution → deduplication → conflict resolution → partial failure handling
- `tests/e2e/playwright/tests/m2/document-preview-sequence.spec.ts` — US-io.document.preview-sequence: toggle step inclusion → inline label editing → duplicate detection warning → apply button → progress indicator → success/commit link → error/retry
- `tests/e2e/playwright/fixtures/m2/` — test fixture documents:
  - `specification.md` — markdown with structured headings, lists
  - `requirements.txt` — plain text with structured content
  - `technical-spec.pdf` — PDF with text layer
  - `documentation.docx` — DOCX with headings and tables
  - `data.json` — nested JSON with object references
  - `schema.xml` — XML with namespace attributes
  - `entities.csv` — CSV with Class,Property columns
  - `classes.xlsx` — XLSX with Class/Parent/Property/Type columns
  - `empty.txt` — empty file (error case)
  - `large-file.txt` — 25MB file (size limit case)

**Test implementation notes:**
- Mock LLM responses at API Gateway level (network intercept) for deterministic tests — use Playwright `route.fulfill()`
- Each test follows the Arrange-Act-Assert pattern
- Tests must be executable immediately after Phase 0 (gRPC migration) completes — they will FAIL initially
- As each implementation phase completes, progressively more tests should pass
- All tests follow the Gherkin scenarios from the attached user stories exactly

**Logging:** Not applicable (tests).

**Dependencies:** Task 0.5 (Docker Compose with gRPC ports must be running).

**Traceability:** Add E2E test suite and all fixture files to `.ai-factory/traceability/traceability.ttl` with `vdo:validates` relationships to the corresponding user stories.

---

#### Task 1.2: Write E2E tests for NL→OWL generation and AI assistance flows

**Deliverable:** Playwright E2E tests covering NL-to-OWL generation, AI-assisted completion, and template application.

**Covered specs & requirements:**
- [`specs/vision.md`](specs/vision.md) §F14.1–F14.8 — NL→OWL generation, iterative refinement, ontology templates, AI-assisted completion (feature-level requirements)

**Files to create:**
- `tests/e2e/playwright/tests/m2/nl-to-owl-generation.spec.ts` — enter NL text → view generated sequence preview → edit steps → apply → verify entities in ontology
- `tests/e2e/playwright/tests/m2/ai-completion.spec.ts` — open class → request AI subclass suggestions → view ranked list → accept/reject suggestions → verify created entities
- `tests/e2e/playwright/tests/m2/ai-property-suggestions.spec.ts` — open class → request property suggestions → verify domain/range hints → accept suggestion
- `tests/e2e/playwright/tests/m2/iterative-refinement.spec.ts` — generate initial ontology → provide feedback → verify refined sequence → apply final version
- `tests/e2e/playwright/tests/m2/ontology-templates.spec.ts` — list available templates → select "Person/Organization" template → preview → apply → verify created entities

**Test implementation notes:**
- Mock LLM responses with deterministic JSON sequences
- Verify that refinement actually changes the sequence (diff between rounds)
- Template test verifies all expected classes and properties are created

**Logging:** Not applicable (tests).

**Dependencies:** Task 0.5 (Docker Compose with gRPC ports).

**Traceability:** Add to `.ai-factory/traceability/traceability.ttl`.

---

### Phase 2: LLM Provider Integration Layer

**Covered ADRs:**
- [`specs/adr/ADR-DES.API.llm-policy-router-strategy.md`](specs/adr/ADR-DES.API.llm-policy-router-strategy.md) — LLM Policy Router for access control by ontology visibility
- [`specs/adr/ADR-DES.SECURITY.prompt-injection-defense.md`](specs/adr/ADR-DES.SECURITY.prompt-injection-defense.md) — two-level prompt injection defense (pre-filtering + hardening)
- [`specs/adr/ADR-IMPL.STACK.microservice-language-stack-strategy.md`](specs/adr/ADR-IMPL.STACK.microservice-language-stack-strategy.md) — Go for I/O-bound services (API Gateway hosts LLM logic)
- [`specs/adr/ADR-IMPL.STACK.api-gateway-go-strategy.md`](specs/adr/ADR-IMPL.STACK.api-gateway-go-strategy.md) — Go for API Gateway, hosts LLM Policy Router
- [`specs/adr/ADR-DES.SECURITY.gitlab-like-organization-model.md`](specs/adr/ADR-DES.SECURITY.gitlab-like-organization-model.md) — organization model for LLM access override settings
- [`specs/adr/ADR-DES.INFRA.airgap-offline-deployment-strategy.md`](specs/adr/ADR-DES.INFRA.airgap-offline-deployment-strategy.md) — local LLM requirement for air-gapped deployments

---

#### Task 2.1: Create shared LLM abstraction package (Go)

**Deliverable:** A reusable Go package providing a unified interface for LLM providers, used by API Gateway and future Go-based LLM consumers.

**Files to create:**
- `src/services/shared/llm/provider.go` — `Provider` interface: `Complete(ctx, Prompt) (Completion, error)`, `StreamComplete(ctx, Prompt) (<-chan Token, error)`
- `src/services/shared/llm/types.go` — `Prompt{Messages[], Model, Temperature, MaxTokens}`, `Completion{Text, TokensUsed, FinishReason}`, `Token{Text, Index}`
- `src/services/shared/llm/config.go` — `LoadConfigFromEnv()`: reads `LLM_PROVIDER`, `LLM_API_KEY`, `LLM_MODEL`, `LLM_BASE_URL`, `LLM_TIMEOUT`, `LLM_MAX_TOKENS`, `LLM_TEMPERATURE`
- `src/services/shared/llm/errors.go` — typed errors: `ErrProviderUnavailable`, `ErrRateLimited`, `ErrContextTooLong`, `ErrInvalidResponse`
- `src/services/shared/llm/retry.go` — exponential backoff with jitter, max 3 retries
- `src/services/shared/llm/metrics.go` — prometheus counters: `llm_requests_total`, `llm_tokens_total`, `llm_errors_total`, `llm_latency_seconds`

**Unit tests (TDD Rule 5):** Write unit tests BEFORE implementation:
- `src/services/shared/llm/provider_test.go` — mock Provider implementation, test interface compliance
- `src/services/shared/llm/types_test.go` — serialization, validation
- `src/services/shared/llm/config_test.go` — env var parsing, defaults, validation
- `src/services/shared/llm/errors_test.go` — error type checking
- `src/services/shared/llm/retry_test.go` — backoff timing, max retries, jitter
- `src/services/shared/llm/metrics_test.go` — counter increments, label correctness

**Logging:** DEBUG for every LLM call (model, prompt length, latency), INFO for config load, WARN on retry, ERROR on permanent failures.

**Dependencies:** None (standalone package).

---

#### Task 2.2: Implement LLM provider registry (runtime-configured backends)

**Deliverable:** A provider registry that instantiates LLM backends at runtime based on environment configuration. No provider names are hardcoded in folder structure or code paths.

**Files to create:**
- `src/services/shared/llm/registry.go` — global registry: `Register(name, factory)`, `NewProvider(config) (Provider, error)`. Reads `LLM_PROVIDER` env var for the default provider name, looks up the registered factory, instantiates.
- `src/services/shared/llm/providers/openai_compatible.go` — adapter for any OpenAI-compatible chat API. Configured via env: `LLM_PROVIDERS_<NAME>_BASE_URL`, `LLM_PROVIDERS_<NAME>_API_KEY`, `LLM_PROVIDERS_<NAME>_MODEL`. Supports custom base URL for proxies and local models.
- `src/services/shared/llm/providers/anthropic.go` — adapter for Anthropic Messages API (`LLM_PROVIDERS_<NAME>_API_KEY`, `LLM_PROVIDERS_<NAME>_MODEL`).
- `src/services/shared/llm/registry_test.go` — unit tests for registry dispatch
- `src/services/shared/llm/providers/openai_compatible_test.go` — mock HTTP server tests
- `src/services/shared/llm/providers/anthropic_test.go` — mock HTTP server tests

**Key implementation details:**
- Providers register themselves via `init()` functions or explicit registration at startup
- `NewProvider()` reads `LLM_PROVIDER` env var (default provider name), looks up the registry, calls the factory with the full config
- Additional provider configurations are loaded from `LLM_PROVIDERS_<NAME>_*` env vars — each named provider gets its own API key, base URL, model, timeout
- OpenAI-compatible adapter handles any API conforming to the OpenAI chat completions format (OpenAI, Ollama, vLLM, local models)
- Anthropic adapter handles the Messages API format
- All providers implement the same `Provider` interface from Task 2.1
- Each provider includes `net/http` client with configurable timeout
- Support system prompt + user prompt combination
- Parse structured JSON responses (for ontology generation) — handle malformed JSON gracefully

**Configuration example (`.env`):**
```bash
# Default provider
LLM_PROVIDER=ollama

# Provider-specific configs (provider names are config keys, not hardcoded)
LLM_PROVIDERS_OLLAMA_BASE_URL=http://localhost:11434/v1
LLM_PROVIDERS_OLLAMA_MODEL=llama3

LLM_PROVIDERS_OPENAI_API_KEY=sk-...
LLM_PROVIDERS_OPENAI_MODEL=gpt-4o

LLM_PROVIDERS_ANTHROPIC_API_KEY=sk-ant-...
LLM_PROVIDERS_ANTHROPIC_MODEL=claude-3-opus-20240229
```

**Unit tests (TDD Rule 5):** Write BEFORE implementation:
- `src/services/shared/llm/registry_test.go` — test registry dispatch, unknown provider error, default provider resolution from env
- `src/services/shared/llm/providers/openai_compatible_test.go` — mock HTTP server, test Complete(), error handling, rate limit
- `src/services/shared/llm/providers/anthropic_test.go` — mock HTTP server, test Complete(), error cases

**Logging:** DEBUG on request/response (mask API keys), INFO on provider initialization (name, model), WARN on 429/5xx, ERROR on connection failures.

**Dependencies:** Task 2.1 (Provider interface).

---

#### Task 2.3: Create prompt template system

**Deliverable:** A template system for LLM prompts used in ontology generation, document extraction, and AI-assisted completion.

**Files to create:**
- `src/services/shared/llm/templates/engine.go` — template engine: load from filesystem, variable substitution (`{{.ClassName}}`), validation
- `src/services/shared/llm/templates/ontology_generation.tmpl` — prompt for NL→OWL generation
- `src/services/shared/llm/templates/document_extraction.tmpl` — prompt for document-to-ontology extraction
- `src/services/shared/llm/templates/class_completion.tmpl` — prompt for AI-assisted class completion
- `src/services/shared/llm/templates/property_suggestion.tmpl` — prompt for relationship hints
- `src/services/shared/llm/templates/refinement.tmpl` — prompt for iterative refinement
- `src/services/shared/llm/templates/engine_test.go` — unit tests

**Key template structure (document_extraction.tmpl example):**
```
You are an ontology engineer. Given the following document content, extract
an OWL ontology structure. Return a JSON array of SequenceStep objects.

Document title: {{.Title}}
Document format: {{.Format}}

Content:
{{.Content}}

Return format:
{
  "steps": [
    {
      "operation": "CREATE_CLASS",
      "entity_id": "...",
      "label": "...",
      "parent_id": "Thing",
      "annotations": {"rdfs:comment": "..."}
    },
    ...
  ]
}

Rules:
- Classes from headings (H1-H2), paragraphs as descriptions
- Properties from lists, tables, or key-value pairs
- Hierarchy from nesting/indentation
- Use English labels, keep original term as rdfs:comment
```

**Unit tests (TDD Rule 5):** Write BEFORE implementation:
- `src/services/shared/llm/templates/engine_test.go` — template loading, variable substitution

**Logging:** DEBUG on template load and render, INFO on template cache miss.

**Dependencies:** Task 2.1 (types).

---

#### Task 2.4: Add LLM observability (metrics, tracing, cost tracking)

**Deliverable:** Complete observability for all LLM interactions.

**Files to create/change:**
- `src/services/shared/llm/observability.go` — OpenTelemetry span creation for each LLM call
- `src/services/shared/llm/cost.go` — cost tracking per provider/model (token pricing table)
- Update `src/services/shared/llm/metrics.go` — add histograms for latency, counters for cost
- `src/services/shared/llm/observability_test.go`

**Key implementation details:**
- Each LLM call creates an OTel span with attributes: `llm.provider`, `llm.model`, `llm.prompt_tokens`, `llm.completion_tokens`, `llm.temperature`
- Token usage counter per provider, per model
- Cost estimation: `prompt_tokens * prompt_price + completion_tokens * completion_price`
- Latency histogram: `llm_request_duration_seconds` with buckets [0.5, 1, 2, 5, 10, 30, 60]
- Error counter by error type: `llm_errors_total{error="rate_limited|timeout|invalid_response|connection"}`

**Logging:** INFO on cost summary per request, DEBUG on span attributes, WARN when cost exceeds threshold.

**Dependencies:** Task 1.2 (providers must be instrumented).

---

#### Task 2.5: Implement LLM Policy Router in API Gateway

**Deliverable:** Policy router in the API Gateway that controls LLM access based on ontology visibility and deployment type, per `ADR-DES.API.llm-policy-router-strategy`.

**Files to create:**
- `src/services/api-gateway/middleware/llm_policy_router.go` — policy evaluation: ontology visibility (Public/Internal/Private) × deployment mode (SaaS/on-premise) × provider type (external/local)
- `src/services/api-gateway/middleware/llm_policy_router_test.go` — unit tests for all policy combinations

**Files to change:**
- `src/services/api-gateway/routes.go` — apply LLM Policy Router middleware to AI-related routes (`/api/v1/ontologies/:id/generate-from-text`, `/api/v1/ontologies/:id/ai/*`, document-extractor proxy routes)

**Policy matrix:**

| Ontology Visibility | SaaS + External LLM | SaaS + Local LLM | On-premise (any LLM) |
|--------------------|-------------------|-----------------|--------------------|
| Public | ✅ Allow | ✅ Allow | ✅ Allow |
| Internal | ❌ Block (default) / Admin override | ✅ Allow | ✅ Allow |
| Private | ❌ Block (default) / Admin override | ✅ Allow | ✅ Allow |

**Key implementation details:**
- Router reads ontology visibility from the ontology-service (gRPC `GetOntology` or cached header)
- Deployment mode determined by `DEPLOYMENT_MODE` env var (`saas`/`on-premise`/`air-gapped`)
- `air-gapped` mode blocks ALL external LLM providers, only allows local
- Admin override via `X-Admin-Override` header + `checkPermission` gRPC call to auth-service
- Blocked requests return 403 with error code `LLM-POLICY-BLOCKED` and reason description
- Cached policy decisions per ontology_id with 60s TTL to reduce auth-service calls

**Unit tests (TDD Rule 5):** Write BEFORE implementation:
- `llm_policy_router_test.go` — test all 12 policy combinations (3 visibility × 2 deployment × 2 provider type), test admin override, test cache expiry

**Logging:** INFO on policy decision (allow/block + reason), DEBUG on ontology visibility lookup, WARN on blocked request (with user_id, ontology_id, provider).

**Dependencies:** Task 2.2 (LLM provider registry), Task 0.4 (API Gateway gRPC for ontology lookup).

---

### Phase 3: Document Extractor Service (Python)

**Covered use cases:**
- [`specs/use-cases/UC-io.import.extract-ontology-from-document.md`](specs/use-cases/UC-io.import.extract-ontology-from-document.md) — извлечение онтологии из загружаемого документа (одиночный файл)
- [`specs/use-cases/UC-io.import.batch-extract-ontology.md`](specs/use-cases/UC-io.import.batch-extract-ontology.md) — пакетное извлечение онтологии из нескольких документов

**Covered ADRs:**
- [`specs/adr/ADR-DES.INFRA.doc-extractor-service-strategy.md`](specs/adr/ADR-DES.INFRA.doc-extractor-service-strategy.md) — Document Extractor service design: two-phase protocol, parsers, gRPC ApplySequence
- [`specs/adr/ADR-DES.API.llm-policy-router-strategy.md`](specs/adr/ADR-DES.API.llm-policy-router-strategy.md) — LLM Policy Router governs extractor's access to LLM providers
- [`specs/adr/ADR-DES.SECURITY.prompt-injection-defense.md`](specs/adr/ADR-DES.SECURITY.prompt-injection-defense.md) — pre-filtering of user documents before LLM submission
- [`specs/adr/ADR-IMPL.STACK.microservice-language-stack-strategy.md`](specs/adr/ADR-IMPL.STACK.microservice-language-stack-strategy.md) — Python for NLP/ML services (document-extractor)
- [`specs/adr/ADR-IMPL.STACK.metrics-python-strategy.md`](specs/adr/ADR-IMPL.STACK.metrics-python-strategy.md) — Python service pattern (FastAPI, uvicorn) used as template
- [`specs/adr/ADR-IMPL.STACK.ontology-rust-strategy.md`](specs/adr/ADR-IMPL.STACK.ontology-rust-strategy.md) — Rust for ontology-service that receives ApplySequence calls

---

#### Task 3.1: Scaffold document-extractor service

**Deliverable:** New Python microservice with FastAPI, health checks, Docker Compose integration, and gRPC client.

**Files to create:**
- `src/services/document-extractor/pyproject.toml` — dependencies: `fastapi`, `uvicorn`, `grpcio`, `grpcio-tools`, `pydantic`, `python-docx`, `PyPDF2`, `pdfplumber`, `openpyxl`, `aiofiles`, `httpx`
- `src/services/document-extractor/main.py` — FastAPI app with lifespan, health, ready, metrics endpoints
- `src/services/document-extractor/Dockerfile` — `python:3.12-slim`, multi-stage, `pip install`, system deps for PDF
- `src/services/document-extractor/parsers/__init__.py`
- `src/services/document-extractor/grpc_client/__init__.py` — gRPC client to ontology-service:9001
- `src/services/document-extractor/grpc_client/ontology.py` — `ApplySequence` RPC wrapper
- `src/services/document-extractor/config.py` — pydantic settings from env: `LLM_PROVIDER`, `LLM_API_KEY`, `LLM_MODEL`, `ONTOLOGY_SERVICE_URL`, `SERVICE_PORT`, `MAX_FILE_SIZE_MB`

**Service port:** 9013 (gRPC internal) + 8092 (HTTP health/metrics)

**Logging:** Structured JSON logging (matching metrics-service pattern), DEBUG for all operations, trace_id propagation from incoming requests.
**Logging:** DEBUG for all operations, trace_id propagation from incoming requests.

**Unit tests (TDD Rule 5):** Write BEFORE implementation — test FastAPI routes, health endpoints, config loading.

**Dependencies:** Task 0.2 (ontology-service gRPC must be running).

---

#### Task 3.2: Implement file parsers — text documents (MD, TXT, PDF, DOCX)

**Deliverable:** Parsers that extract structured content from text-based documents.

**Files to create:**
- `src/services/document-extractor/parsers/base.py` — `BaseParser` abstract class: `parse(file_path) -> ParsedDocument`
- `src/services/document-extractor/parsers/models.py` — `ParsedDocument{title, sections[], metadata}`, `Section{heading, level, paragraphs[], tables[], lists[]}`
- `src/services/document-extractor/parsers/text_parser.py` — MD/TXT parser: headings (H1-H6), bullet lists, numbered lists, paragraphs
- `src/services/document-extractor/parsers/pdf_parser.py` — PDF parser: text extraction with `pdfplumber`, structure detection (font size for headings), password detection, scanned document detection (no text layer)
- `src/services/document-extractor/parsers/docx_parser.py` — DOCX parser: headings (Heading 1-6 styles), paragraphs, tables (→ datatype properties)
- `src/services/document-extractor/parsers/text_parser_test.py`
- `src/services/document-extractor/parsers/pdf_parser_test.py`
- `src/services/document-extractor/parsers/docx_parser_test.py`

**Key implementation details:**
- MD/TXT: `re` for heading detection, markdown structure preservation
- PDF: `pdfplumber` for text extraction (better than PyPDF2 for layout); detect encrypted PDF, detect scanned PDF (no extractable text)
- DOCX: `python-docx` for style-based structure detection
- All parsers return `ParsedDocument` with sections preserving hierarchy
- File size limit: 20 MB (sync), >20 MB → suggest async mode
- File validation: magic bytes check (reject non-matching extensions)

**Unit tests (TDD Rule 5):** Write BEFORE implementation: `text_parser_test.py`, `pdf_parser_test.py`, `docx_parser_test.py` — test each parser with sample fixtures, error cases (empty, password-protected, scanned).

**Logging:** DEBUG on parse start/end (file size, format, section count), INFO on file validation, WARN on password-protected PDF, WARN on scanned PDF, ERROR on parse failures.

**Dependencies:** Task 3.1 (service scaffold).

---

#### Task 3.3: Implement file parsers — structured data (JSON, XML, CSV, XLSX)

**Deliverable:** Parsers that extract structured content from data-oriented file formats.

**Files to create:**
- `src/services/document-extractor/parsers/json_parser.py` — JSON parser: nested keys → class hierarchy, primitives → datatype properties, object refs → object properties
- `src/services/document-extractor/parsers/xml_parser.py` — XML parser: tags → classes, attributes → properties, nesting → hierarchy
- `src/services/document-extractor/parsers/csv_parser.py` — CSV parser: delimiter auto-detection (`,`/`;`/`\t`), header row → column names, LLM column mapping
- `src/services/document-extractor/parsers/xlsx_parser.py` — XLSX parser: `openpyxl`, sheet selection, column auto-mapping (Class/Parent/Property/Type pattern)
- `src/services/document-extractor/parsers/json_parser_test.py`
- `src/services/document-extractor/parsers/xml_parser_test.py`
- `src/services/document-extractor/parsers/csv_parser_test.py`
- `src/services/document-extractor/parsers/xlsx_parser_test.py`

**Key implementation details:**
- JSON: traverse recursively, infer types from JSON values (string→datatype, number→datatype, object→class)
- XML: use `xml.etree.ElementTree`, preserve namespace info
- CSV: try `,` first, fallback to `;` and `\t`; if >50% consistent, use that delimiter
- XLSX: detect "Class/Parent/Property/Type" pattern, otherwise treat as generic table
- All parsers return uniform `ParsedDocument` for downstream LLM processing

**Logging:** DEBUG on parse (format, row/element count, inferred columns), INFO on delimiter detection, WARN on empty sheets/rows.

**Unit tests (TDD Rule 5):** Write BEFORE implementation: `json_parser_test.py`, `xml_parser_test.py`, `csv_parser_test.py`, `xlsx_parser_test.py` — test each parser with sample fixtures, edge cases (empty, malformed, mixed delimiters).

**Dependencies:** Task 3.1 (service scaffold), Task 3.2 (shared models).

---

#### Task 3.4: Implement LLM integration and gRPC ApplySequence client

**Deliverable:** Document extractor sends parsed content to LLM, receives sequence, validates it, and sends to ontology-service via gRPC.

**Files to create:**
- `src/services/document-extractor/llm/client.py` — Python LLM client using `httpx` (runtime-configured provider from env)
- `src/services/document-extractor/llm/prompts.py` — prompt templates for ontology extraction
- `src/services/document-extractor/sequence/model.py` — `SequenceStep` pydantic model, `ApplySequenceRequest` model
- `src/services/document-extractor/sequence/validator.py` — validate LLM output: JSON parse, schema check, entity reference consistency
- `src/services/document-extractor/api/routes.py` — FastAPI routes: `POST /api/v1/documents/extract`, `POST /api/v1/documents/extract/batch`
- `src/services/document-extractor/api/routes_test.py` — route tests with mocked LLM

**Key implementation details:**
- `POST /api/v1/documents/extract`: accept multipart file upload, parse, send to LLM, return sequence preview
- LLM retry: max 3 attempts with exponential backoff, progressively more structured prompts
- Sequence validation: check no circular references, valid operation types, required fields present
- gRPC ApplySequence call: connect to ontology-service:9001, send validated sequence, handle errors
- Response includes: apply status, commit ID, entity count, any warnings

**Unit tests (TDD Rule 5):** Write BEFORE implementation: `client_test.py` (mock LLM HTTP), `validator_test.py` (valid/invalid sequences), `routes_test.py` (mock FastAPI test client with mocked services).

**Logging:** DEBUG on LLM request/response (truncate content >1KB), INFO on sequence validation (step count, warnings), INFO on ApplySequence result (commit ID, entities created), WARN on LLM retries, ERROR on gRPC failures.

**Dependencies:** Tasks 3.2, 3.3 (parsers for all formats), Task 0.2 (ontology-service gRPC).

**Traceability:** Add document-extractor service, all parsers, and routes to `.ai-factory/traceability/traceability.ttl`.

**Dependencies:** Tasks 2.2, 2.3 (parsers for all formats), Task 0.2 (ontology-service gRPC).

---

#### Task 3.5: Implement ApplySequence domain logic in ontology-service

**Deliverable:** The ontology-service gRPC `ApplySequence` RPC (currently `Status::unimplemented`) is fully implemented with atomic Neo4j transaction, source file attachment, and commit creation.

**Files to change:**
- `src/services/ontology-service/src/grpc/mod.rs` — replace `ApplySequence` stub with real implementation
- `src/services/ontology-service/src/services/apply_sequence.rs` — **NEW:** ApplySequence business logic: validate steps, execute Neo4j transaction, create commit
- `src/services/ontology-service/src/repositories/sequence_repo.rs` — **NEW:** Neo4j operations for sequence steps

**Key implementation details:**
- Process steps in order: create classes first, then properties, then individuals, then annotations/relationships
- All steps execute in a **single Neo4j transaction** — if any step fails, the entire transaction rolls back
- For each `create_class` step: create Neo4j node with label, parent relationship (rdfs:subClassOf), annotations as properties
- For `create_object_property`: create relationship type node with domain/range
- For `create_datatype_property`: create property node with domain and xsd range
- Deduplication: if `skip_if_exists=true` and entity_id already exists, skip silently (increment skipped counter)
- After all steps succeed, create a versioning commit via the versioning-service gRPC (or direct DB write for now)
- Attach `source_file` bytes as commit artifact if provided
- Return `ApplySequenceResponse` with commit_id, steps_applied, steps_skipped, steps_failed, errors

**Error handling:**
- Invalid entity references (dangling parent_id) → step-level error, continue remaining steps
- Neo4j constraint violation (duplicate IRI without skip_if_exists) → step-level error
- Transaction conflict → retry up to 3 times with exponential backoff
- Fatal errors (connection lost, auth failure) → abort entire sequence

**Unit tests (TDD Rule 5):** Write BEFORE implementation:
- Unit tests for step ordering (classes before properties), Neo4j transaction behavior (commit/rollback), deduplication, source file attachment, error reporting

**Logging:** DEBUG on each step execution, INFO on transaction commit (step count, entity IDs), WARN on individual step failures (with step index + reason), ERROR on transaction rollback.

**Dependencies:** Tasks 3.4 (gRPC client calls this), 2.2 (LLM provider — the source data comes through LLM).

---

### Phase 4: NL→OWL Generation & AI Assistance

**Covered specs & requirements:**
- [`specs/vision.md`](specs/vision.md) §F14.1–F14.3 — генерация OWL из NL-текста, итеративное уточнение, AI-подсказки для завершения классов/свойств
- [`specs/vision.md`](specs/vision.md) §F14.4–F14.5 — шаблоны онтологий для типовых доменов

**Covered ADRs:**
- [`specs/adr/ADR-DES.SECURITY.prompt-injection-defense.md`](specs/adr/ADR-DES.SECURITY.prompt-injection-defense.md) — post-processing for NL→OWL output, system prompt hardening
- [`specs/adr/ADR-DES.SECURITY.nl-query-opt-in-mandate.md`](specs/adr/ADR-DES.SECURITY.nl-query-opt-in-mandate.md) — explicit user opt-in required for NL query mode
- [`specs/adr/ADR-DES.UI.ai-suggestion-ux-strategy.md`](specs/adr/ADR-DES.UI.ai-suggestion-ux-strategy.md) — ranked list with confidence scores, caching (60s TTL), rate limiting

---

#### Task 4.1: Implement NL→OWL conversion endpoint

**Deliverable:** API endpoint that accepts natural language description and returns OWL ontology structure via gRPC.

**Files to create:**
- `src/services/api-gateway/handlers/nl_to_owl_handler.go` — handler: accept NL text, call LLM, return sequence preview
- `src/services/api-gateway/routes.go` — add route: `POST /api/v1/ontologies/:id/generate-from-text`
- _Note:_ The NL→OWL prompt template already exists as `src/services/shared/llm/templates/ontology_generation.tmpl` (created in Task 2.3). Extend with NL-specific variables if needed.

**Files to change:**
- `src/services/api-gateway/handlers/ontology_handler.go` — add `HandleGenerateFromText` method

**Key implementation details:**
- Accept `{ontology_id, text, model?}` in request body
- Use shared LLM provider (Go) to generate ontology structure
- Return sequence preview (same format as document extraction)
- Support optional `model` parameter for overriding default LLM model
- Propagation of generated sequence to frontend preview (Phase 5)

**Unit tests (TDD Rule 5):** Write BEFORE implementation:
- `src/services/api-gateway/handlers/nl_to_owl_handler_test.go` — test handler with mock LLM provider, test edge cases (empty text, too long text, LLM error)

**Logging:** DEBUG on NL input (>100 chars truncated), INFO on generation result (entity count), WARN on LLM errors.

**Dependencies:** Task 2.2 (LLM providers), Task 2.3 (prompt templates), Task 0.4 (API Gateway gRPC).

---

#### Task 4.2: Implement AI-assisted class/property completion and relationship hints

**Deliverable:** AI endpoints that suggest completions for classes, properties, and relationships.

**Files to create:**
- `src/services/api-gateway/handlers/ai_completion_handler.go` — handlers for completion suggestions
- _Note:_ The completion prompt template already exists as `class_completion.tmpl` (created in Task 2.3). Enrich with ontology context variables. If an extended version is needed, create `class_completion_v2.tmpl` as an enriched variant.

**Files to change:**
- `src/services/api-gateway/routes.go` — add routes:
  - `POST /api/v1/ontologies/:id/ai/suggest-classes` — suggest subclasses for a given class
  - `POST /api/v1/ontologies/:id/ai/suggest-properties` — suggest properties for a class
  - `POST /api/v1/ontologies/:id/ai/suggest-relationships` — suggest relationships between classes

**Key implementation details:**
- Each endpoint receives current ontology context (class tree, existing properties)
- LLM prompt includes ontology prefix, existing class names, property names
- Response: ranked list of suggestions with confidence scores
- Suggestions include: label, suggested parent, description/rationale
- User can accept individual suggestions or dismiss

**Unit tests (TDD Rule 5):** Write BEFORE implementation:
- `src/services/api-gateway/handlers/ai_completion_handler_test.go` — test each suggestion endpoint with mock LLM, test empty context, test large ontology context

**Logging:** DEBUG on request (class name, ontology context size), INFO on suggestion count, WARN on empty suggestions.

**Dependencies:** Tasks 2.2, 2.3, 4.1.

---

#### Task 4.3: Implement iterative refinement workflow

**Deliverable:** User can iteratively refine generated ontology structure through conversation-like feedback.

**Files to create:**
- `src/services/api-gateway/handlers/refinement_handler.go` — refinement endpoint handler
- `src/services/shared/llm/templates/refinement.tmpl` — refinement prompt

**Files to change:**
- `src/services/api-gateway/routes.go` — add route: `POST /api/v1/ontologies/:id/ai/refine`
- `src/services/frontend/src/pages/OntologyWorkspace.vue` (in Phase 5) — refinement UI

**Key implementation details:**
- Accept `{sequence_id, feedback_text, changes[]}` — previous sequence + user feedback
- LLM receives original generation context + user feedback + current sequence
- Returns updated sequence with only the changed steps
- Supports multiple refinement iterations (conversation history in context window)
- Max 5 refinement rounds per sequence (configurable)

**Unit tests (TDD Rule 5):** Write BEFORE implementation:
- `src/services/api-gateway/handlers/refinement_handler_test.go` — test refinement with mock LLM, test max rounds limit, test diff generation

**Logging:** DEBUG on feedback text, INFO on refinement round number, DEBUG on diff between old and new sequence.

**Dependencies:** Tasks 4.1, 2.3 (refinement template).

---

#### Task 4.4: Create ontology domain templates

**Deliverable:** Pre-built ontology templates for common domains, accessible via API and UI.

**Files to create:**
- `src/services/api-gateway/internal/templates/ontologies/person.json` — Person, Organization, Event, Location, Document, etc.
- `src/services/api-gateway/internal/templates/ontologies/product.json` — Product, Category, Manufacturer, Review, Price
- `src/services/api-gateway/internal/templates/ontologies/software.json` — Component, API, Service, Dependency, Version
- `src/services/api-gateway/internal/templates/ontologies/medical.json` — Patient, Diagnosis, Treatment, Medication, Provider
- `src/services/api-gateway/internal/templates/README.md` — description of each template

**Files to change:**
- `src/services/api-gateway/routes.go` — add route: `GET /api/v1/templates/ontologies` — list available templates
- `src/services/api-gateway/routes.go` — add route: `POST /api/v1/ontologies/:id/apply-template` — apply template to ontology

**Key implementation details:**
- Templates stored as JSON files, loaded at gateway startup
- Each template defines: classes (with labels, parents, descriptions), properties (with domains/ranges), annotations
- `apply-template` calls ApplySequence gRPC to atomically add all template entities
- Templates are read-only (cannot be modified via API in M2)
- Template count: 5-8 templates covering common domains

**Unit tests (TDD Rule 5):** Write BEFORE implementation:
- `src/services/api-gateway/handlers/template_handler_test.go` — test template listing, application, missing template error

**Logging:** INFO on template load (count), DEBUG on template application (entity count), WARN on missing template.

**Dependencies:** Task 0.2 (ApplySequence gRPC), Task 3.5 (ApplySequence domain logic).

---

#### Task 4.5: Implement prompt injection defense

**Deliverable:** Two-level prompt injection defense for all LLM-facing endpoints, per `ADR-DES.SECURITY.prompt-injection-defense`.

**Files to create:**
- `src/services/api-gateway/middleware/prompt_injection.go` — pre-filter middleware for user-supplied content sent to LLM
- `src/services/api-gateway/middleware/prompt_injection_test.go` — unit tests for pre-filter rules

**Files to change:**
- `src/services/api-gateway/routes.go` — apply injection defense middleware to all AI routes (`/generate-from-text`, `/ai/*`, `/refine`)

**Key implementation details — Pre-filtering (VEDO-side):**
- Scan user-supplied text for known prompt injection patterns:
  - `Ignore previous instructions`, `Ignore all previous`, `Forget everything` — system prompt override attempts
  - `You are now`, `Act as`, `Pretend you are` — role-play injection
  - Base64-encoded hidden instructions, delimiter injection (`---`, `===\n`)
  - XML tag injection (`<system>`, `<instruction>`)
- Pattern matches are logged as `WARN` with full user input in the audit log
- On high-confidence match: return 400 with code `PROMPT-INJECTION-DETECTED` and a generic error message (do not reveal detection details to user)
- On low-confidence match: include an additional system prompt boundary instruction as defense-in-depth

**Key implementation details — System prompt hardening:**
- All prompt templates (Go `.tmpl` files and Python strings in `llm/prompts.py`) MUST include boundary markers:
  ```
  [SYSTEM BOUNDARY] You are an ontology extraction AI. [SYSTEM BOUNDARY]
  ```
- System prompt must explicitly instruct the LLM to ignore embedded instructions in user content

**Key implementation details — Post-processing:**
- After LLM response received, scan output for embedded instructions before parsing JSON
- If detected: log WARN, retry with additional hardening prompt
- If detected after max retries: return error to user, log to security audit

**Unit tests (TDD Rule 5):** Write BEFORE implementation:
- `prompt_injection_test.go` — test all pre-filter patterns (system override, role-play, base64, delimiter, XML), test false positives (legitimate text containing trigger words), test confidence level detection

**Logging:** WARN on each detected pattern (user_id, ontology_id, pattern_type, input_snippet), ERROR on repeated detection, audit log for all events. Do NOT log the full user input in plain text — use a hash digest for correlation.

**Dependencies:** Tasks 4.1 (NL→OWL endpoint needs protection first), 2.3 (prompt templates need hardening).

---

### Phase 5: Frontend — Preview, Upload & Apply

**Covered ADRs:**
- [`specs/adr/ADR-DES.UI.ai-suggestion-ux-strategy.md`](specs/adr/ADR-DES.UI.ai-suggestion-ux-strategy.md) — AI suggestion UX: ranked lists, confidence scores, caching, rate limiting
- [`specs/adr/ADR-IMPL.STACK.frontend-vue-strategy.md`](specs/adr/ADR-IMPL.STACK.frontend-vue-strategy.md) — Vue 3 + Composition API + Apollo Client + Vite
- [`specs/adr/ADR-DES.INFRA.otel-observability-strategy.md`](specs/adr/ADR-DES.INFRA.otel-observability-strategy.md) — frontend trace context propagation to backend services

---

#### Task 5.1: Build file upload UI component

**Deliverable:** Reusable Vue 3 component for document upload with drag & drop, format validation, and progress feedback.

**Files to create:**
- `src/services/frontend/src/components/ontology/DocumentUploader.vue` — drag & drop zone, file type icons, size display, progress bar
- `src/services/frontend/src/components/ontology/DocumentUploader.spec.ts` — vitest unit tests

**Component API:**
```typescript
// Props
interface DocumentUploaderProps {
  ontologyId: string
  maxFileSizeMb?: number       // default: 20
  allowedFormats?: string[]    // default: all 8 formats
  mode?: 'single' | 'batch'   // default: 'single'
}

// Emits
interface DocumentUploaderEmits {
  'upload-complete': (result: ExtractionPreview) => void
  'upload-error': (error: ExtractionError) => void
  'upload-progress': (progress: number) => void
}
```

**Key implementation details:**
- Drag & drop zone with visual feedback (border highlight on drag-over)
- Click-to-browse fallback
- File type validation by extension + magic bytes (client-side pre-check)
- Size validation with async mode suggestion for >20MB files
- Upload progress bar (using `axios` with `onUploadProgress`)
- Error states: wrong format, too large, network error, server error
- Disabled state while upload in progress
- Supports single + batch mode (switch via prop)

**Unit tests (TDD Rule 5):** Write BEFORE implementation using vitest + @vue/test-utils:
- Test drag-drop acceptance, file validation (type/size), upload progress emission, error states, disabled state

**Logging:** console.debug on file selection (name, size, type), console.info on upload start/complete, console.warn on validation failure.

**Dependencies:** Task 3.4 (document-extractor API routes — `POST /api/v1/documents/extract`), Task 0.4 (API Gateway routing — gateway must proxy extract endpoint).

---

#### Task 5.2: Build sequence preview table component

**Deliverable:** Interactive table for reviewing and editing extracted ontology sequences before applying.

**Files to create:**
- `src/services/frontend/src/components/ontology/SequencePreview.vue` — main preview component
- `src/services/frontend/src/components/ontology/SequencePreviewRow.vue` — single row with toggle, inline edit, source info
- `src/services/frontend/src/components/ontology/SequencePreview.spec.ts` — unit tests

**Component API:**
```typescript
interface SequencePreviewProps {
  steps: SequenceStep[]
  ontologyId: string
  sourceFiles?: string[]  // for batch — show file origin
}

interface SequencePreviewEmits {
  'apply': (steps: SequenceStep[]) => void
  'cancel': () => void
}
```

**Key implementation details (matching US-io.document.preview-sequence):**
- Table columns: checkbox (include/exclude), operation icon (class/property/individual), entity ID, label (editable inline), parent/domain/range
- Inline label editing: click → input field → Enter to save, Escape to cancel
- Duplicate detection: yellow warning badge + tooltip "⚠ class Customer already exists"
- Excluded rows: greyed out, strikethrough
- Source file column (batch mode): shows which file each step came from
- Bottom counter: "Will be imported: N of M steps"
- Group by source file (batch mode) with collapsible sections
- Responsive: table on desktop, card list on mobile

**Unit tests (TDD Rule 5):** Write BEFORE implementation using vitest + @vue/test-utils:
- Test toggle step inclusion, inline label editing, duplicate detection display, source file grouping, counter accuracy

**Logging:** console.debug on step toggle/edit, console.info on apply.

**Dependencies:** Task 5.1 (uploader provides data), Task 3.4 (API for extraction).

---

#### Task 5.3: Build apply workflow with progress and commit feedback

**Deliverable:** Apply button → progress → success/error with commit link.

**Files to create:**
- `src/services/frontend/src/components/ontology/ApplySequenceButton.vue` — apply button with confirmation modal
- `src/services/frontend/src/components/ontology/ApplyProgressModal.vue` — progress modal: "Importing... 2/3"
- `src/services/frontend/src/composables/useApplySequence.ts` — composable: calls API, manages state

**Files to change:**
- `src/services/frontend/src/apollo/queries.ts` — add `APPLY_SEQUENCE` mutation (or REST call)
- `src/services/frontend/src/pages/OntologyWorkspace.vue` — integrate apply workflow into workspace

**Workflow states:**
1. **Idle** — "Apply Import" button (with entity count)
2. **Confirm** — modal: summary of what will be created, final count, commit message input
3. **Applying** — progress bar with step counter "3/15", cancel disabled
4. **Success** — "✅ Imported: 15 entities" + link to commit `#abc123`
5. **Error** — "❌ Import failed: <reason>" + retry button

**Unit tests (TDD Rule 5):** Write BEFORE implementation using vitest + @vue/test-utils:
- Test workflow state transitions, progress tracking, error display, commit link generation

**Logging:** console.info on apply start, console.debug on each step completion, console.error on failure.

**Dependencies:** Task 5.2 (preview must be shown), Task 3.4 (API endpoint).

---

#### Task 5.4: Build batch upload UI with deduplication and conflict resolution

**Deliverable:** Multi-file upload, combined preview, conflict resolution UI.

**Files to create:**
- `src/services/frontend/src/components/ontology/BatchUploader.vue` — multi-file upload orchestrator
- `src/services/frontend/src/components/ontology/ConflictResolver.vue` — conflict resolution modal (name clashes between files)
- `src/services/frontend/src/composables/useBatchUpload.ts` — batch state management

**Files to change:**
- `src/services/frontend/src/components/ontology/DocumentUploader.vue` — batch mode integration
- `src/services/frontend/src/components/ontology/SequencePreview.vue` — source file grouping for batch

**Key implementation details:**
- Upload up to 10 files simultaneously (with per-file progress)
- After all files processed: combined preview with source file badges
- Deduplication: automatic by entity ID, with visual indicator
- Conflict resolution: when same entity ID has different labels across files → modal with options (keep file A, keep file B, custom)
- Partial failure handling: "1 of 3 files failed" notification, successful results still shown
- "Retry failed" button for individual failed files

**Unit tests (TDD Rule 5):** Write BEFORE implementation using vitest + @vue/test-utils:
- Test multi-file upload, deduplication logic, conflict resolution modal, partial failure handling

**Logging:** console.debug on each file completion, console.info on deduplication count, console.warn on conflicts.

**Dependencies:** Tasks 5.1, 5.2, 3.4.

---

### Phase 6: Integration, Validation & Documentation

**Covered ADRs:**
- [`specs/adr/ADR-IMPL.STACK.port-mapping-strategy.md`](specs/adr/ADR-IMPL.STACK.port-mapping-strategy.md) — validates gRPC port alignment in Docker Compose and CI
- [`specs/adr/ADR-IMPL.STACK.antora-docs-adoption.md`](specs/adr/ADR-IMPL.STACK.antora-docs-adoption.md) — Antora for documentation site, module structure for M2 docs
- [`specs/adr/ADR-DES.INFRA.otel-observability-strategy.md`](specs/adr/ADR-DES.INFRA.otel-observability-strategy.md) — validates OpenTelemetry instrumentation across all new services

---

#### Task 6.1: Docker Compose, Makefile, CI, and Docker quality gate

**Deliverable:** Full infrastructure integration for the new document-extractor service and LLM configuration.

**Files to create:**
- `deploy/docker-compose.llm.yaml` — optional LLM service overrides (local LLM container for dev)
- `src/services/document-extractor/.dockerignore`

**Files to change:**
- `deploy/docker-compose.yml` — add `document-extractor` service block:
  ```yaml
  document-extractor:
    build:
      context: ../src/services/document-extractor
    expose:
      - "8092"   # HTTP health
      - "9013"   # gRPC internal
    environment:
      - HOST=0.0.0.0
      - SERVICE_PORT=8092
      - ONTOLOGY_SERVICE_URL=grpc://ontology-service:9001
      - LLM_PROVIDER=${LLM_PROVIDER:-}
      - LLM_API_KEY=${LLM_API_KEY:-}
      - LLM_MODEL=${LLM_MODEL:-}
      - LLM_BASE_URL=${LLM_BASE_URL:-}
      - LLM_PROVIDERS_OLLAMA_BASE_URL=${LLM_PROVIDERS_OLLAMA_BASE_URL:-http://localhost:11434/v1}
      - LLM_PROVIDERS_OLLAMA_MODEL=${LLM_PROVIDERS_OLLAMA_MODEL:-llama3}
      - MAX_FILE_SIZE_MB=20
    depends_on:
      ontology-service:
        condition: service_healthy
    networks:
      - vedo-network
    healthcheck:
      <<: *healthcheck-defaults
      test: ["CMD-SHELL", "/bin/bash -ec 'exec 3<>/dev/tcp/127.0.0.1/8092; printf \"GET /health HTTP/1.1\\r\\nHost: localhost\\r\\nConnection: close\\r\\n\\r\\n\" >&3; grep -q \"200 OK\" <&3'"]
  ```
- `src/Makefile` — add `document-extractor` to `STUB_SERVICES`; add `docker-build-document-extractor` target
- `deploy/ci/gitlab-ci.yml` — add lint/build/test jobs for document-extractor:
  - `document-extractor-lint`: ruff + basedpyright
  - `document-extractor-test`: pytest
  - `document-extractor-build`: docker build
- `.env.example` — add LLM configuration variables
- `deploy/README.md` — add document-extractor and LLM section

**Logging:** Not applicable (infrastructure).

**Docker quality gate (RULES.md §Code Quality #6):** After `docker compose up -d`, verify:
- All containers healthy (`docker compose ps` — all `healthy`)
- Document-extractor `/health` returns 200
- Document-extractor `/ready` returns 200
- LLM connectivity test (if LLM provider configured)
- gRPC health probe on all internal services (9001-9013)

**Dependencies:** Tasks 3.1–3.4 (service must exist), Task 0.5 (gRPC ports aligned).

**Traceability:** Add document-extractor service, Docker Compose changes, CI jobs, and `.env.example` variables to `.ai-factory/traceability/traceability.ttl`.

---

#### Task 6.2: Integration test validation (LLM external interfaces)

**Deliverable:** Run and verify integration tests for LLM external interfaces (TDD Rule 4).

**Tests to execute:**
- `src/services/shared/llm/integration_test.go` — real HTTP calls to configured LLM backend (build tag `integration`)
- All unit tests from Phases 2–5 (already written JIT per task) must pass: `go test ./...`, `cargo test`, `pytest`, `npx vitest`

**Key implementation details:**
- Integration tests are SKIPPED by default; run with `LLM_INTEGRATION_TESTS=1 go test -tags=integration ./src/services/shared/llm/...`
- Test against the configured LLM provider (set via `LLM_PROVIDER` env var; falls back to OpenAI-compatible local endpoint)
- Verify: real prompt → response parsing, timeout handling, token counting accuracy
- All unit tests from implementation phases must pass before this task is marked complete

**Logging:** Not applicable (tests).

**Dependencies:** All implementation tasks from Phases 2–5.

**Traceability:** Add integration test suite to `.ai-factory/traceability/traceability.ttl` with `vdo:validates` links to LLM provider implementations.

---

#### Task 6.3: Traceability and documentation validation

**Deliverable:** Verify traceability.ttl is complete for all M2 artifacts and all Antora docs are generated.

**Validation steps:**
1. Run change impact analysis: query `.ai-factory/traceability/traceability.ttl` for all artifacts added/modified in Phases 0–5
2. Verify every new service, API endpoint, test suite, proto definition, and deployment config has a corresponding `vdo:` entry
3. Verify relationships: `vdo:implements` for code→spec, `vdo:validates` for test→story, `vdo:constrains` for ADR→implementation
4. Fix any missing entries
5. Build Antora docs: `cd src/docs/antora && antora antora-playbook.yml` — verify zero build errors
6. Verify all M2-related pages are rendered: ai-creation, document-extractor, gRPC modules

**Logging:** INFO on traceability query, WARN on missing entries, INFO on Antora build result.

**Dependencies:** Task 6.2 (tests must pass first).

---

#### Task 6.4: Antora documentation

**Deliverable:** User-facing and developer documentation for M2 features.

**Files to create:**
- `src/docs/antora/user-guide/modules/ai-creation/pages/index.adoc` — overview: what is AI-assisted ontology creation
- `src/docs/antora/user-guide/modules/ai-creation/pages/document-extraction.adoc` — how to extract ontology from documents (step-by-step with screenshots)
- `src/docs/antora/user-guide/modules/ai-creation/pages/nl-generation.adoc` — NL→OWL generation guide
- `src/docs/antora/user-guide/modules/ai-creation/pages/templates.adoc` — using ontology templates
- `src/docs/antora/user-guide/modules/ai-creation/pages/batch-upload.adoc` — batch upload guide
- `src/docs/antora/user-guide/modules/ai-creation/pages/refinement.adoc` — iterative refinement guide
- `src/docs/antora/user-guide/modules/ai-creation/nav.adoc` — navigation for AI creation section
- `src/docs/antora/developer-guide/modules/document-extractor/pages/index.adoc` — document-extractor service architecture
- `src/docs/antora/developer-guide/modules/document-extractor/pages/parsers.adoc` — parser design and extension guide
- `src/docs/antora/developer-guide/modules/document-extractor/nav.adoc`
- `src/docs/antora/developer-guide/modules/gRPC/pages/index.adoc` — gRPC usage, proto generation, service contracts
- `src/docs/antora/developer-guide/modules/gRPC/nav.adoc`
- `src/docs/antora/integrator-guide/modules/api/pages/grpc-reference.adoc` — gRPC API reference for external integrators
- `src/docs/antora/integrator-guide/modules/api/pages/llm-config.adoc` — LLM provider configuration

**Files to change:**
- `src/docs/antora/user-guide/antora.yml` — add `ai-creation` module
- `src/docs/antora/developer-guide/antora.yml` — add `document-extractor` and `gRPC` modules
- `src/docs/antora/integrator-guide/antora.yml` — update API module nav
- `src/docs/antora/antora-playbook.yml` — verify all modules included

**Logging:** Not applicable (documentation).

**Dependencies:** All implementation tasks complete (docs are the final deliverable).

---

## Commit Plan

| Commit | Tasks | Message |
|--------|-------|---------|
| 1 | 0.1 | `feat(proto): define gRPC service contracts for all internal services` |
| 2 | 0.2, 0.3 | `feat(grpc): migrate Rust and Go services to gRPC servers` |
| 3 | 0.4 [~], 0.5, 0.6 | `feat(grpc): update API Gateway to gRPC proxy, Docker Compose config, and proto generation` |
| 4 | 1.1, 1.2 | `test(e2e): add E2E test specifications for M2 document extraction and AI flows` |
| 5 | 2.1, 2.2 | `feat(llm): add LLM abstraction layer with runtime-configured provider registry` |
| 6 | 2.3, 2.4, 2.5 | `feat(llm): add prompt templates, observability, and LLM Policy Router` |
| 7 | 3.1, 3.2 | `feat(extractor): scaffold document-extractor service with text parsers` |
| 8 | 3.3, 3.4, 3.5 | `feat(extractor): add structured data parsers, ApplySequence domain logic, and LLM/gRPC integration` |
| 9 | 4.1, 4.2, 4.5 | `feat(ai): implement NL-to-OWL generation, AI-assisted completion, and prompt injection defense` |
| 10 | 4.3, 4.4 | `feat(ai): add iterative refinement and domain templates` |
| 11 | 5.1, 5.2 | `feat(ui): build file upload and sequence preview components` |
| 12 | 5.3, 5.4 | `feat(ui): build apply workflow and batch upload UI` |
| 13 | 6.1 | `feat(infra): add document-extractor to Docker Compose, CI, and quality gate` |
| 14 | 6.2 | `test: integration test validation for LLM external interfaces` |
| 15 | 6.3, 6.4 | `chore: validate traceability and add Antora documentation` |

---

## Dependencies Map

```
Phase 0 (gRPC — PREREQUISITE)
  0.1 (proto defs)
   ├─► 0.2 (Rust gRPC)
   ├─► 0.3 (Go gRPC)
   │
   ├─► 0.5 (Docker/CI update) ← depends on 0.2, 0.3, 0.4
   └─► 0.6 (proto generation + Go gRPC clients) ← depends on 0.5
        │
        └─► 0.4 [~] (API Gateway gRPC proxy) ← depends on 0.2, 0.3, 0.6

Phase 1 (E2E Tests — TDD FIRST)
  1.1, 1.2 ← depend on Phase 0 (infrastructure must be running)
  These are the acceptance gate — implementation phases make them pass.

Phase 2 (LLM Provider Integration)
  2.1 (abstraction)
   └─► 2.2 (providers)
        ├─► 2.4 (observability + integration tests)
        └─► 2.3 (templates)
              └─► 2.5 (LLM Policy Router) ← depends on 2.2, 0.4

Phase 3 (Document Extractor & ApplySequence)
  3.1 (scaffold) ← depends on 0.2, 2.2
   ├─► 3.2 (text parsers)
   ├─► 3.3 (structured parsers)
   └─► 3.4 (LLM/gRPC integration) ← depends on 3.2, 3.3
        └─► 3.5 (ApplySequence domain logic) ← depends on 3.4, 2.2

Phase 4 (NL→OWL & AI)
  4.1, 4.2 ← depend on 2.2, 2.3, 0.6 (gRPC clients)
  4.3 ← depends on 4.1, 2.3
  4.4 ← depends on 0.2, 3.5 (ApplySequence)
  4.5 ← depends on 4.1, 2.3

Phase 5 (Frontend)
  5.1 → 5.2 → 5.3; 5.4 depends on 5.1, 5.2
  5.1 depends on 3.4 (API routes), 0.4 (gateway routing)
  All depend on Phase 3 (API endpoints) and Phase 4

Phase 6 (Integration & Validation)
  6.1 ← depends on Phase 0, 3 (including 3.5)
  6.2 ← depends on all implementation phases (2–5, including 2.5, 3.5, 4.5)
  6.3, 6.4 ← depends on 6.2
```

---

## Risk Register

| Risk | Probability | Impact | Mitigation |
|------|-----------|--------|------------|
| gRPC migration breaks existing API contracts | Medium | High | Run full E2E test suite after each migrated service; keep HTTP health endpoints unchanged |
| LLM provider API changes break extraction | Medium | Medium | Versioned provider implementations; integration tests with mock LLM servers |
| PDF/DOCX parsing fails on complex documents | Medium | Medium | Graceful degradation — return partial structure; allow manual correction in preview |
| LLM generation quality inconsistent | High | Medium | Prompt engineering iterations; user-editable preview; refinement workflow |
| Large document (>20MB) processing timeout | Low | Medium | Async processing mode; streaming upload; configurable timeouts |
| gRPC port conflicts in local dev | Low | Low | `.env` overridable ports; `deploy/README.md` port reference |
