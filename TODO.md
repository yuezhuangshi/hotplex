# HotPlex Architecture Refactoring TODO

> Created: 2026-02-24  
> Last Updated: 2026-02-24 21:00  
> Status: ✅ All Tasks Completed

## Overview

Architecture review and SOLID/DRY compliance refactoring for HotPlex project.

---

## Completed Tasks ✅

### P1: SessionPool Persistence Extraction (Completed)
- Created `internal/persistence/markers.go` with `SessionMarkerStore` interface
- Created `FileMarkerStore` and `InMemoryMarkerStore` implementations
- Refactored `internal/engine/pool.go` to use `markerStore`

### P1-3: Migrate ClaudeCodeProvider to SessionMarkerStore (Completed 2026-02-24)
- Replaced `markerDir string` field with `markerStore persistence.SessionMarkerStore`
- Updated `NewClaudeCodeProvider` to use `persistence.NewDefaultFileMarkerStore()`
- Updated `GetMarkerDir()` to return `markerStore.Dir()`
- Updated `CheckSessionMarker()` to call `markerStore.Exists()`
- Updated `BuildCLIArgs` marker creation to use `markerStore.Create()`
- Removed `os` and `path/filepath` imports, added `internal/persistence` import
- **Build passes**: `go build ./...` ✅
- **Tests pass**: `go test ./...` ✅

### P3: Provider Interface Dead Code Removal (Completed)
- Removed unused methods from `provider/provider.go`: `BuildEnvVars`, `ExtractSessionID`, `GetVersion`
- Removed implementations from `provider/claude_provider.go` and `provider/opencode_provider.go`

### P4: Compile-Time Interface Check (Completed)
- Added `var _ HotPlexClient = (*Engine)(nil)` in `hotplex.go`

### P5: Security Suggestions Refactor (Completed)
- Converted `getSuggestions` switch statement to map lookup in `internal/security/detector.go`

### P2 (Partial): ChatApps Base Utilities Created
- Created `chatapps/base/sender.go` - `SenderWithMutex` utility
- Created `chatapps/base/http.go` - `HTTPClient` utility
- Created `chatapps/base/webhook.go` - `WebhookRunner` utility

### P2-4 (Completed 2026-02-24): ChatApp Adapters Refactored
- ✅ **slack/adapter.go** - Already using `*base.SenderWithMutex` and `*base.WebhookRunner`
- ✅ **discord/adapter.go** - Already using `*base.SenderWithMutex` and `*base.WebhookRunner`
- ✅ **telegram/adapter.go** - Refactored to use base utilities
- ✅ **whatsapp/adapter.go** - Refactored to use base utilities  
- ✅ **dingtalk/adapter.go** - Refactored to use base utilities
- **Build passes**: `go build ./...` ✅
- **Tests pass**: `go test ./...` ✅

---

## Pending Tasks 🚧

### P2-4: Refactor Remaining ChatApp Adapters (3 adapters left)

**Priority**: High  
**Estimated Effort**: 1-1.5 hours (3 adapters remaining)

**Description**: Refactor 3 remaining chatapp adapters to use new base utilities.

**Files to Modify** (NOT YET DONE):
- `/Users/huangzhonghui/HotPlex/chatapps/telegram/adapter.go`
- `/Users/huangzhonghui/HotPlex/chatapps/whatsapp/adapter.go`
- `/Users/huangzhonghui/HotPlex/chatapps/dingtalk/adapter.go`

**Adapters Status**:
| Adapter | Status | Notes |
|---------|--------|-------|
| slack | ✅ Done | Already using base utilities |
| discord | ✅ Done | Already using base utilities |
| telegram | 🚧 Pending | Uses old pattern: `sender func()`, `senderMu sync.RWMutex`, `webhookWg sync.WaitGroup` |
| whatsapp | 🚧 Pending | Uses old pattern: `sender func()`, `senderMu sync.RWMutex`, `webhookWg sync.WaitGroup` |
| dingtalk | 🚧 Pending | Uses old pattern: `sender func()`, `senderMu sync.RWMutex`, `webhookWg sync.WaitGroup` |

**Pattern to Replace (in telegram, whatsapp, dingtalk)**:

