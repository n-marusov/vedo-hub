//! Axum handlers for the read-only SPARQL and CYPHER query endpoints.
//!
//! Provides:
//! - `POST /api/v1/sparql` — minimal SPARQL SELECT → Cypher translation
//! - `POST /api/v1/cypher` — direct Cypher execution with read-only enforcement
//!
//! The API Gateway already validates queries (read-only enforcement, LIMIT
//! injection) before proxying here. This handler applies **defense-in-depth**:
//! it re-validates the query server-side before touching Neo4j so a future
//! gateway misconfiguration cannot leak mutations to the upstream.

use std::sync::Arc;
use std::time::{Duration, Instant};

use axum::{
    extract::State,
    http::{header, StatusCode},
    response::{IntoResponse, Response},
    Json,
};
use serde::{Deserialize, Serialize};
use tracing::{debug, error, info, warn};

use crate::AppState;

/// Default row cap applied when the inbound query has no LIMIT. The gateway
/// injects its own cap (QUERY_MAX_LIMIT, default 1000) before forwarding, so
/// this value is the second line of defense for callers hitting the
/// ontology-service directly (e.g. internal operators).
const DEFAULT_QUERY_LIMIT: usize = 1000;

/// Maximum execution time for any single query. Tuned conservatively so a
/// runaway query cannot starve the connection pool.
const QUERY_TIMEOUT: Duration = Duration::from_secs(30);

/// Mutation keywords rejected by the server-side read-only enforcement.
/// The check uses whole-word matching so identifiers like `CreateResource`
/// are not flagged.
const SPARQL_MUTATION_KEYWORDS: &[&str] = &[
    "INSERT", "DELETE", "LOAD", "CLEAR", "DROP", "ADD", "MOVE", "COPY", "CREATE",
];
const CYPHER_MUTATION_KEYWORDS: &[&str] = &["CREATE", "DELETE", "SET", "REMOVE", "MERGE"];

/// Known read-only Cypher procedures that are safe to call via `CALL ...`.
/// CALL to any procedure not in this list is rejected as potentially unsafe.
/// This is an allowlist — only procedures proven read-only are added.
const READONLY_PROCEDURES: &[&str] = &[
    "db.labels",
    "db.relationshipTypes",
    "db.schema.nodeTypeProperties",
    "db.schema.relTypeProperties",
    "db.schema.visualization",
    "db.propertyKeys",
    "db.indexes",
    "db.constraints",
    "dbms.listConfig",
    "dbms.components",
];

/// Request body shared by both endpoints.
#[derive(Debug, Deserialize)]
pub struct QueryRequest {
    pub query: String,
}

/// Standard query response shape returned by both endpoints.
#[derive(Debug, Serialize)]
pub struct QueryResponse {
    pub results: Vec<serde_json::Value>,
    pub execution_time_ms: u64,
    pub triple_count: usize,
}

/// Error response shape — mirrors the gateway's `{"error":{"code","message"}}`.
fn error_response(status: StatusCode, code: &str, message: &str) -> Response {
    (
        status,
        [(header::CONTENT_TYPE, "application/json")],
        Json(serde_json::json!({
            "error": code,
            "detail": message,
        })),
    )
        .into_response()
}

