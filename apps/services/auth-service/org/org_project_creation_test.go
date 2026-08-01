package org

// Validates: REQ-FUN.ORG.project-creation
//
// CreateProject behavior tests for project scope creation.
// Verifies: parent validation, visibility enforcement, role checks,
// ontology pairing, Owner membership, cleanup on failure.
// BDD: [Condition]_[Action]_[ExpectedResult]

import (
	"testing"
)

// TestCreateProject_WithValidParentGroup_CreatesScope checks that a project
// with a valid parent group and sufficient role (Maintainer+) succeeds.
func TestCreateProject_WithValidParentGroup_CreatesScope(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	// Create parent group with a Maintainer member
	_ = store.UpsertScope(ScopeNode{ID: "parent-group", Type: ScopeGroup, Visibility: VisibilityPrivate, Name: "Parent"})
	_ = store.UpsertMembership(OrgMembership{UserID: "maintainer-uuid", Scope: "parent-group", Role: "Maintainer"})

	child := ScopeNode{
		ID:         "new-project",
		Type:       ScopeProject,
		ParentID:   "parent-group",
		Name:       "Test Project",
		Visibility: VisibilityPrivate,
	}
	if err := svc.CreateScope("maintainer-uuid", child); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	stored, _ := store.GetScope("new-project")
	if stored == nil {
		t.Fatal("project scope not found")
	}
	if stored.Type != ScopeProject {
		t.Fatalf("expected ScopeProject, got %s", stored.Type)
	}
	if stored.Name != "Test Project" {
		t.Fatalf("expected name 'Test Project', got %s", stored.Name)
	}
	if stored.ParentID != "parent-group" {
		t.Fatalf("expected parent 'parent-group', got %s", stored.ParentID)
	}
}

// TestCreateProject_WithMissingParentGroup_ReturnsScopeNotFound checks that
// creating a project under a non-existent group fails.
func TestCreateProject_WithMissingParentGroup_ReturnsScopeNotFound(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	child := ScopeNode{
		ID:         "orphan-project",
		Type:       ScopeProject,
		ParentID:   "nonexistent-group",
		Name:       "Orphan",
		Visibility: VisibilityPrivate,
	}
	err := svc.CreateScope("owner-uuid", child)
	assertErrCode(t, err, "SCOPE_NOT_FOUND")
}

// TestCreateProject_WithVisibilityExceedingParent_ReturnsVisibilityViolation
// checks that a project cannot be more public than its parent group.
func TestCreateProject_WithVisibilityExceedingParent_ReturnsVisibilityViolation(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	_ = store.UpsertScope(ScopeNode{ID: "internal-group", Type: ScopeGroup, Visibility: VisibilityInternal, Name: "Internal Group"})
	_ = store.UpsertMembership(OrgMembership{UserID: "owner-uuid", Scope: "internal-group", Role: "Owner"})

	child := ScopeNode{
		ID:         "public-project",
		Type:       ScopeProject,
		ParentID:   "internal-group",
		Name:       "Too Public",
		Visibility: VisibilityPublic,
	}
	err := svc.CreateScope("owner-uuid", child)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if oe, ok := err.(*OrgError); ok {
		if oe.Code != "VISIBILITY_VIOLATION" {
			t.Fatalf("expected VISIBILITY_VIOLATION, got %s", oe.Code)
		}
	} else {
		t.Fatalf("expected *OrgError, got %T", err)
	}
}

// TestCreateProject_WithInsufficientRole_ReturnsForbidden checks that a user
// without at least Maintainer role in the parent group cannot create a project.
func TestCreateProject_WithInsufficientRole_ReturnsForbidden(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	_ = store.UpsertScope(ScopeNode{ID: "parent-group", Type: ScopeGroup, Visibility: VisibilityPrivate, Name: "Parent"})
	_ = store.UpsertMembership(OrgMembership{UserID: "guest-uuid", Scope: "parent-group", Role: "Guest"})

	child := ScopeNode{
		ID:         "restricted-project",
		Type:       ScopeProject,
		ParentID:   "parent-group",
		Name:       "Restricted",
		Visibility: VisibilityPrivate,
	}
	err := svc.CreateScope("guest-uuid", child)
	assertErrCode(t, err, "FORBIDDEN_INSUFFICIENT_ROLE")
}

// TestCreateProject_WithNoMembership_ReturnsForbidden checks that a user with
// no membership in the parent group cannot create a project.
func TestCreateProject_WithNoMembership_ReturnsForbidden(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	_ = store.UpsertScope(ScopeNode{ID: "parent-group", Type: ScopeGroup, Visibility: VisibilityPrivate, Name: "Parent"})

	child := ScopeNode{
		ID:         "no-access-project",
		Type:       ScopeProject,
		ParentID:   "parent-group",
		Name:       "No Access",
		Visibility: VisibilityPrivate,
	}
	err := svc.CreateScope("stranger-uuid", child)
	assertErrCode(t, err, "FORBIDDEN_INSUFFICIENT_ROLE")
}

