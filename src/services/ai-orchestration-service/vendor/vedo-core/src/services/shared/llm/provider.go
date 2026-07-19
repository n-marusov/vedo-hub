package llm

// Provider defines the unified interface for LLM providers.
// Implementations wrap provider-specific APIs (OpenAI, Anthropic, etc.)
// behind this common contract.
type Provider interface {
	// Complete sends a prompt to the LLM and returns the full completion.
	Complete(ctx Context, p Prompt) (Completion, error)

	// StreamComplete sends a prompt and returns a channel of tokens
	// as they are generated.
	StreamComplete(ctx Context, p Prompt) (<-chan Token, error)
}

// ProviderFactory is a function that creates a new Provider instance
// from the given configuration.
type ProviderFactory func(cfg Config) (Provider, error)
