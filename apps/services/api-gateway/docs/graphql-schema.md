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
| Create / update / delete ontology, classes, properties, individuals | **REST** | `openapi.json` — `/ontologies*` |
| SPARQL and CYPHER execution | **REST** | `openapi.json` — `/sparql`, `/cypher` |
| Ontology metadata (branch, commit, dirty) | **REST** | `openapi.json` — `GET /ontologies/{id}` |
| Versioning reads (commits, branches) and writes (commit, branch, switch, merge, rollback, checkout) | **REST** | `openapi.json` — `/versioning/*` |
| Organization reads and writes (groups, projects, members, visibility, policies) | **REST** | `openapi.json` — `/groups`, `/projects`, `/projects/{id}/members`, etc. |
| Member management (add / update role / remove) | **REST** | `openapi.json` — `/projects/{id}/members*` |
| Draft state coordination | **REST** | `openapi.json` — `/ontologies/{id}/draft` (planned, see ADR) |

## 2. Query root

All field names below are camelCased at the wire level (async-graphql converts Rust snake_case to GraphQL camelCase by default).

### 2.0 Entity interface

Every graph-navigation entity (`Class`, `Property`, `Individual`) implements the `Entity` interface.

```graphql
interface Entity {
  id: String!
  label: String!
  comment: String
  entityType: EntityType!
}

enum EntityType {
  CLASS
  PROPERTY
  INDIVIDUAL
}
```

### 2.1 Classes

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
type Class implements Entity {
  id: String!
  label: String!
  comment: String
  entityType: EntityType!
  parents: [String!]!
  children: [String!]!
  isAbstract: Boolean!
  isDeprecated: Boolean!
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

### 2.2 Properties

| Field | Args | Returns |
|---|---|---|
| `property` | `ontologyId: String!`, `propertyId: String!` | `Property` (nullable) |
| `properties` | `ontologyId: String!`, `q: String`, `propertyType: PropertyType`, `page: Int = 0`, `perPage: Int = 20` | `PropertyConnection!` |

```graphql
enum PropertyType {
  OBJECT
  DATATYPE
  ANNOTATION
}

type Property implements Entity {
  id: String!
  label: String!
  comment: String
  entityType: EntityType!
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

### 2.3 Individuals

| Field | Args | Returns |
|---|---|---|
| `individual` | `ontologyId: String!`, `individualId: String!` | `Individual` (nullable — full detail with property values) |
| `individuals` | `ontologyId: String!`, `classId: String!`, `q: String`, `page: Int = 0`, `perPage: Int = 20` | `IndividualConnection!` |

```graphql
type Individual implements Entity {
  id: String!
  label: String!
  comment: String
  entityType: EntityType!
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

## 3. Forbidden operations (architecture invariant)

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
| `commits`, `branch`, `branches`, `tags`, `compareRevisions` | Non-graph Query resolvers removed — versioning reads migrated to REST | `GET /api/v1/versioning/commits`, `/api/v1/versioning/branches`, `/api/v1/versioning/commits/{id}/delta` |
| `groups`, `projects`, `members` | Non-graph Query resolvers removed — org reads migrated to REST | `GET /api/v1/groups`, `/api/v1/projects`, `/api/v1/projects/{id}/members` |
| `ontology(id)` metadata query | Non-graph Query resolver removed — ontology metadata migrated to REST | `GET /api/v1/ontologies/{id}` |
| Dynamic schema registration of user-defined class/property fields | User-defined ontology types are returned through containers (`literalValues`, `referenceValues`, `properties`), not as generated schema fields | N/A — use container fields |

## 4. DoS protection

GraphQL itself is a DoS vector (recursive queries, expensive resolvers). The ontology-service applies the following SDL/runtime safeguards (see ADR-DES.API.graphql-sparql-split-strategy § "Решение"):

- **Depth limit** on incoming queries (configured in `build_schema`).
- **Query complexity** calculator with a per-request threshold.
- **Timeout** per resolver (axum middleware).
- **Result cap** for `classTree` / `classDescendants` / `autocompleteClasses` (capped at 100).
- **Rate limiting** applied at the API Gateway layer (Redis-backed token bucket, configured in `middleware.Idempotency` and the rate-limit middleware).

Each GraphQL request is logged with `operation_name`, `trace_id`, redacted variables summary, result count, and the security decision (allowed / blocked / rate-limited).

## 5. Related documents

- `openapi.json` — REST contract for all write operations and SPARQL/CYPHER execution.
- `specs/adr/ADR-DES.API.graphql-sparql-split-strategy.md` — split rationale and protocol responsibility table.
- `specs/adr/ADR-DES.API.rest-graphql-mutation-boundary.md` — mutation boundary decision and migration checklist.
- `specs/adr/ADR-DES.API.sparql-dos-protection.md` — CircuitBreakerMiddleware integration with the REST `/sparql` endpoint.
- `specs/adr/ADR-DES.API.protocol-stack-strategy.md` — overall API protocol stack decision (GraphQL = navigation read-only, REST = writes).
