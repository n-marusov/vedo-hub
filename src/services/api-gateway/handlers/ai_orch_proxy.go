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
