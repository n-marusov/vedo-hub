# W3C Turtle (RDF 1.1 Turtle) Reference

> Source: https://www.w3.org/TR/turtle/
> Created: 2026-07-12
> Updated: 2026-07-12

## Overview

Turtle (Terse RDF Triple Language) is a textual syntax for RDF defined as a W3C Recommendation (25 February 2014). It allows an RDF graph to be written in a compact, natural text form with abbreviations for common usage patterns and datatypes. Turtle provides levels of compatibility with N-Triples format and the triple pattern syntax of SPARQL.

A Turtle document serializes an RDF graph as a set of triples (subject, predicate, object). The grammar is a subset of the SPARQL 1.1 Query Language `TriplesBlock` grammar, sharing production and terminal names where possible.

- **Media type:** `text/turtle`
- **File extension:** `.ttl`
- **Encoding:** always UTF-8
- **Identifier IRI:** `http://www.w3.org/ns/formats/Turtle`
- **Specification status:** W3C Recommendation (stable, normative)

## Core Concepts

**Triple:** The fundamental unit — a sequence of (subject, predicate, object) terms separated by whitespace and terminated by `.`.

**RDF Term types** (three defined in RDF Concepts):
- **IRIs** — Internationalized Resource Identifiers, written enclosed in `<` and `>`, or as prefixed names.
- **Literals** — values such as strings, numbers, booleans; composed of a lexical form and optional language tag or datatype IRI.
- **Blank nodes** — unnamed RDF nodes, written as `_:label` or using `[ ... ]` / `[]` syntax.

**Prefix declaration** (`@prefix` or `PREFIX`): associates a short label with a namespace IRI to avoid repeating long IRIs.

**Base declaration** (`@base` or `BASE`): sets the base IRI for resolving relative IRIs.

**`a` keyword:** in predicate position, represents the IRI `http://www.w3.org/1999/02/22-rdf-syntax-ns#type` (rdf:type).

**Collections:** RDF lists using `rdf:first`/`rdf:rest` structure, written as `( ... )`.

**Comments:** `#` outside an IRIREF or String, continuing to end of line. Treated as whitespace.

## Syntax Forms

### Simple Triples

```turtle
<http://example.org/#spiderman> <http://www.perceive.net/schemas/relationship/enemyOf> <http://example.org/#green-goblin> .
```

### Predicate Lists (`;`)

Repeats the subject across multiple predicate-object pairs:

```turtle
<http://example.org/#spiderman> <http://www.perceive.net/schemas/relationship/enemyOf> <http://example.org/#green-goblin> ;
    <http://xmlns.com/foaf/0.1/name> "Spiderman" .
```

### Object Lists (`,`)

Repeats subject and predicate across multiple objects:

```turtle
<http://example.org/#spiderman> <http://xmlns.com/foaf/0.1/name> "Spiderman", "Человек-паук"@ru .
```

### IRIs

- **Absolute IRI:** `<http://example.org/#green-goblin>`
- **Relative IRI:** `<#green-goblin>` — resolved relative to current base IRI
- **Prefixed name:** `foaf:name` — expands to `<http://xmlns.com/foaf/0.1/name>` when `@prefix foaf: <http://xmlns.com/foaf/0.1/>` is declared
- **Empty prefix:** `@prefix : <http://another.example/> .` allows `:subject5 :predicate5 :object5 .`

Prefix declarations:
```turtle
# Turtle style (requires trailing '.')
@prefix somePrefix: <http://www.perceive.net/schemas/relationship/> .

# SPARQL style (NO trailing '.')
PREFIX somePrefix: <http://www.perceive.net/schemas/relationship/>
```

**Note:** Prefixed names are a superset of XML QNames. The local part may include:
- Leading digits: `leg:3032571`
- Non-leading colons: `og:video:height`
- Reserved character escape sequences: `wgs:lat\-long`

### Literals

#### Quoted Literals

Delimiters: `"..."`, `'...'`, `"""..."""` (multi-line), `'''...'''` (multi-line)

