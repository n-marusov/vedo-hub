package main

// Validates: REQ-FUN.API.llm-policy

import (
	"testing"
	"vedo-core/src/services/shared/llm"
)

// =============================================================================
// overrideSummary tests
// =============================================================================

func TestOverrideSummary_NilMap_ReturnsNone(t *testing.T) {
	result := overrideSummary(nil)
	if result != "<none>" {
		t.Errorf("expected '<none>', got %q", result)
	}
}

func TestOverrideSummary_EmptyMap_ReturnsNone(t *testing.T) {
	result := overrideSummary(map[string]string{})
	if result != "<none>" {
		t.Errorf("expected '<none>', got %q", result)
	}
}

func TestOverrideSummary_SinglePair_ReturnsKeyEqualsValue(t *testing.T) {
	result := overrideSummary(map[string]string{"model": "gpt-4"})
	if result != "model=gpt-4" {
		t.Errorf("expected 'model=gpt-4', got %q", result)
	}
}

func TestOverrideSummary_MultiplePairs_ReturnsJoined(t *testing.T) {
	m := map[string]string{
		"model":    "gpt-4",
		"provider": "openai",
	}
	result := overrideSummary(m)
	if result != "model=gpt-4,provider=openai" && result != "provider=openai,model=gpt-4" {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestOverrideSummary_EmptyValue_IncludesKey(t *testing.T) {
	result := overrideSummary(map[string]string{"key": ""})
	if result != "key=" {
		t.Errorf("expected 'key=', got %q", result)
	}
}

// =============================================================================
// getEnv tests
// =============================================================================

func TestGetEnv_NotSet_ReturnsFallback(t *testing.T) {
	t.Setenv("TEST_GETENV_UNSET", "")
	result := getEnv("TEST_GETENV_UNSET", "fallback")
	if result != "fallback" {
		t.Errorf("expected 'fallback', got %q", result)
	}
}

func TestGetEnv_Set_ReturnsValue(t *testing.T) {
	t.Setenv("TEST_GETENV_SET", "actual-value")
	result := getEnv("TEST_GETENV_SET", "fallback")
	if result != "actual-value" {
		t.Errorf("expected 'actual-value', got %q", result)
	}
}

func TestGetEnv_EmptyEnvVar_ReturnsFallback(t *testing.T) {
	// When env var is set to empty string, os.Getenv returns "" which is the same as not set
	t.Setenv("TEST_GETENV_EMPTY", "")
	result := getEnv("TEST_GETENV_EMPTY", "fallback")
	if result != "fallback" {
		t.Errorf("expected 'fallback', got %q", result)
	}
}

// =============================================================================
// getPort tests
// =============================================================================

func TestGetPort_NotSet_ReturnsDefaultWithColon(t *testing.T) {
	t.Setenv("TEST_PORT", "")
	result := getPort("TEST_PORT", "9090")
	if result != ":9090" {
		t.Errorf("expected ':9090', got %q", result)
	}
}

func TestGetPort_Set_ReturnsValueWithColon(t *testing.T) {
	t.Setenv("TEST_PORT", "8080")
	result := getPort("TEST_PORT", "9090")
	if result != ":8080" {
		t.Errorf("expected ':8080', got %q", result)
	}
}

// =============================================================================
// AIOrchestrationService tests (without gRPC — logic-only)
// =============================================================================

func TestIsProviderReady_BothNil_ReturnsError(t *testing.T) {
	svc := NewAIOrchestrationService(nil, nil)
	err := svc.isProviderReady()
	if err == nil {
		t.Error("expected error when both provider and template engine are nil")
	}
}

func TestIsProviderReady_ProviderOnly_ReturnsError(t *testing.T) {
	svc := NewAIOrchestrationService(&mockProvider{}, nil)
	err := svc.isProviderReady()
	if err == nil {
		t.Error("expected error when template engine is nil")
	}
}

func TestNewAIOrchestrationService_NilArgs_DoesNotPanic(t *testing.T) {
	svc := NewAIOrchestrationService(nil, nil)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.provider != nil {
		t.Error("expected nil provider")
	}
	if svc.templateEngine != nil {
		t.Error("expected nil template engine")
	}
}

// mockProvider is a minimal llm.Provider stub for compile-time checks.
type mockProvider struct{}

func (m *mockProvider) Complete(ctx llm.Context, p llm.Prompt) (llm.Completion, error) {
	return llm.Completion{}, nil
}
func (m *mockProvider) StreamComplete(ctx llm.Context, p llm.Prompt) (<-chan llm.Token, error) {
	ch := make(chan llm.Token)
	close(ch)
	return ch, nil
}
