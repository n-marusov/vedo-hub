# VEDO Core GraphQL Schema — Navigation Only

**Endpoint:** `POST /api/v1/graphql` (proxied to `ontology-service`).
**Auth:** Bearer JWT issued by Keycloak (same `Authorization` header as REST).
**Status:** This document reflects the **target end-state** of ADR-DES.API.graphql-sparql-split-strategy and ADR-DES.API.rest-graphql-mutation-boundary.

## 1. Boundary

GraphQL in VEDO Core is a **read-only navigation layer** for the 2D/3D ontology navigator. Every write operation goes through the REST API documented in `openapi.json`.

| Responsibility | Protocol | Layer |
|---|---|---|
| Graph navigation (class tree, neighborhood, autocomplete, breadcrumb, descendants) | GraphQL Query | this schema |
| Read individual class / property / individual | GraphQL Query | this schema |
| Read commits and branches (version history) | GraphQL Query | this schema (proxied to versioning-service) |
| Read org context (groups, projects, members) | GraphQL Query | this schema (proxied to auth-service) |
| Create / update / delete ontology, classes, properties, individuals | **REST** | `openapi.json` — `/ontologies*` |
| SPARQL and CYPHER execution | **REST** | `openapi.json` — `/sparql`, `/cypher` |
| Versioning writes (commit, branch, switch, merge, rollback, checkout) | **REST** | `openapi.json` — `/versioning/*` |
| Member management (add / update role / remove) | **REST** | `openapi.json` — `/ontologies/{id}/members*` |
| Draft state coordination | **REST** | `openapi.json` — `/ontologies/{id}/draft` (planned, see ADR) |

## 2. Query root

All field names below are camelCased at the wire level (async-graphql converts Rust snake_case to GraphQL camelCase by default).

### 2.1 Ontology

| Field | Args | Returns | Notes |
|---|---|---|---|
| `ontology` | `id: String!` | `Ontology!` | Returns ontology metadata, current branch name, HEAD commit SHA, and `dirty` flag. Falls back to branch `main` with empty commit when the versioning-service is unreachable. |

```graphql
type Ontology {
  id: String!
  name: String!
  branch: String!
  commit: String!
  dirty: Boolean!
}
```

### 2.2 Classes

| Field | Args | Returns |
|---|---|---|
| `class` | `ontologyId: String!`, `classId: String!` | `Class` (nullable) |
| `classes` | `ontologyId: String!`, `q: String`, `page: Int = 0`, `perPage: Int = 20` | `ClassConnection!` |
| `classTree` | `ontologyId: String!` | `[ClassTreeNode!]!` |
| `classAncestors` | `ontologyId: String!`, `classId: String!` | `[BreadcrumbItem!]!` |
| `classDescendants` | `ontologyId: String!`, `classId: String!`, `maxDepth: Int = 10` | `[ClassTreeNode!]!` |
| `graphNeighborhood` | `ontologyId: String!`, `classId: String!`, `depth: Int = 2` | `GraphNeighborhood!` |
| `autocompleteClasses` | `ontologyId: String!`, `q: String!`, `limit: Int = 20` | `[ClassSummary!]!` |

```graphql
type Class {
  id: String!
  label: String!
  comment: String
  parents: [String!]!
  children: [String!]!
}

type ClassSummary {
  id: String!
  label: String!
  comment: String
  parents: [String!]!
}

type ClassTreeNode {
  id: String!
  label: String!
  children: [ClassTreeNode!]!
}

type BreadcrumbItem {
  id: String!
  label: String!
}

type GraphNeighborhood {
  nodes: [GraphNode!]!
  edges: [GraphEdge!]!
}

type GraphNode {
  id: String!
  label: String!
}

type GraphEdge {
  sourceId: String!
  targetId: String!
  propertyId: String!
  propertyLabel: String!
}

type ClassConnection {
  items: [ClassSummary!]!
  total: Int!
  page: Int!
  perPage: Int!
}
```

### 2.3 Properties

| Field | Args | Returns |
|---|---|---|
| `property` | `ontologyId: String!`, `propertyId: String!` | `Property` (nullable) |
| `properties` | `ontologyId: String!`, `q: String`, `propertyType: PropertyType`, `page: Int = 0`, `perPage: Int = 20` | `PropertyConnection!` |

