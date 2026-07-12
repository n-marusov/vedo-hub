package supportservice

import (
	"io"
	"log/slog"
	"math/rand"
	"sync"
	"testing"
	"time"
)

// @ctx: SUPPORT-META-001 and SUPPORT-SLA-001 contract tests

var silenceLogsOnce sync.Once

func silenceLogs() {
	silenceLogsOnce.Do(func() {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	})
}

func TestSupportStoreTenantInfoHappyPath(t *testing.T) {
	silenceLogs()
	store := NewSupportStore()
	store.UpsertTenant("req-meta-1", SupportTenant{TenantID: "tenant-1", TenantStatus: "active", DeploymentModel: "saas", SLATier: "enterprise", EscalationPolicy: "policy-a"})

	tenant, err := store.GetTenantInfo("req-meta-2", "tenant-1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if tenant.TenantID != "tenant-1" || tenant.SLATier != "enterprise" {
		t.Fatalf("unexpected tenant payload: %+v", tenant)
	}
}

// @hlv SUPPORT_TENANT_NOT_FOUND
func TestSupportStoreTenantNotFound(t *testing.T) {
	silenceLogs()
	store := NewSupportStore()

	_, err := store.GetTenantInfo("req-meta-3", "tenant-missing")
	if err == nil || err.Code != "SUPPORT_TENANT_NOT_FOUND" {
		t.Fatalf("expected SUPPORT_TENANT_NOT_FOUND, got %+v", err)
	}
}

// @hlv SUPPORT_EMERGENCY_ACCESS_DENIED
func TestSupportStoreEmergencyAccessDeniedForUnauthorizedRole(t *testing.T) {
	silenceLogs()
	store := NewSupportStore()

	_, err := store.RequestEmergencyAccess("req-meta-4", "tenant-1", "SupportEngineer", "INC-1")
	if err == nil || err.Code != "SUPPORT_EMERGENCY_ACCESS_DENIED" {
		t.Fatalf("expected SUPPORT_EMERGENCY_ACCESS_DENIED, got %+v", err)
	}
}

func TestSupportStoreComplianceEvidenceIndexedByTenantAndControlClass(t *testing.T) {
	silenceLogs()
	store := NewSupportStore()
	record := ComplianceEvidence{EvidenceID: "ev-1", TenantID: "tenant-1", ControlClass: "SOC2", Reference: "doc://audit-1"}
	store.AddComplianceEvidence("req-meta-5", record)

	byTenant := store.ListEvidenceByTenant("tenant-1")
	if len(byTenant) != 1 || byTenant[0].EvidenceID != "ev-1" {
		t.Fatalf("unexpected tenant index: %+v", byTenant)
	}

	byControl := store.ListEvidenceByControlClass("SOC2")
	if len(byControl) != 1 || byControl[0].EvidenceID != "ev-1" {
		t.Fatalf("unexpected control-class index: %+v", byControl)
	}
}

func TestSupportStoreConfigOverrideUpsertReplacesByKey(t *testing.T) {
	silenceLogs()
	store := NewSupportStore()
	store.UpsertConfigOverride("req-meta-6", ConfigOverride{TenantID: "tenant-1", Key: "sla.override", Value: "standard", Reason: "temporary", UpdatedBy: "ops", UpdatedAt: "2026-06-10T12:00:00Z"})
	store.UpsertConfigOverride("req-meta-7", ConfigOverride{TenantID: "tenant-1", Key: "sla.override", Value: "enterprise", Reason: "renewal", UpdatedBy: "ops", UpdatedAt: "2026-06-10T12:30:00Z"})

	overrides := store.ListConfigOverrides("tenant-1")
	if len(overrides) != 1 {
		t.Fatalf("expected one override, got %d", len(overrides))
	}
	if overrides[0].Value != "enterprise" {
		t.Fatalf("expected replaced override value, got %+v", overrides[0])
	}
}

