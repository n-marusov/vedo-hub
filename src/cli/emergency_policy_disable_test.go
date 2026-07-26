package cli

import "testing"

// @ctx: SEC-EPD-001 contract and property tests

// @hlv EPD_DUAL_APPROVAL_MISSING
func TestEvaluateEmergencyPolicyDisable_DualApprovalRequired(t *testing.T) {
	result := EvaluateEmergencyPolicyDisable(EmergencyPolicyDisableInput{Scope: "global", TicketID: "INC-1", IncidentID: "P0-1", TriggerSignal: "five_xx_spike", Approvers: []string{"IncidentCommander"}, RequestID: "req-epd-1"})
	if result.Error == nil || result.Error.Code != "EPD_DUAL_APPROVAL_MISSING" {
		t.Fatalf("expected EPD_DUAL_APPROVAL_MISSING, got %+v", result.Error)
	}
}

// @hlv EPD_TRIGGER_NOT_MET
func TestEvaluateEmergencyPolicyDisable_TriggerRequired(t *testing.T) {
	result := EvaluateEmergencyPolicyDisable(EmergencyPolicyDisableInput{Scope: "global", TicketID: "INC-1", IncidentID: "P0-1", TriggerSignal: "normal", Approvers: []string{"IncidentCommander", "SecurityLead"}, RequestID: "req-epd-2"})
	if result.Error == nil || result.Error.Code != "EPD_TRIGGER_NOT_MET" {
		t.Fatalf("expected EPD_TRIGGER_NOT_MET, got %+v", result.Error)
	}
}

// @hlv EPD_TICKET_OR_INCIDENT_MISSING
func TestEvaluateEmergencyPolicyDisable_TicketAndIncidentRequired(t *testing.T) {
	result := EvaluateEmergencyPolicyDisable(EmergencyPolicyDisableInput{Scope: "global", TriggerSignal: "five_xx_spike", Approvers: []string{"IncidentCommander", "SecurityLead"}, RequestID: "req-epd-3"})
	if result.Error == nil || result.Error.Code != "EPD_TICKET_OR_INCIDENT_MISSING" {
		t.Fatalf("expected EPD_TICKET_OR_INCIDENT_MISSING, got %+v", result.Error)
	}
}

// @hlv EPD_TTL_EXCEEDED
func TestEvaluateEmergencyPolicyDisable_TTLExceeded(t *testing.T) {
	result := EvaluateEmergencyPolicyDisable(EmergencyPolicyDisableInput{Scope: "global", TicketID: "INC-1", IncidentID: "P0-1", TriggerSignal: "five_xx_spike", Approvers: []string{"IncidentCommander", "SecurityLead"}, CurrentDisableMinute: 61, RequestID: "req-epd-4"})
	if result.Error == nil || result.Error.Code != "EPD_TTL_EXCEEDED" {
		t.Fatalf("expected EPD_TTL_EXCEEDED, got %+v", result.Error)
	}
}

// @hlv epd_ttl_max_60
func TestEPDProperty_TTLCannotExceedLimitWithoutEscalation(t *testing.T) {
	for minute := 0; minute <= 80; minute++ {
		result := EvaluateEmergencyPolicyDisable(EmergencyPolicyDisableInput{Scope: "global", TicketID: "INC-1", IncidentID: "P0-1", TriggerSignal: "five_xx_spike", Approvers: []string{"IncidentCommander", "SecurityLead"}, CurrentDisableMinute: minute, RequestID: "req-epd-ttl"})
		if minute <= 60 && result.Status != "activated" {
			t.Fatalf("expected minute=%d to activate", minute)
		}
		if minute > 60 && (result.Error == nil || result.Error.Code != "EPD_TTL_EXCEEDED") {
			t.Fatalf("expected minute=%d to fail with EPD_TTL_EXCEEDED", minute)
		}
	}
}

// @hlv epd_allowed_triggers_only
func TestEPDProperty_OnlyAllowedTriggersActivateFlow(t *testing.T) {
	triggers := []string{"five_xx_spike", "auth_failure_spike", "false_block_confirmed", "other"}
	for _, trigger := range triggers {
		result := EvaluateEmergencyPolicyDisable(EmergencyPolicyDisableInput{Scope: "global", TicketID: "INC-1", IncidentID: "P0-1", TriggerSignal: trigger, Approvers: []string{"IncidentCommander", "SecurityLead"}, RequestID: "req-epd-trigger"})
		if trigger == "other" && (result.Error == nil || result.Error.Code != "EPD_TRIGGER_NOT_MET") {
			t.Fatalf("expected disallowed trigger to fail")
		}
		if trigger != "other" && result.Status != "activated" {
			t.Fatalf("expected allowed trigger %s to activate", trigger)
		}
	}
}

// @hlv epd_smoke_suite_required
func TestEPDProperty_ActivationAlwaysRequiresSmokeSuite(t *testing.T) {
	result := EvaluateEmergencyPolicyDisable(EmergencyPolicyDisableInput{Scope: "global", TicketID: "INC-1", IncidentID: "P0-1", TriggerSignal: "five_xx_spike", Approvers: []string{"IncidentCommander", "SecurityLead"}, RequestID: "req-epd-smoke"})
	if result.Status != "activated" {
		t.Fatal("expected activated status")
	}
	if !result.PostDisableSmokeRequired {
		t.Fatal("post disable smoke suite must be required")
	}
}
