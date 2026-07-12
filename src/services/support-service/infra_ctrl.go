package supportservice

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// @hlv:artifact support-service-code implements CTR-006-007
// @ctx: Vault/KMS integration stubs for INFRA-CTRL-001

type InfraCtrlInput struct {
	Operation            string
	KeyID                string
	TenantID             string
	KeyType              string
	RotationScheduleDays *int
	RequestID            string
}

type KeyMetadataRecord struct {
	KeyID                 string    `json:"key_id"`
	TenantID              string    `json:"tenant_id"`
	KeyType               string    `json:"key_type"`
	Backend               string    `json:"backend"`
	RotationScheduleDays  int       `json:"rotation_schedule_days"`
	AccessAuditRef        string    `json:"access_audit_ref"`
	LastRotatedAt         time.Time `json:"last_rotated_at"`
	Revoked               bool      `json:"revoked"`
	PrivilegedAccessClass string    `json:"privileged_access_class"`
}

type InfraCtrlOutput struct {
	Status   string            `json:"status"`
	KeyID    string            `json:"key_id"`
	Backend  string            `json:"backend"`
	Metadata KeyMetadataRecord `json:"metadata"`
}

type KeyControlStubStore struct {
	mu            sync.RWMutex
	vaultUp       bool
	kmsUp         bool
	preferBackend string
	keysByID      map[string]KeyMetadataRecord
}

func NewKeyControlStubStore() *KeyControlStubStore {
	return &KeyControlStubStore{
		vaultUp:       true,
		kmsUp:         true,
		preferBackend: "vault",
		keysByID:      map[string]KeyMetadataRecord{},
	}
}

func (s *KeyControlStubStore) SetBackendAvailability(requestID string, backend string, available bool) *SupportError {
	if backend != "vault" && backend != "kms" {
		return &SupportError{Code: "INFRA_CTRL_BACKEND_UNAVAILABLE", Message: "unsupported backend target"}
	}
	s.mu.Lock()
	old := s.vaultUp
	if backend == "vault" {
		s.vaultUp = available
		old = !available
	}
	if backend == "kms" {
		old = s.kmsUp
		s.kmsUp = available
	}
	s.mu.Unlock()
	slog.Info("support.infra.backend.state_changed", "request_id", requestID, "entity_id", backend, "old", old, "new", available, "event", "state changed")
	return nil
}

func (s *KeyControlStubStore) SetPreferredBackend(requestID string, backend string) *SupportError {
	if backend != "vault" && backend != "kms" {
		return &SupportError{Code: "INFRA_CTRL_BACKEND_UNAVAILABLE", Message: "unsupported backend target"}
	}
	s.mu.Lock()
	old := s.preferBackend
	s.preferBackend = backend
	s.mu.Unlock()
	slog.Info("support.infra.backend.preference.state_changed", "request_id", requestID, "entity_id", "backend_preference", "old", old, "new", backend, "event", "state changed")
	return nil
}

