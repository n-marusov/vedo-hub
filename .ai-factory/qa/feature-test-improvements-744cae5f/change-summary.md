## Change Summary — Ontology Core Test Coverage Gap

**Risk level:** 🔴 High

**Scope:** Planning tests for 3 ontology core domains that currently have zero test coverage
in the traceability graph.

---

### What Changed

This is not a code change — it is a **test coverage gap analysis** for the ontology engine's
core functional areas. Of 184 P0-level requirements, only 12 (6.5%) have any test coverage
tracked in traceability.ttl.

The 172 uncovered P0 REQ include 27 high-impact requirements in three business-critical domains:

| Domain | Uncovered P0 REQ | Key services affected |
|--------|-----------------|----------------------|
| **TBox/ABox Editor** | 13 | ontology-service (Rust), frontend (Vue 3) |
| **Versioning** | 4 | versioning-service (Rust), api-gateway (Go) |
| **Import/Export** | 10 | document-extractor (Python), publisher-service (Rust), api-gateway (Go) |

---

### Affected Areas

| Component | Change type | Description |
|-----------|------------|-------------|
| ontology-service (Rust) | Missing tests | Class/property/individual CRUD, SHACL validation, hierarchy |
| versioning-service (Rust) | Missing tests | Commits, branches, merges, diffs, delta computation |
| publisher-service (Rust) | Missing tests | Snapshot publishing, Turtle export |
| document-extractor (Python) | Missing tests | Excel import, validation import, structured data |
| api-gateway (Go) | Missing tests | REST versioning, validation import routes |
| frontend (Vue 3) | Missing tests | TBox editor UI, ABox editor UI, query export UI |

---

### Evidence

| Finding | Evidence |
|---------|----------|
| 13 P0 Editor REQ uncovered | `REQ-USR.UI.tbox-editor`, `REQ-USR.UI.abox-editor`, `REQ-FUN.API.class-hierarchy-accuracy`, etc. |
| 4 P0 Versioning REQ uncovered | `REQ-FUN.DATA.versioning`, `REQ-FUN.DATA.major-version-migration`, `REQ-FUN.DATA.major-version-policy`, `REQ-FUN.API.rest-versioning` |
| 10 P0 Import/Export REQ uncovered | `REQ-FUN.API.excel-import-block`, `REQ-FUN.API.validation-import`, `REQ-FUN.DATA.gdpr-export-mandate`, etc. |
| Existing test patterns exist | Security tests use `_test.go` + Playwright E2E; pending skeletons exist for US stories |
| Test infrastructure available | Docker Compose test stack, Neo4j, PostgreSQL, Playwright |

---

### Risks

🔴 **Critical** (must verify):

- **Ontology data corruption risk:** Without tests for class/property CRUD, invalid data or
  broken constraints (cycles, IRI collisions) may go undetected in production
- **Versioning data loss risk:** Commit/merge logic without tests could lead to silent data
  loss or incorrect diff computation affecting collaboration workflows

🟡 **Medium** (should verify):

- **Import/Export regression:** Turtle and XLSX round-trips may break silently without coverage
- **API contract drift:** REST versioning endpoints may deviate from API contract without
  integration tests

🟢 **Low** (nice to verify):

- **Performance regression:** Ontology operations have p95 < 120ms SLAs; no performance tests exist

---

### Testing Recommendations

**First priority:**

- [ ] **Unit tests for ontology-service (Rust):** Class create/read/update/delete, property CRUD,
  hierarchy validation, cycle detection, IRI uniqueness
- [ ] **Unit tests for versioning-service (Rust):** Commit creation, branch operations, merge
  resolution, delta computation, rollback
- [ ] **Integration tests for REST endpoints:** TBox/ABox CRUD via API Gateway, versioning
  endpoints, import/export endpoints

**Regression:**

- [ ] E2E smoke tests for ontology workspace: create class → add property → commit → verify export
- [ ] Round-trip tests: export Turtle → import → compare entities
