package proxy

import (
	"context"
	"log/slog"
	"time"

	versioningv1 "vedo-core/src/services/shared/proto/versioning/v1"
)

// VersioningServiceClient wraps the generated proto stub with lazy initialization,
// logging, and context deadlines from the gRPC connection pool.
type VersioningServiceClient struct {
	pool        *GrpcClientPool
	address     string
	stub        versioningv1.VersioningServiceClient
	lastRefresh time.Time
}

// NewVersioningServiceClient creates a new VersioningServiceClient wrapper.
func NewVersioningServiceClient(pool *GrpcClientPool, address string) *VersioningServiceClient {
	return &VersioningServiceClient{
		pool:    pool,
		address: address,
	}
}

// getStub lazily initializes the gRPC client stub.
func (c *VersioningServiceClient) getStub() (versioningv1.VersioningServiceClient, error) {
	if c.stub != nil && time.Since(c.lastRefresh) < 5*time.Minute {
		return c.stub, nil
	}
	conn, err := c.pool.GetConn(c.address)
	if err != nil {
		slog.Error("grpc.versioning.stub_init_failed", "address", c.address, "error", err)
		return nil, err
	}
	c.stub = versioningv1.NewVersioningServiceClient(conn)
	c.lastRefresh = time.Now()
	slog.Debug("grpc.versioning.stub_initialized", "address", c.address)
	return c.stub, nil
}

// CreateCommit sends a CreateCommitRequest to the versioning service.
func (c *VersioningServiceClient) CreateCommit(ctx context.Context, req *versioningv1.CreateCommitRequest) (*versioningv1.CreateCommitResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.versioning.create_commit", "ontology_id", req.GetOntologyId(), "branch_id", req.GetBranchId())
	resp, err := stub.CreateCommit(ctx, req)
	if err != nil {
		slog.Error("grpc.versioning.create_commit_failed", "ontology_id", req.GetOntologyId(), "error", err)
		return nil, err
	}
	return resp, nil
}

// ListCommits sends a ListCommitsRequest to the versioning service.
func (c *VersioningServiceClient) ListCommits(ctx context.Context, req *versioningv1.ListCommitsRequest) (*versioningv1.ListCommitsResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.versioning.list_commits", "ontology_id", req.GetOntologyId(), "branch_id", req.GetBranchId())
	resp, err := stub.ListCommits(ctx, req)
	if err != nil {
		slog.Error("grpc.versioning.list_commits_failed", "ontology_id", req.GetOntologyId(), "error", err)
		return nil, err
	}
	return resp, nil
}

// CreateBranch sends a CreateBranchRequest to the versioning service.
func (c *VersioningServiceClient) CreateBranch(ctx context.Context, req *versioningv1.CreateBranchRequest) (*versioningv1.CreateBranchResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.versioning.create_branch", "ontology_id", req.GetOntologyId(), "name", req.GetName())
	resp, err := stub.CreateBranch(ctx, req)
	if err != nil {
		slog.Error("grpc.versioning.create_branch_failed", "ontology_id", req.GetOntologyId(), "error", err)
		return nil, err
	}
	return resp, nil
}

// DiffCommits sends a GetDiffRequest to the versioning service.
func (c *VersioningServiceClient) DiffCommits(ctx context.Context, req *versioningv1.GetDiffRequest) (*versioningv1.GetDiffResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.versioning.diff_commits", "base_commit", req.GetBaseCommitId(), "target_commit", req.GetTargetCommitId())
	resp, err := stub.GetDiff(ctx, req)
	if err != nil {
		slog.Error("grpc.versioning.diff_commits_failed", "error", err)
		return nil, err
	}
	return resp, nil
}
