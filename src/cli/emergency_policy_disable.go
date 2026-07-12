package cli

import (
	"log/slog"
	"time"
)

// @ctx: SEC-EPD-001 emergency policy disable validation shell

type EmergencyPolicyDisableInput struct {
	Scope                string
	TicketID             string
	IncidentID           string
	TriggerSignal        string
	Approvers            []string
	CurrentDisableMinute int
	RequestID            string
}

type EmergencyPolicyDisableOutput struct {
	Status                   string
	ActivatedAt              string
	TTLMinutes               int
	PostDisableSmokeRequired bool
	EscalationLevel          string
	Error                    *CliError
}

// @hlv:sec [AUTH_BOUNDARY] — emergency disable requires dual authorization controls
func EvaluateEmergencyPolicyDisable(input EmergencyPolicyDisableInput) EmergencyPolicyDisableOutput {
	start := time.Now()
	slog.Info("epd.evaluate.enter", "request_id", input.RequestID, "scope", input.Scope, "incident_id", input.IncidentID)
	defer func() {
		slog.Info("epd.evaluate.exit", "request_id", input.RequestID, "duration_ms", time.Since(start).Milliseconds())
	}()

	if input.TicketID == "" || input.IncidentID == "" {
		slog.Error("epd.evaluate.ticket_or_incident_missing", "request_id", input.RequestID)
		// @hlv EPD_TICKET_OR_INCIDENT_MISSING
		return EmergencyPolicyDisableOutput{Status: "denied", Error: &CliError{Code: "EPD_TICKET_OR_INCIDENT_MISSING", Message: "ticket id and incident id are required"}}
	}

	if input.TriggerSignal != "five_xx_spike" && input.TriggerSignal != "auth_failure_spike" && input.TriggerSignal != "false_block_confirmed" {
		slog.Error("epd.evaluate.trigger_not_met", "request_id", input.RequestID, "trigger_signal", input.TriggerSignal)
		// @hlv EPD_TRIGGER_NOT_MET
		return EmergencyPolicyDisableOutput{Status: "denied", Error: &CliError{Code: "EPD_TRIGGER_NOT_MET", Message: "trigger thresholds are not met"}}
	}

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
		slog.Error("epd.evaluate.dual_approval_missing", "request_id", input.RequestID, "incident_id", input.IncidentID)
		// @hlv EPD_DUAL_APPROVAL_MISSING
		return EmergencyPolicyDisableOutput{Status: "denied", Error: &CliError{Code: "EPD_DUAL_APPROVAL_MISSING", Message: "Security Lead and Incident Commander approvals are required"}}
	}

	if input.CurrentDisableMinute > 60 {
		slog.Error("epd.evaluate.ttl_exceeded", "request_id", input.RequestID, "current_disable_minute", input.CurrentDisableMinute)
		// @hlv EPD_TTL_EXCEEDED
		return EmergencyPolicyDisableOutput{Status: "denied", Error: &CliError{Code: "EPD_TTL_EXCEEDED", Message: "existing global disable exceeded maximum ttl and requires L5 escalation"}}
	}

	slog.Info("epd.evaluate.state_changed", "request_id", input.RequestID, "entity_id", input.IncidentID, "old", "cleared", "new", "activated", "event", "state changed")
	return EmergencyPolicyDisableOutput{
		Status:                   "activated",
		ActivatedAt:              time.Now().UTC().Format(time.RFC3339),
		TTLMinutes:               60,
		PostDisableSmokeRequired: true,
		EscalationLevel:          "L2",
	}
}
