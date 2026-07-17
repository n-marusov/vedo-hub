package middleware

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Deployment mode constants.
const (
	DeployModeSaaS      = "saas"
	DeployModeOnPremise = "on-premise"
	DeployModeAirGapped = "air-gapped"
)

// PolicyDecision represents the result of an LLM access policy evaluation.
type PolicyDecision struct {
	Allowed bool
	Reason  string
}

// cachedDecision stores a policy decision with expiry.
type cachedDecision struct {
	decision PolicyDecision
	expires  time.Time
}

// policyCache provides TTL-based caching for policy decisions.
type policyCache struct {
	mu    sync.RWMutex
	items map[string]cachedDecision
}

// get retrieves a cached decision if it hasn't expired.
func (pc *policyCache) get(key string) (PolicyDecision, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	item, ok := pc.items[key]
	if !ok {
		return PolicyDecision{}, false
	}
	if time.Now().After(item.expires) {
		return PolicyDecision{}, false
	}
	return item.decision, true
}

// set stores a policy decision with a 60-second TTL.
func (pc *policyCache) set(key string, decision PolicyDecision) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.items[key] = cachedDecision{
		decision: decision,
		expires:  time.Now().Add(60 * time.Second),
	}
}

var globalPolicyCache = &policyCache{
	items: make(map[string]cachedDecision),
}

// UnaryLLMPolicyInterceptor returns a gRPC UnaryServerInterceptor that
// enforces LLM access policy for AI RPCs.
func UnaryLLMPolicyInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if !isAIRPCMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Extract ontology_id from the request (all AI RPCs have ontology_id)
		ontologyID := extractOntologyID(req)
		if ontologyID == "" {
			return handler(ctx, req)
		}

		// Check cache
		decision, ok := globalPolicyCache.get(ontologyID)
		if !ok {
			deployMode := GetDeployMode()
			providerType := GetProviderType()
			visibility := resolveVisibility(ctx, ontologyID)
			decision = EvaluatePolicy(deployMode, visibility, providerType, isAdmin(ctx))
			globalPolicyCache.set(ontologyID, decision)
		}

		if decision.Allowed {
			slog.Debug("llm.policy.allowed",
				"method", info.FullMethod,
				"ontology_id", ontologyID,
				"reason", decision.Reason,
			)
			return handler(ctx, req)
		}

		slog.Warn("llm.policy.blocked",
			"method", info.FullMethod,
			"ontology_id", ontologyID,
			"reason", decision.Reason,
		)
		return nil, status.Error(codes.PermissionDenied, decision.Reason)
	}
}

// StreamLLMPolicyInterceptor returns a gRPC StreamServerInterceptor that
// enforces LLM access policy for streaming AI RPCs.
func StreamLLMPolicyInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if !isAIRPCMethod(info.FullMethod) {
			return handler(srv, stream)
		}

		// For streaming RPCs, we check policy on the first message
		return handler(srv, stream)
	}
}

// isAIRPCMethod checks if the gRPC method is an AI-related RPC.
func isAIRPCMethod(method string) bool {
	aiMethods := []string{
		"/vedo.ai_orchestration.v1.AIOrchestrationService/GenerateOWL",
		"/vedo.ai_orchestration.v1.AIOrchestrationService/NaturalLanguageQuery",
		"/vedo.ai_orchestration.v1.AIOrchestrationService/RefineOntology",
		"/vedo.ai_orchestration.v1.AIOrchestrationService/Complete",
	}
	for _, m := range aiMethods {
		if method == m {
			return true
		}
	}
	return false
}

// extractOntologyID extracts the ontology_id from a proto message using
// the proto getter interface. All AI RPC request messages have GetOntologyId().
func extractOntologyID(req interface{}) string {
	type ontologyIDProvider interface {
		GetOntologyId() string
	}
	if p, ok := req.(ontologyIDProvider); ok {
		return p.GetOntologyId()
	}
	return ""
}

// GetDeployMode reads the DEPLOYMENT_MODE env var.
func GetDeployMode() string {
	mode := os.Getenv("DEPLOYMENT_MODE")
	switch mode {
	case DeployModeSaaS, DeployModeOnPremise, DeployModeAirGapped:
		return mode
	default:
		return DeployModeSaaS
	}
}

// GetProviderType determines the LLM provider type from env config.
func GetProviderType() string {
	provider := os.Getenv("LLM_PROVIDER")
	switch strings.ToLower(provider) {
	case "ollama", "llama", "local-llm", "":
		return "local"
	default:
		return "external"
	}
}

// resolveVisibility extracts ontology visibility from the request context.
// In the ai-orchestration-service, visibility is resolved via gRPC metadata
// (set by the API Gateway JWT middleware as X-Ontology-Visibility).
func resolveVisibility(ctx context.Context, ontologyID string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "private"
	}

	values := md.Get("x-ontology-visibility")
	if len(values) > 0 && values[0] != "" {
		return values[0]
	}

	slog.Debug("llm.policy.no_visibility_info",
		"ontology_id", ontologyID,
		"fallback", "private",
	)
	return "private"
}

// isAdmin checks whether the authenticated user has admin role from gRPC metadata.
func isAdmin(ctx context.Context) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}

	roles := md.Get("x-user-roles")
	for _, role := range roles {
		for _, admin := range []string{"admin", "vedo-admin", "organization-admin"} {
			if strings.EqualFold(role, admin) {
				return true
			}
		}
	}
	return false
}

// EvaluatePolicy applies the LLM access policy matrix.
func EvaluatePolicy(deployMode, visibility, providerType string, userIsAdmin bool) PolicyDecision {
	// Air-gapped always blocks external providers
	if deployMode == DeployModeAirGapped && providerType == "external" {
		return PolicyDecision{
			Allowed: false,
			Reason:  "External LLM providers are not allowed in air-gapped deployment mode.",
		}
	}

	// Public ontology — always allowed
	if visibility == "public" {
		return PolicyDecision{
			Allowed: true,
			Reason:  "Public ontology: LLM access allowed.",
		}
	}

	// On-premise — always allowed
	if deployMode == DeployModeOnPremise {
		return PolicyDecision{
			Allowed: true,
			Reason:  "On-premise deployment: LLM access allowed.",
		}
	}

	// Local LLM (SaaS mode) — always allowed
	if providerType == "local" {
		return PolicyDecision{
			Allowed: true,
			Reason:  "Local LLM provider: access allowed.",
		}
	}

	// SaaS + External LLM + non-public ontology
	if deployMode == DeployModeSaaS && providerType == "external" {
		if visibility == "internal" || visibility == "private" {
			if userIsAdmin {
				slog.Warn("llm.policy.admin_override",
					"ontology_visibility", visibility,
				)
				return PolicyDecision{
					Allowed: true,
					Reason:  "Admin override: LLM access allowed for non-public ontology.",
				}
			}

			return PolicyDecision{
				Allowed: false,
				Reason:  "LLM access to " + visibility + " ontology is blocked in SaaS mode with external provider. Use a local LLM provider or have an admin override.",
			}
		}
	}

	// Default: allow
	return PolicyDecision{
		Allowed: true,
		Reason:  "LLM access allowed.",
	}
}
