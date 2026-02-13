package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds application configuration
type Config struct {
	TaskTypes     []string `yaml:"task_types"`
	RelationTypes []string `yaml:"relation_types,omitempty"`
	DataDir       string   `yaml:"-"` // Set from env or default
	ProjectFound  bool     `yaml:"-"` // Whether an existing project was discovered
}

// DefaultRelationTypes returns the default relation types
var DefaultRelationTypes = []string{"blocked_by", "relates_to", "duplicate_of"}

// DefaultConfig returns configuration with defaults
func DefaultConfig() *Config {
	return &Config{
		TaskTypes:     []string{"feature", "bug"},
		RelationTypes: DefaultRelationTypes,
		DataDir:       "./tasks",
	}
}

// Load loads configuration from file and environment
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Check for env override first
	if dir := os.Getenv("MCP_TASKS_DIR"); dir != "" {
		cfg.DataDir = dir
		cfg.ProjectFound = true
		// Try to load config from parent of tasks directory
		configPath := filepath.Join(dir, "..", "mcp-tasks.yaml")
		if data, err := os.ReadFile(configPath); err == nil {
			yaml.Unmarshal(data, cfg)
		}
		return cfg, nil
	}

	// Search for existing project
	projectRoot, err := FindProjectRoot()
	if err != nil {
		return nil, err
	}

	if projectRoot != "" {
		cfg.DataDir = filepath.Join(projectRoot, "tasks")
		cfg.ProjectFound = true
		// Try to load config from project root
		configPath := filepath.Join(projectRoot, "mcp-tasks.yaml")
		if data, err := os.ReadFile(configPath); err == nil {
			yaml.Unmarshal(data, cfg)
		}
	} else {
		// No project found - use cwd default
		cfg.DataDir = "./tasks"
		cfg.ProjectFound = false
		// Still try to load config from cwd
		if data, err := os.ReadFile("mcp-tasks.yaml"); err == nil {
			yaml.Unmarshal(data, cfg)
		}
	}

	return cfg, nil
}

// TasksDir returns the full path to the tasks directory
func (c *Config) TasksDir() string {
	if filepath.IsAbs(c.DataDir) {
		return c.DataDir
	}
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, c.DataDir)
}

// IsValidTaskType checks if task type is in configured list
func (c *Config) IsValidTaskType(t string) bool {
	for _, valid := range c.TaskTypes {
		if t == valid {
			return true
		}
	}
	return false
}

// IsValidRelationType checks if relation type is in configured list
func (c *Config) IsValidRelationType(t string) bool {
	for _, valid := range c.RelationTypes {
		if t == valid {
			return true
		}
	}
	return false
}

// FindProjectRoot searches for an existing project root by looking for
// mcp-tasks.yaml or a tasks directory, starting from cwd and moving up.
// Returns the directory containing the config/tasks, or empty string if not found.
func FindProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := cwd
	for {
		// First priority: mcp-tasks.yaml
		if _, err := os.Stat(filepath.Join(dir, "mcp-tasks.yaml")); err == nil {
			return dir, nil
		}

		// Second priority: tasks directory
		if info, err := os.Stat(filepath.Join(dir, "tasks")); err == nil && info.IsDir() {
			return dir, nil
		}

		// Move to parent
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", nil
		}
		dir = parent
	}
}
