# MCP Task Manager MVP Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a functional MCP server that provides task management tools for Claude and coding agents.

**Architecture:** Go MCP server using mcp-go library with markdown file storage (YAML frontmatter), JSON index cache, and 8 tools split between task management and agent workflow.

**Tech Stack:** Go 1.21+, github.com/mark3labs/mcp-go, gopkg.in/yaml.v3

---

## Task 1: Project Initialization

**Files:**
- Create: `go.mod`
- Create: `go.sum`
- Create: `cmd/mcp-task-manager/main.go`

**Step 1: Initialize Go module**

Run:
```bash
cd /home/gernot/src/mcp-task-manager
go mod init github.com/gpayer/mcp-task-manager
```

Expected: `go.mod` created with module path

**Step 2: Create minimal main.go**

Create `cmd/mcp-task-manager/main.go`:
```go
package main

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	s := server.NewMCPServer(
		"mcp-task-manager",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
```

**Step 3: Fetch dependencies**

Run:
```bash
go mod tidy
```

Expected: `go.sum` created, mcp-go downloaded

**Step 4: Verify it compiles**

Run:
```bash
go build ./cmd/mcp-task-manager
```

Expected: No errors, binary created

**Step 5: Commit**

```bash
git init
git add go.mod go.sum cmd/
git commit -m "feat: initialize project with minimal MCP server"
```

---

## Task 2: Task Model and Types

**Files:**
- Create: `internal/task/task.go`

**Step 1: Create task types**

Create `internal/task/task.go`:
```go
package task

import "time"

// Status represents the current state of a task
type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

// Priority represents task priority level
type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"
)

// PriorityOrder returns numeric order for sorting (lower = higher priority)
func (p Priority) Order() int {
	switch p {
	case PriorityCritical:
		return 0
	case PriorityHigh:
		return 1
	case PriorityMedium:
		return 2
	case PriorityLow:
		return 3
	default:
		return 99
	}
}

// Task represents a single task
type Task struct {
	ID          int       `yaml:"id" json:"id"`
	Title       string    `yaml:"title" json:"title"`
	Description string    `yaml:"-" json:"description"` // Stored in markdown body
	Status      Status    `yaml:"status" json:"status"`
	Priority    Priority  `yaml:"priority" json:"priority"`
	Type        string    `yaml:"type" json:"type"`
	CreatedAt   time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time `yaml:"updated_at" json:"updated_at"`
}

// IsValidStatus checks if status is valid
func IsValidStatus(s string) bool {
	switch Status(s) {
	case StatusTodo, StatusInProgress, StatusDone:
		return true
	}
	return false
}

// IsValidPriority checks if priority is valid
func IsValidPriority(p string) bool {
	switch Priority(p) {
	case PriorityCritical, PriorityHigh, PriorityMedium, PriorityLow:
		return true
	}
	return false
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./internal/task
```

Expected: No errors

**Step 3: Commit**

```bash
git add internal/
git commit -m "feat: add task model and types"
```

---

## Task 3: Configuration

**Files:**
- Create: `internal/config/config.go`
- Create: `mcp-tasks.yaml`

**Step 1: Create config package**

Create `internal/config/config.go`:
```go
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds application configuration
type Config struct {
	TaskTypes []string `yaml:"task_types"`
	DataDir   string   `yaml:"-"` // Set from env or default
}

// DefaultConfig returns configuration with defaults
func DefaultConfig() *Config {
	return &Config{
		TaskTypes: []string{"feature", "bug"},
		DataDir:   "./tasks",
	}
}

// Load loads configuration from file and environment
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Try to load from config file
	configPath := "mcp-tasks.yaml"
	if data, err := os.ReadFile(configPath); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	// Override data dir from environment
	if dir := os.Getenv("MCP_TASKS_DIR"); dir != "" {
		cfg.DataDir = dir
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
```

**Step 2: Create default config file**

Create `mcp-tasks.yaml`:
```yaml
task_types:
  - feature
  - bug
```

**Step 3: Fetch yaml dependency**

