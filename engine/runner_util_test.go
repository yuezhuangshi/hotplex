package engine

import (
	"testing"

	"github.com/hrygo/hotplex/event"
	intengine "github.com/hrygo/hotplex/internal/engine"
	"github.com/hrygo/hotplex/types"
)

func TestEngine_Close(t *testing.T) {
	logger := newTestLogger()
	mockManager := &mockSessionManager{sessions: make(map[string]*intengine.Session)}

	engine := &Engine{
		opts:    EngineOptions{Namespace: "test"},
		logger:  logger,
		manager: mockManager,
	}

	// Close should succeed
	err := engine.Close()
	if err != nil {
		t.Errorf("Close() error: %v", err)
	}
}

func TestEngine_GetCLIVersion(t *testing.T) {
	logger := newTestLogger()

	engine := &Engine{
		opts:    EngineOptions{Namespace: "test"},
		logger:  logger,
		cliPath: "/nonexistent/claude",
	}

	// Should fail with nonexistent CLI
	_, err := engine.GetCLIVersion()
	if err == nil {
		t.Error("GetCLIVersion() should fail for nonexistent CLI")
	}
}

func TestEngine_StopSession_WithMockManager(t *testing.T) {
	logger := newTestLogger()
	mockManager := &mockSessionManager{
		sessions: make(map[string]*intengine.Session),
	}

	engine := &Engine{
		opts:    EngineOptions{Namespace: "test"},
		logger:  logger,
		manager: mockManager,
	}

	err := engine.StopSession("test-session", "test reason")
	if err != nil {
		t.Errorf("StopSession() error: %v", err)
	}
}

func TestWrapSafe_WithNilLogger(t *testing.T) {
	// WrapSafe with nil logger should still work
	cb := func(eventType string, data any) error {
		return nil
	}

	wrapped := event.WrapSafe(nil, cb)
	if wrapped == nil {
		t.Error("WrapSafe should not return nil for non-nil callback")
	}
}

func TestWrapSafe_WithErrorAndNilLogger(t *testing.T) {
	// WrapSafe with nil logger and error callback should not panic
	cb := func(eventType string, data any) error {
		return types.ErrDangerBlocked
	}

	wrapped := event.WrapSafe(nil, cb)
	err := wrapped("test", nil)

	// Should suppress error and return nil
	if err != nil {
		t.Errorf("WrapSafe should suppress error, got: %v", err)
	}
}
