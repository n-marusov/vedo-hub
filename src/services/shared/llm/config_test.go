package llm

// Validates: REQ-FUN.API.llm-policy

import (
	"os"
	"testing"
	"time"
)

func TestConfigDefaults(t *testing.T) {
	// Ensure no env vars are set
	unsetEnvVars(t,
		"LLM_PROVIDER",
		"LLM_API_KEY",
		"LLM_MODEL",
		"LLM_BASE_URL",
		"LLM_TIMEOUT",
		"LLM_MAX_TOKENS",
		"LLM_TEMPERATURE",
	)

	cfg := LoadConfigFromEnv()

	if cfg.Provider != "openai" {
		t.Errorf("expected default provider 'openai', got '%s'", cfg.Provider)
	}
	if cfg.Model != "gpt-4o" {
		t.Errorf("expected default model 'gpt-4o', got '%s'", cfg.Model)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", cfg.Timeout)
	}
	if cfg.MaxTokens != 4096 {
		t.Errorf("expected default max tokens 4096, got %d", cfg.MaxTokens)
	}
	if cfg.Temperature != 0.7 {
		t.Errorf("expected default temperature 0.7, got %f", cfg.Temperature)
	}
}

func TestConfigFromEnv(t *testing.T) {
	setEnvVars(t, map[string]string{
		"LLM_PROVIDER":    "ollama",
		"LLM_API_KEY":     "",
		"LLM_MODEL":       "llama3",
		"LLM_BASE_URL":    "http://localhost:11434/v1",
		"LLM_TIMEOUT":     "60s",
		"LLM_MAX_TOKENS":  "2048",
		"LLM_TEMPERATURE": "0.5",
	})
	defer unsetEnvVars(t,
		"LLM_PROVIDER",
		"LLM_API_KEY",
		"LLM_MODEL",
		"LLM_BASE_URL",
		"LLM_TIMEOUT",
		"LLM_MAX_TOKENS",
		"LLM_TEMPERATURE",
	)

	cfg := LoadConfigFromEnv()

	if cfg.Provider != "ollama" {
		t.Errorf("expected provider 'ollama', got '%s'", cfg.Provider)
	}
	if cfg.APIKey != "" {
		t.Errorf("expected empty API key, got '%s'", cfg.APIKey)
	}
	if cfg.Model != "llama3" {
		t.Errorf("expected model 'llama3', got '%s'", cfg.Model)
	}
	if cfg.BaseURL != "http://localhost:11434/v1" {
		t.Errorf("expected base URL 'http://localhost:11434/v1', got '%s'", cfg.BaseURL)
	}
	if cfg.Timeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", cfg.Timeout)
	}
	if cfg.MaxTokens != 2048 {
		t.Errorf("expected max tokens 2048, got %d", cfg.MaxTokens)
	}
	if cfg.Temperature != 0.5 {
		t.Errorf("expected temperature 0.5, got %f", cfg.Temperature)
	}
}

func TestConfigInvalidTimeout(t *testing.T) {
	setEnvVars(t, map[string]string{
		"LLM_TIMEOUT": "invalid",
	})
	defer unsetEnvVars(t, "LLM_TIMEOUT")

	cfg := LoadConfigFromEnv()
	// Should fall back to default on parse error
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s on invalid input, got %v", cfg.Timeout)
	}
}

func TestConfigInvalidMaxTokens(t *testing.T) {
	setEnvVars(t, map[string]string{
		"LLM_MAX_TOKENS": "not-a-number",
	})
	defer unsetEnvVars(t, "LLM_MAX_TOKENS")

	cfg := LoadConfigFromEnv()
	if cfg.MaxTokens != 4096 {
		t.Errorf("expected default max tokens 4096 on invalid input, got %d", cfg.MaxTokens)
	}
}

func TestConfigInvalidTemperature(t *testing.T) {
	setEnvVars(t, map[string]string{
		"LLM_TEMPERATURE": "invalid",
	})
	defer unsetEnvVars(t, "LLM_TEMPERATURE")

	cfg := LoadConfigFromEnv()
	if cfg.Temperature != 0.7 {
		t.Errorf("expected default temperature 0.7 on invalid input, got %f", cfg.Temperature)
	}
}

func TestConfigClampedTemperature(t *testing.T) {
	setEnvVars(t, map[string]string{
		"LLM_TEMPERATURE": "2.5",
	})
	defer unsetEnvVars(t, "LLM_TEMPERATURE")

	cfg := LoadConfigFromEnv()
	if cfg.Temperature != 2.0 {
		t.Errorf("expected clamped temperature 2.0, got %f", cfg.Temperature)
	}
}

func TestConfigNegativeTemperature(t *testing.T) {
	setEnvVars(t, map[string]string{
		"LLM_TEMPERATURE": "-1.0",
	})
	defer unsetEnvVars(t, "LLM_TEMPERATURE")

	cfg := LoadConfigFromEnv()
	if cfg.Temperature != 0.0 {
		t.Errorf("expected clamped temperature 0.0, got %f", cfg.Temperature)
	}
}

// Helper: set multiple env vars
func setEnvVars(t *testing.T, vars map[string]string) {
	t.Helper()
	for k, v := range vars {
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("failed to set env var %s: %v", k, err)
		}
	}
}

// Helper: unset multiple env vars
func unsetEnvVars(t *testing.T, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if err := os.Unsetenv(k); err != nil {
			t.Fatalf("failed to unset env var %s: %v", k, err)
		}
	}
}
