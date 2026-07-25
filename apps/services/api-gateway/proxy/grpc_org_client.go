package proxy

// OrgServiceClient wraps the generated org gRPC stub with lazy initialization,
// logging, and context deadlines from the gRPC connection pool.
//
// Covers: group/project CRUD, membership management, visibility, policies.

import (
	"context"
	"log/slog"
	"time"

	authv1 "vedo-core/src/services/shared/proto/auth/v1"

	"google.golang.org/grpc/metadata"
)

// OrgServiceClient wraps the generated OrgServiceClient stub for org management RPCs.
type OrgServiceClient struct {
	pool        *GrpcClientPool
	address     string
	stub        authv1.OrgServiceClient
	lastRefresh time.Time
}

// NewOrgServiceClient creates a new OrgServiceClient wrapper.
func NewOrgServiceClient(pool *GrpcClientPool, address string) *OrgServiceClient {
	return &OrgServiceClient{
		pool:    pool,
		address: address,
	}
}

// getStub lazily initializes the gRPC client stub.
func (c *OrgServiceClient) getStub() (authv1.OrgServiceClient, error) {
	if c.stub != nil && time.Since(c.lastRefresh) < 5*time.Minute {
		return c.stub, nil
	}
	conn, err := c.pool.GetConn(c.address)
	if err != nil {
		slog.Error("grpc.org.stub_init_failed", "address", c.address, "error", err)
		return nil, err
	}
	c.stub = authv1.NewOrgServiceClient(conn)
	c.lastRefresh = time.Now()
	slog.Debug("grpc.org.stub_initialized", "address", c.address)
	return c.stub, nil
}

// withAuth injects the JWT from context into gRPC metadata.
func (c *OrgServiceClient) withAuth(ctx context.Context, token string) context.Context {
	if token != "" {
		return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	}
	return ctx
}

// ============================================================================
// Group RPCs
// ============================================================================

func (c *OrgServiceClient) CreateGroup(ctx context.Context, req *authv1.CreateGroupRequest, token string) (*authv1.CreateGroupResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 10*time.Second)
	defer cancel()
	slog.Debug("grpc.org.create_group", "name", req.GetName())
	return stub.CreateGroup(ctx, req)
}

func (c *OrgServiceClient) GetGroup(ctx context.Context, id, token string) (*authv1.GetGroupResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 5*time.Second)
	defer cancel()
	return stub.GetGroup(ctx, &authv1.GetGroupRequest{Id: id})
}

func (c *OrgServiceClient) ListGroups(ctx context.Context, req *authv1.ListGroupsRequest, token string) (*authv1.ListGroupsResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 5*time.Second)
	defer cancel()
	return stub.ListGroups(ctx, req)
}

func (c *OrgServiceClient) UpdateGroup(ctx context.Context, req *authv1.UpdateGroupRequest, token string) (*authv1.UpdateGroupResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 10*time.Second)
	defer cancel()
	return stub.UpdateGroup(ctx, req)
}

func (c *OrgServiceClient) DeleteGroup(ctx context.Context, id, token string) (*authv1.DeleteGroupResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 10*time.Second)
	defer cancel()
	return stub.DeleteGroup(ctx, &authv1.DeleteGroupRequest{Id: id})
}

func (c *OrgServiceClient) ListChildGroups(ctx context.Context, parentID, token string) (*authv1.ListChildGroupsResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 5*time.Second)
	defer cancel()
	return stub.ListChildGroups(ctx, &authv1.ListChildGroupsRequest{ParentId: parentID})
}

// ============================================================================
// Project RPCs
// ============================================================================

func (c *OrgServiceClient) CreateProject(ctx context.Context, req *authv1.CreateProjectRequest, token string) (*authv1.CreateProjectResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 10*time.Second)
	defer cancel()
	return stub.CreateProject(ctx, req)
}

