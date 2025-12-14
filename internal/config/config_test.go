package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DataDir != "./tasks" {
		t.Errorf("DefaultConfig().DataDir = %q, want %q", cfg.DataDir, "./tasks")
	}

	if len(cfg.TaskTypes) != 2 {
		t.Errorf("DefaultConfig().TaskTypes length = %d, want 2", len(cfg.TaskTypes))
	}

	expectedTypes := map[string]bool{"feature": true, "bug": true}
	for _, tt := range cfg.TaskTypes {
		if !expectedTypes[tt] {
			t.Errorf("unexpected task type: %q", tt)
		}
	}
}

func TestIsValidTaskType(t *testing.T) {
	cfg := &Config{
		TaskTypes: []string{"feature", "bug", "chore"},
	}

	tests := []struct {
		taskType string
		valid    bool
	}{
		{"feature", true},
		{"bug", true},
		{"chore", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.taskType, func(t *testing.T) {
			if got := cfg.IsValidTaskType(tt.taskType); got != tt.valid {
				t.Errorf("IsValidTaskType(%q) = %v, want %v", tt.taskType, got, tt.valid)
			}
		})
	}
}

func TestTasksDir_Absolute(t *testing.T) {
	cfg := &Config{DataDir: "/absolute/path"}
	if got := cfg.TasksDir(); got != "/absolute/path" {
		t.Errorf("TasksDir() = %q, want %q", got, "/absolute/path")
	}
}

func TestTasksDir_Relative(t *testing.T) {
	cfg := &Config{DataDir: "./tasks"}
	cwd, _ := os.Getwd()
	expected := filepath.Join(cwd, "./tasks")

	if got := cfg.TasksDir(); got != expected {
		t.Errorf("TasksDir() = %q, want %q", got, expected)
	}
}

func TestLoad_WithEnvOverride(t *testing.T) {
	// Save and restore env
	oldVal := os.Getenv("MCP_TASKS_DIR")
	defer os.Setenv("MCP_TASKS_DIR", oldVal)

	os.Setenv("MCP_TASKS_DIR", "/custom/path")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DataDir != "/custom/path" {
		t.Errorf("Load() DataDir = %q, want %q", cfg.DataDir, "/custom/path")
	}
}
