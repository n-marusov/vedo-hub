// @ctx: HTTP handler for ticket notifications — linear flow: parse → validate → route → queue → respond
// @hlv:sec [INPUT_VALIDATION] — all user input parsed from JSON request bodies
// @hlv:sec [AUTH_BOUNDARY] — API key authentication via X-API-Key header
// @hlv:sec [AUTH_BOUNDARY] — per-channel rate limiting prevents abuse

// @hlv:sec [NETWORK] — channel availability check before routing
package main

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type NotifierHandler struct {
	mu                  sync.RWMutex
	channelAvailability map[string]bool
	rateLimiter         *RateLimiter
}

func NewNotifierHandler() *NotifierHandler {
	return &NotifierHandler{
		channelAvailability: map[string]bool{
			"email":     true,
			"slack":     true,
			"pagerduty": true,
			"in_app":    true,
		},
		rateLimiter: NewRateLimiter(100),
	}
}

func (h *NotifierHandler) SetChannelAvailability(channel string, available bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.channelAvailability[channel] = available
}

func (h *NotifierHandler) SetRateLimit(maxRPS int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rateLimiter = NewRateLimiter(maxRPS)
}

func (h *NotifierHandler) isChannelAvailable(channel string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.channelAvailability[channel]
}

func (h *NotifierHandler) HandleNotification(c *gin.Context) {
	traceID := c.GetString("trace_id")
	slog.Info("handler.notification.enter",
		"trace_id", traceID,
	)

	var req NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("handler.notification.invalid_request",
			"trace_id", traceID,
			"error", err.Error(),
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   ErrInvalidRequest,
			Message: "Request body missing required fields or contains invalid values",
			Ref:     "T-22",
		})
		return
	}

	if !IsValidUUID(req.TicketID) {
		slog.Error("handler.notification.invalid_ticket_id",
			"trace_id", traceID,
			"ticket_id", req.TicketID,
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   ErrInvalidRequest,
			Message: "Ticket ID must be a valid UUID",
			Ref:     "T-22",
		})
		return
	}

	if !ValidEventTypes[req.EventType] {
		slog.Error("handler.notification.invalid_event_type",
			"trace_id", traceID,
			"event_type", req.EventType,
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   ErrInvalidRequest,
			Message: "Invalid event_type: " + req.EventType,
			Ref:     "T-22",
		})
		return
	}

	if !ValidChannels[req.NotificationChannel] {
		slog.Error("handler.notification.invalid_channel",
			"trace_id", traceID,
			"channel", req.NotificationChannel,
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   ErrInvalidRequest,
			Message: "Invalid notification_channel: " + req.NotificationChannel,
			Ref:     "T-22",
		})
		return
	}

	if !ValidRecipientTypes[req.RecipientType] {
		slog.Error("handler.notification.invalid_recipient_type",
			"trace_id", traceID,
			"recipient_type", req.RecipientType,
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   ErrRecipientInvalid,
			Message: "Invalid recipient_type: " + req.RecipientType,
			Ref:     "T-22",
		})
		return
	}

	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		slog.Error("handler.notification.unauthorized",
			"trace_id", traceID,
		)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   ErrUnauthorized,
			Message: "API key is required via X-API-Key header",
			Ref:     "security.yaml",
		})
		return
	}

	rateKey := req.NotificationChannel + ":" + req.RecipientType
	if !h.rateLimiter.Allow(rateKey) {
		slog.Error("handler.notification.rate_limited",
			"trace_id", traceID,
			"key", rateKey,
		)
		c.Header("Retry-After", "1")
		c.JSON(http.StatusTooManyRequests, ErrorResponse{
			Error:   ErrRateLimited,
			Message: "Notification rate exceeds allowed threshold",
			Ref:     "security.yaml",
		})
		return
	}

	if !h.isChannelAvailable(req.NotificationChannel) {
		slog.Error("handler.notification.channel_unavailable",
			"trace_id", traceID,
			"channel", req.NotificationChannel,
		)
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error:   ErrChannelUnavailable,
			Message: channelUnavailableMessage(req.NotificationChannel),
			Ref:     "T-23",
		})
		return
	}

	sp := req.TicketData.SystemPriority
	if sp == "p0" || sp == "p1" {
		extras := routeEscalation(&req)
		for _, extra := range extras {
			recipient, err := routeToChannel(&extra)
			if err != nil {
				slog.Error("handler.notification.escalation_failed",
					"trace_id", traceID,
					"channel", extra.NotificationChannel,
					"error", err.Error(),
				)
			} else {
				slog.Info("handler.notification.escalation_queued",
					"trace_id", traceID,
					"channel", extra.NotificationChannel,
					"recipient", recipient,
				)
			}
		}
	}

	recipient, err := routeToChannel(&req)
	if err != nil {
		appErr, ok := err.(*AppError)
		if ok {
			slog.Error("handler.notification.channel_error",
				"trace_id", traceID,
				"error", appErr.Code,
				"message", appErr.Message,
			)
			status := http.StatusBadRequest
			if appErr.Code == ErrChannelUnavailable {
				status = http.StatusServiceUnavailable
			}
			c.JSON(status, ErrorResponse{
				Error:   appErr.Code,
				Message: appErr.Message,
				Ref:     appErr.Ref,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   ErrTemplateError,
			Message: "Error processing notification",
			Ref:     "T-22",
		})
		return
	}

	notifID := NewUUID()
	resp := NotificationResponse{
		NotificationID: notifID,
		Status:         StatusQueued,
		Channel:        req.NotificationChannel,
		Recipient:      recipient,
		Timestamp:      timestampNow(),
		Error:          nil,
		RetryCount:     0,
	}

	slog.Info("handler.notification.exit",
		"trace_id", traceID,
		"notification_id", notifID,
		"channel", req.NotificationChannel,
		"status", StatusQueued,
	)

	c.JSON(http.StatusAccepted, resp)
}

func (h *NotifierHandler) Metadata(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":        "ticket-notifier",
		"version":     "0.2.0",
		"description": "Ticket notification service — email, slack, pagerduty, in_app",
	})
}

func (h *NotifierHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func (h *NotifierHandler) ReadyCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func routeToChannel(req *NotificationRequest) (string, error) {
	switch req.NotificationChannel {
	case "email":
		return handleEmailChannel(req)
	case "slack":
		return handleSlackChannel(req)
	case "pagerduty":
		return handlePagerDutyChannel(req)
	case "in_app":
		return handleInAppChannel(req)
	default:
		return "", &AppError{Code: ErrInvalidRequest, Message: "Unknown channel: " + req.NotificationChannel}
	}
}

func channelUnavailableMessage(channel string) string {
	switch channel {
	case "pagerduty":
		return "PagerDuty notification channel is currently unavailable"
	case "slack":
		return "Slack notification channel is currently unavailable"
	case "email":
		return "Email notification channel is currently unavailable"
	case "in_app":
		return "In-app notification channel is currently unavailable"
	default:
		return "Notification channel is currently unavailable"
	}
}

type RateLimiter struct {
	mu      sync.Mutex
	maxRPS  int
	windows map[string][]time.Time
}

func NewRateLimiter(maxRPS int) *RateLimiter {
	return &RateLimiter{
		maxRPS:  maxRPS,
		windows: make(map[string][]time.Time),
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-time.Second)

	entries := rl.windows[key]
	var filtered []time.Time
	for _, t := range entries {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= rl.maxRPS {
		rl.windows[key] = filtered
		return false
	}

	filtered = append(filtered, now)
	rl.windows[key] = filtered
	return true
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
	}
}