```go
// BEFORE - Current duplicate pattern:
type Adapter struct {
    *base.Adapter
    config      Config
    webhookPath string
    sender      func(ctx context.Context, sessionID string, msg *base.ChatMessage) error
    senderMu    sync.RWMutex   // Remove this
    webhookWg   sync.WaitGroup // Remove this
    // ... other fields
}

func (a *Adapter) SendMessage(ctx context.Context, sessionID string, msg *base.ChatMessage) error {
    a.senderMu.RLock()
    sender := a.sender
    a.senderMu.RUnlock()
    if sender == nil {
        return fmt.Errorf("sender not configured")
    }
    return sender(ctx, sessionID, msg)
}

func (a *Adapter) SetSender(fn func(ctx context.Context, sessionID string, msg *base.ChatMessage) error) {
    a.senderMu.Lock()
    defer a.senderMu.Unlock()
    a.sender = fn
}

func (a *Adapter) Stop() error {
    done := make(chan struct{})
    go func() {
        a.webhookWg.Wait()
        close(done)
    }()
    select {
    case <-done:
    case <-time.After(5 * time.Second):
        a.Logger().Warn("Timeout waiting for webhook goroutines")
    }
    return a.Adapter.Stop()
}

// AFTER - Using base utilities:
type Adapter struct {
    *base.Adapter
    config      Config
    webhookPath string
    sender      *base.SenderWithMutex  // Changed
    webhook     *base.WebhookRunner    // Added
    // ... other fields (remove senderMu, webhookWg)
}

// In NewAdapter:
a := &Adapter{
    config:      config,
    webhookPath: "/webhook",
    sender:      base.NewSenderWithMutex(),     // Add
    webhook:     base.NewWebhookRunner(logger), // Add
}

// Set default sender after construction
if config.BotToken != "" {
    a.sender.SetSender(a.defaultSender)
}

func (a *Adapter) SendMessage(ctx context.Context, sessionID string, msg *base.ChatMessage) error {
    return a.sender.SendMessage(ctx, sessionID, msg)
}

func (a *Adapter) SetSender(fn func(ctx context.Context, sessionID string, msg *base.ChatMessage) error) {
    a.sender.SetSender(fn)
}

func (a *Adapter) Stop() error {
    a.webhook.Stop()
    return a.Adapter.Stop()
}

// In handleWebhook (replace a.webhookWg.Add(1) pattern):
// BEFORE:
a.webhookWg.Add(1)
go func() {
    defer a.webhookWg.Done()
    if err := a.Handler()(reqCtx, msg); err != nil {
        a.Logger().Error("Handle message failed", "error", err)
    }
}()

// AFTER:
a.webhook.Run(ctx, a.Handler(), msg)
```

**Special Notes per Adapter**:

1. **telegram/adapter.go**:
   - Has `RateLimiter` - keep this
   - In `SendMessage()`, rate limiter check should come BEFORE sender call:
     ```go
     func (a *Adapter) SendMessage(ctx context.Context, sessionID string, msg *base.ChatMessage) error {
         if err := a.rateLimiter.Wait(ctx); err != nil {
             return fmt.Errorf("rate limited: %w", err)
         }
         return a.sender.SendMessage(ctx, sessionID, msg)
     }
     ```
   - Remove `sync` import if no longer needed (check for other sync usage)

2. **whatsapp/adapter.go**:
   - Simpler refactor - just replace sender/senderMu/webhookWg pattern
   - Remove `sync` import

3. **dingtalk/adapter.go**:
   - Has extra fields: `token`, `tokenExpire`, `tokenMu` - keep these
   - Remove `sync.RWMutex` but keep `sync.Mutex` for tokenMu
   - Keep `sync` import for `tokenMu`

**Estimated Code Reduction**: ~50-60 lines per adapter × 3 adapters = 150-180 lines

---

## Verification Commands

After completing each adapter:

```bash
# Build verification
go build ./...

# Test verification
go test ./...

# Race detector (final check)
go test -race ./...
```

---

## Files Created During This Project

| File | Purpose |
|------|---------|
| `internal/persistence/markers.go` | SessionMarkerStore interface and implementations |
| `chatapps/base/sender.go` | SenderWithMutex utility |
| `chatapps/base/http.go` | HTTPClient utility |
| `chatapps/base/webhook.go` | WebhookRunner utility |

## Files Modified During This Project

| File | Changes |
|------|---------|
| `internal/engine/pool.go` | Uses markerStore instead of markerDir |
| `provider/provider.go` | Removed 3 unused interface methods |
| `provider/claude_provider.go` | Removed dead code + migrated to SessionMarkerStore |
| `provider/opencode_provider.go` | Removed dead code methods |
| `hotplex.go` | Added compile-time interface check |
| `internal/security/detector.go` | switch → map for suggestions |
| `internal/engine/pool_test.go` | Updated to use markerStore |

---

## Next Session Quick Start

```bash
cd /Users/huangzhonghui/HotPlex

# Verify current state
go build ./...
go test ./...

# Continue P2-4: Refactor remaining 3 adapters
# Files to edit:
#   - chatapps/telegram/adapter.go
#   - chatapps/whatsapp/adapter.go  
#   - chatapps/dingtalk/adapter.go
```

---

## Architecture Context

```
HotPlex Project Structure:
├── provider/
│   ├── provider.go          # Provider interface
│   ├── claude_provider.go   # Claude Code CLI (✅ P1-3 Done)
│   └── opencode_provider.go # OpenCode CLI
├── chatapps/
│   ├── base/
│   │   ├── sender.go        # SenderWithMutex utility
│   │   ├── webhook.go       # WebhookRunner utility
│   │   └── http.go          # HTTPClient utility
│   ├── slack/adapter.go     # ✅ Already using base utilities
│   ├── discord/adapter.go   # ✅ Already using base utilities
│   ├── telegram/adapter.go  # 🚧 P2-4 Pending
│   ├── whatsapp/adapter.go  # 🚧 P2-4 Pending
│   └── dingtalk/adapter.go  # 🚧 P2-4 Pending
├── internal/
│   ├── persistence/
│   │   └── markers.go       # SessionMarkerStore interface
│   └── engine/
│       └── pool.go          # Uses SessionMarkerStore
└── hotplex.go               # Main entry point
```

---

## Summary

**Done**: 6 tasks (P1, P1-3, P3, P4, P5, P2 partial)
**Remaining**: P2-4 (3 adapters: telegram, whatsapp, dingtalk)
**Estimated Time**: 1-1.5 hours
