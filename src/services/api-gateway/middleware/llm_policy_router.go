package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
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

// LLMPolicyRouter returns a Gin middleware that controls LLM access based on
// ontology visibility and deployment type, per ADR-DES.API.llm-policy-router-strategy.
//
// Policy matrix:
//
//	| Ontology Visibility | SaaS + External LLM | SaaS + Local LLM | On-premise (any LLM) |
//	|--------------------|-------------------|-----------------|--------------------|
//	| Public             | Allow             | Allow           | Allow              |
//	| Internal           | Block/AdminOverride| Allow           | Allow              |
//	| Private            | Block/AdminOverride| Allow           | Allow              |
//
// Air-gapped mode blocks ALL external LLM providers, only allows local.
func LLMPolicyRouter() gin.HandlerFunc {
	cache := &policyCache{
		mu:    &sync.RWMutex{},
		items: make(map[string]cachedDecision),
	}

	return func(c *gin.Context) {
		// Only apply to AI-related routes
		if !isAIRoute(c.Request.URL.Path) {
			c.Next()
			return
		}

		ontologyID := c.Param("id")
		if ontologyID == "" {
			ontologyID = c.Param("ontologyId")
		}
		if ontologyID == "" {
			c.Next()
			return
		}

		// Check cache first
		decision, ok := cache.get(ontologyID)
		if !ok {
			deployMode := getDeployMode()
			visibility := getOntologyVisibility(c, ontologyID)
			providerType := getProviderType()

			decision = evaluatePolicy(deployMode, visibility, providerType, c)
			cache.set(ontologyID, decision)
		}

		if decision.Allowed {
			slog.Debug("llm.policy.allowed",
				"ontology_id", ontologyID,
				"reason", decision.Reason,
				"trace_id", c.GetHeader("X-Trace-Id"),
			)
			c.Next()
			return
		}

		slog.Warn("llm.policy.blocked",
			"ontology_id", ontologyID,
			"reason", decision.Reason,
			"user_id", c.GetHeader("X-User-Id"),
			"trace_id", c.GetHeader("X-Trace-Id"),
		)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": gin.H{
				"code":    "LLM-POLICY-BLOCKED",
				"message": decision.Reason,
			},
		})
	}
}

// isAIRoute checks if the request path is an AI-related route.
func isAIRoute(path string) bool {
	// Exact match for ontology list endpoint
	if path == "/api/v1/ontologies" {
		return true
	}
	// Prefix match for ontology-specific routes
	aiPrefixes := []string{
		"/api/v1/ontologies/",
	}
	for _, p := range aiPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// getDeployMode reads the DEPLOYMENT_MODE env var.
func getDeployMode() string {
	mode := os.Getenv("DEPLOYMENT_MODE")
	switch mode {
	case DeployModeSaaS, DeployModeOnPremise, DeployModeAirGapped:
		return mode
	default:
		return DeployModeSaaS
	}
}

// getProviderType determines the LLM provider type from env config.
// Returns "local" if LLM_PROVIDER is set to a local-only provider (e.g., ollama),
// "external" otherwise.
func getProviderType() string {
	provider := os.Getenv("LLM_PROVIDER")

	switch strings.ToLower(provider) {
	case "ollama", "llama", "local-llm", "":
		return "local"
	default:
		return "external"
	}
}

// getOntologyVisibility determines the ontology visibility from request context.
// In the current phase, this reads from X-Ontology-Visibility header or falls back
// to "private" (most restrictive default).
func getOntologyVisibility(c *gin.Context, ontologyID string) string {
	// Try header first (set by upstream or cached from earlier lookup)
	visibility := c.GetHeader("X-Ontology-Visibility")
	if visibility != "" {
		return visibility
	}

	// Try query parameter
	visibility = c.Query("visibility")
	if visibility != "" {
		return visibility
	}

	// Fall back to most restrictive default
	slog.Debug("llm.policy.no_visibility_header",
		"ontology_id", ontologyID,
		"fallback", "private",
		"trace_id", c.GetHeader("X-Trace-Id"),
	)
	return "private"
}

// evaluatePolicy applies the LLM access policy matrix.
func evaluatePolicy(deployMode, visibility, providerType string, c *gin.Context) PolicyDecision {
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
			// Check admin override
			adminOverride := c.GetHeader("X-Admin-Override")
			if adminOverride == "true" {
				slog.Warn("llm.policy.admin_override",
					"ontology_id", c.Param("id"),
					"user_id", c.GetHeader("X-User-Id"),
					"trace_id", c.GetHeader("X-Trace-Id"),
				)
				return PolicyDecision{
					Allowed: true,
					Reason:  "Admin override: LLM access allowed for non-public ontology.",
				}
			}

			return PolicyDecision{
				Allowed: false,
				Reason:  "LLM access to " + visibility + " ontology is blocked in SaaS mode with external provider. Use a local LLM provider or request admin override.",
			}
		}
	}

	// Default: allow
	return PolicyDecision{
		Allowed: true,
		Reason:  "LLM access allowed.",
	}
}

// policyCache provides TTL-based caching for policy decisions.
type policyCache struct {
	mu    *sync.RWMutex
	items map[string]cachedDecision
}

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

func (pc *policyCache) set(key string, decision PolicyDecision) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.items[key] = cachedDecision{
		decision: decision,
		expires:  time.Now().Add(60 * time.Second),
	}
}