/// `POST /api/v1/sparql` — execute a SPARQL SELECT query.
///
/// Currently supports the basic graph pattern (BGP) form:
/// `SELECT (?s ?p ?o | *) WHERE { ?s ?p ?o }` translated to the Cypher
/// `MATCH (s)-[p]->(o) RETURN s, p, o LIMIT N`. More complex SPARQL
/// (FILTER, OPTIONAL, UNION, property paths) returns 501 Not Implemented so
/// callers can route via the explicit Cypher endpoint.
pub async fn sparql_handler(
    State(state): State<Arc<AppState>>,
    Json(req): Json<QueryRequest>,
) -> Response {
    let start = Instant::now();
    debug!(query_len = req.query.len(), "SPARQL query received");

    if let Some(code) = validate_readonly(&req.query, QueryDialect::Sparql) {
        warn!(code, "SPARQL query rejected by server-side validator");
        return error_response(
            StatusCode::BAD_REQUEST,
            code,
            "Query rejected by server-side read-only enforcement",
        );
    }

    let translation = match translate_sparql_to_cypher(&req.query) {
        Ok(t) => t,
        Err(reason) => {
            warn!(reason, "SPARQL query not yet supported by translator");
            return error_response(
                StatusCode::NOT_IMPLEMENTED,
                "ONT-SPARQL-UNSUPPORTED",
                &format!(
                    "This SPARQL query form is not supported by the M1 translator: {reason}. \
                     Route complex queries via POST /api/v1/cypher."
                ),
            );
        }
    };

    let pool = match require_neo4j(&state) {
        Ok(p) => p,
        Err(resp) => return resp,
    };

    debug!(cypher = %translation.cypher, limit = translation.limit, "Executing translated Cypher");
    // [FIX] Pass limit as a Cypher parameter for defense-in-depth.
    let result = run_readonly_cypher_with_limit(pool, &translation.cypher, translation.limit).await;
    let execution_time_ms = start.elapsed().as_millis() as u64;

    match result {
        Ok(rows) => {
            let triple_count = rows.len();
            info!(
                rows = triple_count,
                execution_time_ms, "SPARQL query executed"
            );
            (
                StatusCode::OK,
                [(header::CONTENT_TYPE, "application/json")],
                Json(QueryResponse {
                    results: rows,
                    execution_time_ms,
                    triple_count,
                }),
            )
                .into_response()
        }
        Err(e) => {
            error!(error = %e, "Neo4j query failed for SPARQL translation");
            error_response(
                StatusCode::INTERNAL_SERVER_ERROR,
                "ONT-DATABASE-ERROR",
                "Query execution failed due to a database error",
            )
        }
    }
}

/// `POST /api/v1/cypher` — execute a (validated) read-only Cypher query.
pub async fn cypher_handler(
    State(state): State<Arc<AppState>>,
    Json(req): Json<QueryRequest>,
) -> Response {
    let start = Instant::now();
    debug!(query_len = req.query.len(), "CYPHER query received");

    if let Some(code) = validate_readonly(&req.query, QueryDialect::Cypher) {
        warn!(code, "CYPHER query rejected by server-side validator");
        return error_response(
            StatusCode::BAD_REQUEST,
            code,
            "Query rejected by server-side read-only enforcement",
        );
    }

    let pool = match require_neo4j(&state) {
        Ok(p) => p,
        Err(resp) => return resp,
    };

    let result = run_readonly_cypher(pool, &req.query).await;
    let execution_time_ms = start.elapsed().as_millis() as u64;

    match result {
        Ok(rows) => {
            let triple_count = rows.len();
            info!(
                rows = triple_count,
                execution_time_ms, "CYPHER query executed"
            );
            (
                StatusCode::OK,
                [(header::CONTENT_TYPE, "application/json")],
                Json(QueryResponse {
                    results: rows,
                    execution_time_ms,
                    triple_count,
                }),
            )
                .into_response()
        }
        Err(e) => {
            error!(error = %e, "Neo4j query failed for CYPHER");
            error_response(
                StatusCode::INTERNAL_SERVER_ERROR,
                "ONT-DATABASE-ERROR",
                "Query execution failed due to a database error",
            )
        }
    }
}

#[derive(Clone, Copy)]
enum QueryDialect {
    Sparql,
    Cypher,
}

/// Server-side defense-in-depth check. Returns the rejection error code when
/// the query contains a whole-word mutation keyword, otherwise `None`.
///
/// This duplicates the gateway validator's logic on purpose — the gateway is
/// a separate process that can be misconfigured or bypassed by internal
/// callers, so the upstream must protect itself.
fn validate_readonly(query: &str, dialect: QueryDialect) -> Option<&'static str> {
    let keywords = match dialect {
        QueryDialect::Sparql => SPARQL_MUTATION_KEYWORDS,
        QueryDialect::Cypher => CYPHER_MUTATION_KEYWORDS,
    };
    let upper = query.to_uppercase();
    for kw in keywords {
        if contains_whole_word(&upper, kw) {
            return Some("ONT-QUERY-READONLY");
        }
    }

    // Additional check for Cypher: CALL to write procedures can bypass keyword
    // detection if the procedure name doesn't contain a whole-word mutation
    // keyword (e.g. `apoc.periodic.commit` won't match any Cypher keyword).
    if matches!(dialect, QueryDialect::Cypher) && upper.contains("CALL") {
        if !contains_known_readonly_procedure(&upper) {
            return Some("ONT-QUERY-READONLY");
        }
    }

    None
}

