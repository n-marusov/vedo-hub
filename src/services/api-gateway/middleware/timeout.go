package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout returns a Gin middleware that sets a context deadline on the request.
// If the upstream takes longer than the timeout, the proxy transport will
// abort and the proxy ErrorHandler will return 504 Gateway Timeout.
// Default is 30s.
func Timeout(timeout time.Duration) gin.HandlerFunc {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		// Log if the context was exceeded (response already handled by proxy
		// ErrorHandler which returns 504 via GATEWAY-UPSTREAM-ERROR).
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			slog.Warn("middleware.timeout",
				"path", c.Request.URL.Path,
				"method", c.Request.Method,
				"timeout", timeout.String(),
				"trace_id", c.GetHeader("X-Trace-Id"),
			)
			// If the proxy error handler hasn't written a response yet
			// (shouldn't happen, but guard anyway), write one here.
			if !c.IsAborted() {
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
					"error": gin.H{
						"code":    "GATEWAY-TIMEOUT",
						"message": "Request timed out while proxying to upstream service",
					},
				})
			}
		}
	}
}