func (c *OrgServiceClient) GetProject(ctx context.Context, id, token string) (*authv1.GetProjectResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 5*time.Second)
	defer cancel()
	return stub.GetProject(ctx, &authv1.GetProjectRequest{Id: id})
}

func (c *OrgServiceClient) ListProjects(ctx context.Context, req *authv1.ListProjectsRequest, token string) (*authv1.ListProjectsResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 5*time.Second)
	defer cancel()
	return stub.ListProjects(ctx, req)
}

func (c *OrgServiceClient) UpdateProject(ctx context.Context, req *authv1.UpdateProjectRequest, token string) (*authv1.UpdateProjectResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 10*time.Second)
	defer cancel()
	return stub.UpdateProject(ctx, req)
}

func (c *OrgServiceClient) DeleteProject(ctx context.Context, id, token string) (*authv1.DeleteProjectResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 10*time.Second)
	defer cancel()
	return stub.DeleteProject(ctx, &authv1.DeleteProjectRequest{Id: id})
}

// ============================================================================
// Membership RPCs
// ============================================================================

func (c *OrgServiceClient) AddMember(ctx context.Context, req *authv1.AddMemberRequest, token string) (*authv1.AddMemberResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 10*time.Second)
	defer cancel()
	return stub.AddMember(ctx, req)
}

func (c *OrgServiceClient) UpdateMemberRole(ctx context.Context, req *authv1.UpdateMemberRoleRequest, token string) (*authv1.UpdateMemberRoleResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 10*time.Second)
	defer cancel()
	return stub.UpdateMemberRole(ctx, req)
}

func (c *OrgServiceClient) RemoveMember(ctx context.Context, req *authv1.RemoveMemberRequest, token string) (*authv1.RemoveMemberResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 10*time.Second)
	defer cancel()
	return stub.RemoveMember(ctx, req)
}

func (c *OrgServiceClient) ListMembers(ctx context.Context, scope, token string) (*authv1.ListMembersResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 5*time.Second)
	defer cancel()
	return stub.ListMembers(ctx, &authv1.ListMembersRequest{Scope: scope})
}

// ============================================================================
// Visibility RPCs
// ============================================================================

func (c *OrgServiceClient) SetVisibility(ctx context.Context, req *authv1.SetVisibilityRequest, token string) (*authv1.SetVisibilityResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 10*time.Second)
	defer cancel()
	return stub.SetVisibility(ctx, req)
}

func (c *OrgServiceClient) GetVisibility(ctx context.Context, scope, token string) (*authv1.GetVisibilityResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 5*time.Second)
	defer cancel()
	return stub.GetVisibility(ctx, &authv1.GetVisibilityRequest{Scope: scope})
}

// ============================================================================
// Policy RPCs
// ============================================================================

func (c *OrgServiceClient) CreatePolicy(ctx context.Context, req *authv1.CreatePolicyRequest, token string) (*authv1.CreatePolicyResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 10*time.Second)
	defer cancel()
	return stub.CreatePolicy(ctx, req)
}

func (c *OrgServiceClient) ListPolicies(ctx context.Context, scope, token string) (*authv1.ListPoliciesResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 5*time.Second)
	defer cancel()
	return stub.ListPolicies(ctx, &authv1.ListPoliciesRequest{Scope: scope})
}

// ============================================================================
// Fork Project RPC
// ============================================================================

func (c *OrgServiceClient) ForkProject(ctx context.Context, req *authv1.ForkProjectRequest, token string) (*authv1.ForkProjectResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 30*time.Second)
	defer cancel()
	slog.Debug("grpc.org.fork_project", "source", req.GetSourceProjectId())
	return stub.ForkProject(ctx, req)
}

func (c *OrgServiceClient) DeletePolicy(ctx context.Context, req *authv1.DeletePolicyRequest, token string) (*authv1.DeletePolicyResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(c.withAuth(ctx, token), 10*time.Second)
	defer cancel()
	return stub.DeletePolicy(ctx, req)
}
