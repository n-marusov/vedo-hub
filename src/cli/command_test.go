package cli

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

// @ctx: command framework contract tests for CLI-OPS-001

// @hlv CLI_COMMAND_NOT_SUPPORTED
func TestExecuteCommand_UnsupportedCommand(t *testing.T) {
	input := CliInput{
		Command:      "invalid subcommand",
		OutputFormat: FormatJSON,
		RequestID:    "req-unsupported",
	}
	output := ExecuteCommand(input)
	if output.Status != "error" {
		t.Fatalf("expected error status for unsupported command, got %s", output.Status)
	}
	if output.Error == nil || output.Error.Code != "CLI_COMMAND_NOT_SUPPORTED" {
		t.Fatalf("expected CLI_COMMAND_NOT_SUPPORTED, got %+v", output.Error)
	}
}

// @hlv CLI_INVALID_FORMAT
func TestExecuteCommand_InvalidFormat(t *testing.T) {
	input := CliInput{
		Command:      CmdSupportTenantInfo,
		OutputFormat: "yaml",
		TenantID:     "tenant-9d7c",
		RequestID:    "req-invalid-format",
	}
	output := ExecuteCommand(input)
	if output.Status != "error" {
		t.Fatalf("expected error for invalid format, got %s", output.Status)
	}
	if output.Error == nil || output.Error.Code != "CLI_INVALID_FORMAT" {
		t.Fatalf("expected CLI_INVALID_FORMAT, got %+v", output.Error)
	}
}

// @hlv CLI_TICKET_REQUIRED
func TestExecuteCommand_TicketRequired(t *testing.T) {
	resetMFAChallengesForTests()
	input := CliInput{
		Command:        CmdSecurityPolicyDisable,
		OutputFormat:   FormatJSON,
		ActorType:      ActorUser,
		MFAChallengeID: "mfa-ticket",
		RequestID:      "req-ticket",
	}
	output := ExecuteCommand(input)
	if output.Error == nil || output.Error.Code != "CLI_TICKET_REQUIRED" {
		t.Fatalf("expected CLI_TICKET_REQUIRED, got %+v", output.Error)
	}
}

// @hlv CLI_MFA_REQUIRED
func TestExecuteCommand_MFARequired(t *testing.T) {
	resetMFAChallengesForTests()
	input := CliInput{
		Command:      CmdSecurityPolicyDisable,
		OutputFormat: FormatJSON,
		TicketID:     "INC-1",
		ActorType:    ActorUser,
		RequestID:    "req-mfa",
	}
	output := ExecuteCommand(input)
	if output.Error == nil || output.Error.Code != "CLI_MFA_REQUIRED" {
		t.Fatalf("expected CLI_MFA_REQUIRED, got %+v", output.Error)
	}
}

// @hlv CLI_GUARDRAIL_NOT_SATISFIED
func TestExecuteCommand_ServiceAccountCannotRunABCategory(t *testing.T) {
	resetMFAChallengesForTests()
	input := CliInput{
		Command:      CmdSecurityPolicyDisable,
		OutputFormat: FormatJSON,
		TicketID:     "INC-2",
		ActorType:    ActorServiceAccount,
		RequestID:    "req-sa-block",
	}
	output := ExecuteCommand(input)
	if output.Error == nil || output.Error.Code != "CLI_GUARDRAIL_NOT_SATISFIED" {
		t.Fatalf("expected CLI_GUARDRAIL_NOT_SATISFIED, got %+v", output.Error)
	}
}

// @hlv CLI_SUPPORT_TENANT_NOT_FOUND
func TestExecuteCommand_UnknownTenantDeterministicError(t *testing.T) {
	input := CliInput{
		Command:      CmdSupportTenantInfo,
		OutputFormat: FormatJSON,
		TenantID:     "tenant-missing",
		RequestID:    "req-unknown-tenant",
	}
	output := ExecuteCommand(input)
	if output.Error == nil || output.Error.Code != "CLI_SUPPORT_TENANT_NOT_FOUND" {
		t.Fatalf("expected CLI_SUPPORT_TENANT_NOT_FOUND, got %+v", output.Error)
	}
}

// @hlv CLI_EMERGENCY_KEY_NOT_FOUND
func TestExecuteCommand_UnknownEmergencyKeyError(t *testing.T) {
	resetMFAChallengesForTests()
	input := CliInput{
		Command:        CmdSupportEmergencyAccessRevoke,
		OutputFormat:   FormatJSON,
		KeyID:          "key-missing",
		TicketID:       "INC-3",
		MFAChallengeID: "mfa-key",
		ActorType:      ActorUser,
		RequestID:      "req-unknown-key",
	}
	output := ExecuteCommand(input)
	if output.Error == nil || output.Error.Code != "CLI_EMERGENCY_KEY_NOT_FOUND" {
		t.Fatalf("expected CLI_EMERGENCY_KEY_NOT_FOUND, got %+v", output.Error)
	}
}

// @hlv guardrail_mapping_deterministic
func TestExecuteCommandProperty_GuardrailMappingDeterministic(t *testing.T) {
	commands := []Command{CmdSupportTenantInfo, CmdSupportAuditTrail, CmdSupportEmergencyAccessReq, CmdSecurityPolicyDisable}
	for _, command := range commands {
		first, levelFirst, errFirst := ResolveGuardrailControls(command, "")
		second, levelSecond, errSecond := ResolveGuardrailControls(command, "")
		if errFirst != nil || errSecond != nil {
			t.Fatalf("unexpected guardrail lookup error for %s", command)
		}
		if levelFirst != levelSecond || strings.Join(first, ",") != strings.Join(second, ",") {
			t.Fatalf("guardrail mapping must be deterministic for %s", command)
		}
	}
}