Run:
```bash
go mod tidy
```

Expected: yaml.v3 added to go.sum

**Step 4: Verify it compiles**

Run:
```bash
go build ./internal/config
```

Expected: No errors

**Step 5: Commit**

```bash
git add internal/config mcp-tasks.yaml go.sum
git commit -m "feat: add configuration with task types"
```

---

## Task 4: Markdown Storage

**Files:**
- Create: `internal/storage/markdown.go`

**Step 1: Create markdown storage**

Create `internal/storage/markdown.go`:
```go
package storage

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gpayer/mcp-task-manager/internal/task"
	"gopkg.in/yaml.v3"
)

// MarkdownStorage handles reading/writing task markdown files
type MarkdownStorage struct {
	dir string
}

// NewMarkdownStorage creates a new markdown storage
func NewMarkdownStorage(dir string) *MarkdownStorage {
	return &MarkdownStorage{dir: dir}
}

// EnsureDir creates the tasks directory if it doesn't exist
func (s *MarkdownStorage) EnsureDir() error {
	return os.MkdirAll(s.dir, 0755)
}

// taskPath returns the file path for a task ID
func (s *MarkdownStorage) taskPath(id int) string {
	return filepath.Join(s.dir, fmt.Sprintf("%03d.md", id))
}

// Save writes a task to a markdown file
func (s *MarkdownStorage) Save(t *task.Task) error {
	if err := s.EnsureDir(); err != nil {
		return err
	}

	// Build frontmatter
	frontmatter := struct {
		ID        int           `yaml:"id"`
		Title     string        `yaml:"title"`
		Status    task.Status   `yaml:"status"`
		Priority  task.Priority `yaml:"priority"`
		Type      string        `yaml:"type"`
		CreatedAt string        `yaml:"created_at"`
		UpdatedAt string        `yaml:"updated_at"`
	}{
		ID:        t.ID,
		Title:     t.Title,
		Status:    t.Status,
		Priority:  t.Priority,
		Type:      t.Type,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(frontmatter); err != nil {
		return err
	}
	buf.WriteString("---\n\n")
	buf.WriteString(t.Description)

	// Atomic write: write to temp, then rename
	tmpPath := s.taskPath(t.ID) + ".tmp"
	if err := os.WriteFile(tmpPath, buf.Bytes(), 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, s.taskPath(t.ID))
}

// Load reads a task from a markdown file
func (s *MarkdownStorage) Load(id int) (*task.Task, error) {
	data, err := os.ReadFile(s.taskPath(id))
	if err != nil {
		return nil, err
	}
	return s.parse(data)
}

// Delete removes a task file
func (s *MarkdownStorage) Delete(id int) error {
	return os.Remove(s.taskPath(id))
}

// LoadAll reads all tasks from the directory
func (s *MarkdownStorage) LoadAll() ([]*task.Task, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var tasks []*task.Task
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		// Skip index file
		if entry.Name() == ".index.json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			continue
		}

		t, err := s.parse(data)
		if err != nil {
			continue
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

// parse extracts task from markdown with frontmatter
func (s *MarkdownStorage) parse(data []byte) (*task.Task, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	// Find opening ---
	if !scanner.Scan() || scanner.Text() != "---" {
		return nil, fmt.Errorf("invalid frontmatter: missing opening ---")
	}

	// Collect frontmatter
	var frontmatterBuf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			break
		}
		frontmatterBuf.WriteString(line)
		frontmatterBuf.WriteString("\n")
	}

	// Parse frontmatter
	var fm struct {
		ID        int    `yaml:"id"`
		Title     string `yaml:"title"`
		Status    string `yaml:"status"`
		Priority  string `yaml:"priority"`
		Type      string `yaml:"type"`
		CreatedAt string `yaml:"created_at"`
		UpdatedAt string `yaml:"updated_at"`
	}
	if err := yaml.Unmarshal(frontmatterBuf.Bytes(), &fm); err != nil {
		return nil, err
	}

	// Collect body (description)
	var bodyBuf bytes.Buffer
	for scanner.Scan() {
		if bodyBuf.Len() > 0 {
			bodyBuf.WriteString("\n")
		}
		bodyBuf.WriteString(scanner.Text())
	}

	// Parse timestamps
	createdAt, _ := parseTime(fm.CreatedAt)
	updatedAt, _ := parseTime(fm.UpdatedAt)

	return &task.Task{
		ID:          fm.ID,
		Title:       fm.Title,
		Description: strings.TrimSpace(bodyBuf.String()),
		Status:      task.Status(fm.Status),
		Priority:    task.Priority(fm.Priority),
		Type:        fm.Type,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

// parseTime tries multiple time formats
func parseTime(s string) (t time.Time, err error) {
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err = time.Parse(f, s); err == nil {
			return
		}
	}
	return
}

// NextID returns the next available task ID
func (s *MarkdownStorage) NextID() (int, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}

	maxID := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".md")
		if id, err := strconv.Atoi(name); err == nil && id > maxID {
			maxID = id
		}
	}

	return maxID + 1, nil
}
```

