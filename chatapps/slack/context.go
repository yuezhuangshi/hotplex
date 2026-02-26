package slack

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

// contextKey is a private type for context keys to avoid collisions
type contextKey string

const (
	// traceIDKey is the context key for trace ID
	traceIDKey contextKey = "trace_id"
)

// generateTraceID generates a unique trace ID for request tracking
func generateTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to simple ID if random generation fails
		return "trace-unknown"
	}
	return "trace-" + hex.EncodeToString(b[:8])
}

// WithTraceID returns a new context with the given trace ID
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromContext extracts the trace ID from context
func TraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return generateTraceID()
}

// NewContextWithTraceID creates a new context with a generated trace ID
// Returns the context and the trace ID
func NewContextWithTraceID(ctx context.Context) (context.Context, string) {
	traceID := generateTraceID()
	return WithTraceID(ctx, traceID), traceID
}
