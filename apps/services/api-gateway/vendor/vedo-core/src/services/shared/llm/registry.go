package llm

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

// Global provider registry.
var (
	providersMu sync.RWMutex
	providers   = make(map[string]ProviderFactory)
)

// Register registers a provider factory with the given name.
// If a factory is already registered under the same name, it panics.
// This is intended to be called from provider init() functions or
// explicit startup registration.
func Register(name string, factory ProviderFactory) {
	providersMu.Lock()
	defer providersMu.Unlock()

	if _, exists := providers[name]; exists {
		panic(fmt.Sprintf("llm: provider %q already registered", name))
	}
	providers[name] = factory
}

// NewProvider creates a new Provider instance based on the LLM_PROVIDER
// environment variable (or the given provider name). It looks up the
// registered factory and calls it with the resolved configuration.
func NewProvider(providerName ...string) (Provider, error) {
	name := getEnv("LLM_PROVIDER", "openai")
	if len(providerName) > 0 && providerName[0] != "" {
		name = providerName[0]
	}

	providersMu.RLock()
	factory, ok := providers[name]
	providersMu.RUnlock()

	if !ok {
		available := AvailableProviders()
		return nil, fmt.Errorf("llm: unknown provider %q (available: %v)", name, available)
	}

	cfg := LoadConfigFromEnv()
	// Override config from provider-specific env vars
	cfg = applyProviderEnvOverrides(name, cfg)

	log.Printf("[INFO] llm: initializing provider %q with model %q", name, cfg.Model)

	return factory(cfg)
}

// AvailableProviders returns the list of registered provider names.
func AvailableProviders() []string {
	providersMu.RLock()
	defer providersMu.RUnlock()

	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	return names
}

// applyProviderEnvOverrides reads provider-specific environment variables
// and overrides the base config. It looks for LLM_PROVIDERS_<NAME>_* vars.
func applyProviderEnvOverrides(name string, cfg Config) Config {
	// Replace hyphens with underscores for env var naming compatibility
	sanitized := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
	prefix := fmt.Sprintf("LLM_PROVIDERS_%s_", sanitized)

	if baseURL := os.Getenv(prefix + "BASE_URL"); baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if apiKey := os.Getenv(prefix + "API_KEY"); apiKey != "" {
		cfg.APIKey = apiKey
	}
	if model := os.Getenv(prefix + "MODEL"); model != "" {
		cfg.Model = model
	}
	if timeout := os.Getenv(prefix + "TIMEOUT"); timeout != "" {
		if d := parseDurationEnv(prefix+"TIMEOUT", 0); d > 0 {
			cfg.Timeout = d
		}
	}

	return cfg
}
