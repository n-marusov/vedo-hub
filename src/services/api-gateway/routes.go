package main

import (
	_ "embed"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"vedo-core/src/services/api-gateway/handlers"
	"vedo-core/src/services/api-gateway/models"
	"vedo-core/src/services/api-gateway/proxy"
)

//go:embed docs/openapi.json
var openAPISpec []byte

// RegisterRoutes defines all API route groups and wires them to the proxy handlers.
// The proxies are configured via environment variables pointing to upstream services.
//
// Gin's trie-based router forbids mixing wildcards (`+"`"+`/*path`) with named params
// (`+"`"+`:id`) or static segments at the same level. We therefore register every
// proxied endpoint explicitly instead of using a catch-all wildcard. Unknown
// `/api/v1/*` paths fall through to `r.NoRoute` (set up by the caller) which
// returns a uniform `GATEWAY-NOT-FOUND` error.
func RegisterRoutes(r *gin.Engine, grpcPool *proxy.GrpcClientPool) {
	// Initialize upstream proxies (HTTP reverse proxy — legacy, being replaced by gRPC)
	ontologyProxy := mustNewProxy(
		getEnv("ONTOLOGY_SERVICE_URL", "http://localhost:8082"),
		"ontology-service",
	)
	versioningProxy := mustNewProxy(
		getEnv("VERSIONING_SERVICE_URL", "http://localhost:8083"),
		"versioning-service",
	)

	// Create gRPC service clients for backend communication.
	ontologyGrpc := proxy.NewOntologyServiceClient(grpcPool, proxy.GrpcAddrOntology)
	versioningGrpc := proxy.NewVersioningServiceClient(grpcPool, proxy.GrpcAddrVersioning)
	authGrpc := proxy.NewAuthServiceClient(grpcPool, proxy.GrpcAddrAuth)
	aiOrchGrpc := proxy.NewAIOrchestrationServiceClient(grpcPool, proxy.GrpcAddrAIOrch)

	// Suppress unused variable warnings for future-use clients
	_ = versioningGrpc
	_ = authGrpc

	// API v1 routes — auth middleware applied at engine level (see main.go)
	api := r.Group("/api/v1")

	// Ontology REST read handlers — use gRPC.
	ontologyHandler := handlers.NewOntologyHandler(ontologyProxy, ontologyGrpc)
	api.GET("/ontologies", ontologyHandler.HandleListOntologies)
	api.GET("/ontologies/:id", ontologyHandler.HandleGetOntology)
	api.GET("/ontologies/:id/classes", ontologyHandler.HandleListClasses)
	api.GET("/ontologies/:id/classes/:classId", ontologyHandler.HandleGetClass)
	api.GET("/ontologies/:id/properties", ontologyHandler.HandleListProperties)
	api.GET("/ontologies/:id/individuals", ontologyHandler.HandleListIndividuals)

	// Ontology write endpoints (POST/PUT/DELETE) — HTTP proxy (legacy, migrating to gRPC).
	api.POST("/ontologies", gin.WrapH(ontologyProxy))
	api.PUT("/ontologies/:id", gin.WrapH(ontologyProxy))
	api.DELETE("/ontologies/:id", gin.WrapH(ontologyProxy))
	api.POST("/ontologies/:id/classes", gin.WrapH(ontologyProxy))
	api.PUT("/ontologies/:id/classes/:classId", gin.WrapH(ontologyProxy))
	api.DELETE("/ontologies/:id/classes/:classId", gin.WrapH(ontologyProxy))
	api.POST("/ontologies/:id/properties", gin.WrapH(ontologyProxy))
	api.PUT("/ontologies/:id/properties/:propertyId", gin.WrapH(ontologyProxy))
	api.DELETE("/ontologies/:id/properties/:propertyId", gin.WrapH(ontologyProxy))
	api.POST("/ontologies/:id/individuals", gin.WrapH(ontologyProxy))
	api.PUT("/ontologies/:id/individuals/:individualId", gin.WrapH(ontologyProxy))
	api.DELETE("/ontologies/:id/individuals/:individualId", gin.WrapH(ontologyProxy))
	api.GET("/ontologies/:id/export", gin.WrapH(ontologyProxy))
	api.POST("/ontologies/:id/import", gin.WrapH(ontologyProxy))

	// Versioning service proxy.
	api.POST("/versioning/commits", gin.WrapH(versioningProxy))
	api.GET("/versioning/commits", gin.WrapH(versioningProxy))
	api.GET("/versioning/commits/:id", gin.WrapH(versioningProxy))
	api.GET("/versioning/commits/:id/delta", gin.WrapH(versioningProxy))
	api.POST("/versioning/commits/:id/checkout", gin.WrapH(versioningProxy))
	api.POST("/versioning/commits/:id/rollback", gin.WrapH(versioningProxy))

	api.POST("/versioning/branches", gin.WrapH(versioningProxy))
	api.GET("/versioning/branches", gin.WrapH(versioningProxy))
	api.GET("/versioning/branches/:id", gin.WrapH(versioningProxy))
	api.DELETE("/versioning/branches/:id", gin.WrapH(versioningProxy))
	api.POST("/versioning/branches/:id/switch", gin.WrapH(versioningProxy))
	api.POST("/versioning/branches/merge", gin.WrapH(versioningProxy))

	// Query routes — SPARQL/CYPHER with read-only enforcement.
	queryMaxLimit := 1000
	if v := getEnv("QUERY_MAX_LIMIT", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			queryMaxLimit = n
		}
	}
	queryHandler := handlers.NewQueryHandler(ontologyProxy, queryMaxLimit, ontologyGrpc)
	api.POST("/sparql", queryHandler.HandleSPARQL)
	api.POST("/cypher", queryHandler.HandleCYPHER)

	// GraphQL endpoint — proxy to ontology-service
	api.Any("/graphql", gin.WrapH(ontologyProxy))

	// Create AI orchestration proxy — thin HTTP→gRPC bridge. All AI business
	// logic now lives in the ai-orchestration-service.
	aiOrchProxy := handlers.NewAIOrchProxy(aiOrchGrpc, ontologyGrpc)

	// AI-related routes — proxied to ai-orchestration-service via gRPC.
	// Policy routing and prompt injection defense are handled server-side.
	api.POST("/ontologies/:id/generate-from-text", aiOrchProxy.HandleGenerateFromText)
	api.POST("/ontologies/:id/ai/suggest-classes", aiOrchProxy.HandleSuggestClasses)
	api.POST("/ontologies/:id/ai/suggest-properties", aiOrchProxy.HandleSuggestProperties)
	api.POST("/ontologies/:id/ai/suggest-relationships", aiOrchProxy.HandleSuggestRelationships)
	api.POST("/ontologies/:id/ai/complete", aiOrchProxy.HandleSuggestClasses) // legacy alias
	api.POST("/ontologies/:id/ai/refine", aiOrchProxy.HandleRefine)

	// Document extractor proxy (HTTP — will later migrate to gRPC)
	// Adds X-Ontology-Id header from the path parameter so the downstream
	// service can perform policy checks and usage logging.
	documentExtractorProxy := mustNewProxy(
		getEnv("DOCUMENT_EXTRACTOR_URL", "http://localhost:8092"),
		"document-extractor",
	)
	api.POST("/ontologies/:id/documents/extract", withOntologyHeader(documentExtractorProxy))
	api.POST("/ontologies/:id/documents/extract/batch", withOntologyHeader(documentExtractorProxy))

	// Ontology template routes — proxied to ai-orchestration-service
	api.GET("/templates/ontologies", aiOrchProxy.HandleListTemplates)
	api.POST("/ontologies/:id/apply-template", aiOrchProxy.HandleApplyTemplate)

	// OpenAPI spec — served locally from embedded spec
	api.GET("/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", openAPISpec)
	})

	// Docs error route — placeholder for future UI
	api.GET("/docs", func(c *gin.Context) {
		slog.Info("docs.redirect", "trace_id", c.GetHeader("X-Trace-Id"))
		c.Redirect(http.StatusFound, "/openapi.json")
	})

	// NoRoute handler — uniform 404 for unknown /api/v1 routes.
	r.NoRoute(func(c *gin.Context) {
		slog.Warn("route.not_found",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"trace_id", c.GetHeader("X-Trace-Id"),
		)
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "GATEWAY-NOT-FOUND",
				Message: "The requested API endpoint does not exist.",
			},
		})
	})
}

// mustNewProxy creates a new proxy or panics on invalid upstream URL.
func mustNewProxy(upstreamURL, serviceName string) *proxy.Proxy {
	p, err := proxy.New(upstreamURL, 0, nil, serviceName)
	if err != nil {
		panic("invalid upstream URL for " + serviceName + ": " + err.Error())
	}
	return p
}

// getEnv returns the value of the environment variable or the fallback.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// withOntologyHeader wraps an http.Handler and injects X-Ontology-Id from
// the Gin path parameter :id before delegating to the underlying handler.
func withOntologyHeader(upstream http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ontologyID := c.Param("id")
		if ontologyID != "" {
			c.Request.Header.Set("X-Ontology-Id", ontologyID)
		}
		// Also propagate user identity and trace headers
		if uid := c.GetHeader("X-User-Id"); uid != "" {
			c.Request.Header.Set("X-User-Id", uid)
		}
		if traceID := c.GetHeader("X-Trace-Id"); traceID != "" {
			c.Request.Header.Set("X-Trace-Id", traceID)
		}
		upstream.ServeHTTP(c.Writer, c.Request)
	}
}
