# SPARQL Query Language for RDF Reference

> Source: https://www.w3.org/TR/rdf-sparql-query/
> Created: 2026-07-13
> Updated: 2026-07-13

## Overview

SPARQL (SPARQL Protocol and RDF Query Language) is the W3C-standardized query language for RDF. RDF is a directed, labeled graph data format; SPARQL expresses queries over RDF graphs stored natively or materialized via middleware. The 2008 W3C Recommendation defines the 1.0 language: pattern matching over triples, optional/alternative patterns, dataset scoping (named graphs), solution modifiers (ORDER BY, DISTINCT, LIMIT, OFFSET), four query forms (SELECT, CONSTRUCT, ASK, DESCRIBE), and an extensible value-testing framework based on XQuery/XPath operators.

SPARQL 1.1 (superseding this spec) adds aggregate functions, sub-queries, property paths, federation, and an update language. This reference covers the 1.0 semantics, which remains the foundation.

## Core Concepts

- **RDF Term**: IRI, literal, or blank node. The set of all RDF Terms = I ∪ RDF-L ∪ RDF-B.
- **Triple Pattern**: Like an RDF triple but any of subject/predicate/object may be a variable. Formally: (RDF-T ∪ V) × (I ∪ V) × (RDF-T ∪ V).
- **Basic Graph Pattern (BGP)**: A set of triple patterns. Matches a subgraph of the data by substituting variables and blank nodes.
- **Group Graph Pattern**: Delimited by `{}`. Contains BGP(s), filters, OPTIONAL, UNION, and GRAPH patterns. The outer-most group is the query pattern.
- **Solution Mapping**: A partial function from variables to RDF terms. A "solution" binds variables so the pattern matches the data.
- **RDF Dataset**: `{ G, (u1, G1), ..., (un, Gn) }` — one default graph `G` (unnamed) plus zero or more named graphs `(ui, Gi)` identified by IRI. The **active graph** is the one currently being matched; outside `GRAPH`, it is the default graph.
- **Variables**: Prefixed by `?` or `$` (interchangeable). Global scope across the query. `$abc` and `?abc` are the same variable.
- **Blank Nodes in Queries**: Act as non-distinguished variables, not references to specific data blank nodes. Labels are scoped to a single BGP; the same label cannot appear in two different BGPs.
- **IRIs**: Written as `<...>` or as prefixed names (`prefix:local`). `PREFIX` declares prefix→IRI bindings; `BASE` sets the base IRI for resolving relative IRIs.
- **Literals**: String with optional language tag (`@lang`) or datatype IRI (`^^xsd:type`). Integers → `xsd:integer`, decimals → `xsd:decimal`, exponents → `xsd:double`, `true`/`false` → `xsd:boolean`.
- **Effective Boolean Value (EBV)**: Used by `FILTER`, `&&`, `||`, `!`. Boolean/numeric literals use their value; strings are false if zero-length; unbound or unsupported types produce a type error.
- **Filter Evaluation with Errors**: `||` returns TRUE if one branch is TRUE (even if the other errors); `&&` returns FALSE if one branch is FALSE; errors otherwise propagate. Three-valued logic (T/F/Error).

## API / Interface (Query Forms)

### Query Structure

```
Prologue? (SelectQuery | ConstructQuery | DescribeQuery | AskQuery)
Prologue ::= BaseDecl? PrefixDecl*
```

Keywords are case-insensitive except `a` (rdf:type shorthand).

### SELECT

Returns variable bindings as a result set.

```
SELECT [DISTINCT|REDUCED] (Var+ | '*') DatasetClause* WhereClause SolutionModifier
```

- `SELECT *` — all variables in the query.
- `DISTINCT` — eliminates duplicate solutions.
- `REDUCED` — permits (but does not require) eliminating duplicates.

### CONSTRUCT

Returns an RDF graph by substituting solutions into a triple template.

```
CONSTRUCT ConstructTemplate DatasetClause* WhereClause SolutionModifier
```

- Triples with unbound variables or illegal positions (e.g., literal as subject) are omitted.
- Ground triples (no variables) in the template always appear in output.
- Blank node labels in templates are scoped per solution; different solutions produce different blank nodes.

### ASK

Returns a boolean — whether the pattern has any solution.

```
ASK DatasetClause* WhereClause
```

### DESCRIBE

Returns an RDF graph describing resources. The description is determined by the query service, not the query. Can take IRIs, variables, or `*`.

