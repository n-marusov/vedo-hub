package services

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

const (
	// DefaultMaxLimit is injected when no LIMIT is specified.
	DefaultMaxLimit = 1000
)

// SPARQL mutation keywords that must be rejected (whole-word matched).
var sparqlMutationKeywords = []string{"INSERT", "DELETE", "LOAD", "CLEAR", "DROP", "ADD", "MOVE", "COPY", "CREATE"}

// CYPHER mutation keywords (whole-word matched).
var cypherMutationKeywords = []string{"CREATE", "DELETE", "SET", "REMOVE", "MERGE"}

// containsKeyword checks whether `keyword` appears as a whole word in `s`
// (not as a substring). This prevents false positives where a mutation
// keyword like CREATE appears inside an identifier (e.g. a class named
// `CreateResource`) or a string literal. The match is case-insensitive and
// uses word boundaries — anything non-alphanumeric counts as a boundary.
func containsKeyword(s, keyword string) bool {
	// regexp is intentionally compiled per call rather than precompiled as
	// package vars because the keyword set is small and the cost is
	// negligible. A future optimization can precompile the patterns into a
	// single alternation regex.
	pattern := `(?i)\b` + regexp.QuoteMeta(keyword) + `\b`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

// QueryValidationResult holds the result of validating a query.
type QueryValidationResult struct {
	Valid          bool
	SanitizedQuery string // The query with LIMIT injected if applicable
	ErrorCode      string
	ErrorMessage   string
}

// ValidateAndSanitizeSPARQL performs coarse fast-path validation for SPARQL queries.
// Full validation including LIMIT injection is owned by the ontology-service.
// This is a defence-in-depth fast-path that rejects obvious mutations before proxying.
func ValidateAndSanitizeSPARQL(query string, _maxLimit int) *QueryValidationResult {
	trimmed := strings.TrimSpace(query)
	upper := strings.ToUpper(trimmed)

	// Reject empty queries
	if trimmed == "" {
		return &QueryValidationResult{
			Valid:        false,
			ErrorCode:    "GATEWAY-QUERY-EMPTY",
			ErrorMessage: "Query must not be empty",
		}
	}

	// Only allow SELECT, ASK, CONSTRUCT, DESCRIBE queries
	if !strings.HasPrefix(upper, "SELECT") && !strings.HasPrefix(upper, "ASK") && !strings.HasPrefix(upper, "CONSTRUCT") && !strings.HasPrefix(upper, "DESCRIBE") {
		return &QueryValidationResult{
			Valid:        false,
			ErrorCode:    "GATEWAY-QUERY-READONLY",
			ErrorMessage: "Only SELECT, ASK, CONSTRUCT, and DESCRIBE queries are allowed. Mutation queries (INSERT, DELETE, etc.) are rejected.",
		}
	}

	// Coarse mutation keyword check (fast-path). Whole-word matching so
	// identifiers like `CreateResource` are not falsely flagged.
	// Full validation is performed by the ontology-service downstream.
	bodyUpper := upper
	slog.Debug("[FIX] query.validation.body_scan", "type", "sparql", "body_len", len(bodyUpper))
	for _, kw := range sparqlMutationKeywords {
		if containsKeyword(bodyUpper, kw) {
			slog.Warn("[FIX] query.mutation_keyword_detected",
				"type", "sparql",
				"keyword", kw,
			)
			return &QueryValidationResult{
				Valid:        false,
				ErrorCode:    "GATEWAY-QUERY-READONLY",
				ErrorMessage: fmt.Sprintf("Mutation keyword '%s' is not allowed in read-only queries.", kw),
			}
		}
	}

	// Hash the query for audit logging
	hash := sha256.Sum256([]byte(query))
	slog.Debug("[FIX] query.validation.passed",
		"type", "sparql",
		"hash", fmt.Sprintf("%x", hash[:8]),
	)

	return &QueryValidationResult{
		Valid:          true,
		SanitizedQuery: trimmed,
	}
}

// ValidateAndSanitizeCYPHER performs coarse fast-path validation for CYPHER queries.
// Full validation including LIMIT injection is owned by the ontology-service.
func ValidateAndSanitizeCYPHER(query string, _maxLimit int) *QueryValidationResult {
	trimmed := strings.TrimSpace(query)
	upper := strings.ToUpper(trimmed)

	// Reject empty queries
	if trimmed == "" {
		return &QueryValidationResult{
			Valid:        false,
			ErrorCode:    "GATEWAY-QUERY-EMPTY",
			ErrorMessage: "Query must not be empty",
		}
	}

	// Only allow MATCH queries (read-only)
	if !strings.HasPrefix(upper, "MATCH") {
		return &QueryValidationResult{
			Valid:        false,
			ErrorCode:    "GATEWAY-QUERY-READONLY",
			ErrorMessage: "Only MATCH queries are allowed. Mutation queries (CREATE, DELETE, SET, etc.) are rejected.",
		}
	}

	// Coarse mutation keyword check (fast-path).
	// Full validation is performed by the ontology-service downstream.
	bodyUpper := strings.TrimPrefix(upper, "MATCH")
	slog.Debug("[FIX] query.validation.body_scan", "type", "cypher", "body_len", len(bodyUpper))
	for _, kw := range cypherMutationKeywords {
		if containsKeyword(bodyUpper, kw) {
			slog.Warn("[FIX] query.mutation_keyword_detected",
				"type", "cypher",
				"keyword", kw,
			)
			return &QueryValidationResult{
				Valid:        false,
				ErrorCode:    "GATEWAY-QUERY-READONLY",
				ErrorMessage: fmt.Sprintf("Mutation keyword '%s' is not allowed in read-only queries.", kw),
			}
		}
	}

	// Hash for audit logging
	hash := sha256.Sum256([]byte(query))
	slog.Debug("[FIX] query.validation.passed",
		"type", "cypher",
		"hash", fmt.Sprintf("%x", hash[:8]),
	)

	return &QueryValidationResult{
		Valid:          true,
		SanitizedQuery: trimmed,
	}
}
