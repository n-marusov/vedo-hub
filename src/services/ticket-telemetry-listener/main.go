// @ctx: Gin router setup for ticket-telemetry-listener — automatic ticket creation from telemetry events
// @hlv:sec [INPUT_VALIDATION] — trace and correlation middleware

package main

import (
	"log/slog"
	"os"
	"time"

	"vedo-core/src/services/ticket-api/ticketapi"

	"github.com/gin-gonic/gin"
)

const serviceName = "ticket-telemetry-listener"
const defaultPort = "8090"
const defaultClassifierURL = "http://ticket-classifier:8000"
const appVersion = "1.5.0"

func main() {
	port := os.Getenv("STUB_PORT")
	if port == "" {
		port = defaultPort
	}

	classifierURL := os.Getenv("CLASSIFIER_URL")
	if classifierURL == "" {
		classifierURL = defaultClassifierURL
	}

	store := ticketapi.NewTicketStore()
	dedup := NewDedupStore()
	pyClient := NewPythonClassifierClient(classifierURL)
	goClass := NewGoFallbackClassifier()
	handler := NewAutoTicketHandler(store, dedup, pyClient, goClass, appVersion)

	router := gin.New()
	router.Use(TraceMiddleware())
	router.Use(RequestLogger())
	router.Use(gin.Recovery())

	router.POST("/tickets/automatic", handler.HandleAutoTicket)
	router.GET("/health", handler.HealthCheck)
	router.GET("/ready", handler.ReadyCheck)
	router.GET("/", handler.Metadata)

	slog.Info("server.starting",
		"service", serviceName,
		"port", port,
		"app_version", appVersion,
	)
	if err := router.Run(":" + port); err != nil {
		slog.Error("server.failed", "error", err.Error())
		os.Exit(1)
	}
}

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = ticketapi.NewUUID()
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-Id", traceID)
		c.Next()
	}
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetString("trace_id")
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		slog.Info("http.request",
			"trace_id", traceID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", duration.Milliseconds(),
		)
	}
}
