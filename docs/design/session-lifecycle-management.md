# HotPlex Session 生命周期管理机制

> 完整的 Session 创建、恢复、清理机制详解

## 核心设计原则

### 端到端一致性映射

```
SessionID ──SHA1──▶ ProviderSessionID (永远确定性，可追溯)

❌ 禁止: Random UUID, resetSessions 机制
✅ 始终: SHA1(sessionID) → ProviderSessionID
```

**关键点**：
- `sessionID` = 逻辑会话 ID（如 `slack:U12345:C67890:Ts12345`）
- `providerSessionID` = CLI 实际使用的 UUID（永远通过 SHA1 生成）
- **绝对不存在 random UUID**，确保端到端可追溯

---

## 一、核心概念与存储结构

### 两种持久化文件

| 文件类型 | 路径 | 用途 | 管理者 |
|---------|------|------|--------|
| **Marker 文件** | `~/.hotplex/sessions/<uuid>.lock` | 标记会话可恢复 | SessionPool (persistence/) |
| **CLI Session 文件** | `~/.claude/projects/<workspace>/<uuid>.jsonl` | 实际对话历史和状态 | CLI Provider (Claude Code) |

Marker 文件是空文件，仅作为"此会话曾经成功启动"的标记。CLI Session 文件包含完整的对话上下文。

---

## 二、Session 生命周期状态机

```
                              用户请求到达
                                   │
                                   ▼
                        ┌─────────────────────┐
                        │ GetOrCreateSession  │
                        └──────────┬──────────┘
                                   │
              ┌────────────────────┼────────────────────┐
              │                    │                    │
              ▼                    │                    ▼
   ┌──────────────────┐            │          ┌──────────────────┐
   │ 内存中存在活跃会话? │            │          │  内存中不存在/已死 │
   │ IsAlive() == true │            │          │                  │
   └────────┬─────────┘            │          └────────┬─────────┘
            │                      │                   │
      是    │                      │                   │ 否
   ┌────────┴────────┐             │         ┌─────────┴─────────┐
   │ Touch() 更新时间戳│             │         │    startSession   │
   │ 直接返回现有会话  │             │         │    (Cold Start)   │
   └──────────────────┘             │         └─────────┬─────────┘
                                    │                   │
                                    │                   ▼
                                    │    ┌──────────────────────────────┐
                                    │    │  1. 生成 ProviderSessionID    │
                                    │    │     SHA1(namespace:session:ID)│
                                    │    └──────────────┬───────────────┘
                                    │                   │
                                    │                   ▼
                                    │    ┌──────────────────────────────┐
                                    │    │  2. 检查 Marker 是否存在      │
                                    │    │    markerStore.Exists()      │
                                    │    └──────────────┬───────────────┘
                                    │                   │
                                    │                   ▼
                                    │    ┌──────────────────────────────┐
                                    │    │  3. 构建 CLI 参数             │
                                    │    │ buildCLIArgs()               │
                                    │    └──────────────┬───────────────┘
                                    │                   │
                                    │                   ▼
                                    │    ┌──────────────────────────────┐
                                    │    │  4. 启动 CLI 进程             │
                                    │    │ cmd.Start() 成功后创建 Marker│
                                    │    └──────────────────────────────┘
```

---

## 三、Resume vs New Session 决策流程

### buildCLIArgs() 决策逻辑

```
markerStore.Exists(providerSessionID) ?
│
├─ YES
│   │
│   └─ VerifySession(providerSessionID, workDir) ?
│       │
│       ├─ YES (CLI 文件存在)
│       │   └─ --resume <uuid>
│       │   （不删除任何文件，直接恢复）
│       │
│       └─ NO (CLI 文件丢失)
│           ├─ markerStore.Delete(providerSessionID)
│           ├─ CleanupSession(providerSessionID)  // 清理僵尸文件
│           └─ --session-id <uuid>  (同样ID，全新会话)
│
└─ NO
    ├─ CleanupSession(providerSessionID)  // 清理可能存在的僵尸文件
    └─ --session-id <uuid>
    （Marker 在 cmd.Start() 成功后创建）
```

### 关键点

1. **Resume 场景**：Marker 存在 + CLI 文件存在 → 直接 `--resume`，不删除任何文件
2. **僵尸 Marker 场景**：Marker 存在 + CLI 文件丢失 → 删除 Marker，创建全新会话（同样 ID）
3. **新会话场景**：Marker 不存在 → 清理可能的僵尸 CLI 文件，创建新会话

---

## 四、Slash Command 行为

### 命令定义

| 命令 | Slack 别名 | 语义 | 执行操作 | 下次访问行为 |
|------|-----------|------|---------|-------------|
| `/reset` | `#reset` | 清空上下文，重新开始 | 1. StopSession()<br>2. DeleteMarker()<br>3. CleanupSession() | 同样 SHA1 ID，全新空会话 |
| `/dc` | `#dc` | 断开连接，保留上下文 | 1. StopSession() | 同样 SHA1 ID，resume 现有会话 |

### /reset 命令详解

