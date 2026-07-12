package cli

import (
	"fmt"
	"log/slog"
	"time"
)

// @ctx: stage 2 emergency readonly/clear workflow scaffolding for SEC-EPD-001 and SEC-SECRET-ROT-001

type IncidentSecretRecord struct {
	SecretID         string `json:"secret_id"`
	SecretCategory   string `json:"secret_category"`
	RotationSLAHours int    `json:"rotation_sla_hours"`
}

type IncidentSecretInventory struct {
	IncidentID  string                 `json:"incident_id"`
	GeneratedAt string                 `json:"generated_at"`
	Records     []IncidentSecretRecord `json:"records"`
}

type RotationPlanStep struct {
	SecretID       string `json:"secret_id"`
	SecretCategory string `json:"secret_category"`
	Action         string `json:"action"`
	DeadlineHours  int    `json:"deadline_hours"`
}

type RotationPlanScaffold struct {
	IncidentID string             `json:"incident_id"`
	Steps      []RotationPlanStep `json:"steps"`
	SmokeSuite []string           `json:"smoke_suite"`
}

type EmergencyWorkflowState struct {
	Mode               string `json:"mode"`
	PolicyDisableState string `json:"policy_disable_state"`
	InventoryCount     int    `json:"inventory_count"`
	RotationStepCount  int    `json:"rotation_step_count"`
}

// @hlv:sec [AUTH_BOUNDARY] — readonly activation requires dual approval via emergency policy disable contract
func StartEmergencyReadonlyWorkflow(input CliInput, traceID string, correlationID string, controls []string) CliOutput {
	start := time.Now()
	slog.Info("cli.emergency.readonly.enter",
		"request_id", input.RequestID,
		"trace_id", traceID,
		"entity_id", input.IncidentID,
		"command", input.Command,
	)
	defer func() {
		slog.Info("cli.emergency.readonly.exit",
			"request_id", input.RequestID,
			"trace_id", traceID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}()

	epd := EvaluateEmergencyPolicyDisable(EmergencyPolicyDisableInput{
		Scope:         "global",
		TicketID:      input.TicketID,
		IncidentID:    input.IncidentID,
		TriggerSignal: input.TriggerSignal,
		Approvers:     input.Approvers,
		RequestID:     input.RequestID,
	})
	if epd.Error != nil {
		slog.Error("cli.emergency.readonly.activate_failed",
			"request_id", input.RequestID,
			"trace_id", traceID,
			"entity_id", input.IncidentID,
			"input_summary", summarizeInput(input),
			"error_code", epd.Error.Code,
		)
		return CliOutput{Status: "error", Error: epd.Error}
	}

	inventory := BuildIncidentSecretInventory(input.RequestID, input.IncidentID, input.SecretCategories)
	rotation := BuildRotationPlanScaffold(input.RequestID, inventory)

	slog.Info("cli.emergency.readonly.state_changed",
		"request_id", input.RequestID,
		"entity_id", input.IncidentID,
		"old", "normal",
		"new", "emergency_readonly",
		"event", "state changed",
	)

	workflowState := EmergencyWorkflowState{
		Mode:               "emergency_readonly",
		PolicyDisableState: epd.Status,
		InventoryCount:     len(inventory.Records),
		RotationStepCount:  len(rotation.Steps),
	}

	return CliOutput{
		Status: "ok",
		Data: &CliData{
			Command:               string(input.Command),
			Result:                fmt.Sprintf("readonly-mode-active inventory=%d rotation_steps=%d", workflowState.InventoryCount, workflowState.RotationStepCount),
			TraceID:               traceID,
			CorrelationID:         correlationID,
			GuardrailsApplied:     controls,
			MFAChallengePerformed: true,
		},
	}
}

// @hlv:sec [AUTH_BOUNDARY] — clear operation requires dual approval and references active incident context
func ClearEmergencyReadonlyWorkflow(input CliInput, traceID string, correlationID string, controls []string) CliOutput {
	start := time.Now()
	slog.Info("cli.emergency.clear.enter",
		"request_id", input.RequestID,
		"trace_id", traceID,
		"entity_id", input.IncidentID,
	)
	defer func() {
		slog.Info("cli.emergency.clear.exit",
			"request_id", input.RequestID,
			"trace_id", traceID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}()

	if input.TicketID == "" || input.IncidentID == "" {
		// @hlv EPD_TICKET_OR_INCIDENT_MISSING
		return CliOutput{Status: "error", Error: &CliError{Code: "EPD_TICKET_OR_INCIDENT_MISSING", Message: "ticket and incident are required to clear emergency mode"}}
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
		// @hlv EPD_DUAL_APPROVAL_MISSING
		return CliOutput{Status: "error", Error: &CliError{Code: "EPD_DUAL_APPROVAL_MISSING", Message: "dual approval required for emergency clear"}}
	}

	slog.Info("cli.emergency.clear.state_changed",
		"request_id", input.RequestID,
		"entity_id", input.IncidentID,
		"old", "emergency_readonly",
		"new", "normal",
		"event", "state changed",
	)

	return CliOutput{
		Status: "ok",
		Data: &CliData{
			Command:               string(input.Command),
			Result:                "emergency-mode-cleared",
			TraceID:               traceID,
			CorrelationID:         correlationID,
			GuardrailsApplied:     controls,
			MFAChallengePerformed: true,
		},
	}
}

// @hlv:sec [SECRET_HANDLING] — incident secret inventory scaffold enumerates impacted secret categories only
func BuildIncidentSecretInventory(requestID string, incidentID string, categories []string) IncidentSecretInventory {
	start := time.Now()
	slog.Info("cli.emergency.inventory.enter", "request_id", requestID, "entity_id", incidentID)
	defer func() {
		slog.Info("cli.emergency.inventory.exit", "request_id", requestID, "entity_id", incidentID, "duration_ms", time.Since(start).Milliseconds())
	}()

	if len(categories) == 0 {
		categories = []string{"service_to_service_credentials"}
	}
	records := make([]IncidentSecretRecord, 0, len(categories))
	for i, category := range categories {
		slaHours, ok := resolveRotationSLA(category)
		if !ok {
			slaHours = 24
		}
		records = append(records, IncidentSecretRecord{
			SecretID:         fmt.Sprintf("%s-secret-%d", incidentID, i+1),
			SecretCategory:   category,
			RotationSLAHours: slaHours,
		})
	}
	return IncidentSecretInventory{IncidentID: incidentID, GeneratedAt: time.Now().UTC().Format(time.RFC3339), Records: records}
}

func BuildRotationPlanScaffold(requestID string, inventory IncidentSecretInventory) RotationPlanScaffold {
	steps := make([]RotationPlanStep, 0, len(inventory.Records))
	for _, record := range inventory.Records {
		steps = append(steps, RotationPlanStep{
			SecretID:       record.SecretID,
			SecretCategory: record.SecretCategory,
			Action:         "rotate_and_revoke_old_secret",
			DeadlineHours:  record.RotationSLAHours,
		})
	}

	slog.Info("cli.emergency.rotation_plan.created",
		"request_id", requestID,
		"entity_id", inventory.IncidentID,
		"target", "secret-rotation-plan",
		"outcome", "created",
	)

	return RotationPlanScaffold{
		IncidentID: inventory.IncidentID,
		Steps:      steps,
		SmokeSuite: []string{"auth", "api", "background jobs"},
	}
}
