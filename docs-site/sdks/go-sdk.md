# Go SDK

The Go SDK provides native integration with HotPlex for Go applications.

## Installation

```bash
go get github.com/hrygo/hotplex
```

## Types

### Engine

The core control plane for AI CLI agent integration.

```go
type Engine struct { ... }

func NewEngine(options EngineOptions) (*Engine, error)
func (e *Engine) Execute(ctx context.Context, cfg *Config, prompt string, callback Callback) error
func (e *Engine) Close() error
func (e *Engine) GetSessionStats(sessionID string) *SessionStats
func (e *Engine) StopSession(sessionID string, reason string) error
```

### Config

Session configuration.

```go
type Config struct {
    WorkDir          string  // Working directory for CLI
    SessionID        string  // Unique session identifier
    TaskInstructions string  // System prompt for the AI
}
```

### EngineOptions

Engine initialization options.

```go
type EngineOptions struct {
    Logger           *slog.Logger
    Timeout          time.Duration
    IdleTimeout      time.Duration
    Namespace        string
    PermissionMode   string
    AllowedTools     []string
    DisallowedTools  []string
    Provider         Provider
    AdminToken       string
}
```

### Callback

Event streaming callback function.

```go
type Callback func(eventType string, data any) error
```

## Error Handling

```go
import "github.com/hrygo/hotplex"

err := engine.Execute(ctx, cfg, prompt, callback)
if errors.Is(err, hotplex.ErrDangerBlocked) {
    // Dangerous operation was blocked
}
if errors.Is(err, hotplex.ErrSessionDead) {
    // Session is no longer alive
}
if errors.Is(err, hotplex.ErrTimeout) {
    // Operation timed out
}
```

## Version

```go
fmt.Println(hotplex.Version)      // "1.0.0"
fmt.Println(hotplex.VersionMajor) // 1
fmt.Println(hotplex.VersionMinor) // 0
fmt.Println(hotplex.VersionPatch) // 0
```

## Examples

See the [_examples](https://github.com/hrygo/hotplex/tree/main/_examples) directory for complete examples:

- `go_claude_basic` - Basic usage
- `go_claude_lifecycle` - Session lifecycle management
- `go_opencode_basic` - OpenCode integration
- `go_error_handling` - Error handling patterns