func TestEvaluateSLAEnterpriseCustomP0WithinSLA(t *testing.T) {
	silenceLogs()
	opened := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	response := opened.Add(20 * time.Minute)

	verdict, err := EvaluateSLA(SLAInput{TenantID: "tenant-enterprise-1", SLATier: "custom", IncidentSeverity: "P0", OpenedAt: opened, FirstResponseAt: response, RequestID: "req-sla-1"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if verdict.Status != "within_sla" {
		t.Fatalf("expected within_sla, got %+v", verdict)
	}
	if verdict.Target.ResponseTimeMinutes != 30 || verdict.Target.TTRMinutes != 120 || !verdict.Target.Support247 {
		t.Fatalf("unexpected target for custom P0: %+v", verdict.Target)
	}
}

// @hlv SLA_TIER_UNKNOWN
func TestEvaluateSLARejectsUnknownTier(t *testing.T) {
	silenceLogs()
	_, err := EvaluateSLA(SLAInput{TenantID: "tenant-2", SLATier: "gold", IncidentSeverity: "P1", OpenedAt: time.Now().UTC(), RequestID: "req-sla-2"})
	if err == nil || err.Code != "SLA_TIER_UNKNOWN" {
		t.Fatalf("expected SLA_TIER_UNKNOWN, got %+v", err)
	}
}

// @hlv SLA_SEVERITY_UNKNOWN
func TestEvaluateSLARejectsUnknownSeverity(t *testing.T) {
	silenceLogs()
	_, err := EvaluateSLA(SLAInput{TenantID: "tenant-2", SLATier: "standard", IncidentSeverity: "PX", OpenedAt: time.Now().UTC(), RequestID: "req-sla-3"})
	if err == nil || err.Code != "SLA_SEVERITY_UNKNOWN" {
		t.Fatalf("expected SLA_SEVERITY_UNKNOWN, got %+v", err)
	}
}

// @hlv SLA_TIMELINE_INVALID
func TestEvaluateSLARejectsInvalidTimeline(t *testing.T) {
	silenceLogs()
	opened := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	response := opened.Add(-1 * time.Minute)

	_, err := EvaluateSLA(SLAInput{TenantID: "tenant-2", SLATier: "enterprise", IncidentSeverity: "P1", OpenedAt: opened, FirstResponseAt: response, RequestID: "req-sla-4"})
	if err == nil || err.Code != "SLA_TIMELINE_INVALID" {
		t.Fatalf("expected SLA_TIMELINE_INVALID, got %+v", err)
	}
}

func TestEvaluateSLACommunityTierBestEffort(t *testing.T) {
	silenceLogs()
	opened := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)

	verdict, err := EvaluateSLA(SLAInput{TenantID: "tenant-community-1", SLATier: "community", IncidentSeverity: "P2", OpenedAt: opened, RequestID: "req-sla-5"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if verdict.Status != "best_effort" {
		t.Fatalf("expected best_effort, got %+v", verdict)
	}
}

// @hlv PBT-SUPPORT-SLA-001
func TestSLAPropertyTierSeverityMapping(t *testing.T) {
	silenceLogs()
	tiers := []string{"community", "standard", "enterprise", "custom"}
	severities := []string{"P0", "P1", "P2", "P3"}
	rng := rand.New(rand.NewSource(42))
	opened := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)

	for i := 0; i < 10000; i++ {
		tier := tiers[rng.Intn(len(tiers))]
		severity := severities[rng.Intn(len(severities))]
		expected := resolveSLATarget(tier, severity)

		verdict, err := EvaluateSLA(SLAInput{TenantID: "tenant-pbt", SLATier: tier, IncidentSeverity: severity, OpenedAt: opened, FirstResponseAt: opened.Add(1 * time.Minute), RequestID: "req-sla-pbt-1"})
		if err != nil {
			t.Fatalf("unexpected error for tier=%s severity=%s: %v", tier, severity, err)
		}
		if verdict.Target != expected {
			t.Fatalf("unexpected target for tier=%s severity=%s: got %+v expected %+v", tier, severity, verdict.Target, expected)
		}
	}
}

// @hlv PBT-SUPPORT-SLA-002
func TestSLAPropertyResponseCannotPrecedeOpenTimestamp(t *testing.T) {
	silenceLogs()
	rng := rand.New(rand.NewSource(99))
	base := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)

	for i := 0; i < 10000; i++ {
		openOffset := time.Duration(rng.Intn(720)) * time.Minute
		opened := base.Add(openOffset)
		responseOffset := time.Duration(rng.Intn(721)-360) * time.Minute
		response := opened.Add(responseOffset)

		_, err := EvaluateSLA(SLAInput{TenantID: "tenant-pbt", SLATier: "enterprise", IncidentSeverity: "P1", OpenedAt: opened, FirstResponseAt: response, RequestID: "req-sla-pbt-2"})
		if response.Before(opened) {
			if err == nil || err.Code != "SLA_TIMELINE_INVALID" {
				t.Fatalf("expected SLA_TIMELINE_INVALID for response=%s opened=%s, got %+v", response, opened, err)
			}
			continue
		}
		if err != nil && err.Code == "SLA_TIMELINE_INVALID" {
			t.Fatalf("unexpected SLA_TIMELINE_INVALID for response=%s opened=%s", response, opened)
		}
	}
}

// @hlv PBT-SUPPORT-SLA-003
func TestSLAPropertyCustomP0P1Always24x7(t *testing.T) {
	silenceLogs()
	rng := rand.New(rand.NewSource(123))
	severities := []string{"P0", "P1"}
	opened := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)

	for i := 0; i < 10000; i++ {
		severity := severities[rng.Intn(len(severities))]
		verdict, err := EvaluateSLA(SLAInput{TenantID: "tenant-pbt", SLATier: "custom", IncidentSeverity: severity, OpenedAt: opened, FirstResponseAt: opened.Add(1 * time.Minute), RequestID: "req-sla-pbt-3"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !verdict.Target.Support247 {
			t.Fatalf("expected custom %s to be 24x7, got %+v", severity, verdict.Target)
		}
	}
}
