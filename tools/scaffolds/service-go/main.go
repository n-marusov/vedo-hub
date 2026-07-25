package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	requestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vedo_requests_total",
		Help: "Total requests",
	}, []string{"method", "path", "status"})
	requestErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vedo_request_errors_total",
		Help: "Total failed requests",
	}, []string{"method", "path"})
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "vedo_request_duration_seconds",
		Help:    "Request duration seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})
)

func initTracer(ctx context.Context, endpoint string) func(context.Context) error {
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(strings.TrimPrefix(endpoint, "http://")), otlptracegrpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("service-go-template"),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return tp.Shutdown
}

func redact(value string) string {
	replacer := strings.NewReplacer("password", "[REDACTED]", "token", "[REDACTED]", "secret", "[REDACTED]")
	return replacer.Replace(value)
}

func traceID(ctx context.Context) string {
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		return spanContext.TraceID().String()
	}
	return "00000000000000000000000000000000"
}

func main() {
	ctx := context.Background()
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://otel-collector:4317"
	}
	shutdown := initTracer(ctx, endpoint)
	defer shutdown(ctx)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	r := gin.New()
	r.Use(otelgin.Middleware("service-go-template"))
	r.Use(func(c *gin.Context) {
		start := time.Now()
		correlationID := c.GetHeader("x-correlation-id")
		if correlationID == "" {
			correlationID = "generated-correlation-id"
		}
		c.Next()
		status := c.Writer.Status()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		requestTotal.WithLabelValues(c.Request.Method, path, http.StatusText(status)).Inc()
		requestDuration.WithLabelValues(c.Request.Method, path).Observe(time.Since(start).Seconds())
		if status >= 400 {
			requestErrors.WithLabelValues(c.Request.Method, path).Inc()
		}
		logger.Info(redact("request_complete"),
			"path", path,
			"method", c.Request.Method,
			"status", status,
			"trace_id", traceID(c.Request.Context()),
			"correlation_id", correlationID,
		)
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"name": "service-go-template", "version": "0.1.0", "description": "Go service template with observability", "stub": true})
	})
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "healthy"}) })
	r.GET("/ready", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ready"}) })
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
