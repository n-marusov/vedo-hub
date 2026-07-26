package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// requestCounter tracks total requests by service and method.
	// Auto-registered with the default Prometheus registry so promhttp.Handler()
	// serves it automatically.
	requestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vedo_service_requests_total",
			Help: "Total number of requests handled by the API Gateway",
		},
		[]string{"service", "method"},
	)
)

// Metrics returns a Gin middleware that increments the request counter
// with the service name and HTTP method labels.
func Metrics(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		requestCounter.WithLabelValues(serviceName, c.Request.Method).Inc()
	}
}
