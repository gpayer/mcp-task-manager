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

func TestFindProjectRoot_ConfigFile(t *testing.T) {
	// Create temp directory structure with mcp-tasks.yaml
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub", "deep")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirs: %v", err)
	}

	// Create config file in tmpDir
	configPath := filepath.Join(tmpDir, "mcp-tasks.yaml")
	if err := os.WriteFile(configPath, []byte("task_types:\n  - feature\n"), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Save and restore cwd
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	// Change to deep subdirectory
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	root, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot() error = %v", err)
	}

	if root != tmpDir {
		t.Errorf("FindProjectRoot() = %q, want %q", root, tmpDir)
	}
}

func TestFindProjectRoot_TasksDirectory(t *testing.T) {
	// Create temp directory structure with tasks/ directory (no config file)
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub", "deep")
	tasksDir := filepath.Join(tmpDir, "tasks")

	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirs: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Save and restore cwd
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	// Change to deep subdirectory
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	root, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot() error = %v", err)
	}

	if root != tmpDir {
		t.Errorf("FindProjectRoot() = %q, want %q", root, tmpDir)
	}
}

func TestFindProjectRoot_ConfigFilePreferredOverTasksDir(t *testing.T) {
	// Create two levels: one with tasks/, parent with mcp-tasks.yaml
	// Should find the one with config file first (it's higher priority at same level)
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")

	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirs: %v", err)
	}

	// Create both: config file and tasks dir at same level (config should win)
	configPath := filepath.Join(subDir, "mcp-tasks.yaml")
	tasksDir := filepath.Join(subDir, "tasks")

	if err := os.WriteFile(configPath, []byte("task_types:\n  - feature\n"), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Save and restore cwd
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	root, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot() error = %v", err)
	}

	// Should find subDir (where both exist) - config file is checked first
	if root != subDir {
		t.Errorf("FindProjectRoot() = %q, want %q", root, subDir)
	}
}

func TestFindProjectRoot_NotFound(t *testing.T) {
	// Create temp directory with nothing
	tmpDir := t.TempDir()

	// Save and restore cwd
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	root, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot() error = %v", err)
	}

	if root != "" {
		t.Errorf("FindProjectRoot() = %q, want empty string", root)
	}
}

func TestFindProjectRoot_InProjectRoot(t *testing.T) {
	// When already in project root, should return that directory
	tmpDir := t.TempDir()

	// Create config file in tmpDir
	configPath := filepath.Join(tmpDir, "mcp-tasks.yaml")
	if err := os.WriteFile(configPath, []byte("task_types:\n  - feature\n"), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Save and restore cwd
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	root, err := FindProjectRoot()
	if err != nil {
		t.Fatalf("FindProjectRoot() error = %v", err)
	}

	if root != tmpDir {
		t.Errorf("FindProjectRoot() = %q, want %q", root, tmpDir)
	}
}

func TestLoad_WithEnvOverride_SetsProjectFound(t *testing.T) {
	// Save and restore env
	oldVal := os.Getenv("MCP_TASKS_DIR")
	defer os.Setenv("MCP_TASKS_DIR", oldVal)

	os.Setenv("MCP_TASKS_DIR", "/custom/path")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.ProjectFound {
		t.Error("Load() ProjectFound = false, want true when env var is set")
	}
}

func TestLoad_FindsProjectRoot(t *testing.T) {
	// Create temp directory structure with mcp-tasks.yaml in parent
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub", "deep")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirs: %v", err)
	}

	// Create config file in tmpDir (parent)
	configPath := filepath.Join(tmpDir, "mcp-tasks.yaml")
	if err := os.WriteFile(configPath, []byte("task_types:\n  - feature\n  - chore\n"), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Clear env var to ensure we use FindProjectRoot
	oldVal := os.Getenv("MCP_TASKS_DIR")
	defer os.Setenv("MCP_TASKS_DIR", oldVal)
	os.Unsetenv("MCP_TASKS_DIR")

	// Save and restore cwd
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	// Change to deep subdirectory
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should find project root and set DataDir
	expectedDataDir := filepath.Join(tmpDir, "tasks")
	if cfg.DataDir != expectedDataDir {
		t.Errorf("Load() DataDir = %q, want %q", cfg.DataDir, expectedDataDir)
	}

	if !cfg.ProjectFound {
		t.Error("Load() ProjectFound = false, want true when project found")
	}

	// Should also load config from project root
	if !cfg.IsValidTaskType("chore") {
		t.Error("Load() should have loaded config from project root, but 'chore' not in TaskTypes")
	}
}

func TestLoad_NoProjectFound(t *testing.T) {
	// Create temp directory with nothing
	tmpDir := t.TempDir()

	// Clear env var
	oldVal := os.Getenv("MCP_TASKS_DIR")
	defer os.Setenv("MCP_TASKS_DIR", oldVal)
	os.Unsetenv("MCP_TASKS_DIR")

	// Save and restore cwd
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should use default "./tasks"
	if cfg.DataDir != "./tasks" {
		t.Errorf("Load() DataDir = %q, want %q", cfg.DataDir, "./tasks")
	}

	if cfg.ProjectFound {
		t.Error("Load() ProjectFound = true, want false when no project found")
	}
}

func TestLoad_LoadsConfigFromProjectRoot(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirs: %v", err)
	}

	// Create config file in tmpDir with custom task types
	configPath := filepath.Join(tmpDir, "mcp-tasks.yaml")
	configContent := "task_types:\n  - feature\n  - bug\n  - docs\n  - refactor\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Clear env var
	oldVal := os.Getenv("MCP_TASKS_DIR")
	defer os.Setenv("MCP_TASKS_DIR", oldVal)
	os.Unsetenv("MCP_TASKS_DIR")

	// Save and restore cwd
	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should load custom task types from config file in project root
	expectedTypes := []string{"feature", "bug", "docs", "refactor"}
	if len(cfg.TaskTypes) != len(expectedTypes) {
		t.Errorf("Load() TaskTypes length = %d, want %d", len(cfg.TaskTypes), len(expectedTypes))
	}

	for _, tt := range expectedTypes {
		if !cfg.IsValidTaskType(tt) {
			t.Errorf("Load() should have task type %q, but it's not valid", tt)
		}
	}
}

func TestLoad_EnvVarLoadsConfigFromParentDir(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	tasksDir := filepath.Join(tmpDir, "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("failed to create tasks dir: %v", err)
	}

	// Create config file in tmpDir (parent of tasks)
	configPath := filepath.Join(tmpDir, "mcp-tasks.yaml")
	configContent := "task_types:\n  - feature\n  - custom\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Set env var to tasks directory
	oldVal := os.Getenv("MCP_TASKS_DIR")
	defer os.Setenv("MCP_TASKS_DIR", oldVal)
	os.Setenv("MCP_TASKS_DIR", tasksDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should load config from parent of env var path
	if !cfg.IsValidTaskType("custom") {
		t.Error("Load() should have loaded config from parent of MCP_TASKS_DIR")
	}
}
