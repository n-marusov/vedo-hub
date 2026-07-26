package auth

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// captureSlog replaces the default slog logger with a text handler writing to
// a buffer for the duration of fn, returning the captured output. The original
// logger is restored automatically when the test finishes.
func captureSlog(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	original := slog.Default()
	slog.SetDefault(slog.New(handler))
	t.Cleanup(func() { slog.SetDefault(original) })
	fn()
	return buf.String()
}

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
	// @ctx: provider order is Vault -> AWS -> env -> error.
	// With a fake vault address, tryVault is attempted first and fails, then
	// the chain falls through to the no-provider error (AWS/env empty).
	output := captureSlog(t, func() {
		_, err := ResolveCredentials("https://vault.example.com", "", "")
		if err == nil {
			t.Fatal("expected error: vault stub must fail in test mode")
		}
	})
	if !strings.Contains(output, "auth.resolve.trying_vault") {
		t.Fatalf("expected vault to be tried first, got: %s", output)
	}
	// The vault address logged must be redacted, never the raw URL.
	if strings.Contains(output, "https://vault.example.com") {
		t.Fatalf("vault address was logged unredacted: %s", output)
	}
}

// @hlv credential_resolution_order_deterministic
func TestResolveCredentials_Order_AWSSecond(t *testing.T) {
	output := captureSlog(t, func() {
		_, err := ResolveCredentials("", "us-east-1", "")
		if err == nil {
			t.Fatal("expected error: aws stub must fail in test mode")
		}
	})
	if !strings.Contains(output, "auth.resolve.trying_aws") {
		t.Fatalf("expected aws to be tried, got: %s", output)
	}
	// Vault must NOT be attempted when no vault address is configured.
	if strings.Contains(output, "auth.resolve.trying_vault") {
		t.Fatalf("vault should not be tried when vaultAddr is empty, got: %s", output)
	}
}

// @hlv credential_resolution_order_deterministic
func TestResolveCredentials_Order_EnvThird(t *testing.T) {
	// @ctx: only env prefix set — falls through to env, then error
	output := captureSlog(t, func() {
		_, err := ResolveCredentials("", "", "VEDO_BACKUP_S3_")
		if err == nil {
			t.Fatal("expected error: env stub must fail in test mode")
		}
	})
	if !strings.Contains(output, "auth.resolve.trying_env") {
		t.Fatalf("expected env to be tried, got: %s", output)
	}
	if strings.Contains(output, "auth.resolve.trying_vault") {
		t.Fatalf("vault should not be tried when vaultAddr is empty, got: %s", output)
	}
	if strings.Contains(output, "auth.resolve.trying_aws") {
		t.Fatalf("aws should not be tried when awsRegion is empty, got: %s", output)
	}
}

// @hlv no_secrets_in_logs
func TestResolveCredentials_NoSecretsInLogs(t *testing.T) {
	// @ctx: credential resolution logs only provider names and redacted
	// addresses — the secret value (if any) never appears in log output.
	// The vault address is redacted to "http...host" form by redactAddr.
	secretAddr := "https://vault.example.com/secret-path"
	output := captureSlog(t, func() {
		_, _ = ResolveCredentials(secretAddr, "", "")
	})
	// The full address must not appear verbatim — redactAddr masks the middle.
	if strings.Contains(output, secretAddr) {
		t.Fatalf("vault address leaked unredacted into log: %s", output)
	}
	// The "addr=" field must be present (shows the redacted form was logged).
	if !strings.Contains(output, "addr=") {
		t.Fatalf("expected redacted addr field in log, got: %s", output)
	}
}

// @hlv structured_logging_only
func TestResolveCredentials_StructuredLogging(t *testing.T) {
	output := captureSlog(t, func() {
		_, _ = ResolveCredentials("https://vault.example.com", "us-east-1", "VEDO_")
	})
	// Structured logging emits key=value pairs, not interpolated strings.
	for _, key := range []string{"vault_configured=true", "aws_configured=true", "env_prefix=VEDO_"} {
		if !strings.Contains(output, key) {
			t.Fatalf("expected structured field %q in log output, got: %s", key, output)
		}
	}
}
