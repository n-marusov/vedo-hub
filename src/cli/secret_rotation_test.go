package cli

import "testing"

// @ctx: SEC-SECRET-ROT-001 contract and property tests

// @hlv ROTATION_TRIGGER_INVALID
func TestEvaluateSecretRotation_InvalidTrigger(t *testing.T) {
	result := EvaluateSecretRotation(SecretRotationInput{IncidentID: "P0-1", Trigger: "manual_refresh", SecretCategory: "service_to_service_credentials", RotatedPercent: 100, OldSecretsRevokedWithinMinutes: 10, RequestID: "req-rot-1"})
	if result.Error == nil || result.Error.Code != "ROTATION_TRIGGER_INVALID" {
		t.Fatalf("expected ROTATION_TRIGGER_INVALID, got %+v", result.Error)
	}
}

// @hlv ROTATION_SCOPE_INCOMPLETE
func TestEvaluateSecretRotation_IncompleteCoverage(t *testing.T) {
	result := EvaluateSecretRotation(SecretRotationInput{IncidentID: "P0-1", Trigger: "leaked_credentials", SecretCategory: "service_to_service_credentials", RotatedPercent: 82, OldSecretsRevokedWithinMinutes: 10, RequestID: "req-rot-2"})
	if result.Error == nil || result.Error.Code != "ROTATION_SCOPE_INCOMPLETE" {
		t.Fatalf("expected ROTATION_SCOPE_INCOMPLETE, got %+v", result.Error)
	}
}

// @hlv ROTATION_SLA_BREACH
func TestEvaluateSecretRotation_SLABreach(t *testing.T) {
	result := EvaluateSecretRotation(SecretRotationInput{IncidentID: "P0-1", Trigger: "leaked_credentials", SecretCategory: "cicd_tokens_registry_signing", ElapsedHours: 5, RotatedPercent: 100, OldSecretsRevokedWithinMinutes: 10, RequestID: "req-rot-3"})
	if result.Error == nil || result.Error.Code != "ROTATION_SLA_BREACH" {
		t.Fatalf("expected ROTATION_SLA_BREACH, got %+v", result.Error)
	}
}

// @hlv ROTATION_EXCEPTION_EXPIRED
func TestEvaluateSecretRotation_ExceptionExpired(t *testing.T) {
	result := EvaluateSecretRotation(SecretRotationInput{IncidentID: "P0-1", Trigger: "leaked_credentials", SecretCategory: "service_to_service_credentials", ExceptionApprovalIssuedHoursAgo: 25, RotatedPercent: 100, OldSecretsRevokedWithinMinutes: 10, RequestID: "req-rot-4"})
	if result.Error == nil || result.Error.Code != "ROTATION_EXCEPTION_EXPIRED" {
		t.Fatalf("expected ROTATION_EXCEPTION_EXPIRED, got %+v", result.Error)
	}
}

// @hlv rotation_sla_mapping_4_8_24
func TestRotationProperty_SLAByCategory(t *testing.T) {
	cases := map[string]int{
		"cicd_tokens_registry_signing":   4,
		"service_to_service_credentials": 8,
		"external_api_keys":              24,
		"customer_session_secrets":       24,
	}
	for category, expected := range cases {
		result := EvaluateSecretRotation(SecretRotationInput{IncidentID: "P0-1", Trigger: "leaked_credentials", SecretCategory: category, ElapsedHours: expected, RotatedPercent: 100, OldSecretsRevokedWithinMinutes: 10, RequestID: "req-rot-sla"})
		if result.Status != "completed" {
			t.Fatalf("expected completed for category %s", category)
		}
		if result.RotationSLAHours != expected {
			t.Fatalf("expected category %s to map to %d hours", category, expected)
		}
	}
}

// @hlv rotation_old_secret_revocation_15_min
func TestRotationProperty_OldSecretsRevokedWithinFifteenMinutes(t *testing.T) {
	for revocationMinutes := 1; revocationMinutes <= 20; revocationMinutes++ {
		result := EvaluateSecretRotation(SecretRotationInput{IncidentID: "P0-1", Trigger: "leaked_credentials", SecretCategory: "service_to_service_credentials", RotatedPercent: 100, OldSecretsRevokedWithinMinutes: revocationMinutes, RequestID: "req-rot-revoke"})
		if revocationMinutes <= 15 && result.Status != "completed" {
			t.Fatalf("expected %d minutes to pass", revocationMinutes)
		}
		if revocationMinutes > 15 && (result.Error == nil || result.Error.Code != "ROTATION_SCOPE_INCOMPLETE") {
			t.Fatalf("expected %d minutes to fail", revocationMinutes)
		}
	}
}

// @hlv rotation_success_requires_full_coverage
func TestRotationProperty_SuccessRequiresOneHundredPercentCoverage(t *testing.T) {
	for _, coverage := range []float64{0, 25, 60, 100} {
		result := EvaluateSecretRotation(SecretRotationInput{IncidentID: "P0-1", Trigger: "leaked_credentials", SecretCategory: "service_to_service_credentials", RotatedPercent: coverage, OldSecretsRevokedWithinMinutes: 10, RequestID: "req-rot-coverage"})
		if coverage == 100 && result.Status != "completed" {
			t.Fatal("coverage 100 should complete")
		}
		if coverage < 100 && (result.Error == nil || result.Error.Code != "ROTATION_SCOPE_INCOMPLETE") {
			t.Fatalf("coverage %.0f must fail with ROTATION_SCOPE_INCOMPLETE", coverage)
		}
	}
}