```
DESCRIBE (VarOrIRIref+ | '*') DatasetClause* WhereClause? SolutionModifier
```

### Dataset Clauses

```
FROM <iri>          — adds graph to default graph (merge of all FROM clauses)
FROM NAMED <iri>    — adds a named graph
GRAPH (Var|IRI)     — switches active graph to a named graph (by IRI or variable)
```

### Solution Modifiers

Applied in this order: ORDER BY → Projection → DISTINCT/REDUCED → OFFSET → LIMIT.

```
SolutionModifier ::= OrderClause? LimitOffsetClauses?
OrderClause      ::= 'ORDER' 'BY' OrderCondition+
OrderCondition   ::= (('ASC'|'DESC') BrackettedExpression) | Constraint | Var
LIMIT INTEGER
OFFSET INTEGER
```

- **ORDER BY**: Sorting order — unbound (lowest) < blank nodes < IRIs < literals. Plain literals < `xsd:string`. Total order is not defined for all term pairs.
- **OFFSET**: Skips first N solutions (useless without ORDER BY).
- **LIMIT**: Upper bound on returned solutions. `LIMIT 0` returns nothing. Must be non-negative.

## Built-in Functions and Operators

### Grammar Access to Built-ins

```
BuiltInCall ::= 'STR' '(' E ')'
             | 'LANG' '(' E ')'
             | 'LANGMATCHES' '(' E ',' E ')'
             | 'DATATYPE' '(' E ')'
             | 'BOUND' '(' Var ')'
             | 'sameTerm' '(' E ',' E ')'
             | 'isIRI' '(' E ')' | 'isURI' '(' E ')'
             | 'isBLANK' '(' E ')'
             | 'isLITERAL' '(' E ')'
             | REGEX '(' E ',' E (',' E)? ')'
```

### Function Summary

| Function | Signature | Returns | Notes |
|----------|-----------|---------|-------|
| `bound` | `xsd:boolean bound(variable)` | true if var is bound | Used for negation-as-failure with OPTIONAL. Does not error on unbound. |
| `isIRI` / `isURI` | `xsd:boolean isIRI(RDF term)` | true if IRI | Aliases. |
| `isBlank` | `xsd:boolean isBlank(RDF term)` | true if blank node | |
| `isLiteral` | `xsd:boolean isLiteral(RDF term)` | true if literal | |
| `str` | `simple literal str(literal)` / `simple literal str(IRI)` | lexical form of literal or codepoint form of IRI | Useful with `regex` on typed literals. |
| `lang` | `simple literal lang(literal)` | language tag or `""` | |
| `datatype` | `IRI datatype(typed literal)` / `IRI datatype(simple literal)` | datatype IRI, or `xsd:string` for plain literals | |
| `sameTerm` | `xsd:boolean sameTerm(RDF term, RDF term)` | true if same RDF term (lexical+datatype IRI+language tag) | Does NOT error on unsupported datatypes, unlike `=`. |
| `langMatches` | `xsd:boolean langMatches(simple literal tag, simple literal range)` | true if tag matches range per RFC 4647 §3.3.1 | `"*"` matches any non-empty tag. |
| `regex` | `xsd:boolean regex(simple literal text, simple literal pattern [, simple literal flags])` | XPath `fn:matches` result | Flags: `"i"` (case-insensitive), `"m"`, `"x"`, `"s"`. Operates on simple literals (use `str()` on others). |

### Operator Summary

| Operator | Operand Types | Result | Notes |
|----------|--------------|--------|-------|
| `A \|\| B`, `A && B` | `xsd:boolean (EBV)` | `xsd:boolean` | Three-valued logic with errors. |
| `!A` | `xsd:boolean (EBV)` | `xsd:boolean` | `fn:not(A)`. |
| `A = B`, `A != B` | numeric, simple literal, xsd:string, xsd:boolean, xsd:dateTime | `xsd:boolean` | For unsupported datatypes, `=` may error (use `sameTerm`). |
| `A = B`, `A != B` | RDF term | `xsd:boolean` | `RDFterm-equal` — tests RDF term equality; errors if both literals with unequal values but unsupported datatype. |
| `A < B`, `A > B`, `A <= B`, `A >= B` | numeric, simple literal, xsd:string, xsd:boolean, xsd:dateTime | `xsd:boolean` | |
| `A + B`, `A - B`, `A * B`, `A / B` | numeric | numeric | `A / B` returns `xsd:decimal` if both operands `xsd:integer`. |
| `+A`, `-A` | numeric | numeric | Unary. |

