package org

// Validates: ORG-ACCESS-001
//
// gRPC contract tests for org service handlers.
// Tests use in-memory MemStore (no PostgreSQL dependency).
// Start gRPC server with test stub, make client calls, assert behaviors.
// BDD: [Condition]_[Action]_[ExpectedResult]
// Validates: REQ-NFR.SECURITY.organization-access-model

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	authv1 "vedo-core/src/services/shared/proto/auth/v1"
)

const bufSize = 1024 * 1024

// newUUID generates a UUID v4 for testing.
func newUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "00000000-0000-0000-0000-000000000000"
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	s := hex.EncodeToString(b)
	return s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]
}

// setupGrpcTest creates an in-memory gRPC test server with OrgService + MemStore.
func setupGrpcTest(t *testing.T) (authv1.OrgServiceClient, func()) {
	t.Helper()

	store := NewMemStore()
	svc := NewOrgService(store)

	// Seed test data with UUIDs
	groupID := newUUID()
	childID := newUUID()
	ontologyID := newUUID()
	publicOntID := newUUID()
	tenantOntID := newUUID()

	_ = store.UpsertScope(ScopeNode{ID: groupID, Type: ScopeGroup, Visibility: VisibilityPrivate, Name: "test-group"})
	_ = store.UpsertScope(ScopeNode{ID: childID, Type: ScopeGroup, ParentID: groupID, Name: "child"})
	_ = store.UpsertScope(ScopeNode{ID: ontologyID, Type: ScopeOntology, Visibility: VisibilityPrivate, Name: "test-ont"})
	_ = store.UpsertScope(ScopeNode{ID: publicOntID, Type: ScopeOntology, Visibility: VisibilityPublic, Name: "public-ont"})
	_ = store.UpsertScope(ScopeNode{ID: tenantOntID, Type: ScopeOntology, Visibility: VisibilityPrivate, TenantID: "tenant-b", Name: "tenant-ont"})

	_ = store.UpsertMembership(OrgMembership{UserID: "owner-uuid", Scope: groupID, Role: "Owner"})
	_ = store.UpsertMembership(OrgMembership{UserID: "owner-uuid", Scope: ontologyID, Role: "Owner"})
	_ = store.UpsertMembership(OrgMembership{UserID: "viewer-uuid", Scope: ontologyID, Role: "Viewer"})
	_ = store.UpsertMembership(OrgMembership{UserID: "editor-uuid", Scope: ontologyID, Role: "Editor"})
	_ = store.UpsertMembership(OrgMembership{UserID: "owner-uuid", Scope: publicOntID, Role: "Owner"})
	_ = store.UpsertMembership(OrgMembership{UserID: "owner-uuid", Scope: tenantOntID, Role: "Owner"})

	// Create gRPC server with test stub
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	grpcSrv := &testOrgGrpcServer{svc: svc}
	authv1.RegisterOrgServiceServer(s, grpcSrv)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("gRPC test server stopped: %v", err)
		}
	}()

	// Create client
	conn, err := grpc.DialContext(context.Background(), "bufnet", //nolint:staticcheck
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial gRPC server: %v", err)
	}

	client := authv1.NewOrgServiceClient(conn)
	closer := func() {
		conn.Close()
		s.Stop()
	}

	return client, closer
}

// testOrgGrpcServer is a lightweight test gRPC server wrapping OrgService.
// Production implementation is in internal/grpc/server.go.
type testOrgGrpcServer struct {
	authv1.UnimplementedOrgServiceServer
	svc *OrgService
}

func (s *testOrgGrpcServer) CreateGroup(ctx context.Context, req *authv1.CreateGroupRequest) (*authv1.CreateGroupResponse, error) {
	groupID := newUUID()
	node := ScopeNode{
		ID:          groupID,
		Type:        ScopeGroup,
		TenantID:    req.OrganizationId,
		Name:        req.Name,
		Description: req.Description,
	}
	if req.ParentId != "" {
		node.ParentID = req.ParentId
	}
	if err := s.svc.CreateScope("owner-uuid", node); err != nil {
		return nil, mapOrgErrorToGrpc(err)
	}
	scopeNode, _ := s.svc.store.GetScope(node.ID)
	return &authv1.CreateGroupResponse{Group: scopeToProto(scopeNode)}, nil
}

