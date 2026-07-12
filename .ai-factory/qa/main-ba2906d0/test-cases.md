# Test Cases: vedo-hub Test Coverage

> **Branch:** `main` | **Based on:** `change-summary.md`, `test-plan.md` | **Date:** 2026-07-12

---

## Priority Legend

| Priority | Definition |
|----------|------------|
| **Critical** | Core business logic failure — blocks deployment |
| **High** | Important functionality — should block unless mitigated |
| **Medium** | Supporting functionality — should be included |
| **Low** | Nice-to-have — cosmetic or rare edge cases |

---

## 1. Core Domain Services (Rust) — Unit & Integration Tests

### TC-001: Ontology — Create class with valid IRI
| Field | Value |
|-------|-------|
| **Priority** | Critical |
| **Preconditions** | ontology-service running, Neo4j connection available |
| **Test data** | `{"iri": "http://example.org/ontology#Person", "label": "Person", "description": "A person entity"}` |
| **Steps** | 1. POST `/api/v1/ontologies/{id}/classes` with valid JSON payload<br>2. Verify HTTP 201<br>3. GET `/api/v1/ontologies/{id}/classes/{classId}` |
| **Expected result** | Class created and retrievable. Response includes `iri`, `label`, `description`, `created_at`, `version` |
| **Test data notes** | Use unique ontology and class IRIs per run |

### TC-002: Ontology — Create class with duplicate IRI
| Field | Value |
|-------|-------|
| **Priority** | Critical |
| **Preconditions** | Existing class with IRI `http://example.org/ontology#Person` |
| **Test data** | Same IRI as existing class |
| **Steps** | 1. POST same class IRI again<br>2. Verify HTTP 409 Conflict |
| **Expected result** | Error response with code `DUPLICATE_IRI`, message about existing resource |

### TC-003: Ontology — Create class with invalid IRI format
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | ontology-service running |
| **Test data** | `{"iri": "not-a-valid-iri", "label": "Bad IRI"}` |
| **Steps** | 1. POST with invalid IRI<br>2. Verify HTTP 400 |
| **Expected result** | Validation error with explanation of valid IRI format |

### TC-004: Ontology — Create relationship between two classes
| Field | Value |
|-------|-------|
| **Priority** | Critical |
| **Preconditions** | Two existing classes with known IDs |
| **Test data** | `{"source": "{classA}", "target": "{classB}", "type": "subClassOf"}` |
| **Steps** | 1. POST `/relationships` with valid relationship data<br>2. Verify HTTP 201<br>3. GET relationship by ID |
| **Expected result** | Relationship created and retrievable. Response includes source, target, type metadata |

### TC-005: Ontology — Delete class with relationships
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | Class has one or more relationships to other classes |
| **Steps** | 1. DELETE `/classes/{classId}`<br>2. Verify response code |
| **Expected result** | HTTP 409 if cascade delete is not allowed, or HTTP 200 with cascade. Relationships referencing the class are also removed |
| **Test data notes** | Both cascade and no-cascade modes should be tested |

### TC-006: Ontology — Search classes by label
| Field | Value |
|-------|-------|
| **Priority** | Medium |
| **Preconditions** | Multiple classes exist with labels containing "Person" |
| **Test data** | Search query: `Person` |
| **Steps** | 1. GET `/classes?search=Person`<br>2. Verify results |
| **Expected result** | Returns all classes whose labels contain "Person". Results sorted by relevance |

### TC-007: Versioning — Create commit with changes
| Field | Value |
|-------|-------|
| **Priority** | Critical |
| **Preconditions** | Existing ontology with at least one change made since last commit |
| **Test data** | `{"message": "Added Person class", "author": "user@example.com"}` |
| **Steps** | 1. Make a change to ontology<br>2. POST `/api/v1/ontologies/{id}/commits` with commit message<br>3. Verify HTTP 201 |
| **Expected result** | Commit created. Response includes `commit_id`, `message`, `author`, `timestamp`, `parent_commit_id` |

### TC-008: Versioning — Create branch from commit
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | Existing commit in ontology |
| **Test data** | `{"name": "feature/new-classes", "source_commit": "{commitId}"}` |
| **Steps** | 1. POST `/branches` with valid branch data<br>2. Verify HTTP 201<br>3. GET `/branches` to list |
| **Expected result** | Branch created and visible in branch list. First commit in branch equals source commit |

### TC-009: Versioning — Merge branch with conflicts
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | Two branches with conflicting changes to the same class |
| **Steps** | 1. POST `/merges` between branches<br>2. Verify response |
| **Expected result** | HTTP 409 Conflict with details about conflicting axioms. Merge resolution required before proceeding |

