package main

import (
	"context"
	"crypto/tls"
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
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const serviceName = "commenting-service"
const defaultHTTPPort = "8084"
const defaultGRPCPort = "9004"

func getPort(envVar, fallback string) string {
	if p := os.Getenv(envVar); p != "" {
		return p
	}
	return fallback
}

func grpcServerOptions() []grpc.ServerOption {
	if os.Getenv("GRPC_TLS_ENABLED") != "true" {
		slog.Debug("grpc.tls.disabled")
		return nil
	}

	certFile := os.Getenv("GRPC_TLS_CERT_FILE")
	if certFile == "" {
		certFile = "/etc/vedo/tls/grpc-server.crt"
	}
	keyFile := os.Getenv("GRPC_TLS_KEY_FILE")
	if keyFile == "" {
		keyFile = "/etc/vedo/tls/grpc-server.key"
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		slog.Warn("grpc.tls.server_config_failed", "error", err)
		return nil
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.NoClientCert, // Phase 1: server-side TLS only
	})

	slog.Info("grpc.tls.server_enabled", "cert", certFile)
	return []grpc.ServerOption{grpc.Creds(creds)}
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
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        serviceName,
			"version":     "0.2.0",
			"description": "Commenting and collaboration service",
		})
	})

	// ---- gRPC Server ----
	grpcOpts := grpcServerOptions()
	grpcSrv := grpc.NewServer(grpcOpts...)
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcSrv, healthSrv)
	healthSrv.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
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

	slog.Info("commenting-service running", "http_port", httpPort, "grpc_port", grpcPort,
		"note", "gRPC CommentingService RPCs not yet registered; run 'make proto-generate' to enable")

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

// NOTE: After running `make proto-generate`, register CommentingServiceServer:
//
//	import commentingv1 "vedo-core/src/services/shared/proto/commenting/v1"
//	commentingv1.RegisterCommentingServiceServer(grpcSrv, &grpcServer.CommentingGrpcServer{})