**Step 2: Add missing import**

The file needs the `time` import. Update the imports:
```go
import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gpayer/mcp-task-manager/internal/task"
	"gopkg.in/yaml.v3"
)
```

**Step 3: Verify it compiles**

Run:
```bash
go build ./internal/storage
```

Expected: No errors

**Step 4: Commit**

```bash
git add internal/storage
git commit -m "feat: add markdown file storage"
```

---

## Task 5: JSON Index Cache

**Files:**
- Create: `internal/storage/index.go`

**Step 1: Create index cache**

Create `internal/storage/index.go`:
```go
package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/gpayer/mcp-task-manager/internal/task"
)

// Index is an in-memory cache of all tasks
type Index struct {
	tasks   map[int]*task.Task
	dir     string
	storage *MarkdownStorage
}

// NewIndex creates a new index for the given directory
func NewIndex(dir string, storage *MarkdownStorage) *Index {
	return &Index{
		tasks:   make(map[int]*task.Task),
		dir:     dir,
		storage: storage,
	}
}

// indexPath returns path to the index file
func (idx *Index) indexPath() string {
	return filepath.Join(idx.dir, ".index.json")
}

// Rebuild scans all markdown files and rebuilds the index
func (idx *Index) Rebuild() error {
	tasks, err := idx.storage.LoadAll()
	if err != nil {
		return err
	}

	idx.tasks = make(map[int]*task.Task)
	for _, t := range tasks {
		idx.tasks[t.ID] = t
	}

	return idx.Save()
}

// Save persists the index to disk
func (idx *Index) Save() error {
	if err := os.MkdirAll(idx.dir, 0755); err != nil {
		return err
	}

	tasks := idx.All()
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := idx.indexPath() + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, idx.indexPath())
}

// Load reads the index from disk (or rebuilds if missing/corrupt)
func (idx *Index) Load() error {
	data, err := os.ReadFile(idx.indexPath())
	if err != nil {
		// Index missing, rebuild from files
		return idx.Rebuild()
	}

	var tasks []*task.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		// Index corrupt, rebuild from files
		return idx.Rebuild()
	}

	idx.tasks = make(map[int]*task.Task)
	for _, t := range tasks {
		idx.tasks[t.ID] = t
	}

	return nil
}

// Get returns a task by ID
func (idx *Index) Get(id int) (*task.Task, bool) {
	t, ok := idx.tasks[id]
	return t, ok
}

// Set adds or updates a task in the index
func (idx *Index) Set(t *task.Task) {
	idx.tasks[t.ID] = t
}

// Delete removes a task from the index
func (idx *Index) Delete(id int) {
	delete(idx.tasks, id)
}

// All returns all tasks sorted by ID
func (idx *Index) All() []*task.Task {
	tasks := make([]*task.Task, 0, len(idx.tasks))
	for _, t := range idx.tasks {
		tasks = append(tasks, t)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
	return tasks
}

// Filter returns tasks matching the given criteria
func (idx *Index) Filter(status *task.Status, priority *task.Priority, taskType *string) []*task.Task {
	var result []*task.Task
	for _, t := range idx.tasks {
		if status != nil && t.Status != *status {
			continue
		}
		if priority != nil && t.Priority != *priority {
			continue
		}
		if taskType != nil && t.Type != *taskType {
			continue
		}
		result = append(result, t)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// NextTodo returns the highest priority todo task
func (idx *Index) NextTodo() *task.Task {
	var candidates []*task.Task
	for _, t := range idx.tasks {
		if t.Status == task.StatusTodo {
			candidates = append(candidates, t)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Sort by priority (lower order = higher priority), then by creation date
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Priority.Order() != candidates[j].Priority.Order() {
			return candidates[i].Priority.Order() < candidates[j].Priority.Order()
		}
		return candidates[i].CreatedAt.Before(candidates[j].CreatedAt)
	})

	return candidates[0]
}

// NextID returns the next available task ID
func (idx *Index) NextID() int {
	maxID := 0
	for id := range idx.tasks {
		if id > maxID {
			maxID = id
		}
	}
	return maxID + 1
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./internal/storage
```

