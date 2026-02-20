# 🔥 hotplex

<p align="center">
  <a href="https://github.com/hrygo/hotplex/actions/workflows/ci.yml"><img src="https://github.com/hrygo/hotplex/actions/workflows/ci.yml/badge.svg" alt="Build Status"></a>
  <a href="https://github.com/hrygo/hotplex/releases"><img src="https://img.shields.io/github/v/release/hrygo/hotplex?style=flat-square" alt="Latest Release"></a>
  <a href="https://pkg.go.dev/github.com/hrygo/hotplex"><img src="https://img.shields.io/badge/go-reference-007d9c?style=flat-square&logo=go" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/hrygo/hotplex"><img src="https://goreportcard.com/badge/github.com/hrygo/hotplex?style=flat-square" alt="Go Report Card"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/hrygo/hotplex?style=flat-square" alt="License"></a>
</p>

> **从 5000ms 🐢 锐减至 200ms 🚀。hotplex 让你的 AI 智能体时刻保持“热启动”。**

*Read this in other languages: [English](README.md), [简体中文](README_zh.md).*

**hotplex** 是一个高性能的**进程多路复用器 (Process Multiplexer)**，专为在长生命周期的服务器或 Web 环境中运行繁重的本地 AI CLI 代理（如 Claude Code、OpenCode、Aider）而设计。

它通过在后台保持繁重的 Node.js 或 Python CLI 进程常驻，并将并发的请求流（热重载多路复用）受控地路由到它们的标准输入/输出 (Stdin/Stdout) 管道中，从而彻底解决了“冷启动”问题。

## 🚀 为什么选用 hotplex？

通常，从后端服务（如 Go API）运行本地 CLI 代理意味着必须为**每一次**交互都生成一个全新的操作系统进程。

*   **问题所在：** 像 `claude` (Claude Code) 这样的工具是重量级的 Node.js 应用。每次执行 `npx @anthropic-ai/claude-code` 都需要花费 **3-5 秒**，仅仅是为了启动 V8 引擎、读取文件系统上下文以及进行身份验证。对于实时的 Web UI 而言，这种延迟会让智能体感觉极其缓慢且不响应。
*   **解决方案：** hotplex 为每个用户/会话仅仅启动一次 CLI 进程，将其保存在后台（被包裹在安全的进程组 `pgid` 中），并建立持久的双向管道。当用户发送新消息时，hotplex 会立即通过 `Stdin` 注入消息，并通过 `Stdout` 将 JSON 响应流式传回。延迟从 **5000 毫秒锐减至 200 毫秒以内**。

## 💡 愿景与应用场景

hotplex 的核心价值在于**赋能 AI 应用程序轻松集成强大的 CLI 代理**（如 **Claude Code**, **Aider**, 或 **OpenCode**），使其成为 AI 的外部“肌肉”。通过 hotplex，你可以获得亚秒级的响应速度，无需重复造轮子。

*   🌐 **Web 版 AI 智能体**: 为 Claude Code 构建功能完备的 Web UI。用户通过浏览器交互，hotplex 在后端管理持久的沙盒化 CLI 进程。
*   🔧 **DevOps 工具链自动化**: 将 AI 直接集成到运维流程中。让智能体通过持久会话执行脚本、分析日志并排查故障。
*   🚀 **CI/CD 智能集成**: 在流水线中引入智能代码审查与漏洞修复，彻底免去笨重 Node.js 工具的启动延迟。
*   🕵️ **智能 AIOps**: 打造运维机器人，持续监控系统并通过受控的终端会话自主执行安全修复指令。

## ✨ 特性概览

*   **⚡ 极速热启动：** 初始启动后，后续指令响应达到毫秒级 (~200ms)。
*   **♻️ 会话池自动回收 (GC)：** 自动跟踪空闲进程并清理释放资源。
*   **🛡️ 原生工具约束：** 在引擎层级通过 CLI 原生参数硬性限制智能体能力（例如禁用 `Bash` 或网络权限）。
*   **🔌 WebSocket 网关：** 包含 `hotplexd` 独立服务器，原生支持 WebSockets 接入。
*   **📦 原生 Go SDK：** 提供高层次 Go API，方便直接嵌入到您的后端服务中。
*   **🔥 安全性防火墙：** 内置 `danger.go` 拦截器，在指令下发前阻断破坏性命令（如 `rm -rf /`）。
*   **🔒 上下文隔离：** 基于 PGID 的进程隔离与确定性的 UUID v5 会话空间。

