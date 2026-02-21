package hotplex

import (
	"context"

	"github.com/hrygo/hotplex/event"
	"github.com/hrygo/hotplex/types"
)

// HotPlexClient defines the public API for the HotPlex engine.
// It abstracts the underlying process management and provides a clean, session-aware
// interface for callers to interact with Claude Code CLI agents.
type HotPlexClient interface {
	// Execute runs a command or prompt within the HotPlex sandbox and streams events back via the Callback.
	// It uses Hot-Multiplexing to reuse existing processes if a matching SessionID is provided in the Config.
	Execute(ctx context.Context, cfg *types.Config, prompt string, callback event.Callback) error

	// Close gracefully terminates all managed sessions in the pool and releases all OS-level resources.
	// This includes sweeping process groups (PGID) to ensure no zombie processes remain.
	Close() error
}