// @hlv:sec [AUTH_BOUNDARY] - privileged key metadata operations require approved support path
// @hlv:sec [SECRET_HANDLING] - key metadata outputs must never include plaintext key material
func (s *KeyControlStubStore) Execute(input *InfraCtrlInput) (InfraCtrlOutput, *SupportError) {
	start := time.Now()
	slog.Info("support.infra.ctrl.enter", "request_id", input.RequestID, "entity_id", input.KeyID, "operation", input.Operation)
	defer func() {
		slog.Info("support.infra.ctrl.exit", "request_id", input.RequestID, "entity_id", input.KeyID, "duration_ms", time.Since(start).Milliseconds())
	}()

	if input.Operation != "create_key_stub" && input.Operation != "get_key_metadata" && input.Operation != "rotate_key_stub" && input.Operation != "revoke_key_stub" {
		slog.Error("support.infra.ctrl.failed", "request_id", input.RequestID, "entity_id", input.KeyID, "input_summary", "unsupported operation", "error_code", "INFRA_CTRL_BACKEND_UNAVAILABLE")
		return InfraCtrlOutput{}, &SupportError{Code: "INFRA_CTRL_BACKEND_UNAVAILABLE", Message: "unsupported operation for key control stub"}
	}

	schedule := 90
	if input.RotationScheduleDays != nil {
		schedule = *input.RotationScheduleDays
		if schedule <= 0 {
			slog.Error("support.infra.ctrl.failed", "request_id", input.RequestID, "entity_id", input.KeyID, "input_summary", fmt.Sprintf("rotation_schedule_days=%d", schedule), "error_code", "INFRA_CTRL_ROTATION_POLICY_INVALID")
			return InfraCtrlOutput{}, &SupportError{Code: "INFRA_CTRL_ROTATION_POLICY_INVALID", Message: "rotation schedule must be greater than zero"}
		}
	}

	backend, backendErr := s.resolveBackend(input.RequestID)
	if backendErr != nil {
		return InfraCtrlOutput{}, backendErr
	}

	if input.Operation == "create_key_stub" {
		record := KeyMetadataRecord{
			KeyID:                 input.KeyID,
			TenantID:              input.TenantID,
			KeyType:               input.KeyType,
			Backend:               backend,
			RotationScheduleDays:  schedule,
			AccessAuditRef:        fmt.Sprintf("worm://audit/%s", input.KeyID),
			LastRotatedAt:         time.Now().UTC(),
			PrivilegedAccessClass: "emergency",
		}
		s.mu.Lock()
		s.keysByID[input.KeyID] = record
		s.mu.Unlock()
		slog.Info("support.infra.key.state_changed", "request_id", input.RequestID, "entity_id", input.KeyID, "old", "missing", "new", "created", "event", "state changed")
		return InfraCtrlOutput{Status: "ok", KeyID: input.KeyID, Backend: backend, Metadata: record}, nil
	}

	s.mu.RLock()
	record, ok := s.keysByID[input.KeyID]
	s.mu.RUnlock()
	if !ok {
		slog.Error("support.infra.ctrl.failed", "request_id", input.RequestID, "entity_id", input.KeyID, "input_summary", "missing key id", "error_code", "INFRA_CTRL_KEY_NOT_FOUND")
		return InfraCtrlOutput{}, &SupportError{Code: "INFRA_CTRL_KEY_NOT_FOUND", Message: "key metadata not found"}
	}

	if input.Operation == "get_key_metadata" {
		return InfraCtrlOutput{Status: "ok", KeyID: record.KeyID, Backend: backend, Metadata: record}, nil
	}

	if input.Operation == "rotate_key_stub" {
		record.Backend = backend
		record.LastRotatedAt = time.Now().UTC()
		record.AccessAuditRef = fmt.Sprintf("worm://audit/%s/rotate", input.KeyID)
		s.mu.Lock()
		s.keysByID[input.KeyID] = record
		s.mu.Unlock()
		slog.Info("support.infra.key.state_changed", "request_id", input.RequestID, "entity_id", input.KeyID, "old", "active", "new", "rotated", "event", "state changed")
		return InfraCtrlOutput{Status: "ok", KeyID: record.KeyID, Backend: backend, Metadata: record}, nil
	}

	record.Backend = backend
	record.Revoked = true
	record.AccessAuditRef = fmt.Sprintf("worm://audit/%s/revoke", input.KeyID)
	s.mu.Lock()
	s.keysByID[input.KeyID] = record
	s.mu.Unlock()
	slog.Info("support.infra.key.state_changed", "request_id", input.RequestID, "entity_id", input.KeyID, "old", "active", "new", "revoked", "event", "state changed")
	return InfraCtrlOutput{Status: "ok", KeyID: record.KeyID, Backend: backend, Metadata: record}, nil
}

// @hlv:sec [NETWORK] - backend selection simulates Vault/KMS connectivity and failover paths
func (s *KeyControlStubStore) resolveBackend(requestID string) (string, *SupportError) {
	s.mu.RLock()
	vaultUp := s.vaultUp
	kmsUp := s.kmsUp
	prefer := s.preferBackend
	s.mu.RUnlock()
	if prefer == "vault" {
		if vaultUp {
			slog.Info("support.infra.backend.call", "request_id", requestID, "entity_id", "vault", "target", "vault", "outcome", "available")
			return "vault", nil
		}
		if kmsUp {
			slog.Warn("support.infra.backend.call", "request_id", requestID, "entity_id", "kms", "target", "kms", "outcome", "failover")
			return "kms", nil
		}
	}
	if prefer == "kms" {
		if kmsUp {
			slog.Info("support.infra.backend.call", "request_id", requestID, "entity_id", "kms", "target", "kms", "outcome", "available")
			return "kms", nil
		}
		if vaultUp {
			slog.Warn("support.infra.backend.call", "request_id", requestID, "entity_id", "vault", "target", "vault", "outcome", "failover")
			return "vault", nil
		}
	}
	slog.Error("support.infra.backend.call.failed", "request_id", requestID, "entity_id", "backend", "input_summary", "vault and kms unavailable", "error_code", "INFRA_CTRL_BACKEND_UNAVAILABLE")
	return "", &SupportError{Code: "INFRA_CTRL_BACKEND_UNAVAILABLE", Message: "Vault/KMS backend is unavailable"}
}
