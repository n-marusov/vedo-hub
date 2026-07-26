// Package grpcserver provides the gRPC server implementation for auth-service.
//
// AuthGrpcServer implements authv1.AuthServiceServer (legacy auth RPCs).
// OrgGrpcServer implements authv1.OrgServiceServer (org management RPCs).
package grpcserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"log/slog"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"vedo-core/src/services/auth-service/org"
	authv1 "vedo-core/src/services/shared/proto/auth/v1"
)

// ============================================================================
// Auth Service Server (legacy)
// ============================================================================

// AuthGrpcServer implements the authv1.AuthServiceServer gRPC interface.
type AuthGrpcServer struct {
	authv1.UnimplementedAuthServiceServer
}

func (s *AuthGrpcServer) TokenIntrospect(ctx context.Context, req *authv1.TokenIntrospectRequest) (*authv1.TokenIntrospectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TokenIntrospect not implemented")
}

func (s *AuthGrpcServer) CheckPermission(ctx context.Context, req *authv1.CheckPermissionRequest) (*authv1.CheckPermissionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckPermission not implemented")
}

func (s *AuthGrpcServer) GetUserRoles(ctx context.Context, req *authv1.GetUserRolesRequest) (*authv1.GetUserRolesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUserRoles not implemented")
}

func (s *AuthGrpcServer) ListPermissions(ctx context.Context, req *authv1.ListPermissionsRequest) (*authv1.ListPermissionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPermissions not implemented")
}

// ============================================================================
// Org Service Server
// ============================================================================

// OrgGrpcServer implements the authv1.OrgServiceServer gRPC interface.
type OrgGrpcServer struct {
	authv1.UnimplementedOrgServiceServer
	svc *org.OrgService
}

// NewOrgGrpcServer creates a new OrgGrpcServer with the given OrgService.
func NewOrgGrpcServer(svc *org.OrgService) *OrgGrpcServer {
	return &OrgGrpcServer{svc: svc}
}

// ============================================================================
// Group RPCs
// ============================================================================

func (s *OrgGrpcServer) CreateGroup(ctx context.Context, req *authv1.CreateGroupRequest) (*authv1.CreateGroupResponse, error) {
	requesterID := extractUserID(ctx)

	groupID := newUUID()
	visibility := normalizeVisibility(req.GetVisibility())
	scopeNode := org.ScopeNode{
		ID:          groupID,
		Type:        org.ScopeGroup,
		TenantID:    req.OrganizationId,
		Name:        req.Name,
		Description: req.Description,
		Visibility:  org.Visibility(visibility),
	}
	if req.ParentId != "" {
		scopeNode.ParentID = req.ParentId
	}

	slog.Debug("grpc.org.create_group", "name", req.GetName(), "visibility", scopeNode.Visibility)

	if err := s.svc.CreateScope(requesterID, scopeNode); err != nil {
		return nil, mapOrgError(err)
	}

	scope, _ := s.svc.Store().GetScope(scopeNode.ID)
	log.Printf(`{"event":"grpc.request","method":"CreateGroup","scope":"%s","user":"%s"}`, scopeNode.ID, requesterID)
	return &authv1.CreateGroupResponse{
		Group: scopeNodeToProto(scope),
	}, nil
}

// normalizeVisibility converts a visibility string to canonical PascalCase.
// Accepts "private", "internal", "public" (lowercase or any case) and
// returns "Private", "Internal", "Public". Unknown values are returned unchanged.
func normalizeVisibility(v string) string {
	if v == "" {
		return string(org.VisibilityPrivate)
	}
	switch strings.ToLower(v) {
	case "private":
		return string(org.VisibilityPrivate)
	case "internal":
		return string(org.VisibilityInternal)
	case "public":
		return string(org.VisibilityPublic)
	default:
		return v
	}
}

