package llm

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// LLMCostTotal tracks the estimated cost of LLM API calls.
var LLMCostTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "llm_cost_total_usd",
		Help: "Estimated total cost of LLM API calls in USD.",
	},
	[]string{"provider", "model"},
)

// ModelPricing represents token pricing for an LLM model.
type ModelPricing struct {
	// PromptPrice is the price per 1K input tokens in USD.
	PromptPrice float64
	// CompletionPrice is the price per 1K output tokens in USD.
	CompletionPrice float64
}

// defaultPricing is a map of known model pricing.
// Prices are in USD per 1K tokens.
// Source: provider pricing pages (as of 2026-07).
var defaultPricing = map[string]map[string]ModelPricing{
	"openai": {
		"gpt-4o":        {PromptPrice: 0.0025, CompletionPrice: 0.01},
		"gpt-4o-mini":   {PromptPrice: 0.00015, CompletionPrice: 0.0006},
		"gpt-4-turbo":   {PromptPrice: 0.01, CompletionPrice: 0.03},
		"gpt-4":         {PromptPrice: 0.03, CompletionPrice: 0.06},
		"gpt-3.5-turbo": {PromptPrice: 0.0005, CompletionPrice: 0.0015},
	},
	"anthropic": {
		"claude-3-opus-20240229":     {PromptPrice: 0.015, CompletionPrice: 0.075},
		"claude-3-sonnet-20240229":   {PromptPrice: 0.003, CompletionPrice: 0.015},
		"claude-3-haiku-20240307":    {PromptPrice: 0.00025, CompletionPrice: 0.00125},
		"claude-3-5-sonnet-20241022": {PromptPrice: 0.003, CompletionPrice: 0.015},
	},
	"ollama": {
		"*": {PromptPrice: 0.0, CompletionPrice: 0.0}, // Local models are free
	},
}

// EstimateCost estimates the cost of an LLM call based on provider, model, and token usage.
// It returns the estimated cost in USD.
func EstimateCost(provider, model string, promptTokens, completionTokens int) float64 {
	providerPrices, ok := defaultPricing[provider]
	if !ok {
		return 0
	}

	// Look for specific model pricing
	pricing, ok := providerPrices[model]
	if !ok {
		// Fall back to wildcard, then first available
		pricing, ok = providerPrices["*"]
		if !ok {
			pricing, ok = findFirstPricing(providerPrices)
			if !ok {
				return 0
			}
		}
	}

	promptCost := float64(promptTokens) / 1000 * pricing.PromptPrice
	completionCost := float64(completionTokens) / 1000 * pricing.CompletionPrice

	return promptCost + completionCost
}

// GetModelPricing returns the pricing for a specific provider and model.
// Returns nil if no pricing is found.
func GetModelPricing(provider, model string) *ModelPricing {
	providerPrices, ok := defaultPricing[provider]
	if !ok {
		return nil
	}

	pricing, ok := providerPrices[model]
	if !ok {
		// Try wildcard
		pricing, ok = providerPrices["*"]
		if !ok {
			return nil
		}
	}

	return &pricing
}

// findFirstPricing returns the first pricing entry from the map.
func findFirstPricing(prices map[string]ModelPricing) (ModelPricing, bool) {
	for _, p := range prices {
		return p, true
	}
	return ModelPricing{}, false
}
