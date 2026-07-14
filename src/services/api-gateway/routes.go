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
// Gin's trie-based router forbids mixing wildcards (`/*path`) with named params
// (`:id`) or static segments at the same level. We therefore register every
// proxied endpoint explicitly instead of using a catch-all wildcard. Unknown
// `/api/v1/*` paths fall through to `r.NoRoute` (set up by the caller) which
// returns a uniform `GATEWAY-NOT-FOUND` error.
func RegisterRoutes(r *gin.Engine) {
	// Initialize upstream proxies
	ontologyProxy := mustNewProxy(
		getEnv("ONTOLOGY_SERVICE_URL", "http://localhost:8082"),
		"ontology-service",
	)
	versioningProxy := mustNewProxy(
		getEnv("VERSIONING_SERVICE_URL", "http://localhost:8083"),
		"versioning-service",
	)

	// API v1 routes — auth middleware applied at engine level (see main.go)
	api := r.Group("/api/v1")

	// Ontology REST read handlers — registered on the gateway itself so query
	// parameter validation, pagination normalization, and error formatting stay
	// consistent across the read surface.
	ontologyHandler := handlers.NewOntologyHandler(ontologyProxy)
	api.GET("/ontologies", ontologyHandler.HandleListOntologies)
	api.GET("/ontologies/:id", ontologyHandler.HandleGetOntology)
	api.GET("/ontologies/:id/classes", ontologyHandler.HandleListClasses)
	api.GET("/ontologies/:id/classes/:classId", ontologyHandler.HandleGetClass)
	api.GET("/ontologies/:id/properties", ontologyHandler.HandleListProperties)
	api.GET("/ontologies/:id/individuals", ontologyHandler.HandleListIndividuals)

	// Ontology write endpoints (POST/PUT/DELETE) are exposed by the
	// ontology-service itself. Proxy them by path so the gateway stays the
	// single facade the frontend talks to.
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
	// Max row limit defaults to 1000 and is configurable via QUERY_MAX_LIMIT.
	queryMaxLimit := 1000
	if v := getEnv("QUERY_MAX_LIMIT", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			queryMaxLimit = n
		}
	}
	queryHandler := handlers.NewQueryHandler(ontologyProxy, queryMaxLimit)
	api.POST("/sparql", queryHandler.HandleSPARQL)
	api.POST("/cypher", queryHandler.HandleCYPHER)

	// GraphQL endpoint — proxy to ontology-service
	api.Any("/graphql", gin.WrapH(ontologyProxy))

	// OpenAPI spec — served locally from embedded spec
	api.GET("/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", openAPISpec)
	})

	// Docs error route — placeholder for future UI
	api.GET("/docs", func(c *gin.Context) {
		slog.Info("docs.redirect", "trace_id", c.GetHeader("X-Trace-Id"))
		c.Redirect(http.StatusFound, "/openapi.json")
	})

	// NoRoute handler — uniform 404 for unknown /api/v1 routes. Using NoRoute
	// avoids the gin trie wildcard conflict that a catch-all `/*path` would
	// introduce at the same level as registered static segments.
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
