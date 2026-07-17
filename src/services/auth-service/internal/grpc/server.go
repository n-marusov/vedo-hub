// Package grpcserver provides the gRPC server implementation for auth-service.
//
// The AuthServiceServer implementation will be enabled once protoc generates
// the Go code from proto/auth/v1/auth.proto.  Run `make proto-generate`
// from the repo root and then uncomment the imports and registration in main.go.
//
// Usage (after proto generation):
//
//	import authv1 "vedo-core/src/services/shared/proto/auth/v1"
//	authv1.RegisterAuthServiceServer(grpcSrv, &AuthGrpcServer{})
package grpcserver

// AuthGrpcServer implements the authv1.AuthServiceServer gRPC interface.
// Full implementation will be added in a later phase.
//
// After proto generation, uncomment the import and implement:
//
//	type AuthGrpcServer struct {
//		authv1.UnimplementedAuthServiceServer
//	}
//
//	func (s *AuthGrpcServer) TokenIntrospect(ctx context.Context, req *authv1.TokenIntrospectRequest) (*authv1.TokenIntrospectResponse, error) {
//		return nil, status.Errorf(codes.Unimplemented, "method TokenIntrospect not implemented")
//	}
//
//	func (s *AuthGrpcServer) CheckPermission(ctx context.Context, req *authv1.CheckPermissionRequest) (*authv1.CheckPermissionResponse, error) {
//		return nil, status.Errorf(codes.Unimplemented, "method CheckPermission not implemented")
//	}
//
//	func (s *AuthGrpcServer) GetUserRoles(ctx context.Context, req *authv1.GetUserRolesRequest) (*authv1.GetUserRolesResponse, error) {
//		return nil, status.Errorf(codes.Unimplemented, "method GetUserRoles not implemented")
//	}
//
//
//	func (s *AuthGrpcServer) ListPermissions(ctx context.Context, req *authv1.ListPermissionsRequest) (*authv1.ListPermissionsResponse, error) {
//		return nil, status.Errorf(codes.Unimplemented, "method ListPermissions not implemented")
//	}