```graphql
enum PropertyType {
  OBJECT
  DATATYPE
}

type Property {
  id: String!
  label: String!
  comment: String
  propertyType: PropertyType!
  domains: [String!]!
  ranges: [String!]!
  xsdType: String
  characteristics: PropertyCharacteristics!
  annotations: [Annotation!]!
}

type PropertySummary {
  id: String!
  label: String!
  propertyType: PropertyType!
  xsdType: String
  domains: [String!]!
}

type PropertyConnection {
  items: [PropertySummary!]!
  total: Int!
  page: Int!
  perPage: Int!
}
```

### 2.4 Individuals

| Field | Args | Returns |
|---|---|---|
| `individual` | `ontologyId: String!`, `individualId: String!` | `Individual` (nullable — full detail with property values) |
| `individuals` | `ontologyId: String!`, `classId: String!`, `q: String`, `page: Int = 0`, `perPage: Int = 20` | `IndividualConnection!` |

```graphql
type Individual {
  id: String!
  label: String!
  comment: String
  classId: String!
  classLabel: String!
  literalValues: [LiteralValue!]!
  referenceValues: [ReferenceValue!]!
}

type LiteralValue {
  propertyId: String!
  propertyLabel: String!
  value: String!
  xsdType: String
  valueId: String
}

type ReferenceValue {
  propertyId: String!
  propertyLabel: String!
  targetId: String!
  targetLabel: String!
  edgeId: String!
}

type IndividualConnection {
  items: [IndividualSummary!]!
  total: Int!
  page: Int!
  perPage: Int!
}
```

User-defined class/property/individual fields are NOT exposed as static GraphQL fields. The dynamic ontology schema is returned through the `literalValues` / `referenceValues` / `properties` containers — see ADR-DES.API.graphql-sparql-split-strategy § "Динамическая регистрация пользовательских типов в GraphQL-схеме" (forbidden).

### 2.5 Versioning (read-only)

These resolvers proxy read calls to the versioning-service. Write operations on versions go through REST `/api/v1/versioning/*`.

| Field | Args | Returns |
|---|---|---|
| `commits` | `ontologyId: String!`, `branchId: String`, `page: Int = 0`, `perPage: Int = 20` | `CommitConnection!` |
| `branch` | `ontologyId: String!`, `branchId: String!` | `Branch` (nullable) |
| `branches` | `ontologyId: String!`, `referenceBranchId: String` | `BranchConnection!` |

```graphql
type Commit {
  id: String!
  branchId: String!
  parentCommitId: String
  message: String!
  authorId: String!
  authorName: String!
  totalChanges: Int!
  createdAt: String!
}

type CommitConnection {
  items: [Commit!]!
  total: Int!
  page: Int!
  perPage: Int!
}

type Branch {
  id: String!
  name: String!
  ontologyId: String!
  headCommitId: String
  createdAt: String!
  isProtected: Boolean!
  lastCommitMessage: String
  lastCommitAuthor: String
  aheadCount: Int
  behindCount: Int
}

type BranchConnection {
  items: [Branch!]!
  total: Int!
}
```

### 2.6 Organization (read-only)

| Field | Args | Returns |
|---|---|---|
| `groups` | `q: String` | `[Group!]!` |
| `projects` | `q: String`, `sortBy: String`, `sortDir: String`, `page: Int`, `perPage: Int` | `[Project!]!` |
| `members` | `ontologyId: String!` | `[Member!]!` |

```graphql
type Group {
  id: String!
  name: String!
  description: String
  parentGroupId: String
  visibility: String
  memberCount: Int
  projectCount: Int
}

type Project {
  id: String!
  name: String!
  description: String
  visibility: String
  memberCount: Int
  updatedAt: String
}

type Member {
  userId: String!
  scope: String!
  role: String!
  username: String
  avatarUrl: String
  addedAt: String
}
```

## 3. Mutation root — DEPRECATED

The target end-state of VEDO Core is **`EmptyMutation`** — no GraphQL `Mutation` type at all (see ADR-DES.API.rest-graphql-mutation-boundary.md § "Запрещено в GraphQL"). The current schema still exposes the following **deprecated placeholder** resolvers while the frontend migrates to REST. They emit a `#[deprecated]` warning at compile time of the ontology-service and will be removed once the frontend stops invoking them.

