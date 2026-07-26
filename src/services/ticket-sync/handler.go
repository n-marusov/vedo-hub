// @ctx: HTTP handlers — POST /ticket-sync + POST /webhook/gitlab
// @hlv:sec [INPUT_VALIDATION] — all user input parsed from JSON request bodies
// @hlv:sec [AUTH_BOUNDARY] — authentication required for state-changing endpoints

package main

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type TicketSyncHandler struct {
	engine *SyncEngine
}

func NewTicketSyncHandler(engine *SyncEngine) *TicketSyncHandler {
	return &TicketSyncHandler{
		engine: engine,
	}
}

func (h *TicketSyncHandler) SyncTicket(c *gin.Context) {
	traceID := c.GetString("trace_id")
	slog.Info("handler.sync_ticket.enter",
		"trace_id", traceID,
	)
	// @hlv log_entry_exit

	// @hlv:sec [INPUT_VALIDATION] — parse and validate request body
	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("handler.sync_ticket.invalid_request",
			"trace_id", traceID,
			"error", err.Error(),
		)
		// @hlv SYNC-INVALID-REQUEST
		// @hlv log_all_errors
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   ErrInvalidRequest,
			Message: "Request body missing required fields or contains invalid values",
			Ref:     "T-19",
		})
		return
	}

	result := h.engine.Execute(req)

	status := h.resolveHTTPStatus(result)

	if result.Error != nil && *result.Error != "" {
		slog.Error("handler.sync_ticket.failed",
			"trace_id", traceID,
			"sync_id", result.SyncID,
			"sync_status", string(result.SyncStatus),
			"error", *result.Error,
		)
		c.JSON(status, ErrorResponse{
			Error:   *result.Error,
			Message: result.getErrorMessage(),
			Ref:     "T-18",
		})
		// @hlv log_all_errors
		return
	}

	slog.Info("handler.sync_ticket.exit",
		"trace_id", traceID,
		"sync_id", result.SyncID,
		"sync_status", string(result.SyncStatus),
	)
	c.JSON(status, result)
}

func (h *TicketSyncHandler) WebhookGitLab(c *gin.Context) {
	traceID := c.GetString("trace_id")
	slog.Info("handler.webhook_gitlab.enter",
		"trace_id", traceID,
	)
	// @hlv log_entry_exit

	slog.Info("handler.webhook_gitlab.accepted",
		"trace_id", traceID,
		"content_type", c.GetHeader("Content-Type"),
	)
	// @hlv log_external_calls

	c.JSON(http.StatusAccepted, gin.H{
		"status":  "accepted",
		"message": "Webhook received but not yet processed",
	})

	slog.Info("handler.webhook_gitlab.exit",
		"trace_id", traceID,
	)
}

// @hlv SYNC-INVALID-REQUEST -> 400, SYNC-UNAUTHORIZED -> 401,
// @hlv SYNC-GITLAB-UNAVAILABLE -> 503, SYNC-RATE-LIMITED -> 429,
// @hlv SYNC-CONFLICT -> 409, SYNC-PAYLOAD-TOO-LARGE -> 413
func (h *TicketSyncHandler) resolveHTTPStatus(resp SyncResponse) int {
	if resp.Error == nil || *resp.Error == "" {
		return http.StatusAccepted
	}

	switch *resp.Error {
	case ErrInvalidRequest:
		return http.StatusBadRequest
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrGitLabUnavailable:
		return http.StatusServiceUnavailable
	case ErrRateLimited:
		return http.StatusTooManyRequests
	case ErrConflict:
		return http.StatusConflict
	case ErrPayloadTooLarge:
		return http.StatusRequestEntityTooLarge
	default:
		if resp.SyncStatus == SyncStatusConflict {
			return http.StatusConflict
		}
		if resp.SyncStatus == SyncStatusFailed {
			return http.StatusServiceUnavailable
		}
		return http.StatusAccepted
	}
}

func (r SyncResponse) getErrorMessage() string {
	if r.Error == nil {
		return ""
	}
	switch *r.Error {
	case ErrInvalidRequest:
		return "Ticket ID must be a valid UUID"
	case ErrUnauthorized:
		return "Service lacks authorization for GitLab API access"
	case ErrGitLabUnavailable:
		return "GitLab API is currently unavailable"
	case ErrRateLimited:
		return "GitLab API rate limit exceeded"
	case ErrConflict:
		return "Synchronization conflict detected"
	case ErrPayloadTooLarge:
		return "Issue payload exceeds GitLab API limits"
	default:
		return *r.Error
	}
}

func NewUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = NewUUID()
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-Id", traceID)
		c.Next()
	}
	// @hlv request_correlation
	// @hlv no_sensitive_in_logs
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
		// @hlv log_external_calls
	}
}
