package middleware

// Validates: REQ-FUN.API.llm-policy
// Validates: REQ-FUN.DATA.ontology-visibility-levels

import (
	"os"
	"testing"
)

// =============================================================================
// EvaluatePolicy tests — LLM access policy matrix
// =============================================================================

func TestEvaluatePolicy_AirGappedExternalProvider_Blocks(t *testing.T) {
	decision := EvaluatePolicy("air-gapped", "private", "external", false)
	if decision.Allowed {
		t.Errorf("expected blocked for air-gapped + external provider, got allowed: %s", decision.Reason)
	}
}

func TestEvaluatePolicy_AirGappedLocalProvider_Allows(t *testing.T) {
	decision := EvaluatePolicy("air-gapped", "private", "local", false)
	if !decision.Allowed {
		t.Errorf("expected allowed for air-gapped + local provider, got blocked: %s", decision.Reason)
	}
}

func TestEvaluatePolicy_PublicOntology_AlwaysAllowed(t *testing.T) {
	// Public ontology is always allowed EXCEPT in air-gapped mode with external provider,
	// where the air-gapped constraint takes precedence.
	cases := []struct {
		mode     string
		provider string
	}{
		{"saas", "external"},
		{"saas", "local"},
		{"on-premise", "external"},
		{"on-premise", "local"},
		{"air-gapped", "local"},
	}
	for _, tc := range cases {
		decision := EvaluatePolicy(tc.mode, "public", tc.provider, false)
		if !decision.Allowed {
			t.Errorf("expected allowed for public ontology (mode=%s, provider=%s), got blocked: %s",
				tc.mode, tc.provider, decision.Reason)
		}
	}
}

func TestEvaluatePolicy_OnPremise_AlwaysAllowed(t *testing.T) {
	visibilities := []string{"public", "internal", "private"}
	providers := []string{"external", "local"}
	for _, vis := range visibilities {
		for _, provider := range providers {
			decision := EvaluatePolicy("on-premise", vis, provider, false)
			if !decision.Allowed {
				t.Errorf("expected allowed for on-premise (vis=%s, provider=%s), got blocked: %s",
					vis, provider, decision.Reason)
			}
		}
	}
}

func TestEvaluatePolicy_LocalLLM_AlwaysAllowed(t *testing.T) {
	visibilities := []string{"public", "internal", "private"}
	modes := []string{"saas", "air-gapped"}
	for _, vis := range visibilities {
		for _, mode := range modes {
			decision := EvaluatePolicy(mode, vis, "local", false)
			if !decision.Allowed {
				t.Errorf("expected allowed for local LLM (mode=%s, vis=%s), got blocked: %s",
					mode, vis, decision.Reason)
			}
		}
	}
}

func TestEvaluatePolicy_SaaSExternalPrivate_BlocksWithoutAdmin(t *testing.T) {
	decision := EvaluatePolicy("saas", "private", "external", false)
	if decision.Allowed {
		t.Errorf("expected blocked for SaaS + external + private (non-admin), got allowed: %s", decision.Reason)
	}
}

func TestEvaluatePolicy_SaaSExternalInternal_BlocksWithoutAdmin(t *testing.T) {
	decision := EvaluatePolicy("saas", "internal", "external", false)
	if decision.Allowed {
		t.Errorf("expected blocked for SaaS + external + internal (non-admin), got blocked: %s", decision.Reason)
	}
}

func TestEvaluatePolicy_SaaSExternalPrivate_AllowsWithAdmin(t *testing.T) {
	decision := EvaluatePolicy("saas", "private", "external", true)
	if !decision.Allowed {
		t.Errorf("expected allowed for SaaS + external + private (admin), got blocked: %s", decision.Reason)
	}
}

func TestEvaluatePolicy_SaaSExternalInternal_AllowsWithAdmin(t *testing.T) {
	decision := EvaluatePolicy("saas", "internal", "external", true)
	if !decision.Allowed {
		t.Errorf("expected allowed for SaaS + external + internal (admin), got blocked: %s", decision.Reason)
	}
}

func TestEvaluatePolicy_SaaSExternalPublic_AlwaysAllowed(t *testing.T) {
	decision := EvaluatePolicy("saas", "public", "external", false)
	if !decision.Allowed {
		t.Errorf("expected allowed for SaaS + external + public, got blocked: %s", decision.Reason)
	}
}

// =============================================================================
// GetDeployMode tests
// =============================================================================

func TestGetDeployMode_Default_ReturnsSaaS(t *testing.T) {
	os.Unsetenv("DEPLOYMENT_MODE")
	mode := GetDeployMode()
	if mode != "saas" {
		t.Errorf("expected 'saas', got %q", mode)
	}
}

func TestGetDeployMode_OnPremise_ReturnsOnPremise(t *testing.T) {
	os.Setenv("DEPLOYMENT_MODE", "on-premise")
	defer os.Unsetenv("DEPLOYMENT_MODE")
	mode := GetDeployMode()
	if mode != "on-premise" {
		t.Errorf("expected 'on-premise', got %q", mode)
	}
}

