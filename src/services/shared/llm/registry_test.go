package llm

import (
	"os"
	"strings"
	"testing"
)

func TestRegisterAndNewProvider(t *testing.T) {
	Register("mock-test-reg", func(cfg Config) (Provider, error) {
		return &mockProvider{
			completeFunc: func(ctx Context, p Prompt) (Completion, error) {
				return Completion{Text: "mock", TokensUsed: 5}, nil
			},
		}, nil
	})

	provider, err := NewProvider("mock-test-reg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	comp, err := provider.Complete(Context{}, Prompt{Messages: []Message{{Role: RoleUser, Content: "hi"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comp.Text != "mock" {
		t.Errorf("expected 'mock', got '%s'", comp.Text)
	}
}

func TestRegistryUnknownProvider(t *testing.T) {
	_, err := NewProvider("non-existent-provider")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
	if !strings.Contains(err.Error(), "unknown provider") {
		t.Errorf("expected 'unknown provider' in error, got '%s'", err.Error())
	}
}

func TestRegistryDuplicatePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()

	Register("dup-test", func(cfg Config) (Provider, error) {
		return nil, nil
	})
	Register("dup-test", func(cfg Config) (Provider, error) {
		return nil, nil
	})
}

func TestRegistryDefaultFromEnv(t *testing.T) {
	os.Setenv("LLM_PROVIDER", "mock-env-test")
	defer os.Unsetenv("LLM_PROVIDER")

	Register("mock-env-test", func(cfg Config) (Provider, error) {
		return &mockProvider{}, nil
	})

	provider, err := NewProvider()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestAvailableProviders(t *testing.T) {
	available := AvailableProviders()
	if len(available) == 0 {
		t.Log("no providers registered (expected in isolation)")
	}
}

func TestProviderEnvOverrides(t *testing.T) {
	os.Setenv("LLM_PROVIDERS_OVERRIDE_TEST_BASE_URL", "http://custom:8080/v1")
	os.Setenv("LLM_PROVIDERS_OVERRIDE_TEST_API_KEY", "custom-key")
	os.Setenv("LLM_PROVIDERS_OVERRIDE_TEST_MODEL", "custom-model")
	defer func() {
		os.Unsetenv("LLM_PROVIDERS_OVERRIDE_TEST_BASE_URL")
		os.Unsetenv("LLM_PROVIDERS_OVERRIDE_TEST_API_KEY")
		os.Unsetenv("LLM_PROVIDERS_OVERRIDE_TEST_MODEL")
	}()

	cfg := LoadConfigFromEnv()
	cfg = applyProviderEnvOverrides("override-test", cfg)

	if cfg.BaseURL != "http://custom:8080/v1" {
		t.Errorf("expected custom base URL, got '%s'", cfg.BaseURL)
	}
	if cfg.APIKey != "custom-key" {
		t.Errorf("expected custom API key, got '%s'", cfg.APIKey)
	}
	if cfg.Model != "custom-model" {
		t.Errorf("expected custom model, got '%s'", cfg.Model)
	}
}
