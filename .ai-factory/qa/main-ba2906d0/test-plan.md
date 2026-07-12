# Test Plan: Test Coverage for vedo-hub

> **Branch:** `main` | **Based on:** `change-summary.md` | **Date:** 2026-07-12

## Test Scope

### In Scope

Based on the critical and high-severity risks identified in the change summary, the following areas are in scope for new test development:

1. **Core domain services (Rust)** — ontology-service, versioning-service, publisher-service
2. **Frontend unit testing** — frontend (Vue 3 + TypeScript)
3. **Commenting service (Go)** — real-time collaboration
4. **CLI integration tests** — filling stub functions with real assertions
5. **Python services** — metrics-service, ticket-classifier
6. **Cross-service integration tests** — database connectivity, API contracts
7. **Regression test suite** — automated regression coverage

### Out of Scope

- Performance/load testing (separate effort, post-MVP)
- Chaos engineering / resilience testing (separate effort)
- Visual regression testing (nice-to-have, post-MVP)
- Infrastructure/Pipeline testing (already covered by shell scripts)

## Test Types

### 1. Unit Tests — Core Domain Logic (Priority: Critical)

| Area | Component | Current Coverage | Target Coverage | Effort |
|------|-----------|-----------------|-----------------|--------|
| Ontology CRUD | ontology-service (Rust) | 0 tests | 25+ tests | 3-4 days |
| Git-like versioning | versioning-service (Rust) | 0 tests | 30+ tests | 4-5 days |
| Ontology publishing | publisher-service (Rust) | 0 tests | 15+ tests | 2-3 days |
| Public browse | public-browse-api (Rust) | 0 tests | 10+ tests | 1-2 days |

**Rationale:** These are the core domain services implementing the primary business value. Zero test coverage is a critical risk. Each service needs:
- Unit tests for business logic functions
- Integration tests for Neo4j/PostgreSQL database operations
- Error handling and edge case tests

### 2. Unit Tests — Go Services with Gaps (Priority: High)

| Service | Current | Target | Notes |
|---------|---------|--------|-------|
| commenting-service | 0 tests | 25+ tests | Real-time collaboration, WebSocket handling |
| api-gateway | 24 tests | 40+ tests | Add integration tests for proxy routing |
| support-service | 21 tests | 35+ tests | Add more edge cases for SLA/WORM |
| ticket-notifier | 13 tests | 25+ tests | Add notification channel integration tests |
| ticket-sync | 12 tests | 20+ tests | Add more sync failure scenarios |
| ticket-telemetry-listener | 21 tests | 30+ tests | Add more telemetry format edge cases |

### 3. Frontend Unit Tests (Priority: High)

| Component | Current | Target | Framework |
|-----------|---------|--------|-----------|
| Vue components | 0 tests | 30+ tests | Vitest |
| Composables (logic) | 0 tests | 20+ tests | Vitest |
| Pinia stores | 0 tests | 10+ tests | Vitest |
| Apollo GraphQL operations | 0 tests | 15+ tests | Vitest + MSW |
| Vue Flow graph components | 0 tests | 15+ tests | Vitest + happy-dom |

**Total target: 90+ tests across the frontend.**

**Rationale:** The UI is a complex single-page application with graph visualization (Vue Flow), real-time updates (GraphQL subscriptions), and multi-panel layout. Without unit tests, regressions are likely after every component change.

### 4. Python Service Tests (Priority: Medium)

| Service | Current | Target | Framework |
|---------|---------|--------|-----------|
| metrics-service | 0 tests | 15+ tests | pytest |
| ticket-classifier | 0 tests | 20+ tests | pytest |

**Rationale:** Both services are currently stub HTTP servers. When implemented with actual business logic (metrics aggregation, ML classification), tests are essential.

### 5. CLI Integration Tests — Fill Stubs (Priority: Medium)

| Test | Current | Target | Status |
|------|---------|--------|--------|
| TestCliIntegration_UnsupportedCommand | Stub (empty) | Full implementation | ⚠️ Fill |
| TestCliIntegration_InvalidFormat | Stub (empty) | Full implementation | ⚠️ Fill |
| TestCliIntegration_ComposeDiagnosticsFailure | Stub (empty) | Full implementation | ⚠️ Fill |
| TestCliIntegration_AuditRedactionFailure | Stub (empty) | Full implementation | ⚠️ Fill |
| TestCliIntegration_NoSecretsInOutput | Stub (empty) | Full implementation | ⚠️ Fill |
| TestCliIntegration_CredentialChainOrder | Stub (empty) | Full implementation | ⚠️ Fill |

### 6. Integration / Contract Tests (Priority: High)

| Area | Current | Target | Notes |
|------|---------|--------|-------|
| Neo4j integration | 0 tests | 15+ tests | Ontology CRUD operations |
| PostgreSQL integration | 0 tests | 10+ tests | Version store operations |
| Redis integration | 0 tests | 5+ tests | Caching and locks |
| RabbitMQ integration | 0 tests | 5+ tests | Message queue operations |
| API contract tests | 0 tests | 20+ tests | OpenAPI schema conformance |
| gRPC inter-service tests | 0 tests | 10+ tests | Service-to-service contracts |

### 7. Regression Test Suite (Priority: High)

| Type | Scope | Automation |
|------|-------|-----------|
| Weekly regression | Full E2E suite + critical paths | CI pipeline |
| Smoke tests per build | Health checks, core flows | CI pipeline |
| Security regression | Authorization tests (existing) | CI pipeline (+ expand) |

## Verification Checklist

- [ ] **HIGH** — ontology-service: implement 25+ unit and integration tests
- [ ] **HIGH** — versioning-service: implement 30+ unit and integration tests
- [ ] **HIGH** — publisher-service: implement 15+ unit and integration tests
- [ ] **HIGH** — commenting-service: implement 25+ tests for collaboration logic
- [ ] **HIGH** — frontend: add Vitest and write 90+ component/composable/store tests
- [ ] **HIGH** — Add database integration tests (Neo4j, PostgreSQL, Redis)
- [ ] **HIGH** — Add API contract tests (OpenAPI schema validation)
- [ ] **MEDIUM** — Fill 6 CLI integration test stubs with real assertions
- [ ] **MEDIUM** — metrics-service: add pytest and write 15+ tests
- [ ] **MEDIUM** — ticket-classifier: add pytest and write 20+ tests
- [ ] **MEDIUM** — api-gateway: add integration tests for proxy routing
- [ ] **MEDIUM** — Expand support-service tests with more edge cases
- [ ] **LOW** — public-browse-api (Rust): add basic 10+ tests
- [ ] **LOW** — publish-browse-ui: add basic smoke tests