- Language tag: `"Spiderman"@en`
- Datatype IRI: `"That Seventies Show"^^xsd:string`
- No language tag or datatype → datatype is `xsd:string`

Restrictions on `\` — may not appear except in escape sequences.

| Delimiter | Cannot contain |
|-----------|---------------|
| `"` | `"`, LF (U+000A), CR (U+000D) |
| `'` | `'`, LF, CR |
| `"""` | the sequence `"""` |
| `'''` | the sequence `'''` |

#### Numbers

| Datatype | Abbreviated | Lexical equivalent | Regex |
|----------|-------------|--------------------|-------|
| `xsd:integer` | `-5` | `"-5"^^xsd:integer` | `[+-]?[0-9]+` |
| `xsd:decimal` | `-5.0` | `"-5.0"^^xsd:decimal` | `[+-]?[0-9]*\.[0-9]+` |
| `xsd:double` | `4.2E9` | `"4.2E9"^^xsd:double` | mantissa: `[+-]?[0-9]+\.[0-9]+` or `[+-]?\.[0-9]+` or `[+-]?[0-9]`; exponent: `[+-]?[0-9]+` |

#### Booleans

`true` or `false` (case-sensitive) → datatype `xsd:boolean`.

```turtle
<http://somecountry.example/census2007> :isLandlocked false .
```

### Blank Nodes

Labeled blank nodes: `_:alice`, `_:bob` — a fresh blank node is allocated for each unique label; repeated use identifies the same node.

```turtle
_:alice foaf:knows _:bob .
_:bob foaf:knows _:alice .
```

### Nested Unlabeled Blank Nodes (`[ ... ]`)

`[]` allocates a fresh blank node. `[ predicateObjectList ]` allocates a fresh blank node that serves as the subject of the embedded triples.

```turtle
# Someone knows someone else, who has the name "Bob".
[] foaf:knows [ foaf:name "Bob" ] .
```

Nested:
```turtle
[ foaf:name "Alice" ] foaf:knows [
    foaf:name "Bob" ;
    foaf:knows [
        foaf:name "Eve" ] ;
    foaf:mbox <bob@example.com> ] .
```

### Collections (`( ... )`)

RDF Collections (linked list with `rdf:first`/`rdf:rest`).

```turtle
:subject :predicate ( :a :b :c ) .    # non-empty collection
:subject :predicate2 () .              # empty collection → rdf:nil
```

Collections can be nested and involve other syntactic forms:

```turtle
(1 [:p :q] ( 2 ) ) :p2 :q2 .
```

## Turtle vs SPARQL

| Feature | Turtle | SPARQL |
|---------|--------|--------|
| Literals as subject | No | Yes |
| Variables (`?name`, `$name`) | No | Yes |
| Prefix/base declarations | Anywhere outside triple | Prologue only |
| `@prefix`/`@base` case | Case-sensitive | N/A |
| `PREFIX`/`BASE` case | Case-insensitive | Case-insensitive |
| `true`/`false` case | Case-sensitive | Case-insensitive |

## Escape Sequences

### Numeric escape sequences (in IRIs and Strings)

| Sequence | Range |
|----------|-------|
| `\u` HEX HEX HEX HEX | U+0000 to U+FFFF |
| `\U` HEX HEX HEX HEX HEX HEX HEX HEX | U+0000 to U+10FFFF |

### String escape sequences (in Strings only)

| Sequence | Code point |
|----------|-----------|
| `\t` | U+0009 |
| `\b` | U+0008 |
| `\n` | U+000A |
| `\r` | U+000D |
| `\f` | U+000C |
| `\"` | U+0022 |
| `\'` | U+0027 |
| `\\` | U+005C |

### Reserved character escape sequences (in local names only)

`\` followed by one of: `~.-!$&'()*+,;=/?#@%_` — represents the character to the right of the `\`.

### Escape context table

