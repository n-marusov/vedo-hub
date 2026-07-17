package proxy

import (
	"context"
	"log/slog"
	"time"

	ai_orchestrationv1 "vedo-core/src/services/shared/proto/ai-orchestration/v1"
)

// GrpcAddrAIOrch is the gRPC address for the ai-orchestration-service.
const GrpcAddrAIOrch = "ai-orchestration-service:9014"

// AIOrchestrationServiceClient wraps the generated proto stub with lazy
// initialization, logging, and context deadlines from the gRPC connection pool.
type AIOrchestrationServiceClient struct {
	pool        *GrpcClientPool
	address     string
	stub        ai_orchestrationv1.AIOrchestrationServiceClient
	lastRefresh time.Time
}

// NewAIOrchestrationServiceClient creates a new AIOrchestrationServiceClient wrapper.
func NewAIOrchestrationServiceClient(pool *GrpcClientPool, address string) *AIOrchestrationServiceClient {
	return &AIOrchestrationServiceClient{
		pool:    pool,
		address: address,
	}
}

// getStub lazily initializes the gRPC client stub.
func (c *AIOrchestrationServiceClient) getStub() (ai_orchestrationv1.AIOrchestrationServiceClient, error) {
	if c.stub != nil && time.Since(c.lastRefresh) < 5*time.Minute {
		return c.stub, nil
	}
	conn, err := c.pool.GetConn(c.address)
	if err != nil {
		slog.Error("grpc.ai_orch.stub_init_failed", "address", c.address, "error", err)
		return nil, err
	}
	c.stub = ai_orchestrationv1.NewAIOrchestrationServiceClient(conn)
	c.lastRefresh = time.Now()
	slog.Debug("grpc.ai_orch.stub_initialized", "address", c.address)
	return c.stub, nil
}

// GenerateOWL sends an NL→OWL generation request to the ai-orchestration-service.
func (c *AIOrchestrationServiceClient) GenerateOWL(ctx context.Context, req *ai_orchestrationv1.GenerateOWLRequest) (*ai_orchestrationv1.GenerateOWLResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	gctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.ai_orch.generate_owl",
		"ontology_id", req.GetOntologyId(),
		"text_length", len(req.GetText()),
	)
	resp, err := stub.GenerateOWL(gctx, req)
	if err != nil {
		slog.Error("grpc.ai_orch.generate_owl_failed",
			"ontology_id", req.GetOntologyId(),
			"error", err,
		)
		return nil, err
	}
	return resp, nil
}

// Complete sends an AI completion request to the ai-orchestration-service.
func (c *AIOrchestrationServiceClient) Complete(ctx context.Context, req *ai_orchestrationv1.CompleteRequest) (*ai_orchestrationv1.CompleteResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	gctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.ai_orch.complete",
		"ontology_id", req.GetOntologyId(),
		"type", req.GetType(),
	)
	stream, err := stub.Complete(gctx, req)
	if err != nil {
		slog.Error("grpc.ai_orch.complete_failed",
			"ontology_id", req.GetOntologyId(),
			"error", err,
		)
		return nil, err
	}
	// Collect all streaming responses and return the last batch
	var lastResp *ai_orchestrationv1.CompleteResponse
	for {
		resp, err := stream.Recv()
		if err != nil {
			break
		}
		lastResp = resp
	}
	return lastResp, nil
}

// RefineOntology sends a refinement request to the ai-orchestration-service.
func (c *AIOrchestrationServiceClient) RefineOntology(ctx context.Context, req *ai_orchestrationv1.RefineOntologyRequest) (*ai_orchestrationv1.RefineOntologyResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	gctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.ai_orch.refine",
		"ontology_id", req.GetOntologyId(),
		"session_id", req.GetSessionId(),
	)
	stream, err := stub.RefineOntology(gctx, req)
	if err != nil {
		slog.Error("grpc.ai_orch.refine_failed",
			"ontology_id", req.GetOntologyId(),
			"error", err,
		)
		return nil, err
	}
	// Collect all streaming responses
	var lastResp *ai_orchestrationv1.RefineOntologyResponse
	for {
		resp, err := stream.Recv()
		if err != nil {
			break
		}
		lastResp = resp
	}
	return lastResp, nil
}

// ListTemplates sends a template listing request to the ai-orchestration-service.
func (c *AIOrchestrationServiceClient) ListTemplates(ctx context.Context, req *ai_orchestrationv1.ListTemplatesRequest) (*ai_orchestrationv1.ListTemplatesResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	gctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.ai_orch.list_templates", "domain_filter", req.GetDomainFilter())
	resp, err := stub.ListTemplates(gctx, req)
	if err != nil {
		slog.Error("grpc.ai_orch.list_templates_failed", "error", err)
		return nil, err
	}
	return resp, nil
}

// GetTemplate sends a single template retrieval request.
func (c *AIOrchestrationServiceClient) GetTemplate(ctx context.Context, req *ai_orchestrationv1.GetTemplateRequest) (*ai_orchestrationv1.GetTemplateResponse, error) {
	stub, err := c.getStub()
	if err != nil {
		return nil, err
	}
	gctx, cancel := context.WithTimeout(ctx, c.pool.GetUpstreamTimeout())
	defer cancel()
	slog.Debug("grpc.ai_orch.get_template", "template_id", req.GetTemplateId())
	resp, err := stub.GetTemplate(gctx, req)
	if err != nil {
		slog.Error("grpc.ai_orch.get_template_failed",
			"template_id", req.GetTemplateId(),
			"error", err,
		)
		return nil, err
	}
	return resp, nil
}
