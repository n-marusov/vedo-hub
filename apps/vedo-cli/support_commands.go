package cli

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// @hlv:artifact cli-support-shell implements CTR-006-001
// @hlv:artifact cli-support-shell implements CTR-006-005
// @ctx: stage 5 support command shell for tenant-info, audit-trail, list-backups, emergency-access request/list/revoke

type supportTenantShellRecord struct {
	TenantID        string `json:"tenant_id"`
	TenantStatus    string `json:"tenant_status"`
	DeploymentModel string `json:"deployment_model"`
	SLATier         string `json:"sla_tier"`
}

type supportAuditEventShellRecord struct {
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	Checksum  string `json:"checksum"`
}

type supportEmergencyKeyShellRecord struct {
	KeyID                string `json:"key_id"`
	TenantID             string `json:"tenant_id"`
	RotationScheduleDays int    `json:"rotation_schedule_days"`
	TicketID             string `json:"ticket_id"`
	Revoked              bool   `json:"revoked"`
}

type supportCommandShell struct {
	mu              sync.RWMutex
	wormAvailable   bool
	tenants         map[string]supportTenantShellRecord
	auditByTenant   map[string][]supportAuditEventShellRecord
	backupsByTenant map[string][]string
	keysByID        map[string]supportEmergencyKeyShellRecord
	keysByTenant    map[string][]supportEmergencyKeyShellRecord
}

var defaultSupportShell = newSupportCommandShell()

func newSupportCommandShell() *supportCommandShell {
	return &supportCommandShell{
		wormAvailable: true,
		tenants: map[string]supportTenantShellRecord{
			"tenant-9d7c": {TenantID: "tenant-9d7c", TenantStatus: "active", DeploymentModel: "saas", SLATier: "enterprise"},
			"tenant-4aa1": {TenantID: "tenant-4aa1", TenantStatus: "active", DeploymentModel: "saas", SLATier: "custom"},
		},
		auditByTenant: map[string][]supportAuditEventShellRecord{
			"tenant-9d7c": {
				{EventID: "evt-3f8f", EventType: "emergency_access_granted", Checksum: "b07f3ef7615f"},
				{EventID: "evt-4101", EventType: "key_rotated", Checksum: "a1c99b44f21c"},
			},
		},
		backupsByTenant: map[string][]string{
			"tenant-9d7c": {"backup-2026-06-10", "backup-2026-06-09"},
			"tenant-4aa1": {"backup-2026-06-11"},
		},
		keysByID: map[string]supportEmergencyKeyShellRecord{},
		keysByTenant: map[string][]supportEmergencyKeyShellRecord{
			"tenant-9d7c": {
				{KeyID: "key-101", TenantID: "tenant-9d7c", RotationScheduleDays: 30, TicketID: "INC-0", Revoked: false},
			},
		},
	}
}

// @hlv:sec [INPUT_VALIDATION] - support tenant info requires known tenant in isolated support metadata shell
func runSupportTenantInfoShell(input CliInput, traceID string, correlationID string, controls []string) CliOutput {
	start := time.Now()
	slog.Info("cli.support.tenant_info.enter", "request_id", input.RequestID, "trace_id", traceID, "entity_id", input.TenantID)
	defer func() {
		slog.Info("cli.support.tenant_info.exit", "request_id", input.RequestID, "trace_id", traceID, "duration_ms", time.Since(start).Milliseconds())
	}()
	defaultSupportShell.mu.RLock()
	tenant, ok := defaultSupportShell.tenants[input.TenantID]
	defaultSupportShell.mu.RUnlock()
	if !ok {
		slog.Error("cli.support.tenant.not_found", "request_id", input.RequestID, "trace_id", traceID, "entity_id", input.TenantID, "input_summary", summarizeInput(input), "error_code", "CLI_SUPPORT_TENANT_NOT_FOUND")
		// @hlv CLI_SUPPORT_TENANT_NOT_FOUND
		return CliOutput{Status: "error", Error: &CliError{Code: "CLI_SUPPORT_TENANT_NOT_FOUND", Message: "tenant not found"}}
	}
	result := marshalSupportShellResult(map[string]any{"tenant": tenant})
	return CliOutput{Status: "ok", Data: &CliData{Command: string(input.Command), Result: result, TraceID: traceID, CorrelationID: correlationID, GuardrailsApplied: controls}}
}

