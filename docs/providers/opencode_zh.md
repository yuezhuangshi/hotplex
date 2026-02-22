# OpenCode Provider 集成指南

*查看其他语言: [English](opencode.md), [简体中文](opencode_zh.md).*

本文档介绍了如何在 HotPlex 生态系统中集成和使用 **OpenCode** CLI 智能体。

## 概述

与针对交互式 REPL 风格（Stdin 通讯）优化的 Claude Code 不同，**OpenCode** 采用了“服务器优先”的架构设计。它提供了多种集成方式：
1. **CLI 模式**：通过执行 `opencode run` 处理单次任务。
2. **服务器模式**：将 `opencode serve` 作为持久化后台进程运行。
3. **ACP 协议**：使用 **Agent Client Protocol (智能体客户端协议)** 进行结构化的多轮交互。

HotPlex 利用这些模式提供亚秒级的响应速度和持久化的会话管理。

## 安装

确保已全局安装 OpenCode：

```bash
npm install -g opencode
```

验证安装：
```bash
opencode --version
```

## 配置

可以通过 `hotplex.ProviderConfig` 配置 `OpenCodeProvider`。

### Provider 配置参数

| 参数           | 类型     | 说明                                                        |
| :------------- | :------- | :---------------------------------------------------------- |
| `use_http_api` | `bool`   | 如果为 true，HotPlex 将与后台的 `opencode serve` 实例通信。 |
| `port`         | `int`    | OpenCode 服务器端口（默认：4096）。                         |
| `plan_mode`    | `bool`   | 开启 OpenCode 的规划模式（对只读操作自动授权）。            |
| `provider`     | `string` | LLM 供应商（如 `openai`, `anthropic`, `siliconflow`）。     |
| `model`        | `string` | 具体模型 ID（如 `zhipu/glm-5-code-plan`, `gpt-4o`）。       |

## 架构选型

### 1. CLI 模式 (默认)
在此模式下，HotPlex 为每个请求执行 `opencode run`。
- **优点**：简单，无需维护后台服务器。
- **缺点**：延迟较高（冷启动），因为 CLI 必须在每轮对话时重新初始化。

### 2. HTTP 服务器模式 (推荐)
HotPlex 启动并管理一个 `opencode serve` 进程。请求通过 HTTP API 发送，事件通过 SSE (Server-Sent Events) 捕获。
- **优点**：亚秒级延迟，多轮对话状态更稳健。
- **缺点**：需要管理后台进程资源。

## Go 语言使用示例

```go
package main

import (
	"context"
	"fmt"
	"github.com/hrygo/hotplex"
)

func main() {
	// 1. 初始化 OpenCode Provider
	opencodePrv, _ := hotplex.NewOpenCodeProvider(hotplex.ProviderConfig{
		Type:         hotplex.ProviderTypeOpenCode,
		DefaultModel: "zhipu/glm-5-code-plan", // 推荐用于代码规划任务
		OpenCode: &hotplex.OpenCodeConfig{
			PlanMode:   true,  // 开启规划模式
			UseHTTPAPI: true,  // 使用服务器模式以获得低延迟
		},
	}, nil)

	// 2. 注入引擎
	engine, _ := hotplex.NewEngine(hotplex.EngineOptions{
		Provider: opencodePrv,
	})
	defer engine.Close()

	// 3. 执行任务
	ctx := context.Background()
	engine.Execute(ctx, &hotplex.Config{SessionID: "my-task"}, "分析当前仓库", 
		func(eventType string, data any) error {
			if eventType == "answer" {
				fmt.Print(data.(*hotplex.EventWithMeta).EventData)
			}
			return nil
		})
}
```

## 协议映射

OpenCode 事件会自动映射到 HotPlex 的统一事件模型中：

| OpenCode Part 类型 | HotPlex 事件类型 |
| :----------------- | :--------------- |
| `text`             | `answer`         |
| `reasoning`        | `thinking`       |
| `tool` (input)     | `tool_use`       |
| `tool` (output)    | `tool_result`    |
| `step-start`       | `step_start`     |
| `step-finish`      | `step_finish`    |

## 进阶集成模式 (最新最佳实践)

### 1. 智能体客户端协议 (ACP)
对于深度集成（IDE 或自定义控制平面），推荐使用 `opencode acp`。
- **通信方式**：通过 `stdin`/`stdout` 进行 JSON-RPC 交互。
- **能力**：允许客户端精细控制工具授权并观察智能体详细的推理链。
- **延迟**：持久化的进程模式消除了冷启动延迟。

### 2. 仓库级上下文 (`AGENTS.md`)
建议在项目根目录首次运行 `opencode run "/init"`。这将生成：
- **`AGENTS.md`**：AI 用于理解项目结构的概览图。
- **`opencode.json`**：规则定义和模型偏好。
HotPlex 可以动态监测或向这些文件中注入规则，以实现更智能的编排。

### 3. MCP (Model Context Protocol) 支持
OpenCode 支持 MCP 服务器。您可以通过在配置中添加 MCP 服务器来扩展 OpenCode 的能力，使其能够与外部数据库、API 或自定义企业工具进行交互。

### 4. 会话持久化与状态相关
OpenCode 将所有对话历史存储在 `~/.local/share/opencode/opencode.db` 中。
- 配合 `opencode run` 使用 `--session <id>` 参数可以恢复特定任务。
- HotPlex 可以通过检查此数据库来提供跨会话的状态审计与分析。
