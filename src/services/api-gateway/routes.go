package main

import (
	_ "embed"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"vedo-core/src/services/api-gateway/handlers"
	"vedo-core/src/services/api-gateway/models"
	"vedo-core/src/services/api-gateway/proxy"
)

//go:embed docs/openapi.json
var openAPISpec []byte

// RegisterRoutes defines all API route groups and wires them to the proxy handlers.
// The proxies are configured via environment variables pointing to upstream services.
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

	// Ontology REST read handlers — registered before the catch-all proxy
	// so Gin matches these specific routes first.
	ontologyHandler := handlers.NewOntologyHandler(ontologyProxy)

	// Specific REST read endpoints
	api.GET("/ontologies", ontologyHandler.HandleListOntologies)
	api.GET("/ontologies/:id", ontologyHandler.HandleGetOntology)
	api.GET("/ontologies/:id/classes", ontologyHandler.HandleListClasses)
	api.GET("/ontologies/:id/classes/:classId", ontologyHandler.HandleGetClass)
	api.GET("/ontologies/:id/properties", ontologyHandler.HandleListProperties)
	api.GET("/ontologies/:id/individuals", ontologyHandler.HandleListIndividuals)

	// Ontology service proxy — catch-all for routes not matched above
	ontology := api.Group("/ontologies")
	{
		ontology.Any("/*path", gin.WrapH(ontologyProxy))
		ontology.Any("", gin.WrapH(ontologyProxy))
	}

	// Versioning service proxy — all /api/v1/versioning/* routes
	versioning := api.Group("/versioning")
	{
		versioning.Any("/*path", gin.WrapH(versioningProxy))
		versioning.Any("", gin.WrapH(versioningProxy))
	}

	// Query routes — SPARQL/CYPHER with read-only enforcement and rate limiting
	queryHandler := handlers.NewQueryHandler(ontologyProxy, 1000)
	api.POST("/sparql", queryHandler.HandleSPARQL)
	api.POST("/cypher", queryHandler.HandleCYPHER)

	// OpenAPI spec — served locally from embedded spec
	api.GET("/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json", openAPISpec)
	})

	// Docs error route — placeholder for future UI
	api.GET("/docs", func(c *gin.Context) {
		slog.Info("docs.redirect", "trace_id", c.GetHeader("X-Trace-Id"))
		c.Redirect(http.StatusFound, "/openapi.json")
	})

	// Not found handler for unknown /api/v1 routes — returns consistent error format
	api.Any("/*path", func(c *gin.Context) {
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
