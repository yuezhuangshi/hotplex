# API Reference

Complete API reference for HotPlex.

## Engine API

### `NewEngine(opts EngineOptions) (*Engine, error)`

Creates a new HotPlex Engine instance.

**Parameters:**
- `opts` - Engine configuration options

**Returns:**
- `*Engine` - Engine instance
- `error` - Error if initialization fails

---

### `engine.Execute(ctx, cfg, prompt, callback) error`

Executes a prompt with the AI agent.

**Parameters:**
- `ctx context.Context` - Execution context
- `cfg *Config` - Session configuration
- `prompt string` - User prompt
- `callback Callback` - Event streaming callback

**Returns:**
- `error` - Execution error

---

### `engine.Close() error`

Closes the engine and terminates all sessions.

---

### `engine.GetSessionStats(sessionID string) *SessionStats`

Returns statistics for a session.

---

### `engine.StopSession(sessionID, reason string) error`

Stops a running session.

---

## Config Types

### `Config`

```go
type Config struct {
    WorkDir          string  // Working directory
    SessionID        string  // Session identifier
    TaskInstructions string  // System instructions
}
```

### `EngineOptions`

```go
type EngineOptions struct {
    Logger          *slog.Logger  // Custom logger
    Timeout         time.Duration // Execution timeout
    IdleTimeout     time.Duration // Session idle timeout
    Namespace       string        // Namespace for logging
    PermissionMode  string        // Permission mode
    AllowedTools    []string      // Allowed tool list
    DisallowedTools []string      // Disallowed tool list
    Provider        Provider      // Custom provider
    AdminToken      string        // Admin token for bypass
}
```

## Event Types

| Type | Description | Data Type |
|------|-------------|-----------|
| `thinking` | AI is processing | `string` |
| `tool_use` | Tool invocation | `EventMeta` |
| `tool_result` | Tool result | `EventMeta` |
| `answer` | Text response | `string` |
| `error` | Error | `string` |
| `session_stats` | Statistics | `SessionStatsData` |
| `danger_block` | Security block | `DangerEvent` |

## Error Types

| Error | Description |
|-------|-------------|
| `ErrDangerBlocked` | Dangerous operation blocked by WAF |
| `ErrInvalidConfig` | Invalid configuration |
| `ErrSessionNotFound` | Session not found |
| `ErrSessionDead` | Session is dead |
| `ErrTimeout` | Operation timed out |
| `ErrInputTooLarge` | Input exceeds maximum size |
| `ErrProcessStart` | Failed to start CLI process |
| `ErrPipeClosed` | Pipe closed unexpectedly |

## Health Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /health` | Overall health status |
| `GET /health/ready` | Readiness probe (K8s) |
| `GET /health/live` | Liveness probe (K8s) |
| `GET /metrics` | Prometheus metrics |
