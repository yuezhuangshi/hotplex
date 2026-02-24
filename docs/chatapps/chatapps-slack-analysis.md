# Slack 双向全双工通讯技术方案

> 本文档深入分析 HotPlex Slack 适配器的技术实现方案，包括 API 调用、认证机制、消息格式等细节，并对比 OpenClaw 最佳实践给出改进建议。

---

## 1. 架构概述

### 1.1 通讯模型

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Slack 双向通讯架构                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│   用户 ──发送消息──▶ Slack 服务器 ──HTTP POST──▶ HotPlex           │
│                        (Events API)           (handleEvent)           │
│                                                                      │
│   用户 ◀──回复消息── Slack 服务器 ◀──HTTP POST── HotPlex           │
│                        (Web API)            (chat.postMessage)       │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### 1.2 当前实现状态

| 功能 | 状态 | 说明 |
|------|------|------|
| 消息接收 (Webhook) | ✅ 已实现 | Events API HTTP Webhook |
| 消息发送 (API) | ✅ 已实现 | chat.postMessage |
| 签名验证 | ✅ 已实现 | HMAC-SHA256 |
| Session 管理 | ✅ 已实现 | Channel:User 组合 |
| Socket Mode | ❌ 待实现 | WebSocket 实时连接 |
| 消息分片 | ❌ 待实现 | 4000 字符自动分片 |
| 线程支持 | ⚠️ 待完善 | 基础 thread_ts 存储 |
| Rate Limit | ❌ 待实现 | 指数退避重试机制 |

### 1.3 OpenClaw 参考架构

OpenClaw 采用更完善的架构设计，支持以下特性：

```
src/slack/
├── monitor/              # 事件接收 (Bolt.js)
│   ├── provider.ts     # 主入口: 支持 socket/http 双模式
│   └── events/          # 事件处理
├── send.ts              # 消息发送 (含分片)
├── streaming.ts          # Slack 原生流式消息
├── draft-stream.ts       # 草稿流 (实时预览)
├── threading.ts         # 线程管理
├── client.ts            # Slack WebClient (含重试)
└── format.ts            # Markdown → mrkdwn 转换
```

---

## 2. API 端点分析

### 2.1 当前使用的 API

| 用途 | API 端点 | 认证方式 | 代码位置 |
|------|---------|---------|----------|
| 发送消息 | `POST /api/chat.postMessage` | Bot Token (xoxb-) | slack/adapter.go:243 |
| 验证签名 | - | HMAC-SHA256 | slack/adapter.go:221 |

### 2.2 消息发送实现

```go
// 代码位置: slack/adapter.go:243-271
func (a *Adapter) SendToChannel(ctx context.Context, channelID, text string) error {
    payload := map[string]any{"channel": channelID, "text": text}
    body, err := json.Marshal(payload)
    
    req, err := http.NewRequestWithContext(ctx, "POST", 
        "https://slack.com/api/chat.postMessage", bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+a.config.BotToken)
    
    resp, err := http.DefaultClient.Do(req)
    // 处理响应
}
```

**请求头**:
```
Content-Type: application/json
Authorization: Bearer xoxb-xxx-xxx
```

### 2.3 OpenClaw 消息分片实现

```typescript
// OpenClaw send.ts
const SLACK_TEXT_LIMIT = 4000;

// 自动分片发送
for (const chunk of chunks) {
    await client.chat.postMessage({
        channel: channelID,
        text: chunk,
        thread_ts: threadTS,  // 保持线程
    });
}
```

**关键点**:
- 单条消息限制 4000 字符
- 自动分片后保持线程上下文
- 支持 markdown 到 mrkdwn 格式转换

---

## 3. 消息接收 (Webhook)

### 3.1 Events API 回调处理

