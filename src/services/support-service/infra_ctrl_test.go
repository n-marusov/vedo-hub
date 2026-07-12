package supportservice

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

// @ctx: INFRA-CTRL-001 contract and property tests for Vault/KMS stubs

func TestCreateKeyMetadataStubSucceeds(t *testing.T) {
	silenceLogs()
	store := NewKeyControlStubStore()
	ninety := 90
	out, err := store.Execute(&InfraCtrlInput{Operation: "create_key_stub", KeyID: "key-11f7", TenantID: "tenant-9d7c", KeyType: "api_gateway_admin", RotationScheduleDays: &ninety, RequestID: "req-infra-1"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if out.Status != "ok" || out.Backend == "" {
		t.Fatalf("expected successful create output, got %+v", out)
	}
}

// @hlv INFRA_CTRL_KEY_NOT_FOUND
func TestUnknownKeyLookupReturnsNotFound(t *testing.T) {
	silenceLogs()
	store := NewKeyControlStubStore()
	_, err := store.Execute(&InfraCtrlInput{Operation: "get_key_metadata", KeyID: "key-missing", TenantID: "tenant-9d7c", RequestID: "req-infra-2"})
	if err == nil || err.Code != "INFRA_CTRL_KEY_NOT_FOUND" {
		t.Fatalf("expected INFRA_CTRL_KEY_NOT_FOUND, got %+v", err)
	}
}

// @hlv INFRA_CTRL_ROTATION_POLICY_INVALID
func TestInvalidRotationPolicyRejected(t *testing.T) {
	silenceLogs()
	store := NewKeyControlStubStore()
	zero := 0
	_, err := store.Execute(&InfraCtrlInput{Operation: "create_key_stub", KeyID: "key-77", TenantID: "tenant-a", KeyType: "pg_admin", RotationScheduleDays: &zero, RequestID: "req-infra-3"})
	if err == nil || err.Code != "INFRA_CTRL_ROTATION_POLICY_INVALID" {
		t.Fatalf("expected INFRA_CTRL_ROTATION_POLICY_INVALID, got %+v", err)
	}
}

// @hlv INFRA_CTRL_BACKEND_UNAVAILABLE
func TestBackendOutageReturnsAvailabilityError(t *testing.T) {
	silenceLogs()
	store := NewKeyControlStubStore()
	if err := store.SetBackendAvailability("req-infra-4", "vault", false); err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}
	if err := store.SetBackendAvailability("req-infra-5", "kms", false); err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}
	_, execErr := store.Execute(&InfraCtrlInput{Operation: "create_key_stub", KeyID: "key-88", TenantID: "tenant-a", KeyType: "neo4j_admin", RequestID: "req-infra-6"})
	if execErr == nil || execErr.Code != "INFRA_CTRL_BACKEND_UNAVAILABLE" {
		t.Fatalf("expected INFRA_CTRL_BACKEND_UNAVAILABLE, got %+v", execErr)
	}
}

// @hlv PBT-INFRA-CTRL-001
func TestMissingScheduleDefaultsToNinetyDays(t *testing.T) {
	silenceLogs()
	store := NewKeyControlStubStore()
	rng := rand.New(rand.NewSource(1001))
	for i := 0; i < 10000; i++ {
		keyID := fmt.Sprintf("key-default-%d-%d", i, rng.Int63())
		out, err := store.Execute(&InfraCtrlInput{Operation: "create_key_stub", KeyID: keyID, TenantID: "tenant-default", KeyType: "root_shell", RotationScheduleDays: nil, RequestID: "req-infra-pbt-1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out.Metadata.RotationScheduleDays != 90 {
			t.Fatalf("expected default schedule 90, got %d", out.Metadata.RotationScheduleDays)
		}
	}
}

// @hlv PBT-INFRA-CTRL-002
func TestBackendFailoverPreservesKeyIDs(t *testing.T) {
	silenceLogs()
	store := NewKeyControlStubStore()
	rng := rand.New(rand.NewSource(1002))
	for i := 0; i < 10000; i++ {
		keyID := fmt.Sprintf("key-failover-%d-%d", i, rng.Int63())
		_, createErr := store.Execute(&InfraCtrlInput{Operation: "create_key_stub", KeyID: keyID, TenantID: "tenant-failover", KeyType: "api_gateway_admin", RequestID: "req-infra-pbt-2-create"})
		if createErr != nil {
			t.Fatalf("unexpected create error: %v", createErr)
		}
		_ = store.SetBackendAvailability("req-infra-pbt-2-vault", "vault", false)
		_ = store.SetBackendAvailability("req-infra-pbt-2-kms", "kms", true)
		_, rotateErr := store.Execute(&InfraCtrlInput{Operation: "rotate_key_stub", KeyID: keyID, TenantID: "tenant-failover", RequestID: "req-infra-pbt-2-rotate"})
		if rotateErr != nil {
			t.Fatalf("unexpected rotate error: %v", rotateErr)
		}
		out, getErr := store.Execute(&InfraCtrlInput{Operation: "get_key_metadata", KeyID: keyID, TenantID: "tenant-failover", RequestID: "req-infra-pbt-2-get"})
		if getErr != nil {
			t.Fatalf("unexpected get error: %v", getErr)
		}
		if out.KeyID != keyID {
			t.Fatalf("expected key_id stable across failover, got %s want %s", out.KeyID, keyID)
		}
		_ = store.SetBackendAvailability("req-infra-pbt-2-vault-up", "vault", true)
	}
}

// @hlv PBT-INFRA-CTRL-003
func TestOutputsNeverIncludeKeyPlaintext(t *testing.T) {
	silenceLogs()
	store := NewKeyControlStubStore()
	rng := rand.New(rand.NewSource(1003))
	for i := 0; i < 10000; i++ {
		keyID := fmt.Sprintf("key-opaque-%d", i)
		out, err := store.Execute(&InfraCtrlInput{Operation: "create_key_stub", KeyID: keyID, TenantID: "tenant-opaque", KeyType: "pg_admin", RequestID: "req-infra-pbt-3"})
		if err != nil {
			t.Fatalf("unexpected create error: %v", err)
		}
		payload, marshalErr := json.Marshal(out)
		if marshalErr != nil {
			t.Fatalf("unexpected marshal error: %v", marshalErr)
		}
		token := fmt.Sprintf("plaintext-%d", rng.Int63())
		serialized := string(payload)
		if strings.Contains(serialized, token) || strings.Contains(strings.ToLower(serialized), "plaintext") || strings.Contains(strings.ToLower(serialized), "secret_material") {
			t.Fatalf("output leaked plaintext indicator: %s", serialized)
		}
	}
}
