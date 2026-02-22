---
layout: home

hero:
  name: "HotPlex"
  text: "AI Agent Control Plane"
  tagline: Transforming AI CLI Agents into Production-Ready Interactive Services
  image:
    src: /logo.svg
    alt: HotPlex
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/hrygo/hotplex

features:
  - icon: 🔄
    title: Cli-as-a-Service
    details: Shift from "run-and-exit" to persistent sessions with continuous instruction flow and context preservation.

  - icon: 🧩
    title: Ease of Integration
    details: A unified Go SDK and protocol gateway that plugs top-tier Agent capabilities into your product instantly.

  - icon: 🚀
    title: Zero Spin-up Overhead
    details: Eliminate the long wait times for Node.js/Python runtimes to provide sub-second user feedback.

  - icon: 🛡️
    title: Fast & Balanced Security
    details: Command-level WAF and PGID isolation provide a protective glove for AI shell operations.

  - icon: 🔌
    title: Ready for Scale
    details: Support for native Go embedding or standalone Proxy mode with WebSocket and OpenCode-compatible protocols.

  - icon: 📊
    title: Production Observability
    details: Built-in OpenTelemetry tracing, Prometheus metrics, and health check endpoints.
---

## Quick Start

### Install

```bash
go get github.com/hrygo/hotplex
```

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/hrygo/hotplex"
)

func main() {
    opts := hotplex.EngineOptions{
        Timeout:         5 * time.Minute,
        PermissionMode:  "bypassPermissions",
        AllowedTools:    []string{"Bash", "Edit", "Read", "FileSearch"},
    }
    engine, _ := hotplex.NewEngine(opts)
    defer engine.Close()

    cfg := &hotplex.Config{
        WorkDir:          "/tmp/ai-sandbox",
        SessionID:        "user-123",
        TaskInstructions: "You are a senior Go systems engineer.",
    }

    ctx := context.Background()
    err := engine.Execute(ctx, cfg, "Refactor the main.go", 
        func(eventType string, data any) error {
            if eventType == "answer" {
                fmt.Printf("🤖 %v\n", data)
            }
            return nil
        })
}
```

## Multi-Language SDKs

HotPlex provides official SDKs for multiple languages:

| SDK | Status | Description |
|-----|--------|-------------|
| [Go SDK](/sdks/go-sdk) | ✅ v1.0 | Native Go integration |
| [Python SDK](/sdks/python-sdk) | ✅ v1.0 | WebSocket client for Python |
| [TypeScript SDK](/sdks/typescript-sdk) | ✅ v1.0 | TypeScript/JavaScript client |

## Roadmap

We are actively evolving HotPlex to become the definitive execution engine for the Local AI ecosystem.

- [x] Provider Abstraction (Claude Code, OpenCode)
- [x] Event Hooks System
- [x] OpenTelemetry Integration
- [x] Prometheus Metrics
- [x] Docker Remote Execution
- [x] Multi-Language SDKs
- [ ] L2/L3 Isolation (H2 2026)
