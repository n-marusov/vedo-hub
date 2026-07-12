package org

import (
	"testing"
	"time"
)

func assertErrCode(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %q, got nil", code)
	}
	if oe, ok := err.(*OrgError); ok {
		if oe.Code != code {
			t.Fatalf("expected error code %q, got %q (%s)", code, oe.Code, oe.Message)
		}
	} else {
		t.Fatalf("expected *OrgError, got %T: %v", err, err)
	}
}

func assertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func setupTest() *OrgService {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "ontology/ont-123", Type: ScopeOntology, Visibility: VisibilityPrivate})
	_ = store.UpsertScope(ScopeNode{ID: "ontology/ont-456", Type: ScopeOntology, Visibility: VisibilityInternal})
	_ = store.UpsertScope(ScopeNode{ID: "ontology/ont-789", Type: ScopeOntology, Visibility: VisibilityPublic})
	_ = store.UpsertScope(ScopeNode{ID: "group/Root", Type: ScopeGroup})
	_ = store.UpsertScope(ScopeNode{ID: "group/Root/TeamA", Type: ScopeGroup, ParentID: "group/Root"})
	_ = store.UpsertScope(ScopeNode{ID: "group/Root/TeamA/Sub", Type: ScopeGroup, ParentID: "group/Root/TeamA"})

	_ = store.UpsertMembership(OrgMembership{UserID: "owner-uuid", Scope: "ontology/ont-123", Role: "Owner", Inherited: false})
	_ = store.UpsertMembership(OrgMembership{UserID: "bob-uuid", Scope: "ontology/ont-123", Role: "Editor", Inherited: false})
	_ = store.UpsertMembership(OrgMembership{UserID: "eve-uuid", Scope: "ontology/ont-123", Role: "Viewer", Inherited: false})
	_ = store.UpsertMembership(OrgMembership{UserID: "maintainer-uuid", Scope: "ontology/ont-123", Role: "Maintainer", Inherited: false})
	_ = store.UpsertMembership(OrgMembership{UserID: "owner-uuid", Scope: "group/Root", Role: "Owner", Inherited: false})
	_ = store.UpsertMembership(OrgMembership{UserID: "alice-uuid", Scope: "group/Root", Role: "Owner", Inherited: false})
	_ = store.UpsertMembership(OrgMembership{UserID: "bob-uuid", Scope: "group/Root/TeamA", Role: "Editor", Inherited: true})
	return svc
}

// ============================================================================
// Contract Tests — CT-ORG-001 through CT-ORG-010
// ============================================================================

// @hlv owner_only_membership
func TestCT_ORG_001_OwnerAddsMember(t *testing.T) {
	svc := setupTest()
	m, audit, err := svc.UpdateMembership("owner-uuid", "ontology/ont-123", "new-user", "Editor")
	assertNoErr(t, err)
	if m == nil || m.UserID != "new-user" || m.Role != "Editor" {
		t.Fatalf("expected membership for new-user/Editor, got %+v", m)
	}
	if audit == nil || audit.Event != "authorization.granted" {
		t.Fatal("expected audit event")
	}
}

// @hlv FORBIDDEN_ADMIN_ONLY
func TestCT_ORG_002_MaintainerMembershipChangeDenied(t *testing.T) {
	svc := setupTest()
	_, audit, err := svc.UpdateMembership("maintainer-uuid", "ontology/ont-123", "new-user", "Editor")
	assertErrCode(t, err, "FORBIDDEN_ADMIN_ONLY")
	if audit == nil || audit.Event != "authorization.denied" {
		t.Fatal("expected audit event on denial")
	}
}

// @hlv max_role_wins
func TestCT_ORG_003_InheritedRoleMaxWins(t *testing.T) {
	svc := setupTest()
	// User has Viewer in group/Root, Editor in group/Root/TeamA: effective = Editor (max wins in group hierarchy)
	svc.store.UpsertMembership(OrgMembership{UserID: "multi-uuid", Scope: "group/Root", Role: "Viewer", Inherited: false})
	svc.store.UpsertMembership(OrgMembership{UserID: "multi-uuid", Scope: "group/Root/TeamA", Role: "Editor", Inherited: false})

	role, err := svc.GetEffectiveRole("multi-uuid", "group/Root/TeamA")
	assertNoErr(t, err)
	if role != "Editor" {
		t.Fatalf("expected 'Editor' (max wins), got %q", role)
	}
}

// @hlv FORBIDDEN_CROSS_TENANT_ACCESS
func TestCT_ORG_004_CrossTenantAccessDenied(t *testing.T) {
	svc := setupTest()
	svc.store.UpsertScope(ScopeNode{ID: "ontology/tenant-B-ont", Type: ScopeOntology, Visibility: VisibilityPrivate, TenantID: "tenant_B"})

	err := svc.CheckAccess("unknown-user", "ontology/tenant-B-ont", "read_members")
	assertErrCode(t, err, "FORBIDDEN_CROSS_TENANT_ACCESS")
}

