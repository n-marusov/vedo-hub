package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"vedo-core/src/services/ai-orchestration-service/internal/handler"
	"vedo-core/src/services/ai-orchestration-service/internal/middleware"
	"vedo-core/src/services/shared/llm"
	ai_orchestrationv1 "vedo-core/src/services/shared/proto/ai-orchestration/v1"
)

// AIOrchestrationService implements the AI Orchestration gRPC server.
// It delegates AI business logic to the handler package for each RPC.
type AIOrchestrationService struct {
	ai_orchestrationv1.UnimplementedAIOrchestrationServiceServer
	provider       llm.Provider
	templateEngine *llm.TemplateEngine
}

// NewAIOrchestrationService creates a new AIOrchestrationService.
// provider and templateEngine may be nil — the service starts without
// AI capabilities and returns LLM_UNAVAILABLE errors until configured.
func NewAIOrchestrationService(provider llm.Provider, templateEngine *llm.TemplateEngine) *AIOrchestrationService {
	return &AIOrchestrationService{
		provider:       provider,
		templateEngine: templateEngine,
	}
}

// isProviderReady checks whether the LLM provider and template engine are
// configured. Returns an error with a descriptive message if not.
func (s *AIOrchestrationService) isProviderReady() error {
	if s.provider == nil {
		return status.Error(codes.Unavailable, "LLM provider is not configured")
	}
	if s.templateEngine == nil {
		return status.Error(codes.Unavailable, "LLM template engine is not configured")
	}
	return nil
}

// =============================================================================
// GenerateOWL — NL→OWL conversion (unary)
// =============================================================================

func (s *AIOrchestrationService) GenerateOWL(ctx context.Context, req *ai_orchestrationv1.GenerateOWLRequest) (*ai_orchestrationv1.GenerateOWLResponse, error) {
	start := time.Now()
	slog.Debug("ai.generate_owl",
		"ontology_id", req.GetOntologyId(),
		"text_length", len(req.GetText()),
		"trace_id", req.GetTraceId(),
	)

	if err := s.isProviderReady(); err != nil {
		return nil, err
	}

	resp, err := handler.GenerateOWL(ctx, req, s.provider, s.templateEngine)
	if err != nil {
		slog.Error("ai.generate_owl.failed",
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return nil, err
	}

	slog.Info("ai.generate_owl.completed",
		"ontology_id", req.GetOntologyId(),
		"step_count", resp.GetTotalSteps(),
		"duration_ms", time.Since(start).Milliseconds(),
	)
	return resp, nil
}

// =============================================================================
// NaturalLanguageQuery — NL queries against ontology context (unary)
// =============================================================================

func (s *AIOrchestrationService) NaturalLanguageQuery(ctx context.Context, req *ai_orchestrationv1.NaturalLanguageQueryRequest) (*ai_orchestrationv1.NaturalLanguageQueryResponse, error) {
	start := time.Now()
	slog.Debug("ai.natural_language_query",
		"ontology_id", req.GetOntologyId(),
		"query_length", len(req.GetQuery()),
		"trace_id", req.GetTraceId(),
	)

	if err := s.isProviderReady(); err != nil {
		return nil, err
	}

	// Use GenerateOWL-style logic but rendered with a simpler prompt.
	llmCtx := llm.Context{
		BaseCtx:    ctx,
		TraceID:    req.GetTraceId(),
		MaxRetries: 2,
	}

	prompt := llm.Prompt{
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: "You are an ontology expert. Answer questions about ontologies based on the given context. Be concise and accurate."},
			{Role: llm.RoleUser, Content: req.GetQuery()},
		},
		Temperature: 0.2,
		MaxTokens:   1024,
	}
	if req.GetModel() != "" {
		prompt.Model = req.GetModel()
	}

	completion, err := s.provider.Complete(llmCtx, prompt)
	if err != nil {
		slog.Error("ai.natural_language_query.llm_error", "error", err)
		return nil, status.Error(codes.Internal, "LLM query failed")
	}

	slog.Info("ai.natural_language_query.completed",
		"duration_ms", time.Since(start).Milliseconds(),
		"tokens_used", completion.TokensUsed,
	)

	return &ai_orchestrationv1.NaturalLanguageQueryResponse{
		Answer:       completion.Text,
		TokensUsed:   int32(completion.TokensUsed),
		PromptTokens: int32(completion.PromptTokens),
	}, nil
}

// =============================================================================
// RefineOntology — Iterative refinement (server-streaming)
// =============================================================================