/// Whole-word match so identifiers like `CreateResource` or `DeleteMe` are
/// not flagged. Word boundaries are defined by anything non-alphanumeric.
fn contains_whole_word(haystack: &str, needle: &str) -> bool {
    let needle_bytes = needle.as_bytes();
    if needle_bytes.is_empty() {
        return false;
    }
    let mut start = 0usize;
    while let Some(idx) = haystack[start..].find(needle) {
        let abs = start + idx;
        let before_ok = abs == 0 || !haystack.as_bytes()[abs - 1].is_ascii_alphanumeric();
        let after_idx = abs + needle.len();
        let after_ok =
            after_idx == haystack.len() || !haystack.as_bytes()[after_idx].is_ascii_alphanumeric();
        if before_ok && after_ok {
            return true;
        }
        start = abs + needle.len();
        if start >= haystack.len() {
            break;
        }
    }
    false
}

/// Checks if a query contains a `CALL` to a known read-only procedure.
/// Only procedures in the `READONLY_PROCEDURES` allowlist are permitted.
/// Returns `false` when the CALL targets an unknown (potentially write) procedure.
fn contains_known_readonly_procedure(upper: &str) -> bool {
    // Find CALL keyword
    let mut search_start = 0usize;
    while let Some(call_pos) = upper[search_start..].find("CALL") {
        let abs = search_start + call_pos;
        let after_call = &upper[abs + 4..];

        // Skip whitespace after CALL
        let proc_start = after_call.trim_start();
        if proc_start.is_empty() {
            search_start = abs + 4;
            continue;
        }

        // Extract procedure name: from the start up to '(' or whitespace
        let proc_name_end = proc_start
            .find(|c: char| c == '(' || c.is_whitespace() || c == ')')
            .unwrap_or(proc_start.len());
        let proc_name = &proc_start[..proc_name_end];

        // Check against allowlist (comparison is already uppercase)
        let is_allowed = READONLY_PROCEDURES
            .iter()
            .any(|&allowed| proc_name == allowed.to_uppercase().as_str());
        if !is_allowed {
            tracing::warn!(
                procedure = %proc_name,
                "CALL to unknown procedure rejected by server-side validator"
            );
            return false;
        }

        // Move past this CALL to check for more CALLs
        search_start = abs + 4;
    }
    true
}

/// Resolves the Neo4j pool from app state, returning the pool on success or
/// an error Response when the database is not configured.
fn require_neo4j(state: &AppState) -> Result<&crate::neo4j::Neo4jPool, Response> {
    state.neo4j.as_ref().map(|p| p).ok_or_else(|| {
        warn!("Query rejected: Neo4j not configured");
        error_response(
            StatusCode::SERVICE_UNAVAILABLE,
            "ONT-NEO4J-NOT-CONFIGURED",
            "Neo4j database is not configured",
        )
    })
}

/// Executes a Cypher query against the pool with a bounded timeout. Each row
/// is serialized to a `serde_json::Value` keyed by column name so callers
/// receive a JSON array of objects.
async fn run_readonly_cypher(
    pool: &crate::neo4j::Neo4jPool,
    query: &str,
) -> Result<Vec<serde_json::Value>, String> {
    let fut = pool.graph().execute(neo4rs::query(query));
    let mut result = tokio::time::timeout(QUERY_TIMEOUT, fut)
        .await
        .map_err(|_| {
            warn!(timeout_secs = QUERY_TIMEOUT.as_secs(), "Query timed out");
            format!("Query exceeded {QUERY_TIMEOUT:?} timeout")
        })?
        .map_err(|e| {
            error!(error = %e, "Neo4j execute failed");
            e.to_string()
        })?;

    let mut rows: Vec<serde_json::Value> = Vec::new();
    while let Ok(Some(row)) = result.next().await {
        // neo4rs 0.7 Row does not expose column names publicly, but `to::<T>`
        // deserializes the full attributes BoltMap — for serde_json::Value this
        // produces a JSON object keyed by RETURN column name.
        let value: serde_json::Value = row
            .to::<serde_json::Value>()
            .unwrap_or(serde_json::Value::Null);
        rows.push(value);
    }
    Ok(rows)
}