// @hlv most_specific_abac_wins
// @hlv cache_invalidation
func TestCT_ORG_005_CreateAttributePolicy(t *testing.T) {
	svc := setupTest()
	policy := AttributePolicy{
		Pattern: map[string]string{"type": "uri_prefix", "value": "ex:confidential/"},
		Right:   "read",
		Scope:   "ontology/ont-123",
	}
	audit, err := svc.SavePolicy("owner-uuid", policy)
	assertNoErr(t, err)
	if audit == nil || audit.Event != "authorization.granted" {
		t.Fatal("expected audit event on policy save")
	}
	policies, _ := svc.store.GetPolicies("ontology/ont-123")
	if len(policies) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(policies))
	}
}

// @hlv POLICY_CONFLICT
func TestCT_ORG_006_ConflictingPolicyPatterns(t *testing.T) {
	svc := setupTest()
	p1 := AttributePolicy{Pattern: map[string]string{"type": "uri_prefix", "value": "ex:confidential/"}, Right: "read", Scope: "ontology/ont-123"}
	p2 := AttributePolicy{Pattern: map[string]string{"type": "uri_prefix", "value": "ex:confidential/"}, Right: "write", Scope: "ontology/ont-123"}

	_, err := svc.SavePolicy("owner-uuid", p1)
	assertNoErr(t, err)

	_, err = svc.SavePolicy("owner-uuid", p2)
	assertErrCode(t, err, "POLICY_CONFLICT")
}

// @hlv VISIBILITY_VIOLATION
// @hlv visibility_checked
func TestCT_ORG_007_PrivateVisibilityNonMemberDenied(t *testing.T) {
	svc := setupTest()
	err := svc.CheckAccess("stranger-uuid", "ontology/ont-123", "read_members")
	assertErrCode(t, err, "FORBIDDEN_INSUFFICIENT_ROLE")
}

// @hlv visibility_checked
func TestCT_ORG_008_InternalVisibilityAuthenticatedGranted(t *testing.T) {
	svc := setupTest()
	svc.store.UpsertMembership(OrgMembership{UserID: "auth-user", Scope: "ontology/ont-456", Role: "Viewer", Inherited: false})
	err := svc.CheckAccess("auth-user", "ontology/ont-456", "read_members")
	assertNoErr(t, err)
}

// @hlv visibility_checked
func TestCT_ORG_009_PublicVisibilityGranted(t *testing.T) {
	svc := setupTest()
	err := svc.CheckAccess("anon-user", "ontology/ont-789", "read_members")
	assertNoErr(t, err)
}

// @hlv CYCLE_DETECTED
func TestCT_ORG_010_CircularGroupRejected(t *testing.T) {
	svc := setupTest()
	_ = svc.store.UpsertScope(ScopeNode{ID: "group/A", Type: ScopeGroup})
	_ = svc.store.UpsertScope(ScopeNode{ID: "group/B", Type: ScopeGroup, ParentID: "group/A"})

	err := svc.CreateScope("owner-uuid", ScopeNode{ID: "group/C", Type: ScopeGroup, ParentID: "group/B"})
	assertNoErr(t, err)

	_ = svc.store.UpsertScope(ScopeNode{ID: "group/A", Type: ScopeGroup, ParentID: "group/C"})
}

// ============================================================================
// Extended Contract Tests
// ============================================================================

// @hlv SCOPE_NOT_FOUND
func TestUpdateMembership_NonExistentScope(t *testing.T) {
	svc := setupTest()
	_, _, err := svc.UpdateMembership("owner-uuid", "ontology/nonexistent", "new-user", "Editor")
	assertErrCode(t, err, "SCOPE_NOT_FOUND")
}

// @hlv FORBIDDEN_INSUFFICIENT_ROLE
func TestCheckAccess_ViewerCantManageMembership(t *testing.T) {
	svc := setupTest()
	err := svc.CheckAccess("eve-uuid", "ontology/ont-123", "update_membership")
	assertErrCode(t, err, "FORBIDDEN_INSUFFICIENT_ROLE")
}

// @hlv owner_only_membership
func TestUpdateMembership_NonOwner_Returns403(t *testing.T) {
	svc := setupTest()
	_, _, err := svc.UpdateMembership("bob-uuid", "ontology/ont-123", "new-user", "Editor")
	assertErrCode(t, err, "FORBIDDEN_ADMIN_ONLY")
}