### Constructor Functions (XPath casts)

`xsd:datetime(?lit)`, `xsd:integer(?x)`, etc. — IRI-named function calls that cast a value to an XML Schema datatype. Casting rules: `Y` (always valid), `N` (never), `M` (depends on lexical value). Casting IRI → `xsd:string` produces the codepoint string.

## Usage Patterns

### Simple SELECT query

```sparql
PREFIX foaf: <http://xmlns.com/foaf/0.1/>
SELECT ?name ?mbox
WHERE { ?x foaf:name ?name .
        ?x foaf:mbox ?mbox }
```

### Predicate-object lists and object lists (syntactic sugar)

```sparql
# Semicolon: shared subject
?x foaf:name ?name ;
   foaf:mbox ?mbox .

# Comma: shared subject + predicate
?x foaf:nick "Alice" , "Alice_" .

# Combined
?x foaf:name ?name ; foaf:nick "Alice" , "Alice_" .
```

### `a` keyword (rdf:type)

```sparql
?x a :Class1 .
# equivalent to: ?x rdf:type :Class1 .
```

### Blank node syntax

```sparql
[ :p "v" ] .                # fresh blank node, _:b0 :p "v" .
[ :p "v" ] :q "w" .         # _:b0 :p "v" .  _:b0 :q "w" .
:x :q [ :p "v" ] .          # :x :q _:b0 .  _:b0 :p "v" .
[ foaf:name ?name ; foaf:mbox <mailto:alice@example.org> ]
```

### RDF Collections

```sparql
(1 ?x 3 4) :p "w" .
# expands to: _:b0 rdf:first 1 ; rdf:rest _:b1 .
#              _:b1 rdf:first ?x ; rdf:rest _:b2 .
#              _:b2 rdf:first 3 ; rdf:rest _:b3 .
#              _:b3 rdf:first 4 ; rdf:rest rdf:nil .
#              _:b0 :p "w" .
()    # equivalent to rdf:nil
```

### OPTIONAL — left join

```sparql
SELECT ?name ?mbox
WHERE { ?x foaf:name ?name .
        OPTIONAL { ?x foaf:mbox ?mbox } }
# Solutions where mbox is absent still appear with mbox unbound.
```

### UNION — disjunction

```sparql
SELECT ?title
WHERE { { ?book dc10:title ?title } UNION { ?book dc11:title ?title } }
```

### Negation via OPTIONAL + !bound

```sparql
SELECT ?name
WHERE { ?x foaf:givenName ?name .
        OPTIONAL { ?x dc:date ?date }
        FILTER (!bound(?date)) }   # people with NO known date
```

### Named graphs

```sparql
SELECT ?src ?bobNick
FROM NAMED <http://example.org/foaf/aliceFoaf>
FROM NAMED <http://example.org/foaf/bobFoaf>
WHERE { GRAPH ?src { ?x foaf:mbox <mailto:bob@work.example> .
                     ?x foaf:nick ?bobNick } }
```

Restricting to a specific named graph:

```sparql
WHERE { GRAPH data:bobFoaf { ?x foaf:mbox <mailto:bob@work.example> ;
                             ?x foaf:nick ?nick } }
```

### FROM (default graph from URL)

```sparql
SELECT ?name
FROM <http://example.org/foaf/aliceFoaf>
WHERE { ?x foaf:name ?name }
```

### FILTER with regex

```sparql
SELECT ?title
WHERE { ?x dc:title ?title
        FILTER regex(?title, "^SPARQL") }
```

Case-insensitive:

```sparql
WHERE { ?x dc:title ?title
        FILTER regex(?title, "web", "i") }
```

### FILTER with numeric conditions

```sparql
SELECT ?title ?price
WHERE { ?x ns:price ?price .
        ?x dc:title ?title .
        FILTER (?price < 30.5) }
```

### FILTER with language tags

```sparql
SELECT ?title
WHERE { ?x dc:title "That Seventies Show"@en ;
        ?x dc:title ?title .
        FILTER langMatches(lang(?title), "FR") }
```

### CONSTRUCT — building a graph

```sparql
PREFIX foaf: <http://xmlns.com/foaf/0.1/>
PREFIX vcard: <http://www.w3.org/2001/vcard-rdf/3.0#>
CONSTRUCT { <http://example.org/person#Alice> vcard:FN ?name }
WHERE { ?x foaf:name ?name }
```

### CONSTRUCT with blank nodes in template