/// [FIX] Executes a Cypher query against the pool with a parameterized LIMIT.
/// The `$vedo_limit` parameter is bound to `limit` (as i64) so the LIMIT
/// value cannot be a source of Cypher injection.
async fn run_readonly_cypher_with_limit(
    pool: &crate::neo4j::Neo4jPool,
    query: &str,
    limit: usize,
) -> Result<Vec<serde_json::Value>, String> {
    let q = neo4rs::query(query).param("vedo_limit", limit as i64);
    let fut = pool.graph().execute(q);
    let mut result = tokio::time::timeout(QUERY_TIMEOUT, fut)
        .await
        .map_err(|_| {
            warn!(
                timeout_secs = QUERY_TIMEOUT.as_secs(),
                "[FIX] Parameterized query timed out"
            );
            format!("Query exceeded {QUERY_TIMEOUT:?} timeout")
        })?
        .map_err(|e| {
            error!(error = %e, "[FIX] Neo4j execute failed");
            e.to_string()
        })?;

    let mut rows: Vec<serde_json::Value> = Vec::new();
    while let Ok(Some(row)) = result.next().await {
        let value: serde_json::Value = row
            .to::<serde_json::Value>()
            .unwrap_or(serde_json::Value::Null);
        rows.push(value);
    }
    Ok(rows)
}

// ── SPARQL → Cypher translation (minimal M1 subset) ──────────────────────────

#[derive(Debug)]
struct Translation {
    cypher: String,
    limit: usize,
}

/// Translates the basic SPARQL BGP form to Cypher. Returns `Err(reason)` when
/// the query shape is not supported by the M1 translator.
fn translate_sparql_to_cypher(query: &str) -> Result<Translation, String> {
    let normalized = query.trim();
    let upper = normalized.to_uppercase();

    // Only SELECT queries are accepted (gateway already rejects INSERT/DELETE/etc,
    // but we double-check here for defense in depth).
    if !upper.starts_with("SELECT") {
        return Err("only SELECT queries are supported by the M1 translator".into());
    }

    // Extract optional LIMIT clause from the SPARQL query.
    let limit = extract_sparql_limit(&upper).unwrap_or(DEFAULT_QUERY_LIMIT);

    // Extract the WHERE clause body — text between the first `{` and the
    // matching `}` before any LIMIT clause.
    let where_body = extract_where_block(normalized)
        .ok_or_else(|| "could not locate a WHERE { ... } block".to_string())?;

    // Reject FILTER, OPTIONAL, UNION, property paths and other constructs.
    // This runs before triple tokenization so a FILTER clause (which the
    // naive tokenizer would otherwise count as a 3-token pseudo-triple) is
    // caught with the correct error message.
    for unsupported in [
        "FILTER", "OPTIONAL", "UNION", "DISTINCT", "GROUP", "ORDER", "OFFSET",
    ] {
        if upper.contains(unsupported) {
            return Err(format!(
                "{unsupported} clauses are not supported by the M1 translator"
            ));
        }
    }

    // Tokenize the WHERE body into triple patterns separated by `.`.
    let triples = parse_triples(&where_body);
    if triples.is_empty() {
        return Err("WHERE block contains no triple patterns".into());
    }
    if triples.len() > 1 {
        return Err("the M1 translator supports a single triple pattern per WHERE block".into());
    }
    let triple = &triples[0];

    let (s, p, o) = (&triple.subject, &triple.predicate, &triple.object);
    // Only the variable triple `?s ?p ?o` (in any order, with any var names)
    // is supported — translate to a full graph pattern.
    if s.starts_with('?') && p.starts_with('?') && o.starts_with('?') {
        let s_name = var_to_ident(s).ok_or_else(|| {
            format!("SPARQL variable '{s}' contains characters outside [a-zA-Z0-9_]")
        })?;
        let p_name = var_to_ident(p).ok_or_else(|| {
            format!("SPARQL variable '{p}' contains characters outside [a-zA-Z0-9_]")
        })?;
        let o_name = var_to_ident(o).ok_or_else(|| {
            format!("SPARQL variable '{o}' contains characters outside [a-zA-Z0-9_]")
        })?;
        // [FIX] Use Cypher parameter for LIMIT to avoid interpolation.
        // `limit` is a parsed usize but we still parameterize for defense-in-depth.
        let cypher = format!(
            "MATCH ({s_name})-[{p_name}]->({o_name}) \
             RETURN {s_name}, {p_name}, {o_name} LIMIT $vedo_limit"
        );
        tracing::debug!(
            cypher = %cypher,
            limit,
            "[FIX] SPARQL→Cypher translation with parameterized LIMIT"
        );
        return Ok(Translation { cypher, limit });
    }

    Err("only the variable triple pattern ?s ?p ?o is supported by the M1 translator".into())
}

