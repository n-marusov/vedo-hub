package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	ontologyv1 "vedo-core/src/services/shared/proto/ontology/v1"

	"vedo-core/src/services/api-gateway/proxy"
	"vedo-core/src/services/api-gateway/services"
)

// QueryRequest is the expected JSON body for SPARQL/CYPHER queries.
type QueryRequest struct {
	Query string `json:"query"`
}

// QueryHandler handles SPARQL and CYPHER query endpoints with read-only
// enforcement, mandatory LIMIT injection, rate limiting, and audit logging.
type QueryHandler struct {
	ontologyProxy *proxy.Proxy
	grpcClient    *proxy.OntologyServiceClient
	rateLimiter   *services.RateLimiter
	maxLimit      int
}

// NewQueryHandler creates a new QueryHandler with optional gRPC client.
func NewQueryHandler(ontologyProxy *proxy.Proxy, maxLimit int, grpcClient *proxy.OntologyServiceClient) *QueryHandler {
	if maxLimit <= 0 {
		maxLimit = 1000
	}
	return &QueryHandler{
		ontologyProxy: ontologyProxy,
		grpcClient:    grpcClient,
		rateLimiter:   services.NewRateLimiter(maxLimit),
		maxLimit:      maxLimit,
	}
}

// HandleSPARQL handles POST /api/v1/sparql.
func (h *QueryHandler) HandleSPARQL(c *gin.Context) {
	h.handleQuery(c, "sparql")
}

// HandleCYPHER handles POST /api/v1/cypher.
func (h *QueryHandler) HandleCYPHER(c *gin.Context) {
	h.handleQuery(c, "cypher")
}

func (h *QueryHandler) handleQuery(c *gin.Context, queryType string) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")
	userID := c.GetHeader("X-User-Id")
	sourceIP := c.ClientIP()

	// Parse the request body
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("query.invalid_body",
			"type", queryType,
			"trace_id", traceID,
			"error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "GATEWAY-QUERY-INVALID-BODY",
				"message": "Invalid request body. Expected {\"query\": \"...\"}.",
			},
		})
		return
	}

	// Validate query
	var validation *services.QueryValidationResult
	switch queryType {
	case "sparql":
		validation = services.ValidateAndSanitizeSPARQL(req.Query, h.maxLimit)
	case "cypher":
		validation = services.ValidateAndSanitizeCYPHER(req.Query, h.maxLimit)
	}

	if !validation.Valid {
		slog.Warn("query.rejected",
			"type", queryType,
			"trace_id", traceID,
			"user_id", userID,
			"error_code", validation.ErrorCode,
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    validation.ErrorCode,
				"message": validation.ErrorMessage,
			},
		})
		return
	}

	// Rate limiting
	roles := c.GetStringSlice("auth_roles")
	tier := services.ResolveTier(roles)
	rateLimitKey := userID
	if rateLimitKey == "" {
		rateLimitKey = sourceIP
	}

	allowed, retryAfter := h.rateLimiter.Allow(rateLimitKey, tier)
	if !allowed {
		seconds := int(retryAfter.Seconds()) + 1
		slog.Warn("query.rate_limited",
			"type", queryType,
			"tier", tier,
			"trace_id", traceID,
			"user_id", userID,
			"retry_after", seconds,
		)
		c.Header("Retry-After", fmt.Sprintf("%d", seconds))
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": gin.H{
				"code":    "GATEWAY-RATE-LIMITED",
				"message": fmt.Sprintf("Rate limit exceeded. Retry after %d seconds.", seconds),
			},
		})
		return
	}

	// Audit logging
	queryHash := sha256Hash(req.Query)
	slog.Info("query.executing",
		"type", queryType,
		"trace_id", traceID,
		"user_id", userID,
		"source_ip", sourceIP,
		"query_hash", queryHash[:16],
		"tier", tier,
	)

	// Try gRPC first; fallback to HTTP proxy
	if h.grpcClient != nil {
		if err := h.forwardViaGRPC(c, queryType, validation.SanitizedQuery); err == nil {
			duration := time.Since(start).Milliseconds()
			slog.Debug("query.completed.via_grpc",
				"type", queryType,
				"trace_id", traceID,
				"duration_ms", duration,
			)
			return
		}
		slog.Debug("query.grpc_fallback_to_http",
			"type", queryType,
			"trace_id", traceID,
		)
	}

	// Fallback: Replace the request body with the sanitized query and forward to proxy
	sanitizedBody, _ := json.Marshal(QueryRequest{Query: validation.SanitizedQuery})
	c.Request.Body = io.NopCloser(bytes.NewReader(sanitizedBody))
	c.Request.ContentLength = int64(len(sanitizedBody))

	// Forward to upstream via proxy
	h.ontologyProxy.ServeHTTP(c.Writer, c.Request)

	// Log completion
	duration := time.Since(start).Milliseconds()
	slog.Debug("query.completed",
		"type", queryType,
		"trace_id", traceID,
		"duration_ms", duration,
	)
}

// forwardViaGRPC attempts to execute the query via gRPC.
// Currently falls back to HTTP since server-side gRPC stubs are still
// being implemented. Will be activated when server-side migration is complete.
func (h *QueryHandler) forwardViaGRPC(c *gin.Context, queryType, sanitizedQuery string) error {
	ontologyID := c.Param("id")
	if ontologyID == "" {
		return errGRPCNotAvailable
	}

	switch queryType {
	case "sparql":
		_, err := h.grpcClient.ExecuteSPARQL(c.Request.Context(), &ontologyv1.ExecuteSPARQLRequest{
			OntologyId: ontologyID,
			Query:      sanitizedQuery,
			MaxResults: int32(h.maxLimit),
		})
		if err != nil {
			slog.Debug("query.grpc_sparql_failed", "error", err)
			return err
		}
	case "cypher":
		_, err := h.grpcClient.ExecuteCYPHER(c.Request.Context(), &ontologyv1.ExecuteCYPHERRequest{
			OntologyId: ontologyID,
			Query:      sanitizedQuery,
			MaxResults: int32(h.maxLimit),
		})
		if err != nil {
			slog.Debug("query.grpc_cypher_failed", "error", err)
			return err
		}
	}
	return nil
}

func sha256Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}
