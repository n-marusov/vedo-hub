package cli

import (
	"log/slog"
	"time"
)

// @ctx: SEC-SECRET-ROT-001 incident-driven rotation validation shell

type SecretRotationInput struct {
	IncidentID                      string
	Trigger                         string
	SecretCategory                  string
	RequestedBy                     string
	ExceptionApprovalIssuedHoursAgo int
	ElapsedHours                    int
	RotatedPercent                  float64
	OldSecretsRevokedWithinMinutes  int
	RequestID                       string
}

type SecretRotationOutput struct {
	Status                         string
	RotationSLAHours               int
	RotatedPercent                 float64
	OldSecretsRevokedWithinMinutes int
	SmokeSuite                     []string
	Error                          *CliError
}

// @hlv:sec [SECRET_HANDLING] — secret rotation path validates revocation and timeline requirements
func EvaluateSecretRotation(input SecretRotationInput) SecretRotationOutput {
	start := time.Now()
	slog.Info("rotation.evaluate.enter", "request_id", input.RequestID, "incident_id", input.IncidentID, "secret_category", input.SecretCategory)
	defer func() {
		slog.Info("rotation.evaluate.exit", "request_id", input.RequestID, "duration_ms", time.Since(start).Milliseconds())
	}()

	if input.Trigger != "cicd_compromise" && input.Trigger != "leaked_credentials" && input.Trigger != "compromised_workstation" {
		slog.Error("rotation.evaluate.trigger_invalid", "request_id", input.RequestID, "trigger", input.Trigger)
		// @hlv ROTATION_TRIGGER_INVALID
		return SecretRotationOutput{Status: "failed", Error: &CliError{Code: "ROTATION_TRIGGER_INVALID", Message: "trigger does not match mandatory rotation criteria"}}
	}

	if input.ExceptionApprovalIssuedHoursAgo > 24 {
		slog.Error("rotation.evaluate.exception_expired", "request_id", input.RequestID, "hours", input.ExceptionApprovalIssuedHoursAgo)
		// @hlv ROTATION_EXCEPTION_EXPIRED
		return SecretRotationOutput{Status: "failed", Error: &CliError{Code: "ROTATION_EXCEPTION_EXPIRED", Message: "rotation exception exceeded allowed approval window"}}
	}

	slaHours, ok := resolveRotationSLA(input.SecretCategory)
	if !ok {
		slog.Error("rotation.evaluate.scope_incomplete", "request_id", input.RequestID, "secret_category", input.SecretCategory)
		// @hlv ROTATION_SCOPE_INCOMPLETE
		return SecretRotationOutput{Status: "failed", Error: &CliError{Code: "ROTATION_SCOPE_INCOMPLETE", Message: "unknown category prevents complete coverage accounting"}}
	}

	if input.RotatedPercent < 100.0 {
		slog.Error("rotation.evaluate.scope_incomplete", "request_id", input.RequestID, "rotated_percent", input.RotatedPercent)
		// @hlv ROTATION_SCOPE_INCOMPLETE
		return SecretRotationOutput{Status: "failed", Error: &CliError{Code: "ROTATION_SCOPE_INCOMPLETE", Message: "rotation does not cover full affected secret set"}}
	}

	if input.ElapsedHours > slaHours {
		slog.Error("rotation.evaluate.sla_breach", "request_id", input.RequestID, "elapsed_hours", input.ElapsedHours, "sla_hours", slaHours)
		// @hlv ROTATION_SLA_BREACH
		return SecretRotationOutput{Status: "failed", Error: &CliError{Code: "ROTATION_SLA_BREACH", Message: "rotation exceeded category SLA"}}
	}

	if input.OldSecretsRevokedWithinMinutes > 15 {
		slog.Error("rotation.evaluate.scope_incomplete.revocation", "request_id", input.RequestID, "revocation_minutes", input.OldSecretsRevokedWithinMinutes)
		// @hlv ROTATION_SCOPE_INCOMPLETE
		return SecretRotationOutput{Status: "failed", Error: &CliError{Code: "ROTATION_SCOPE_INCOMPLETE", Message: "old secrets revocation exceeds 15 minute deadline"}}
	}

	slog.Info("rotation.evaluate.state_changed", "request_id", input.RequestID, "entity_id", input.IncidentID, "old", "started", "new", "completed", "event", "state changed")
	return SecretRotationOutput{
		Status:                         "completed",
		RotationSLAHours:               slaHours,
		RotatedPercent:                 100.0,
		OldSecretsRevokedWithinMinutes: input.OldSecretsRevokedWithinMinutes,
		SmokeSuite:                     []string{"auth", "api", "background jobs"},
	}
}

func resolveRotationSLA(category string) (int, bool) {
	switch category {
	case "cicd_tokens_registry_signing":
		return 4, true
	case "service_to_service_credentials":
		return 8, true
	case "external_api_keys", "customer_session_secrets":
		return 24, true
	default:
		return 0, false
	}
}