#[derive(Debug, Clone)]
struct TriplePattern {
    subject: String,
    predicate: String,
    object: String,
}

/// Splits a WHERE block body into individual triple patterns terminated by `.`.
fn parse_triples(body: &str) -> Vec<TriplePattern> {
    let mut out = Vec::new();
    for stmt in body.split('.') {
        let toks: Vec<&str> = stmt.split_whitespace().collect();
        if toks.len() != 3 {
            continue;
        }
        out.push(TriplePattern {
            subject: toks[0].to_string(),
            predicate: toks[1].to_string(),
            object: toks[2].to_string(),
        });
    }
    out
}

/// Extracts the WHERE { ... } block contents (without the surrounding braces).
/// Returns `None` when the braces cannot be balanced.
fn extract_where_block(query: &str) -> Option<String> {
    let where_idx = query.to_uppercase().find("WHERE")?;
    let after_where = &query[where_idx..];
    let open = after_where.find('{')?;
    // Find the matching close brace, ignoring LIMIT clauses that follow.
    let mut depth = 0i32;
    let mut close = None;
    for (i, ch) in after_where[open..].char_indices() {
        match ch {
            '{' => depth += 1,
            '}' => {
                depth -= 1;
                if depth == 0 {
                    close = Some(open + i);
                    break;
                }
            }
            _ => {}
        }
    }
    let close = close?;
    Some(after_where[open + 1..close].trim().to_string())
}

/// Parses `LIMIT <n>` from an uppercased SPARQL query, if present.
fn extract_sparql_limit(upper: &str) -> Option<usize> {
    let limit_idx = upper.find("LIMIT")?;
    let after = &upper[limit_idx + "LIMIT".len()..];
    let token = after.split_whitespace().next()?;
    token.parse::<usize>().ok()
}

