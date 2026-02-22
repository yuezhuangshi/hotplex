# Getting Started

HotPlex is a high-performance AI Agent Control Plane that transforms AI CLI agents (like Claude Code, OpenCode) into production-ready interactive services.

## What is HotPlex?

Instead of reinventing the wheel, HotPlex leverages existing, powerful AI CLI agents and bridges them into production-ready systems by converting them into long-lived, interactive services (**Cli-as-a-Service**).

### Key Benefits

- **Zero Spin-up Overhead**: Eliminate the multi-second latency of restarting CLI environments
- **Hot-Multiplexing**: Persistent sessions with continuous instruction flow and context preservation
- **Unified Integration Layer**: Whether building professional AI products or automated pipelines, HotPlex makes elite agent capabilities as easy to call as a standard API
- **Production-Ready Security**: Command-level WAF and PGID isolation provide a "protective glove" for AI shell operations

## Installation

### Go Module

```bash
go get github.com/hrygo/hotplex
```

### Requirements

- Go 1.24+
- AI CLI tool installed (Claude Code or OpenCode)

## Architecture Overview

HotPlex decouples the **access layer** from the **execution engine layer**, leveraging bounded Go channels and WaitGroups to achieve deterministic, safe concurrent I/O handling at scale.

### Components

| Component | Description |
|-----------|-------------|
| **Access Layer** | Native Go SDK calls or remote API connections (`hotplexd`). Supports WebSocket and OpenCode-compatible protocols. |
| **Engine Layer** | Singleton resource manager managing the session pool, configuration overrides, and security WAF. |
| **Process Layer** | Sub-process worker isolated in PGID-level workspaces, locked to specific directory boundaries. |

## Next Steps

- [Quick Start](/guide/quick-start) - Build your first HotPlex application
- [Architecture](/guide/architecture) - Learn about the system design
- [Security](/guide/security) - Understand the security model
- [Go SDK](/sdks/go-sdk) - Detailed SDK documentation
