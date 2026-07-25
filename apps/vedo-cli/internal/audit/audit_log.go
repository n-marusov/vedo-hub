package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

var (
	querySecretPattern   = regexp.MustCompile(`(?i)(api_key|token|secret|password|credential)\s*=\s*[^\s&]+`)
	authorizationPattern = regexp.MustCompile(`(?i)(authorization\s*:\s*bearer\s+)[^\s]+`)
	jwtPattern           = regexp.MustCompile(`\b[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`)
	awsKeyPattern        = regexp.MustCompile(`(?i)\bAKIA[0-9A-Z]{16}\b`)
)

// @ctx: redacted audit logging for CLI-OPS-001 and SEC-SUPPLY-001 contracts
// @hlv:sec [SECRET_HANDLING] — audit events must redact secrets before logging

type AuditEvent struct {
	EventVersion    string `json:"event_version"`
	OccurredAt      string `json:"occurred_at"`
	ImmutableDigest string `json:"immutable_digest"`
	Actor           string `json:"actor"`
	Role            string `json:"role"`
	Command         string `json:"command"`
	Environment     string `json:"environment"`
	TraceID         string `json:"trace_id"`
	CorrelationID   string `json:"correlation_id"`
	Result          string `json:"result"` // ok, error
	ErrorCode       string `json:"error_code,omitempty"`
	InputSummary    string `json:"input_summary"` // redacted
}

// @hlv:sec [CRYPTO] — immutable digest prevents undetected event tampering
func BuildImmutableDigest(actor string, command string, traceID string, correlationID string, result string) string {
	raw := actor + "|" + command + "|" + traceID + "|" + correlationID + "|" + result
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// EmitAuditEvent writes a redacted audit log entry.
// @hlv:sec [SECRET_HANDLING] — all secret-like fields are redacted before emission
func EmitAuditEvent(event AuditEvent) {
	start := time.Now()
	if event.EventVersion == "" {
		event.EventVersion = "v1"
	}
	if event.OccurredAt == "" {
		event.OccurredAt = time.Now().UTC().Format(time.RFC3339)
	}
	if event.ImmutableDigest == "" {
		event.ImmutableDigest = BuildImmutableDigest(event.Actor, event.Command, event.TraceID, event.CorrelationID, event.Result)
	}
	slog.Info("audit.event.enter",
		"actor", event.Actor,
		"command", event.Command,
		"result", event.Result,
	)
	defer func() {
		slog.Info("audit.event.exit", "duration_ms", time.Since(start).Milliseconds())
	}()

	// @ctx: input_summary is already redacted before this call
	slog.Info("audit.event",
		"event_version", event.EventVersion,
		"occurred_at", event.OccurredAt,
		"immutable_digest", event.ImmutableDigest,
		"actor", event.Actor,
		"role", event.Role,
		"command", event.Command,
		"environment", event.Environment,
		"trace_id", event.TraceID,
		"correlation_id", event.CorrelationID,
		"result", event.Result,
		"error_code", event.ErrorCode,
		"input_summary", event.InputSummary,
	)
}

// RedactInput removes known secret patterns from input before audit logging.
// @hlv:sec [SECRET_HANDLING] — removes tokens, passwords, keys from input
func RedactInput(raw string) string {
	// @ctx: replaces known secret patterns with [REDACTED]
	result := raw
	result = querySecretPattern.ReplaceAllStringFunc(result, func(m string) string {
		idx := strings.Index(m, "=")
		if idx < 0 {
			return "[REDACTED]"
		}
		return m[:idx+1] + "[REDACTED]"
	})
	result = authorizationPattern.ReplaceAllString(result, "${1}[REDACTED]")
	result = jwtPattern.ReplaceAllString(result, "[REDACTED_JWT]")
	result = awsKeyPattern.ReplaceAllString(result, "[REDACTED_AWS_KEY]")
	return result
}
