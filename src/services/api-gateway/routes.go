package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes defines all API route groups.
// Actual proxy handlers are implemented in Phase 6 (Task 6.1).
// These placeholder handlers return 501 Not Implemented.
func RegisterRoutes(r *gin.Engine) {
	// API v1 routes — auth middleware applied at engine level (see main.go)
	api := r.Group("/api/v1")

	// Ontology service proxy routes
	ontology := api.Group("/ontologies")
	{
		ontology.POST("", notImplemented("POST /api/v1/ontologies"))
		ontology.GET("", notImplemented("GET /api/v1/ontologies"))
		ontology.GET("/:id", notImplemented("GET /api/v1/ontologies/:id"))

		// Classes
		ontology.POST("/:id/classes", notImplemented("POST /api/v1/ontologies/:id/classes"))
		ontology.GET("/:id/classes", notImplemented("GET /api/v1/ontologies/:id/classes"))
		ontology.GET("/:id/classes/:classId", notImplemented("GET /api/v1/ontologies/:id/classes/:classId"))
		ontology.PUT("/:id/classes/:classId", notImplemented("PUT /api/v1/ontologies/:id/classes/:classId"))
		ontology.DELETE("/:id/classes/:classId", notImplemented("DELETE /api/v1/ontologies/:id/classes/:classId"))

		// Properties
		ontology.POST("/:id/properties", notImplemented("POST /api/v1/ontologies/:id/properties"))
		ontology.GET("/:id/properties", notImplemented("GET /api/v1/ontologies/:id/properties"))
		ontology.PUT("/:id/properties/:propId", notImplemented("PUT /api/v1/ontologies/:id/properties/:propId"))
		ontology.DELETE("/:id/properties/:propId", notImplemented("DELETE /api/v1/ontologies/:id/properties/:propId"))

		// Individuals
		ontology.POST("/:id/individuals", notImplemented("POST /api/v1/ontologies/:id/individuals"))
		ontology.GET("/:id/individuals", notImplemented("GET /api/v1/ontologies/:id/individuals"))
		ontology.PUT("/:id/individuals/:indivId", notImplemented("PUT /api/v1/ontologies/:id/individuals/:indivId"))
		ontology.DELETE("/:id/individuals/:indivId", notImplemented("DELETE /api/v1/ontologies/:id/individuals/:indivId"))

		// Import/Export
		ontology.POST("/:id/import", notImplemented("POST /api/v1/ontologies/:id/import"))
		ontology.GET("/:id/export", notImplemented("GET /api/v1/ontologies/:id/export"))
	}

	// Versioning service proxy routes
	versioning := api.Group("/versioning")
	{
		versioning.POST("/:ontologyId/branches", notImplemented("POST /api/v1/versioning/:ontologyId/branches"))
		versioning.GET("/:ontologyId/branches", notImplemented("GET /api/v1/versioning/:ontologyId/branches"))
		versioning.DELETE("/:ontologyId/branches/:branchId", notImplemented("DELETE /api/v1/versioning/:ontologyId/branches/:branchId"))

		versioning.POST("/:ontologyId/commits", notImplemented("POST /api/v1/versioning/:ontologyId/commits"))
		versioning.GET("/:ontologyId/commits", notImplemented("GET /api/v1/versioning/:ontologyId/commits"))
		versioning.GET("/:ontologyId/commits/:commitId", notImplemented("GET /api/v1/versioning/:ontologyId/commits/:commitId"))

		versioning.POST("/:ontologyId/checkout", notImplemented("POST /api/v1/versioning/:ontologyId/checkout"))
		versioning.POST("/:ontologyId/rollback", notImplemented("POST /api/v1/versioning/:ontologyId/rollback"))
	}

	// Query routes
	api.POST("/sparql", notImplemented("POST /api/v1/sparql"))
	api.POST("/cypher", notImplemented("POST /api/v1/cypher"))

	// OpenAPI spec
	api.GET("/openapi.json", notImplemented("GET /api/v1/openapi.json"))
}

// notImplemented returns a Gin handler that returns 501 Not Implemented.
func notImplemented(route string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": gin.H{
				"code":    "GATEWAY-NOT-IMPLEMENTED",
				"message": "Route " + route + " is not yet implemented",
			},
		})
	}
}