/// Strips the leading `?` from a SPARQL variable to make a Cypher identifier.
/// [FIX] Validates the resulting identifier against `^[a-zA-Z_][a-zA-Z0-9_]*$`
/// and returns `None` when the variable contains characters outside that set.
/// This prevents Cypher injection via malicious SPARQL variable names.
fn var_to_ident(var: &str) -> Option<&str> {
    let stripped = var.trim_start_matches('?');
    if stripped.is_empty() {
        return None;
    }
    let mut chars = stripped.chars();
    let first = chars.next().unwrap();
    if !(first.is_ascii_alphabetic() || first == '_') {
        tracing::warn!(
            var = %var,
            "[FIX] Rejecting SPARQL variable with invalid first character"
        );
        return None;
    }
    if !chars.all(|c| c.is_ascii_alphanumeric() || c == '_') {
        tracing::warn!(
            var = %var,
            "[FIX] Rejecting SPARQL variable with invalid characters"
        );
        return None;
    }
    Some(stripped)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_validate_readonly_cypher_rejects_mutations() {
        assert_eq!(
            validate_readonly("MATCH (n) DELETE n", QueryDialect::Cypher),
            Some("ONT-QUERY-READONLY")
        );
        assert_eq!(
            validate_readonly("CREATE (n:Foo)", QueryDialect::Cypher),
            Some("ONT-QUERY-READONLY")
        );
    }

    #[test]
    fn test_validate_readonly_cypher_allows_class_named_createresource() {
        // Regression guard for the naive `Contains` false positive.
        assert_eq!(
            validate_readonly("MATCH (n:CreateResource) RETURN n", QueryDialect::Cypher),
            None,
        );
        assert_eq!(
            validate_readonly("MATCH (n:DeleteMe) RETURN n", QueryDialect::Cypher),
            None,
        );
    }

    #[test]
    fn test_validate_readonly_sparql_rejects_insert() {
        assert_eq!(
            validate_readonly("INSERT DATA { <a> <b> <c> }", QueryDialect::Sparql),
            Some("ONT-QUERY-READONLY")
        );
    }

    #[test]
    fn test_translate_sparql_basic_bgp() {
        let t = translate_sparql_to_cypher("SELECT ?s ?p ?o WHERE { ?s ?p ?o }")
            .expect("basic BGP must translate");
        assert!(t.cypher.contains("MATCH (s)-[p]->(o)"));
        assert!(t.cypher.contains("RETURN s, p, o"));
        // [FIX] LIMIT is now parameterized as $vedo_limit
        assert!(t.cypher.contains("LIMIT $vedo_limit"));
        assert_eq!(t.limit, 1000);
    }

    #[test]
    fn test_translate_sparql_preserves_explicit_limit() {
        let t = translate_sparql_to_cypher("SELECT ?s ?p ?o WHERE { ?s ?p ?o } LIMIT 50")
            .expect("explicit LIMIT must be honored");
        assert!(t.cypher.contains("LIMIT $vedo_limit"));
        assert_eq!(t.limit, 50);
    }

    #[test]
    fn test_translate_sparql_rejects_malicious_variable_name() {
        // [FIX] SPARQL variable names that would inject Cypher must be rejected.
        // `?s-p` contains a hyphen which is outside [a-zA-Z0-9_].
        let err = translate_sparql_to_cypher("SELECT ?s-p ?p ?o WHERE { ?s-p ?p ?o }");
        match err {
            Err(reason) => {
                assert!(
                    reason.contains("invalid") || reason.contains("characters"),
                    "expected rejection of malicious variable, got: {reason}"
                );
            }
            Ok(t) => {
                assert!(
                    !t.cypher.contains("-"),
                    "malicious variable leaked into Cypher: {}",
                    t.cypher
                );
            }
        }
    }

    #[test]
    fn test_translate_sparql_limit_is_parameterized() {
        // [FIX] LIMIT must appear as a Cypher parameter, not an integer literal.
        let t = translate_sparql_to_cypher("SELECT ?s ?p ?o WHERE { ?s ?p ?o }")
            .expect("basic BGP must translate");
        assert!(
            t.cypher.contains("LIMIT $vedo_limit"),
            "LIMIT must be parameterized as $vedo_limit, got: {}",
            t.cypher
        );
        // Ensure no integer literal follows LIMIT (which would mean interpolation).
        assert!(!t.cypher.contains("LIMIT 1000"));
    }

    #[test]
    fn test_var_to_ident_rejects_invalid_chars() {
        // [FIX] var_to_ident must reject characters outside [a-zA-Z_][a-zA-Z0-9_]*
        assert_eq!(var_to_ident("?s"), Some("s"));
        assert_eq!(var_to_ident("?subject_name"), Some("subject_name"));
        assert_eq!(var_to_ident("?_under"), Some("_under"));
        // Invalid: starts with digit
        assert_eq!(var_to_ident("?1s"), None);
        // Invalid: contains parentheses (injection attempt)
        assert_eq!(var_to_ident("?s)-[p"), None);
        // Invalid: contains hyphen
        assert_eq!(var_to_ident("?s-name"), None);
        // Invalid: contains dot
        assert_eq!(var_to_ident("?s.name"), None);
        // Empty after stripping ?
        assert_eq!(var_to_ident("?"), None);
    }

    #[test]
    fn test_translate_sparql_rejects_filter() {
        let err =
            translate_sparql_to_cypher("SELECT ?s ?p ?o WHERE { ?s ?p ?o . FILTER(?s = <a>) }")
                .expect_err("FILTER must be rejected");
        assert!(err.contains("FILTER"));
    }

    #[test]
    fn test_translate_sparql_rejects_multi_pattern() {
        let err = translate_sparql_to_cypher("SELECT ?s ?o WHERE { ?s ?p ?o . ?o ?p2 ?o2 }")
            .expect_err("multi-pattern must be rejected");
        assert!(err.contains("single triple"));
    }

    #[test]
    fn test_contains_whole_word_boundaries() {
        assert!(contains_whole_word("INSERT DATA", "INSERT"));
        assert!(contains_whole_word("(CREATE)", "CREATE"));
        // Identifier glued to the keyword — must NOT match.
        assert!(!contains_whole_word("CREATERESOURCE", "CREATE"));
        assert!(!contains_whole_word("CreateResource", "CREATE"));
        assert!(!contains_whole_word("DELETEME", "DELETE"));
    }

    #[test]
    fn test_validate_readonly_cypher_rejects_comments_containing_keywords() {
        // Line comment with mutation keyword — must be rejected.
        assert_eq!(
            validate_readonly("MATCH (n) // CREATE\nRETURN n", QueryDialect::Cypher),
            Some("ONT-QUERY-READONLY"),
            "comment with CREATE must be rejected"
        );
        // Block comment with mutation keyword.
        assert_eq!(
            validate_readonly("MATCH (n) /* DELETE */ RETURN n", QueryDialect::Cypher),
            Some("ONT-QUERY-READONLY"),
            "block comment with DELETE must be rejected"
        );
        // MERGE inside a comment.
        assert_eq!(
            validate_readonly(
                "MATCH (n) // MERGE with neighbor\nRETURN n",
                QueryDialect::Cypher
            ),
            Some("ONT-QUERY-READONLY"),
            "comment with MERGE must be rejected"
        );
    }

    #[test]
    fn test_validate_readonly_sparql_rejects_comments_containing_keywords() {
        assert_eq!(
            validate_readonly(
                "SELECT ?s WHERE { ?s ?p ?o } # INSERT DATA",
                QueryDialect::Sparql
            ),
            Some("ONT-QUERY-READONLY"),
            "comment with INSERT must be rejected for SPARQL"
        );
    }

    #[test]
    fn test_validate_readonly_false_positive_string_literal() {
        // String literals containing mutation keywords cause false positives:
        // the keyword is not a Cypher/SPARQL mutation statement, but the
        // word-boundary validator cannot distinguish it from a real keyword.
        // This test documents the known limitation — see Task 2.1 for the fix.
        assert_eq!(
            validate_readonly(
                "MATCH (n) WHERE n.name = 'CREATE' RETURN n",
                QueryDialect::Cypher
            ),
            Some("ONT-QUERY-READONLY"),
            "string literal 'CREATE' is falsely rejected (known limitation)"
        );
        assert_eq!(
            validate_readonly(
                "MATCH (n) WHERE n.label CONTAINS 'DELETE' RETURN n",
                QueryDialect::Cypher
            ),
            Some("ONT-QUERY-READONLY"),
            "string literal 'DELETE' is falsely rejected (known limitation)"
        );
    }

    #[tokio::test]
    async fn test_error_response_format_sanitized() {
        // The error_response function must not expose raw internal error details.
        // This test verifies the output shape; the handler fix (Task 2.2) will
        // ensure callers pass sanitized messages instead of raw DB errors.
        let resp = error_response(
            StatusCode::INTERNAL_SERVER_ERROR,
            "ONT-DATABASE-ERROR",
            "Query execution failed",
        );
        let (parts, body) = resp.into_parts();
        assert_eq!(parts.status, StatusCode::INTERNAL_SERVER_ERROR);

        // Read the body JSON.
        let bytes = axum::body::to_bytes(body, 1024)
            .await
            .expect("body must be readable");
        let json: serde_json::Value =
            serde_json::from_slice(&bytes).expect("body must be valid JSON");

        assert_eq!(json["error"], "ONT-DATABASE-ERROR");
        assert_eq!(json["detail"], "Query execution failed");
        // The response must NOT contain internal error markers.
        let detail = json["detail"].as_str().unwrap_or("");
        let forbidden = ["Neo4j", "neo4rs", "crypto", "signature", "runtime error"];
        for pattern in &forbidden {
            assert!(
                !detail.contains(pattern),
                "error detail must not contain '{}': got '{}'",
                pattern,
                detail
            );
        }
    }
}
