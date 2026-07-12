// @ctx: ticket-api entry point — Gin router setup with middleware, handlers, and metrics
// @hlv:artifact ticket-api-code implements TICKET-CORE-001

package main

import (
	"log/slog"
	"os"

	"vedo-core/src/services/ticket-api/ticketapi"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("STUB_PORT")
	if port == "" {
		port = "8088"
	}
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	appVer := os.Getenv("APP_VERSION")
	if appVer == "" {
		appVer = "0.2.0"
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
	// @hlv structured_logging_only

	store := ticketapi.NewTicketStore()
	audit := ticketapi.NewAuditStore()
	handler := ticketapi.NewTicketHandler(store, audit, env, appVer)

	router := gin.New()
	router.Use(ticketapi.TraceMiddleware())
	router.Use(ticketapi.RequestLogger())
	router.Use(gin.Recovery())

	tickets := router.Group("/api/v1/tickets")
	{
		tickets.POST("", handler.CreateTicket)
		tickets.GET("", handler.ListTickets)
		tickets.GET("/:id", handler.GetTicket)
		tickets.PATCH("/:id", handler.UpdateTicket)
		tickets.POST("/:id/comments", handler.AddComment)
		tickets.GET("/:id/audit", handler.GetAuditLog)
	}

	router.GET("/health", handler.HealthCheck)
	router.GET("/ready", handler.ReadyCheck)
	router.GET("/", handler.Metadata)

	slog.Info("ticket-api.starting",
		"port", port,
		"env", env,
		"version", appVer,
	)

	if err := router.Run(":" + port); err != nil {
		slog.Error("ticket-api.failed", "error", err.Error())
		os.Exit(1)
	}
}