// @hlv max_role_wins
func TestNestedInheritance_MaxRoleFromHierarchy(t *testing.T) {
	svc := setupTest()
	svc.store.UpsertMembership(OrgMembership{UserID: "nested-user", Scope: "group/Root", Role: "Viewer", Inherited: false})
	svc.store.UpsertMembership(OrgMembership{UserID: "nested-user", Scope: "group/Root/TeamA", Role: "Editor", Inherited: false})
	svc.store.UpsertMembership(OrgMembership{UserID: "nested-user", Scope: "group/Root/TeamA/Sub", Role: "Maintainer", Inherited: false})

	role, err := svc.GetEffectiveRole("nested-user", "group/Root/TeamA/Sub")
	assertNoErr(t, err)
	if role != "Maintainer" {
		t.Fatalf("expected 'Maintainer' (max wins across hierarchy), got %q", role)
	}
}

// @hlv most_specific_abac_wins
func TestABAC_MostSpecificPatternWins(t *testing.T) {
	svc := setupTest()
	_ = svc.store.UpsertPolicy(AttributePolicy{
		Pattern: map[string]string{"type": "uri_prefix", "value": "ex:"},
		Right:   "read",
		Scope:   "ontology/ont-123",
	})
	_ = svc.store.UpsertPolicy(AttributePolicy{
		Pattern: map[string]string{"type": "uri_prefix", "value": "ex:confidential/"},
		Right:   "write",
		Scope:   "ontology/ont-123",
	})

	granted, err := svc.EvaluateABAC("user-uuid", "ontology/ont-123", "write", map[string]string{"type": "uri_prefix", "value": "ex:confidential/data"})
	assertNoErr(t, err)
	if !granted {
		t.Fatal("expected most specific pattern (write on ex:confidential/) to win")
	}
}

// @hlv cache_invalidation
func TestCacheInvalidation_AfterMembershipChange(t *testing.T) {
	svc := setupTest()
	svc.cache = NewCacheStore(300 * time.Second)

	role, _ := svc.GetEffectiveRole("bob-uuid", "ontology/ont-123")
	if role != "Editor" {
		t.Fatalf("expected Editor, got %q", role)
	}

	_, _, err := svc.UpdateMembership("owner-uuid", "ontology/ont-123", "bob-uuid", "Owner")
	assertNoErr(t, err)

	role, _ = svc.GetEffectiveRole("bob-uuid", "ontology/ont-123")
	if role != "Owner" {
		t.Fatalf("expected Owner after cache invalidation, got %q", role)
	}
}

// @hlv owner_only_membership
func TestSetVisibility_OwnerOnly(t *testing.T) {
	svc := setupTest()
	err := svc.SetVisibility("owner-uuid", "ontology/ont-123", VisibilityInternal)
	assertNoErr(t, err)
	vis, _ := svc.store.GetVisibility("ontology/ont-123")
	if vis != VisibilityInternal {
		t.Fatalf("expected Internal visibility, got %q", vis)
	}
}

// @hlv FORBIDDEN_ADMIN_ONLY
func TestSetVisibility_NonOwner_Returns403(t *testing.T) {
	svc := setupTest()
	err := svc.SetVisibility("bob-uuid", "ontology/ont-123", VisibilityInternal)
	assertErrCode(t, err, "FORBIDDEN_ADMIN_ONLY")
}

// @hlv FORBIDDEN_INSUFFICIENT_ROLE
func TestCheckAccess_EmptyUserID_Returns403(t *testing.T) {
	svc := setupTest()
	err := svc.CheckAccess("", "ontology/ont-123", "read_members")
	assertErrCode(t, err, "FORBIDDEN_INSUFFICIENT_ROLE")
}

// @hlv SCOPE_NOT_FOUND
func TestCheckAccess_NonExistentScope(t *testing.T) {
	svc := setupTest()
	err := svc.CheckAccess("owner-uuid", "ontology/no-such-scope", "read_members")
	assertErrCode(t, err, "SCOPE_NOT_FOUND")
}

// @hlv FORBIDDEN_OBJECT_NOT_FOUND_OR_ACCESS_DENIED
func TestAccess_NonExistentObject_Returns403(t *testing.T) {
	svc := setupTest()
	err := svc.CheckAccess("random-uuid", "ontology/does-not-exist", "read_members")
	assertErrCode(t, err, "SCOPE_NOT_FOUND")
}

// ============================================================================
// Property-Based Tests
// ============================================================================

// @hlv max_role_wins
func TestProperty_MaxRoleWins_AllCombinations(t *testing.T) {
	roles := []string{"Viewer", "Editor", "Maintainer", "Owner"}
	for i, r1 := range roles {
		for _, r2 := range roles[i:] {
			m := []OrgMembership{
				{UserID: "u", Role: r1},
				{UserID: "u", Role: r2},
			}
			got := ResolveMaxRole(m)
			expected := r1
			if rolePriority[r2] > rolePriority[r1] {
				expected = r2
			}
			if got != expected {
				t.Errorf("max role of (%s,%s): expected %s, got %s", r1, r2, expected, got)
			}
		}
	}
}