| Context | Numeric escapes | String escapes | Reserved char escapes |
|---------|----------------|----------------|----------------------|
| IRIs (as RDF terms or in declarations) | yes | no | no |
| Local names | no | no | yes |
| Strings | yes | yes | no |

**Note on %-encoding:** `%` followed by two hex characters in IRIs and local names is NOT decoded during processing. `<http://a.example/%66oo-bar>` designates the IRI `http://a.example/%66oo-bar`, NOT `http://a.example/foo-bar`.

## Grammar

The entry point is `turtleDoc`. The grammar is LL(1) and LALR(1) when uppercased names are used as terminals. Tokenizing uses longest match.

### Key productions

| Rule | Definition |
|------|-----------|
| `turtleDoc` | `statement*` |
| `statement` | `directive \| triples '.'` |
| `directive` | `prefixID \| base \| sparqlPrefix \| sparqlBase` |
| `prefixID` | `'@prefix' PNAME_NS IRIREF '.'` |
| `base` | `'@base' IRIREF '.'` |
| `sparqlBase` | `"BASE" IRIREF` |
| `sparqlPrefix` | `"PREFIX" PNAME_NS IRIREF` |
| `triples` | `subject predicateObjectList \| blankNodePropertyList predicateObjectList?` |
| `predicateObjectList` | `verb objectList (';' (verb objectList)?)*` |
| `objectList` | `object (',' object)*` |
| `verb` | `predicate \| 'a'` |
| `subject` | `iri \| BlankNode \| collection` |
| `predicate` | `iri` |
| `object` | `iri \| BlankNode \| collection \| blankNodePropertyList \| literal` |
| `literal` | `RDFLiteral \| NumericLiteral \| BooleanLiteral` |
| `blankNodePropertyList` | `'[' predicateObjectList ']'` |
| `collection` | `'(' object* ')'` |
| `NumericLiteral` | `INTEGER \| DECIMAL \| DOUBLE` |
| `RDFLiteral` | `String (LANGTAG \| '^^' iri)?` |
| `BooleanLiteral` | `'true' \| 'false'` |
| `String` | `STRING_LITERAL_QUOTE \| STRING_LITERAL_SINGLE_QUOTE \| STRING_LITERAL_LONG_SINGLE_QUOTE \| STRING_LITERAL_LONG_QUOTE` |
| `iri` | `IRIREF \| PrefixedName` |
| `PrefixedName` | `PNAME_LN \| PNAME_NS` |
| `BlankNode` | `BLANK_NODE_LABEL \| ANON` |

### Key terminal productions

| Rule | Definition |
|------|-----------|
| `IRIREF` | `'<' ([^#x00-#x20<>"{}|^\`] \| UCHAR)* '>'` |
| `PNAME_NS` | `PN_PREFIX? ':'` |
| `PNAME_LN` | `PNAME_NS PN_LOCAL` |
| `BLANK_NODE_LABEL` | `'_:' (PN_CHARS_U \| [0-9]) ((PN_CHARS \| '.')* PN_CHARS)?` |
| `LANGTAG` | `'@' [a-zA-Z]+ ('-' [a-zA-Z0-9]+)*` |
| `INTEGER` | `[+-]? [0-9]+` |
| `DECIMAL` | `[+-]? [0-9]* '.' [0-9]+` |
| `DOUBLE` | `[+-]? ([0-9]+'.'[0-9]* EXPONENT \| '.'[0-9]+ EXPONENT \| [0-9]+ EXPONENT)` |
| `EXPONENT` | `[eE] [+-]? [0-9]+` |
| `ANON` | `'[' WS* ']'` |
| `UCHAR` | `'\u' HEX*4 \| '\U' HEX*8` |
| `ECHAR` | `'\' [tbnrf"'\`]` |
| `WS` | `#x20 \| #x9 \| #xD \| #xA` |

### Character class productions

