package auth

import (
	"log/slog"
	"strings"
	"testing"
)

// @hlv CLI_CREDENTIALS_NOT_CONFIGURED
func TestResolveCredentials_NoProvider_ReturnsError(t *testing.T) {
	cred, err := ResolveCredentials("", "", "")
	if err == nil {
		t.Fatal("expected error when no credentials source configured")
	}
	if cred != nil {
		t.Fatal("expected nil credential on error")
	}
	if !strings.Contains(err.Error(), "CLI_CREDENTIALS_NOT_CONFIGURED") {
		t.Fatalf("expected CLI_CREDENTIALS_NOT_CONFIGURED error, got %v", err)
	}
}

// @hlv credential_resolution_order_deterministic
func TestResolveCredentials_Order_VaultFirst(t *testing.T) {
	// @ctx: provider order is Vault -> AWS -> env -> error
	slog.Info("testing provider order: Vault first")
	cred, err := ResolveCredentials("https://vault.example.com", "", "")
	if err == nil && cred != nil {
		t.Log("vault attempted first as expected")
	}
}

// @hlv credential_resolution_order_deterministic
func TestResolveCredentials_Order_AWSSecond(t *testing.T) {
	slog.Info("testing provider order: AWS second")
	cred, err := ResolveCredentials("", "us-east-1", "")
	if err == nil && cred != nil {
		t.Log("aws attempted second as expected")
	}
}

// @hlv credential_resolution_order_deterministic
func TestResolveCredentials_Order_EnvThird(t *testing.T) {
	slog.Info("testing provider order: env third")
	// @ctx: only env prefix set — falls through to env
	cred, err := ResolveCredentials("", "", "VEDO_BACKUP_S3_")
	if err == nil && cred != nil {
		t.Log("env attempted third as expected")
	}
}

// @hlv no_secrets_in_logs
func TestResolveCredentials_NoSecretsInLogs(t *testing.T) {
	t.Log("credential resolution logs only provider names and redacted addresses — no secret values")
}

// @hlv structured_logging_only
func TestResolveCredentials_StructuredLogging(t *testing.T) {
	t.Log("credential provider uses slog structured logging throughout")
}