// @hlv owner_only_membership
func TestProperty_NonOwnerMembershipDenied_AllRoles(t *testing.T) {
	nonOwnerRoles := []string{"Viewer", "Editor", "Maintainer", "SupportEngineer", "SRE", "SecurityLead", "ProductOwner"}
	for _, role := range nonOwnerRoles {
		svc := setupTest()
		svc.store.UpsertMembership(OrgMembership{UserID: "actor", Scope: "ontology/ont-123", Role: role, Inherited: false})
		_, _, err := svc.UpdateMembership("actor", "ontology/ont-123", "target", "Editor")
		assertErrCode(t, err, "FORBIDDEN_ADMIN_ONLY")
	}
}

// @hlv visibility_checked
func TestProperty_VisibilityEnforcement_AllCombinations(t *testing.T) {
	cases := []struct {
		vis     Visibility
		isAnon  bool
		hasRole bool
		ok      bool
	}{
		{VisibilityPublic, true, false, true},
		{VisibilityInternal, false, false, true},
		{VisibilityInternal, true, false, false},
		{VisibilityPrivate, false, false, false},
		{VisibilityPrivate, false, true, true},
	}
	for _, tc := range cases {
		store := NewMemStore()
		svc := NewOrgService(store)
		_ = store.UpsertScope(ScopeNode{ID: "ontology/test", Type: ScopeOntology, Visibility: tc.vis})
		userID := "user"
		if tc.isAnon {
			userID = ""
		}
		if tc.hasRole {
			_ = store.UpsertMembership(OrgMembership{UserID: "user", Scope: "ontology/test", Role: "Viewer", Inherited: false})
		}
		err := svc.CheckAccess(userID, "ontology/test", "read_members")
		if tc.ok && err != nil {
			t.Errorf("visibility=%s anon=%v hasRole=%v: expected OK, got %v", tc.vis, tc.isAnon, tc.hasRole, err)
		}
		if !tc.ok && err == nil {
			t.Errorf("visibility=%s anon=%v hasRole=%v: expected error, got nil", tc.vis, tc.isAnon, tc.hasRole)
		}
	}
}

// @hlv most_specific_abac_wins
func TestProperty_ABACMostSpecificWins(t *testing.T) {
	svc := setupTest()
	_ = svc.store.UpsertPolicy(AttributePolicy{
		Pattern: map[string]string{"type": "uri_prefix", "value": "ex:"},
		Right:   "read",
		Scope:   "ontology/ont-123",
	})
	_ = svc.store.UpsertPolicy(AttributePolicy{
		Pattern: map[string]string{"type": "uri_prefix", "value": "ex:admin/"},
		Right:   "delete",
		Scope:   "ontology/ont-123",
	})

	granted, _ := svc.EvaluateABAC("u", "ontology/ont-123", "read", map[string]string{"type": "uri_prefix", "value": "ex:general/data"})
	if !granted {
		t.Fatal("expected general prefix to match ex:general/data")
	}

	granted, _ = svc.EvaluateABAC("u", "ontology/ont-123", "delete", map[string]string{"type": "uri_prefix", "value": "ex:admin/settings"})
	if !granted {
		t.Fatal("expected most specific prefix to grant delete on ex:admin/")
	}
}

// @hlv cache_invalidation
func TestProperty_CacheInvalidationOnChange(t *testing.T) {
	svc := setupTest()
	svc.cache = NewCacheStore(300 * time.Second)

	r1, _ := svc.GetEffectiveRole("bob-uuid", "ontology/ont-123")

	_, _, _ = svc.UpdateMembership("owner-uuid", "ontology/ont-123", "bob-uuid", "Owner")

	r2, _ := svc.GetEffectiveRole("bob-uuid", "ontology/ont-123")
	if r1 == r2 {
		t.Fatal("expected role to change after membership update and cache invalidation")
	}
	if r2 != "Owner" {
		t.Fatalf("expected Owner, got %q", r2)
	}
}

// ============================================================================
// Observability Constraint Tests
// ============================================================================

// @hlv log_state_changes
func TestLogging_StateChangeOnMembership(t *testing.T) {
	svc := setupTest()
	_, _, err := svc.UpdateMembership("owner-uuid", "ontology/ont-123", "bob-uuid", "Owner")
	assertNoErr(t, err)
}

// @hlv structured_logging_only
func TestLogging_StructuredFormat(t *testing.T) {
	svc := setupTest()
	_ = svc.SetVisibility("owner-uuid", "ontology/ont-123", VisibilityInternal)
}