```sparql
CONSTRUCT { ?x  vcard:N _:v .
            _:v vcard:givenName ?gname .
            _:v vcard:familyName ?fname }
WHERE { { ?x foaf:firstname ?gname } UNION { ?x foaf:givenname ?gname } .
        { ?x foaf:surname   ?fname } UNION { ?x foaf:family_name ?fname } }
# Each solution gets its own fresh _:v blank node.
```

### ASK

```sparql
ASK { ?x foaf:name "Alice" }     # → true if any solution
```

### ORDER BY + LIMIT + OFFSET

```sparql
SELECT ?name
WHERE { ?x foaf:name ?name }
ORDER BY ?name
LIMIT 5
OFFSET 10
```

### DISTINCT vs REDUCED

```sparql
SELECT DISTINCT ?name WHERE { ?x foaf:name ?name }   # forces deduplication
SELECT REDUCED  ?name WHERE { ?x foaf:name ?name }   # may deduplicate
```

### Extension function

```sparql
PREFIX func: <http://example.org/functions#>
SELECT ?name ?id
WHERE { ?x foaf:name ?name ; func:empId ?id .
        FILTER (func:even(?id)) }
```

## Configuration

### Grammar (Rule Numbers)

| Rule | Definition |
|------|------------|
| `[1]` | `Query ::= Prologue (SelectQuery \| ConstructQuery \| DescribeQuery \| AskQuery)` |
| `[2]` | `Prologue ::= BaseDecl? PrefixDecl*` |
| `[3]` | `BaseDecl ::= 'BASE' IRI_REF` |
| `[4]` | `PrefixDecl ::= 'PREFIX' PNAME_NS IRI_REF` |
| `[5]` | `SelectQuery ::= 'SELECT' ('DISTINCT'\|'REDUCED')? (Var+\|'*') DatasetClause* WhereClause SolutionModifier` |
| `[6]` | `ConstructQuery ::= 'CONSTRUCT' ConstructTemplate DatasetClause* WhereClause SolutionModifier` |
| `[7]` | `DescribeQuery ::= 'DESCRIBE' (VarOrIRIref+\|'*') DatasetClause* WhereClause? SolutionModifier` |
| `[8]` | `AskQuery ::= 'ASK' DatasetClause* WhereClause` |
| `[9]` | `DatasetClause ::= 'FROM' (DefaultGraphClause \| NamedGraphClause)` |
| `[13]` | `WhereClause ::= 'WHERE'? GroupGraphPattern` |
| `[14]` | `SolutionModifier ::= OrderClause? LimitOffsetClauses?` |
| `[20]` | `GroupGraphPattern ::= '{' TriplesBlock? ((GraphPatternNotTriples\|Filter) '.'? TriplesBlock?)* '}'` |
| `[22]` | `GraphPatternNotTriples ::= OptionalGraphPattern \| GroupOrUnionGraphPattern \| GraphGraphPattern` |
| `[23]` | `OptionalGraphPattern ::= 'OPTIONAL' GroupGraphPattern` |
| `[24]` | `GraphGraphPattern ::= 'GRAPH' VarOrIRIref GroupGraphPattern` |
| `[25]` | `GroupOrUnionGraphPattern ::= GroupGraphPattern ('UNION' GroupGraphPattern)*` |
| `[26]` | `Filter ::= 'FILTER' Constraint` |
| `[57]` | `BuiltInCall ::= STR/LANG/LANGMATCHES/DATATYPE/BOUND/sameTerm/isIRI/isURI/isBLANK/isLITERAL/RegexExpression` |
| `[58]` | `RegexExpression ::= 'REGEX' '(' E ',' E (',' E)? ')'` |

### Literal Datatype Shortcuts

| Syntax | Datatype |
|--------|----------|
| `42` | `xsd:integer` |
| `1.3` | `xsd:decimal` |
| `1.0e6` | `xsd:double` |
| `true` / `false` | `xsd:boolean` |
| `"foo"` | plain literal (no datatype) |
| `"foo"@en` | plain literal with language tag |
| `"foo"^^xsd:string` | typed literal |

### Escape Sequences (Strings)

`\t` U+0009, `\n` U+000A, `\r` U+000D, `\b` U+0008, `\f` U+000C, `\"` U+0022, `\'` U+0027, `\\` U+005C. Codepoint escapes `\uXXXX` and `\UXXXXXXXX` can appear anywhere in the query string and are processed before grammar parsing.

### IRI Syntax

