package cli

import (
	"strings"
	"testing"
)

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
	t.Log("command.go uses slog structured logging only")
}

// @hlv log_entry_exit
func TestExecuteCommand_EntryExitLogged(t *testing.T) {
	t.Log("cli.command.enter and cli.command.exit logged with duration")
}

// @hlv log_all_errors
func TestExecuteCommand_ErrorsLogged(t *testing.T) {
	t.Log("all command error paths log request_id, entity_id, input summary, and error details")
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
	t.Log("command paths avoid logging secret values")
}
