# Neo4j Cypher Reference

> Source: https://neo4j.com/docs/cypher-manual/current/introduction/cypher-overview/
> Source fallback: https://github.com/neo4j/docs-cypher/blob/dev/modules/ROOT/pages/introduction/cypher-overview.adoc
> Additional sources: https://github.com/neo4j/docs-cypher/tree/dev/modules/ROOT/pages/queries, https://github.com/neo4j/docs-cypher/tree/dev/modules/ROOT/pages/syntax, https://github.com/neo4j/docs-cypher/tree/dev/modules/ROOT/pages/patterns, https://github.com/neo4j/docs-cypher/tree/dev/modules/ROOT/pages/clauses, https://github.com/neo4j/docs-cypher/tree/dev/modules/cheat-sheet/pages
> Created: 2026-07-13
> Updated: 2026-07-13

## Overview

Cypher is Neo4j's declarative graph query language. It was created by Neo4j engineers as an SQL-equivalent language for graph databases. Like SQL, Cypher focuses on what to retrieve rather than how to retrieve it, but it expresses graph traversal with visual pattern syntax such as `(nodes)-[:CONNECT_TO]->(otherNodes)`.

Cypher queries are built around property graph entities: nodes, relationships, and paths. Queries visually describe graph patterns to match or create. Parentheses represent nodes, square brackets represent relationships, arrows represent relationship direction, labels and relationship types constrain matching, and properties appear in map-like `{key: value}` structures.

Compared with SQL, Cypher is schema-flexible and places `RETURN` near the end of the query. Neo4j can enforce open schema using graph types, constraints, and indexes, but nodes and relationships are not required to share the same property set unless constraints enforce it. Cypher can be extended with Neo4j's APOC Core library, which adds procedures and functions for data integration, graph algorithms, and data conversion.

## Core Concepts

**Node**: A graph data entity. In Cypher, nodes use parentheses: `(n)`. Labels act like tags for matching specific node categories: `(n:Person)`. Properties are key-value data inside braces: `(n:Person {name: 'Anna'})`. A variable such as `n` lets later clauses refer to the same node.

```cypher
MATCH (n:Person {name:'Anna'})
RETURN n.born AS birthYear
```

**Relationship**: A directed connection between two nodes. A relationship must have a start node, an end node, and exactly one type. Relationships use square brackets and arrows: `-[r:KNOWS]->`. Relationship properties can be filtered inline or with `WHERE`.

```cypher
MATCH (:Person {name: 'Anna'})-[r:KNOWS WHERE r.since < 2020]->(friend:Person)
RETURN count(r) As numberOfFriends
```

**Path**: A sequence of connected nodes and relationships. Paths can be matched and bound to variables. Variable-length traversal uses quantified relationships or quantified path patterns.

```cypher
MATCH p = SHORTEST 1 (:Person {name: 'Anna'})-[:KNOWS]-+(:Person {nationality: 'Canadian'})
RETURN p
```

**Pattern**: A node/relationship shape used by `MATCH`, `CREATE`, or `MERGE`. Writing a query is effectively drawing a pattern through the graph.

```cypher
MATCH (actor:Actor)-[:ACTED_IN]->(movie:Movie {title: 'The Matrix'})
RETURN actor.name
```

**Variable**: A symbolic name bound to a node, relationship, path, or expression result. Variables are passed through query clauses; `WITH` controls which variables stay in scope.

**Label**: A tag assigned to nodes, such as `Person` or `Movie`. A node may have multiple labels.

**Relationship type**: The single type assigned to a relationship, such as `ACTED_IN` or `KNOWS`.

**Property**: A key-value value stored on a node or relationship. Properties are accessed with dot syntax such as `movie.title` or dynamically with subscript syntax such as `node[propertyName]`.

## API / Interface

Cypher is a query language interface rather than a library API. The main interface is a text query submitted to Neo4j through Neo4j Browser, Cypher Shell, drivers, HTTP APIs, or embedded tooling.

### Common query shape

```cypher
MATCH <pattern>
WHERE <predicate>
WITH <expressions>
RETURN <expressions>
ORDER BY <expression>
SKIP <number>
LIMIT <number>
```

### Read clauses

| Clause | Purpose |
|--------|---------|
| `MATCH` | Specify graph patterns to search for. |
| `OPTIONAL MATCH` | Search for patterns while returning `null` for missing parts. |
| `FILTER` | Adds filters to queries; Cypher 25 only, introduced in Neo4j 2025.06. |