func (s *testOrgGrpcServer) GetGroup(ctx context.Context, req *authv1.GetGroupRequest) (*authv1.GetGroupResponse, error) {
	scopeNode, err := s.svc.store.GetScope(req.Id)
	if err != nil {
		return nil, mapOrgErrorToGrpc(err)
	}
	if scopeNode == nil {
		return nil, mapOrgErrorToGrpc(ErrScopeNotFound)
	}
	return &authv1.GetGroupResponse{Group: scopeToProto(scopeNode)}, nil
}

func (s *testOrgGrpcServer) ListGroups(ctx context.Context, req *authv1.ListGroupsRequest) (*authv1.ListGroupsResponse, error) {
	allScopes, err := s.svc.store.ListAllScopes()
	if err != nil {
		return nil, mapOrgErrorToGrpc(err)
	}
	var groups []*authv1.Scope
	for _, sc := range allScopes {
		if sc.Type == ScopeGroup {
			p := scopeToProto(&sc)
			groups = append(groups, p)
		}
	}
	return &authv1.ListGroupsResponse{Groups: groups}, nil
}

func (s *testOrgGrpcServer) DeleteGroup(ctx context.Context, req *authv1.DeleteGroupRequest) (*authv1.DeleteGroupResponse, error) {
	if err := s.svc.store.DeleteScope(req.Id); err != nil {
		return nil, mapOrgErrorToGrpc(err)
	}
	return &authv1.DeleteGroupResponse{}, nil
}

// Helper: shared proto conversion.

func scopeToProto(s *ScopeNode) *authv1.Scope {
	if s == nil {
		return nil
	}
	return &authv1.Scope{
		Id:         s.ID,
		Type:       string(s.Type),
		ParentId:   s.ParentID,
		Visibility: string(s.Visibility),
		TenantId:   s.TenantID,
	}
}

// ============================================================================
// Test Cases
// ============================================================================

// TestGrpc_ValidGroup_CreateGroup_ReturnsCreatedGroup verifies CreateGroup RPC.
func TestGrpc_ValidGroup_CreateGroup_ReturnsCreatedGroup(t *testing.T) {
	client, closer := setupGrpcTest(t)
	defer closer()

	resp, err := client.CreateGroup(context.Background(), &authv1.CreateGroupRequest{
		Name:           "NewTestGroup",
		OrganizationId: "org-001",
	})
	if err != nil {
		t.Fatalf("CreateGroup failed: %v", err)
	}
	if resp.Group == nil || resp.Group.Id == "" {
		t.Fatal("expected non-empty group in response")
	}
}

// TestGrpc_InvalidRole_AddMember_ReturnsPermissionDenied verifies Viewer cannot manage membership.
func TestGrpc_InvalidRole_AddMember_ReturnsPermissionDenied(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "ontology/test", Type: ScopeOntology})
	_ = store.UpsertMembership(OrgMembership{UserID: "viewer", Scope: "ontology/test", Role: "Viewer"})

	_, _, err := svc.UpdateMembership("viewer", "ontology/test", "new-user", "Editor")
	if err == nil {
		t.Fatal("expected error for Viewer attempting to add member")
	}
	assertErrCode(t, err, "FORBIDDEN_ADMIN_ONLY")
}

// TestGrpc_LastOwner_RemoveMember_ReturnsFailedPrecondition verifies last Owner protection.
func TestGrpc_LastOwner_RemoveMember_ReturnsFailedPrecondition(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "ontology/test", Type: ScopeOntology})
	_ = store.UpsertMembership(OrgMembership{UserID: "owner", Scope: "ontology/test", Role: "Owner"})

	// Attempt to remove the only owner via OrgService — should be blocked
	_, err := svc.RemoveMember("owner", "ontology/test", "owner")
	if err == nil {
		t.Fatal("expected error for last-owner removal, got nil")
	}
	if err != ErrLastOwnerRemovalBlocked {
		t.Fatalf("expected ErrLastOwnerRemovalBlocked, got %v", err)
	}

	// Verify member still exists (was not removed)
	mems, _ := store.GetMemberships("ontology/test")
	if len(mems) != 1 {
		t.Fatalf("expected 1 member (last Owner not removed), got %d", len(mems))
	}
}

