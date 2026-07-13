package services

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"strings"
)

const (
	// DefaultMaxLimit is injected when no LIMIT is specified.
	DefaultMaxLimit = 1000

	// SPARQL mutation keywords that must be rejected.
	sparqlMutationKeywords = "INSERT|DELETE|LOAD|CLEAR|DROP|ADD|MOVE|COPY|CREATE"
	// CYPHER mutation keywords.
	cypherMutationKeywords = "CREATE|DELETE|SET|REMOVE|MERGE"
)

// QueryValidationResult holds the result of validating a query.
type QueryValidationResult struct {
	Valid          bool
	SanitizedQuery string // The query with LIMIT injected if applicable
	ErrorCode      string
	ErrorMessage   string
}

// ValidateAndSanitizeSPARQL validates and sanitizes a SPARQL query.
func ValidateAndSanitizeSPARQL(query string, maxLimit int) *QueryValidationResult {
	trimmed := strings.TrimSpace(query)
	upper := strings.ToUpper(trimmed)

	if maxLimit <= 0 {
		maxLimit = DefaultMaxLimit
	}

	// Reject empty queries
	if trimmed == "" {
		return &QueryValidationResult{
			Valid:        false,
			ErrorCode:    "GATEWAY-QUERY-EMPTY",
			ErrorMessage: "Query must not be empty",
		}
	}

	// Only allow SELECT and ASK queries
	if !strings.HasPrefix(upper, "SELECT") && !strings.HasPrefix(upper, "ASK") && !strings.HasPrefix(upper, "CONSTRUCT") && !strings.HasPrefix(upper, "DESCRIBE") {
		return &QueryValidationResult{
			Valid:        false,
			ErrorCode:    "GATEWAY-QUERY-READONLY",
			ErrorMessage: "Only SELECT, ASK, CONSTRUCT, and DESCRIBE queries are allowed. Mutation queries (INSERT, DELETE, etc.) are rejected.",
		}
	}

	// Additional safety check: reject any mutation keywords in the query body
	re := strings.NewReplacer("SELECT", "", "ASK", "", "CONSTRUCT", "", "DESCRIBE", "")
	bodyUpper := re.Replace(upper)
	for _, kw := range strings.Split(sparqlMutationKeywords, "|") {
		if strings.Contains(bodyUpper, kw) {
			return &QueryValidationResult{
				Valid:        false,
				ErrorCode:    "GATEWAY-QUERY-READONLY",
				ErrorMessage: fmt.Sprintf("Mutation keyword '%s' is not allowed in read-only queries.", kw),
			}
		}
	}

	// Inject LIMIT if not present (for SELECT queries)
	result := trimmed
	if strings.HasPrefix(upper, "SELECT") && !strings.Contains(upper, "LIMIT") {
		result = fmt.Sprintf("%s LIMIT %d", strings.TrimRight(trimmed, "; "), maxLimit)
	}

	// Hash the query for audit logging
	hash := sha256.Sum256([]byte(query))
	slog.Debug("query.validation",
		"type", "sparql",
		"hash", fmt.Sprintf("%x", hash[:8]),
		"injected_limit", !strings.Contains(upper, "LIMIT"),
	)

	return &QueryValidationResult{
		Valid:          true,
		SanitizedQuery: result,
	}
}

// ValidateAndSanitizeCYPHER validates and sanitizes a CYPHER query.
func ValidateAndSanitizeCYPHER(query string, maxLimit int) *QueryValidationResult {
	trimmed := strings.TrimSpace(query)
	upper := strings.ToUpper(trimmed)

	if maxLimit <= 0 {
		maxLimit = DefaultMaxLimit
	}

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

	// Check for mutation keywords in the query body beyond the prefix
	bodyUpper := strings.TrimPrefix(upper, "MATCH")
	for _, kw := range strings.Split(cypherMutationKeywords, "|") {
		if strings.Contains(bodyUpper, kw) {
			return &QueryValidationResult{
				Valid:        false,
				ErrorCode:    "GATEWAY-QUERY-READONLY",
				ErrorMessage: fmt.Sprintf("Mutation keyword '%s' is not allowed in read-only queries.", kw),
			}
		}
	}

	// Inject LIMIT if not present
	result := trimmed
	if !strings.Contains(upper, "LIMIT") {
		result = fmt.Sprintf("%s LIMIT %d", strings.TrimRight(trimmed, "; "), maxLimit)
	}

	// Hash for audit logging
	hash := sha256.Sum256([]byte(query))
	slog.Debug("query.validation",
		"type", "cypher",
		"hash", fmt.Sprintf("%x", hash[:8]),
		"injected_limit", !strings.Contains(upper, "LIMIT"),
	)

	return &QueryValidationResult{
		Valid:          true,
		SanitizedQuery: result,
	}
}