func (s *OrgGrpcServer) GetGroup(ctx context.Context, req *authv1.GetGroupRequest) (*authv1.GetGroupResponse, error) {
	scope, err := s.svc.Store().GetScope(req.Id)
	if err != nil {
		return nil, mapOrgError(err)
	}
	if scope == nil {
		return nil, status.Errorf(codes.NotFound, "SCOPE_NOT_FOUND: group %s does not exist", req.Id)
	}
	return &authv1.GetGroupResponse{Group: scopeNodeToProto(scope)}, nil
}

func (s *OrgGrpcServer) ListGroups(ctx context.Context, req *authv1.ListGroupsRequest) (*authv1.ListGroupsResponse, error) {
	scopes, err := s.svc.Store().ListAllScopes()
	if err != nil {
		return nil, mapOrgError(err)
	}

	var groups []*authv1.Scope
	for i := range scopes {
		if scopes[i].Type == org.ScopeGroup {
			groups = append(groups, scopeNodeToProto(&scopes[i]))
		}
	}

	return &authv1.ListGroupsResponse{Groups: groups}, nil
}

func (s *OrgGrpcServer) UpdateGroup(ctx context.Context, req *authv1.UpdateGroupRequest) (*authv1.UpdateGroupResponse, error) {
	requesterID := extractUserID(ctx)

	if err := s.svc.CheckAccess(requesterID, req.Id, "update_membership"); err != nil {
		return nil, mapOrgError(err)
	}

	scope, err := s.svc.Store().GetScope(req.Id)
	if err != nil || scope == nil {
		return nil, status.Errorf(codes.NotFound, "SCOPE_NOT_FOUND: group %s does not exist", req.Id)
	}

	if req.GetName() != "" {
		scope.Name = req.GetName()
	}
	if req.GetDescription() != "" {
		scope.Description = req.GetDescription()
	}
	if req.GetVisibility() != "" {
		scope.Visibility = org.Visibility(normalizeVisibility(req.GetVisibility()))
	}

	slog.Debug("grpc.org.update_group", "id", req.GetId(), "name", scope.Name, "visibility", scope.Visibility)

	if err := s.svc.Store().UpsertScope(*scope); err != nil {
		return nil, mapOrgError(err)
	}
	s.svc.Cache().InvalidateScope(scope.ID)

	updated, _ := s.svc.Store().GetScope(scope.ID)
	log.Printf(`{"event":"grpc.request","method":"UpdateGroup","scope":"%s","user":"%s"}`, scope.ID, requesterID)
	return &authv1.UpdateGroupResponse{Group: scopeNodeToProto(updated)}, nil
}

func (s *OrgGrpcServer) DeleteGroup(ctx context.Context, req *authv1.DeleteGroupRequest) (*authv1.DeleteGroupResponse, error) {
	requesterID := extractUserID(ctx)

	if err := s.svc.CheckAccess(requesterID, req.Id, "delete_scope"); err != nil {
		return nil, mapOrgError(err)
	}

	if err := s.svc.Store().DeleteScope(req.Id); err != nil {
		return nil, mapOrgError(err)
	}

	s.svc.Cache().InvalidateScope(req.Id)
	log.Printf(`{"event":"grpc.request","method":"DeleteGroup","scope":"%s"}`, req.Id)
	return &authv1.DeleteGroupResponse{}, nil
}

func (s *OrgGrpcServer) ListChildGroups(ctx context.Context, req *authv1.ListChildGroupsRequest) (*authv1.ListChildGroupsResponse, error) {
	children, err := s.svc.Store().ListChildScopes(req.ParentId)
	if err != nil {
		return nil, mapOrgError(err)
	}

	var groups []*authv1.Scope
	for i := range children {
		if children[i].Type == org.ScopeGroup {
			groups = append(groups, scopeNodeToProto(&children[i]))
		}
	}

	return &authv1.ListChildGroupsResponse{Groups: groups}, nil
}

// ============================================================================
// Project RPCs
// ============================================================================

