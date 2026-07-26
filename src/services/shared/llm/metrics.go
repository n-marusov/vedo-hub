package llm

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// LLM metrics for Prometheus monitoring.
var (
	// LLMRequestsTotal counts total LLM API requests by provider and model.
	LLMRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_requests_total",
			Help: "Total number of LLM API requests.",
		},
		[]string{"provider", "model"},
	)

	// LLMTokensTotal counts tokens used by provider, model, and type (prompt/completion).
	LLMTokensTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_tokens_total",
			Help: "Total number of LLM tokens used.",
		},
		[]string{"provider", "model", "type"},
	)

	// LLMErrorsTotal counts LLM errors by provider, model, and error type.
	LLMErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_errors_total",
			Help: "Total number of LLM errors.",
		},
		[]string{"provider", "model", "error"},
	)

	// LLMLatencySeconds measures LLM request latency by provider and model.
	LLMLatencySeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llm_latency_seconds",
			Help:    "LLM request latency in seconds.",
			Buckets: []float64{0.5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"provider", "model"},
	)
)

// ResetLLMMetrics resets all LLM metrics to zero.
// Useful for testing to ensure clean state between test cases.
func ResetLLMMetrics() {
	LLMRequestsTotal.Reset()
	LLMTokensTotal.Reset()
	LLMErrorsTotal.Reset()
	LLMLatencySeconds.Reset()
}
