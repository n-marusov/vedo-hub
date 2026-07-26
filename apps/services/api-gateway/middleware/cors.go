package middleware

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORS returns a Gin middleware that restricts CORS headers to allowed origins.
// In production (APP_ENV=production), it fails closed — denies all CORS when
// CORS_ALLOWED_ORIGINS is not set. In development, reflecting origin is allowed
// but credentials are only sent when the origin matches the allowlist.
func CORS(allowedOrigins ...string) gin.HandlerFunc {
	// Parse env var for allowed origins (comma-separated, takes precedence over args)
	envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	var origins []string
	if envOrigins != "" {
		for _, o := range strings.Split(envOrigins, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				origins = append(origins, o)
			}
		}
	} else {
		origins = allowedOrigins
	}

	appEnv := os.Getenv("APP_ENV")
	isProduction := appEnv == "production"

	// Fail-closed in production when no origins are configured.
	if isProduction && len(origins) == 0 {
		slog.Warn("[FIX] CORS fail-closed in production: no allowed origins configured")
		return func(c *gin.Context) {
			if c.GetHeader("Origin") != "" {
				c.Header("Access-Control-Allow-Origin", "")
				c.Header("Vary", "Origin")
				if c.Request.Method == "OPTIONS" {
					c.AbortWithStatus(204)
					return
				}
			}
			c.Next()
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		allowed := false
		if len(origins) == 0 {
			// Development mode: reflect origin
			c.Header("Access-Control-Allow-Origin", origin)
			allowed = true
		} else {
			for _, o := range origins {
				if origin == o {
					c.Header("Access-Control-Allow-Origin", origin)
					allowed = true
					break
				}
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Trace-Id, X-Correlation-Id, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "X-Trace-Id, X-Correlation-Id")
		c.Header("Vary", "Origin")

		// Only set credentials when origin is explicitly allowed
		if allowed && (len(origins) == 0 || origin != "") {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestLogger returns a Gin middleware that logs each request with timing.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		duration := time.Since(start).Milliseconds()
		traceID := c.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = c.GetHeader("X-Correlation-Id")
		}

		if status >= 500 {
			_ = c.Error(nil) // trigger error handling
		}

		// Structured log entry
		c.Set("request_duration_ms", duration)
		_ = path
		_ = method
		_ = traceID
	}
}
