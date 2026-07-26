package cli

import "testing"

// @ctx: SEC-PAM-001 contract and property tests

// @hlv PAM_TICKET_REQUIRED
func TestEvaluatePAMAccess_TicketRequired(t *testing.T) {
	result := EvaluatePAMAccess(PAMRequest{ActorRole: "SRE", RequestedTTLMinutes: 30, MFAFactor: "fido2", RequestID: "req-1", EntityID: "tenant-9d7c"})
	if result.Error == nil || result.Error.Code != "PAM_TICKET_REQUIRED" {
		t.Fatalf("expected PAM_TICKET_REQUIRED, got %+v", result.Error)
	}
}

// @hlv PAM_MFA_FACTOR_NOT_COMPLIANT
func TestEvaluatePAMAccess_MFACompliantRequired(t *testing.T) {
	result := EvaluatePAMAccess(PAMRequest{ActorRole: "SecurityAdmin", TicketID: "INC-1", RequestedTTLMinutes: 30, MFAFactor: "totp", RequestID: "req-2", EntityID: "tenant-9d7c"})
	if result.Error == nil || result.Error.Code != "PAM_MFA_FACTOR_NOT_COMPLIANT" {
		t.Fatalf("expected PAM_MFA_FACTOR_NOT_COMPLIANT, got %+v", result.Error)
	}
}

// @hlv PAM_TTL_EXCEEDED
func TestEvaluatePAMAccess_TTLExceeded(t *testing.T) {
	result := EvaluatePAMAccess(PAMRequest{ActorRole: "SecurityAdmin", TicketID: "INC-2", RequestedTTLMinutes: 120, MFAFactor: "fido2", RequestID: "req-3", EntityID: "tenant-9d7c"})
	if result.Error == nil || result.Error.Code != "PAM_TTL_EXCEEDED" {
		t.Fatalf("expected PAM_TTL_EXCEEDED, got %+v", result.Error)
	}
}

// @hlv PAM_BREAK_GLASS_APPROVAL_MISSING
func TestEvaluatePAMAccess_BreakGlassApprovalsRequired(t *testing.T) {
	result := EvaluatePAMAccess(PAMRequest{ActorRole: "SecurityAdmin", TicketID: "INC-3", RequestedTTLMinutes: 30, MFAFactor: "fido2", BreakGlass: true, Approvers: []string{"IncidentCommander"}, RequestID: "req-4", EntityID: "tenant-9d7c"})
	if result.Error == nil || result.Error.Code != "PAM_BREAK_GLASS_APPROVAL_MISSING" {
		t.Fatalf("expected PAM_BREAK_GLASS_APPROVAL_MISSING, got %+v", result.Error)
	}
}

// @hlv pam_ttl_max_60
func TestPAMProperty_TTLApprovalBoundedBySixtyMinutes(t *testing.T) {
	for ttl := 1; ttl <= 120; ttl++ {
		result := EvaluatePAMAccess(PAMRequest{ActorRole: "SecurityAdmin", TicketID: "INC-5", RequestedTTLMinutes: ttl, MFAFactor: "fido2", RequestID: "req-ttl", EntityID: "tenant-9d7c"})
		if ttl <= 60 && result.Decision != "approved" {
			t.Fatalf("expected ttl=%d to be approved", ttl)
		}
		if ttl > 60 && (result.Error == nil || result.Error.Code != "PAM_TTL_EXCEEDED") {
			t.Fatalf("expected ttl=%d to be denied by PAM_TTL_EXCEEDED", ttl)
		}
	}
}

// @hlv pam_break_glass_dual_approval
func TestPAMProperty_BreakGlassRequiresDualApprovals(t *testing.T) {
	cases := [][]string{{"IncidentCommander"}, {"SecurityLead"}, {"SecurityLead", "IncidentCommander"}}
	for _, approvers := range cases {
		result := EvaluatePAMAccess(PAMRequest{ActorRole: "SecurityAdmin", TicketID: "INC-6", RequestedTTLMinutes: 30, MFAFactor: "fido2", BreakGlass: true, Approvers: approvers, RequestID: "req-approvers", EntityID: "tenant-9d7c"})
		if len(approvers) == 2 && result.Decision != "approved" {
			t.Fatalf("expected dual approvers to be approved")
		}
		if len(approvers) != 2 && (result.Error == nil || result.Error.Code != "PAM_BREAK_GLASS_APPROVAL_MISSING") {
			t.Fatalf("expected single approver case to be denied")
		}
	}
}

// @hlv pam_session_recording_required
func TestPAMProperty_ApprovedSessionsAlwaysRequireRecording(t *testing.T) {
	result := EvaluatePAMAccess(PAMRequest{ActorRole: "SecurityAdmin", TicketID: "INC-7", RequestedTTLMinutes: 30, MFAFactor: "fido2", RequestID: "req-7", EntityID: "tenant-9d7c"})
	if result.Decision != "approved" {
		t.Fatal("expected approved result")
	}
	if !result.SessionRecordingRequired || !result.CommandAuditRequired {
		t.Fatal("approved result must require recording and command audit")
	}
}