- `IRI_REF ::= '<' ([^<>"{}|^`\\]-[#x00-#x20])* '>'`
- Prefixed names: `PNAME_NS ::= PN_PREFIX? ':'`, `PNAME_LN ::= PNAME_NS PN_LOCAL`.
- SPARQL local names allow leading digits (XML local names do not).
- Relative IRIs are resolved against `BASE` per RFC 3986 §5.1.
- Comments: `#` to end of line, treated as whitespace.

## Best Practices

1. **Always use PREFIX declarations** instead of full IRIs for readability and to avoid typos that silently fail to match.
2. **Prefer `sameTerm` over `=`** when comparing typed literals with unsupported datatypes — `=` can produce a type error and silently filter solutions.
3. **Use `!bound(?v)` with OPTIONAL for negation** — SPARQL 1.0 has no `NOT EXISTS` or `MINUS` (added in 1.1).
4. **Pair LIMIT/OFFSET with ORDER BY** — without ORDER BY the slice is undefined.
5. **Keep blank node labels within one BGP** — the same label cannot appear in two BGPs in one query.
6. **Use `str()` before `regex()` on typed literals** — `regex()` operates only on simple literals.
7. **Use `*` carefully in SELECT** — it returns all variables, including those from inner OPTIONALs or UNIONs; this can produce wide result sets.
8. **Validate IRIs against RFC 3987** — `<abc##def>` is illegal; illegal IRIs may cause parser failures.
9. **Treat `REDUCED` as a hint**, not a guarantee — for strict deduplication always use `DISTINCT`.
10. **Use `GRAPH ?g` to discover which named graph contributed a solution** — bind `?g` and project it.
11. **Be careful with file: URIs in FROM** — servers may dereference them and expose local file contents.
12. **Security: similar-looking IRIs (Cyrillic vs Latin)** may spoof — verify character codepoints when constructing queries from user input.

## Common Pitfalls

- **`"cat"` does not match `"cat"@en`** — a plain literal and a language-tagged literal are distinct RDF terms. Match language-tagged literals explicitly: `{ ?v ?p "cat"@en }`.
- **`42` does not match `"42"`** — the integer is a typed literal `xsd:integer`; the string is a plain literal. Use `str()` and cast if needed.
- **`= ` between unsupported-datatype literals errors**, silently filtering the solution — use `sameTerm` for exact RDF-term comparison.
- **OPTIONAL is left-associative**: `P OPTIONAL { A } OPTIONAL { B }` means `(P OPTIONAL { A }) OPTIONAL { B }`. Multiple OPTIONALs each add bindings independently.
- **FILTER scope is the whole group** — placing it before or after triple patterns in the same group does not change behavior.
- **UNION with different variables leaves the other unbound** — `{ ?x :p ?a } UNION { ?y :p ?b }` yields solutions binding either `?a` or `?b`, not both.
- **Blank nodes in query results are not stable identifiers** — they are scoped to the result document and may differ across runs.
- **CONSTRUCT drops triples with unbound variables** — if a template variable is optional and unbound, the corresponding triple(s) are omitted; explicit `BOUND` checks may be needed.
- **`regex` on typed or language-tagged literals fails** — wrap with `str(?lit)`.
- **Two `FROM` clauses with the same IRI** do not guarantee identical blank node identity in the merged default graph.
- **Numeric/cat errors short-circuit**: `||` only requires one TRUE; ensure both branches are type-safe when semantics matter.
- **`OPTIONAL { { ... FILTER(... ?x ...) } }`** — the spec notes double-nested OPTIONAL+FILTER combinations have ambiguous simplification; results may differ between engines — restructure when possible.

## Version Notes

- This reference covers **SPARQL 1.0** (W3C Recommendation 15 January 2008).
- **SPARQL 1.1** (superseding) adds: aggregate functions (`COUNT`, `SUM`, `AVG`, `GROUP BY`, `HAVING`), sub-queries, property paths (`*`, `+`, `?`, sequence, inverse), `SERVICE` (federated query), `NOT EXISTS` / `MINUS`, `BIND` / `VALUES`, update language (`INSERT` / `DELETE`), COALESCE, IF, CONCAT, SUBSTR, and more. Engines implementing 1.1 are backward-compatible with 1.0 queries unless they use features removed in 1.1.
- The 1.0 `RDFterm-equal` errors on unsupported-datatype literals; 1.0 implementations return errors, leaving the solution-filtering behavior implementation-defined for some cases. Prefer `sameTerm`.
- Media type: `application/sparql-query`. Recommended file extension: `.rq`.
- Encoding: UTF-8.