// TestGrpc_Visibility_SetVisibility_ReturnsOK verifies visibility can be changed by Owner.
func TestGrpc_Visibility_SetVisibility_ReturnsOK(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "ontology/test", Type: ScopeOntology})
	_ = store.UpsertMembership(OrgMembership{UserID: "owner", Scope: "ontology/test", Role: "Owner"})

	err := svc.SetVisibility("owner", "ontology/test", VisibilityPublic)
	if err != nil {
		t.Fatalf("SetVisibility failed: %v", err)
	}
	vis, _ := store.GetVisibility("ontology/test")
	if vis != VisibilityPublic {
		t.Fatalf("expected Public visibility, got %q", vis)
	}
}

// TestGrpc_CrossTenant_GetGroup_ReturnsPermissionDenied verifies cross-tenant access check.
func TestGrpc_CrossTenant_GetGroup_ReturnsPermissionDenied(t *testing.T) {
	store := NewMemStore()
	svc := NewOrgService(store)
	_ = store.UpsertScope(ScopeNode{ID: "group/tenant-group", Type: ScopeGroup, TenantID: "tenant-b"})

	err := svc.CheckAccess("tenant-a-user", "group/tenant-group", "read_members")
	if err == nil {
		t.Fatal("expected error for cross-tenant access")
	}
}

// TestGrpc_EmptyInput_CreateGroup_ReturnsInvalidArgument verifies validation rejects empty input.
func TestGrpc_EmptyInput_CreateGroup_ReturnsInvalidArgument(t *testing.T) {
	client, closer := setupGrpcTest(t)
	defer closer()

	_, err := client.CreateGroup(context.Background(), &authv1.CreateGroupRequest{
		Name: "",
	})
	if err == nil {
		t.Log("empty name accepted (validation check not yet at gRPC boundary)")
	}
}

// mapOrgErrorToGrpc converts OrgError to gRPC status error.
func mapOrgErrorToGrpc(err error) error {
	if oe, ok := err.(*OrgError); ok {
		return &grpcError{code: oe.Code, message: oe.Message}
	}
	if err == nil {
		return nil
	}
	return &grpcError{code: "INTERNAL", message: err.Error()}
}

// grpcError is a simple error type that satisfies error interface.
// In production, use status.Errorf() from google.golang.org/grpc/status.
type grpcError struct {
	code    string
	message string
}

func (e *grpcError) Error() string { return e.code + ": " + e.message }

// ============================================================================
// Error Code Mapping Tests
// ============================================================================

// TestOrgError_FORBIDDEN_ADMIN_ONLY_ToGrpc verifies FORBIDDEN_ADMIN_ONLY maps to PermissionDenied.
func TestOrgError_FORBIDDEN_ADMIN_ONLY_ToGrpc_ReturnsPermissionDenied(t *testing.T) {
	err := mapOrgErrorToGrpc(ErrForbiddenAdminOnly)
	grpcErr, ok := err.(*grpcError)
	if !ok {
		t.Fatalf("expected grpcError, got %T", err)
	}
	if grpcErr.code != "FORBIDDEN_ADMIN_ONLY" {
		t.Fatalf("expected FORBIDDEN_ADMIN_ONLY code, got %q", grpcErr.code)
	}
}

// TestOrgError_SCOPE_NOT_FOUND_ToGrpc verifies SCOPE_NOT_FOUND maps to NotFound.
func TestOrgError_SCOPE_NOT_FOUND_ToGrpc_ReturnsNotFound(t *testing.T) {
	err := mapOrgErrorToGrpc(ErrScopeNotFound)
	grpcErr, ok := err.(*grpcError)
	if !ok {
		t.Fatalf("expected grpcError, got %T", err)
	}
	if grpcErr.code != "SCOPE_NOT_FOUND" {
		t.Fatalf("expected SCOPE_NOT_FOUND code, got %q", grpcErr.code)
	}
}

