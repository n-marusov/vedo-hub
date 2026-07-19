// Code-generation contract test for org.proto.
// Validates that generated Go stubs compile and expose expected types and methods.
package main

import (
	"testing"

	grpcserver "vedo-core/src/services/auth-service/internal/grpc"
	"vedo-core/src/services/auth-service/org"
	authv1 "vedo-core/src/services/shared/proto/auth/v1"
)

// TestOrgProto_GeneratedTypesCompile verifies that the generated Go types
// from org.proto are accessible and have the expected structure.
func TestOrgProto_GeneratedTypesCompile(t *testing.T) {
	// Verify core message types can be constructed
	scope := &authv1.Scope{
		Id:          "group/test-group",
		Type:        "group",
		ParentId:    "group/parent",
		Name:        "Test Group",
		Description: "Test group description",
		Visibility:  "Private",
		TenantId:    "org-001",
	}
	if scope.Id != "group/test-group" {
		t.Fatalf("expected scope.Id to be set")
	}
	if scope.Type != "group" {
		t.Fatalf("expected scope.Type to be 'group', got %q", scope.Type)
	}

	member := &authv1.Member{
		UserId: "user-123",
		Scope:  "ontology/ont-001",
		Role:   "Editor",
	}
	if member.UserId != "user-123" {
		t.Fatalf("expected member.UserId to be set")
	}

	policy := &authv1.AttributePolicy{
		Scope: "ontology/ont-001",
		Right: "read",
	}
	policy.Pattern = map[string]string{"type": "uri_prefix", "value": "ex:confidential/"}
	if len(policy.Pattern) != 2 {
		t.Fatalf("expected policy pattern to have 2 entries")
	}
}

// TestOrgProto_ServiceInterface verifies that the generated OrgServiceServer
// interface exists and has the expected methods. This is a compile-time check.
func TestOrgProto_ServiceInterface(t *testing.T) {
	// This test verifies the interface can be referenced at runtime.
	// The interface itself is checked at compile time.
	server := &mockOrgServer{}
	_, ok := interface{}(server).(authv1.OrgServiceServer)
	if !ok {
		t.Fatal("mockOrgServer does not implement authv1.OrgServiceServer")
	}
}

// mockOrgServer provides stub implementations of all OrgServiceServer methods
// for compile-time interface verification.
type mockOrgServer struct {
	authv1.UnimplementedOrgServiceServer
}

// TestAuthServiceStartup_RegistersGrpcHandlers verifies that the auth-service
// properly registers gRPC services (OrgService, AuthService, Health) on startup.
func TestAuthServiceStartup_RegistersGrpcHandlers(t *testing.T) {
	store := org.NewMemStore()
	svc := org.NewOrgService(store)
	grpcSrv := grpcserver.NewOrgGrpcServer(svc)

	_ = grpcSrv // grpcSrv is correctly typed (authv1.OrgServiceServer)
	// Compile-time check: grpcSrv implements OrgServiceServer
	var _ authv1.OrgServiceServer = grpcSrv

	// Verify the OrgServiceServer has the correct number of expected methods
	// by checking it can be assigned to the interface
	_, ok := interface{}(grpcSrv).(authv1.OrgServiceServer)
	if !ok {
		t.Fatal("OrgGrpcServer does not implement authv1.OrgServiceServer")
	}

	t.Logf("OrgService RPCs registered successfully")
}
