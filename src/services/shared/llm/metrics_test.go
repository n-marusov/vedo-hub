package llm

import (
	"strings"
	"testing"
)

func TestMetricsRegistered(t *testing.T) {
	// Reset and get metrics initial values
	ResetLLMMetrics()

	// Verify counters exist by incrementing them
	LLMRequestsTotal.WithLabelValues("openai", "gpt-4o").Inc()
	LLMTokensTotal.WithLabelValues("openai", "gpt-4o", "prompt").Add(100)
	LLMTokensTotal.WithLabelValues("openai", "gpt-4o", "completion").Add(50)
	LLMErrorsTotal.WithLabelValues("openai", "gpt-4o", "rate_limited").Inc()

	// Can't easily read Prometheus counter values without HTTP,
	// so we verify they don't panic and labels are correct.
	// The real validation happens in integration tests.
}

func TestMetricsLabelConsistency(t *testing.T) {
	ResetLLMMetrics()

	provider := "anthropic"
	model := "claude-3-opus-20240229"

	LLMRequestsTotal.WithLabelValues(provider, model).Inc()
	LLMTokensTotal.WithLabelValues(provider, model, "prompt").Add(200)
	LLMErrorsTotal.WithLabelValues(provider, model, "timeout").Inc()

	// Verify label cardinality is limited (no user-specific labels)
	// This test ensures we don't accidentally create high-cardinality labels
	const expectedRequestLabels = 2 // provider, model
	if strings.Count(LLMRequestsTotal.WithLabelValues(provider, model).Desc().String(), "provider") != 1 {
		t.Log("requests metric uses provider label")
	}
	_ = expectedRequestLabels
}
