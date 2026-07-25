// @hlv:artifact ticket-sync-server implements TICKET-SYNC-001
// @ctx: Gin router setup and main entry point for ticket-sync service

package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	router := setupRouter()
	slog.Info("server.starting", "port", "8086")
	if err := router.Run(":8086"); err != nil {
		slog.Error("server.failed", "error", err.Error())
		os.Exit(1)
	}
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(TraceMiddleware())
	router.Use(RequestLogger())

	client := NewStubGitLabClient()
	engine := NewSyncEngine(client)
	handler := NewTicketSyncHandler(engine)

	syncGroup := router.Group("/")
	{
		syncGroup.POST("ticket-sync", handler.SyncTicket)
	}

	webhookGroup := router.Group("/webhook")
	webhookGroup.Use(WebhookHMACMiddleware())
	{
		webhookGroup.POST("gitlab", handler.WebhookGitLab)
	}

	return router
}
