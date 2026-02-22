# OpenCode Provider Integration Guide

*Read this in other languages: [English](opencode.md), [简体中文](opencode_zh.md).*

This document describes how to integrate and use the **OpenCode** CLI agent within the HotPlex ecosystem.

## Overview

Unlike Claude Code, which is optimized for interactive REPL-style Stdin communication, **OpenCode** is designed with a "Server-First" architecture. It provides multiple ways to integrate:
1. **CLI Mode**: Executing `opencode run` for one-off tasks.
2. **Server Mode**: Running `opencode serve` as a persistent background process.
3. **ACP Protocol**: Using the **Agent Client Protocol** for structured, multi-turn interactions.

HotPlex leverages these modes to provide sub-second response times and persistent session management.

## Installation

Ensure you have OpenCode installed globally:

```bash
npm install -g opencode
```

Verify the installation:
```bash
opencode --version
```

## Configuration

The `OpenCodeProvider` can be configured via `hotplex.ProviderConfig`.

### Provider Options

| Option         | Type     | Description                                                       |
| :------------- | :------- | :---------------------------------------------------------------- |
| `use_http_api` | `bool`   | If true, HotPlex talks to a background `opencode serve` instance. |
| `port`         | `int`    | The port for the OpenCode server (default: 4096).                 |
| `plan_mode`    | `bool`   | Enables OpenCode's planning mode (auto-permission for read-only). |
| `provider`     | `string` | The LLM provider (e.g., `openai`, `anthropic`, `siliconflow`).    |
| `model`        | `string` | The specific model ID (e.g., `zhipu/glm-5-code-plan`, `gpt-4o`).  |

## Architecture

### 1. CLI Mode (Default)
In this mode, HotPlex executes `opencode run` for each request. 
- **Pros**: Simple, no background server needed.
- **Cons**: Higher latency (Cold Start) as the CLI must initialize for every turn.

### 2. HTTP Server Mode (Recommended)
HotPlex starts and manages an `opencode serve` process. Requests are sent via the HTTP API, and events are captured via SSE (Server-Sent Events).
- **Pros**: Sub-second latency, robust multi-turn state.
- **Cons**: Requires managing a background process.

## Example: Using OpenCode in Go

```go
package main

import (
	"context"
	"fmt"
	"github.com/hrygo/hotplex"
)

func main() {
	// 1. Initialize OpenCode Provider
	opencodePrv, _ := hotplex.NewOpenCodeProvider(hotplex.ProviderConfig{
		Type:         hotplex.ProviderTypeOpenCode,
		DefaultModel: "zhipu/glm-5-code-plan", // Recommended for code-plan tasks
		OpenCode: &hotplex.OpenCodeConfig{
			PlanMode:   true,  // Enable planning mode
			UseHTTPAPI: true,  // Use server mode for low latency
		},
	}, nil)

	// 2. Wrap in Engine
	engine, _ := hotplex.NewEngine(hotplex.EngineOptions{
		Provider: opencodePrv,
	})
	defer engine.Close()

	// 3. Execute
	ctx := context.Background()
	engine.Execute(ctx, &hotplex.Config{SessionID: "my-task"}, "Analyze this repo", 
		func(eventType string, data any) error {
			if eventType == "answer" {
				fmt.Print(data.(*hotplex.EventWithMeta).EventData)
			}
			return nil
		})
}
```

## Protocol Mapping

OpenCode events are automatically mapped to HotPlex's unified event model:

| OpenCode Part Type | HotPlex Event Type |
| :----------------- | :----------------- |
| `text`             | `answer`           |
| `reasoning`        | `thinking`         |
| `tool` (input)     | `tool_use`         |
| `tool` (output)    | `tool_result`      |
| `step-start`       | `step_start`       |
| `step-finish`      | `step_finish`      |

## Advanced Integration Patterns (Latest Best Practices)

### 1. Agent Client Protocol (ACP)
For deep integration (IDEs or custom control planes), use `opencode acp`. 
- **Communication**: JSON-RPC over `stdin`/`stdout`.
- **Capability**: Allows the client to control tool authorization and observe granular agent reasoning.
- **Latency**: Persistent process eliminates cold start.

### 2. Repository-Level Context (`AGENTS.md`)
Always run `opencode run "/init"` in your project root first. This generates:
- **`AGENTS.md`**: A map for the AI to understand your project structure.
- **`opencode.json`**: Rule definitions and model preferences.
HotPlex can dynamically monitor or inject rules into these files for smarter orchestration.

### 3. MCP (Model Context Protocol) Support
OpenCode supports MCP servers. You can extend OpenCode's capabilities by adding MCP servers to its configuration, allowing it to interact with external databases, APIs, or custom enterprise tools.

### 4. Session Persistence & State
OpenCode stores all conversation history in `~/.local/share/opencode/opencode.db`. 
- Use the `--session <id>` flag with `opencode run` to resume specific tasks.
- HotPlex can inspect this database to provide cross-session reporting and analytics.
