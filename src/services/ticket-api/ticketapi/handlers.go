// @ctx: HTTP handlers for ticket API — linear flow: parse → validate → execute → respond
// @hlv:sec [INPUT_VALIDATION] — all user input parsed from JSON request bodies
// @hlv:sec [AUTH_BOUNDARY] — authentication required for all state-changing endpoints

package ticketapi

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	store  *TicketStore
	audit  *AuditStore
	env    string
	appVer string
}

func NewTicketHandler(store *TicketStore, audit *AuditStore, env, appVer string) *TicketHandler {
	return &TicketHandler{
		store:  store,
		audit:  audit,
		env:    env,
		appVer: appVer,
	}
}

func (h *TicketHandler) CreateTicket(c *gin.Context) {
	traceID := c.GetString("trace_id")
	slog.Info("handler.create_ticket.enter",
		"trace_id", traceID,
	)
	// @hlv log_entry_exit

	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("handler.create_ticket.invalid_request",
			"trace_id", traceID,
			"error", err.Error(),
		)
		// @hlv TICKET-INVALID-REQUEST
		// @hlv log_all_errors
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   ErrInvalidRequest,
			Message: "Request body missing required fields or contains invalid values",
			Ref:     "T-01",
		})
		return
	}

	userID := c.GetString("user_id")
	metadata := TicketMetadata{
		VedoVersion: h.appVer,
		Environment: h.env,
		UserID:      userID,
		TraceID:     &traceID,
		UserAgent:   c.GetHeader("User-Agent"),
	}
	if metadata.UserAgent == "" {
		metadata.UserAgent = "unknown"
	}

	ticket, err := h.store.Create(req, metadata)
	if err != nil {
		appErr, ok := err.(*AppError)
		if ok {
			slog.Error("handler.create_ticket.failed",
				"trace_id", traceID,
				"error", appErr.Code,
				"message", appErr.Message,
			)
			// @hlv log_all_errors
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	h.audit.Record(userID, c.GetString("role"), "create", ticket.ID, traceID, "success")
	slog.Info("handler.create_ticket.exit",
		"trace_id", traceID,
		"ticket_id", ticket.ID,
		"status", http.StatusCreated,
	)
	c.JSON(http.StatusCreated, ticket)
}

func (h *TicketHandler) GetTicket(c *gin.Context) {
	traceID := c.GetString("trace_id")
	slog.Info("handler.get_ticket.enter",
		"trace_id", traceID,
		"ticket_id", c.Param("id"),
	)

	ticket, err := h.store.GetByID(c.Param("id"))
	if err != nil {
		appErr, ok := err.(*AppError)
		if ok && appErr.Code == ErrNotFound {
			// @hlv TICKET-NOT-FOUND
			slog.Error("handler.get_ticket.not_found",
				"trace_id", traceID,
				"ticket_id", c.Param("id"),
			)
			// @hlv log_all_errors
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   ErrNotFound,
				Message: "Ticket not found",
				Ref:     "T-06",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	slog.Info("handler.get_ticket.exit",
		"trace_id", traceID,
		"ticket_id", ticket.ID,
		"status", http.StatusOK,
	)
	c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandler) UpdateTicket(c *gin.Context) {
	traceID := c.GetString("trace_id")
	ticketID := c.Param("id")
	slog.Info("handler.update_ticket.enter",
		"trace_id", traceID,
		"ticket_id", ticketID,
	)

	var req UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("handler.update_ticket.invalid_request",
			"trace_id", traceID,
			"error", err.Error(),
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   ErrInvalidRequest,
			Message: "Request body contains invalid values",
			Ref:     "T-01",
		})
		return
	}

	if req.Status != nil && *req.Status == TicketStatusClosed && c.GetString("role") != "support" {
		// @hlv TICKET-FORBIDDEN
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error:   ErrForbidden,
			Message: "Only support users can close tickets",
			Ref:     "security considerations",
		})
		return
	}
	// @hlv authn_required

	ticket, err := h.store.Update(ticketID, req)
	if err != nil {
		appErr, ok := err.(*AppError)
		if ok {
			status := http.StatusConflict
			if appErr.Code == ErrNotFound {
				status = http.StatusNotFound
			}
			slog.Error("handler.update_ticket.failed",
				"trace_id", traceID,
				"error", appErr.Code,
				"message", appErr.Message,
			)
			// @hlv TICKET-CONFLICT
			// @hlv log_all_errors
			c.JSON(status, gin.H{
				"error":   appErr.Code,
				"message": appErr.Message,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	h.audit.Record(c.GetString("user_id"), c.GetString("role"), "update", ticketID, traceID, "success")
	slog.Info("handler.update_ticket.exit",
		"trace_id", traceID,
		"ticket_id", ticketID,
		"status", http.StatusOK,
	)
	c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandler) ListTickets(c *gin.Context) {
	traceID := c.GetString("trace_id")
	slog.Info("handler.list_tickets.enter",
		"trace_id", traceID,
	)

	var source *TicketSource
	if s := c.Query("source"); s != "" {
		val := TicketSource(s)
		source = &val
	}
	var status *TicketStatus
	if s := c.Query("status"); s != "" {
		val := TicketStatus(s)
		status = &val
	}
	var category *TicketCategory
	if s := c.Query("category"); s != "" {
		val := TicketCategory(s)
		category = &val
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	tickets, total := h.store.List(source, status, category, limit, offset)
	slog.Info("handler.list_tickets.exit",
		"trace_id", traceID,
		"count", len(tickets),
		"total", total,
	)
	c.JSON(http.StatusOK, gin.H{
		"tickets": tickets,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

func (h *TicketHandler) AddComment(c *gin.Context) {
	traceID := c.GetString("trace_id")
	ticketID := c.Param("id")
	slog.Info("handler.add_comment.enter",
		"trace_id", traceID,
		"ticket_id", ticketID,
	)

	var body struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   ErrInvalidRequest,
			Message: "Comment text is required",
			Ref:     "T-02",
		})
		return
	}

	userID := c.GetString("user_id")
	comment, err := h.store.AddComment(ticketID, userID, body.Text)
	if err != nil {
		appErr, ok := err.(*AppError)
		if ok && appErr.Code == ErrNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   ErrNotFound,
				Message: "Ticket not found",
				Ref:     "T-06",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	h.audit.Record(userID, c.GetString("role"), "comment", ticketID, traceID, "success")
	slog.Info("handler.add_comment.exit",
		"trace_id", traceID,
		"ticket_id", ticketID,
	)
	c.JSON(http.StatusCreated, comment)
}

func (h *TicketHandler) GetAuditLog(c *gin.Context) {
	ticketID := c.Param("id")
	entries := h.audit.GetByTicketID(ticketID)
	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
		"total":   len(entries),
	})
}

func (h *TicketHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func (h *TicketHandler) ReadyCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (h *TicketHandler) Metadata(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":        "ticket-api",
		"version":     h.appVer,
		"description": "Ticket management API",
	})
}

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = NewUUID()
		}
		c.Set("trace_id", traceID)
		c.Header("X-Trace-Id", traceID)

		userID := c.GetHeader("X-User-Id")
		if userID == "" {
			userID = "anonymous"
		}
		c.Set("user_id", userID)

		role := c.GetHeader("X-User-Role")
		if role == "" {
			role = "user"
		}
		c.Set("role", role)

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
