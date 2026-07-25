package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"vedo-core/src/services/shared/llm"
	ai_orchestrationv1 "vedo-core/src/services/shared/proto/ai-orchestration/v1"
)

const serviceName = "ai-orchestration-service"

// Default ports — overridden by GRPC_PORT and HEALTH_PORT env vars.
const (
	defaultGrpcPort   = "9014"
	defaultHealthPort = "8093"
)

// getEnv returns the value of the environment variable or the fallback.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getPort returns the port string (with colon prefix) from env or default.
func getPort(envKey, defaultVal string) string {
	port := os.Getenv(envKey)
	if port == "" {
		port = defaultVal
	}
	return ":" + port
}

// overrideSummary returns a stable, log-friendly string representation of a
// string map. Empty maps produce "<none>".
func overrideSummary(m map[string]string) string {
	if len(m) == 0 {
		return "<none>"
	}
	pairs := make([]string, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(pairs, ",")
}

func main() {
	// Structured JSON logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Parse ports
	grpcPort := getPort("GRPC_PORT", defaultGrpcPort)
	healthPort := getPort("HEALTH_PORT", defaultHealthPort)

	slog.Info("server.starting",
		"service", serviceName,
		"grpc_port", grpcPort,
		"health_port", healthPort,
	)

	// Initialize LLM provider from environment config.
	// The service can start without an LLM provider — AI features will
	// return an LLM_UNAVAILABLE error until one is configured.
	aiProvider, err := llm.NewProvider()
	if err != nil {
		slog.Warn("llm.init.provider_unavailable",
			"error", err,
			"hint", "Set LLM_PROVIDER, LLM_API_KEY, and LLM_MODEL environment variables",
		)
	}

	// Initialize template engine for prompt rendering.
	templateEngine, err := llm.NewTemplateEngine()
	if err != nil {
		slog.Warn("llm.init.template_engine_unavailable",
			"error", err,
		)
	} else {
		avail := templateEngine.Available()
		slog.Info("llm.init.templates_loaded", "count", len(avail), "templates", avail)
	}

	// Create the gRPC AI orchestration server.
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryLoggingInterceptor()),
		grpc.StreamInterceptor(streamLoggingInterceptor()),
	)

	// Register the AI orchestration service implementation.
	aiSvc := NewAIOrchestrationService(aiProvider, templateEngine)
	ai_orchestrationv1.RegisterAIOrchestrationServiceServer(grpcServer, aiSvc)

	// Register gRPC health service for Kubernetes / Docker Compose probes.
	healthSvc := &healthServer{}
	grpc_health_v1.RegisterHealthServer(grpcServer, healthSvc)

	// Enable gRPC reflection for debugging (grpcurl, etc.).
	reflection.Register(grpcServer)

	// Start gRPC listener.
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		slog.Error("server.grpc_listen_failed", "error", err, "port", grpcPort)
		panic(err)
	}

	go func() {
		slog.Info("server.grpc_serving", "address", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("server.grpc_serve_error", "error", err)
			panic(err)
		}
	}()

	// Start HTTP health/metrics server.
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{"status":"healthy","service":"%s"}`, serviceName)
	})
	healthMux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `{"status":"ready","service":"%s"}`, serviceName)
	})
	healthMux.Handle("/metrics", promhttp.Handler())

	healthSrv := &http.Server{
		Addr:         healthPort,
		Handler:      healthMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		slog.Info("server.health_serving", "address", healthPort)
		if err := healthSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server.health_listen_error", "error", err)
			panic(err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("server.shutting_down", "signal", sig.String(), "service", serviceName)

	// Shutdown gRPC server (graceful stop with drain)
	grpcServer.GracefulStop()
	slog.Info("server.grpc_stopped", "service", serviceName)

	// Shutdown HTTP health server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := healthSrv.Shutdown(ctx); err != nil {
		slog.Error("server.health_shutdown_error", "error", err, "service", serviceName)
	}
	slog.Info("server.stopped", "service", serviceName)
}

// healthServer implements the gRPC Health/V1 service for liveness probes.
type healthServer struct {
	grpc_health_v1.UnimplementedHealthServer
}

func (s *healthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (s *healthServer) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return stream.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
}

// unaryLoggingInterceptor logs all unary gRPC calls with trace context.
func unaryLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		slog.Debug("grpc.unary_call", "method", info.FullMethod)
		resp, err := handler(ctx, req)
		duration := time.Since(start)
		if err != nil {
			slog.Warn("grpc.unary_error",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
				"error", err,
			)
		} else {
			slog.Debug("grpc.unary_ok",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
			)
		}
		return resp, err
	}
}

// streamLoggingInterceptor logs all streaming gRPC calls.
func streamLoggingInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		slog.Debug("grpc.stream_start", "method", info.FullMethod)
		err := handler(srv, stream)
		duration := time.Since(start)
		if err != nil {
			slog.Warn("grpc.stream_error",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
				"error", err,
			)
		} else {
			slog.Debug("grpc.stream_end",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
			)
		}
		return err
	}
}
