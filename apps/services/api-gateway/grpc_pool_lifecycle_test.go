package main

// Validates: REQ-CON.SECURITY.write-path-invariant

// Regression tests for the gRPC client pool lifecycle contract:
//
//  1. RegisterRoutes(r, pool) MUST NOT close the pool — connections
//     established before route registration must remain usable. The
//     original bug placed `defer grpcPool.Close()` inside RegisterRoutes,
//     which shut the pool down before the HTTP server even started, so gRPC
//     clients could never connect (HTTP proxy masked the failure).
//  2. The graceful shutdown path (pool.Close() alongside srv.Shutdown) MUST
//     close pooled connections — the correct place for pool teardown is the
//     shutdown handler in main.go, not route registration.
//
// These tests are unit-level: the pool dials are non-blocking (grpc.Dial
// without WithBlock returns immediately), so no external infrastructure is
// required and no //go:build integration tag is needed.

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/connectivity"

	"vedo-core/src/services/api-gateway/proxy"
)

// TestGRPCPool_AfterRegisterRoutes_ShouldStayOpen locks in the lifecycle
// fix: connections established before RegisterRoutes must survive route
// registration. This test fails if `defer grpcPool.Close()` is reintroduced
// inside RegisterRoutes.
func TestGRPCPool_AfterRegisterRoutes_ShouldStayOpen(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pool := proxy.NewGrpcClientPool(5 * time.Second)

	// Establish a connection before registration, as the real main.go does
	// (grpcPool := proxy.NewGrpcClientPool(...) then RegisterRoutes(r, pool)).
	conn, err := pool.GetConn("localhost:1")
	if err != nil {
		t.Fatalf("get conn before RegisterRoutes: %v", err)
	}
	defer conn.Close()

	r := gin.New()
	RegisterRoutes(r, pool)

	// If the bug is present, RegisterRoutes closed the pool and conn is now
	// in SHUTDOWN state.
	state := conn.GetState()
	if state == connectivity.Shutdown {
		t.Fatal("gRPC pool was closed during RegisterRoutes — connections must stay open until graceful shutdown")
	}
	t.Logf("pool state after RegisterRoutes: %s (expected: not SHUTDOWN)", state)
}

// TestGRPCPool_AfterShutdown_ShouldClose verifies the graceful shutdown path
// closes pooled connections — mirroring main.go's shutdown handler.
func TestGRPCPool_AfterShutdown_ShouldClose(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pool := proxy.NewGrpcClientPool(5 * time.Second)

	conn, err := pool.GetConn("localhost:1")
	if err != nil {
		t.Fatalf("get conn: %v", err)
	}

	// Simulate the graceful shutdown handler from main.go.
	pool.Close()

	state := conn.GetState()
	if state != connectivity.Shutdown {
		t.Fatalf("expected connection SHUTDOWN after pool.Close(), got %s", state)
	}
}

// TestGRPCPool_NewPool_ShouldStartOpen verifies a freshly created pool is
// not in a closed state before any routes are registered.
func TestGRPCPool_NewPool_ShouldStartOpen(t *testing.T) {
	pool := proxy.NewGrpcClientPool(5 * time.Second)

	conn, err := pool.GetConn("localhost:1")
	if err != nil {
		t.Fatalf("get conn from fresh pool: %v", err)
	}
	defer conn.Close()

	if state := conn.GetState(); state == connectivity.Shutdown {
		t.Fatal("freshly created pool must not be in SHUTDOWN state")
	}
}