| Mutation (still present) | Replaces | Migration target (REST) |
|---|---|---|
| `updateDraft(ontologyId, changes: DraftInput!)` | frontend `useDraftState` composable | `PUT /api/v1/ontologies/{id}/draft` (planned — see ADR) |
| `updateMemberRole(ontologyId, userId, role)` | frontend `MembersPage` role editor | `PUT /api/v1/ontologies/{id}/members/{userId}` |
| `removeMember(ontologyId, userId)` | frontend `MembersPage` remove action | `DELETE /api/v1/ontologies/{id}/members/{userId}` |

Frontend Apollo client code that still calls these mutations should be migrated to `axios` (or another REST client) using the endpoints in `openapi.json`. After all call sites are migrated:

1. Remove `MutationRoot` from `build_schema` in `src/services/ontology-service/src/graphql/schema.rs` and replace with `EmptyMutation`.
2. Delete `src/services/ontology-service/src/graphql/mutation.rs`.
3. Remove `UPDATE_DRAFT_MUTATION`, `UPDATE_MEMBER_ROLE_MUTATION`, `REMOVE_MEMBER_MUTATION` from `src/services/frontend/src/apollo/queries.ts` and update `useDraftState` / `MembersPage` consumers.

## 4. Forbidden operations (architecture invariant)

The following GraphQL operations are strictly forbidden by ADR-DES.API.graphql-sparql-split-strategy and ADR-DES.API.rest-graphql-mutation-boundary. Any PR adding one of these is an architectural violation.

| Operation | Why forbidden | REST replacement |
|---|---|---|
| `sparqlQuery(query: String!)` | Bypasses API Gateway CircuitBreakerMiddleware, rate limiting, query complexity checks — see ADR-DES.API.sparql-dos-protection | `POST /api/v1/sparql` |
| `cypherQuery(query: String!)` | Same — bypasses DoS protection | `POST /api/v1/cypher` |
| `createClass`, `createProperty`, `createIndividual` | Write operations must route through REST for idempotency-key, auth, audit log | `POST /api/v1/ontologies/{id}/classes` (and `/properties`, `/individuals`) |
| `updateClass`, `updateProperty`, `updateIndividual` | Same | `PUT /api/v1/ontologies/{id}/classes/{classId}` (etc.) |
| `deleteClass`, `deleteProperty`, `deleteIndividual` | Same | `DELETE /api/v1/ontologies/{id}/classes/{classId}` (etc.) |
| `createOntology`, `updateOntology`, `deleteOntology` | Same | `POST/PUT/DELETE /api/v1/ontologies[/{id}]` |
| `createCommit`, `createBranch`, `mergeBranches`, `switchBranch`, `deleteBranch`, `checkoutCommit`, `rollbackToCommit` | Same — versioning writes go through REST | `/api/v1/versioning/*` (see `openapi.json`) |
| `addMember`, `importOntology`, `exportOntology` | Same | `/api/v1/ontologies/{id}/members`, `/import`, `/export` |
| Dynamic schema registration of user-defined class/property fields | User-defined ontology types are returned through containers (`literalValues`, `referenceValues`, `properties`), not as generated schema fields | N/A — use container fields |

## 5. DoS protection

GraphQL itself is a DoS vector (recursive queries, expensive resolvers). The ontology-service applies the following SDL/runtime safeguards (see ADR-DES.API.graphql-sparql-split-strategy § "Решение"):

- **Depth limit** on incoming queries (configured in `build_schema`).
- **Query complexity** calculator with a per-request threshold.
- **Timeout** per resolver (axum middleware).
- **Result cap** for `classTree` / `classDescendants` / `autocompleteClasses` (capped at 100).
- **Rate limiting** applied at the API Gateway layer (Redis-backed token bucket, configured in `middleware.Idempotency` and the rate-limit middleware).

Each GraphQL request is logged with `operation_name`, `trace_id`, redacted variables summary, result count, and the security decision (allowed / blocked / rate-limited).

## 6. Related documents

- `openapi.json` — REST contract for all write operations and SPARQL/CYPHER execution.
- `specs/adr/ADR-DES.API.graphql-sparql-split-strategy.md` — split rationale and protocol responsibility table.
- `specs/adr/ADR-DES.API.rest-graphql-mutation-boundary.md` — mutation boundary decision and migration checklist.
- `specs/adr/ADR-DES.API.sparql-dos-protection.md` — CircuitBreakerMiddleware integration with the REST `/sparql` endpoint.
- `specs/adr/ADR-DES.API.protocol-stack-strategy.md` — overall API protocol stack decision (GraphQL = navigation read-only, REST = writes).