### TC-010: Versioning — Get diff between two commits
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | Two commits with different ontology states |
| **Steps** | 1. GET `/diff?from={commitA}&to={commitB}`<br>2. Verify response |
| **Expected result** | Structured diff showing added, removed, and modified axioms. Format matches ontology diff specification |

### TC-011: Publisher — Publish ontology version
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | Commit with ontology state exists |
| **Test data** | `{"commit_id": "{commitId}", "visibility": "public"}` |
| **Steps** | 1. POST `/publish` with commit and visibility<br>2. Verify HTTP 201<br>3. GET `/published/{ontologyId}` |
| **Expected result** | Ontology published and accessible via public browse API |

### TC-012: Publisher — Publish with invalid commit ID
| Field | Value |
|-------|-------|
| **Priority** | Medium |
| **Preconditions** | None |
| **Test data** | `{"commit_id": "non-existent-uuid", "visibility": "public"}` |
| **Steps** | 1. POST `/publish` with non-existent commit<br>2. Verify response |
| **Expected result** | HTTP 404 with error code `COMMIT_NOT_FOUND` |

---

## 2. Commenting Service (Go) — Unit & Integration Tests

### TC-013: Comment — Add comment to ontology class
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | Authenticated user, existing ontology class |
| **Test data** | `{"class_id": "{classId}", "body": "This class should be renamed", "mentions": ["user2@example.com"]}` |
| **Steps** | 1. POST `/api/v1/comments` with valid comment<br>2. Verify HTTP 201<br>3. GET `/comments?class_id={classId}` |
| **Expected result** | Comment created and retrievable. Response includes `id`, `author`, `body`, `created_at`, `mentions` |

### TC-014: Comment — Add comment with empty body
| Field | Value |
|-------|-------|
| **Priority** | Medium |
| **Preconditions** | Authenticated user |
| **Test data** | `{"class_id": "{classId}", "body": ""}` |
| **Steps** | 1. POST with empty body<br>2. Verify response |
| **Expected result** | HTTP 400 with validation error for empty body |

### TC-015: Comment — WebSocket real-time notification
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | Two WebSocket connections established (User A and User B) |
| **Test data** | User A posts comment on class |
| **Steps** | 1. User A POST comment<br>2. Verify User A receives confirmation<br>3. Verify User B receives real-time notification via WebSocket |
| **Expected result** | Both users receive the comment event. Notification includes comment metadata |

### TC-016: Comment — Delete own comment
| Field | Value |
|-------|-------|
| **Priority** | Medium |
| **Preconditions** | Existing comment by current user |
| **Steps** | 1. DELETE `/comments/{commentId}`<br>2. Verify HTTP 200<br>3. GET comment by ID |
| **Expected result** | Comment marked as deleted (soft delete). Body replaced with "[deleted]" |

### TC-017: Comment — Delete another user's comment (unauthorized)
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | Existing comment by different user |
| **Steps** | 1. DELETE `/comments/{commentId}` as another user<br>2. Verify response |
| **Expected result** | HTTP 403 with error code `FORBIDDEN` |

---

## 3. Frontend (Vue 3 + TypeScript) — Unit & Component Tests

### TC-018: Frontend — Login page renders OAuth providers
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | Vitest + vue-test-utils set up |
| **Steps** | 1. Mount `LoginPage` component<br>2. Verify all 5 OAuth provider buttons are rendered<br>3. Verify login form elements exist |
| **Expected result** | Component renders without errors. OAuth buttons for Google, Yandex, VK, ESIA, Gosuslugi are present |

### TC-019: Frontend — Dashboard displays greeting and widgets
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | Mocked Apollo client with authenticated user |
| **Steps** | 1. Mount `DashboardPage` component<br>2. Verify greeting with user name<br>3. Verify widget components render |
| **Expected result** | "Welcome, {user}" greeting displayed. Dataset/Ontology widgets show correct mock data |

### TC-020: Frontend — Workspace 3-panel layout
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | Mocked ontology and class data |
| **Steps** | 1. Mount `WorkspacePage` component<br>2. Verify class hierarchy panel<br>3. Verify graph visualization panel (Vue Flow)<br>4. Verify property/axiom panel |
| **Expected result** | Three-panel layout renders correctly. Graph shows nodes and edges. Panels are resizable |

### TC-021: Frontend — Ontology search composable
| Field | Value |
|-------|-------|
| **Priority** | Medium |
| **Preconditions** | Vitest with mocked Apollo query |
| **Steps** | 1. Call `useOntologySearch()` composable<br>2. Set search term to "Person"<br>3. Wait for results |
| **Expected result** | Returns filtered list of classes matching search term. Loading state transitions correctly. Error state handled |