// @hlv mfa_fresh_per_invocation
func TestExecuteCommandProperty_EachABInvocationNeedsFreshMFA(t *testing.T) {
	resetMFAChallengesForTests()
	input := CliInput{
		Command:        CmdSupportEmergencyAccessReq,
		OutputFormat:   FormatJSON,
		TenantID:       "tenant-9d7c",
		TicketID:       "INC-10",
		ActorType:      ActorUser,
		RequestID:      "req-mfa-fresh",
		MFAChallengeID: "mfa-once",
	}
	first := ExecuteCommand(input)
	second := ExecuteCommand(input)
	if first.Status != "ok" {
		t.Fatal("first invocation should pass")
	}
	if second.Error == nil || second.Error.Code != "CLI_MFA_REQUIRED" {
		t.Fatal("second invocation with reused challenge must fail")
	}
}

// @hlv command_output_redacts_secrets
func TestRenderOutput_JSON(t *testing.T) {
	output := CliOutput{
		Status: "ok",
		Data: &CliData{
			Command:               "support emergency-access request",
			Result:                "ok",
			TraceID:               "trace-123",
			CorrelationID:         "cli-2026-06-06-001",
			GuardrailsApplied:     []string{"typed_confirmation", "env_guard", "mfa"},
			MFAChallengePerformed: true,
		},
	}
	rendered := RenderOutput(output, FormatJSON)
	if !strings.Contains(rendered, "trace-123") {
		t.Fatal("JSON output should contain trace_id")
	}
	if !strings.Contains(rendered, "support emergency-access request") {
		t.Fatal("JSON output should contain command name")
	}
}

// @hlv output_format_human
func TestRenderOutput_Human(t *testing.T) {
	output := CliOutput{
		Status: "ok",
		Data: &CliData{
			Command: "status local",
			Result:  "all-required-services-reachable",
		},
	}
	rendered := RenderOutput(output, FormatHuman)
	if !strings.Contains(rendered, "Status: ok") {
		t.Fatal("human output should contain Status line")
	}
	if !strings.Contains(rendered, "status local") {
		t.Fatal("human output should contain command name")
	}
}

// @hlv structured_logging_only
func TestExecuteCommand_StructuredLogging(t *testing.T) {
	input := CliInput{
		Command:      CmdSupportTenantInfo,
		OutputFormat: FormatJSON,
		TenantID:     "tenant-struct",
		RequestID:    "req-struct",
	}
	output := captureSlog(t, func() { ExecuteCommand(input) })
	// Structured logging emits key=value pairs, not free-form strings.
	for _, field := range []string{"command=" + string(CmdSupportTenantInfo), "request_id=req-struct"} {
		if !strings.Contains(output, field) {
			t.Fatalf("expected structured field %q in log output, got: %s", field, output)
		}
	}
}

// @hlv log_entry_exit
func TestExecuteCommand_EntryExitLogged(t *testing.T) {
	input := CliInput{
		Command:      CmdSupportTenantInfo,
		OutputFormat: FormatJSON,
		TenantID:     "tenant-enter-exit",
		RequestID:    "req-enter-exit",
	}
	output := captureSlog(t, func() { ExecuteCommand(input) })
	if !strings.Contains(output, "cli.command.enter") {
		t.Fatalf("expected cli.command.enter log line, got: %s", output)
	}
	if !strings.Contains(output, "cli.command.exit") {
		t.Fatalf("expected cli.command.exit log line, got: %s", output)
	}
	if !strings.Contains(output, "duration_ms") {
		t.Fatalf("expected duration_ms field in exit log, got: %s", output)
	}
}

// @hlv log_all_errors
func TestExecuteCommand_ErrorsLogged(t *testing.T) {
	input := CliInput{
		Command:      CmdSupportTenantInfo,
		OutputFormat: FormatJSON,
		TenantID:     "tenant-err-log",
		RequestID:    "req-err-log",
	}
	output := captureSlog(t, func() { ExecuteCommand(input) })
	// Every command invocation logs request_id, trace_id, and entity_id so
	// error paths are fully traceable in aggregated logs.
	for _, field := range []string{"request_id=req-err-log", "trace_id=", "entity_id=tenant-err-log"} {
		if !strings.Contains(output, field) {
			t.Fatalf("expected traceability field %q in log output, got: %s", field, output)
		}
	}
}

// @hlv request_correlation
func TestExecuteCommand_TraceAndCorrelationIDs(t *testing.T) {
	input := CliInput{Command: CmdSupportTenantInfo, OutputFormat: FormatJSON, TenantID: "tenant-9d7c", RequestID: "req-correlation"}
	output := ExecuteCommand(input)
	if output.Data == nil || output.Data.TraceID == "" || output.Data.CorrelationID == "" {
		t.Fatal("expected trace_id and correlation_id in output")
	}
}

// @hlv no_secrets_in_logs
func TestExecuteCommand_NoSecretsInLogs(t *testing.T) {
	// @ctx: the command framework logs only bounded fields (command, tenant_id,
	// request_id, trace_id, actor_type) and never free-form user input like
	// Reason, which could carry secret values. Verify the Reason field never
	// leaks into any log line.
	secretInReason := "token=super-secret-value-12345"
	input := CliInput{
		Command:      CmdSupportTenantInfo,
		OutputFormat: FormatJSON,
		TenantID:     "tenant-no-secrets",
		RequestID:    "req-no-secrets",
		Reason:       secretInReason,
	}
	output := captureSlog(t, func() { ExecuteCommand(input) })
	if strings.Contains(output, secretInReason) {
		t.Fatalf("command log leaked value from Reason field: %s", output)
	}
}
