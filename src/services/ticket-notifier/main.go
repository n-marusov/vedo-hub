// @ctx: Gin service entry point for ticket-notifier — middleware, routes, startup

package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("STUB_PORT")
	if port == "" {
		port = "8091"
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(TraceMiddleware())
	r.Use(RequestLogger())

	handler := NewNotifierHandler()

	r.POST("/notifications", handler.HandleNotification)
	r.GET("/", handler.Metadata)
	r.GET("/health", handler.HealthCheck)
	r.GET("/ready", handler.ReadyCheck)

	slog.Info("server.starting",
		"service", "ticket-notifier",
		"port", port,
	)
	if err := r.Run(":" + port); err != nil {
		slog.Error("server.failed", "error", err)
		os.Exit(1)
	}
}
