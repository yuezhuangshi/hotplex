package brain

import (
	"context"
	"log/slog"

	"github.com/hrygo/hotplex/brain/llm"
)

// Init initializes the global Brain from environmental variables.
// It detects the provider and sets the Global Brain instance.
func Init(logger *slog.Logger) error {
	config := LoadConfigFromEnv()

	if !config.Enabled {
		logger.Debug("Native Brain is disabled or missing configuration. Skipping.")
		return nil
	}

	switch config.Provider {
	case "openai":
		// This uses OpenAI SDK for OpenAI, DeepSeek, Groq, etc.
		client := llm.NewOpenAIClient(config.APIKey, config.Endpoint, config.Model, logger)
		SetGlobal(&brainWrapper{client: client})
		logger.Info("Native Brain initialized", "provider", config.Provider, "model", config.Model)
	default:
		// Fallback for unknown provider
		logger.Warn("Unknown brain provider specified. Brain disabled.", "provider", config.Provider)
	}

	return nil
}

// brainWrapper satisfies the Brain interface using a client implementation.
type brainWrapper struct {
	client interface {
		Chat(ctx context.Context, prompt string) (string, error)
		Analyze(ctx context.Context, prompt string, target any) error
	}
}

func (w *brainWrapper) Chat(ctx context.Context, prompt string) (string, error) {
	return w.client.Chat(ctx, prompt)
}

func (w *brainWrapper) Analyze(ctx context.Context, prompt string, target any) error {
	return w.client.Analyze(ctx, prompt, target)
}
