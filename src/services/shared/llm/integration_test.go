//go:build integration

package llm

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// LLM integration tests — real HTTP calls to configured LLM backend.
//
// These tests are SKIPPED by default. Run with:
//
//	LLM_INTEGRATION_TESTS=1 go test -tags=integration ./src/services/shared/llm/...
//
// Configuration is loaded from environment variables (see LoadConfigFromEnv).

func skipIfNotIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("LLM_INTEGRATION_TESTS") != "1" {
		t.Skip("LLM_INTEGRATION_TESTS not set; skipping integration test")
	}
}

func TestIntegration_ProviderCompletion(t *testing.T) {
	skipIfNotIntegration(t)

	cfg := LoadConfigFromEnv()
	if cfg.Provider == "" {
		t.Fatal("LLM_PROVIDER must be set for integration tests")
	}
	if cfg.APIKey == "" {
		t.Fatal("LLM_API_KEY must be set for integration tests")
	}

	provider, err := NewProvider(cfg.Provider)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	ctx := Context{
		BaseCtx:    context.Background(),
		MaxRetries: 2,
	}

	prompt := Prompt{
		Messages: []Message{
			{Role: RoleSystem, Content: "You are a helpful assistant."},
			{Role: RoleUser, Content: "Say hello in one word."},
		},
		MaxTokens: 50,
	}

	completion, err := provider.Complete(ctx, prompt)
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	t.Logf("Completion: %s (tokens: %d)", completion.Text, completion.TokensUsed)
	if completion.Text == "" {
		t.Error("response should not be empty")
	}
	if completion.TokensUsed <= 0 {
		t.Error("token count should be > 0")
	}
}

func TestIntegration_ProviderStreamCompletion(t *testing.T) {
	skipIfNotIntegration(t)

	cfg := LoadConfigFromEnv()
	if cfg.Provider == "" {
		t.Fatal("LLM_PROVIDER must be set for integration tests")
	}
	if cfg.APIKey == "" {
		t.Fatal("LLM_API_KEY must be set for integration tests")
	}

	provider, err := NewProvider(cfg.Provider)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	ctx := Context{
		BaseCtx:    context.Background(),
		MaxRetries: 2,
	}

	prompt := Prompt{
		Messages: []Message{
			{Role: RoleSystem, Content: "You are a helpful assistant."},
			{Role: RoleUser, Content: "Count from 1 to 3."},
		},
		MaxTokens: 50,
	}

	tokenChan, err := provider.StreamComplete(ctx, prompt)
	if err != nil {
		t.Fatalf("StreamComplete failed: %v", err)
	}

	var received []string
	for token := range tokenChan {
		received = append(received, token.Text)
	}

	t.Logf("Streamed %d tokens", len(received))
	if len(received) == 0 {
		t.Error("should receive at least one token")
	}
	fullResponse := strings.Join(received, "")
	if !containsStr(fullResponse, "1") {
		t.Errorf("response should contain '1', got: %s", fullResponse)
	}
}

func TestIntegration_TimeoutHandling(t *testing.T) {
	skipIfNotIntegration(t)

	cfg := LoadConfigFromEnv()
	cfg.Timeout = 1 * time.Nanosecond // Force immediate timeout

	// Set env temporarily to test timeout configuration
	os.Setenv("LLM_TIMEOUT", "1ns")
	defer os.Unsetenv("LLM_TIMEOUT")

	provider, err := NewProvider(cfg.Provider)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	ctx := Context{
		BaseCtx:    context.Background(),
		MaxRetries: 1,
	}

	prompt := Prompt{
		Messages: []Message{
			{Role: RoleUser, Content: "Write a long essay about AI."},
		},
		MaxTokens: 10,
	}

	_, err = provider.Complete(ctx, prompt)
	if err == nil {
		t.Error("should fail with timeout")
	} else {
		t.Logf("Timeout error: %v", err)
	}
}

func TestIntegration_TokenCounting(t *testing.T) {
	skipIfNotIntegration(t)

	cfg := LoadConfigFromEnv()
	if cfg.Provider == "" {
		t.Fatal("LLM_PROVIDER must be set for integration tests")
	}
	if cfg.APIKey == "" {
		t.Fatal("LLM_API_KEY must be set for integration tests")
	}

	provider, err := NewProvider(cfg.Provider)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	ctx := Context{
		BaseCtx:    context.Background(),
		MaxRetries: 2,
	}

	prompt := Prompt{
		Messages: []Message{
			{Role: RoleSystem, Content: "You are a helpful assistant."},
			{Role: RoleUser, Content: "What is OWL in the context of the Semantic Web? Answer in 2-3 sentences."},
		},
		MaxTokens: 100,
	}

	completion, err := provider.Complete(ctx, prompt)
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	t.Logf("Prompt tokens: %d, Completion tokens (reported): %d",
		completion.PromptTokens, completion.TokensUsed-completion.PromptTokens)
	t.Logf("Total tokens: %d", completion.TokensUsed)

	if completion.TokensUsed <= 0 {
		t.Error("total token count should be > 0")
	}
	if completion.PromptTokens < 0 {
		t.Error("prompt tokens should be >= 0")
	}
	if completion.TokensUsed < completion.PromptTokens {
		t.Errorf("total (%d) should be >= prompt (%d)", completion.TokensUsed, completion.PromptTokens)
	}
}
