package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
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

// contextKey is a private type for context keys to avoid collisions.
type contextKey string

const (
	contextKeyUserID contextKey = "user_id"
)

// KeyFunc defines the interface for JWT verification. Implementations can
// verify tokens against a JWKS endpoint (Keycloak) or a local PEM key.
// Structured for future pluggable JWT verification — Phase 2+.
type KeyFunc func(token string) (string, error)

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

// authMiddleware enforces that X-User-Id is provided for all /api/v1/ routes.
// In production the API Gateway has already verified the JWT and forwarded
// the user identity via X-User-Id. This middleware provides defense-in-depth
// by rejecting requests that lack the header unless AUTH_DISABLED=true.
//
// The KeyFunc parameter is reserved for future direct JWT verification
// against Keycloak JWKS (see api-gateway/auth/keyfunc.go for reference).
func authMiddleware(next http.Handler, authDisabled bool, _ KeyFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only enforce auth for API routes; health/metrics stay open.
		if !strings.HasPrefix(r.URL.Path, "/api/v1/") {
			next.ServeHTTP(w, r)
			return
		}

		if authDisabled {
			userID := r.Header.Get("X-User-Id")
			if userID == "" {
				userID = "anonymous"
			}
			slog.Debug("[FIX] Auth disabled - trusting X-User-Id header",
				"user_id", userID,
				"path", r.URL.Path,
			)
			ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		userID := r.Header.Get("X-User-Id")
		if userID == "" {
			slog.Warn("[FIX] Auth rejected - missing X-User-Id header",
				"path", r.URL.Path,
			)
			writeJSON(w, http.StatusUnauthorized, ErrorResponse{
				Error:   "UNAUTHORIZED",
				Message: "Missing authentication: X-User-Id header is required",
			})
			return
		}

		slog.Debug("[FIX] Auth verified",
			"user_id", userID,
			"path", r.URL.Path,
		)
		ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getUserID extracts the verified user ID from the request context.
// Falls back to the X-User-Id header for backwards compatibility
// (e.g., when proxied through a gateway that sets the header).
func getUserID(r *http.Request) string {
	if userID, ok := r.Context().Value(contextKeyUserID).(string); ok && userID != "" {
		return userID
	}
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		return "anonymous"
	}
	return userID
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

	// Auth configuration
	authDisabled := os.Getenv("AUTH_DISABLED") == "true"
	if authDisabled {
		slog.Warn("[FIX] AUTH_DISABLED=true — X-User-Id header will be trusted without verification. Do not use in production.")
	}

	// Structured JSON logger
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	slog.Info("Starting commenting-service",
		"port", port,
		"service", serviceName,
		"auth_disabled", authDisabled,
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

	// Wrapping middleware for request counting, logging, and auth
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

	// Apply auth middleware (defense-in-depth). KeyFunc is nil for now —
	// direct JWKS verification can be plugged in Phase 2+.
	handler := authMiddleware(wrapped, authDisabled, nil)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
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