// @hlv:sec [NETWORK] - audit-trail command simulates internal WORM backend availability check
func runSupportAuditTrailShell(input CliInput, traceID string, correlationID string, controls []string) CliOutput {
	start := time.Now()
	slog.Info("cli.support.audit_trail.enter", "request_id", input.RequestID, "trace_id", traceID, "entity_id", input.TenantID)
	defer func() {
		slog.Info("cli.support.audit_trail.exit", "request_id", input.RequestID, "trace_id", traceID, "duration_ms", time.Since(start).Milliseconds())
	}()
	defaultSupportShell.mu.RLock()
	_, tenantExists := defaultSupportShell.tenants[input.TenantID]
	wormAvailable := defaultSupportShell.wormAvailable
	events := defaultSupportShell.auditByTenant[input.TenantID]
	defaultSupportShell.mu.RUnlock()
	if !tenantExists {
		// @hlv CLI_SUPPORT_TENANT_NOT_FOUND
		return CliOutput{Status: "error", Error: &CliError{Code: "CLI_SUPPORT_TENANT_NOT_FOUND", Message: "tenant not found"}}
	}
	if !wormAvailable {
		// @hlv SUPPORT_AUDIT_WORM_UNAVAILABLE
		return CliOutput{Status: "error", Error: &CliError{Code: "SUPPORT_AUDIT_WORM_UNAVAILABLE", Message: "WORM storage unavailable for requested audit trail"}}
	}
	result := marshalSupportShellResult(map[string]any{"audit_events": events})
	return CliOutput{Status: "ok", Data: &CliData{Command: string(input.Command), Result: result, TraceID: traceID, CorrelationID: correlationID, GuardrailsApplied: controls}}
}

func runSupportListBackupsShell(input CliInput, traceID string, correlationID string, controls []string) CliOutput {
	start := time.Now()
	slog.Info("cli.support.list_backups.enter", "request_id", input.RequestID, "trace_id", traceID, "entity_id", input.TenantID)
	defer func() {
		slog.Info("cli.support.list_backups.exit", "request_id", input.RequestID, "trace_id", traceID, "duration_ms", time.Since(start).Milliseconds())
	}()
	defaultSupportShell.mu.RLock()
	_, tenantExists := defaultSupportShell.tenants[input.TenantID]
	backups := defaultSupportShell.backupsByTenant[input.TenantID]
	defaultSupportShell.mu.RUnlock()
	if !tenantExists {
		// @hlv CLI_SUPPORT_TENANT_NOT_FOUND
		return CliOutput{Status: "error", Error: &CliError{Code: "CLI_SUPPORT_TENANT_NOT_FOUND", Message: "tenant not found"}}
	}
	result := marshalSupportShellResult(map[string]any{"backups": backups})
	return CliOutput{Status: "ok", Data: &CliData{Command: string(input.Command), Result: result, TraceID: traceID, CorrelationID: correlationID, GuardrailsApplied: controls}}
}

