package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"vedo-core/src/services/api-gateway/auth"
	corsmw "vedo-core/src/services/api-gateway/middleware"
)

const serviceName = "api-gateway"
const defaultPort = "8080"

// getPort returns the port from SERVICE_PORT env var or default.
func getPort() string {
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = defaultPort
	}
	return port
}

// getUpstreamTimeout returns the timeout applied to proxied upstream requests.
// Reads UPSTREAM_TIMEOUT (seconds); falls back to 30s when unset or invalid.
// A non-positive value also falls back to the 30s default.
func getUpstreamTimeout() time.Duration {
	raw := os.Getenv("UPSTREAM_TIMEOUT")
	if raw == "" {
		return 30 * time.Second
	}
	secs, err := strconv.Atoi(raw)
	if err != nil || secs <= 0 {
		return 30 * time.Second
	}
	return time.Duration(secs) * time.Second
}

func main() {
	// Initialize Gin with default middleware (logging, recovery)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsmw.CORS())

	// Health and metrics endpoints (no auth)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": serviceName})
	})
	r.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready", "service": serviceName})
	})
	r.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, "# HELP vedo_service_requests_total Total service requests\n# TYPE vedo_service_requests_total counter\nvedo_service_requests_total{service=\"%s\"} 0\n", serviceName)
	})

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        serviceName,
			"version":     "0.2.0",
			"description": "API Gateway for VEDO Core",
		})
	})

	// Auth middleware configuration — exempt health/metrics/ready
	authConfig := &auth.Config{
		KeyFunc:          nil, // will be set when Keycloak integration is wired
		ExemptPrefixes:   append(auth.DefaultExemptPrefixes(), "/api/v1/public/"),
		ExactExemptPaths: auth.DefaultExactExemptPaths(),
		AuditWriter:      &auth.SlogAuditWriter{},
		AdminRoles:       auth.DefaultAdminRoles(),
	}

	// Apply auth middleware to all routes except exempt paths
	r.Use(auth.NewMiddleware(authConfig))

	// Timeout middleware for upstream requests (UPSTREAM_TIMEOUT seconds, 30s default)
	r.Use(corsmw.Timeout(getUpstreamTimeout()))

	// Register API route groups with proxy handlers
	RegisterRoutes(r)

	// Start server
	port := getPort()
	if err := r.Run(":" + port); err != nil {
		panic(err)
	}
}
