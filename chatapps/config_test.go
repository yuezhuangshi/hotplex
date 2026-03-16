package chatapps

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hrygo/hotplex/provider"
)

// TestConfigInheritance tests the YAML config inheritance mechanism
func TestConfigInheritance(t *testing.T) {
	// Create temp directory for test configs
	tmpDir, err := os.MkdirTemp("", "hotplex-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Create base subdirectory
	baseDir := filepath.Join(tmpDir, "base")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatalf("Failed to create base dir: %v", err)
	}

	// Write parent config (base)
	parentConfig := `
platform: slack
mode: socket
system_prompt: "Base system prompt"
task_instructions: "Base task instructions"
engine:
  timeout: 30m
  idle_timeout: 1h
  work_dir: /base/work
provider:
  type: claude-code
  default_model: sonnet
security:
  verify_signature: true
  permission:
    dm_policy: allow
    group_policy: block
    bot_user_id: BASE_BOT_ID
features:
  chunking:
    enabled: true
    max_chars: 4000
message_store:
  enabled: true
  type: sqlite
`
	parentPath := filepath.Join(baseDir, "slack.yaml")
	if err := os.WriteFile(parentPath, []byte(parentConfig), 0644); err != nil {
		t.Fatalf("Failed to write parent config: %v", err)
	}

	// Write child config that inherits from parent
	childConfig := `
inherits: ./base/slack.yaml
platform: slack
mode: http
system_prompt: "Child system prompt"
security:
  permission:
    group_policy: multibot
    bot_user_id: CHILD_BOT_ID
engine:
  work_dir: /child/work
features:
  chunking:
    max_chars: 2000
`
	childPath := filepath.Join(tmpDir, "slack.yaml")
	if err := os.WriteFile(childPath, []byte(childConfig), 0644); err != nil {
		t.Fatalf("Failed to write child config: %v", err)
	}

	// Load config
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	loader, err := NewConfigLoader(tmpDir, logger)
	if err != nil {
		t.Fatalf("Failed to create config loader: %v", err)
	}

	// Verify config was loaded
	cfg := loader.GetConfig("slack")
	if cfg == nil {
		t.Fatal("Expected slack config to be loaded")
	}

	// Verify child overrides parent
	if cfg.Mode != "http" {
		t.Errorf("Expected Mode 'http', got '%s'", cfg.Mode)
	}
	if cfg.SystemPrompt != "Child system prompt" {
		t.Errorf("Expected SystemPrompt 'Child system prompt', got '%s'", cfg.SystemPrompt)
	}
	if cfg.Security.Permission.GroupPolicy != "multibot" {
		t.Errorf("Expected GroupPolicy 'multibot', got '%s'", cfg.Security.Permission.GroupPolicy)
	}
	if cfg.Security.Permission.BotUserID != "CHILD_BOT_ID" {
		t.Errorf("Expected BotUserID 'CHILD_BOT_ID', got '%s'", cfg.Security.Permission.BotUserID)
	}
	if cfg.Engine.WorkDir != "/child/work" {
		t.Errorf("Expected Engine.WorkDir '/child/work', got '%s'", cfg.Engine.WorkDir)
	}
	if cfg.Features.Chunking.MaxChars != 2000 {
		t.Errorf("Expected Features.Chunking.MaxChars 2000, got %d", cfg.Features.Chunking.MaxChars)
	}

	// Verify parent values are inherited when not overridden
	if cfg.TaskInstructions != "Base task instructions" {
		t.Errorf("Expected TaskInstructions 'Base task instructions', got '%s'", cfg.TaskInstructions)
	}
	if cfg.Engine.Timeout != 30*time.Minute {
		t.Errorf("Expected Engine.Timeout 30m, got %v", cfg.Engine.Timeout)
	}
	if cfg.Engine.IdleTimeout != time.Hour {
		t.Errorf("Expected Engine.IdleTimeout 1h, got %v", cfg.Engine.IdleTimeout)
	}
	if cfg.Provider.Type != provider.ProviderTypeClaudeCode {
		t.Errorf("Expected Provider.Type 'claude-code', got '%s'", cfg.Provider.Type)
	}
	if cfg.Provider.DefaultModel != "sonnet" {
		t.Errorf("Expected Provider.DefaultModel 'sonnet', got '%s'", cfg.Provider.DefaultModel)
	}
	if !*cfg.Security.VerifySignature {
		t.Error("Expected VerifySignature to be true")
	}
	if cfg.Security.Permission.DMPolicy != "allow" {
		t.Errorf("Expected DMPolicy 'allow', got '%s'", cfg.Security.Permission.DMPolicy)
	}
	if !*cfg.Features.Chunking.Enabled {
		t.Error("Expected Features.Chunking.Enabled to be true")
	}
	if !*cfg.MessageStore.Enabled {
		t.Error("Expected MessageStore.Enabled to be true")
	}
	if cfg.MessageStore.Type != "sqlite" {
		t.Errorf("Expected MessageStore.Type 'sqlite', got '%s'", cfg.MessageStore.Type)
	}
}