### Projecting clauses

| Clause | Purpose |
|--------|---------|
| `RETURN ... [AS]` | Defines what to include in the result set. |
| `WITH ... [AS]` | Chains query parts and passes selected variables/results forward. |
| `UNWIND ... [AS]` | Expands a list into rows. |
| `FOR` | GQL-aligned syntax with the same semantics as `UNWIND`; Cypher 25 only, introduced in Neo4j 2026.04. |
| `LET` | Binds values to variables; Cypher 25 only, introduced in Neo4j 2025.06. |
| `FINISH` | Defines a query to have no result. |

### Reading sub-clauses

| Clause | Purpose |
|--------|---------|
| `WHERE` | Adds constraints to `MATCH` / `OPTIONAL MATCH` patterns or filters `WITH` results. |
| `SEARCH` | Filters `MATCH` / `OPTIONAL MATCH` results using approximate nearest neighbor vector search. |
| `ORDER BY [ASC[ENDING] | DESC[ENDING]]` | Sorts output; ascending is default. |
| `SKIP` / `OFFSET` | Defines from which row to start including rows. |
| `LIMIT` | Constrains the number of output rows. |

### Write clauses

| Clause | Purpose |
|--------|---------|
| `CREATE` | Create nodes and relationships. |
| `INSERT` | Synonym to `CREATE`; requires multiple labels separated by `&`, not `:`. |
| `DELETE` | Delete nodes, relationships, or paths; nodes must have associated relationships explicitly deleted. |
| `DETACH DELETE` | Delete nodes and automatically delete all associated relationships. |
| `SET` | Update labels and properties on nodes and relationships. |
| `REMOVE` | Remove properties and labels from nodes and relationships. |
| `FOREACH` | Update data within a list, path components, or aggregation results. |

### Read/write, subquery, and set operation clauses

| Clause | Purpose |
|--------|---------|
| `MERGE` | Ensures a pattern exists; matches it or creates it. |
| `ON CREATE` | Used with `MERGE`; actions when a pattern is created. |
| `ON MATCH` | Used with `MERGE`; actions when an existing pattern is matched. |
| `CALL ... [YIELD ...]` | Invokes a database procedure and returns results. |
| `CALL { ... }` | Evaluates a subquery, often for post-`UNION` processing or aggregations. |
| `CALL { ... } IN TRANSACTIONS` | Evaluates a subquery in separate transactions, typically for large modifications/imports. |
| `UNION` | Combines multiple query result sets and removes duplicates. |
| `UNION ALL` | Combines multiple query result sets and retains duplicates. |
| `USE` | Selects which graph a query or query part executes against. |
| `LOAD CSV` | Imports data from CSV files. |

### Schema and administration clauses

| Clause family | Purpose |
|---------------|---------|
| `CREATE | SHOW | DROP INDEX` | Manage indexes. |
| `CREATE | SHOW | DROP CONSTRAINT` | Manage constraints. |
| `CREATE | ALTER | SHOW | RENAME | DROP GRAPH TYPE` | Manage graph types. |
| `SHOW` | List information about databases and configuration, including aliases, constraints, indexes, privileges, procedures, roles, servers, settings, transactions, users, and more. |

## Usage Patterns

### Match nodes by label and property

```cypher
MATCH (movie:Movie)
WHERE movie.rating > 7
RETURN movie.title
```

### Match a relationship pattern

```cypher
MATCH (actor:Actor)-[:ACTED_IN]->(movie:Movie {title: 'The Matrix'})
RETURN actor.name
```

### Match all nodes

```cypher
MATCH ()
RETURN count(*) AS numNodes
```

### Filter by label in a node pattern

```cypher
MATCH (:Stop)
RETURN count(*) AS numStops
```

### Match a fixed-length path

```cypher
MATCH (s:Stop)-[:CALLS_AT]->(:Station {name: 'Denmark Hill'})
RETURN s.arrives AS arrivalTime
```

### Use inline `WHERE` inside a pattern

```cypher
MATCH (n:Station {name: 'Denmark Hill'})<-[:CALLS_AT]-
        (s:Stop WHERE s.departs = time('22:37'))-[:NEXT]->
        (:Stop)-[:CALLS_AT]->(d:Station)
RETURN d.name AS nextCallingPoint
```

### Match variable-length paths with quantified relationships