### TC-022: Frontend — Vue Flow node selection
| Field | Value |
|-------|-------|
| **Priority** | Medium |
| **Preconditions** | Graph component mounted with test nodes |
| **Steps** | 1. Click on a node in the graph<br>2. Verify node is selected (highlighted)<br>3. Verify property panel updates with selected node data |
| **Expected result** | Selected node gets visual highlight. Properties panel shows the selected class details |

### TC-023: Frontend — Form validation (create class)
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | Create class dialog/modal open |
| **Steps** | 1. Submit form with empty IRI<br>2. Submit form with invalid IRI format<br>3. Submit form with valid data |
| **Expected result** | Empty IRI shows validation error. Invalid IRI shows format error. Valid data submits successfully |

### TC-024: Frontend — Pinia store for ontology state
| Field | Value |
|-------|-------|
| **Priority** | Medium |
| **Preconditions** | Mocked store with initial state |
| **Steps** | 1. Dispatch `addClass` action<br>2. Verify store state updates<br>3. Verify getters reflect new state |
| **Expected result** | Store state is immutable (not mutated directly). Getters return correct computed values |

---

## 4. CLI Integration Tests — Fill Stubs

### TC-025: CLI — Unsupported command returns error
| Field | Value |
|-------|-------|
| **Priority** | Medium |
| **Preconditions** | vedo-cli binary built |
| **Test data** | `vedo-cli unsupported-command` |
| **Steps** | 1. Execute CLI with invalid subcommand<br>2. Verify exit code != 0<br>3. Verify stderr contains error message |
| **Expected result** | Non-zero exit code. Error message lists available commands |

### TC-026: CLI — Invalid output format
| Field | Value |
|-------|-------|
| **Priority** | Medium |
| **Preconditions** | vedo-cli binary built |
| **Test data** | `vedo-cli ticket list --format invalid` |
| **Steps** | 1. Execute CLI with invalid format flag<br>2. Verify exit code != 0<br>3. Verify error mentions valid formats |
| **Expected result** | Non-zero exit code. Error message lists supported formats (json, yaml, table) |

### TC-027: CLI — No secrets leaked in output
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | vedo-cli binary built, test credentials configured |
| **Test data** | `vedo-cli ticket get {id}` with sensitive fields |
| **Steps** | 1. Execute CLI command<br>2. Check all output for regex patterns matching secrets (API keys, tokens, passwords) |
| **Expected result** | No secrets/credentials visible in stdout or stderr. Sensitive fields are redacted |

### TC-028: CLI — Credential chain resolution
| Field | Value |
|-------|-------|
| **Priority** | Medium |
| **Preconditions** | Simulated credential sources (env, file) |
| **Steps** | 1. Set env var for credential<br>2. Execute CLI command that uses credential<br>3. Verify correct source used |
| **Expected result** | Credential resolution follows priority chain: Vault > AWS > env > error |

---

## 5. Python Services

### TC-029: Metrics — Health endpoint returns status
| Field | Value |
|-------|-------|
| **Priority** | Medium |
| **Preconditions** | metrics-service running |
| **Steps** | 1. GET `/health`<br>2. Verify HTTP 200 |
| **Expected result** | Response includes `{"status": "ok", "service": "metrics-service"}` |

### TC-030: Classifier — Ticket classification request
| Field | Value |
|-------|-------|
| **Priority** | Medium |
| **Preconditions** | ticket-classifier running |
| **Test data** | `{"title": "Login broken", "description": "Cannot log in with Google", "category": "bug"}` |
| **Steps** | 1. POST `/classify` with ticket data<br>2. Verify response |
| **Expected result** | Returns predicted category, confidence score, and suggested priority |

---

## 6. Database Integration Tests

### TC-031: Neo4j — Create and query ontology graph
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | Neo4j test instance available |
| **Steps** | 1. Create class nodes in Neo4j<br>2. Create relationships between nodes<br>3. Query graph with Cypher |
| **Expected result** | Nodes and relationships stored correctly. Cypher query returns connected subgraph matching expected structure |

### TC-032: PostgreSQL — Version store CRUD
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | PostgreSQL test instance available, schema migrated |
| **Steps** | 1. Insert version record<br>2. Query by ontology ID<br>3. Query by commit ID<br>4. Verify audit trail |
| **Expected result** | Version records stored with correct schema. Queries return expected data. Audit trail shows full history |

---

## 7. Security — Expansion of Existing Tests

### TC-033: BOLA — Cross-tenant resource access (10 concurrent requests)
| Field | Value |
|-------|-------|
| **Priority** | High |
| **Preconditions** | Multiple tenants configured |
| **Steps** | 1. Send 10 concurrent cross-tenant requests<br>2. Verify all return 403 |
| **Expected result** | All 10 requests return 403 with `BOLA_CROSS_TENANT`. No rate limiting false positives |
