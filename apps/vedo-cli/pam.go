package cli

import (
	"log/slog"
	"time"
)

// @ctx: SEC-PAM-001 privileged access policy checks and extension points

type PAMRequest struct {
	ActorRole           string
	TicketID            string
	Scope               string
	RequestedTTLMinutes int
	MFAFactor           string
	BreakGlass          bool
	Approvers           []string
	RequestID           string
	EntityID            string
}

type PAMResult struct {
	Decision                 string
	GrantedTTLMinutes        int
	ApprovalChain            []string
	SessionRecordingRequired bool
	CommandAuditRequired     bool
	Error                    *CliError
}

// @hlv:sec [AUTH_BOUNDARY] — privileged session policy enforces role and MFA controls
func EvaluatePAMAccess(input PAMRequest) PAMResult {
	start := time.Now()
	slog.Info("pam.evaluate.enter", "request_id", input.RequestID, "entity_id", input.EntityID, "role", input.ActorRole, "scope", input.Scope)
	defer func() {
		slog.Info("pam.evaluate.exit", "request_id", input.RequestID, "entity_id", input.EntityID, "duration_ms", time.Since(start).Milliseconds())
	}()

	if input.TicketID == "" {
		slog.Error("pam.evaluate.ticket_missing", "request_id", input.RequestID, "entity_id", input.EntityID)
		// @hlv PAM_TICKET_REQUIRED
		return PAMResult{Decision: "denied", Error: &CliError{Code: "PAM_TICKET_REQUIRED", Message: "ticket id is required"}}
	}

	if input.RequestedTTLMinutes > 60 || input.RequestedTTLMinutes < 1 {
		slog.Error("pam.evaluate.ttl_exceeded", "request_id", input.RequestID, "entity_id", input.EntityID, "requested_ttl_minutes", input.RequestedTTLMinutes)
		// @hlv PAM_TTL_EXCEEDED
		return PAMResult{Decision: "denied", Error: &CliError{Code: "PAM_TTL_EXCEEDED", Message: "requested ttl exceeds policy max"}}
	}

	if input.MFAFactor != "fido2" && input.MFAFactor != "webauthn" {
		slog.Error("pam.evaluate.mfa_non_compliant", "request_id", input.RequestID, "entity_id", input.EntityID, "mfa_factor", input.MFAFactor)
		// @hlv PAM_MFA_FACTOR_NOT_COMPLIANT
		return PAMResult{Decision: "denied", Error: &CliError{Code: "PAM_MFA_FACTOR_NOT_COMPLIANT", Message: "phishing-resistant MFA is required for privileged role"}}
	}

	if input.BreakGlass {
		hasIncidentCommander := false
		hasSecurityLead := false
		for _, approver := range input.Approvers {
			if approver == "IncidentCommander" {
				hasIncidentCommander = true
			}
			if approver == "SecurityLead" {
				hasSecurityLead = true
			}
		}
		if !hasIncidentCommander || !hasSecurityLead {
			slog.Error("pam.evaluate.break_glass.approval_missing", "request_id", input.RequestID, "entity_id", input.EntityID)
			// @hlv PAM_BREAK_GLASS_APPROVAL_MISSING
			return PAMResult{Decision: "denied", Error: &CliError{Code: "PAM_BREAK_GLASS_APPROVAL_MISSING", Message: "break-glass requires Security Lead and Incident Commander approvals"}}
		}
	}

	slog.Info("pam.evaluate.state_changed",
		"request_id", input.RequestID,
		"entity_id", input.EntityID,
		"old", "denied",
		"new", "approved",
		"event", "state changed",
	)

	return PAMResult{
		Decision:                 "approved",
		GrantedTTLMinutes:        input.RequestedTTLMinutes,
		ApprovalChain:            []string{"jit-policy-engine"},
		SessionRecordingRequired: true,
		CommandAuditRequired:     true,
	}
}
