package hotplex

import (
	"errors"
	"log/slog"
	"os"
	"testing"
)

func TestNewEventWithMeta(t *testing.T) {
	meta := &EventMeta{
		DurationMs:   100,
		ToolName:     "bash",
		ToolID:       "tool-123",
		Status:       "success",
		InputTokens:  50,
		OutputTokens: 100,
	}

	event := NewEventWithMeta("tool_use", "executing command", meta)

	if event.EventType != "tool_use" {
		t.Errorf("EventType = %q, want tool_use", event.EventType)
	}
	if event.EventData != "executing command" {
		t.Errorf("EventData = %q, want 'executing command'", event.EventData)
	}
	if event.Meta.DurationMs != 100 {
		t.Errorf("Meta.DurationMs = %d, want 100", event.Meta.DurationMs)
	}
	if event.Meta.ToolName != "bash" {
		t.Errorf("Meta.ToolName = %q, want bash", event.Meta.ToolName)
	}
}

func TestNewEventWithMeta_NilMeta(t *testing.T) {
	event := NewEventWithMeta("thinking", "AI is thinking", nil)

	if event == nil {
		t.Fatal("NewEventWithMeta returned nil")
	}
	if event.Meta == nil {
		t.Fatal("Meta should not be nil")
	}
	// Meta should be zero-valued struct
	if event.Meta.DurationMs != 0 {
		t.Errorf("Meta.DurationMs = %d, want 0 (zero value)", event.Meta.DurationMs)
	}
}

func TestWrapSafe_NilCallback(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	wrapped := WrapSafe(logger, nil)
	if wrapped != nil {
		t.Error("WrapSafe with nil callback should return nil")
	}
}

func TestWrapSafe_SuccessfulCallback(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	called := false

	cb := func(eventType string, data any) error {
		called = true
		return nil
	}

	wrapped := WrapSafe(logger, cb)
	err := wrapped("test", "data")

	if err != nil {
		t.Errorf("wrapped callback returned error: %v", err)
	}
	if !called {
		t.Error("callback was not called")
	}
}

func TestWrapSafe_ErrorCallback(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	expectedErr := errors.New("callback error")

	cb := func(eventType string, data any) error {
		return expectedErr
	}

	wrapped := WrapSafe(logger, cb)
	err := wrapped("test", "data")

	// WrapSafe should suppress the error (return nil)
	if err != nil {
		t.Errorf("WrapSafe should suppress error, got: %v", err)
	}
}

func TestEventMeta_Fields(t *testing.T) {
	meta := &EventMeta{
		DurationMs:       1000,
		TotalDurationMs:  5000,
		ToolName:         "Write",
		ToolID:           "tool-abc-123",
		Status:           "success",
		ErrorMsg:         "",
		InputTokens:      200,
		OutputTokens:     500,
		CacheWriteTokens: 100,
		CacheReadTokens:  50,
		InputSummary:     "Created new file",
		OutputSummary:    "File written successfully",
		FilePath:         "/path/to/file.go",
		LineCount:        42,
		Progress:         100,
		TotalSteps:       5,
		CurrentStep:      5,
	}

	// Verify all fields are accessible
	if meta.ToolName != "Write" {
		t.Errorf("ToolName = %q, want Write", meta.ToolName)
	}
	if meta.LineCount != 42 {
		t.Errorf("LineCount = %d, want 42", meta.LineCount)
	}
	if meta.Progress != 100 {
		t.Errorf("Progress = %d, want 100", meta.Progress)
	}
}

func TestSessionStatsData_Fields(t *testing.T) {
	data := &SessionStatsData{
		SessionID:            "session-123",
		StartTime:            1700000000,
		EndTime:              1700001000,
		TotalDurationMs:      10000,
		ThinkingDurationMs:   2000,
		ToolDurationMs:       5000,
		GenerationDurationMs: 3000,
		InputTokens:          1000,
		OutputTokens:         500,
		CacheWriteTokens:     200,
		CacheReadTokens:      100,
		TotalTokens:          1500,
		ToolCallCount:        10,
		ToolsUsed:            []string{"bash", "Write", "Edit"},
		FilesModified:        5,
		FilePaths:            []string{"/a.go", "/b.go"},
		TotalCostUSD:         0.05,
		ModelUsed:            "claude-code",
		IsError:              false,
		ErrorMessage:         "",
	}

	if data.SessionID != "session-123" {
		t.Errorf("SessionID = %q, want session-123", data.SessionID)
	}
	if data.TotalCostUSD != 0.05 {
		t.Errorf("TotalCostUSD = %f, want 0.05", data.TotalCostUSD)
	}
	if len(data.ToolsUsed) != 3 {
		t.Errorf("len(ToolsUsed) = %d, want 3", len(data.ToolsUsed))
	}
}