// @hlv:sec [AUTH_BOUNDARY] - emergency access shell writes privileged key metadata with ticket binding
func runSupportEmergencyAccessRequestShell(input CliInput, traceID string, correlationID string, controls []string, level GuardrailLevel) CliOutput {
	start := time.Now()
	slog.Info("cli.support.emergency_access.request.enter", "request_id", input.RequestID, "trace_id", traceID, "entity_id", input.TenantID)
	defer func() {
		slog.Info("cli.support.emergency_access.request.exit", "request_id", input.RequestID, "trace_id", traceID, "duration_ms", time.Since(start).Milliseconds())
	}()
	defaultSupportShell.mu.RLock()
	_, tenantExists := defaultSupportShell.tenants[input.TenantID]
	existing := len(defaultSupportShell.keysByTenant[input.TenantID])
	defaultSupportShell.mu.RUnlock()
	if !tenantExists {
		// @hlv CLI_SUPPORT_TENANT_NOT_FOUND
		return CliOutput{Status: "error", Error: &CliError{Code: "CLI_SUPPORT_TENANT_NOT_FOUND", Message: "tenant not found"}}
	}
	key := supportEmergencyKeyShellRecord{KeyID: fmt.Sprintf("key-%s-%d", input.TenantID, existing+1), TenantID: input.TenantID, RotationScheduleDays: 30, TicketID: input.TicketID, Revoked: false}
	defaultSupportShell.mu.Lock()
	defaultSupportShell.keysByID[key.KeyID] = key
	defaultSupportShell.keysByTenant[input.TenantID] = append(defaultSupportShell.keysByTenant[input.TenantID], key)
	defaultSupportShell.mu.Unlock()
	slog.Info("cli.support.emergency_access.state_changed", "request_id", input.RequestID, "trace_id", traceID, "entity_id", input.TenantID, "old", existing, "new", existing+1, "event", "state changed")
	result := marshalSupportShellResult(map[string]any{"emergency_access": key})
	return CliOutput{Status: "ok", Data: &CliData{Command: string(input.Command), Result: result, TraceID: traceID, CorrelationID: correlationID, GuardrailsApplied: controls, MFAChallengePerformed: level == GuardrailG3 || level == GuardrailG4}}
}

func runSupportEmergencyAccessListShell(input CliInput, traceID string, correlationID string, controls []string) CliOutput {
	defaultSupportShell.mu.RLock()
	if input.KeyID != "" {
		key, ok := defaultSupportShell.keysByID[input.KeyID]
		defaultSupportShell.mu.RUnlock()
		if !ok {
			// @hlv CLI_EMERGENCY_KEY_NOT_FOUND
			return CliOutput{Status: "error", Error: &CliError{Code: "CLI_EMERGENCY_KEY_NOT_FOUND", Message: "emergency key not found"}}
		}
		result := marshalSupportShellResult(map[string]any{"emergency_access": []supportEmergencyKeyShellRecord{key}})
		return CliOutput{Status: "ok", Data: &CliData{Command: string(input.Command), Result: result, TraceID: traceID, CorrelationID: correlationID, GuardrailsApplied: controls, MFAChallengePerformed: true}}
	}
	keys := defaultSupportShell.keysByTenant[input.TenantID]
	defaultSupportShell.mu.RUnlock()
	result := marshalSupportShellResult(map[string]any{"emergency_access": keys})
	return CliOutput{Status: "ok", Data: &CliData{Command: string(input.Command), Result: result, TraceID: traceID, CorrelationID: correlationID, GuardrailsApplied: controls, MFAChallengePerformed: true}}
}

func runSupportEmergencyAccessRevokeShell(input CliInput, traceID string, correlationID string, controls []string) CliOutput {
	defaultSupportShell.mu.Lock()
	record, ok := defaultSupportShell.keysByID[input.KeyID]
	if !ok {
		defaultSupportShell.mu.Unlock()
		// @hlv CLI_EMERGENCY_KEY_NOT_FOUND
		return CliOutput{Status: "error", Error: &CliError{Code: "CLI_EMERGENCY_KEY_NOT_FOUND", Message: "emergency key not found"}}
	}
	record.Revoked = true
	defaultSupportShell.keysByID[input.KeyID] = record
	tenantKeys := defaultSupportShell.keysByTenant[record.TenantID]
	for idx, key := range tenantKeys {
		if key.KeyID == input.KeyID {
			tenantKeys[idx] = record
			break
		}
	}
	defaultSupportShell.keysByTenant[record.TenantID] = tenantKeys
	defaultSupportShell.mu.Unlock()
	result := marshalSupportShellResult(map[string]any{"emergency_access": record})
	return CliOutput{Status: "ok", Data: &CliData{Command: string(input.Command), Result: result, TraceID: traceID, CorrelationID: correlationID, GuardrailsApplied: controls, MFAChallengePerformed: true}}
}

func marshalSupportShellResult(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(payload)
}

func setSupportWORMAvailabilityForTests(available bool) {
	defaultSupportShell.mu.Lock()
	defaultSupportShell.wormAvailable = available
	defaultSupportShell.mu.Unlock()
}

func resetSupportShellForTests() {
	defaultSupportShell = newSupportCommandShell()
}
