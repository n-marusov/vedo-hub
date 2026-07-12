package supportservice

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// @ctx: SUPPORT-META-001 WORM contract and property tests

func TestWORMAuditTrailIncludesChecksums(t *testing.T) {
	silenceLogs()
	store := NewWORMStore(7)

	event, err := store.WriteImmutableEvent("req-worm-1", "tenant-1", "emergency_access_granted", `{"ticket_id":"INC-1"}`)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if event.Checksum == "" {
		t.Fatal("expected checksum to be populated")
	}

	events, err := store.ListTenantEvents("req-worm-2", "tenant-1", time.Time{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(events) != 1 || events[0].Checksum == "" {
		t.Fatalf("expected one checksummed event, got %+v", events)
	}
	if !events[0].RetentionUntil.After(events[0].CreatedAt) {
		t.Fatalf("expected retention_after_created_at, got created=%s retention=%s", events[0].CreatedAt, events[0].RetentionUntil)
	}
}

// @hlv SUPPORT_AUDIT_WORM_UNAVAILABLE
func TestWORMOutageReturnsAvailabilityError(t *testing.T) {
	silenceLogs()
	store := NewWORMStore(7)
	store.SetAvailability("req-worm-3", false)

	_, err := store.ListTenantEvents("req-worm-4", "tenant-1", time.Time{})
	if err == nil || err.Code != "SUPPORT_AUDIT_WORM_UNAVAILABLE" {
		t.Fatalf("expected SUPPORT_AUDIT_WORM_UNAVAILABLE, got %+v", err)
	}
}

// @hlv SUPPORT_DATA_INTEGRITY_CHECKSUM_MISMATCH
func TestWORMChecksumMismatchReturnsIntegrityError(t *testing.T) {
	silenceLogs()
	store := NewWORMStore(7)
	event, err := store.WriteImmutableEvent("req-worm-5", "tenant-1", "backup_catalog_updated", `{"backup_id":"b-1"}`)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	verifyErr := store.VerifyChecksum("req-worm-6", event.EventID, `{"backup_id":"tampered"}`)
	if verifyErr == nil || verifyErr.Code != "SUPPORT_DATA_INTEGRITY_CHECKSUM_MISMATCH" {
		t.Fatalf("expected SUPPORT_DATA_INTEGRITY_CHECKSUM_MISMATCH, got %+v", verifyErr)
	}
}

// @hlv PBT-SUPPORT-META-001
func TestSupportMetaPropertyChecksumStableAndTamperDetected(t *testing.T) {
	silenceLogs()
	store := NewWORMStore(7)
	rng := rand.New(rand.NewSource(7))

	for i := 0; i < 10000; i++ {
		payload := fmt.Sprintf(`{"tenant":"t-%d","nonce":%d}`, i, rng.Int63())
		event, err := store.WriteImmutableEvent("req-worm-pbt-1", "tenant-pbt", "incident_update", payload)
		if err != nil {
			t.Fatalf("unexpected write error: %v", err)
		}

		if verifyErr := store.VerifyChecksum("req-worm-pbt-2", event.EventID, payload); verifyErr != nil {
			t.Fatalf("expected checksum to verify for original payload: %+v", verifyErr)
		}

		tampered := payload + "-tampered"
		verifyErr := store.VerifyChecksum("req-worm-pbt-3", event.EventID, tampered)
		if verifyErr == nil || verifyErr.Code != "SUPPORT_DATA_INTEGRITY_CHECKSUM_MISMATCH" {
			t.Fatalf("expected mismatch for tampered payload, got %+v", verifyErr)
		}
	}
}

// @hlv PBT-SUPPORT-META-002
func TestSupportMetaPropertySupportPlaneAvailableDuringTenantOutage(t *testing.T) {
	silenceLogs()
	store := NewSupportStore()
	rng := rand.New(rand.NewSource(17))

	for i := 0; i < 10000; i++ {
		tenantID := fmt.Sprintf("tenant-%d", i)
		store.UpsertTenant("req-meta-pbt-1", SupportTenant{TenantID: tenantID, TenantStatus: "active", DeploymentModel: "saas", SLATier: "standard", EscalationPolicy: "default"})

		tenantDBOutage := rng.Intn(2) == 0
		if tenantDBOutage {
			store.AddIncident("req-meta-pbt-2", IncidentRecord{IncidentID: fmt.Sprintf("inc-%d", i), TenantID: tenantID, Severity: "P1", OpenedAt: time.Now().UTC()})
		}

		_, err := store.GetTenantInfo("req-meta-pbt-3", tenantID)
		if err != nil {
			t.Fatalf("support plane should remain available during outage=%t: %v", tenantDBOutage, err)
		}
	}
}

// @hlv PBT-SUPPORT-META-003
func TestSupportMetaPropertyMetadataClassesRouteToStorageAndRetention(t *testing.T) {
	silenceLogs()
	supportStore := NewSupportStore()
	wormStore := NewWORMStore(7)
	rng := rand.New(rand.NewSource(71))
	categories := []string{"tenant", "contact", "incident", "compliance", "override", "emergency", "audit"}

	for i := 0; i < 10000; i++ {
		tenantID := fmt.Sprintf("tenant-route-%d", i)
		category := categories[rng.Intn(len(categories))]

		supportStore.UpsertTenant("req-meta-pbt-4", SupportTenant{TenantID: tenantID, TenantStatus: "active", DeploymentModel: "saas", SLATier: "enterprise", EscalationPolicy: "default"})
		exerciseCategory(t, supportStore, wormStore, tenantID, category, i)
	}
}

func exerciseCategory(t *testing.T, supportStore *SupportStore, wormStore *WORMStore, tenantID, category string, i int) {
	t.Helper()
	switch category {
	case "tenant":
		if _, err := supportStore.GetTenantInfo("req-meta-pbt-5", tenantID); err != nil {
			t.Fatalf("tenant metadata missing: %v", err)
		}
	case "contact":
		supportStore.AddContact("req-meta-pbt-6", tenantID, SupportContact{Name: "N", Email: "n@example.com", Role: "SupportEngineer"})
		if len(supportStore.ListContacts(tenantID)) == 0 {
			t.Fatal("contact metadata not routed to support store")
		}
	case "incident":
		supportStore.AddIncident("req-meta-pbt-7", IncidentRecord{IncidentID: fmt.Sprintf("inc-%d", i), TenantID: tenantID, Severity: "P2", OpenedAt: time.Now().UTC()})
		if len(supportStore.ListIncidents(tenantID)) == 0 {
			t.Fatal("incident metadata not routed to support store")
		}
	case "compliance":
		supportStore.AddComplianceEvidence("req-meta-pbt-8", ComplianceEvidence{EvidenceID: fmt.Sprintf("ev-%d", i), TenantID: tenantID, ControlClass: "SOC2", Reference: "doc://x"})
		if len(supportStore.ListEvidenceByControlClass("SOC2")) == 0 {
			t.Fatal("compliance evidence not indexed by control class")
		}
	case "override":
		supportStore.UpsertConfigOverride("req-meta-pbt-9", ConfigOverride{TenantID: tenantID, Key: "sla.override", Value: "custom", Reason: "case", UpdatedBy: "ops", UpdatedAt: "2026-06-10T12:00:00Z"})
		if len(supportStore.ListConfigOverrides(tenantID)) == 0 {
			t.Fatal("config override not stored in support metadata")
		}
	case "emergency":
		if _, err := supportStore.RequestEmergencyAccess("req-meta-pbt-10", tenantID, "SRE", "INC-42"); err != nil {
			t.Fatalf("expected emergency access metadata to be accepted for SRE: %v", err)
		}
	case "audit":
		event, err := wormStore.WriteImmutableEvent("req-meta-pbt-11", tenantID, "metadata_read", "{}")
		if err != nil {
			t.Fatalf("audit event write failed: %v", err)
		}
		if !event.RetentionUntil.After(event.CreatedAt.AddDate(6, 11, 0)) {
			t.Fatalf("expected ~7y retention, got created=%s retention=%s", event.CreatedAt, event.RetentionUntil)
		}
	}
}