```go
// 代码位置: slack/adapter.go:115-163
func (a *Adapter) handleEvent(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    
    // 1. 签名验证
    if a.config.SigningSecret != "" {
        signature := r.Header.Get("X-Slack-Signature")
        timestamp := r.Header.Get("X-Slack-Request-Timestamp")
        if !a.verifySignature(body, timestamp, signature) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
    }
    
    // 2. 解析事件
    var event Event
    json.Unmarshal(body, &event)
    
    // 3. URL 验证 (Slack 首次配置时)
    if event.Challenge != "" {
        w.Write([]byte(event.Challenge))
        return
    }
    
    // 4. 处理消息事件
    if event.Type == "event_callback" {
        a.handleEventCallback(r.Context(), event.Event)
    }
}
```

### 3.2 签名验证算法

```go
// 代码位置: slack/adapter.go:221-241
func (a *Adapter) verifySignature(body []byte, timestamp, signature string) bool {
    // 1. 检查时间戳 (5 分钟内有效)
    parsedTS := strings.TrimPrefix(timestamp, "v0=")
    if now-ts > 60*5 {
        return false
    }
    
    // 2. 构建签名 base string
    baseString := fmt.Sprintf("v0:%s:%s", parsedTS, string(body))
    
    // 3. HMAC-SHA256 计算
    h := hmac.New(sha256.New, []byte(a.config.SigningSecret))
    h.Write([]byte(baseString))
    signatureComputed := "v0=" + hex.EncodeToString(h.Sum(nil))
    
    // 4. 常数时间比较
    return hmac.Equal([]byte(signatureComputed), []byte(signature))
}
```

**签名算法**:
1. 提取时间戳: `v0:timestamp:body`
2. 使用 SigningSecret 作为密钥进行 HMAC-SHA256 签名
3. 添加 `v0=` 前缀
4. 使用 `hmac.Equal` 常数时间比较防止 timing attack

### 3.3 消息解析

```go
// 代码位置: slack/adapter.go:165-196
type MessageEvent struct {
    Type        string `json:"type"`
    Channel     string `json:"channel"`
    ChannelType string `json:"channel_type"`
    User        string `json:"user"`
    Text        string `json:"text"`
    TS          string `json:"ts"`
    EventTS     string `json:"event_ts"`
    BotID       string `json:"bot_id,omitempty"`
    SubType     string `json:"subtype,omitempty"`
}

// Session 创建
sessionID := a.GetOrCreateSession(msgEvent.Channel+":"+msgEvent.User, msgEvent.User)

// 消息构建
msg := &base.ChatMessage{
    Platform:  "slack",
    SessionID: sessionID,
    UserID:    msgEvent.User,
    Content:   msgEvent.Text,
    MessageID: msgEvent.TS,
    Metadata: map[string]any{
        "channel_id":   msgEvent.Channel,
        "channel_type": msgEvent.ChannelType,
    },
}
```

**关键字段映射**:

| Slack 事件字段 | 用途 | 保存位置 |
|---------------|------|---------|
| `user` | 用户 ID | `msg.UserID` |
| `channel` | 频道 ID | `msg.Metadata.channel_id` |
| `text` | 消息内容 | `msg.Content` |
| `ts` | 消息时间戳 | `msg.MessageID` |

---

## 4. Session 管理

### 4.1 Session 结构

```go
// 代码位置: chatapps/base/adapter.go
type Session struct {
    SessionID  string
    UserID     string
    Platform   string
    LastActive time.Time
}
```

### 4.2 Session 映射

```go
// 代码位置: slack/adapter.go:180
sessionID := a.GetOrCreateSession(msgEvent.Channel+":"+msgEvent.User, msgEvent.User)
```

**Session Key 规则**: `{channelId}:{userId}`

### 4.3 过期清理

```go
// 代码位置: chatapps/base/adapter.go:309-329
func (a *Adapter) cleanupSessions() {
    ticker := time.NewTicker(5 * time.Minute)
    for {
        select {
        case <-a.cleanupDone:
            return
        case <-ticker.C:
            // 清理超过 sessionTimeout 的会话
        }
    }
}
```

- 清理周期: 5 分钟
- 会话超时: 30 分钟

### 4.4 OpenClaw 线程上下文解析

