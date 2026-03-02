package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// OpenAIClient implements OpenAI-compatible LLM interactions.
// It can be used for OpenAI, DeepSeek, Groq, etc.
type OpenAIClient struct {
	client *openai.Client
	model  string
	logger *slog.Logger
}

// NewOpenAIClient creates a new OpenAI compatible client.
func NewOpenAIClient(apiKey, endpoint, model string, logger *slog.Logger) *OpenAIClient {
	config := openai.DefaultConfig(apiKey)
	if endpoint != "" {
		config.BaseURL = endpoint
	}

	return &OpenAIClient{
		client: openai.NewClientWithConfig(config),
		model:  model,
		logger: logger,
	}
}

// Chat generates a simple plain text completion.
func (c *OpenAIClient) Chat(ctx context.Context, prompt string) (string, error) {
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("openai chat error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("zero choices in response")
	}

	return resp.Choices[0].Message.Content, nil
}

// Analyze requests JSON formatted output from the model.
// It uses "ResponseFormat: {Type: JSON_OBJECT}" to ensure model compatibility for structured reasoning.
func (c *OpenAIClient) Analyze(ctx context.Context, prompt string, target any) error {
	// Instruct the model to return JSON if it's not explicitly in the prompt
	if !strings.Contains(strings.ToLower(prompt), "json") {
		prompt = prompt + "\n\nIMPORTANT: Return ONLY valid JSON."
	}

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)

	if err != nil {
		return fmt.Errorf("openai analyze error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return fmt.Errorf("zero choices in response")
	}

	content := resp.Choices[0].Message.Content
	err = json.Unmarshal([]byte(content), target)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON content: %w. CONTENT: %s", err, content)
	}

	return nil
}
