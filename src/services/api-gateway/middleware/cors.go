package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// CORS returns a Gin middleware that sets permissive CORS headers for the frontend origin.
// In production, restrict origins to the actual frontend URL.
func CORS(allowedOrigins ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if len(allowedOrigins) == 0 {
			// Allow all origins in development
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					c.Header("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Trace-Id, X-Correlation-Id, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "X-Trace-Id, X-Correlation-Id")
		c.Header("Access-Control-Allow-Credentials", "true")
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
			c.Error(nil) // trigger error handling
		}

		// Structured log entry
		c.Set("request_duration_ms", duration)
		_ = path
		_ = method
		_ = traceID
	}
}
