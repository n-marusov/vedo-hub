package cli

import (
	"strings"
	"testing"
)

// @ctx: stage 5 CLI support command tests for tenant-info, audit-trail, backups, and emergency access shell

func TestSupportTenantInfoCommandReturnsTenantFields(t *testing.T) {
	resetSupportShellForTests()
	output := ExecuteCommand(CliInput{Command: CmdSupportTenantInfo, OutputFormat: FormatJSON, TenantID: "tenant-9d7c", RequestID: "req-support-tenant-1"})
	if output.Status != "ok" || output.Data == nil {
		t.Fatalf("expected ok output, got %+v", output.Error)
	}
	if !strings.Contains(output.Data.Result, "tenant-9d7c") || !strings.Contains(output.Data.Result, "enterprise") {
		t.Fatalf("expected tenant metadata in result, got %s", output.Data.Result)
	}
}

func TestSupportAuditTrailIncludesChecksums(t *testing.T) {
	resetSupportShellForTests()
	output := ExecuteCommand(CliInput{Command: CmdSupportAuditTrail, OutputFormat: FormatJSON, TenantID: "tenant-9d7c", RequestID: "req-support-audit-1"})
	if output.Status != "ok" || output.Data == nil {
		t.Fatalf("expected ok output, got %+v", output.Error)
	}
	if !strings.Contains(output.Data.Result, "checksum") || !strings.Contains(output.Data.Result, "event_id") {
		t.Fatalf("expected checksum fields in audit trail output, got %s", output.Data.Result)
	}
}

// @hlv SUPPORT_AUDIT_WORM_UNAVAILABLE
func TestSupportAuditTrailReturnsWORMUnavailable(t *testing.T) {
	resetSupportShellForTests()
	setSupportWORMAvailabilityForTests(false)
	output := ExecuteCommand(CliInput{Command: CmdSupportAuditTrail, OutputFormat: FormatJSON, TenantID: "tenant-9d7c", RequestID: "req-support-audit-2"})
	if output.Error == nil || output.Error.Code != "SUPPORT_AUDIT_WORM_UNAVAILABLE" {
		t.Fatalf("expected SUPPORT_AUDIT_WORM_UNAVAILABLE, got %+v", output.Error)
	}
}

func TestSupportListBackupsReturnsBackupRecords(t *testing.T) {
	resetSupportShellForTests()
	output := ExecuteCommand(CliInput{Command: CmdSupportListBackups, OutputFormat: FormatJSON, TenantID: "tenant-9d7c", RequestID: "req-support-backups-1"})
	if output.Status != "ok" || output.Data == nil {
		t.Fatalf("expected ok output, got %+v", output.Error)
	}
	if !strings.Contains(output.Data.Result, "backup-2026-06-10") {
		t.Fatalf("expected backups in result payload, got %s", output.Data.Result)
	}
}

// @hlv CLI_EMERGENCY_KEY_NOT_FOUND
func TestSupportEmergencyAccessListUnknownKeyReturnsNotFound(t *testing.T) {
	resetSupportShellForTests()
	resetMFAChallengesForTests()
	output := ExecuteCommand(CliInput{Command: CmdSupportEmergencyAccessList, OutputFormat: FormatJSON, TenantID: "tenant-9d7c", KeyID: "key-missing", MFAChallengeID: "mfa-list-1", ActorType: ActorUser, RequestID: "req-support-em-list-1"})
	if output.Error == nil || output.Error.Code != "CLI_EMERGENCY_KEY_NOT_FOUND" {
		t.Fatalf("expected CLI_EMERGENCY_KEY_NOT_FOUND, got %+v", output.Error)
	}
}

func TestSupportEmergencyAccessRequestListAndRevokeFlow(t *testing.T) {
	resetSupportShellForTests()
	resetMFAChallengesForTests()

	request := ExecuteCommand(CliInput{Command: CmdSupportEmergencyAccessReq, OutputFormat: FormatJSON, TenantID: "tenant-9d7c", TicketID: "INC-2026-114", MFAChallengeID: "mfa-req-1", ActorType: ActorUser, RequestID: "req-support-em-req-1"})
	if request.Status != "ok" || request.Data == nil {
		t.Fatalf("expected request to succeed, got %+v", request.Error)
	}
	if !strings.Contains(request.Data.Result, "key-tenant-9d7c") {
		t.Fatalf("expected generated key in request result, got %s", request.Data.Result)
	}

	list := ExecuteCommand(CliInput{Command: CmdSupportEmergencyAccessList, OutputFormat: FormatJSON, TenantID: "tenant-9d7c", MFAChallengeID: "mfa-list-2", ActorType: ActorUser, RequestID: "req-support-em-list-2"})
	if list.Status != "ok" || list.Data == nil {
		t.Fatalf("expected list to succeed, got %+v", list.Error)
	}
	if !strings.Contains(list.Data.Result, "key-tenant-9d7c") {
		t.Fatalf("expected listed emergency key, got %s", list.Data.Result)
	}

	revoke := ExecuteCommand(CliInput{Command: CmdSupportEmergencyAccessRevoke, OutputFormat: FormatJSON, KeyID: "key-tenant-9d7c-2", TicketID: "INC-2026-114", MFAChallengeID: "mfa-revoke-1", ActorType: ActorUser, RequestID: "req-support-em-revoke-1"})
	if revoke.Status != "ok" || revoke.Data == nil {
		t.Fatalf("expected revoke to succeed, got %+v", revoke.Error)
	}
	if !strings.Contains(revoke.Data.Result, "\"revoked\":true") {
		t.Fatalf("expected revoked marker, got %s", revoke.Data.Result)
	}
}
