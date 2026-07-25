package supportservice

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// @hlv:artifact support-service-code implements CTR-006-005
// @hlv:artifact support-service-code implements CTR-006-006
// @ctx: support metadata schema and SLA evaluation for SUPPORT-META-001 and SUPPORT-SLA-001

type SupportError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *SupportError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type SupportTenant struct {
	TenantID         string `json:"tenant_id"`
	TenantStatus     string `json:"tenant_status"`
	DeploymentModel  string `json:"deployment_model"`
	SLATier          string `json:"sla_tier"`
	EscalationPolicy string `json:"escalation_policy"`
}

type SupportContact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type IncidentRecord struct {
	IncidentID string    `json:"incident_id"`
	TenantID   string    `json:"tenant_id"`
	Severity   string    `json:"severity"`
	OpenedAt   time.Time `json:"opened_at"`
}

type ComplianceEvidence struct {
	EvidenceID   string `json:"evidence_id"`
	TenantID     string `json:"tenant_id"`
	ControlClass string `json:"control_class"`
	Reference    string `json:"reference"`
}

type ConfigOverride struct {
	TenantID  string `json:"tenant_id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Reason    string `json:"reason"`
	UpdatedBy string `json:"updated_by"`
	UpdatedAt string `json:"updated_at"`
}

type EmergencyAccessRecord struct {
	KeyID                 string `json:"key_id"`
	RotationScheduleDays  int    `json:"rotation_schedule_days"`
	RequestedBy           string `json:"requested_by"`
	TicketID              string `json:"ticket_id"`
	ApprovalPolicyMatched bool   `json:"approval_policy_matched"`
}

type SLAInput struct {
	TenantID         string
	SLATier          string
	IncidentSeverity string
	OpenedAt         time.Time
	FirstResponseAt  time.Time
	RequestID        string
}

type SLATarget struct {
	ResponseTimeMinutes int  `json:"response_time_minutes"`
	TTDMinutes          int  `json:"ttd_minutes"`
	TTRMinutes          int  `json:"ttr_minutes"`
	Support247          bool `json:"support_24_7"`
}

type SLAVerdict struct {
	Status       string    `json:"status"`
	Target       SLATarget `json:"sla_target"`
	BreachReason string    `json:"breach_reason,omitempty"`
}

type SupportStore struct {
	mu                sync.RWMutex
	tenants           map[string]SupportTenant
	contactsByTenant  map[string][]SupportContact
	incidentsByTenant map[string][]IncidentRecord
	evidenceByControl map[string][]ComplianceEvidence
	evidenceByTenant  map[string][]ComplianceEvidence
	emergencyByTenant map[string][]EmergencyAccessRecord
	overridesByTenant map[string][]ConfigOverride
}

func NewSupportStore() *SupportStore {
	return &SupportStore{
		tenants:           map[string]SupportTenant{},
		contactsByTenant:  map[string][]SupportContact{},
		incidentsByTenant: map[string][]IncidentRecord{},
		evidenceByControl: map[string][]ComplianceEvidence{},
		evidenceByTenant:  map[string][]ComplianceEvidence{},
		emergencyByTenant: map[string][]EmergencyAccessRecord{},
		overridesByTenant: map[string][]ConfigOverride{},
	}
}

func (s *SupportStore) UpsertTenant(requestID string, tenant SupportTenant) {
	start := time.Now()
	slog.Info("support.store.tenant.upsert.enter", "request_id", requestID, "entity_id", tenant.TenantID)
	defer func() {
		slog.Info("support.store.tenant.upsert.exit", "request_id", requestID, "entity_id", tenant.TenantID, "duration_ms", time.Since(start).Milliseconds())
	}()

	s.mu.Lock()
	old := s.tenants[tenant.TenantID]
	s.tenants[tenant.TenantID] = tenant
	s.mu.Unlock()

	slog.Info("support.store.tenant.state_changed", "request_id", requestID, "entity_id", tenant.TenantID, "old", old.SLATier, "new", tenant.SLATier, "event", "state changed")
}

func (s *SupportStore) GetTenantInfo(requestID string, tenantID string) (SupportTenant, *SupportError) {
	start := time.Now()
	slog.Info("support.store.tenant.get.enter", "request_id", requestID, "entity_id", tenantID)
	defer func() {
		slog.Info("support.store.tenant.get.exit", "request_id", requestID, "entity_id", tenantID, "duration_ms", time.Since(start).Milliseconds())
	}()

	s.mu.RLock()
	tenant, ok := s.tenants[tenantID]
	s.mu.RUnlock()
	if !ok {
		slog.Error("support.store.tenant.get.failed", "request_id", requestID, "entity_id", tenantID, "error_code", "SUPPORT_TENANT_NOT_FOUND")
		return SupportTenant{}, &SupportError{Code: "SUPPORT_TENANT_NOT_FOUND", Message: "tenant identity not found in Support DB"}
	}
	return tenant, nil
}

func (s *SupportStore) AddContact(requestID string, tenantID string, contact SupportContact) {
	s.mu.Lock()
	s.contactsByTenant[tenantID] = append(s.contactsByTenant[tenantID], contact)
	count := len(s.contactsByTenant[tenantID])
	s.mu.Unlock()
	slog.Info("support.store.contact.state_changed", "request_id", requestID, "entity_id", tenantID, "old", "count", "new", count, "event", "state changed")
}

func (s *SupportStore) ListContacts(tenantID string) []SupportContact {
	s.mu.RLock()
	defer s.mu.RUnlock()
	contacts := s.contactsByTenant[tenantID]
	clone := make([]SupportContact, len(contacts))
	copy(clone, contacts)
	return clone
}

func (s *SupportStore) AddIncident(requestID string, incident IncidentRecord) {
	s.mu.Lock()
	s.incidentsByTenant[incident.TenantID] = append(s.incidentsByTenant[incident.TenantID], incident)
	s.mu.Unlock()
	slog.Info("support.store.incident.state_changed", "request_id", requestID, "entity_id", incident.TenantID, "old", "none", "new", incident.IncidentID, "event", "state changed")
}

func (s *SupportStore) ListIncidents(tenantID string) []IncidentRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	incidents := s.incidentsByTenant[tenantID]
	clone := make([]IncidentRecord, len(incidents))
	copy(clone, incidents)
	return clone
}

func (s *SupportStore) AddComplianceEvidence(requestID string, evidence ComplianceEvidence) {
	s.mu.Lock()
	s.evidenceByControl[evidence.ControlClass] = append(s.evidenceByControl[evidence.ControlClass], evidence)
	s.evidenceByTenant[evidence.TenantID] = append(s.evidenceByTenant[evidence.TenantID], evidence)
	s.mu.Unlock()
	slog.Info("support.store.evidence.state_changed", "request_id", requestID, "entity_id", evidence.TenantID, "old", "indexing_pending", "new", evidence.ControlClass, "event", "state changed")
}

func (s *SupportStore) ListEvidenceByTenant(tenantID string) []ComplianceEvidence {
	s.mu.RLock()
	defer s.mu.RUnlock()
	evidence := s.evidenceByTenant[tenantID]
	clone := make([]ComplianceEvidence, len(evidence))
	copy(clone, evidence)
	return clone
}

func (s *SupportStore) ListEvidenceByControlClass(controlClass string) []ComplianceEvidence {
	s.mu.RLock()
	defer s.mu.RUnlock()
	evidence := s.evidenceByControl[controlClass]
	clone := make([]ComplianceEvidence, len(evidence))
	copy(clone, evidence)
	return clone
}

func (s *SupportStore) UpsertConfigOverride(requestID string, override ConfigOverride) {
	s.mu.Lock()
	overrides := s.overridesByTenant[override.TenantID]
	replaced := false
	for idx, item := range overrides {
		if item.Key == override.Key {
			overrides[idx] = override
			replaced = true
			break
		}
	}
	if !replaced {
		overrides = append(overrides, override)
	}
	s.overridesByTenant[override.TenantID] = overrides
	count := len(overrides)
	s.mu.Unlock()
	slog.Info("support.store.override.state_changed", "request_id", requestID, "entity_id", override.TenantID, "old", "count", "new", count, "event", "state changed")
}

func (s *SupportStore) ListConfigOverrides(tenantID string) []ConfigOverride {
	s.mu.RLock()
	defer s.mu.RUnlock()
	overrides := s.overridesByTenant[tenantID]
	clone := make([]ConfigOverride, len(overrides))
	copy(clone, overrides)
	return clone
}

// @hlv:sec [AUTH_BOUNDARY] — emergency access request enforces least-privilege support roles
func (s *SupportStore) RequestEmergencyAccess(requestID string, tenantID string, actorRole string, ticketID string) (EmergencyAccessRecord, *SupportError) {
	allowed := actorRole == "SecurityLead" || actorRole == "SRE"
	if !allowed {
		slog.Error("support.store.emergency_access.denied", "request_id", requestID, "entity_id", tenantID, "input_summary", fmt.Sprintf("actor_role=%s ticket_set=%t", actorRole, ticketID != ""), "error_code", "SUPPORT_EMERGENCY_ACCESS_DENIED")
		return EmergencyAccessRecord{}, &SupportError{Code: "SUPPORT_EMERGENCY_ACCESS_DENIED", Message: "emergency key request denied by role or approval policy"}
	}
	record := EmergencyAccessRecord{
		KeyID:                 fmt.Sprintf("key-%s-%d", tenantID, len(s.emergencyByTenant[tenantID])+1),
		RotationScheduleDays:  30,
		RequestedBy:           actorRole,
		TicketID:              ticketID,
		ApprovalPolicyMatched: true,
	}
	s.mu.Lock()
	s.emergencyByTenant[tenantID] = append(s.emergencyByTenant[tenantID], record)
	s.mu.Unlock()
	slog.Info("support.store.emergency_access.state_changed", "request_id", requestID, "entity_id", tenantID, "old", "none", "new", record.KeyID, "event", "state changed")
	return record, nil
}

func EvaluateSLA(input SLAInput) (SLAVerdict, *SupportError) {
	start := time.Now()
	slog.Info("support.sla.evaluate.enter", "request_id", input.RequestID, "entity_id", input.TenantID, "sla_tier", input.SLATier, "severity", input.IncidentSeverity)
	defer func() {
		slog.Info("support.sla.evaluate.exit", "request_id", input.RequestID, "entity_id", input.TenantID, "duration_ms", time.Since(start).Milliseconds())
	}()

	if input.SLATier != "community" && input.SLATier != "standard" && input.SLATier != "enterprise" && input.SLATier != "custom" {
		slog.Error("support.sla.evaluate.failed", "request_id", input.RequestID, "entity_id", input.TenantID, "input_summary", "unsupported tier", "error_code", "SLA_TIER_UNKNOWN")
		return SLAVerdict{}, &SupportError{Code: "SLA_TIER_UNKNOWN", Message: "sla_tier must be one of community|standard|enterprise|custom"}
	}
	if input.IncidentSeverity != "P0" && input.IncidentSeverity != "P1" && input.IncidentSeverity != "P2" && input.IncidentSeverity != "P3" {
		slog.Error("support.sla.evaluate.failed", "request_id", input.RequestID, "entity_id", input.TenantID, "input_summary", "unsupported severity", "error_code", "SLA_SEVERITY_UNKNOWN")
		return SLAVerdict{}, &SupportError{Code: "SLA_SEVERITY_UNKNOWN", Message: "incident_severity must be one of P0|P1|P2|P3"}
	}
	if !input.FirstResponseAt.IsZero() && input.FirstResponseAt.Before(input.OpenedAt) {
		slog.Error("support.sla.evaluate.failed", "request_id", input.RequestID, "entity_id", input.TenantID, "input_summary", "timeline invalid", "error_code", "SLA_TIMELINE_INVALID")
		return SLAVerdict{}, &SupportError{Code: "SLA_TIMELINE_INVALID", Message: "first_response_at cannot be earlier than opened_at"}
	}

	target := resolveSLATarget(input.SLATier, input.IncidentSeverity)
	if input.SLATier == "community" {
		return SLAVerdict{Status: "best_effort", Target: target, BreachReason: "non-contractual community tier"}, nil
	}

	if input.FirstResponseAt.IsZero() {
		return SLAVerdict{Status: "breached", Target: target, BreachReason: "first response missing"}, nil
	}

	responseMinutes := int(input.FirstResponseAt.Sub(input.OpenedAt).Minutes())
	if responseMinutes > target.ResponseTimeMinutes {
		return SLAVerdict{Status: "breached", Target: target, BreachReason: "response window exceeded"}, nil
	}
	return SLAVerdict{Status: "within_sla", Target: target}, nil
}

func resolveSLATarget(tier string, severity string) SLATarget {
	if tier == "custom" {
		if severity == "P0" {
			return SLATarget{ResponseTimeMinutes: 30, TTDMinutes: 10, TTRMinutes: 120, Support247: true}
		}
		if severity == "P1" {
			return SLATarget{ResponseTimeMinutes: 120, TTDMinutes: 30, TTRMinutes: 240, Support247: true}
		}
		return SLATarget{ResponseTimeMinutes: 240, TTDMinutes: 60, TTRMinutes: 480, Support247: false}
	}
	if tier == "enterprise" {
		if severity == "P0" {
			return SLATarget{ResponseTimeMinutes: 60, TTDMinutes: 20, TTRMinutes: 240, Support247: true}
		}
		return SLATarget{ResponseTimeMinutes: 180, TTDMinutes: 60, TTRMinutes: 480, Support247: true}
	}
	if tier == "standard" {
		return SLATarget{ResponseTimeMinutes: 240, TTDMinutes: 120, TTRMinutes: 1440, Support247: false}
	}
	return SLATarget{ResponseTimeMinutes: 0, TTDMinutes: 0, TTRMinutes: 0, Support247: false}
}
