package llm

import (
	"context"
	"log"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// tracerName is the name used for the OTel tracer.
const tracerName = "vedo-core/src/services/shared/llm"

// ObservabilityProvider wraps a Provider with OpenTelemetry tracing,
// Prometheus metrics, and cost tracking.
type ObservabilityProvider struct {
	inner  Provider
	name   string
	model  string
	tracer trace.Tracer
}

// WrapWithObservability wraps a Provider with observability instrumentation.
// It creates OTel spans and records Prometheus metrics for every LLM call.
func WrapWithObservability(name string, provider Provider) *ObservabilityProvider {
	tracer := otel.Tracer(tracerName)
	return &ObservabilityProvider{
		inner:  provider,
		name:   name,
		model:  "",
		tracer: tracer,
	}
}

// Complete sends a prompt to the LLM with full observability instrumentation.
func (p *ObservabilityProvider) Complete(ctx Context, prompt Prompt) (Completion, error) {
	// Create base context for OTel if not set
	otelCtx := context.Background()
	if ctx.BaseCtx != nil {
		otelCtx = ctx.BaseCtx
	}

	// Start OTel span
	_, span := p.tracer.Start(otelCtx, "llm."+p.name+".complete",
		trace.WithAttributes(
			attribute.String("llm.provider", p.name),
			attribute.String("llm.model", resolveModel(prompt.Model, p.model)),
			attribute.Int("llm.prompt_messages", len(prompt.Messages)),
			attribute.Float64("llm.temperature", prompt.Temperature),
		),
	)
	defer span.End()

	// Call the inner provider
	completion, err := p.inner.Complete(ctx, prompt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(
			attribute.String("llm.error", err.Error()),
		)

		// Record error metrics
		LLMErrorsTotal.WithLabelValues(p.name, p.model, classifyError(err)).Inc()

		log.Printf("[ERROR] llm/observability: provider=%q error=%v", p.name, err)
		return completion, err
	}

	// Set success attributes
	span.SetAttributes(
		attribute.Int("llm.prompt_tokens", completion.PromptTokens),
		attribute.Int("llm.completion_tokens", completion.TokensUsed),
		attribute.String("llm.finish_reason", string(completion.FinishReason)),
	)
	span.SetStatus(codes.Ok, "")

	// Record metrics
	LLMRequestsTotal.WithLabelValues(p.name, p.model).Inc()
	if completion.PromptTokens > 0 {
		LLMTokensTotal.WithLabelValues(p.name, p.model, "prompt").Add(float64(completion.PromptTokens))
	}
	if completion.TokensUsed > 0 {
		LLMTokensTotal.WithLabelValues(p.name, p.model, "completion").Add(float64(completion.TokensUsed))
	}
	if completion.PromptTokens > 0 || completion.TokensUsed > 0 {
		cost := EstimateCost(p.name, p.model, completion.PromptTokens, completion.TokensUsed)
		LLMCostTotal.WithLabelValues(p.name, p.model).Add(cost)
		log.Printf("[INFO] llm/cost: provider=%q model=%q cost=$%.6f (prompt=%d comp=%d)",
			p.name, p.model, cost, completion.PromptTokens, completion.TokensUsed)
	}

	return completion, nil
}

// StreamComplete sends a prompt and returns a channel of tokens with tracing.
func (p *ObservabilityProvider) StreamComplete(ctx Context, prompt Prompt) (<-chan Token, error) {
	otelCtx := context.Background()
	if ctx.BaseCtx != nil {
		otelCtx = ctx.BaseCtx
	}

	_, span := p.tracer.Start(otelCtx, "llm."+p.name+".stream",
		trace.WithAttributes(
			attribute.String("llm.provider", p.name),
			attribute.String("llm.model", resolveModel(prompt.Model, p.model)),
			attribute.Int("llm.prompt_messages", len(prompt.Messages)),
		),
	)
	defer span.End()

	ch, err := p.inner.StreamComplete(ctx, prompt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(codes.Ok, "")
	LLMRequestsTotal.WithLabelValues(p.name, p.model).Inc()

	return ch, nil
}

// classifyError categorizes an error for metrics labelling.
func classifyError(err error) string {
	switch {
	case isError(err, "rate limited"):
		return "rate_limited"
	case isError(err, "timeout"):
		return "timeout"
	case isError(err, "context too long"):
		return "context_too_long"
	case isError(err, "invalid response"):
		return "invalid_response"
	case isError(err, "provider unavailable"):
		return "provider_unavailable"
	default:
		return "unknown"
	}
}

func isError(err error, substr string) bool {
	return err != nil && strings.Contains(err.Error(), substr)
}

// resolveModel returns the model to use, preferring the prompt-level model.
func resolveModel(promptModel, providerModel string) string {
	if promptModel != "" {
		return promptModel
	}
	return providerModel
}