```cypher
MATCH (:Station {name: 'Peckham Rye'})-[link:LINK]-+
        (:Station {name: 'Clapham Junction'})
RETURN reduce(acc = 0.0, l IN link | round(acc + l.distance, 2)) AS
         totalDistance
```

### Match variable-length paths with quantified path patterns

```cypher
MATCH (:Station {name: 'Peckham Rye'})
      (()-[link:LINK]-(s) WHERE link.distance <= 2)+
      (:Station {name: 'London Victoria'})
UNWIND s AS station
RETURN station.name AS callingPoint
```

### Use `EXISTS` subquery in a path filter

```cypher
MATCH (:Station {name: 'Denmark Hill'})<-[:CALLS_AT]-(s1:Stop)-[:NEXT]->+
        (sN:Stop WHERE NOT EXISTS { (sN)-[:NEXT]->(:Stop) })-[:CALLS_AT]->
        (d:Station)
RETURN s1.departs AS departure, sN.arrives AS arrival,
       d.name AS finalDestination
```

### Create nodes and relationships

```cypher
CREATE (keanu:Person {name: 'Keanu Reeves', born: 1964})
CREATE (matrix:Movie {title: 'The Matrix', released: 1999})
CREATE (keanu)-[:ACTED_IN {roles: ['Neo']}]->(matrix)
```

### Ensure data exists with `MERGE`

```cypher
MERGE (matrix:Movie {title: 'The Matrix', rating: 10})
MERGE (keanu:Person {name: 'Keanu Reeves'})
MERGE (keanu)-[:ACTED_IN]->(matrix)
```

### Use `MERGE` with conditional updates

```cypher
MERGE (p:Person {name: 'Keanu Reeves'})
ON CREATE SET p.createdAt = timestamp()
ON MATCH SET p.lastSeenAt = timestamp()
RETURN p
```

### Chain query parts with `WITH`

```cypher
MATCH (p:Person)-[:ACTED_IN]->(m:Movie)
WITH p, count(m) AS movieCount
WHERE movieCount > 5
RETURN p.name, movieCount
ORDER BY movieCount DESC
```

### Preserve all variables with `WITH *`

```cypher
MATCH (p:Person)-[:ACTED_IN]->(m:Movie)
WITH *
RETURN p.name, m.title
```

### Expand a list with `UNWIND`

```cypher
UNWIND ['Neo', 'Trinity', 'Morpheus'] AS name
CREATE (:Character {name: name})
```

### Use `RETURN DISTINCT`

```cypher
MATCH (p:Person)-[:ACTED_IN]->(:Movie)
RETURN DISTINCT p.name AS actor
```

### Use `RETURN *`

```cypher
MATCH p = (actor:Actor)-[:ACTED_IN]->(movie:Movie)
RETURN *
```

### Update properties with `SET`

```cypher
MATCH (p:Person {name: 'Keanu Reeves'})
SET p.popularity = 10
RETURN p
```

### Update properties dynamically

```cypher
MATCH (n:Person {name: 'Anna'})
SET n[$propertyName] = $propertyValue
RETURN n
```

### Replace or mutate property maps with `SET`

```cypher
MATCH (n:Person {name: 'Anna'})
SET n = {name: 'Anna', born: 1980}
RETURN n
```

```cypher
MATCH (n:Person {name: 'Anna'})
SET n += {city: 'London'}
RETURN n
```

### Delete a relationship, then a node

```cypher
MATCH (p:Person {name: 'Anna'})-[r]-()
DELETE r
```

```cypher
MATCH (p:Person {name: 'Anna'})
DELETE p
```

### Delete a node and its relationships

```cypher
MATCH (p:Person {name: 'Anna'})
DETACH DELETE p
```

### Use a `CALL` subquery per input row

```cypher
MATCH (team:Team)
CALL (team) {
  MATCH (team)<-[:PLAYS_FOR]-(player:Player)
  RETURN collect(player) AS players
}
RETURN team.name, players
```

### Combine query results

```cypher
MATCH (p:Person)
RETURN p.name AS name
UNION
MATCH (m:Movie)
RETURN m.title AS name
```

## Configuration

No Cypher language configuration file is required. The practical configuration surface depends on Neo4j server settings, driver settings, indexes, constraints, graph types, and procedure/plugin availability such as APOC Core.