Expected: No errors

**Step 3: Commit**

```bash
git add internal/storage/index.go
git commit -m "feat: add JSON index cache"
```

---

## Task 6: Task Service

**Files:**
- Create: `internal/task/service.go`

**Step 1: Create task service**

Create `internal/task/service.go`:
```go
package task

import (
	"fmt"
	"time"
)

// Storage interface for task persistence
type Storage interface {
	Save(t *Task) error
	Load(id int) (*Task, error)
	Delete(id int) error
	EnsureDir() error
}

// Index interface for task indexing
type Index interface {
	Load() error
	Save() error
	Get(id int) (*Task, bool)
	Set(t *Task)
	Delete(id int)
	All() []*Task
	Filter(status *Status, priority *Priority, taskType *string) []*Task
	NextTodo() *Task
	NextID() int
}

// Service provides task management operations
type Service struct {
	storage    Storage
	index      Index
	validTypes []string
}

// NewService creates a new task service
func NewService(storage Storage, index Index, validTypes []string) *Service {
	return &Service{
		storage:    storage,
		index:      index,
		validTypes: validTypes,
	}
}

// Initialize loads the index and ensures storage is ready
func (s *Service) Initialize() error {
	if err := s.storage.EnsureDir(); err != nil {
		return err
	}
	return s.index.Load()
}

// Create creates a new task
func (s *Service) Create(title, description string, priority Priority, taskType string) (*Task, error) {
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if !IsValidPriority(string(priority)) {
		return nil, fmt.Errorf("invalid priority: %s", priority)
	}
	if !s.isValidType(taskType) {
		return nil, fmt.Errorf("invalid task type: %s", taskType)
	}

	now := time.Now().UTC()
	t := &Task{
		ID:          s.index.NextID(),
		Title:       title,
		Description: description,
		Status:      StatusTodo,
		Priority:    priority,
		Type:        taskType,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.storage.Save(t); err != nil {
		return nil, err
	}

	s.index.Set(t)
	if err := s.index.Save(); err != nil {
		return nil, err
	}

	return t, nil
}

// Get returns a task by ID
func (s *Service) Get(id int) (*Task, error) {
	t, ok := s.index.Get(id)
	if !ok {
		return nil, fmt.Errorf("task not found: %d", id)
	}
	return t, nil
}

// Update modifies a task
func (s *Service) Update(id int, title, description *string, status *Status, priority *Priority, taskType *string) (*Task, error) {
	t, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	if title != nil {
		if *title == "" {
			return nil, fmt.Errorf("title cannot be empty")
		}
		t.Title = *title
	}
	if description != nil {
		t.Description = *description
	}
	if status != nil {
		if !IsValidStatus(string(*status)) {
			return nil, fmt.Errorf("invalid status: %s", *status)
		}
		t.Status = *status
	}
	if priority != nil {
		if !IsValidPriority(string(*priority)) {
			return nil, fmt.Errorf("invalid priority: %s", *priority)
		}
		t.Priority = *priority
	}
	if taskType != nil {
		if !s.isValidType(*taskType) {
			return nil, fmt.Errorf("invalid task type: %s", *taskType)
		}
		t.Type = *taskType
	}

	t.UpdatedAt = time.Now().UTC()

	if err := s.storage.Save(t); err != nil {
		return nil, err
	}

	s.index.Set(t)
	if err := s.index.Save(); err != nil {
		return nil, err
	}

	return t, nil
}

// Delete removes a task
func (s *Service) Delete(id int) error {
	if _, err := s.Get(id); err != nil {
		return err
	}

	if err := s.storage.Delete(id); err != nil {
		return err
	}

	s.index.Delete(id)
	return s.index.Save()
}

// List returns all tasks, optionally filtered
func (s *Service) List(status *Status, priority *Priority, taskType *string) []*Task {
	return s.index.Filter(status, priority, taskType)
}

// GetNextTask returns the highest priority todo task
func (s *Service) GetNextTask() *Task {
	return s.index.NextTodo()
}

// StartTask moves a task from todo to in_progress
func (s *Service) StartTask(id int) (*Task, error) {
	t, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	if t.Status != StatusTodo {
		return nil, fmt.Errorf("task %d is not in todo status (current: %s)", id, t.Status)
	}

	status := StatusInProgress
	return s.Update(id, nil, nil, &status, nil, nil)
}

// CompleteTask moves a task from in_progress to done
func (s *Service) CompleteTask(id int) (*Task, error) {
	t, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	if t.Status != StatusInProgress {
		return nil, fmt.Errorf("task %d is not in progress (current: %s)", id, t.Status)
	}

	status := StatusDone
	return s.Update(id, nil, nil, &status, nil, nil)
}

// isValidType checks if task type is valid
func (s *Service) isValidType(t string) bool {
	for _, valid := range s.validTypes {
		if t == valid {
			return true
		}
	}
	return false
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./internal/task
```