func (s *OrgGrpcServer) CreateProject(ctx context.Context, req *authv1.CreateProjectRequest) (*authv1.CreateProjectResponse, error) {
	requesterID := extractUserID(ctx)

	projectID := newUUID()
	node := org.ScopeNode{
		ID:          projectID,
		Type:        org.ScopeProject,
		TenantID:    req.OrganizationId,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.svc.CreateScope(requesterID, node); err != nil {
		return nil, mapOrgError(err)
	}

	// Generate a UUID v4 for the paired Ontology (REQ-FUN.DATA.ontology-identifier-standard).
	ontologyID := newUUID()
	if err := s.svc.Store().CreateOntology(org.Ontology{
		ProjectScope: projectID,
		OntologyID:   ontologyID,
	}); err != nil {
		slog.Error("project.create.ontology_pairing_failed", "project_id", projectID, "err", err)
		// Best-effort cleanup: remove the scope if the ontology pairing failed.
		if delErr := s.svc.Store().DeleteScope(projectID); delErr != nil {
			slog.Error("project.create.cleanup_failed", "project_id", projectID, "err", delErr)
		}
		return nil, status.Errorf(codes.Internal, "PROJECT_CREATE_ONTOLOGY_PAIRING_FAILED: %v", err)
	}

	scope, _ := s.svc.Store().GetScope(projectID)
	proto := scopeNodeToProto(scope)
	if proto != nil {
		proto.OntologyId = ontologyID
	}
	slog.Info("project.create", "project_id", projectID, "ontology_id", ontologyID)
	log.Printf(`{"event":"grpc.request","method":"CreateProject","scope":"%s","ontology_id":"%s"}`, projectID, ontologyID)
	return &authv1.CreateProjectResponse{Project: proto}, nil
}

func (s *OrgGrpcServer) GetProject(ctx context.Context, req *authv1.GetProjectRequest) (*authv1.GetProjectResponse, error) {
	scope, err := s.svc.Store().GetScope(req.Id)
	if err != nil {
		return nil, mapOrgError(err)
	}
	// Accept both 'project' (canonical) and 'ontology' (legacy alias) types.
	if scope == nil || (scope.Type != org.ScopeProject && scope.Type != org.ScopeOntology) {
		return nil, status.Errorf(codes.NotFound, "SCOPE_NOT_FOUND: project %s does not exist", req.Id)
	}
	proto := scopeNodeToProto(scope)
	if proto != nil {
		if ont, _ := s.svc.Store().GetOntologyByProjectScope(scope.ID); ont != nil {
			proto.OntologyId = ont.OntologyID
		}
	}
	return &authv1.GetProjectResponse{Project: proto}, nil
}

func (s *OrgGrpcServer) ListProjects(ctx context.Context, req *authv1.ListProjectsRequest) (*authv1.ListProjectsResponse, error) {
	scopes, err := s.svc.Store().ListAllScopes()
	if err != nil {
		return nil, mapOrgError(err)
	}

	var projects []*authv1.Scope
	for i := range scopes {
		// Accept both 'project' (canonical) and 'ontology' (legacy alias) types.
		if scopes[i].Type == org.ScopeProject || scopes[i].Type == org.ScopeOntology {
			proto := scopeNodeToProto(&scopes[i])
			if proto != nil {
				if ont, _ := s.svc.Store().GetOntologyByProjectScope(scopes[i].ID); ont != nil {
					proto.OntologyId = ont.OntologyID
				}
			}
			projects = append(projects, proto)
		}
	}

	return &authv1.ListProjectsResponse{Projects: projects, Total: int32(len(projects))}, nil
}

func (s *OrgGrpcServer) UpdateProject(ctx context.Context, req *authv1.UpdateProjectRequest) (*authv1.UpdateProjectResponse, error) {
	requesterID := extractUserID(ctx)

	if err := s.svc.CheckAccess(requesterID, req.Id, "update_membership"); err != nil {
		return nil, mapOrgError(err)
	}

	scope, err := s.svc.Store().GetScope(req.Id)
	if err != nil || scope == nil {
		return nil, status.Errorf(codes.NotFound, "SCOPE_NOT_FOUND: project %s does not exist", req.Id)
	}

	log.Printf(`{"event":"grpc.request","method":"UpdateProject","scope":"%s"}`, req.Id)
	return &authv1.UpdateProjectResponse{Project: scopeNodeToProto(scope)}, nil
}

func (s *OrgGrpcServer) DeleteProject(ctx context.Context, req *authv1.DeleteProjectRequest) (*authv1.DeleteProjectResponse, error) {
	requesterID := extractUserID(ctx)

	if err := s.svc.CheckAccess(requesterID, req.Id, "delete_scope"); err != nil {
		return nil, mapOrgError(err)
	}

	if err := s.svc.Store().DeleteScope(req.Id); err != nil {
		return nil, mapOrgError(err)
	}

	s.svc.Cache().InvalidateScope(req.Id)
	return &authv1.DeleteProjectResponse{}, nil
}

// ============================================================================
// Membership RPCs
// ============================================================================

func (s *OrgGrpcServer) AddMember(ctx context.Context, req *authv1.AddMemberRequest) (*authv1.AddMemberResponse, error) {
	requesterID := extractUserID(ctx)

	m, _, err := s.svc.UpdateMembership(requesterID, req.Scope, req.UserId, req.Role)
	if err != nil {
		return nil, mapOrgError(err)
	}

	log.Printf(`{"event":"grpc.request","method":"AddMember","scope":"%s","target":"%s"}`, req.Scope, req.UserId)
	return &authv1.AddMemberResponse{Member: membershipToProto(m)}, nil
}

func (s *OrgGrpcServer) UpdateMemberRole(ctx context.Context, req *authv1.UpdateMemberRoleRequest) (*authv1.UpdateMemberRoleResponse, error) {
	requesterID := extractUserID(ctx)

	m, _, err := s.svc.UpdateMembership(requesterID, req.Scope, req.UserId, req.Role)
	if err != nil {
		return nil, mapOrgError(err)
	}

	log.Printf(`{"event":"grpc.request","method":"UpdateMemberRole","scope":"%s","target":"%s"}`, req.Scope, req.UserId)
	return &authv1.UpdateMemberRoleResponse{Member: membershipToProto(m)}, nil
}

func (s *OrgGrpcServer) RemoveMember(ctx context.Context, req *authv1.RemoveMemberRequest) (*authv1.RemoveMemberResponse, error) {
	requesterID := extractUserID(ctx)

	if err := s.svc.CheckAccess(requesterID, req.Scope, "update_membership"); err != nil {
		return nil, mapOrgError(err)
	}

	if err := s.svc.Store().DeleteMembership(req.Scope, req.UserId); err != nil {
		return nil, mapOrgError(err)
	}

	s.svc.Cache().InvalidateScope(req.Scope)
	log.Printf(`{"event":"grpc.request","method":"RemoveMember","scope":"%s","target":"%s"}`, req.Scope, req.UserId)
	return &authv1.RemoveMemberResponse{}, nil
}

func (s *OrgGrpcServer) ListMembers(ctx context.Context, req *authv1.ListMembersRequest) (*authv1.ListMembersResponse, error) {
	members, err := s.svc.Store().GetMemberships(req.Scope)
	if err != nil {
		return nil, mapOrgError(err)
	}

	var protoMembers []*authv1.Member
	for i := range members {
		protoMembers = append(protoMembers, membershipToProto(&members[i]))
	}

	return &authv1.ListMembersResponse{Members: protoMembers}, nil
}

// ============================================================================
// Visibility RPCs
// ============================================================================

func (s *OrgGrpcServer) SetVisibility(ctx context.Context, req *authv1.SetVisibilityRequest) (*authv1.SetVisibilityResponse, error) {
	requesterID := extractUserID(ctx)

	vis := org.Visibility(req.Visibility)
	if err := s.svc.SetVisibility(requesterID, req.Scope, vis); err != nil {
		return nil, mapOrgError(err)
	}

	s.svc.Cache().InvalidateScope(req.Scope)
	log.Printf(`{"event":"grpc.request","method":"SetVisibility","scope":"%s","visibility":"%s"}`, req.Scope, req.Visibility)
	return &authv1.SetVisibilityResponse{}, nil
}

func (s *OrgGrpcServer) GetVisibility(ctx context.Context, req *authv1.GetVisibilityRequest) (*authv1.GetVisibilityResponse, error) {
	vis, err := s.svc.Store().GetVisibility(req.Scope)
	if err != nil {
		return nil, mapOrgError(err)
	}

	return &authv1.GetVisibilityResponse{Visibility: string(vis)}, nil
}

// ============================================================================
// Policy RPCs
// ============================================================================

func (s *OrgGrpcServer) CreatePolicy(ctx context.Context, req *authv1.CreatePolicyRequest) (*authv1.CreatePolicyResponse, error) {
	requesterID := extractUserID(ctx)

	policy := org.AttributePolicy{
		Scope:   req.Scope,
		Pattern: req.Pattern,
		Right:   req.Right,
	}

	_, err := s.svc.SavePolicy(requesterID, policy)
	if err != nil {
		return nil, mapOrgError(err)
	}

	log.Printf(`{"event":"grpc.request","method":"CreatePolicy","scope":"%s","right":"%s"}`, req.Scope, req.Right)
	return &authv1.CreatePolicyResponse{}, nil
}

func (s *OrgGrpcServer) ListPolicies(ctx context.Context, req *authv1.ListPoliciesRequest) (*authv1.ListPoliciesResponse, error) {
	policies, err := s.svc.Store().GetPolicies(req.Scope)
	if err != nil {
		return nil, mapOrgError(err)
	}

	var protoPolicies []*authv1.AttributePolicy
	for i := range policies {
		protoPolicies = append(protoPolicies, policyToProto(&policies[i]))
	}

	return &authv1.ListPoliciesResponse{Policies: protoPolicies}, nil
}

func (s *OrgGrpcServer) DeletePolicy(ctx context.Context, req *authv1.DeletePolicyRequest) (*authv1.DeletePolicyResponse, error) {
	if err := s.svc.Store().DeletePolicy(req.Scope, req.PolicyId); err != nil {
		return nil, mapOrgError(err)
	}

	return &authv1.DeletePolicyResponse{}, nil
}

// ============================================================================
// Access Check RPC
// ============================================================================

func (s *OrgGrpcServer) CheckAccess(ctx context.Context, req *authv1.CheckAccessRequest) (*authv1.CheckAccessResponse, error) {
	if err := s.svc.CheckAccess(req.UserId, req.Scope, req.Action); err != nil {
		return &authv1.CheckAccessResponse{
			Allowed: false,
			Reason:  err.Error(),
		}, nil
	}

	return &authv1.CheckAccessResponse{
		Allowed: true,
	}, nil
}

// ============================================================================
// Fork Project RPC
// ============================================================================

func (s *OrgGrpcServer) ForkProject(ctx context.Context, req *authv1.ForkProjectRequest) (*authv1.ForkProjectResponse, error) {
	requesterID := extractUserID(ctx)
	sourceID := req.SourceProjectId

	// Generate IDs for the new project
	newProjectID := newUUID()
	ontologyID := newUUID()

	// Create new project scope with upstream link
	node := org.ScopeNode{
		ID:                newProjectID,
		Type:              org.ScopeProject,
		TenantID:          req.OrganizationId,
		UpstreamProjectID: sourceID, // link to source
	}

	if err := s.svc.CreateScope(requesterID, node); err != nil {
		return nil, mapOrgError(err)
	}

	// Create paired Ontology
	if err := s.svc.Store().CreateOntology(org.Ontology{
		ProjectScope: newProjectID,
		OntologyID:   ontologyID,
	}); err != nil {
		// Compensating action: delete new scope
		_ = s.svc.Store().DeleteScope(newProjectID)
		return nil, status.Errorf(codes.Internal, "FORK_ONTOLOGY_PAIRING_FAILED: %v", err)
	}

	// Grant caller Owner role
	mem := org.OrgMembership{
		UserID: requesterID,
		Scope:  newProjectID,
		Role:   "Owner",
	}
	if err := s.svc.Store().UpsertMembership(mem); err != nil {
		_ = s.svc.Store().DeleteScope(newProjectID)
		return nil, status.Errorf(codes.Internal, "FORK_MEMBERSHIP_FAILED: %v", err)
	}

	// Build response
	scope, _ := s.svc.Store().GetScope(newProjectID)
	proto := scopeNodeToProto(scope)
	if proto != nil {
		proto.OntologyId = ontologyID
		proto.UpstreamProjectId = sourceID
	}

	slog.Info("fork.completed", "source", sourceID, "new_project", newProjectID, "ontology_id", ontologyID)
	return &authv1.ForkProjectResponse{Project: proto}, nil
}

// ============================================================================
// Error Mapping
// ============================================================================

// mapOrgError converts an org package error to a gRPC status error.
func mapOrgError(err error) error {
	if oe, ok := err.(*org.OrgError); ok {
		switch oe.Code {
		case "FORBIDDEN_ADMIN_ONLY", "FORBIDDEN_INSUFFICIENT_ROLE",
			"FORBIDDEN_CROSS_TENANT_ACCESS", "VISIBILITY_VIOLATION":
			return status.Errorf(codes.PermissionDenied, "%s: %s", oe.Code, oe.Message)
		case "SCOPE_NOT_FOUND", "FORBIDDEN_OBJECT_NOT_FOUND_OR_ACCESS_DENIED":
			return status.Errorf(codes.NotFound, "%s: %s", oe.Code, oe.Message)
		case "CYCLE_DETECTED":
			return status.Errorf(codes.InvalidArgument, "%s: %s", oe.Code, oe.Message)
		case "POLICY_CONFLICT":
			return status.Errorf(codes.AlreadyExists, "%s: %s", oe.Code, oe.Message)
		default:
			return status.Errorf(codes.Internal, "%s: %s", oe.Code, oe.Message)
		}
	}
	if err == nil {
		return nil
	}
	return status.Errorf(codes.Internal, "INTERNAL: %v", err)
}

// ============================================================================
// Helpers
// ============================================================================

func extractUserID(ctx context.Context) string {
	// In production, extract from gRPC metadata propagated from JWT claims
	return "system"
}

func scopeNodeToProto(s *org.ScopeNode) *authv1.Scope {
	if s == nil {
		return nil
	}
	// Extract the display name from the scope ID.
	// With UUID identifiers (ADR-DES.DATA.uuid-identifiers-for-groups-projects-mandate),
	// the ID is the UUID directly. For backward compatibility with legacy
	// "group/name" format, extract the part after the slash if present.
	name := s.ID
	if idx := strings.Index(s.ID, "/"); idx >= 0 && idx+1 < len(s.ID) {
		name = s.ID[idx+1:]
	}
	return &authv1.Scope{
		Id:                s.ID,
		Type:              string(s.Type),
		Name:              name,
		ParentId:          s.ParentID,
		Visibility:        string(s.Visibility),
		TenantId:          s.TenantID,
		UpstreamProjectId: s.UpstreamProjectID,
	}
}

func membershipToProto(m *org.OrgMembership) *authv1.Member {
	if m == nil {
		return nil
	}
	return &authv1.Member{
		UserId: m.UserID,
		Scope:  m.Scope,
		Role:   m.Role,
	}
}

func policyToProto(p *org.AttributePolicy) *authv1.AttributePolicy {
	if p == nil {
		return nil
	}
	return &authv1.AttributePolicy{
		Scope:   p.Scope,
		Pattern: p.Pattern,
		Right:   p.Right,
	}
}

// newUUID generates a UUID v4 (RFC 4122) using crypto/rand.
// Used for ontology_id per REQ-FUN.DATA.ontology-identifier-standard.
func newUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand.Read never fails on modern kernels; this fallback
		// returns a placeholder if something goes wrong.
		return "00000000-0000-0000-0000-000000000000"
	}
	// Set version 4 bits and RFC 4122 variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	s := hex.EncodeToString(b)
	return s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]
}