```typescript
// OpenClaw threading.ts
function resolveSlackThreadContext(params) {
    const { thread_ts, ts, event_ts, parent_user_id } = params.message;
    
    // 判断是否为线程回复
    const isThreadReply = thread_ts && (thread_ts !== ts || parent_user_id);
    
    return {
        isThreadReply,
        replyToId: thread_ts ?? ts,
        messageThreadId: isThreadReply ? thread_ts : undefined,
    };
}
```

---

## 5. Socket Mode vs HTTP Webhook

### 5.1 两种模式对比

| 模式 | 连接方式 | 公网 URL | 延迟 | 适用场景 |
|------|---------|---------|------|---------|
| **HTTP Webhook** | HTTP POST | 需要 | 较高 | 公共 Slack App |
| **Socket Mode** | WebSocket | 不需要 | 低 | 内部工具/开发 |

### 5.2 OpenClaw Socket Mode 实现

```typescript
// OpenClaw monitor/provider.ts
const slackMode = opts.mode ?? account.config.mode ?? "socket";

const app = new App(
  slackMode === "socket"
    ? {
        token: botToken,
        appToken,           // xapp-* token (必需)
        socketMode: true,  // WebSocket 模式
      }
    : {
        token: botToken,
        receiver: new HTTPReceiver({ signingSecret }),
      }
);

// 启动
if (slackMode === "socket") {
    await app.start();
}
```

### 5.3 建议实现

```go
type Config struct {
    BotToken      string  // xoxb-* (必需)
    AppToken     string  // xapp-* (Socket Mode 需要)
    SigningSecret string  // HTTP Mode 需要
    Mode          string  // "socket" | "http"
}
```

---

## 6. 消息分片机制

### 6.1 Slack 限制

- 单条消息最大: 4000 字符
- 消息块最大: 50 个
- 总消息大小: 100KB

### 6.2 OpenClaw 分片策略

```typescript
// OpenClaw send.ts
const SLACK_TEXT_LIMIT = 4000;

// 1. Markdown 转换为 mrkdwn
const chunks = markdownToSlackMrkdwnChunks(markdown, chunkLimit);

// 2. 分片发送
for (const chunk of chunks) {
    await client.chat.postMessage({
        channel: channelID,
        text: chunk,
        thread_ts: threadTS,  // 保持线程
    });
}
```

### 6.3 建议实现

```go
const SlackTextLimit = 4000

func (a *Adapter) sendWithChunking(ctx context.Context, channelID, text string, threadTS string) error {
    chunks := chunkMessage(text, SlackTextLimit)
    
    for i, chunk := range chunks {
        // 添加分片编号
        msg := fmt.Sprintf("[%d/%d]\n%s", i+1, len(chunks), chunk)
        
        err := a.SendToChannel(ctx, channelID, msg)
        if err != nil {
            return err
        }
    }
    return nil
}
```

---

## 7. 认证机制详解

### 7.1 Token 类型

| Token | 前缀 | 用途 |
|-------|------|------|
| Bot Token | `xoxb-` | API 调用 (发消息、读频道) |
| User Token | `xoxp-` | 用户权限操作 |
| App Token | `xapp-` | Socket Mode WebSocket 连接 |

### 7.2 配置参数

```go
type Config struct {
    BotToken       string  // Bot Token (xoxb-*)
    AppToken      string  // App Token (xapp-*) for Socket Mode
    SigningSecret string  // Signing Secret for HTTP Mode
    ServerAddr    string  // HTTP 服务地址
    SystemPrompt  string  // 系统提示词
}
```

### 7.3 OpenClaw SDK 内置重试

```typescript
// OpenClaw client.ts
const SLACK_DEFAULT_RETRY_OPTIONS = {
    retries: 2,
    factor: 2,
    minTimeout: 500,
    maxTimeout: 3000,
    randomize: true,
};
```

---

## 8. Slack 原生流式消息

### 8.1 ChatStreamer API

OpenClaw 使用 Slack 最新的 ChatStreamer API 实现实时流式输出：

```typescript
// OpenClaw streaming.ts
const streamer = client.chatStream({
    channel,
    thread_ts: threadTs,
});

// 实时追加
await streamer.append({ markdown_text: text });

// 完成
await streamer.stop({ markdown_text: finalText });
```

