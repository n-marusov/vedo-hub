package audit

// Validates: REQ-NFR.SECURITY.audit-access-audit
// Validates: REQ-FUN.INTEGRATION.audit-integration
// Validates: REQ-FUN.INTEGRATION.audit-log-content

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// captureSlog replaces the default slog logger with a text handler writing to
// a buffer for the duration of fn, returning the captured output. The original
// logger is restored automatically when the test finishes.
func captureSlog(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	original := slog.Default()
	slog.SetDefault(slog.New(handler))
	t.Cleanup(func() { slog.SetDefault(original) })
	fn()
	return buf.String()
}

// @hlv no_secrets_in_logs
func TestEmitAuditEvent_NoSecretsInLogs(t *testing.T) {
	secretValue := "super-secret-token-value"
	event := AuditEvent{
		Actor:         "admin",
		Role:          "platform",
		Command:       "status local",
		Environment:   "local",
		TraceID:       "trace-123",
		CorrelationID: "cli-2026-06-06-001",
		Result:        "ok",
		InputSummary:  RedactInput("token=" + secretValue),
	}
	output := captureSlog(t, func() { EmitAuditEvent(event) })
	if strings.Contains(output, secretValue) {
		t.Fatalf("audit log leaked secret value: %s", output)
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Fatalf("expected redaction marker in audit log, got: %s", output)
	}
}

// @hlv structured_logging_only
func TestEmitAuditEvent_StructuredLogging(t *testing.T) {
	event := AuditEvent{
		Actor:   "admin",
		Command: "audit-test-command",
		Result:  "ok",
		TraceID: "trace-struct",
	}
	output := captureSlog(t, func() { EmitAuditEvent(event) })
	// Structured logging emits key=value pairs, not free-form interpolation.
	for _, key := range []string{"actor=admin", "command=audit-test-command", "result=ok"} {
		if !strings.Contains(output, key) {
			t.Fatalf("expected structured field %q in log output, got: %s", key, output)
		}
	}
}

// @hlv log_entry_exit
func TestEmitAuditEvent_EntryExitLogged(t *testing.T) {
	event := AuditEvent{Actor: "admin", Command: "enter-exit-test", Result: "ok"}
	output := captureSlog(t, func() { EmitAuditEvent(event) })
	if !strings.Contains(output, "audit.event.enter") {
		t.Fatalf("expected audit.event.enter log line, got: %s", output)
	}
	if !strings.Contains(output, "audit.event.exit") {
		t.Fatalf("expected audit.event.exit log line, got: %s", output)
	}
	if !strings.Contains(output, "duration_ms") {
		t.Fatalf("expected duration_ms field in exit log, got: %s", output)
	}
}

// @hlv log_state_changes
func TestEmitAuditEvent_StateChangeLogged(t *testing.T) {
	event := AuditEvent{
		Actor:   "state-actor",
		Role:    "operator",
		Command: "commit-apply",
		Result:  "ok",
		TraceID: "trace-state",
	}
	output := captureSlog(t, func() { EmitAuditEvent(event) })
	// A state-change record must capture who did what and the outcome.
	for _, required := range []string{
		"actor=state-actor",
		"command=commit-apply",
		"result=ok",
	} {
		if !strings.Contains(output, required) {
			t.Fatalf("state-change log missing %q, got: %s", required, output)
		}
	}
}

// @hlv log_all_errors
func TestAuditEvent_ErrorPathLogs(t *testing.T) {
	event := AuditEvent{
		Actor:        "admin",
		Command:      "auth resolve-credentials",
		Result:       "error",
		ErrorCode:    "CLI_CREDENTIALS_NOT_CONFIGURED",
		InputSummary: "redacted",
	}
	output := captureSlog(t, func() { EmitAuditEvent(event) })
	if !strings.Contains(output, "result=error") {
		t.Fatalf("expected result=error in log, got: %s", output)
	}
	if !strings.Contains(output, "CLI_CREDENTIALS_NOT_CONFIGURED") {
		t.Fatalf("expected error_code in audit log, got: %s", output)
	}
}

// @hlv no_secrets_in_logs
func TestRedactInput_RemovesSecretlikeValues(t *testing.T) {
	raw := "api_key=sample123 token=maskme Authorization: Bearer eyJ0eXAiOiJKV1QifQ.abc.def key=AKIA1234567890ABCDEf"
	redacted := RedactInput(raw)
	for _, forbidden := range []string{"sample123", "maskme", "eyJ0eXAiOiJKV1QifQ.abc.def", "AKIA1234567890ABCDEf"} {
		if strings.Contains(redacted, forbidden) {
			t.Fatalf("redacted output still contains secret fragment: %s", forbidden)
		}
	}
	for _, expected := range []string{"api_key=[REDACTED]", "token=[REDACTED]", "Authorization: Bearer [REDACTED]", "[REDACTED_AWS_KEY]"} {
		if !strings.Contains(redacted, expected) {
			t.Fatalf("expected redaction marker not found: %s", expected)
		}
	}
}