Expected: No errors

**Step 3: Commit**

```bash
git add internal/task/service.go
git commit -m "feat: add task service with business logic"
```

---

## Task 7: MCP Tool Handlers - Management Tools

**Files:**
- Create: `internal/tools/tools.go`
- Create: `internal/tools/management.go`

**Step 1: Create tools registration**

Create `internal/tools/tools.go`:
```go
package tools

import (
	"github.com/gpayer/mcp-task-manager/internal/task"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Register registers all MCP tools with the server
func Register(s *server.MCPServer, svc *task.Service, validTypes []string) {
	registerManagementTools(s, svc, validTypes)
	registerWorkflowTools(s, svc)
}
```

**Step 2: Create management tools**

Create `internal/tools/management.go`:
```go
package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gpayer/mcp-task-manager/internal/task"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerManagementTools(s *server.MCPServer, svc *task.Service, validTypes []string) {
	// create_task
	createTool := mcp.NewTool("create_task",
		mcp.WithDescription("Create a new task"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Task title"),
		),
		mcp.WithString("description",
			mcp.Description("Task description (markdown supported)"),
		),
		mcp.WithString("priority",
			mcp.Required(),
			mcp.Description("Task priority"),
			mcp.Enum("critical", "high", "medium", "low"),
		),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Task type"),
			mcp.Enum(validTypes...),
		),
	)
	s.AddTool(createTool, createTaskHandler(svc))

	// get_task
	getTool := mcp.NewTool("get_task",
		mcp.WithDescription("Get a task by ID"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
	)
	s.AddTool(getTool, getTaskHandler(svc))

	// update_task
	updateTool := mcp.NewTool("update_task",
		mcp.WithDescription("Update an existing task"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
		mcp.WithString("title",
			mcp.Description("New title"),
		),
		mcp.WithString("description",
			mcp.Description("New description"),
		),
		mcp.WithString("status",
			mcp.Description("New status"),
			mcp.Enum("todo", "in_progress", "done"),
		),
		mcp.WithString("priority",
			mcp.Description("New priority"),
			mcp.Enum("critical", "high", "medium", "low"),
		),
		mcp.WithString("type",
			mcp.Description("New task type"),
			mcp.Enum(validTypes...),
		),
	)
	s.AddTool(updateTool, updateTaskHandler(svc))

	// delete_task
	deleteTool := mcp.NewTool("delete_task",
		mcp.WithDescription("Delete a task"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
	)
	s.AddTool(deleteTool, deleteTaskHandler(svc))

	// list_tasks
	listTool := mcp.NewTool("list_tasks",
		mcp.WithDescription("List tasks with optional filters"),
		mcp.WithString("status",
			mcp.Description("Filter by status"),
			mcp.Enum("todo", "in_progress", "done"),
		),
		mcp.WithString("priority",
			mcp.Description("Filter by priority"),
			mcp.Enum("critical", "high", "medium", "low"),
		),
		mcp.WithString("type",
			mcp.Description("Filter by task type"),
			mcp.Enum(validTypes...),
		),
	)
	s.AddTool(listTool, listTasksHandler(svc))
}

func createTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title := req.Params.Arguments["title"].(string)
		description := ""
		if d, ok := req.Params.Arguments["description"].(string); ok {
			description = d
		}
		priority := task.Priority(req.Params.Arguments["priority"].(string))
		taskType := req.Params.Arguments["type"].(string)

		t, err := svc.Create(title, description, priority, taskType)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return taskResult(t)
	}
}

func getTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := int(req.Params.Arguments["id"].(float64))

		t, err := svc.Get(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return taskResult(t)
	}
}

func updateTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := int(req.Params.Arguments["id"].(float64))

		var title, description, taskType *string
		var status *task.Status
		var priority *task.Priority

		if v, ok := req.Params.Arguments["title"].(string); ok {
			title = &v
		}
		if v, ok := req.Params.Arguments["description"].(string); ok {
			description = &v
		}
		if v, ok := req.Params.Arguments["status"].(string); ok {
			s := task.Status(v)
			status = &s
		}
		if v, ok := req.Params.Arguments["priority"].(string); ok {
			p := task.Priority(v)
			priority = &p
		}
		if v, ok := req.Params.Arguments["type"].(string); ok {
			taskType = &v
		}

		t, err := svc.Update(id, title, description, status, priority, taskType)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return taskResult(t)
	}
}

func deleteTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := int(req.Params.Arguments["id"].(float64))

		if err := svc.Delete(id); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Task %d deleted", id)), nil
	}
}

func listTasksHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var status *task.Status
		var priority *task.Priority
		var taskType *string

		if v, ok := req.Params.Arguments["status"].(string); ok {
			s := task.Status(v)
			status = &s
		}
		if v, ok := req.Params.Arguments["priority"].(string); ok {
			p := task.Priority(v)
			priority = &p
		}
		if v, ok := req.Params.Arguments["type"].(string); ok {
			taskType = &v
		}

		tasks := svc.List(status, priority, taskType)

		if len(tasks) == 0 {
			return mcp.NewToolResultText("No tasks found"), nil
		}

		data, err := json.MarshalIndent(tasks, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}

func taskResult(t *task.Task) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}
```

