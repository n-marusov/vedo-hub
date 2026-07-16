package main

import (
	_ "embed"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"vedo-core/src/services/api-gateway/handlers"
	"vedo-core/src/services/api-gateway/middleware"
	"vedo-core/src/services/api-gateway/models"
	"vedo-core/src/services/api-gateway/proxy"
	"vedo-core/src/services/shared/llm"
)

//go:embed docs/openapi.json
var openAPISpec []byte

// RegisterRoutes defines all API route groups and wires them to the proxy handlers.
// The proxies are configured via environment variables pointing to upstream services.
//
// Gin's trie-based router forbids mixing wildcards (`/*path`) with named params
// (`:id`) or static segments at the same level. We therefore register every
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

	// Suppress unused variable warnings for future-use clients
	_ = versioningGrpc
	_ = authGrpc

	// API v1 routes — auth middleware applied at engine level (see main.go)
	api := r.Group("/api/v1")

	// Ontology REST read handlers — use gRPC (falling back to HTTP proxy
	// when server-side stubs are not yet fully migrated).
	ontologyHandler := handlers.NewOntologyHandler(ontologyProxy, ontologyGrpc)
	api.GET("/ontologies", ontologyHandler.HandleListOntologies)
	api.GET("/ontologies/:id", ontologyHandler.HandleGetOntology)
	api.GET("/ontologies/:id/classes", ontologyHandler.HandleListClasses)
	api.GET("/ontologies/:id/classes/:classId", ontologyHandler.HandleGetClass)
	api.GET("/ontologies/:id/properties", ontologyHandler.HandleListProperties)
	api.GET("/ontologies/:id/individuals", ontologyHandler.HandleListIndividuals)

	// Ontology write endpoints (POST/PUT/DELETE) are exposed by the
	// ontology-service itself. Proxy them by path so the gateway stays the
	// single facade the frontend talks to.
	// NOTE: These still use the HTTP reverse proxy. Migrating write endpoints
	// to gRPC will be done when server-side gRPC stubs are fully implemented.
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

	// Versioning service proxy — explicit endpoints mirroring the
	// versioning-service REST API (see src/services/versioning-service/src/routes.rs).
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

	// Query routes — SPARQL/CYPHER with read-only enforcement and rate limiting.
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

	// Initialize LLM provider and template engine for AI features.
	aiProvider, _ := initLLMProvider()
	templateEngine, _ := llm.NewTemplateEngine()
	if templateEngine != nil {
		avail := templateEngine.Available()
		slog.Info("llm.init.templates_loaded", "count", len(avail), "templates", avail)
	}

	// Create AI handler instances.
	nl2owlHandler := handlers.NewNlToOwlHandler(aiProvider, templateEngine)
	aiCompletionH := handlers.NewAiCompletionHandler(aiProvider, templateEngine, ontologyGrpc)
	refinementH := handlers.NewRefinementHandler(aiProvider, templateEngine)
	templateHandler := handlers.NewTemplateHandler(aiProvider, templateEngine, ontologyGrpc)

	// AI-related routes with LLM Policy Router and Prompt Injection Pre-filter middleware.
	// These routes control LLM access based on ontology visibility and deployment mode,
	// and pre-filter user-supplied text for prompt injection attempts.
	// The ontology gRPC client is used to resolve visibility server-side, preventing
	// clients from forging X-Ontology-Visibility.
	aiRoutes := api.Group("", middleware.LLMPolicyRouter(ontologyGrpc), middleware.PromptInjectionPreFilter())
	{
		// NL→OWL generation from text (Phase 4, Task 4.1)
		aiRoutes.POST("/ontologies/:id/generate-from-text", nl2owlHandler.HandleGenerateFromText)

		// AI-assisted completion (Phase 4, Task 4.2)
		aiRoutes.POST("/ontologies/:id/ai/suggest-classes", aiCompletionH.HandleSuggestClasses)
		aiRoutes.POST("/ontologies/:id/ai/suggest-properties", aiCompletionH.HandleSuggestProperties)
		aiRoutes.POST("/ontologies/:id/ai/suggest-relationships", aiCompletionH.HandleSuggestRelationships)
		// Legacy route — kept for backward compatibility
		aiRoutes.POST("/ontologies/:id/ai/complete", aiCompletionH.HandleSuggestClasses)

		// Iterative refinement (Phase 4, Task 4.3)
		aiRoutes.POST("/ontologies/:id/ai/refine", refinementH.HandleRefine)

		// Document extractor proxy (Phase 3/4)
		documentExtractorProxy := mustNewProxy(
			getEnv("DOCUMENT_EXTRACTOR_URL", "http://localhost:8092"),
			"document-extractor",
		)
		aiRoutes.POST("/ontologies/:id/documents/extract", gin.WrapH(documentExtractorProxy))
		aiRoutes.POST("/ontologies/:id/documents/extract/batch", gin.WrapH(documentExtractorProxy))
	}

	// Ontology template routes (Phase 4, Task 4.4)
	api.GET("/templates/ontologies", templateHandler.HandleListTemplates)
	api.POST("/ontologies/:id/apply-template", templateHandler.HandleApplyTemplate)

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

// initLLMProvider creates an LLM provider from environment configuration.
// Returns nil provider (with logged warning) on failure so the gateway can
// start without LLM features being available.
func initLLMProvider() (llm.Provider, error) {
	provider, err := llm.NewProvider()
	if err != nil {
		slog.Warn("llm.init.provider_unavailable",
			"error", err,
			"hint", "Set LLM_PROVIDER, LLM_API_KEY, and LLM_MODEL environment variables",
		)
		return nil, err
	}

	slog.Info("llm.init.provider_ready", "provider", getEnv("LLM_PROVIDER", "openai"))
	return provider, nil
}
