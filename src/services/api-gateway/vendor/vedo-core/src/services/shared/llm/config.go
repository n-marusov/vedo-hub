package llm

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the LLM provider configuration loaded from environment variables.
type Config struct {
	// Provider is the name of the default LLM provider (e.g., "openai", "ollama").
	Provider string

	// APIKey is the API key for the default provider.
	APIKey string

	// Model is the model name to use (e.g., "gpt-4o", "llama3").
	Model string

	// BaseURL is the base URL for the API endpoint.
	BaseURL string

	// Timeout is the HTTP client timeout for LLM requests.
	Timeout time.Duration

	// MaxTokens is the maximum number of tokens in the response.
	MaxTokens int

	// Temperature controls randomness in generation (0.0-2.0).
	Temperature float64
}

// LoadConfigFromEnv reads LLM configuration from environment variables.
// It applies default values when variables are missing or invalid.
func LoadConfigFromEnv() Config {
	cfg := Config{
		Provider:    getEnv("LLM_PROVIDER", "openai"),
		APIKey:      getEnv("LLM_API_KEY", ""),
		Model:       getEnv("LLM_MODEL", "gpt-4o"),
		BaseURL:     getEnv("LLM_BASE_URL", ""),
		Timeout:     parseDurationEnv("LLM_TIMEOUT", 30*time.Second),
		MaxTokens:   parseIntEnv("LLM_MAX_TOKENS", 4096),
		Temperature: parseTemperatureEnv("LLM_TEMPERATURE", 0.7),
	}

	return cfg
}

// getEnv returns the value of the environment variable or the default.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// parseIntEnv parses an integer environment variable, returning the default on failure.
func parseIntEnv(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return n
}

// parseDurationEnv parses a duration environment variable (e.g., "30s"), returning the default on failure.
func parseDurationEnv(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return d
}

// parseTemperatureEnv parses a temperature value, clamping to [0.0, 2.0].
func parseTemperatureEnv(key string, defaultVal float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return defaultVal
	}
	if f < 0.0 {
		return 0.0
	}
	if f > 2.0 {
		return 2.0
	}
	return f
}

// ParseProviderName extracts a provider name from an LLM_PROVIDERS_<NAME>_* key.
// For example, "LLM_PROVIDERS_OPENAI_API_KEY" returns "OPENAI".
func ParseProviderName(envKey, prefix string) (string, bool) {
	if !strings.HasPrefix(envKey, prefix) {
		return "", false
	}
	suffix := strings.TrimPrefix(envKey, prefix)
	parts := strings.SplitN(suffix, "_", 2)
	if len(parts) < 2 {
		return "", false
	}
	return parts[0], true
}
