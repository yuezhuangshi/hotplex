# Quick Start

Build your first HotPlex application in 5 minutes.

## Prerequisites

- Go 1.24+
- Claude Code CLI installed (`npm install -g @anthropic-ai/claude-code`)

## Basic Example

Create a file `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/hrygo/hotplex"
)

func main() {
    // 1. Initialize engine singleton
    opts := hotplex.EngineOptions{
        Timeout:         5 * time.Minute,
        PermissionMode:  "bypassPermissions",
        AllowedTools:    []string{"Bash", "Edit", "Read", "FileSearch"},
    }
    engine, err := hotplex.NewEngine(opts)
    if err != nil {
        panic(err)
    }
    defer engine.Close()

    // 2. Configure persistent session routing
    cfg := &hotplex.Config{
        WorkDir:          "/tmp/ai-sandbox",
        SessionID:        "user-123",
        TaskInstructions: "You are a senior Go systems engineer.",
    }

    // 3. Execute with streaming callback
    ctx := context.Background()
    err = engine.Execute(ctx, cfg, "List the files in the current directory", 
        func(eventType string, data any) error {
            switch eventType {
            case "answer":
                fmt.Printf("🤖 %v\n", data)
            case "tool_use":
                fmt.Printf("🔧 Tool: %v\n", data)
            case "error":
                fmt.Printf("❌ Error: %v\n", data)
            }
            return nil
        })
    
    if err != nil {
        fmt.Printf("Execution failed: %v\n", err)
    }
}
```

## Run

```bash
go mod init myapp
go mod tidy
go run main.go
```

## Multi-Turn Conversation

HotPlex maintains session state automatically:

```go
// First turn - creates session
engine.Execute(ctx, cfg, "Create a file called hello.txt", callback)

// Second turn - reuses same session
engine.Execute(ctx, cfg, "Add 'Hello, World!' to the file", callback)

// Third turn - context is preserved
engine.Execute(ctx, cfg, "Read the file back to me", callback)
```

## Event Types

| Event | Description |
|-------|-------------|
| `thinking` | AI is processing |
| `tool_use` | Tool invocation started |
| `tool_result` | Tool execution completed |
| `answer` | Text response from AI |
| `error` | Error occurred |
| `session_stats` | Session statistics |

## Next Steps

- [Go SDK Reference](/sdks/go-sdk)
- [Architecture Deep Dive](/guide/architecture)
- [Security Model](/guide/security)