// TestCreateProject_DeveloperInParentGroup_ReturnsForbidden checks that a
// Developer (below Maintainer) cannot create a project.
func TestCreateProject_DeveloperInParentGroup_ReturnsForbidden(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	_ = store.UpsertScope(ScopeNode{ID: "parent-group", Type: ScopeGroup, Visibility: VisibilityPrivate, Name: "Parent"})
	_ = store.UpsertMembership(OrgMembership{UserID: "dev-uuid", Scope: "parent-group", Role: "Developer"})

	child := ScopeNode{
		ID:         "dev-project",
		Type:       ScopeProject,
		ParentID:   "parent-group",
		Name:       "Dev Attempt",
		Visibility: VisibilityPrivate,
	}
	err := svc.CreateScope("dev-uuid", child)
	assertErrCode(t, err, "FORBIDDEN_INSUFFICIENT_ROLE")
}

// TestCreateProject_InheritsVisibilityFromParentWhenEmpty checks that a project
// without explicit visibility inherits from its parent group.
func TestCreateProject_InheritsVisibilityFromParentWhenEmpty(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	_ = store.UpsertScope(ScopeNode{ID: "internal-group", Type: ScopeGroup, Visibility: VisibilityInternal, Name: "Internal Group"})
	_ = store.UpsertMembership(OrgMembership{UserID: "owner-uuid", Scope: "internal-group", Role: "Owner"})

	child := ScopeNode{
		ID:       "inherited-project",
		Type:     ScopeProject,
		ParentID: "internal-group",
		Name:     "Inherited Vis",
		// No visibility set — should inherit from parent
	}
	if err := svc.CreateScope("owner-uuid", child); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	stored, _ := store.GetScope("inherited-project")
	if stored == nil {
		t.Fatal("project scope not found")
	}
	if stored.Visibility != VisibilityInternal {
		t.Fatalf("expected inherited visibility Internal, got %s", stored.Visibility)
	}
}

// TestCreateScope_ProjectWithMaintainerRole_Succeeds checks that Maintainer
// role is sufficient to create a project (not just Owner).
func TestCreateScope_ProjectWithMaintainerRole_Succeeds(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	_ = store.UpsertScope(ScopeNode{ID: "parent-group", Type: ScopeGroup, Visibility: VisibilityPrivate, Name: "Parent"})
	_ = store.UpsertMembership(OrgMembership{UserID: "maintainer-uuid", Scope: "parent-group", Role: "Maintainer"})

	child := ScopeNode{
		ID:         "maintainer-project",
		Type:       ScopeProject,
		ParentID:   "parent-group",
		Name:       "Maintainer Project",
		Visibility: VisibilityPrivate,
	}
	if err := svc.CreateScope("maintainer-uuid", child); err != nil {
		t.Fatalf("expected Maintainer to be allowed, got %v", err)
	}
}

// TestCreateProject_DefaultVisibilityPrivate checks that a top-level project
// (no parent) defaults to Private visibility.
func TestCreateProject_DefaultVisibilityPrivate(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	child := ScopeNode{
		ID:   "standalone-project",
		Type: ScopeProject,
		Name: "Standalone",
	}
	if err := svc.CreateScope("owner-uuid", child); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	stored, _ := store.GetScope("standalone-project")
	if stored == nil {
		t.Fatal("project scope not found")
	}
	if stored.Visibility != VisibilityPrivate {
		t.Fatalf("expected default Private, got %s", stored.Visibility)
	}
}

// TestCreateGroup_AutoAddsCreatorAsOwner checks that when a user creates a
// group via CreateScope, they automatically become an Owner member.
func TestCreateGroup_AutoAddsCreatorAsOwner(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	group := ScopeNode{
		ID:         "new-group",
		Type:       ScopeGroup,
		Name:       "Creator Group",
		Visibility: VisibilityPrivate,
	}
	if err := svc.CreateScope("creator-uuid", group); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	memberships, _ := store.GetMemberships("new-group")
	role := ResolveMaxRole(memberships)
	if role != "Owner" {
		t.Fatalf("expected creator to be Owner, got %q", role)
	}
}

// TestCreateGroup_CreatorCanCreateProjectInOwnGroup is the regression test for
// the GUI E2E flow: a group creator must be able to create a project inside
// their own group without an explicit membership assignment.
func TestCreateGroup_CreatorCanCreateProjectInOwnGroup(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)

	group := ScopeNode{
		ID:         "own-group",
		Type:       ScopeGroup,
		Name:       "Own Group",
		Visibility: VisibilityPrivate,
	}
	if err := svc.CreateScope("creator-uuid", group); err != nil {
		t.Fatalf("expected group creation to succeed, got %v", err)
	}

	project := ScopeNode{
		ID:         "own-project",
		Type:       ScopeProject,
		ParentID:   "own-group",
		Name:       "Own Project",
		Visibility: VisibilityPrivate,
	}
	if err := svc.CreateScope("creator-uuid", project); err != nil {
		t.Fatalf("expected creator to create project in own group, got %v", err)
	}
}