// TestCircularInheritanceDetection tests that circular inheritance is detected
func TestCircularInheritanceDetection(t *testing.T) {
	// Create temp directory for test configs
	tmpDir, err := os.MkdirTemp("", "hotplex-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Create config A that inherits from B
	configA := `
inherits: ./b.yaml
platform: slack
`
	if err := os.WriteFile(filepath.Join(tmpDir, "a.yaml"), []byte(configA), 0644); err != nil {
		t.Fatalf("Failed to write config A: %v", err)
	}

	// Create config B that inherits from A (circular)
	configB := `
inherits: ./a.yaml
platform: slack
`
	if err := os.WriteFile(filepath.Join(tmpDir, "b.yaml"), []byte(configB), 0644); err != nil {
		t.Fatalf("Failed to write config B: %v", err)
	}

	// Load config - should fail with circular inheritance error
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	loader, err := NewConfigLoader(tmpDir, logger)
	if err != nil {
		t.Fatalf("Unexpected error creating loader: %v", err)
	}

	// Circular inheritance should prevent loading - config should be nil or empty
	cfg := loader.GetConfig("slack")
	if cfg != nil {
		t.Error("Expected config to not be loaded due to circular inheritance")
	}
}

// TestDeepInheritance tests multi-level inheritance
func TestDeepInheritance(t *testing.T) {
	// Create temp directory for test configs
	tmpDir, err := os.MkdirTemp("", "hotplex-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Create base subdirectory
	baseDir := filepath.Join(tmpDir, "base")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatalf("Failed to create base dir: %v", err)
	}

	// Grandparent config
	grandparentConfig := `
platform: slack
mode: socket
system_prompt: "Grandparent prompt"
engine:
  timeout: 15m
security:
  permission:
    dm_policy: block
`
	if err := os.WriteFile(filepath.Join(baseDir, "grandparent.yaml"), []byte(grandparentConfig), 0644); err != nil {
		t.Fatalf("Failed to write grandparent config: %v", err)
	}

	// Parent config inherits from grandparent
	parentConfig := `
inherits: ./grandparent.yaml
system_prompt: "Parent prompt"
engine:
  timeout: 30m
  idle_timeout: 2h
security:
  permission:
    dm_policy: allow
    group_policy: multibot
`
	if err := os.WriteFile(filepath.Join(baseDir, "parent.yaml"), []byte(parentConfig), 0644); err != nil {
		t.Fatalf("Failed to write parent config: %v", err)
	}

	// Child config inherits from parent
	childConfig := `
inherits: ./base/parent.yaml
mode: http
engine:
  work_dir: /child/work
security:
  permission:
    bot_user_id: CHILD_BOT
`
	if err := os.WriteFile(filepath.Join(tmpDir, "slack.yaml"), []byte(childConfig), 0644); err != nil {
		t.Fatalf("Failed to write child config: %v", err)
	}

	// Load config
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	loader, err := NewConfigLoader(tmpDir, logger)
	if err != nil {
		t.Fatalf("Failed to create config loader: %v", err)
	}

	cfg := loader.GetConfig("slack")
	if cfg == nil {
		t.Fatal("Expected slack config to be loaded")
	}

	// Verify values from different levels
	// From child
	if cfg.Mode != "http" {
		t.Errorf("Expected Mode 'http' from child, got '%s'", cfg.Mode)
	}
	if cfg.Engine.WorkDir != "/child/work" {
		t.Errorf("Expected WorkDir '/child/work' from child, got '%s'", cfg.Engine.WorkDir)
	}
	if cfg.Security.Permission.BotUserID != "CHILD_BOT" {
		t.Errorf("Expected BotUserID 'CHILD_BOT' from child, got '%s'", cfg.Security.Permission.BotUserID)
	}

	// From parent (overrides grandparent)
	if cfg.SystemPrompt != "Parent prompt" {
		t.Errorf("Expected SystemPrompt 'Parent prompt' from parent, got '%s'", cfg.SystemPrompt)
	}
	if cfg.Engine.Timeout != 30*time.Minute {
		t.Errorf("Expected Timeout 30m from parent, got %v", cfg.Engine.Timeout)
	}
	if cfg.Engine.IdleTimeout != 2*time.Hour {
		t.Errorf("Expected IdleTimeout 2h from parent, got %v", cfg.Engine.IdleTimeout)
	}
	if cfg.Security.Permission.DMPolicy != "allow" {
		t.Errorf("Expected DMPolicy 'allow' from parent, got '%s'", cfg.Security.Permission.DMPolicy)
	}
	if cfg.Security.Permission.GroupPolicy != "multibot" {
		t.Errorf("Expected GroupPolicy 'multibot' from parent, got '%s'", cfg.Security.Permission.GroupPolicy)
	}

	// From grandparent (inherited through parent)
	if cfg.Platform != "slack" {
		t.Errorf("Expected Platform 'slack' from grandparent, got '%s'", cfg.Platform)
	}
}

// TestNoInheritance tests config without inheritance works normally
func TestNoInheritance(t *testing.T) {
	// Create temp directory for test configs
	tmpDir, err := os.MkdirTemp("", "hotplex-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	config := `
platform: slack
mode: socket
system_prompt: "Standalone config"
engine:
  timeout: 30m
`
	if err := os.WriteFile(filepath.Join(tmpDir, "slack.yaml"), []byte(config), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load config
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	loader, err := NewConfigLoader(tmpDir, logger)
	if err != nil {
		t.Fatalf("Failed to create config loader: %v", err)
	}

	cfg := loader.GetConfig("slack")
	if cfg == nil {
		t.Fatal("Expected slack config to be loaded")
	}

	if cfg.SystemPrompt != "Standalone config" {
		t.Errorf("Expected SystemPrompt 'Standalone config', got '%s'", cfg.SystemPrompt)
	}
	if cfg.Inherits != "" {
		t.Errorf("Expected Inherits to be empty, got '%s'", cfg.Inherits)
	}
}

// TestMergeConfigs tests the mergeConfigs function directly
func TestMergeConfigs(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name     string
		parent   *PlatformConfig
		child    *PlatformConfig
		expected PlatformConfig
	}{
		{
			name:     "nil parent returns child",
			parent:   nil,
			child:    &PlatformConfig{Platform: "slack", Mode: "http"},
			expected: PlatformConfig{Platform: "slack", Mode: "http"},
		},
		{
			name:     "nil child returns parent",
			parent:   &PlatformConfig{Platform: "slack", Mode: "socket"},
			child:    nil,
			expected: PlatformConfig{Platform: "slack", Mode: "socket"},
		},
		{
			name:     "child string overrides parent",
			parent:   &PlatformConfig{SystemPrompt: "parent prompt"},
			child:    &PlatformConfig{SystemPrompt: "child prompt"},
			expected: PlatformConfig{SystemPrompt: "child prompt"},
		},
		{
			name:     "child empty string keeps parent",
			parent:   &PlatformConfig{SystemPrompt: "parent prompt"},
			child:    &PlatformConfig{SystemPrompt: ""},
			expected: PlatformConfig{SystemPrompt: "parent prompt"},
		},
		{
			name:     "child bool overrides parent",
			parent:   &PlatformConfig{Security: SecurityConfig{VerifySignature: &trueVal}},
			child:    &PlatformConfig{Security: SecurityConfig{VerifySignature: &falseVal}},
			expected: PlatformConfig{Security: SecurityConfig{VerifySignature: &falseVal}},
		},
		{
			name:     "child slice replaces parent",
			parent:   &PlatformConfig{Engine: EngineConfig{AllowedTools: []string{"tool1", "tool2"}}},
			child:    &PlatformConfig{Engine: EngineConfig{AllowedTools: []string{"tool3"}}},
			expected: PlatformConfig{Engine: EngineConfig{AllowedTools: []string{"tool3"}}},
		},
		{
			name:     "child map merges with parent",
			parent:   &PlatformConfig{Options: map[string]any{"key1": "val1"}},
			child:    &PlatformConfig{Options: map[string]any{"key2": "val2"}},
			expected: PlatformConfig{Options: map[string]any{"key1": "val1", "key2": "val2"}},
		},
		{
			name:     "child map overrides parent key",
			parent:   &PlatformConfig{Options: map[string]any{"key1": "val1"}},
			child:    &PlatformConfig{Options: map[string]any{"key1": "val2"}},
			expected: PlatformConfig{Options: map[string]any{"key1": "val2"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeConfigs(tt.parent, tt.child)

			// Compare key fields based on test case
			if tt.expected.Platform != "" && result.Platform != tt.expected.Platform {
				t.Errorf("Platform: expected '%s', got '%s'", tt.expected.Platform, result.Platform)
			}
			if tt.expected.Mode != "" && result.Mode != tt.expected.Mode {
				t.Errorf("Mode: expected '%s', got '%s'", tt.expected.Mode, result.Mode)
			}
			if tt.expected.SystemPrompt != "" && result.SystemPrompt != tt.expected.SystemPrompt {
				t.Errorf("SystemPrompt: expected '%s', got '%s'", tt.expected.SystemPrompt, result.SystemPrompt)
			}
			if tt.expected.Security.VerifySignature != nil {
				if *result.Security.VerifySignature != *tt.expected.Security.VerifySignature {
					t.Errorf("VerifySignature: expected %v, got %v", *tt.expected.Security.VerifySignature, *result.Security.VerifySignature)
				}
			}
			if tt.expected.Engine.AllowedTools != nil {
				if len(result.Engine.AllowedTools) != len(tt.expected.Engine.AllowedTools) {
					t.Errorf("AllowedTools length: expected %d, got %d", len(tt.expected.Engine.AllowedTools), len(result.Engine.AllowedTools))
				}
			}
			if tt.expected.Options != nil {
				for k, v := range tt.expected.Options {
					if result.Options[k] != v {
						t.Errorf("Options[%s]: expected '%v', got '%v'", k, v, result.Options[k])
					}
				}
			}
		})
	}
}

// TestMergeConfigsPointerFields tests merging of pointer fields
func TestMergeConfigsPointerFields(t *testing.T) {
	trueVal := true
	falseVal := false

	// Test Chunking.Enabled
	parent := &PlatformConfig{
		Features: FeaturesConfig{
			Chunking: ChunkingConfig{Enabled: &trueVal, MaxChars: 4000},
		},
	}
	child := &PlatformConfig{
		Features: FeaturesConfig{
			Chunking: ChunkingConfig{Enabled: &falseVal, MaxChars: 2000},
		},
	}

	result := mergeConfigs(parent, child)

	if *result.Features.Chunking.Enabled != false {
		t.Errorf("Expected Chunking.Enabled false, got true")
	}
	if result.Features.Chunking.MaxChars != 2000 {
		t.Errorf("Expected Chunking.MaxChars 2000, got %d", result.Features.Chunking.MaxChars)
	}
}

// TestInheritanceWithEnvVars tests that env vars are expanded in inherited configs
func TestInheritanceWithEnvVars(t *testing.T) {
	// Set test env var
	_ = os.Setenv("HOTPLEX_TEST_BOT_ID", "TEST_BOT_123")
	t.Cleanup(func() { _ = os.Unsetenv("HOTPLEX_TEST_BOT_ID") })

	// Create temp directory for test configs
	tmpDir, err := os.MkdirTemp("", "hotplex-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Create base subdirectory
	baseDir := filepath.Join(tmpDir, "base")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatalf("Failed to create base dir: %v", err)
	}

	// Parent config with env var
	parentConfig := `
platform: slack
security:
  permission:
    bot_user_id: ${HOTPLEX_TEST_BOT_ID}
`
	if err := os.WriteFile(filepath.Join(baseDir, "slack.yaml"), []byte(parentConfig), 0644); err != nil {
		t.Fatalf("Failed to write parent config: %v", err)
	}

	// Child config
	childConfig := `
inherits: ./base/slack.yaml
`
	if err := os.WriteFile(filepath.Join(tmpDir, "slack.yaml"), []byte(childConfig), 0644); err != nil {
		t.Fatalf("Failed to write child config: %v", err)
	}

	// Load config
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	loader, err := NewConfigLoader(tmpDir, logger)
	if err != nil {
		t.Fatalf("Failed to create config loader: %v", err)
	}

	cfg := loader.GetConfig("slack")
	if cfg == nil {
		t.Fatal("Expected slack config to be loaded")
	}

	// Verify env var was expanded
	if cfg.Security.Permission.BotUserID != "TEST_BOT_123" {
		t.Errorf("Expected BotUserID 'TEST_BOT_123', got '%s'", cfg.Security.Permission.BotUserID)
	}
}

// TestInheritanceProvider verifies provider-specific HTTP client config is not required
func TestInheritanceProvider(t *testing.T) {
	// This test verifies the provider config doesn't cause issues with HTTPClient fields
	tmpDir, err := os.MkdirTemp("", "hotplex-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	baseDir := filepath.Join(tmpDir, "base")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatalf("Failed to create base dir: %v", err)
	}

	parentConfig := `
platform: slack
provider:
  type: claude-code
  default_model: sonnet
  default_permission_mode: bypass-permissions
`
	if err := os.WriteFile(filepath.Join(baseDir, "slack.yaml"), []byte(parentConfig), 0644); err != nil {
		t.Fatalf("Failed to write parent config: %v", err)
	}

	childConfig := `
inherits: ./base/slack.yaml
security:
  permission:
    bot_user_id: TEST_BOT
`
	if err := os.WriteFile(filepath.Join(tmpDir, "slack.yaml"), []byte(childConfig), 0644); err != nil {
		t.Fatalf("Failed to write child config: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	loader, err := NewConfigLoader(tmpDir, logger)
	if err != nil {
		t.Fatalf("Failed to create config loader: %v", err)
	}

	cfg := loader.GetConfig("slack")
	if cfg == nil {
		t.Fatal("Expected slack config to be loaded")
	}

	// Verify provider config is inherited
	if cfg.Provider.Type != provider.ProviderTypeClaudeCode {
		t.Errorf("Expected Provider.Type 'claude-code', got '%s'", cfg.Provider.Type)
	}
	if cfg.Provider.DefaultModel != "sonnet" {
		t.Errorf("Expected Provider.DefaultModel 'sonnet', got '%s'", cfg.Provider.DefaultModel)
	}
}
