package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const serviceName = "auth-service"
const defaultHTTPPort = "8081"
const defaultGRPCPort = "9003"

func getPort(envVar, fallback string) string {
	if p := os.Getenv(envVar); p != "" {
		return p
	}
	return fallback
}

func main() {
	httpPort := getPort("SERVICE_PORT", defaultHTTPPort)
	grpcPort := getPort("GRPC_PORT", defaultGRPCPort)

	// ---- HTTP Server (health, ready, metrics) ----
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": serviceName})
	})
	r.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready", "service": serviceName})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// ---- gRPC Server (auth service) ----
	grpcSrv := grpc.NewServer()

	// Health check service for gRPC
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcSrv, healthSrv)
	healthSrv.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection for debugging
	reflection.Register(grpcSrv)

	// ---- Start both servers ----
	httpSrv := &http.Server{
		Addr:         ":" + httpPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	grpcListener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		slog.Error("gRPC listener failed", "error", err)
		panic(err)
	}

	go func() {
		slog.Info("HTTP server starting", "port", httpPort, "service", serviceName)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	go func() {
		slog.Info("gRPC server starting", "port", grpcPort, "service", serviceName)
		if err := grpcSrv.Serve(grpcListener); err != nil {
			panic(err)
		}
	}()

	slog.Info("auth-service running", "http_port", httpPort, "grpc_port", grpcPort,
		"note", "gRPC AuthService RPCs not yet registered; run 'make proto-generate' to enable")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("shutting down", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	grpcSrv.GracefulStop()
	_ = httpSrv.Shutdown(ctx)
	slog.Info("server stopped", "service", serviceName)
}

// NOTE: After running `make proto-generate`, register the AuthServiceServer:
//
//	import authv1 "vedo-core/src/services/shared/proto/auth/v1"
//	authv1.RegisterAuthServiceServer(grpcSrv, &grpcServer.AuthGrpcServer{})
