package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"vedo-core/src/services/api-gateway/auth"
	corsmw "vedo-core/src/services/api-gateway/middleware"
	"vedo-core/src/services/api-gateway/proxy"
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

// overrideSummary returns a stable, log-friendly string representation of the
// RequiredRoleLevel map. Empty maps produce "<none>" so operators can spot
// the dangerous production misconfiguration at a glance.
func overrideSummary(overrides map[string]int) string {
	if len(overrides) == 0 {
		return "<none>"
	}
	pairs := make([]string, 0, len(overrides))
	for k, v := range overrides {
		pairs = append(pairs, fmt.Sprintf("%s=%d", k, v))
	}
	return strings.Join(pairs, ",")
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
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Metrics middleware — counts every request after health/metrics/ready
	r.Use(corsmw.Metrics(serviceName))

	// Root endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        serviceName,
			"version":     "0.2.0",
			"description": "API Gateway for VEDO Core",
		})
	})

	// Auth middleware configuration — exempt health/metrics/ready.
	// KeyFunc is loaded once at startup from JWT_PUBLIC_KEY_PATH or
	// KEYCLOAK_JWKS_URL. The gateway refuses to boot when JWT verification
	// is not configured so the auth boundary never degrades to accept-all.
	keyFunc, err := auth.LoadKeyFuncFromEnv()
	if err != nil {
		slog.Error("auth.keyfunc.missing",
			"error", err,
			"hint", "set JWT_PUBLIC_KEY_PATH or KEYCLOAK_JWKS_URL",
		)
		panic("JWT verification not configured: " + err.Error())
	}

	requiredRoleLevel := map[string]int{
		// Read-only query endpoints accept Viewer-role (level 0) auth so the
		// rate limiter — not the BFLA gate — returns 429 when the tier quota
		// is exceeded. Mutation keywords are still rejected by the gateway
		// query validator before the request reaches the upstream.
		"POST:/api/v1/sparql":  0,
		"POST:/api/v1/cypher":  0,
		"POST:/api/v1/graphql": 0,
	}
	slog.Info("auth.config.init",
		"role_override_count", len(requiredRoleLevel),
		"endpoints", overrideSummary(requiredRoleLevel),
	)

	authConfig := &auth.Config{
		KeyFunc:           keyFunc,
		ExemptPrefixes:    append(auth.DefaultExemptPrefixes(), "/api/v1/public/"),
		ExactExemptPaths:  auth.DefaultExactExemptPaths(),
		AuditWriter:       &auth.SlogAuditWriter{},
		AdminRoles:        auth.DefaultAdminRoles(),
		RequiredRoleLevel: requiredRoleLevel,
	}

	// Apply auth middleware to all routes except exempt paths
	r.Use(auth.NewMiddleware(authConfig))

	// Timeout middleware for upstream requests (UPSTREAM_TIMEOUT seconds, 30s default)
	r.Use(corsmw.Timeout(getUpstreamTimeout()))

	// Initialize gRPC client pool for internal service communication.
	grpcPool := proxy.NewGrpcClientPool(getUpstreamTimeout())

	// Register API route groups with proxy handlers
	RegisterRoutes(r, grpcPool)

	// Start server with timeouts
	port := getPort()
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server.starting", "port", port, "service", serviceName)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	sig := <-quit
	slog.Info("server.shutting_down", "signal", sig.String(), "service", serviceName)

	// Close gRPC connections before HTTP server shutdown
	grpcPool.Close()
	slog.Info("grpc.pool.closed", "service", serviceName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server.shutdown_error", "error", err, "service", serviceName)
		panic(err)
	}
	slog.Info("server.stopped", "service", serviceName)
}
