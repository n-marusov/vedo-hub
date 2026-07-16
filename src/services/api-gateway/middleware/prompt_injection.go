package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/unicode/norm"

	"vedo-core/src/services/api-gateway/models"
)

// Common prompt injection patterns for pre-filter middleware.
// These detect attempts to override the system prompt or change the
// LLM's role before the request reaches the handler.
var (
	// System override attempts — "ignore previous instructions", "forget everything", etc.
	reIgnoreInstructions = regexp.MustCompile(`(?i)(?:ignore|disregard|skip|forget)\s+(?:all\s+)?(?:previous|prior|above|given)\s+(?:instructions?|commands?|directions?|rules?|context|prompts?|messages?)`)

	// Role-play injection — "you are now", "act as", "pretend you are"
	reNewRole = regexp.MustCompile(`(?i)(?:you\s+are\s+(?:now\s+)?|act\s+(?:as\s+)?|from\s+now\s+on\s+(?:you\s+are\s+)?)(?:DAN|hypnotized|freed|unleashed|unconstrained|ungoverned|unlimited|unrestricted|autonomous|override|admin(?:istrator)?|root|superuser|chatgpt|gpt-4|gpt-3\.5)`)

	// Delimiter injection — separator tokens that break prompt boundaries
	reDelimiterInjection = regexp.MustCompile(`(?m)^[-=]{3,}\s*$`)

	// XML tag injection — <system>, <instruction>, <prompt> tags that override
	reXMLInjection = regexp.MustCompile(`(?i)<\s*(?:system|instruction|prompt|assistant|user|role|context)\s*(?:\s+[^>]*)?>`)
)

// PromptInjectionPreFilter returns a Gin middleware that pre-filters
// user-supplied text in AI-related request bodies for prompt injection patterns.
//
// Pre-filtering happens BEFORE the request reaches the handler or LLM:
//   - High-confidence match -> 400 PROMPT-INJECTION-DETECTED (blocked)
//   - Low-confidence match -> pass through (handler-level sanitization still applies)
//   - No match -> pass through
func PromptInjectionPreFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-Id")
		path := c.Request.URL.Path

		// Only inspect JSON request bodies
		contentType := c.GetHeader("Content-Type")
		if !strings.HasPrefix(contentType, "application/json") {
			c.Next()
			return
		}

		// Read the body once
		body, err := c.GetRawData()
		if err != nil || len(body) == 0 {
			c.Next()
			return
		}

		// Normalize body to NFC for homoglyph detection
		normalized := norm.NFC.String(string(body))

		// Check for high-confidence injection patterns
		highMatch := containsHighConfidenceInjection(normalized)
		lowMatch := containsLowConfidenceInjection(normalized)

		if highMatch {
			slog.Warn("middleware.prompt_injection.blocked",
				"path", path,
				"trace_id", traceID,
				"confidence", "high",
			)
			c.AbortWithStatusJSON(http.StatusBadRequest, models.ErrorResponse{
				Error: models.ErrorDetail{
					Code:    "PROMPT-INJECTION-DETECTED",
					Message: "Request rejected due to security policy.",
				},
			})
			return
		}

		if lowMatch {
			slog.Warn("middleware.prompt_injection.low_confidence",
				"path", path,
				"trace_id", traceID,
				"confidence", "low",
			)
			// Low confidence: pass through — handler-level SanitizeLLMInput
			// and prompt hardening provide defense-in-depth.
		}

		// Restore the body for downstream handlers
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		c.Next()
	}
}

// containsHighConfidenceInjection checks for definitive injection patterns.
func containsHighConfidenceInjection(input string) bool {
	if reIgnoreInstructions.MatchString(input) {
		return true
	}
	if reNewRole.MatchString(input) {
		return true
	}
	return false
}

// containsLowConfidenceInjection checks for suspicious but not definitive patterns.
func containsLowConfidenceInjection(input string) bool {
	if reDelimiterInjection.MatchString(input) {
		return true
	}
	if reXMLInjection.MatchString(input) {
		return true
	}
	return false
}
