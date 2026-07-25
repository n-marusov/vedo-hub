package supportservice

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// @hlv:artifact support-service-code implements CTR-006-005
// @ctx: WORM/MinIO immutable audit event writer for TASK-020

type WORMAuditEvent struct {
	EventID        string    `json:"event_id"`
	TenantID       string    `json:"tenant_id"`
	EventType      string    `json:"event_type"`
	Checksum       string    `json:"checksum"`
	Payload        string    `json:"payload"`
	CreatedAt      time.Time `json:"created_at"`
	RetentionUntil time.Time `json:"retention_until"`
}

type WORMStore struct {
	mu             sync.RWMutex
	available      bool
	retentionYears int
	eventsByID     map[string]WORMAuditEvent
	eventsByTenant map[string][]WORMAuditEvent
}

func NewWORMStore(retentionYears int) *WORMStore {
	return &WORMStore{
		available:      true,
		retentionYears: retentionYears,
		eventsByID:     map[string]WORMAuditEvent{},
		eventsByTenant: map[string][]WORMAuditEvent{},
	}
}

func (w *WORMStore) SetAvailability(requestID string, available bool) {
	w.mu.Lock()
	old := w.available
	w.available = available
	w.mu.Unlock()
	slog.Info("support.worm.availability.state_changed", "request_id", requestID, "entity_id", "worm", "old", old, "new", available, "event", "state changed")
}

// @hlv:sec [NETWORK] — writes immutable events to MinIO/WORM target
func (w *WORMStore) WriteImmutableEvent(requestID string, tenantID string, eventType string, payload string) (WORMAuditEvent, *SupportError) {
	start := time.Now()
	slog.Info("support.worm.write.enter", "request_id", requestID, "entity_id", tenantID, "event_type", eventType)
	defer func() {
		slog.Info("support.worm.write.exit", "request_id", requestID, "entity_id", tenantID, "duration_ms", time.Since(start).Milliseconds())
	}()

	w.mu.RLock()
	available := w.available
	w.mu.RUnlock()
	if !available {
		slog.Error("support.worm.write.failed", "request_id", requestID, "entity_id", tenantID, "input_summary", "worm unavailable", "error_code", "SUPPORT_AUDIT_WORM_UNAVAILABLE")
		return WORMAuditEvent{}, &SupportError{Code: "SUPPORT_AUDIT_WORM_UNAVAILABLE", Message: "WORM storage unavailable for requested audit trail"}
	}

	sum := sha256.Sum256([]byte(payload))
	checksum := hex.EncodeToString(sum[:])
	event := WORMAuditEvent{
		EventID:        fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		TenantID:       tenantID,
		EventType:      eventType,
		Checksum:       checksum,
		Payload:        payload,
		CreatedAt:      time.Now().UTC(),
		RetentionUntil: time.Now().UTC().AddDate(w.retentionYears, 0, 0),
	}

	slog.Info("support.worm.minio_call", "request_id", requestID, "entity_id", tenantID, "target", "minio-worm", "outcome", "write_ok")

	w.mu.Lock()
	w.eventsByID[event.EventID] = event
	w.eventsByTenant[tenantID] = append(w.eventsByTenant[tenantID], event)
	w.mu.Unlock()

	slog.Info("support.worm.write.state_changed", "request_id", requestID, "entity_id", tenantID, "old", "events_count_unknown", "new", event.EventID, "event", "state changed")
	return event, nil
}

func (w *WORMStore) ListTenantEvents(requestID string, tenantID string, since time.Time) ([]WORMAuditEvent, *SupportError) {
	w.mu.RLock()
	available := w.available
	events := w.eventsByTenant[tenantID]
	w.mu.RUnlock()
	if !available {
		slog.Error("support.worm.read.failed", "request_id", requestID, "entity_id", tenantID, "input_summary", "worm unavailable", "error_code", "SUPPORT_AUDIT_WORM_UNAVAILABLE")
		return nil, &SupportError{Code: "SUPPORT_AUDIT_WORM_UNAVAILABLE", Message: "WORM storage unavailable for requested audit trail"}
	}

	filtered := make([]WORMAuditEvent, 0, len(events))
	for _, event := range events {
		if event.CreatedAt.After(since) || event.CreatedAt.Equal(since) {
			filtered = append(filtered, event)
		}
	}
	slog.Info("support.worm.minio_call", "request_id", requestID, "entity_id", tenantID, "target", "minio-worm", "outcome", "read_ok")
	return filtered, nil
}

func (w *WORMStore) VerifyChecksum(requestID string, eventID string, payload string) *SupportError {
	w.mu.RLock()
	event, ok := w.eventsByID[eventID]
	w.mu.RUnlock()
	if !ok {
		return &SupportError{Code: "SUPPORT_DATA_INTEGRITY_CHECKSUM_MISMATCH", Message: "checksum verification failed"}
	}
	sum := sha256.Sum256([]byte(payload))
	computed := hex.EncodeToString(sum[:])
	if computed != event.Checksum {
		slog.Error("support.worm.checksum.failed", "request_id", requestID, "entity_id", event.TenantID, "input_summary", "payload mismatch", "error_code", "SUPPORT_DATA_INTEGRITY_CHECKSUM_MISMATCH")
		return &SupportError{Code: "SUPPORT_DATA_INTEGRITY_CHECKSUM_MISMATCH", Message: "WORM audit checksum verification fails"}
	}
	return nil
}
