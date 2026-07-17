package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"vedo-core/src/services/api-gateway/proxy"
	ai_orchestrationv1 "vedo-core/src/services/shared/proto/ai-orchestration/v1"
	ontologyv1 "vedo-core/src/services/shared/proto/ontology/v1"
)

// AIOrchProxy provides thin HTTP↔gRPC translation for AI routes.
// It contains NO AI business logic — all AI logic lives in
// ai-orchestration-service. This is the post-migration proxy layer.
type AIOrchProxy struct {
	aiOrch       *proxy.AIOrchestrationServiceClient
	ontologyGrpc *proxy.OntologyServiceClient
}

// NewAIOrchProxy creates a new AIOrchProxy.
func NewAIOrchProxy(aiOrch *proxy.AIOrchestrationServiceClient, ontologyGrpc *proxy.OntologyServiceClient) *AIOrchProxy {
	return &AIOrchProxy{
		aiOrch:       aiOrch,
		ontologyGrpc: ontologyGrpc,
	}
}

// HandleGenerateFromText proxies POST /ontologies/:id/generate-from-text
// to the ai-orchestration-service's GenerateOWL RPC.
func (p *AIOrchProxy) HandleGenerateFromText(c *gin.Context) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")
	ontologyID := c.Param("id")

	if ontologyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "GATEWAY-MISSING-ONTOLOGY-ID", "message": "Ontology ID is required."},
		})
		return
	}

	var reqBody struct {
		Text  string `json:"text"`
		Model string `json:"model,omitempty"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "GATEWAY-INVALID-REQUEST", "message": "Invalid request body. 'text' field is required."},
		})
		return
	}

	resp, err := p.aiOrch.GenerateOWL(c.Request.Context(), &ai_orchestrationv1.GenerateOWLRequest{
		OntologyId: ontologyID,
		Text:       reqBody.Text,
		Model:      reqBody.Model,
		TraceId:    traceID,
	})
	if err != nil {
		aiOrchErrToHTTP(c, err, "generate-from-text")
		return
	}

	// Convert proto response to JSON response matching the original API format
	c.JSON(http.StatusOK, gin.H{
		"ontology_id": ontologyID,
		"steps":       resp.GetSteps(),
		"total_steps": resp.GetTotalSteps(),
		"warnings":    resp.GetWarnings(),
	})

	slog.Info("proxy.generate_from_text.completed",
		"ontology_id", ontologyID,
		"step_count", resp.GetTotalSteps(),
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// HandleSuggestClasses proxies to the Complete RPC with COMPLETION_TYPE_SUGGEST_CLASSES.
func (p *AIOrchProxy) HandleSuggestClasses(c *gin.Context) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")
	ontologyID := c.Param("id")

	if ontologyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "GATEWAY-MISSING-ONTOLOGY-ID", "message": "Ontology ID is required."},
		})
		return
	}

	var reqBody struct {
		ClassID     string `json:"class_id"`
		Model       string `json:"model,omitempty"`
		ContextSize int    `json:"context_size,omitempty"`
	}
	_ = c.ShouldBindJSON(&reqBody)

	resp, err := p.aiOrch.Complete(c.Request.Context(), &ai_orchestrationv1.CompleteRequest{
		Type:        ai_orchestrationv1.CompleteRequest_COMPLETION_TYPE_SUGGEST_CLASSES,
		OntologyId:  ontologyID,
		PartialText: reqBody.ClassID,
		Model:       reqBody.Model,
		TraceId:     traceID,
	})
	if err != nil {
		aiOrchErrToHTTP(c, err, "suggest-classes")
		return
	}

	// Parse the batch suggestions from the response
	suggestions := parseSuggestionBatch(resp)

	c.JSON(http.StatusOK, gin.H{
		"suggestions":       suggestions,
		"total_suggestions": len(suggestions),
		"ontology_id":       ontologyID,
	})

	slog.Info("proxy.suggest_classes.completed",
		"ontology_id", ontologyID,
		"count", len(suggestions),
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// HandleSuggestProperties proxies to the Complete RPC with COMPLETION_TYPE_SUGGEST_PROPERTIES.
func (p *AIOrchProxy) HandleSuggestProperties(c *gin.Context) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")
	ontologyID := c.Param("id")

	if ontologyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "GATEWAY-MISSING-ONTOLOGY-ID", "message": "Ontology ID is required."},
		})
		return
	}

	var reqBody struct {
		ClassID string `json:"class_id"`
		Model   string `json:"model,omitempty"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil || reqBody.ClassID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "GATEWAY-INVALID-REQUEST", "message": "Invalid request body. 'class_id' field is required."},
		})
		return
	}

	resp, err := p.aiOrch.Complete(c.Request.Context(), &ai_orchestrationv1.CompleteRequest{
		Type:        ai_orchestrationv1.CompleteRequest_COMPLETION_TYPE_SUGGEST_PROPERTIES,
		OntologyId:  ontologyID,
		PartialText: reqBody.ClassID,
		Model:       reqBody.Model,
		TraceId:     traceID,
	})
	if err != nil {
		aiOrchErrToHTTP(c, err, "suggest-properties")
		return
	}

	suggestions := parseSuggestionBatch(resp)

	c.JSON(http.StatusOK, gin.H{
		"suggestions":       suggestions,
		"total_suggestions": len(suggestions),
		"ontology_id":       ontologyID,
	})

	slog.Info("proxy.suggest_properties.completed",
		"ontology_id", ontologyID,
		"count", len(suggestions),
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// HandleSuggestRelationships proxies to the Complete RPC with COMPLETION_TYPE_SUGGEST_RELATIONSHIPS.
func (p *AIOrchProxy) HandleSuggestRelationships(c *gin.Context) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")
	ontologyID := c.Param("id")

	if ontologyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "GATEWAY-MISSING-ONTOLOGY-ID", "message": "Ontology ID is required."},
		})
		return
	}

	var reqBody struct {
		SourceClassID string `json:"source_class_id"`
		TargetClassID string `json:"target_class_id,omitempty"`
		Model         string `json:"model,omitempty"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil || reqBody.SourceClassID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "GATEWAY-INVALID-REQUEST", "message": "Invalid request body. 'source_class_id' field is required."},
		})
		return
	}

	resp, err := p.aiOrch.Complete(c.Request.Context(), &ai_orchestrationv1.CompleteRequest{
		Type:        ai_orchestrationv1.CompleteRequest_COMPLETION_TYPE_SUGGEST_RELATIONSHIPS,
		OntologyId:  ontologyID,
		PartialText: reqBody.SourceClassID,
		ContextJson: reqBody.TargetClassID,
		Model:       reqBody.Model,
		TraceId:     traceID,
	})
	if err != nil {
		aiOrchErrToHTTP(c, err, "suggest-relationships")
		return
	}

	suggestions := parseSuggestionBatch(resp)

	c.JSON(http.StatusOK, gin.H{
		"suggestions":       suggestions,
		"total_suggestions": len(suggestions),
		"ontology_id":       ontologyID,
	})

	slog.Info("proxy.suggest_relationships.completed",
		"ontology_id", ontologyID,
		"count", len(suggestions),
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// HandleRefine proxies POST /ontologies/:id/ai/refine to the RefineOntology RPC.
func (p *AIOrchProxy) HandleRefine(c *gin.Context) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")
	ontologyID := c.Param("id")

	if ontologyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "GATEWAY-MISSING-ONTOLOGY-ID", "message": "Ontology ID is required."},
		})
		return
	}

	var reqBody struct {
		SequenceID   string `json:"sequence_id"`
		FeedbackText string `json:"feedback_text"`
		Model        string `json:"model,omitempty"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil || reqBody.SequenceID == "" || reqBody.FeedbackText == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "GATEWAY-INVALID-REQUEST", "message": "'sequence_id' and 'feedback_text' are required."},
		})
		return
	}

	resp, err := p.aiOrch.RefineOntology(c.Request.Context(), &ai_orchestrationv1.RefineOntologyRequest{
		OntologyId: ontologyID,
		SessionId:  reqBody.SequenceID,
		Feedback:   reqBody.FeedbackText,
		Model:      reqBody.Model,
		TraceId:    traceID,
	})
	if err != nil {
		aiOrchErrToHTTP(c, err, "refine")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sequence_id": resp.GetSessionId(),
		"changes":     resp.GetSteps(),
		"summary":     resp.GetMessage(),
		"ontology_id": ontologyID,
	})

	slog.Info("proxy.refine.completed",
		"ontology_id", ontologyID,
		"session_id", resp.GetSessionId(),
		"change_count", len(resp.GetSteps()),
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// HandleListTemplates proxies GET /templates/ontologies to the ListTemplates RPC.
func (p *AIOrchProxy) HandleListTemplates(c *gin.Context) {
	start := time.Now()
	domainFilter := c.Query("domain")

	resp, err := p.aiOrch.ListTemplates(c.Request.Context(), &ai_orchestrationv1.ListTemplatesRequest{
		DomainFilter: domainFilter,
	})
	if err != nil {
		aiOrchErrToHTTP(c, err, "list-templates")
		return
	}

	c.JSON(http.StatusOK, resp.GetTemplates())

	slog.Info("proxy.list_templates.completed",
		"count", len(resp.GetTemplates()),
		"domain_filter", domainFilter,
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// HandleApplyTemplate applies a template by fetching it from the
// ai-orchestration-service and calling ApplySequence on the ontology service.
func (p *AIOrchProxy) HandleApplyTemplate(c *gin.Context) {
	start := time.Now()
	traceID := c.GetHeader("X-Trace-Id")
	ontologyID := c.Param("id")

	if ontologyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "GATEWAY-MISSING-ONTOLOGY-ID", "message": "Ontology ID is required."},
		})
		return
	}

	var reqBody struct {
		TemplateID string `json:"template_id"`
		BranchID   string `json:"branch_id,omitempty"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil || reqBody.TemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "GATEWAY-INVALID-REQUEST", "message": "'template_id' is required."},
		})
		return
	}

	// Fetch template from ai-orchestration-service
	tplResp, err := p.aiOrch.GetTemplate(c.Request.Context(), &ai_orchestrationv1.GetTemplateRequest{
		TemplateId: reqBody.TemplateID,
	})
	if err != nil {
		aiOrchErrToHTTP(c, err, "get-template")
		return
	}

	tpl := tplResp.GetTemplate()
	if tpl == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "GATEWAY-TEMPLATE-NOT-FOUND", "message": fmt.Sprintf("Template %q not found.", reqBody.TemplateID)},
		})
		return
	}

	// Convert template steps to ontology service proto format
	protoSteps := convertTemplateStepsToOntology(tpl.GetSteps())

	// Apply via ontology service gRPC
	applyResp, err := p.ontologyGrpc.ApplySequence(c.Request.Context(), &ontologyv1.ApplySequenceRequest{
		OntologyId:    ontologyID,
		BranchId:      reqBody.BranchID,
		Steps:         protoSteps,
		CommitMessage: fmt.Sprintf("Apply template: %s (%s)", tpl.GetName(), tpl.GetDomain()),
	})
	if err != nil {
		slog.Error("proxy.apply_template.grpc_error",
			"trace_id", traceID,
			"ontology_id", ontologyID,
			"template_id", reqBody.TemplateID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "GATEWAY-GRPC-ERROR", "message": fmt.Sprintf("Failed to apply template: %s", err.Error())},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"template_id":   reqBody.TemplateID,
		"ontology_id":   ontologyID,
		"commit_id":     applyResp.GetCommitId(),
		"steps_applied": applyResp.GetStepsApplied(),
	})

	slog.Info("proxy.apply_template.completed",
		"ontology_id", ontologyID,
		"template_id", reqBody.TemplateID,
		"commit_id", applyResp.GetCommitId(),
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// parseSuggestionBatch extracts suggestion items from the gRPC CompleteResponse.
func parseSuggestionBatch(resp *ai_orchestrationv1.CompleteResponse) []interface{} {
	if resp == nil {
		return []interface{}{}
	}
	if batch := resp.GetBatch(); batch != nil {
		suggestions := make([]interface{}, len(batch.GetSuggestions()))
		for i, s := range batch.GetSuggestions() {
			// Try to parse as JSON object, fall back to raw string
			var obj interface{}
			if err := json.Unmarshal([]byte(s), &obj); err == nil {
				suggestions[i] = obj
			} else {
				suggestions[i] = s
			}
		}
		return suggestions
	}
	return []interface{}{}
}

// convertTemplateStepsToOntology converts ai-orchestration proto SequenceSteps
// to ontology service proto SequenceStep messages.
func convertTemplateStepsToOntology(steps []*ai_orchestrationv1.SequenceStep) []*ontologyv1.SequenceStep {
	protoSteps := make([]*ontologyv1.SequenceStep, 0, len(steps))
	for _, s := range steps {
		op := protoOpFromAI(s.GetOperation())
		protoSteps = append(protoSteps, &ontologyv1.SequenceStep{
			Operation:   op,
			EntityId:    s.GetEntityId(),
			Label:       s.GetLabel(),
			ParentId:    s.GetParentId(),
			DomainId:    s.GetDomainId(),
			RangeId:     s.GetRangeId(),
			Annotations: s.GetAnnotations(),
		})
	}
	return protoSteps
}

// protoOpFromAI converts ai-orchestration proto operation enum to ontology proto enum.
func protoOpFromAI(op ai_orchestrationv1.SequenceStep_Operation) ontologyv1.SequenceStep_Operation {
	switch op {
	case ai_orchestrationv1.SequenceStep_OPERATION_CREATE_CLASS:
		return ontologyv1.SequenceStep_OPERATION_CREATE_CLASS
	case ai_orchestrationv1.SequenceStep_OPERATION_CREATE_OBJECT_PROPERTY:
		return ontologyv1.SequenceStep_OPERATION_CREATE_OBJECT_PROPERTY
	case ai_orchestrationv1.SequenceStep_OPERATION_CREATE_DATATYPE_PROPERTY:
		return ontologyv1.SequenceStep_OPERATION_CREATE_DATATYPE_PROPERTY
	case ai_orchestrationv1.SequenceStep_OPERATION_CREATE_INDIVIDUAL:
		return ontologyv1.SequenceStep_OPERATION_CREATE_INDIVIDUAL
	case ai_orchestrationv1.SequenceStep_OPERATION_ADD_ANNOTATION:
		return ontologyv1.SequenceStep_OPERATION_ADD_ANNOTATION
	case ai_orchestrationv1.SequenceStep_OPERATION_SET_PARENT:
		return ontologyv1.SequenceStep_OPERATION_SET_PARENT
	case ai_orchestrationv1.SequenceStep_OPERATION_SET_DOMAIN:
		return ontologyv1.SequenceStep_OPERATION_SET_DOMAIN
	case ai_orchestrationv1.SequenceStep_OPERATION_SET_RANGE:
		return ontologyv1.SequenceStep_OPERATION_SET_RANGE
	default:
		return ontologyv1.SequenceStep_OPERATION_UNSPECIFIED
	}
}

// aiOrchErrToHTTP converts ai-orchestration gRPC errors to HTTP JSON responses.
func aiOrchErrToHTTP(c *gin.Context, err error, action string) {
	traceID := c.GetHeader("X-Trace-Id")
	st, ok := status.FromError(err)
	if !ok {
		slog.Error("proxy.ai_orch.non_grpc_error",
			"action", action,
			"error", err,
			"trace_id", traceID,
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "GATEWAY-INTERNAL", "message": "Internal server error"},
		})
		return
	}

	var httpCode int
	var errCode string

	switch st.Code() {
	case codes.InvalidArgument:
		httpCode = http.StatusBadRequest
		errCode = "GATEWAY-INVALID-REQUEST"
	case codes.NotFound:
		httpCode = http.StatusNotFound
		errCode = "GATEWAY-NOT-FOUND"
	case codes.PermissionDenied:
		httpCode = http.StatusForbidden
		errCode = "LLM-POLICY-BLOCKED"
	case codes.Unavailable:
		httpCode = http.StatusServiceUnavailable
		errCode = "GATEWAY-LLM-UNAVAILABLE"
	default:
		httpCode = http.StatusInternalServerError
		errCode = "GATEWAY-INTERNAL"
	}

	msg := st.Message()
	// Put human-readable fallback on generic error codes
	if errCode == "GATEWAY-INTERNAL" {
		msg = fmt.Sprintf("AI service error: %s", msg)
	}
	if errCode == "GATEWAY-LLM-UNAVAILABLE" {
		msg = "AI service is not available. Set LLM_PROVIDER, LLM_API_KEY, and LLM_MODEL environment variables."
	}

	slog.Warn("proxy.ai_orch.error",
		"action", action,
		"http_code", httpCode,
		"err_code", errCode,
		"grpc_code", st.Code(),
		"message", st.Message(),
		"trace_id", traceID,
	)

	c.JSON(httpCode, gin.H{
		"error": gin.H{"code": errCode, "message": msg},
	})
}

// unused import guard
var _ = strings.TrimSpace
