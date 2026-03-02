package brain

import (
	"context"
)

// Brain represents the core "System 1" intelligence for HotPlex.
// It provides fast, structured, and low-cost reasoning capabilities.
type Brain interface {
	// Chat generates a plain text response for a given prompt.
	// Best used for simple questions, greetings, or summarization.
	Chat(ctx context.Context, prompt string) (string, error)

	// Analyze performs structured analysis and returns the result in the target struct.
	// The target must be a pointer to a struct that can be unmarshaled from JSON.
	// Useful for intent routing, safety checks, and complex data extraction.
	Analyze(ctx context.Context, prompt string, target any) error
}

var (
	globalBrain Brain
)

// Global returns the globally configured Brain instance.
// If no brain is configured, it returns nil.
func Global() Brain {
	return globalBrain
}

// SetGlobal sets the global Brain instance.
func SetGlobal(b Brain) {
	globalBrain = b
}