## 📦 架构设计

hotplex 采用两层架构设计：

![hotplex Architecture](docs/images/topology.svg)

1.  **核心 SDK (`pkg/hotplex`)**：引擎本体，提供 `Engine` 单例、`SessionPool` 和 `Detector`（安全防火墙）。它接收 CLI 的 JSON 流并触发强类型的 Go 事件回调。
2.  **独立服务端 (`cmd/hotplexd`)**：基于 SDK 封装的轻量级 WebSocket 服务器。

#### 🌊 异步事件流机制

hotplex 充分利用 Go 的并发特性实现真正的全双工流式交互：

![hotplex Event Flow](docs/images/async-stream.svg)

*注意：当前 MVP 版本针对 **Claude Code** 的全双工 JSON 协议（同时开启 `--input-format stream-json` 与 `--output-format stream-json`）进行了深度优化，但设计上已预留 `Provider` 接口以支持 OpenCode 和 Aider。*

## ⚡ 快速开始

### 1. 运行 WebSocket 独立服务器

如果你想直接运行服务器并通过前端或脚本连接：

```bash
# 安装 Claude Code (推荐：官方原生安装脚本)
# macOS / Linux / WSL:
curl -fsSL https://claude.ai/install.sh | bash

# 或通过 Homebrew 安装:
brew install claude-code

# 或通过 NPM 安装 (传统方式):
npm install -g @anthropic-ai/claude-code

# 编译并运行守护进程
cd cmd/hotplexd
go build -o hotplexd main.go
./hotplexd
```
服务器运行在 `ws://localhost:8080/ws/v1/agent`。参考 `_examples/websocket_client/client.js` 查看集成示例。

### 2. 使用 Go SDK 原生集成

安装依赖包：
```bash
go get github.com/hrygo/hotplex
```

并在代码中引入：
```go
import "github.com/hrygo/hotplex/pkg/hotplex"

opts := hotplex.EngineOptions{
    Timeout: 5 * time.Minute,
    Logger:  logger,
    PermissionMode: "bypass-permissions", // v0.2.0+ 推荐的权限处理模式
    AllowedTools: []string{"Bash", "Edit"}, // 在引擎层级限制工具能力
}

engine, _ := hotplex.NewEngine(opts)
defer engine.Close()

cfg := &hotplex.Config{
    WorkDir:          "/tmp/sandbox",
    SessionID:        "user_123_session", // 确定性多路复用 ID
    TaskSystemPrompt: "你是一名资深开发工程师。",
}

ctx := context.Background()

// 1. 发送 Prompt 并处理流式回调
err := engine.Execute(ctx, cfg, "重构 main.go 文件", func(eventType string, data any) error {
    if eventType == "answer" {
         fmt.Println("智能体正在响应...")
    }
    return nil
})
```

## 🔒 安全性设计

hotplex 致力于在执行 LLM 生成的代码时提供多重安全保障：

1.  **🛡️ 能力治理**：优先使用原生工具限制 (`AllowedTools`)，确保智能体仅拥有最小权限。
2.  **🔥 指令预检 (WAF)**：正则防御层级，实时拦截 `mkfs`, `dd`, `rm -rf /` 等危险指令。
3.  **⚰️ 进程组清理 (PGID)**：会话终止时对整个进程组发送 `SIGKILL`，杜绝孤儿进程残留。
4.  **🏗️ 目录锁定**：严格限制智能体在指定的 `WorkDir` 路径内活动。

---

## 🗺️ 路线图 (Roadmap)
- [ ] 提取 Provider 接口 (增加对 `Aider` 的通用支持)
- [ ] 远程 Docker 沙盒执行驱动
- [ ] 提供管理 REST API，用于会话自省与管理
- [ ] 集成 [Firebase Genkit](https://firebase.google.com/docs/genkit) (Go SDK)

## 👋 参与贡献
欢迎任何形式的贡献！请阅读 [CONTRIBUTING.md](CONTRIBUTING.md) 了解开发指引。

## 📄 开源协议
hotplex 基于 [MIT 协议](LICENSE) 开源。