**Step 3: Verify it compiles**

Run:
```bash
go build ./internal/tools
```

Expected: No errors

**Step 4: Commit**

```bash
git add internal/tools
git commit -m "feat: add MCP management tools (create, get, update, delete, list)"
```

---

## Task 8: MCP Tool Handlers - Workflow Tools

**Files:**
- Create: `internal/tools/workflow.go`

**Step 1: Create workflow tools**

Create `internal/tools/workflow.go`:
```go
package tools

import (
	"context"

	"github.com/gpayer/mcp-task-manager/internal/task"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerWorkflowTools(s *server.MCPServer, svc *task.Service) {
	// get_next_task
	nextTool := mcp.NewTool("get_next_task",
		mcp.WithDescription("Get the highest priority todo task for an agent to work on"),
	)
	s.AddTool(nextTool, getNextTaskHandler(svc))

	// start_task
	startTool := mcp.NewTool("start_task",
		mcp.WithDescription("Move a task from todo to in_progress"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("Task ID to start"),
		),
	)
	s.AddTool(startTool, startTaskHandler(svc))

	// complete_task
	completeTool := mcp.NewTool("complete_task",
		mcp.WithDescription("Move a task from in_progress to done"),
		mcp.WithNumber("id",
			mcp.Required(),
			mcp.Description("Task ID to complete"),
		),
	)
	s.AddTool(completeTool, completeTaskHandler(svc))
}

func getNextTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		t := svc.GetNextTask()
		if t == nil {
			return mcp.NewToolResultText("No tasks available"), nil
		}
		return taskResult(t)
	}
}

func startTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := int(req.Params.Arguments["id"].(float64))

		t, err := svc.StartTask(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return taskResult(t)
	}
}

func completeTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := int(req.Params.Arguments["id"].(float64))

		t, err := svc.CompleteTask(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return taskResult(t)
	}
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./internal/tools
```