| Rule | Definition |
|------|-----------|
| `PN_CHARS_BASE` | `[A-Z] \| [a-z] \| [#x00C0-#x00D6] \| [#x00D8-#x00F6] \| [#x00F8-#x02FF] \| [#x0370-#x037D] \| [#x037F-#x1FFF] \| [#x200C-#x200D] \| [#x2070-#x218F] \| [#x2C00-#x2FEF] \| [#x3001-#xD7FF] \| [#xF900-#xFDCF] \| [#xFDF0-#xFFFD] \| [#x10000-#xEFFFF]` |
| `PN_CHARS_U` | `PN_CHARS_BASE \| '_'` |
| `PN_CHARS` | `PN_CHARS_U \| '-' \| [0-9] \| #x00B7 \| [#x0300-#x036F] \| [#x203F-#x2040]` |
| `PN_PREFIX` | `PN_CHARS_BASE ((PN_CHARS \| '.')* PN_CHARS)?` |
| `PN_LOCAL` | `(PN_CHARS_U \| ':' \| [0-9] \| PLX) ((PN_CHARS \| '.' \| ':' \| PLX)* (PN_CHARS \| ':' \| PLX))?` |
| `PLX` | `PERCENT \| PN_LOCAL_ESC` |
| `PERCENT` | `'%' HEX HEX` |
| `HEX` | `[0-9] \| [A-F] \| [a-f]` |
| `PN_LOCAL_ESC` | `'\' ('_' \| '~' \| '.' \| '-' \| '!' \| '$' \| '&' \| "'" \| '(' \| ')' \| '*' \| '+' \| ',' \| ';' \| '=' \| '/' \| '?' \| '#' \| '@' \| '%')` |

## Parsing

### Parser State

Five items in parser state:
1. **`baseURI`** (IRI) — set by `base` production
2. **`namespaces`** (Map[prefix → IRI]) — set by `prefixID` production
3. **`bnodeLabels`** (Map[string → blank node]) — maps blank node labels to allocated nodes
4. **`curSubject`** (RDF_Term) — bound by `subject` production
5. **`curPredicate`** (RDF_Term) — bound by `verb` production; if `a`, bound to `rdf:type`

### Triple Emission

Each object `N` produces triple: `curSubject curPredicate N`.

### Property Lists (`blankNodePropertyList`)

- Beginning: save `curSubject` and `curPredicate`, set `curSubject` to a fresh blank node `B`
- The embedded `predicateObjectList` generates triples with `B` as subject
- Finishing: restore `curSubject` and `curPredicate`
- The node produced is blank node `B`

### Collections

- Beginning: save `curSubject` and `curPredicate`
- Each object gets a fresh blank node `B` with `curPredicate = rdf:first`
- Between objects: `objectN-1 rdf:rest objectN`
- Finishing: `curSubject rdf:rest rdf:nil`, restore `curSubject` and `curPredicate`
- Empty collection → `rdf:nil`

### RDF Term Constructors

| Production | Type | Procedure |
|-----------|------|-----------|
| IRIREF | IRI | Characters between `<` `>`, unescape numeric sequences, resolve relative IRIs |
| PNAME_NS | prefix / IRI | As prefix: key in namespaces map. As PrefixedName: look up namespace |
| PNAME_LN | IRI | Look up prefix namespace, unescape reserved chars in local part, concatenate |
| STRING_LITERAL_* | lexical form | Characters between delimiters, unescape sequences |
| LANGTAG | language tag | Characters after `@` |
| RDFLiteral | literal | Lexical form from String; if `^^iri`, datatype=iri; if LANGTAG, datatype=`rdf:langString`; else datatype=`xsd:string` |
| INTEGER | literal | Lexical form=input, datatype=`xsd:integer` |
| DECIMAL | literal | Lexical form=input, datatype=`xsd:decimal` |
| DOUBLE | literal | Lexical form=input, datatype=`xsd:double` |
| BooleanLiteral | literal | Lexical form=`true`/`false`, datatype=`xsd:boolean` |
| BLANK_NODE_LABEL | blank node | Look up in bnodeLabels; allocate if new |
| ANON | blank node | Generate fresh blank node |
| blankNodePropertyList | blank node | Generate fresh blank node |
| collection (non-empty) | blank node | Generate blank node for list head |
| collection (empty) | IRI | `rdf:nil` |