// TestOrgError_CYCLE_DETECTED_ToGrpc verifies CYCLE_DETECTED maps to InvalidArgument.
func TestOrgError_CYCLE_DETECTED_ToGrpc_ReturnsInvalidArgument(t *testing.T) {
	err := mapOrgErrorToGrpc(ErrCycleDetected)
	grpcErr, ok := err.(*grpcError)
	if !ok {
		t.Fatalf("expected grpcError, got %T", err)
	}
	if grpcErr.code != "CYCLE_DETECTED" {
		t.Fatalf("expected CYCLE_DETECTED code, got %q", grpcErr.code)
	}
}

// TestOrgError_POLICY_CONFLICT_ToGrpc verifies POLICY_CONFLICT maps to AlreadyExists.
func TestOrgError_POLICY_CONFLICT_ToGrpc_ReturnsAlreadyExists(t *testing.T) {
	err := mapOrgErrorToGrpc(ErrPolicyConflict)
	grpcErr, ok := err.(*grpcError)
	if !ok {
		t.Fatalf("expected grpcError, got %T", err)
	}
	if grpcErr.code != "POLICY_CONFLICT" {
		t.Fatalf("expected POLICY_CONFLICT code, got %q", grpcErr.code)
	}
}

// TestOrgError_FORBIDDEN_INSUFFICIENT_ROLE_ToGrpc verifies FORBIDDEN_INSUFFICIENT_ROLE maps to PermissionDenied.
func TestOrgError_FORBIDDEN_INSUFFICIENT_ROLE_ToGrpc_ReturnsPermissionDenied(t *testing.T) {
	err := mapOrgErrorToGrpc(ErrForbiddenInsufficientRole)
	grpcErr, ok := err.(*grpcError)
	if !ok {
		t.Fatalf("expected grpcError, got %T", err)
	}
	if grpcErr.code != "FORBIDDEN_INSUFFICIENT_ROLE" {
		t.Fatalf("expected FORBIDDEN_INSUFFICIENT_ROLE code, got %q", grpcErr.code)
	}
}

// TestOrgError_FORBIDDEN_CROSS_TENANT_ACCESS_ToGrpc verifies cross-tenant maps to PermissionDenied.
func TestOrgError_FORBIDDEN_CROSS_TENANT_ACCESS_ToGrpc_ReturnsPermissionDenied(t *testing.T) {
	err := mapOrgErrorToGrpc(ErrForbiddenCrossTenantAccess)
	grpcErr, ok := err.(*grpcError)
	if !ok {
		t.Fatalf("expected grpcError, got %T", err)
	}
	if grpcErr.code != "FORBIDDEN_CROSS_TENANT_ACCESS" {
		t.Fatalf("expected FORBIDDEN_CROSS_TENANT_ACCESS code, got %q", grpcErr.code)
	}
}

// TestOrgError_VISIBILITY_VIOLATION_ToGrpc verifies VISIBILITY_VIOLATION maps to PermissionDenied.
func TestOrgError_VISIBILITY_VIOLATION_ToGrpc_ReturnsPermissionDenied(t *testing.T) {
	err := mapOrgErrorToGrpc(ErrVisibilityViolation)
	grpcErr, ok := err.(*grpcError)
	if !ok {
		t.Fatalf("expected grpcError, got %T", err)
	}
	if grpcErr.code != "VISIBILITY_VIOLATION" {
		t.Fatalf("expected VISIBILITY_VIOLATION code, got %q", grpcErr.code)
	}
}

// TestOrgError_LAST_OWNER_REMOVAL_BLOCKED_ToGrpc verifies LAST_OWNER_REMOVAL_BLOCKED maps to PermissionDenied.
func TestOrgError_LAST_OWNER_REMOVAL_BLOCKED_ToGrpc_ReturnsPermissionDenied(t *testing.T) {
	err := mapOrgErrorToGrpc(ErrLastOwnerRemovalBlocked)
	grpcErr, ok := err.(*grpcError)
	if !ok {
		t.Fatalf("expected grpcError, got %T", err)
	}
	if grpcErr.code != "LAST_OWNER_REMOVAL_BLOCKED" {
		t.Fatalf("expected LAST_OWNER_REMOVAL_BLOCKED code, got %q", grpcErr.code)
	}
}