Expected: No errors

**Step 3: Commit**

```bash
git add internal/tools/workflow.go
git commit -m "feat: add MCP workflow tools (get_next_task, start_task, complete_task)"
```

---

## Task 9: Wire Everything Together in Main

**Files:**
- Modify: `cmd/mcp-task-manager/main.go`

**Step 1: Update main.go with full wiring**

Replace `cmd/mcp-task-manager/main.go` with:
```go
package main

import (
	"log"

	"github.com/gpayer/mcp-task-manager/internal/config"
	"github.com/gpayer/mcp-task-manager/internal/storage"
	"github.com/gpayer/mcp-task-manager/internal/task"
	"github.com/gpayer/mcp-task-manager/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize storage
	tasksDir := cfg.TasksDir()
	mdStorage := storage.NewMarkdownStorage(tasksDir)
	index := storage.NewIndex(tasksDir, mdStorage)

	// Initialize task service
	svc := task.NewService(mdStorage, index, cfg.TaskTypes)
	if err := svc.Initialize(); err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	// Create MCP server
	s := server.NewMCPServer(
		"mcp-task-manager",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	// Register tools
	tools.Register(s, svc, cfg.TaskTypes)

	// Start server
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./cmd/mcp-task-manager
```

Expected: No errors, binary created

**Step 3: Commit**

```bash
git add cmd/mcp-task-manager/main.go
git commit -m "feat: wire up all components in main"
```

---

## Task 10: Manual Integration Test

**Step 1: Build the server**

Run:
```bash
go build -o mcp-task-manager ./cmd/mcp-task-manager
```

Expected: Binary created

**Step 2: Test with MCP inspector (optional)**

If you have mcp-inspector or similar tool, test the server. Otherwise, proceed to add to Claude config.

**Step 3: Add to Claude Desktop config**

Add to `~/.config/claude-desktop/claude_desktop_config.json` (or appropriate config location):
```json
{
  "mcpServers": {
    "task-manager": {
      "command": "/home/gernot/src/mcp-task-manager/mcp-task-manager"
    }
  }
}
```

**Step 4: Verify tools appear in Claude**

Restart Claude Desktop and verify the 8 tools appear:
- create_task
- get_task
- update_task
- delete_task
- list_tasks
- get_next_task
- start_task
- complete_task

**Step 5: Test basic workflow**

In Claude, test:
1. Create a task
2. List tasks
3. Get the task
4. Start the task
5. Complete the task

**Step 6: Commit final state**

```bash
git add .
git commit -m "feat: complete MVP implementation"
```

---

## Summary

This plan creates a functional MCP Task Manager MVP with:

- 8 MCP tools (5 management + 3 workflow)
- Markdown file storage with YAML frontmatter
- JSON index cache (self-healing)
- Configurable task types
- Clean separation of concerns (config, storage, task service, tools)

Total: 10 tasks, approximately 30-40 implementation steps.
