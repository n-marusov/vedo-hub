package proxy

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// GrpcClientPool manages gRPC connections to backend services.
// Connections are pooled and lazily established on first use,
// with periodic health checking to detect stale connections.
type GrpcClientPool struct {
	mu              sync.RWMutex
	conns           map[string]*grpc.ClientConn
	upstreamTimeout time.Duration
}

// NewGrpcClientPool creates a new gRPC connection pool.
// upstreamTimeout is applied to gRPC context deadlines.
func NewGrpcClientPool(upstreamTimeout time.Duration) *GrpcClientPool {
	return &GrpcClientPool{
		conns:           make(map[string]*grpc.ClientConn),
		upstreamTimeout: upstreamTimeout,
	}
}

// GetConn returns a gRPC connection to the named service address.
// Connections are cached and re-used. If the cached connection
// is in TRANSIENT_FAILURE or SHUTDOWN, it is replaced.
func (p *GrpcClientPool) GetConn(address string) (*grpc.ClientConn, error) {
	p.mu.RLock()
	conn, ok := p.conns[address]
	p.mu.RUnlock()

	if ok {
		state := conn.GetState()
		if state != connectivity.TransientFailure && state != connectivity.Shutdown {
			return conn, nil
		}
		slog.Warn("grpc.pool.replacing_stale_conn", "address", address, "state", state.String())
		conn.Close()
	}

	// Establish new connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(50 * 1024 * 1024)), // 50MB
		grpc.WithUserAgent("vedo-api-gateway/0.2.0"),
	}

	newConn, err := grpc.DialContext(ctx, address, dialOpts...)
	if err != nil {
		slog.Error("grpc.pool.dial_failed", "address", address, "error", err)
		return nil, err
	}

	p.mu.Lock()
	// If another goroutine created a connection while we were dialing, use theirs
	if existing, ok := p.conns[address]; ok {
		state := existing.GetState()
		if state != connectivity.TransientFailure && state != connectivity.Shutdown {
			p.mu.Unlock()
			newConn.Close()
			return existing, nil
		}
	}
	p.conns[address] = newConn
	p.mu.Unlock()

	slog.Info("grpc.pool.connected", "address", address)
	return newConn, nil
}

// CheckHealth performs a gRPC health check on the given connection.
// Returns true if the backend reports SERVING.
func (p *GrpcClientPool) CheckHealth(conn *grpc.ClientConn) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthClient := grpc_health_v1.NewHealthClient(conn)
	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		slog.Warn("grpc.pool.health_check_failed", "error", err)
		return false
	}
	return resp.Status == grpc_health_v1.HealthCheckResponse_SERVING
}

// GetUpstreamTimeout returns the configured timeout for upstream calls.
func (p *GrpcClientPool) GetUpstreamTimeout() time.Duration {
	return p.upstreamTimeout
}

// Close gracefully closes all pooled connections.
func (p *GrpcClientPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for addr, conn := range p.conns {
		slog.Debug("grpc.pool.closing", "address", addr)
		if err := conn.Close(); err != nil {
			slog.Warn("grpc.pool.close_error", "address", addr, "error", err)
		}
	}
	p.conns = make(map[string]*grpc.ClientConn)
}

// Backend addresses for gRPC connections.
// These match the ADR-mandated internal gRPC ports.
const (
	GrpcAddrOntology     = "ontology-service:9001"
	GrpcAddrVersioning   = "versioning-service:9002"
	GrpcAddrAuth         = "auth-service:9003"
	GrpcAddrCommenting   = "commenting-service:9004"
	GrpcAddrPublisher    = "publisher-service:9005"
	GrpcAddrPublicBrowse = "public-browse-api:9011"
)

// NOTE: After running `make proto-generate`, create proto-specific clients:
//
//	import ontologyv1 "vedo-core/src/services/shared/proto/ontology/v1"
//
//	func (p *GrpcClientPool) OntologyClient(address string) (ontologyv1.OntologyServiceClient, error) {
//	    conn, err := p.GetConn(address)
//	    if err != nil {
//	        return nil, err
//	    }
//	    return ontologyv1.NewOntologyServiceClient(conn), nil
//	}
//
// See services/ontology_client.go for the full client implementations.
