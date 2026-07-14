package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	serviceName  = "commenting-service"
	defaultPort  = "8085"
	defaultDBURL = "postgres://postgres:password@localhost:5432/vedo_comments?sslmode=disable"
)

var requestTotal atomic.Uint64

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeMetrics(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w,
		"# HELP vedo_service_requests_total Total service requests\n"+
			"# TYPE vedo_service_requests_total counter\n"+
			"vedo_service_requests_total{service=%q} %d\n",
		serviceName, requestTotal.Load(),
	)
}

func main() {
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = defaultPort
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = os.Getenv("COMMENTING_DATABASE_URL")
	}
	if databaseURL == "" {
		databaseURL = defaultDBURL
	}

	// Structured JSON logger
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	slog.Info("Starting commenting-service",
		"port", port,
		"service", serviceName,
	)

	// Initialize store
	store, err := NewCommentStore(databaseURL)
	if err != nil {
		slog.Error("Failed to initialize store", "error", err)
		// Start without DB for development resilience
		slog.Warn("Running without database — only health/metrics endpoints available")
		store = nil
	}

	mux := http.NewServeMux()

	// Health and information endpoints
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		requestTotal.Add(1)
		writeJSON(w, http.StatusOK, map[string]any{
			"name":        serviceName,
			"version":     "0.3.0",
			"description": "Commenting service — CRUD for entity-level comments, threads, feeds",
			"stub":        false,
		})
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		requestTotal.Add(1)
		dbStatus := "disabled"
		if store != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := store.db.PingContext(ctx); err == nil {
				dbStatus = "connected"
			} else {
				dbStatus = "disconnected"
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":   "healthy",
			"service":  serviceName,
			"database": dbStatus,
		})
	})

	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		requestTotal.Add(1)
		dbReady := false
		if store != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			dbReady = store.db.PingContext(ctx) == nil
		}
		if store != nil && !dbReady {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{
				"status":  "not ready",
				"service": serviceName,
				"reason":  "database not reachable",
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "ready",
			"service": serviceName,
		})
	})

	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		requestTotal.Add(1)
		writeMetrics(w)
	})

	// Register comment CRUD routes if store is available
	if store != nil {
		eventBus := newEventBus()
		handlers := newCommentHandlers(store, eventBus)
		registerCommentRoutes(mux, handlers)
		slog.Info("Comment CRUD routes registered")
	} else {
		// Return 503 for comment endpoints when no DB
		mux.HandleFunc("POST /api/v1/", dbUnavailableHandler)
		mux.HandleFunc("GET /api/v1/", dbUnavailableHandler)
		mux.HandleFunc("PUT /api/v1/", dbUnavailableHandler)
		mux.HandleFunc("DELETE /api/v1/", dbUnavailableHandler)
	}

	// Wrapping middleware for request counting and logging
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestTotal.Add(1)
		slog.Debug("Request received",
			"method", r.Method,
			"path", r.URL.Path,
			"trace_id", r.Header.Get("X-Trace-Id"),
			"correlation_id", r.Header.Get("X-Correlation-Id"),
		)
		mux.ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      wrapped,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("Listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	if store != nil {
		if err := store.Close(); err != nil {
			slog.Error("Failed to close database", "error", err)
		}
	}

	slog.Info("Server exited")
}

func dbUnavailableHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{
		Error:   "DATABASE_UNAVAILABLE",
		Message: "Commenting service is running without database connection. API endpoints are unavailable.",
	})
}