func TestGetDeployMode_AirGapped_ReturnsAirGapped(t *testing.T) {
	os.Setenv("DEPLOYMENT_MODE", "air-gapped")
	defer os.Unsetenv("DEPLOYMENT_MODE")
	mode := GetDeployMode()
	if mode != "air-gapped" {
		t.Errorf("expected 'air-gapped', got %q", mode)
	}
}

func TestGetDeployMode_Invalid_DefaultsToSaaS(t *testing.T) {
	os.Setenv("DEPLOYMENT_MODE", "unknown-mode")
	defer os.Unsetenv("DEPLOYMENT_MODE")
	mode := GetDeployMode()
	if mode != "saas" {
		t.Errorf("expected 'saas' for unknown mode, got %q", mode)
	}
}

// =============================================================================
// GetProviderType tests
// =============================================================================

func TestGetProviderType_Default_ReturnsLocal(t *testing.T) {
	os.Unsetenv("LLM_PROVIDER")
	pt := GetProviderType()
	if pt != "local" {
		t.Errorf("expected 'local', got %q", pt)
	}
}

func TestGetProviderType_OpenAI_ReturnsExternal(t *testing.T) {
	os.Setenv("LLM_PROVIDER", "openai")
	defer os.Unsetenv("LLM_PROVIDER")
	pt := GetProviderType()
	if pt != "external" {
		t.Errorf("expected 'external', got %q", pt)
	}
}

func TestGetProviderType_Anthropic_ReturnsExternal(t *testing.T) {
	os.Setenv("LLM_PROVIDER", "anthropic")
	defer os.Unsetenv("LLM_PROVIDER")
	pt := GetProviderType()
	if pt != "external" {
		t.Errorf("expected 'external', got %q", pt)
	}
}

func TestGetProviderType_Ollama_ReturnsLocal(t *testing.T) {
	os.Setenv("LLM_PROVIDER", "ollama")
	defer os.Unsetenv("LLM_PROVIDER")
	pt := GetProviderType()
	if pt != "local" {
		t.Errorf("expected 'local', got %q", pt)
	}
}

func TestGetProviderType_Empty_ReturnsLocal(t *testing.T) {
	os.Setenv("LLM_PROVIDER", "")
	defer os.Unsetenv("LLM_PROVIDER")
	pt := GetProviderType()
	if pt != "local" {
		t.Errorf("expected 'local' for empty, got %q", pt)
	}
}

// =============================================================================
// isAIRPCMethod tests
// =============================================================================

func TestIsAIRPCMethod_GenerateOWL_ReturnsTrue(t *testing.T) {
	if !isAIRPCMethod("/vedo.ai_orchestration.v1.AIOrchestrationService/GenerateOWL") {
		t.Error("expected GenerateOWL to be an AI RPC method")
	}
}

func TestIsAIRPCMethod_NaturalLanguageQuery_ReturnsTrue(t *testing.T) {
	if !isAIRPCMethod("/vedo.ai_orchestration.v1.AIOrchestrationService/NaturalLanguageQuery") {
		t.Error("expected NaturalLanguageQuery to be an AI RPC method")
	}
}

func TestIsAIRPCMethod_RefineOntology_ReturnsTrue(t *testing.T) {
	if !isAIRPCMethod("/vedo.ai_orchestration.v1.AIOrchestrationService/RefineOntology") {
		t.Error("expected RefineOntology to be an AI RPC method")
	}
}

func TestIsAIRPCMethod_Complete_ReturnsTrue(t *testing.T) {
	if !isAIRPCMethod("/vedo.ai_orchestration.v1.AIOrchestrationService/Complete") {
		t.Error("expected Complete to be an AI RPC method")
	}
}

func TestIsAIRPCMethod_HealthCheck_ReturnsFalse(t *testing.T) {
	if isAIRPCMethod("/grpc.health.v1.Health/Check") {
		t.Error("expected Health/Check to NOT be an AI RPC method")
	}
}

func TestIsAIRPCMethod_EmptyString_ReturnsFalse(t *testing.T) {
	if isAIRPCMethod("") {
		t.Error("expected empty string to NOT be an AI RPC method")
	}
}

// =============================================================================
// extractOntologyID tests
// =============================================================================

type mockRequest struct {
	ontologyID string
}

func (m mockRequest) GetOntologyId() string {
	return m.ontologyID
}

type nonProtoRequest struct{}

func TestExtractOntologyID_ValidRequest_ReturnsID(t *testing.T) {
	req := mockRequest{ontologyID: "ont-123"}
	id := extractOntologyID(req)
	if id != "ont-123" {
		t.Errorf("expected 'ont-123', got %q", id)
	}
}

func TestExtractOntologyID_NonProtoType_ReturnsEmpty(t *testing.T) {
	req := nonProtoRequest{}
	id := extractOntologyID(req)
	if id != "" {
		t.Errorf("expected empty, got %q", id)
	}
}

func TestExtractOntologyID_EmptyID_ReturnsEmpty(t *testing.T) {
	req := mockRequest{ontologyID: ""}
	id := extractOntologyID(req)
	if id != "" {
		t.Errorf("expected empty, got %q", id)
	}
}