| Area | Notes |
|------|-------|
| Schema | Neo4j can enforce open schema with graph types, constraints, and indexes. |
| APOC Core | Neo4j supports APOC Core procedures and functions that extend Cypher. |
| Parameters | Queries may use parameters for safer, reusable query execution. |
| Multiple graphs | `USE` selects the graph for a query or query part. |
| Imports | `LOAD CSV` imports CSV data; `CALL { ... } IN TRANSACTIONS` helps with large imports. |

## Best Practices

1. Model graph data around the questions the application needs to answer. The official basic query guide emphasizes creating an appropriate data model before loading data.
2. Use labels and relationship types to constrain patterns early, for example `(p:Person)` and `[:ACTED_IN]`.
3. Use parameters for application-supplied values rather than constructing query strings manually.
4. Use `WITH` deliberately to control query scope. Variables not referenced by `WITH` or carried by `WITH *` are dropped from scope.
5. Use `DISTINCT` when the query must remove duplicate rows or nodes.
6. Prefer `MERGE` when a pattern must exist and should not be blindly duplicated by `CREATE`.
7. Add `ON CREATE` and `ON MATCH` with `MERGE` when created and matched cases need different updates.
8. Use `CALL { ... } IN TRANSACTIONS` for large modifying/import workloads to avoid excessive memory pressure.
9. Use `ORDER BY`, `SKIP`/`OFFSET`, and `LIMIT` for deterministic pagination-like output.
10. Use indexes and constraints where appropriate; the manual lists dedicated schema clauses for managing them.

## Common Pitfalls

1. **Deleting nodes with relationships using `DELETE` only**: `DELETE` requires all associated relationships to be explicitly deleted first. Use `DETACH DELETE` when the node and all connected relationships should be removed.
2. **Using `DETACH DELETE` for large deletes**: The cheat sheet notes `DETACH DELETE` is useful for small example datasets but not suitable for deleting large amounts of data; use batched transactional patterns instead.
3. **Expecting `DETACH DELETE` to delete schema**: It removes data, not indexes or schema. To remove all data including indexes and constraints, recreate the database with `CREATE OR REPLACE DATABASE [name]`.
4. **Assuming relationship types behave like node labels**: Nodes may have multiple labels, but relationships have exactly one type.
5. **Dropping variables accidentally with `WITH`**: Any variable not explicitly referenced by `WITH` or included via `WITH *` is no longer in scope.
6. **Using `CREATE` when uniqueness is intended**: `CREATE` creates new data; use `MERGE` to match-or-create a pattern.
7. **Confusing `SET n = map` and `SET n += map`**: `SET n = map` replaces all properties; `SET n += map` mutates by updating/adding properties.
8. **Ignoring `null` behavior in optional matches**: `OPTIONAL MATCH` uses `nulls` for missing pattern parts; downstream predicates and expressions need to handle that.
9. **Assuming query output order without `ORDER BY`**: Use `ORDER BY` when order matters.
10. **Using dynamic labels/types without valid values**: Dynamic label/type expressions must evaluate to `STRING NOT NULL | LIST<STRING NOT NULL> NOT NULL`.

## Version Notes

- `FILTER` is marked Cypher 25 only and introduced in Neo4j 2025.06.
- `LET` is marked Cypher 25 only and introduced in Neo4j 2025.06.
- `FOR` is marked Cypher 25 only and introduced in Neo4j 2026.04; it is GQL-aligned syntax with the same semantics as `UNWIND`.
- `INSERT` is documented as a synonym to `CREATE`; unlike `CREATE`, it requires multiple labels to be separated by `&`, not `:`.
- `CALL` subqueries can import variables through a variable scope clause, `CALL (<variable>)`; importing `WITH` is noted as deprecated in the cheat-sheet source.
- Cypher contains two list concatenation operators: `||` and `+`; they are functionally equivalent, but `||` is GQL conformant and `+` is not.

## Critical Links

- Cypher overview: https://neo4j.com/docs/cypher-manual/current/introduction/cypher-overview/
- Clauses: https://neo4j.com/docs/cypher-manual/current/clauses/
- Syntax: https://neo4j.com/docs/cypher-manual/current/syntax/
- Patterns: https://neo4j.com/docs/cypher-manual/current/patterns/
- Queries: https://neo4j.com/docs/cypher-manual/current/queries/
- Subqueries: https://neo4j.com/docs/cypher-manual/current/subqueries/
- APOC Core: https://neo4j.com/docs/apoc/current/
- Official source repository used as fallback: https://github.com/neo4j/docs-cypher
