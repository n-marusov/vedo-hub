// @ctx: POST /tickets/automatic — linear flow: parse v validate v classify v dedupe v create v respond
// @hlv:sec [INPUT_VALIDATION] — all telemetry event input validated before processing
// @hlv:sec [DESERIALIZATION] — JSON body deserialized with gin binding

package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"vedo-core/llm/src/services/ticket-api/ticketapi"
)

var (
	dedupSigRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+\+[a-zA-Z0-9_-]+\+[a-zA-Z0-9_-]+$`)
	uuidRegex     = regexp.MustCompile(`^[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}$`)
)

type AutoTicketHandler struct {
	store    *ticketapi.TicketStore
	dedup    *DedupStore
	pyClient *PythonClassifierClient
	goClass  *GoFallbackClassifier
	appVer   string
}

func NewAutoTicketHandler(
	store *ticketapi.TicketStore,
	dedup *DedupStore,
	pyClient *PythonClassifierClient,
	goClass *GoFallbackClassifier,
	appVer string,
) *AutoTicketHandler {
	return &AutoTicketHandler{
		store:    store,
		dedup:    dedup,
		pyClient: pyClient,
		goClass:  goClass,
		appVer:   appVer,
	}
}

func (h *AutoTicketHandler) HandleAutoTicket(c *gin.Context) {
	traceID := c.GetString("trace_id")
	slog.Info("handler.auto_ticket.enter",
		"trace_id", traceID,
	)

	var event TelemetryEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		slog.Error("handler.auto_ticket.invalid_event",
			"trace_id", traceID,
			"error", err.Error(),
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   ErrAutoTicketInvalidEvent,
			Message: "Invalid telemetry event: " + err.Error(),
			Ref:     "T-30",
		})
		return
	}

	if err := h.validateEvent(event); err != nil {
		code, status := err.codeAndStatus()
		slog.Error("handler.auto_ticket.validation_failed",
			"trace_id", traceID,
			"error", code,
			"message", err.Message,
		)
		c.JSON(status, ErrorResponse{
			Error:   code,
			Message: err.Message,
			Ref:     err.Ref,
		})
		return
	}

	category := h.classify(event, traceID)

	existingID := h.dedup.Check(event.DedupeSignature, event.Timestamp)
	if existingID != nil {
		existingTicket, err := h.store.GetByID(*existingID)
		if err == nil {
			slog.Info("handler.auto_ticket.duplicate",
				"trace_id", traceID,
				"signature", event.DedupeSignature,
				"existing_ticket_id", *existingID,
			)
			c.JSON(http.StatusOK, existingTicket)
			return
		}
	}

	severity := mapSeverity(event.Severity)

	req := ticketapi.CreateTicketRequest{
		Title:        event.Title,
		Description:  event.Description,
		Category:     category,
		UserSeverity: severity,
		Source:       ticketapi.TicketSourceTelemetry,
		Channel:      ticketapi.TicketChannelTelemetry,
		Attachments:  []ticketapi.TicketAttachment{},
	}

	meta := ticketapi.TicketMetadata{
		VedoVersion: h.appVer,
		Environment: event.Environment,
		UserID:      "",
		TraceID:     &event.TraceID,
		UserAgent:   DefaultUserAgent,
	}

	ticket, err := h.store.Create(req, meta)
	if err != nil {
		slog.Error("handler.auto_ticket.create_failed",
			"trace_id", traceID,
			"error", err.Error(),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   ErrAutoTicketInvalidEvent,
			Message: "Failed to create ticket: " + err.Error(),
			Ref:     "T-31",
		})
		return
	}

	ticket.Labels = append(ticket.Labels, event.Labels...)
	if event.GrafanaLink != nil {
		ticket.TelemetryLink = event.GrafanaLink
	}

	h.dedup.Record(event.DedupeSignature, ticket.ID, ticket.CreatedAt)

	slog.Info("handler.auto_ticket.exit",
		"trace_id", traceID,
		"ticket_id", ticket.ID,
		"category", ticket.Category,
		"priority", ticket.SystemPriority,
		"status", http.StatusCreated,
	)
	c.JSON(http.StatusCreated, ticket)
}

type validationError struct {
	Code    string
	Message string
	Ref     string
}

func (e *validationError) codeAndStatus() (string, int) {
	switch e.Code {
	case ErrAutoTicketInvalidDedupSig:
		return e.Code, http.StatusBadRequest
	case ErrAutoTicketMetadataTooLarge:
		return e.Code, http.StatusRequestEntityTooLarge
	case ErrAutoTicketTraceIDInvalid:
		return e.Code, http.StatusBadRequest
	default:
		return ErrAutoTicketInvalidEvent, http.StatusBadRequest
	}
}

func (h *AutoTicketHandler) validateEvent(event TelemetryEvent) *validationError {
	if !dedupSigRegex.MatchString(event.DedupeSignature) {
		return &validationError{
			Code:    ErrAutoTicketInvalidDedupSig,
			Message: "Dedupe signature must follow format: service_name+error_type+error_message_hash (32 char hex)",
			Ref:     "T-36",
		}
	}

	if !uuidRegex.MatchString(event.TraceID) {
		return &validationError{
			Code:    ErrAutoTicketTraceIDInvalid,
			Message: "Trace ID must be a valid UUID v4",
			Ref:     "T-37",
		}
	}

	if event.Metadata != nil {
		data, err := json.Marshal(event.Metadata)
		if err == nil && len(data) > MaxMetadataSize {
			return &validationError{
				Code:    ErrAutoTicketMetadataTooLarge,
				Message: "Metadata exceeds maximum size of 10 KB",
				Ref:     "T-38",
			}
		}
	}

	validTypes := map[string]bool{"alert": true, "log": true, "metric": true}
	if !validTypes[event.EventType] {
		return &validationError{
			Code:    ErrAutoTicketInvalidEvent,
			Message: "Invalid event_type: must be alert, log, or metric",
			Ref:     "T-32",
		}
	}

	return nil
}

func (h *AutoTicketHandler) classify(event TelemetryEvent, traceID string) ticketapi.TicketCategory {
	if event.Category != "" {
		cat := ticketapi.TicketCategory(event.Category)
		if ticketapi.IsValidCategory(cat) {
			return cat
		}
	}

	if h.pyClient != nil {
		result, err := h.pyClient.Classify(event)
		if err == nil && result != nil {
			slog.Info("handler.auto_ticket.classified",
				"trace_id", traceID,
				"source", "python",
				"category", result.Category,
			)
			return result.Category
		}
		slog.Warn("handler.auto_ticket.classifier_unavailable",
			"trace_id", traceID,
			"error", err.Error(),
		)
	}

	result := h.goClass.Classify(event)
	slog.Info("handler.auto_ticket.classified",
		"trace_id", traceID,
		"source", "go_fallback",
		"category", result.Category,
	)
	return result.Category
}

func (h *AutoTicketHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func (h *AutoTicketHandler) ReadyCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func (h *AutoTicketHandler) Metadata(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":        "ticket-telemetry-listener",
		"version":     h.appVer,
		"description": "Automatic ticket creation from telemetry events",
	})
}

func mapSeverity(s string) ticketapi.TicketUserSeverity {
	switch strings.ToLower(s) {
	case "critical":
		return ticketapi.TicketSeverityCritical
	case "high":
		return ticketapi.TicketSeverityHigh
	case "medium":
		return ticketapi.TicketSeverityMedium
	case "low":
		return ticketapi.TicketSeverityLow
	default:
		return ticketapi.TicketSeverityMedium
	}
}
