package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"vedo-core/src/services/api-gateway/handlers"
	"vedo-core/src/services/api-gateway/middleware"
	"vedo-core/src/services/api-gateway/models"
	"vedo-core/src/services/api-gateway/proxy"
)

//go:embed docs/openapi.json
var openAPISpec []byte

//go:embed swagger/*
var swaggerEmbedFS embed.FS

// swaggerFS is the pre-computed http.FileSystem rooted at swagger/.
// Initialised in init() via fs.Sub. Nil pointer = panic (programmer error).
var swaggerFS http.FileSystem

func init() {
	sub, err := fs.Sub(swaggerEmbedFS, "swagger")
	if err != nil {
		panic("swagger embed init: " + err.Error())
	}
	swaggerFS = http.FS(sub)
}

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

	_ = versioningGrpc
	_ = authGrpc

	// API v1 routes — auth middleware applied at engine level (see main.go)
	api := r.Group("/api/v1")

	// Org management routes — groups, projects, members, visibility, policies.
	orgClient := proxy.NewOrgServiceClient(grpcPool, proxy.GrpcAddrAuth)
	orgHandler := handlers.NewOrgHandler(orgClient)

	// Idempotency middleware for org write endpoints (REQ-FUN.API.write-idempotency).
	// Uses in-memory store by default; Redis backend when REDIS_URL is set.
	idemStore := middleware.NewMemIdempotencyStore()
	idemMiddleware := middleware.Idempotency(&middleware.IdempotencyConfig{
		Store:         idemStore,
		CriticalPaths: []string{"/api/v1/groups/", "/api/v1/projects/", "/api/v1/ontologies/"}, // org + draft write paths require Idempotency-Key
	})

	api.GET("/groups", orgHandler.HandleListGroups)
	api.GET("/groups/:id", orgHandler.HandleGetGroup)
	api.GET("/groups/:id/subgroups", orgHandler.HandleListChildGroups)
	api.GET("/groups/:id/members", orgHandler.HandleListMembers)

	api.GET("/projects", orgHandler.HandleListProjects)
	api.GET("/projects/:id", orgHandler.HandleGetProject)

	api.GET("/projects/:id/members", orgHandler.HandleListMembers)
	api.GET("/projects/:id/visibility", orgHandler.HandleGetVisibility)
	api.GET("/projects/:id/policies", orgHandler.HandleListPolicies)

	// Org write endpoints with idempotency middleware
	orgWrite := api.Group("")
	orgWrite.Use(idemMiddleware)

	orgWrite.POST("/groups", orgHandler.HandleCreateGroup)
	orgWrite.PUT("/groups/:id", orgHandler.HandleUpdateGroup)
	orgWrite.DELETE("/groups/:id", orgHandler.HandleDeleteGroup)

	orgWrite.POST("/projects", orgHandler.HandleCreateProject)
	orgWrite.PUT("/projects/:id", orgHandler.HandleUpdateProject)
	orgWrite.DELETE("/projects/:id", orgHandler.HandleDeleteProject)

	orgWrite.POST("/projects/:id/members", orgHandler.HandleAddMember)
	orgWrite.PUT("/projects/:id/members/:userId", orgHandler.HandleUpdateMemberRole)
	orgWrite.DELETE("/projects/:id/members/:userId", orgHandler.HandleRemoveMember)

	orgWrite.PUT("/projects/:id/visibility", orgHandler.HandleSetVisibility)
	orgWrite.POST("/projects/:id/policies", orgHandler.HandleCreatePolicy)
	orgWrite.DELETE("/projects/:id/policies/:policyId", orgHandler.HandleDeletePolicy)

	orgWrite.POST("/projects/:id/fork", orgHandler.HandleForkProject)
	orgWrite.PUT("/projects/:id/move", orgHandler.HandleMoveProject)
	// TODO (Task 3.1): after `buf generate` regenerates protobuf Go types for
	// MoveProjectRequest/Response, uncomment the gRPC path in HandleMoveProject
	// and wire the auth-service implementation.

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
	api.POST("/ontologies/:id/validate", gin.WrapH(ontologyProxy))

	// Draft-state coordination — requires Idempotency-Key (under orgWrite).
	orgWrite.PUT("/ontologies/:id/draft", gin.WrapH(ontologyProxy))

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

	// Metrics-service proxy — ontology metrics read endpoints (Python FastAPI, port 8084).
	metricsProxy := mustNewProxy(
		getEnv("METRICS_SERVICE_URL", "http://localhost:8084"),
		"metrics-service",
	)
	api.GET("/metrics/ontologies", gin.WrapH(metricsProxy))
	api.GET("/metrics/ontologies/:ontology_id", gin.WrapH(metricsProxy))

	// GraphQL endpoint — proxy to ontology-service
	api.Any("/graphql", gin.WrapH(ontologyProxy))

	// Commenting-service proxy — comment CRUD and feed endpoints.
	commentingProxy := mustNewProxy(
		getEnv("COMMENTING_SERVICE_URL", "http://localhost:8087"),
		"commenting-service",
	)
	api.GET("/ontologies/:id/comments", gin.WrapH(commentingProxy))
	api.POST("/ontologies/:id/comments", gin.WrapH(commentingProxy))

	// Create AI orchestration proxy — thin HTTP to gRPC bridge. All AI business
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

	// OpenAPI spec — served locally from embedded spec
	api.GET("/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", openAPISpec)
	})

	// Docs routes — Swagger UI (dev-only, controlled by ENABLE_SWAGGER_UI).
	// Per ADR-DES.API.swagger-ui-dev-only-strategy: enabled only in dev;
	// staging and production return 404.
	api.GET("/docs", docHandler)
	api.GET("/docs/*filepath", swaggerStaticHandler)

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

// isSwaggerUIEnabled checks ENABLE_SWAGGER_UI env var (default: false).
// Swagger UI is a dev-only feature per ADR-DES.API.swagger-ui-dev-only-strategy.
func isSwaggerUIEnabled() bool {
	v := os.Getenv("ENABLE_SWAGGER_UI")
	return v == "true" || v == "1"
}

// docHandler serves the Swagger UI index page when ENABLE_SWAGGER_UI=true,
// otherwise returns 404 with a descriptive error message.
func docHandler(c *gin.Context) {
	if !isSwaggerUIEnabled() {
		slog.Warn("swagger.ui.disabled",
			"trace_id", c.GetHeader("X-Trace-Id"),
			"hint", "set ENABLE_SWAGGER_UI=true to enable Swagger UI in dev",
		)
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "SWAGGER-UI-DISABLED",
				Message: "Swagger UI is not available in this environment.",
			},
		})
		return
	}
	slog.Info("swagger.ui.served", "path", "/api/v1/docs", "trace_id", c.GetHeader("X-Trace-Id"))
	c.FileFromFS("index.html", swaggerFS)
}

// swaggerStaticHandler serves embedded Swagger UI static assets (CSS/JS)
// when ENABLE_SWAGGER_UI=true. The Gin catch-all param :filepath
// captures the path after /docs, e.g. "/swagger-ui.css".
func swaggerStaticHandler(c *gin.Context) {
	if !isSwaggerUIEnabled() {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    "SWAGGER-UI-DISABLED",
				Message: "Swagger UI is not available in this environment.",
			},
		})
		return
	}
	file := c.Param("filepath")
	// Gin's catch-all param includes a leading "/" — strip it.
	if len(file) > 0 && file[0] == '/' {
		file = file[1:]
	}
	c.FileFromFS(file, swaggerFS)
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
