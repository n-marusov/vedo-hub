package main

import (
	"os"

	"github.com/gin-gonic/gin"

	"vedo-core/src/services/api-gateway/handlers"
	"vedo-core/src/services/api-gateway/proxy"
)

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

	// Ontology service proxy — all /api/v1/ontologies/* routes
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

	// OpenAPI spec — served by ontology service
	api.GET("/openapi.json", gin.WrapH(ontologyProxy))
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
