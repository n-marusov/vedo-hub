// @ctx: telemetry event types and error codes for automatic ticket creation
// @hlv:sec [INPUT_VALIDATION] — telemetry event input type

package main

import "time"

type TelemetryEvent struct {
	EventType       string         `json:"event_type" binding:"required"`
	Source          string         `json:"source" binding:"required"`
	Environment     string         `json:"environment" binding:"required"`
	Timestamp       time.Time      `json:"timestamp" binding:"required"`
	Severity        string         `json:"severity" binding:"required"`
	Title           string         `json:"title" binding:"required"`
	Description     string         `json:"description" binding:"required"`
	Category        string         `json:"category"`
	TraceID         string         `json:"trace_id" binding:"required"`
	SpanID          string         `json:"span_id"`
	Labels          []string       `json:"labels"`
	Metadata        map[string]any `json:"metadata"`
	DedupeSignature string         `json:"dedupe_signature" binding:"required"`
	GrafanaLink     *string        `json:"grafana_link"`
}

type DedupEntry struct {
	TicketID  string    `json:"ticket_id"`
	CreatedAt time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Ref     string `json:"ticket_management_system_ref,omitempty"`
}

const (
	ErrAutoTicketInvalidEvent      = "AUTO-TICKET-INVALID-EVENT"
	ErrAutoTicketDuplicate         = "AUTO-TICKET-DUPLICATE"
	ErrAutoTicketClassifierUnavail = "AUTO-TICKET-CLASSIFIER-UNAVAILABLE"
	ErrAutoTicketInvalidDedupSig   = "AUTO-TICKET-INVALID-DEDUP-SIGNATURE"
	ErrAutoTicketMetadataTooLarge  = "AUTO-TICKET-METADATA-TOO-LARGE"
	ErrAutoTicketTraceIDInvalid    = "AUTO-TICKET-TRACE-ID-INVALID"
)

const MaxMetadataSize = 10 * 1024
const DefaultUserAgent = "telemetry-collector/v1.0"
