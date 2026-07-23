## Test Plan: Ontology Core Test Coverage

**Date:** 2026-07-23
**Branch:** feature/test-improvements
**Environment:** Local dev (Docker Compose test stack) + CI

---

### 1. Testing Goal

Verify that the ontology engine's three core domains (TBox/ABox Editor, Versioning, Import/Export)
function correctly by implementing automated tests for the 27 uncovered P0 requirements.
Goal: raise RCS for ontology core from 0% to ≥ 80% coverage within these domains.

---

### 2. Test Scope

**In Scope** — we test:

- **TBox Editor** — Class CRUD, property CRUD, hierarchy validation, SHACL validation,
  annotation management, cycle detection, IRI uniqueness
- **ABox Editor** — Individual CRUD, property value assignment, batch operations, inline editing
- **Versioning Service** — Commit creation and history, branch operations (create/switch/merge),
  diff computation, rollback, merge conflict resolution
- **Import/Export** — Turtle export (canonical), DOCX export, XLSX import/export, snapshot
  publishing, RDF/XML import
- **REST API** — TBox/ABox CRUD endpoints via API Gateway, versioning endpoints, import/export
  endpoints

**Out of Scope** — we don't test:

- Real-time collaboration (WebSocket, node locking) — covered by commenting-service tests
- LLM integration (NL→OWL, AI completion) — covered by pending E2E skeletons
- Performance / load testing — SLA verification requires dedicated infrastructure
- Security / authorization — covered by existing BOLA/BFLA test suites

---

### 3. Test Types

| Type | Priority | Area |
|------|----------|------|
| Unit (Rust) | 🔴 High | ontology-service class/property CRUD, versioning-service commit/branch logic |
| Integration (Rust → Neo4j/Postgres) | 🔴 High | Cross-service data flow: create class → verify in Neo4j, commit → verify in Postgres |
| Integration (Go → Rust gRPC) | 🟡 Medium | API Gateway proxies to ontology/versioning service |
| E2E (Playwright) | 🟡 Medium | Full user workflow: login → create ontology → edit → commit → export |
| Negative / Edge cases | 🟡 Medium | Invalid data, cycle detection, IRI collision, empty ontology |
| Round-trip (Import/Export) | 🟡 Medium | Export Turtle → import → compare entity count and structure |

---

### 4. Test Data

| Category | Data | Purpose |
|----------|------|---------|
| Valid ontology | 5 classes (Person, Student, Employee, Organization, Project) + properties | Happy path CRUD |
| Cyclic hierarchy | Class A → subclass B → subclass A | Cycle detection negative test |
| Large ontology | 1000 classes, 500 properties | Scale test (optional) |
| Turtle file | canonical serialization of test ontology | Export/import round-trip |
| XLSX file | ontology with classes + properties in columns | XLSX import |

---

### 5. Preconditions

- [ ] Docker Compose test stack is running (`docker compose -f deploy/docker-compose.test.yml up -d`)
- [ ] Neo4j is accessible with test database
- [ ] PostgreSQL is accessible with test schema
- [ ] API Gateway is running on localhost:9000
- [ ] Frontend is running on localhost:3000
- [ ] Test users are seeded (alice viewer, bob editor, eve admin)
- [ ] Rust test toolchain is installed (cargo, clippy, nextest)
- [ ] Go test toolchain is installed
- [ ] Node.js and Playwright are installed

---

### 6. Acceptance Criteria

- [ ] All 🔴 high-priority unit tests pass
- [ ] Integration tests pass against running Docker Compose stack
- [ ] Negative tests verify expected error codes
- [ ] Round-trip tests confirm data integrity
- [ ] Coverage for 27 P0 REQ is registered in traceability.ttl with `vdo:validates` links
- [ ] All new test files include `// Validates: REQ-...` annotations

---

### 7. Plan Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Neo4j/PostgreSQL not available in test env | High — integration tests cannot run | Use in-memory stubs for unit tests; integration tests behind feature flag |
| Large test volume (27 REQ x multiple scenarios) | Medium — may exceed implementation budget | Prioritize unit tests over E2E; use table-driven tests |
| Rust compilation in CI is slow | Medium — increases pipeline time | Use incremental compilation, `cargo nextest` for test-only runs |
| Existing pending skeletons conflict with new tests | Low — test files are already created with `test.describe.skip` | Remove `.skip` when implementing; merge into existing files |

---

### 8. Checklist

| Check | Priority |
|-------|----------|
| ontology-service: class CRUD unit tests | High |
| ontology-service: hierarchy validation (cycle detection) | High |
| ontology-service: property CRUD (data + object) | High |
| ontology-service: IRI uniqueness enforcement | High |
| ontology-service: SHACL validation on class edit | Medium |
| ontology-service: individual CRUD | High |
| ontology-service: batch individual operations | Medium |
| versioning-service: commit creation and history | High |
| versioning-service: branch create/switch/merge | High |
| versioning-service: diff computation | High |
| versioning-service: merge conflict resolution | Medium |
| versioning-service: rollback | Medium |
| import/export: canonical Turtle export | High |
| import/export: DOCX export | Medium |
| import/export: XLSX import/export | Medium |
| import/export: snapshot publishing | Medium |
| API Gateway: TBox/ABox CRUD proxy tests | Medium |
| API Gateway: versioning endpoint tests | Medium |
| API Gateway: import/upload endpoint tests | Medium |
| E2E: full ontology lifecycle (create → edit → commit → export) | Medium |
| Round-trip: export Turtle → import → compare | Medium |
| TTL update: register all new validates links | High |