func (s *AIOrchestrationService) RefineOntology(req *ai_orchestrationv1.RefineOntologyRequest, stream ai_orchestrationv1.AIOrchestrationService_RefineOntologyServer) error {
	start := time.Now()
	slog.Debug("ai.refine_ontology",
		"ontology_id", req.GetOntologyId(),
		"session_id", req.GetSessionId(),
		"trace_id", req.GetTraceId(),
	)

	if err := s.isProviderReady(); err != nil {
		return err
	}

	resp, err := handler.RefineOntology(stream.Context(), req, s.provider, s.templateEngine)
	if err != nil {
		slog.Error("ai.refine_ontology.failed",
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return err
	}

	slog.Info("ai.refine_ontology.completed",
		"session_id", req.GetSessionId(),
		"change_count", len(resp.GetSteps()),
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return stream.Send(resp)
}

// =============================================================================
// Complete — AI-assisted completion (server-streaming)
// =============================================================================

func (s *AIOrchestrationService) Complete(req *ai_orchestrationv1.CompleteRequest, stream ai_orchestrationv1.AIOrchestrationService_CompleteServer) error {
	start := time.Now()
	slog.Debug("ai.complete",
		"type", req.GetType(),
		"ontology_id", req.GetOntologyId(),
		"trace_id", req.GetTraceId(),
	)

	if err := s.isProviderReady(); err != nil {
		return err
	}

	batch, err := handler.CompleteSuggest(stream.Context(), req, s.provider, s.templateEngine, nil)
	if err != nil {
		slog.Error("ai.complete.failed",
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		return err
	}

	// Stream the batch as a batched response (non-streaming fallback).
	// In production, individual tokens would be streamed for better UX.
	batchJSON, _ := json.Marshal(batch.GetSuggestions())

	err = stream.Send(&ai_orchestrationv1.CompleteResponse{
		Completion: &ai_orchestrationv1.CompleteResponse_Batch{
			Batch: batch,
		},
	})
	if err != nil {
		slog.Error("ai.complete.stream_send_error", "error", err)
		return err
	}

	slog.Info("ai.complete.completed",
		"ontology_id", req.GetOntologyId(),
		"suggestion_count", len(batch.GetSuggestions()),
		"duration_ms", time.Since(start).Milliseconds(),
		"token_hint", string(batchJSON[:min(len(batchJSON), 100)]),
	)

	return nil
}

// =============================================================================
// ListTemplates — List available ontology templates
// =============================================================================

func (s *AIOrchestrationService) ListTemplates(ctx context.Context, req *ai_orchestrationv1.ListTemplatesRequest) (*ai_orchestrationv1.ListTemplatesResponse, error) {
	slog.Debug("ai.list_templates",
		"domain_filter", req.GetDomainFilter(),
	)

	return handler.ListTemplates(ctx, req)
}

// =============================================================================
// GetTemplate — Get a single template by ID
// =============================================================================

func (s *AIOrchestrationService) GetTemplate(ctx context.Context, req *ai_orchestrationv1.GetTemplateRequest) (*ai_orchestrationv1.GetTemplateResponse, error) {
	slog.Debug("ai.get_template",
		"template_id", req.GetTemplateId(),
	)

	return handler.GetTemplate(ctx, req)
}

// =============================================================================
// CheckPolicy — LLM policy check (used by document-extractor)
// =============================================================================

func (s *AIOrchestrationService) CheckPolicy(ctx context.Context, req *ai_orchestrationv1.CheckPolicyRequest) (*ai_orchestrationv1.CheckPolicyResponse, error) {
	slog.Debug("ai.check_policy",
		"ontology_id", req.GetOntologyId(),
		"action", req.GetAction(),
		"user_id", req.GetUserId(),
		"trace_id", req.GetTraceId(),
	)

	deployMode := middleware.GetDeployMode()
	providerType := middleware.GetProviderType()
	visibility := req.GetOntologyId() // Simple default — in production, resolve from ontology service

	decision := middleware.EvaluatePolicy(deployMode, visibility, providerType, false)

	return &ai_orchestrationv1.CheckPolicyResponse{
		Allowed:        decision.Allowed,
		Provider:       providerType,
		Model:          providerType + "-default",
		Reason:         decision.Reason,
		RequireConsent: false,
	}, nil
}

// =============================================================================
// LogLLMUsage — Centralized LLM usage audit
// =============================================================================

func (s *AIOrchestrationService) LogLLMUsage(ctx context.Context, req *ai_orchestrationv1.LogLLMUsageRequest) (*ai_orchestrationv1.LogLLMUsageResponse, error) {
	slog.Info("ai.log_llm_usage",
		"ontology_id", req.GetOntologyId(),
		"action", req.GetAction(),
		"provider", req.GetProvider(),
		"model", req.GetModel(),
		"tokens_in", req.GetTokensIn(),
		"tokens_out", req.GetTokensOut(),
		"cost", req.GetCost(),
		"duration_ms", req.GetDurationMs(),
		"user_id", req.GetUserId(),
		"trace_id", req.GetTraceId(),
	)

	return &ai_orchestrationv1.LogLLMUsageResponse{
		Recorded: true,
	}, nil
}
