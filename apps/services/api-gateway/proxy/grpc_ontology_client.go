package proxy

import (
	"context"
	"log/slog"
	"time"

	ontologyv1 "vedo-core/src/services/shared/proto/ontology/v1"
)

// GetOntologyVisibility resolves ontology visibility from the server,
// implementing middleware.OntologyVisibilityResolver.
// Returns the visibility as a string: "public", "internal", "private", or "".
func (c *OntologyServiceClient) GetOntologyVisibility(ctx context.Context, ontologyID string) (string, error) {
	stub, err := c.getStub()
	if err != nil {
		return "", err
	}
	gctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()

	resp, err := stub.GetOntology(gctx, &ontologyv1.GetOntologyRequest{
		OntologyId: ontologyID,
	})
	if err != nil {
		slog.Error("grpc.ontology.get_ontology_visibility_failed",
			"ontology_id", ontologyID, "error", err,
		)
		return "", err
	}
	if resp.GetOntology() == nil {
		return "", nil
	}

	switch resp.GetOntology().GetVisibility() {
	case ontologyv1.Ontology_VISIBILITY_PUBLIC:
		return "public", nil
	case ontologyv1.Ontology_VISIBILITY_INTERNAL:
		return "internal", nil
	case ontologyv1.Ontology_VISIBILITY_PRIVATE:
		return "private", nil
	default:
		return "private", nil
	}
}

// OntologyServiceClient wraps the generated proto stub with lazy initialization,
// logging, and context deadlines from the gRPC connection pool.
type OntologyServiceClient struct {
	pool        *GrpcClientPool
	address     string
	stub        ontologyv1.OntologyServiceClient
	lastRefresh time.Time
}

// NewOntologyServiceClient creates a new OntologyServiceClient wrapper.
func NewOntologyServiceClient(pool *GrpcClientPool, address string) *OntologyServiceClient {
	return &OntologyServiceClient{
		pool:    pool,
		address: address,
	}
}

// getStub lazily initializes the gRPC client stub.
func (c *OntologyServiceClient) getStub() (ontologyv1.OntologyServiceClient, error) {
	if c.stub != nil && time.Since(c.lastRefresh) < 5*time.Minute {
		return c.stub, nil
	}
	conn, err := c.pool.GetConn(c.address)
	if err != nil {
		slog.Error("grpc.ontology.stub_init_failed", "address", c.address, "error", err)
		return nil, err
	}
	c.stub = ontologyv1.NewOntologyServiceClient(conn)
	c.lastRefresh = time.Now()
	slog.Debug("grpc.ontology.stub_initialized", "address", c.address)
	return c.stub, nil
}

// CreateClass sends a CreateClassRequest to the ontology service.
func (c *OntologyServiceClient) CreateClass(ctx context.Context, req *ontologyv1.CreateClassRequest) (*ontologyv1.CreateClassResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.ontology.create_class", "ontology_id", req.GetOntologyId(), "label", req.GetLabel())
	resp, err := stub.CreateClass(ctx, req)
	if err != nil {
		slog.Error("grpc.ontology.create_class_failed", "ontology_id", req.GetOntologyId(), "error", err)
		return nil, err
	}
	return resp, nil
}

// GetClass sends a GetClassRequest to the ontology service.
func (c *OntologyServiceClient) GetClass(ctx context.Context, req *ontologyv1.GetClassRequest) (*ontologyv1.GetClassResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.ontology.get_class", "ontology_id", req.GetOntologyId(), "class_id", req.GetClassId())
	resp, err := stub.GetClass(ctx, req)
	if err != nil {
		slog.Error("grpc.ontology.get_class_failed", "ontology_id", req.GetOntologyId(), "class_id", req.GetClassId(), "error", err)
		return nil, err
	}
	return resp, nil
}

// ListClasses sends a ListClassesRequest to the ontology service.
func (c *OntologyServiceClient) ListClasses(ctx context.Context, req *ontologyv1.ListClassesRequest) (*ontologyv1.ListClassesResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.ontology.list_classes", "ontology_id", req.GetOntologyId(), "search", req.GetSearch())
	resp, err := stub.ListClasses(ctx, req)
	if err != nil {
		slog.Error("grpc.ontology.list_classes_failed", "ontology_id", req.GetOntologyId(), "error", err)
		return nil, err
	}
	return resp, nil
}

// ApplySequence sends an ApplySequenceRequest to the ontology service.
func (c *OntologyServiceClient) ApplySequence(ctx context.Context, req *ontologyv1.ApplySequenceRequest) (*ontologyv1.ApplySequenceResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.ontology.apply_sequence", "ontology_id", req.GetOntologyId(), "step_count", len(req.GetSteps()))
	resp, err := stub.ApplySequence(ctx, req)
	if err != nil {
		slog.Error("grpc.ontology.apply_sequence_failed", "ontology_id", req.GetOntologyId(), "error", err)
		return nil, err
	}
	return resp, nil
}

// GetOntology sends a GetOntologyRequest to the ontology service.
func (c *OntologyServiceClient) GetOntology(ctx context.Context, req *ontologyv1.GetOntologyRequest) (*ontologyv1.GetOntologyResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.ontology.get_ontology", "ontology_id", req.GetOntologyId())
	resp, err := stub.GetOntology(ctx, req)
	if err != nil {
		slog.Error("grpc.ontology.get_ontology_failed", "ontology_id", req.GetOntologyId(), "error", err)
		return nil, err
	}
	return resp, nil
}

// ExecuteSPARQL sends an ExecuteSPARQLRequest to the ontology service.
func (c *OntologyServiceClient) ExecuteSPARQL(ctx context.Context, req *ontologyv1.ExecuteSPARQLRequest) (*ontologyv1.ExecuteSPARQLResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.ontology.execute_sparql", "ontology_id", req.GetOntologyId(), "max_results", req.GetMaxResults())
	resp, err := stub.ExecuteSPARQL(ctx, req)
	if err != nil {
		slog.Error("grpc.ontology.execute_sparql_failed", "ontology_id", req.GetOntologyId(), "error", err)
		return nil, err
	}
	return resp, nil
}

// ExecuteCYPHER sends an ExecuteCYPHERRequest to the ontology service.
func (c *OntologyServiceClient) ExecuteCYPHER(ctx context.Context, req *ontologyv1.ExecuteCYPHERRequest) (*ontologyv1.ExecuteCYPHERResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.ontology.execute_cypher", "ontology_id", req.GetOntologyId(), "max_results", req.GetMaxResults())
	resp, err := stub.ExecuteCYPHER(ctx, req)
	if err != nil {
		slog.Error("grpc.ontology.execute_cypher_failed", "ontology_id", req.GetOntologyId(), "error", err)
		return nil, err
	}
	return resp, nil
}
