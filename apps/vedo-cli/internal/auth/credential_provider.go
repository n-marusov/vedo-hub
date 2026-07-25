package auth

import (
	"errors"
	"log/slog"
	"time"
)

// @ctx: credentials provider chain for CLI-OPS-001 contract
// Deterministic order: Vault → AWS Secrets Manager → environment variables → explicit error

type CredentialProvider string

const (
	ProviderVault              CredentialProvider = "vault"
	ProviderAWSSecretsManager  CredentialProvider = "aws_secrets_manager"
	ProviderEnv                CredentialProvider = "env"
	ProviderNone               CredentialProvider = ""
)

type ResolvedCredential struct {
	Provider CredentialProvider
	Value    string
	Source   string
}

// ResolveCredentials follows the deterministic provider chain order.
// @hlv:sec [SECRET_HANDLING] — credentials resolved from external sources
// @hlv:sec [AUTH_BOUNDARY] — credential access boundary
func ResolveCredentials(vaultAddr, awsRegion, envPrefix string) (*ResolvedCredential, error) {
	start := time.Now()
	slog.Info("auth.resolve.enter",
		"vault_configured", vaultAddr != "",
		"aws_configured", awsRegion != "",
		"env_prefix", envPrefix,
	)
	defer func() {
		slog.Info("auth.resolve.exit", "duration_ms", time.Since(start).Milliseconds())
	}()

	// @ctx: try Vault first
	if vaultAddr != "" {
		slog.Info("auth.resolve.trying_vault", "addr", redactAddr(vaultAddr))
		cred, err := tryVault(vaultAddr)
		if err == nil {
			slog.Info("auth.resolve.vault_selected")
			return cred, nil
		}
		slog.Warn("auth.resolve.vault_failed", "error", err.Error())
	}

	// @ctx: try AWS Secrets Manager second
	if awsRegion != "" {
		slog.Info("auth.resolve.trying_aws", "region", awsRegion)
		cred, err := tryAWS(awsRegion)
		if err == nil {
			slog.Info("auth.resolve.aws_selected")
			return cred, nil
		}
		slog.Warn("auth.resolve.aws_failed", "error", err.Error())
	}

	// @ctx: try environment variables third
	if envPrefix != "" {
		slog.Info("auth.resolve.trying_env", "prefix", envPrefix)
		cred, err := tryEnv(envPrefix)
		if err == nil {
			slog.Info("auth.resolve.env_selected")
			return cred, nil
		}
		slog.Warn("auth.resolve.env_failed", "error", err.Error())
	}

	// @ctx: no provider resolved — explicit error
	slog.Error("auth.resolve.no_provider")
	// @hlv CLI_CREDENTIALS_NOT_CONFIGURED
	return nil, errors.New("CLI_CREDENTIALS_NOT_CONFIGURED: no credentials source configured")
}

// tryVault attempts to resolve credentials from HashiCorp Vault.
// @hlv:sec [SECRET_HANDLING] — Vault secret retrieved
func tryVault(addr string) (*ResolvedCredential, error) {
	slog.Debug("auth.vault.resolve", "addr", redactAddr(addr))
	return nil, errors.New("vault not available in stub mode")
}

// tryAWS attempts to resolve credentials from AWS Secrets Manager.
// @hlv:sec [SECRET_HANDLING] — AWS secret retrieved
func tryAWS(region string) (*ResolvedCredential, error) {
	slog.Debug("auth.aws.resolve", "region", region)
	return nil, errors.New("aws not available in stub mode")
}

// tryEnv attempts to resolve credentials from environment variables with given prefix.
// @hlv:sec [SECRET_HANDLING] — env var read, must not be logged
func tryEnv(prefix string) (*ResolvedCredential, error) {
	slog.Debug("auth.env.resolve", "prefix", prefix)
	return nil, errors.New("env not configured in stub mode")
}

// @hlv:sec [SECRET_HANDLING] — redact sensitive portions of addresses
func redactAddr(addr string) string {
	if len(addr) > 8 {
		return addr[:4] + "..." + addr[len(addr)-4:]
	}
	return "***"
}