### 8.2 Draft Stream (草稿流)

```typescript
// OpenClaw draft-stream.ts
const stream = createSlackDraftStream({
    target: channelID,
    token: botToken,
    maxChars: 4000,
    throttleMs: 1000,
});

// 实时更新 (节流 1 秒)
stream.update("正在输入...");

// 发送完成
await stream.flush();
```

---

## 9. Markdown → mrkdwn 转换

### 9.1 格式差异

| Markdown | mrkdwn |
|----------|--------|
| `**粗体**` | `*粗体*` |
| `*斜体*` | `_斜体_` |
| `[链接](url)` | `<url\|文本>` |
| `` `代码` `` | `` `代码` `` |
| ```代码块``` | ```代码块``` |

### 9.2 OpenClaw 转换实现

```typescript
// OpenClaw format.ts
function escapeSlackMrkdwnText(text: string): string {
    return text
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;");
}

function buildSlackLink(link, text) {
    return { open: `<${href}|`, close: ">" };
}
```

---

## 10. 潜在问题与改进建议

### 10.1 当前限制

| 问题 | 严重程度 | 说明 |
|------|---------|------|
| 仅支持 HTTP Webhook | 高 | 缺少 Socket Mode 低延迟连接 |
| 无消息分片机制 | 高 | 超过 4000 字符会发送失败 |
| 线程支持不完整 | 中 | 未解析 thread_ts 上下文 |
| 无 Rate Limit 处理 | 中 | 429 错误无退避重试 |
| 无 Markdown 转换 | 低 | 纯文本发送 |

### 10.2 改进建议

#### Phase 1: 基础能力提升

1. **Socket Mode 支持**
   - 添加 AppToken 配置
   - 实现 WebSocket 连接管理
   - 支持连接断开重连

2. **消息分片**
   - 实现 4000 字符自动分片
   - 保持线程上下文
   - 添加分片编号

3. **线程支持完善**
   - 解析 thread_ts 字段
   - 回复到正确线程
   - 支持 reply_broadcast

#### Phase 2: 体验优化

4. **Rate Limit 处理**
   - 实现指数退避
   - 监听 Retry-After 头
   - 添加重试队列

5. **Markdown 支持**
   - 实现 mrkdwn 转换
   - 支持代码块高亮
   - 链接格式转换

#### Phase 3: 高级功能

6. **流式消息**
   - 实现 Draft Stream
   - 实时预览输入
   - 节流更新

7. **Block Kit**
   - 支持富文本消息
   - 按钮交互
   - 内联图片

---

## 11. 测试验证清单

### 11.1 配置验证

- [ ] Bot Token (xoxb-*) 正确配置
- [ ] App Token (xapp-*) 配置 (Socket Mode)
- [ ] Signing Secret 配置 (HTTP Mode)
- [ ] Slack App 已安装到工作区

### 11.2 功能验证

- [ ] 接收文本消息正常
- [ ] 发送文本消息正常
- [ ] 发送 Markdown 消息正常
- [ ] 长消息自动分片
- [ ] 线程消息正确回复
- [ ] Session 隔离正确

### 11.3 Socket Mode 验证

- [ ] WebSocket 连接建立
- [ ] 消息实时接收
- [ ] 连接断开重连
- [ ] 低延迟消息收发

### 11.4 异常处理

- [ ] 签名验证失败处理
- [ ] 429 Rate Limit 处理
- [ ] Token 过期处理
- [ ] 网络错误重试

---

## 12. 参考资料

- [Slack API 文档](https://api.slack.com/)
- [Slack Events API](https://api.slack.com/apis/connectors)
- [Socket Mode](https://api.slack.com/apis/socket-mode)
- [Bolt.js 框架](https://github.com/slackapi/bolt-js)
- [OpenClaw Slack 源码](https://github.com/openclaw/openclaw/tree/main/src/slack)
- [ChatStreamer API](https://docs.slack.dev/ai/developing-ai-apps#streaming)

---

*本文档最后更新: 2026-02-24*
