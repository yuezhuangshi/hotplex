package brain

import (
	"os"
)

// Config holds the configuration for the Global Brain.
type Config struct {
	// Enabled is automatically determined based on APIKey presence.
	Enabled bool
	// Provider supports "openai" (default), "anthropic", "gemini".
	Provider string
	// APIKey is the secret for accessing the provider API.
	APIKey string
	// Endpoint is the optional base URL for the API (e.g. for DeepSeek/Groq).
	Endpoint string
	// Model is the specific model to use (default: gpt-4o-mini).
	Model string
	// Timeout is the maximum duration for a brain request.
	// Defaults to 10 seconds for standard requests.
	TimeoutS int
}

// LoadConfigFromEnv loads the brain configuration from environment variables.
func LoadConfigFromEnv() Config {
	apiKey := os.Getenv("HOTPLEX_BRAIN_API_KEY")

	return Config{
		Enabled:  apiKey != "",
		Provider: getEnv("HOTPLEX_BRAIN_PROVIDER", "openai"),
		APIKey:   apiKey,
		Endpoint: os.Getenv("HOTPLEX_BRAIN_ENDPOINT"),
		Model:    getEnv("HOTPLEX_BRAIN_MODEL", "gpt-4o-mini"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