## Embedding Turtle in HTML

Turtle can be embedded in HTML `<script>` tags:

```html
<script type="text/turtle">
@prefix dc: <http://purl.org/dc/terms/> .
@prefix frbr: <http://purl.org/vocab/frbr/core#> .

<http://books.example.com/works/45U8QJGZSQKDH8N> a frbr:Work ;
     dc:creator "Wil Wheaton"@en ;
     dc:title "Just a Geek"@en .
</script>
```

- `<` and `>` do not need escaping inside script tags
- Each script data block is its own Turtle document with scoped `@prefix`/`@base` declarations
- In XHTML (CDocumentation: `application/xhtml+xml`), wrap in CDATA sections inside Turtle comments:

```turtle
# <![CDATA[
@prefix frbr: <http://purl.org/vocab/frbr/core#> .
<http://books.example.com/works/45U8QJGZSQKDH8N> a frbr:Work .
# ]]>
```

- If `]]>` appears in document content, escape with string escapes: `\u005d\u005d\u003e`

## Best Practices

1. **Use `@prefix`/`@base` over `PREFIX`/`BASE`** until RDF 1.1 Turtle parsers are widely deployed (per spec recommendation).
2. **Remember the trailing `.`** — `@prefix` and `@base` require it; `PREFIX` and `BASE` must NOT have it.
3. **Booleans are case-sensitive** — `TrUe` is NOT a valid boolean in Turtle (unlike SPARQL).
4. **No whitespace between sign and number** in signed numeric literals.
5. **%-encoding is NOT decoded** — `<http://a.example/%66oo-bar>` stays as-is, not resolved to `http://a.example/foo-bar`.
6. **Use longest match tokenizing** — the grammar is LL(1)/LALR(1).
7. **Declare prefixes before use** — though Turtle allows declarations anywhere outside triples, placing them at the top improves readability.
8. **Use `a` as shorthand** for `rdf:type` in predicate position to keep documents compact.
9. **Use predicate lists (`;`) and object lists (`,`) consistently** to reduce repetition and improve readability.
10. **Empty prefix `:` is valid** — `@prefix : <http://example.org/> .` allows `:subject :predicate :object .`

## Common Pitfalls

- **Forgetting the `.` after `@prefix`/`@base`** while adding it after `PREFIX`/`BASE` — the two styles have different requirements.
- **Using `TrUe`/`FALSE`** as booleans — case-sensitive, only lowercase `true`/`false` are valid.
- **Mixing Turtle and SPARQL syntax rules** — e.g., SPARQL allows variables and literals as subjects; Turtle does not.
- **Expecting `%-` decoding** — percent-encoded sequences in IRIs/local names are preserved, not decoded.
- **Placing `@prefix`/`@base` inside SPARQL syntax** — in SPARQL, declarations go only in the Prologue; in Turtle, they can go anywhere outside triples.
- **Whitespace in signed numbers** — `+ 5` is not a valid integer; must be `+5`.
- **Unescaped `\` in literals** — backslash is only valid as part of an escape sequence.
- **`]]>` in XHTML** — must be escaped with `\u005d\u005d\u003e` when using CDATA sections.

## Version Notes

- **RDF 1.1 Turtle** — W3C Recommendation, 25 February 2014. This is the current stable version.
- Added SPARQL-style `BASE` and `PREFIX` directives (case-insensitive, no trailing `.`).
- Title changed from "Turtle" to "RDF 1.1 Turtle" during the CR → PR transition.
- Local part of prefix names can now include `:` (non-leading).
- Added Turtle embedded in HTML support.
- Added reserved character escape sequences for local names.
- String escape sequences limited to strings; numeric escape sequences limited to IRIs and strings.
- Adopted SPARQL's syntax for prefixed names (digits in first character of local name, dots in non-first/non-last positions).
- Previous version: January 2008 Team Submission.