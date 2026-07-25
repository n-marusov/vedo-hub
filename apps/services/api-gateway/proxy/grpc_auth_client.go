package proxy

import (
	"context"
	"log/slog"
	"time"

	authv1 "vedo-core/src/services/shared/proto/auth/v1"
)

// AuthServiceClient wraps the generated proto stub with lazy initialization,
// logging, and context deadlines from the gRPC connection pool.
type AuthServiceClient struct {
	pool        *GrpcClientPool
	address     string
	stub        authv1.AuthServiceClient
	lastRefresh time.Time
}

// NewAuthServiceClient creates a new AuthServiceClient wrapper.
func NewAuthServiceClient(pool *GrpcClientPool, address string) *AuthServiceClient {
	return &AuthServiceClient{
		pool:    pool,
		address: address,
	}
}

// getStub lazily initializes the gRPC client stub.
func (c *AuthServiceClient) getStub() (authv1.AuthServiceClient, error) {
	if c.stub != nil && time.Since(c.lastRefresh) < 5*time.Minute {
		return c.stub, nil
	}
	conn, err := c.pool.GetConn(c.address)
	if err != nil {
		slog.Error("grpc.auth.stub_init_failed", "address", c.address, "error", err)
		return nil, err
	}
	c.stub = authv1.NewAuthServiceClient(conn)
	c.lastRefresh = time.Now()
	slog.Debug("grpc.auth.stub_initialized", "address", c.address)
	return c.stub, nil
}

// TokenIntrospect sends a TokenIntrospectRequest to the auth service.
func (c *AuthServiceClient) TokenIntrospect(ctx context.Context, req *authv1.TokenIntrospectRequest) (*authv1.TokenIntrospectResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.auth.token_introspect", "token_length", len(req.GetToken()))
	resp, err := stub.TokenIntrospect(ctx, req)
	if err != nil {
		slog.Error("grpc.auth.token_introspect_failed", "error", err)
		return nil, err
	}
	return resp, nil
}

// CheckPermission sends a CheckPermissionRequest to the auth service.
func (c *AuthServiceClient) CheckPermission(ctx context.Context, req *authv1.CheckPermissionRequest) (*authv1.CheckPermissionResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.auth.check_permission", "subject", req.GetSubject(), "action", req.GetAction(), "resource_id", req.GetResourceId())
	resp, err := stub.CheckPermission(ctx, req)
	if err != nil {
		slog.Error("grpc.auth.check_permission_failed", "subject", req.GetSubject(), "action", req.GetAction(), "error", err)
		return nil, err
	}
	return resp, nil
}