```
/reset = 清空上下文，保持 ID 一致性

执行操作:
1. engine.StopSession(sessionID)              // 终止进程
2. pool.DeleteMarker(providerSessionID)       // 删除 marker
3. provider.CleanupSession(providerSessionID) // 删除 CLI 文件

下次访问:
- markerStore.Exists() → NO
- 使用同样的 SHA1(sessionID) → providerSessionID
- 创建全新空会话，ID 保持不变

用户看到的效果:
✅ 完全清空对话历史
✅ 会话 ID 不变（端到端可追溯）
```

### /dc (Disconnect) 命令详解

```
/dc = 断开连接，保留上下文供后续恢复

执行操作:
1. engine.StopSession(sessionID, "user_requested_disconnect")  // 仅终止进程

保留:
- Marker 文件 ✓
- CLI Session 文件 ✓

下次访问:
- markerStore.Exists() → YES
- VerifySession() → YES (CLI 文件存在)
- --resume <uuid> → 恢复现有会话上下文

用户看到的效果:
✅ 进程断开，节省资源
✅ 下次发消息自动恢复上下文
✅ 对话历史完整保留
```

### 命令对比图

```
                    /reset                          /dc
                 ─────────────────────────────────────────────────────
    Marker      │  删除                            保留
    CLI 文件    │  删除                            保留
    进程        │  终止                            终止
                 ─────────────────────────────────────────────────────
    下次访问    │  marker不存在                    marker存在
               │  → 同样ID, 全新会话              → resume现有会话
```

**关键点**：两个命令都**不改变 ID 映射**，始终保持 `SHA1(sessionID) → providerSessionID`。

---

## 五、异常处理与恢复

### Resume 失败（CLI 进程意外退出）

当 `cmd.Wait()` 检测到 Resume 的进程意外退出时：

```go
if isResuming && sess.GetStatus() != SessionStatusDead {
    // 1. 删除 Marker
    markerStore.Delete(providerSessionID)

    // 2. 删除 CLI session 文件
    provider.CleanupSession(providerSessionID, workDir)

    // 3. 不标记 random UUID
    // 下次请求将用同样的 SHA1 ID 创建全新会话
}
```

**结果**：下次请求 → 同样的 SHA1 ID → 全新空会话

---

## 六、不存在的机制

### ❌ 已删除的 Random UUID 机制

旧代码存在 `resetSessions` map 和 `ResetProviderSessionID()` 方法，用于在某些场景下生成 random UUID。这是**错误的设计**，已被完全删除：

```go
// ❌ 已删除 - 错误设计
type SessionPool struct {
    resetSessions map[string]bool  // 死代码
}

func (sm *SessionPool) ResetProviderSessionID(sessionID string) {
    sm.resetSessions[sessionID] = true  // 死代码
}
```

### ❌ 已删除的基于时间的 Marker 清理

旧代码使用 `cleanupStaleMarkers()` 基于文件 age 清理 marker。这是**错误的设计**：

```go
// ❌ 已删除 - 错误设计
if fileAge > sm.timeout*2 {
    markerStore.Delete(sessionID)
}
```

**为什么删除**：用户可能长时间后继续会话，不应该基于时间强制清理。正确的清理发生在用户访问时（按需清理）。

---

## 七、文件系统布局

```
$HOME/
├── .hotplex/
│   └── sessions/                          ← Marker 文件目录
│       ├── 550e8400-e29b-41d4-a716-446655440000.lock
│       └── ...
│
└── .claude/
    └── projects/                          ← CLI Session 文件目录
        ├── -Users-hotplex-projects-myapp/
        │   ├── 550e8400-e29b-41d4-a716-446655440000.jsonl
        │   └── ...
        └── ...
```

---

## 八、防止 "No conversation found" 的机制

```
┌────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                                 │
│   第 1 层: VerifySession() - 创建/恢复前验证                                                    │
│   ────────────────────────────────────────────                                                  │
│   Marker 存在时, 先检查 CLI 文件是否真实存在                                                     │
│   不存在则删除 Marker, 创建全新会话（同样 ID）                                                    │
│                                                                                                 │
│   第 2 层: Marker 创建时机 - 防止僵尸 Marker                                                    │
│   ────────────────────────────────────────────                                                  │
│   Marker 在 cmd.Start() 成功后才创建                                                            │
│   失败的启动不会留下 Marker                                                                     │
│                                                                                                 │
│   第 3 层: Resume 失败自动清理                                                                  │
│   ────────────────────────────────────────────                                                  │
│   Resume 失败时自动删除 Marker 和 CLI 文件                                                      │
│   下次请求用同样的 SHA1 ID 创建全新会话                                                          │
│                                                                                                 │
└────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 九、相关代码文件

| 文件 | 职责 |
|------|------|
| `internal/engine/pool.go` | SessionPool 实现: 创建、恢复、GC |
| `internal/engine/session.go` | Session 实例: 进程管理、I/O |
| `internal/persistence/markers.go` | Marker 文件存储接口 |
| `provider/provider.go` | Provider 接口定义 |
| `provider/claude_provider.go` | Claude Code Provider 实现，包含 VerifySession |
| `chatapps/command/types.go` | 命令常量定义 (`/reset`, `/dc`) |
| `chatapps/command/reset_executor.go` | /reset 命令实现 |
| `chatapps/command/disconnect_executor.go` | /dc 命令实现 |
| `chatapps/slack/slash_commands.go` | Slack `#reset` / `#dc` 别名处理 |
| `chatapps/feishu/command_handler.go` | 飞书命令映射 |
