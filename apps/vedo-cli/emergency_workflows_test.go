package cli

import "testing"

// @ctx: stage 2 emergency workflow tests for readonly/clear and rotation scaffolding

// @hlv EPD_TICKET_OR_INCIDENT_MISSING
func TestStartEmergencyReadonlyWorkflow_TicketAndIncidentRequired(t *testing.T) {
	resetMFAChallengesForTests()
	out := StartEmergencyReadonlyWorkflow(CliInput{
		Command:        CmdEmergencyReadonly,
		OutputFormat:   FormatJSON,
		ActorType:      ActorUser,
		MFAChallengeID: "mfa-readonly-1",
		RequestID:      "req-emergency-1",
		TriggerSignal:  "five_xx_spike",
		Approvers:      []string{"IncidentCommander", "SecurityLead"},
	}, "trace-1", "corr-1", []string{"typed_confirmation", "env_guard", "mfa"})
	if out.Error == nil || out.Error.Code != "EPD_TICKET_OR_INCIDENT_MISSING" {
		t.Fatalf("expected EPD_TICKET_OR_INCIDENT_MISSING, got %+v", out.Error)
	}
}

// @hlv EPD_DUAL_APPROVAL_MISSING
func TestClearEmergencyReadonlyWorkflow_DualApprovalRequired(t *testing.T) {
	out := ClearEmergencyReadonlyWorkflow(CliInput{
		Command:      CmdEmergencyClear,
		OutputFormat: FormatJSON,
		TicketID:     "INC-001",
		IncidentID:   "P0-001",
		Approvers:    []string{"IncidentCommander"},
		RequestID:    "req-emergency-2",
	}, "trace-2", "corr-2", []string{"typed_confirmation", "env_guard", "mfa", "cooldown_delay"})
	if out.Error == nil || out.Error.Code != "EPD_DUAL_APPROVAL_MISSING" {
		t.Fatalf("expected EPD_DUAL_APPROVAL_MISSING, got %+v", out.Error)
	}
}

// @hlv EPD_DUAL_APPROVAL_MISSING
// @hlv EPD_TICKET_OR_INCIDENT_MISSING
func TestEmergencyReadonlyAndClear_HappyPath(t *testing.T) {
	resetMFAChallengesForTests()
	start := StartEmergencyReadonlyWorkflow(CliInput{
		Command:          CmdEmergencyReadonly,
		OutputFormat:     FormatJSON,
		TicketID:         "INC-013",
		IncidentID:       "P0-556",
		TriggerSignal:    "five_xx_spike",
		Approvers:        []string{"IncidentCommander", "SecurityLead"},
		SecretCategories: []string{"service_to_service_credentials", "external_api_keys"},
		RequestID:        "req-emergency-3",
	}, "trace-3", "corr-3", []string{"typed_confirmation", "env_guard", "mfa"})
	if start.Status != "ok" || start.Data == nil {
		t.Fatalf("expected readonly start to succeed, got %+v", start.Error)
	}
	if start.Data.Result == "" {
		t.Fatal("expected readonly result payload")
	}

	clear := ClearEmergencyReadonlyWorkflow(CliInput{
		Command:      CmdEmergencyClear,
		OutputFormat: FormatJSON,
		TicketID:     "INC-013",
		IncidentID:   "P0-556",
		Approvers:    []string{"IncidentCommander", "SecurityLead"},
		RequestID:    "req-emergency-4",
	}, "trace-4", "corr-4", []string{"typed_confirmation", "env_guard", "mfa", "cooldown_delay"})
	if clear.Status != "ok" || clear.Data == nil {
		t.Fatalf("expected clear to succeed, got %+v", clear.Error)
	}
}

// @hlv rotation_sla_mapping_4_8_24
func TestBuildIncidentSecretInventory_MapsCategoriesToSLA(t *testing.T) {
	inventory := BuildIncidentSecretInventory("req-inv-1", "P0-778", []string{"cicd_tokens_registry_signing", "service_to_service_credentials", "customer_session_secrets"})
	if len(inventory.Records) != 3 {
		t.Fatalf("expected 3 inventory records, got %d", len(inventory.Records))
	}
	expected := []int{4, 8, 24}
	for idx, record := range inventory.Records {
		if record.RotationSLAHours != expected[idx] {
			t.Fatalf("record %d expected SLA %d, got %d", idx, expected[idx], record.RotationSLAHours)
		}
	}
}

// @hlv rotation_success_requires_full_coverage
func TestBuildRotationPlanScaffold_CreatesStepPerInventoryRecord(t *testing.T) {
	inventory := IncidentSecretInventory{
		IncidentID: "P0-779",
		Records: []IncidentSecretRecord{
			{SecretID: "P0-779-secret-1", SecretCategory: "service_to_service_credentials", RotationSLAHours: 8},
			{SecretID: "P0-779-secret-2", SecretCategory: "external_api_keys", RotationSLAHours: 24},
		},
	}
	plan := BuildRotationPlanScaffold("req-plan-1", inventory)
	if len(plan.Steps) != len(inventory.Records) {
		t.Fatalf("expected %d rotation steps, got %d", len(inventory.Records), len(plan.Steps))
	}
	if len(plan.SmokeSuite) != 3 {
		t.Fatalf("expected smoke suite size 3, got %d", len(plan.SmokeSuite))
	}
}
