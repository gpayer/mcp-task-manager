# CLI Tool Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add CLI subcommands to mcp-task-manager while maintaining backwards-compatible MCP server mode.

**Architecture:** CLI mode is triggered when `len(os.Args) > 1`. New `internal/cli` package handles all CLI logic using flaggy for argument parsing. Reuses existing `task.Service` for all business logic.

**Tech Stack:** Go, flaggy (CLI parsing), existing task.Service

---

## Task 1: Add flaggy dependency

**Files:**
- Modify: `go.mod`

**Step 1: Add flaggy dependency**

Run:
```bash
cd /home/gernot/src/mcp-task-manager && go get github.com/integrii/flaggy
```

**Step 2: Verify dependency added**

Run:
```bash
grep flaggy go.mod
```

Expected: Line showing `github.com/integrii/flaggy`

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add flaggy for CLI argument parsing"
```

---

## Task 2: Create CLI output helpers

**Files:**
- Create: `internal/cli/output.go`
- Create: `internal/cli/output_test.go`

**Step 1: Write the test file**

Create `internal/cli/output_test.go`:

```go
package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/gpayer/mcp-task-manager/internal/task"
)

func TestFormatTaskDetail(t *testing.T) {
	created := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	updated := time.Date(2025, 1, 15, 11, 0, 0, 0, time.UTC)
	tk := &task.Task{
		ID:          1,
		Title:       "Test task",
		Description: "A test description",
		Status:      task.StatusTodo,
		Priority:    task.PriorityHigh,
		Type:        "feature",
		CreatedAt:   created,
		UpdatedAt:   updated,
	}

	output := FormatTaskDetail(tk)

	if !strings.Contains(output, "Task #1") {
		t.Error("expected Task #1 in output")
	}
	if !strings.Contains(output, "Test task") {
		t.Error("expected title in output")
	}
	if !strings.Contains(output, "todo") {
		t.Error("expected status in output")
	}
	if !strings.Contains(output, "high") {
		t.Error("expected priority in output")
	}
	if !strings.Contains(output, "A test description") {
		t.Error("expected description in output")
	}
}

func TestFormatTaskTable(t *testing.T) {
	tasks := []*task.Task{
		{ID: 1, Title: "First task", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature"},
		{ID: 2, Title: "Second task", Status: task.StatusDone, Priority: task.PriorityLow, Type: "bug"},
	}

	output := FormatTaskTable(tasks)

	if !strings.Contains(output, "ID") {
		t.Error("expected header row")
	}
	if !strings.Contains(output, "First task") {
		t.Error("expected first task")
	}
	if !strings.Contains(output, "Second task") {
		t.Error("expected second task")
	}
}

func TestFormatTaskTableEmpty(t *testing.T) {
	output := FormatTaskTable([]*task.Task{})
	if !strings.Contains(output, "No tasks") {
		t.Error("expected 'No tasks' message for empty list")
	}
}

func TestFormatMessage(t *testing.T) {
	output := FormatMessage("Task #3 deleted.", 3)
	if output != "Task #3 deleted." {
		t.Errorf("expected 'Task #3 deleted.', got %q", output)
	}
}

func TestFormatJSON(t *testing.T) {
	tk := &task.Task{ID: 1, Title: "Test", Status: task.StatusTodo, Priority: task.PriorityMedium, Type: "feature"}
	var buf bytes.Buffer
	err := FormatJSON(&buf, tk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), `"id":1`) {
		t.Error("expected JSON with id:1")
	}
}

func TestFormatJSONMessage(t *testing.T) {
	var buf bytes.Buffer
	err := FormatJSONMessage(&buf, "Task deleted", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, `"message"`) {
		t.Error("expected message field")
	}
	if !strings.Contains(output, `"id":5`) {
		t.Error("expected id field")
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/cli/... -v
```

Expected: FAIL (package doesn't exist yet)

**Step 3: Write the implementation**

Create `internal/cli/output.go`:

```go
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/gpayer/mcp-task-manager/internal/task"
)

// FormatTaskDetail formats a single task for human-readable output
func FormatTaskDetail(t *task.Task) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Task #%d\n", t.ID))
	sb.WriteString(fmt.Sprintf("Title:       %s\n", t.Title))
	sb.WriteString(fmt.Sprintf("Status:      %s\n", t.Status))
	sb.WriteString(fmt.Sprintf("Priority:    %s\n", t.Priority))
	sb.WriteString(fmt.Sprintf("Type:        %s\n", t.Type))
	sb.WriteString(fmt.Sprintf("Created:     %s\n", t.CreatedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Updated:     %s\n", t.UpdatedAt.Format("2006-01-02 15:04:05")))
	if t.Description != "" {
		sb.WriteString(fmt.Sprintf("\nDescription:\n%s\n", t.Description))
	}
	return sb.String()
}

// FormatTaskTable formats a list of tasks as a table
func FormatTaskTable(tasks []*task.Task) string {
	if len(tasks) == 0 {
		return "No tasks found."
	}

	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tTitle\tStatus\tPriority\tType")
	for _, t := range tasks {
		title := t.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", t.ID, title, t.Status, t.Priority, t.Type)
	}
	w.Flush()
	return sb.String()
}

// FormatMessage formats a simple message
func FormatMessage(msg string, id int) string {
	return msg
}

// FormatJSON writes a value as JSON to the writer
func FormatJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// FormatJSONMessage writes a message with ID as JSON
func FormatJSONMessage(w io.Writer, msg string, id int) error {
	return FormatJSON(w, map[string]interface{}{
		"message": msg,
		"id":      id,
	})
}
```

**Step 4: Run tests to verify they pass**

Run:
```bash
go test ./internal/cli/... -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli/
git commit -m "feat(cli): add output formatting helpers"
```

---

## Task 3: Create CLI main structure with version command

**Files:**
- Create: `internal/cli/cli.go`
- Create: `internal/cli/cli_test.go`

**Step 1: Write test for version command**

Create `internal/cli/cli_test.go`:

```go
package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunWithArgs([]string{"mcp-task-manager", "version"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "mcp-task-manager") {
		t.Error("expected version output to contain 'mcp-task-manager'")
	}
}

func TestHelpFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunWithArgs([]string{"mcp-task-manager", "--help"}, &stdout, &stderr)

	// flaggy exits with 0 on --help
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/cli/... -v -run TestVersion
```

Expected: FAIL (RunWithArgs doesn't exist)

**Step 3: Write the implementation**

Create `internal/cli/cli.go`:

```go
package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/integrii/flaggy"
)

// Version is set at build time
var Version = "dev"

// Run executes the CLI with os.Args
func Run() {
	code := RunWithArgs(os.Args, os.Stdout, os.Stderr)
	os.Exit(code)
}

// RunWithArgs executes the CLI with given arguments (for testing)
func RunWithArgs(args []string, stdout, stderr io.Writer) int {
	// Reset flaggy for fresh parsing
	flaggy.ResetParser()

	flaggy.SetName("mcp-task-manager")
	flaggy.SetDescription("Task manager for Claude and coding agents")
	flaggy.SetVersion(Version)

	// Version subcommand
	versionCmd := flaggy.NewSubcommand("version")
	versionCmd.Description = "Show version information"
	flaggy.AttachSubcommand(versionCmd, 1)

	// Parse with custom args
	flaggy.ParseArgs(args[1:])

	// Handle subcommands
	if versionCmd.Used {
		fmt.Fprintf(stdout, "mcp-task-manager %s\n", Version)
		return 0
	}

	return 0
}
```

**Step 4: Run tests to verify they pass**

Run:
```bash
go test ./internal/cli/... -v -run TestVersion
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli/
git commit -m "feat(cli): add CLI structure with version command"
```

---

## Task 4: Add list command

**Files:**
- Modify: `internal/cli/cli.go`
- Modify: `internal/cli/cli_test.go`
- Create: `internal/cli/commands.go`

**Step 1: Write test for list command**

Add to `internal/cli/cli_test.go`:

```go
func TestListCommand(t *testing.T) {
	// Create temp directory for tasks
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunWithArgs([]string{"mcp-task-manager", "list"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
	}
	// Empty list should show "No tasks"
	if !strings.Contains(stdout.String(), "No tasks") {
		t.Errorf("expected 'No tasks' message, got: %s", stdout.String())
	}
}

func TestListCommandJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunWithArgs([]string{"mcp-task-manager", "list", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "[") {
		t.Error("expected JSON array output")
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/cli/... -v -run TestListCommand
```

Expected: FAIL

**Step 3: Write commands.go**

Create `internal/cli/commands.go`:

```go
package cli

import (
	"fmt"
	"io"

	"github.com/gpayer/mcp-task-manager/internal/config"
	"github.com/gpayer/mcp-task-manager/internal/storage"
	"github.com/gpayer/mcp-task-manager/internal/task"
)

// initService initializes the task service
func initService() (*task.Service, *config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	tasksDir := cfg.TasksDir()
	mdStorage := storage.NewMarkdownStorage(tasksDir)
	index := storage.NewIndex(tasksDir, mdStorage)
	svc := task.NewService(mdStorage, index, cfg.TaskTypes)

	if err := svc.Initialize(); err != nil {
		return nil, nil, fmt.Errorf("failed to initialize: %w", err)
	}

	return svc, cfg, nil
}

// cmdList handles the list command
func cmdList(stdout, stderr io.Writer, jsonOutput bool, status, priority, taskType string) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	var statusPtr *task.Status
	var priorityPtr *task.Priority
	var typePtr *string

	if status != "" {
		s := task.Status(status)
		statusPtr = &s
	}
	if priority != "" {
		p := task.Priority(priority)
		priorityPtr = &p
	}
	if taskType != "" {
		typePtr = &taskType
	}

	tasks := svc.List(statusPtr, priorityPtr, typePtr)

	if jsonOutput {
		if err := FormatJSON(stdout, tasks); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
	} else {
		fmt.Fprint(stdout, FormatTaskTable(tasks))
	}

	return 0
}
```

**Step 4: Update cli.go to add list command**

Update `internal/cli/cli.go` - add after version subcommand setup:

```go
// Add these variables at the top of RunWithArgs, after the flaggy setup
var jsonOutput bool

// List subcommand
listCmd := flaggy.NewSubcommand("list")
listCmd.Description = "List tasks with optional filters"
var listStatus, listPriority, listType string
listCmd.String(&listStatus, "s", "status", "Filter by status (todo|in_progress|done)")
listCmd.String(&listPriority, "p", "priority", "Filter by priority (critical|high|medium|low)")
listCmd.String(&listType, "t", "type", "Filter by type")
listCmd.Bool(&jsonOutput, "j", "json", "Output as JSON")
flaggy.AttachSubcommand(listCmd, 1)
```

And add handler after version check:

```go
if listCmd.Used {
    return cmdList(stdout, stderr, jsonOutput, listStatus, listPriority, listType)
}
```

**Step 5: Run tests to verify they pass**

Run:
```bash
go test ./internal/cli/... -v -run TestListCommand
```

Expected: PASS

**Step 6: Commit**

```bash
git add internal/cli/
git commit -m "feat(cli): add list command with filters"
```

---

## Task 5: Add get and next commands

**Files:**
- Modify: `internal/cli/cli.go`
- Modify: `internal/cli/cli_test.go`
- Modify: `internal/cli/commands.go`

**Step 1: Write tests**

Add to `internal/cli/cli_test.go`:

```go
func TestGetCommandNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunWithArgs([]string{"mcp-task-manager", "get", "999"}, &stdout, &stderr)

	if code != 1 {
		t.Errorf("expected exit code 1 for not found, got %d", code)
	}
	if !strings.Contains(stderr.String(), "not found") {
		t.Errorf("expected 'not found' error, got: %s", stderr.String())
	}
}

func TestNextCommandNoTasks(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunWithArgs([]string{"mcp-task-manager", "next"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(stdout.String(), "No tasks available") {
		t.Errorf("expected 'No tasks available', got: %s", stdout.String())
	}
}
```

**Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/cli/... -v -run "TestGetCommand|TestNextCommand"
```

Expected: FAIL

**Step 3: Add commands to commands.go**

Add to `internal/cli/commands.go`:

```go
// cmdGet handles the get command
func cmdGet(stdout, stderr io.Writer, jsonOutput bool, id int) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	t, err := svc.Get(id)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if jsonOutput {
		if err := FormatJSON(stdout, t); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
	} else {
		fmt.Fprint(stdout, FormatTaskDetail(t))
	}

	return 0
}

// cmdNext handles the next command
func cmdNext(stdout, stderr io.Writer, jsonOutput bool) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	t := svc.GetNextTask()
	if t == nil {
		if jsonOutput {
			FormatJSON(stdout, map[string]string{"message": "No tasks available"})
		} else {
			fmt.Fprintln(stdout, "No tasks available.")
		}
		return 0
	}

	if jsonOutput {
		if err := FormatJSON(stdout, t); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
	} else {
		fmt.Fprint(stdout, FormatTaskDetail(t))
	}

	return 0
}
```

**Step 4: Add subcommands to cli.go**

Add to `internal/cli/cli.go`:

```go
// Get subcommand
getCmd := flaggy.NewSubcommand("get")
getCmd.Description = "Get task details by ID"
var getID int
getCmd.AddPositionalValue(&getID, "id", 1, true, "Task ID")
var getJSON bool
getCmd.Bool(&getJSON, "j", "json", "Output as JSON")
flaggy.AttachSubcommand(getCmd, 1)

// Next subcommand
nextCmd := flaggy.NewSubcommand("next")
nextCmd.Description = "Get highest priority todo task"
var nextJSON bool
nextCmd.Bool(&nextJSON, "j", "json", "Output as JSON")
flaggy.AttachSubcommand(nextCmd, 1)
```

And handlers:

```go
if getCmd.Used {
    return cmdGet(stdout, stderr, getJSON, getID)
}
if nextCmd.Used {
    return cmdNext(stdout, stderr, nextJSON)
}
```

**Step 5: Run tests**

Run:
```bash
go test ./internal/cli/... -v
```

Expected: PASS

**Step 6: Commit**

```bash
git add internal/cli/
git commit -m "feat(cli): add get and next commands"
```

---

## Task 6: Add create command

**Files:**
- Modify: `internal/cli/cli.go`
- Modify: `internal/cli/cli_test.go`
- Modify: `internal/cli/commands.go`

**Step 1: Write test**

Add to `internal/cli/cli_test.go`:

```go
func TestCreateCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunWithArgs([]string{"mcp-task-manager", "create", "My new task"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "My new task") {
		t.Errorf("expected task title in output, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "medium") {
		t.Error("expected default priority 'medium'")
	}
}

func TestCreateCommandWithFlags(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	var stdout, stderr bytes.Buffer
	code := RunWithArgs([]string{"mcp-task-manager", "create", "Bug fix", "-p", "high", "-t", "bug", "-d", "Fix the login"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "high") {
		t.Error("expected priority 'high'")
	}
	if !strings.Contains(stdout.String(), "bug") {
		t.Error("expected type 'bug'")
	}
}
```

**Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/cli/... -v -run TestCreateCommand
```

Expected: FAIL

**Step 3: Add command to commands.go**

Add to `internal/cli/commands.go`:

```go
// cmdCreate handles the create command
func cmdCreate(stdout, stderr io.Writer, jsonOutput bool, title, priority, taskType, description string) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	t, err := svc.Create(title, description, task.Priority(priority), taskType)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if jsonOutput {
		if err := FormatJSON(stdout, t); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
	} else {
		fmt.Fprint(stdout, FormatTaskDetail(t))
	}

	return 0
}
```

**Step 4: Add subcommand to cli.go**

```go
// Create subcommand
createCmd := flaggy.NewSubcommand("create")
createCmd.Description = "Create a new task"
var createTitle string
var createPriority = "medium"
var createType = "feature"
var createDesc string
var createJSON bool
createCmd.AddPositionalValue(&createTitle, "title", 1, true, "Task title")
createCmd.String(&createPriority, "p", "priority", "Priority (default: medium)")
createCmd.String(&createType, "t", "type", "Type (default: feature)")
createCmd.String(&createDesc, "d", "description", "Task description")
createCmd.Bool(&createJSON, "j", "json", "Output as JSON")
flaggy.AttachSubcommand(createCmd, 1)
```

Handler:

```go
if createCmd.Used {
    return cmdCreate(stdout, stderr, createJSON, createTitle, createPriority, createType, createDesc)
}
```

**Step 5: Run tests**

Run:
```bash
go test ./internal/cli/... -v -run TestCreateCommand
```

Expected: PASS

**Step 6: Commit**

```bash
git add internal/cli/
git commit -m "feat(cli): add create command"
```

---

## Task 7: Add update command

**Files:**
- Modify: `internal/cli/cli.go`
- Modify: `internal/cli/cli_test.go`
- Modify: `internal/cli/commands.go`

**Step 1: Write test**

Add to `internal/cli/cli_test.go`:

```go
func TestUpdateCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	// First create a task
	var stdout, stderr bytes.Buffer
	RunWithArgs([]string{"mcp-task-manager", "create", "Original title"}, &stdout, &stderr)

	// Then update it
	stdout.Reset()
	stderr.Reset()
	code := RunWithArgs([]string{"mcp-task-manager", "update", "1", "--title", "Updated title"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Updated title") {
		t.Errorf("expected updated title in output, got: %s", stdout.String())
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/cli/... -v -run TestUpdateCommand
```

Expected: FAIL

**Step 3: Add command to commands.go**

Add to `internal/cli/commands.go`:

```go
// cmdUpdate handles the update command
func cmdUpdate(stdout, stderr io.Writer, jsonOutput bool, id int, title, status, priority, taskType, description string) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	var titlePtr, descPtr, typePtr *string
	var statusPtr *task.Status
	var priorityPtr *task.Priority

	if title != "" {
		titlePtr = &title
	}
	if description != "" {
		descPtr = &description
	}
	if status != "" {
		s := task.Status(status)
		statusPtr = &s
	}
	if priority != "" {
		p := task.Priority(priority)
		priorityPtr = &p
	}
	if taskType != "" {
		typePtr = &taskType
	}

	t, err := svc.Update(id, titlePtr, descPtr, statusPtr, priorityPtr, typePtr)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if jsonOutput {
		if err := FormatJSON(stdout, t); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
	} else {
		fmt.Fprint(stdout, FormatTaskDetail(t))
	}

	return 0
}
```

**Step 4: Add subcommand to cli.go**

```go
// Update subcommand
updateCmd := flaggy.NewSubcommand("update")
updateCmd.Description = "Update an existing task"
var updateID int
var updateTitle, updateStatus, updatePriority, updateType, updateDesc string
var updateJSON bool
updateCmd.AddPositionalValue(&updateID, "id", 1, true, "Task ID")
updateCmd.String(&updateTitle, "", "title", "New title")
updateCmd.String(&updateStatus, "s", "status", "New status")
updateCmd.String(&updatePriority, "p", "priority", "New priority")
updateCmd.String(&updateType, "t", "type", "New type")
updateCmd.String(&updateDesc, "d", "description", "New description")
updateCmd.Bool(&updateJSON, "j", "json", "Output as JSON")
flaggy.AttachSubcommand(updateCmd, 1)
```

Handler:

```go
if updateCmd.Used {
    return cmdUpdate(stdout, stderr, updateJSON, updateID, updateTitle, updateStatus, updatePriority, updateType, updateDesc)
}
```

**Step 5: Run tests**

Run:
```bash
go test ./internal/cli/... -v -run TestUpdateCommand
```

Expected: PASS

**Step 6: Commit**

```bash
git add internal/cli/
git commit -m "feat(cli): add update command"
```

---

## Task 8: Add delete command

**Files:**
- Modify: `internal/cli/cli.go`
- Modify: `internal/cli/cli_test.go`
- Modify: `internal/cli/commands.go`

**Step 1: Write test**

Add to `internal/cli/cli_test.go`:

```go
func TestDeleteCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	// Create a task first
	var stdout, stderr bytes.Buffer
	RunWithArgs([]string{"mcp-task-manager", "create", "To be deleted"}, &stdout, &stderr)

	// Delete it
	stdout.Reset()
	stderr.Reset()
	code := RunWithArgs([]string{"mcp-task-manager", "delete", "1"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "deleted") {
		t.Errorf("expected 'deleted' message, got: %s", stdout.String())
	}

	// Verify it's gone
	stdout.Reset()
	stderr.Reset()
	code = RunWithArgs([]string{"mcp-task-manager", "get", "1"}, &stdout, &stderr)
	if code != 1 {
		t.Error("expected task to be deleted")
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/cli/... -v -run TestDeleteCommand
```

Expected: FAIL

**Step 3: Add command to commands.go**

Add to `internal/cli/commands.go`:

```go
// cmdDelete handles the delete command
func cmdDelete(stdout, stderr io.Writer, jsonOutput bool, id int) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if err := svc.Delete(id); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	msg := fmt.Sprintf("Task #%d deleted.", id)
	if jsonOutput {
		FormatJSONMessage(stdout, msg, id)
	} else {
		fmt.Fprintln(stdout, msg)
	}

	return 0
}
```

**Step 4: Add subcommand to cli.go**

```go
// Delete subcommand
deleteCmd := flaggy.NewSubcommand("delete")
deleteCmd.Description = "Delete a task"
var deleteID int
var deleteJSON bool
deleteCmd.AddPositionalValue(&deleteID, "id", 1, true, "Task ID")
deleteCmd.Bool(&deleteJSON, "j", "json", "Output as JSON")
flaggy.AttachSubcommand(deleteCmd, 1)
```

Handler:

```go
if deleteCmd.Used {
    return cmdDelete(stdout, stderr, deleteJSON, deleteID)
}
```

**Step 5: Run tests**

Run:
```bash
go test ./internal/cli/... -v -run TestDeleteCommand
```

Expected: PASS

**Step 6: Commit**

```bash
git add internal/cli/
git commit -m "feat(cli): add delete command"
```

---

## Task 9: Add start and complete commands

**Files:**
- Modify: `internal/cli/cli.go`
- Modify: `internal/cli/cli_test.go`
- Modify: `internal/cli/commands.go`

**Step 1: Write tests**

Add to `internal/cli/cli_test.go`:

```go
func TestStartCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	// Create a task
	var stdout, stderr bytes.Buffer
	RunWithArgs([]string{"mcp-task-manager", "create", "Task to start"}, &stdout, &stderr)

	// Start it
	stdout.Reset()
	stderr.Reset()
	code := RunWithArgs([]string{"mcp-task-manager", "start", "1"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "started") {
		t.Errorf("expected 'started' message, got: %s", stdout.String())
	}
}

func TestCompleteCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	// Create and start a task
	var stdout, stderr bytes.Buffer
	RunWithArgs([]string{"mcp-task-manager", "create", "Task to complete"}, &stdout, &stderr)
	RunWithArgs([]string{"mcp-task-manager", "start", "1"}, &stdout, &stderr)

	// Complete it
	stdout.Reset()
	stderr.Reset()
	code := RunWithArgs([]string{"mcp-task-manager", "complete", "1"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "completed") {
		t.Errorf("expected 'completed' message, got: %s", stdout.String())
	}
}
```

**Step 2: Run tests to verify they fail**

Run:
```bash
go test ./internal/cli/... -v -run "TestStartCommand|TestCompleteCommand"
```

Expected: FAIL

**Step 3: Add commands to commands.go**

Add to `internal/cli/commands.go`:

```go
// cmdStart handles the start command
func cmdStart(stdout, stderr io.Writer, jsonOutput bool, id int) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if _, err := svc.StartTask(id); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	msg := fmt.Sprintf("Task #%d started.", id)
	if jsonOutput {
		FormatJSONMessage(stdout, msg, id)
	} else {
		fmt.Fprintln(stdout, msg)
	}

	return 0
}

// cmdComplete handles the complete command
func cmdComplete(stdout, stderr io.Writer, jsonOutput bool, id int) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if _, err := svc.CompleteTask(id); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	msg := fmt.Sprintf("Task #%d completed.", id)
	if jsonOutput {
		FormatJSONMessage(stdout, msg, id)
	} else {
		fmt.Fprintln(stdout, msg)
	}

	return 0
}
```

**Step 4: Add subcommands to cli.go**

```go
// Start subcommand
startCmd := flaggy.NewSubcommand("start")
startCmd.Description = "Start a task (todo -> in_progress)"
var startID int
var startJSON bool
startCmd.AddPositionalValue(&startID, "id", 1, true, "Task ID")
startCmd.Bool(&startJSON, "j", "json", "Output as JSON")
flaggy.AttachSubcommand(startCmd, 1)

// Complete subcommand
completeCmd := flaggy.NewSubcommand("complete")
completeCmd.Description = "Complete a task (in_progress -> done)"
var completeID int
var completeJSON bool
completeCmd.AddPositionalValue(&completeID, "id", 1, true, "Task ID")
completeCmd.Bool(&completeJSON, "j", "json", "Output as JSON")
flaggy.AttachSubcommand(completeCmd, 1)
```

Handlers:

```go
if startCmd.Used {
    return cmdStart(stdout, stderr, startJSON, startID)
}
if completeCmd.Used {
    return cmdComplete(stdout, stderr, completeJSON, completeID)
}
```

**Step 5: Run tests**

Run:
```bash
go test ./internal/cli/... -v
```

Expected: PASS

**Step 6: Commit**

```bash
git add internal/cli/
git commit -m "feat(cli): add start and complete commands"
```

---

## Task 10: Integrate CLI into main.go

**Files:**
- Modify: `cmd/mcp-task-manager/main.go`

**Step 1: Update main.go**

Replace `cmd/mcp-task-manager/main.go` with:

```go
package main

import (
	"log"
	"os"

	"github.com/gpayer/mcp-task-manager/internal/cli"
	"github.com/gpayer/mcp-task-manager/internal/config"
	"github.com/gpayer/mcp-task-manager/internal/storage"
	"github.com/gpayer/mcp-task-manager/internal/task"
	"github.com/gpayer/mcp-task-manager/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// CLI mode if any arguments provided
	if len(os.Args) > 1 {
		cli.Run()
		return
	}

	// MCP server mode
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	tasksDir := cfg.TasksDir()
	mdStorage := storage.NewMarkdownStorage(tasksDir)
	index := storage.NewIndex(tasksDir, mdStorage)

	svc := task.NewService(mdStorage, index, cfg.TaskTypes)
	if err := svc.Initialize(); err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	s := server.NewMCPServer(
		"mcp-task-manager",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	tools.Register(s, svc, cfg.TaskTypes)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
```

**Step 2: Build and test**

Run:
```bash
cd /home/gernot/src/mcp-task-manager && go build ./cmd/mcp-task-manager
./mcp-task-manager --help
./mcp-task-manager version
./mcp-task-manager list
```

Expected: Help output, version, and task list

**Step 3: Run all tests**

Run:
```bash
go test ./... -v
```

Expected: All tests PASS

**Step 4: Commit**

```bash
git add cmd/mcp-task-manager/main.go
git commit -m "feat: integrate CLI mode into main.go"
```

---

## Task 11: Final cleanup and documentation

**Files:**
- Modify: `README.md`

**Step 1: Update README with CLI usage**

Add CLI section to README.md after the MCP tools section.

**Step 2: Run full test suite**

Run:
```bash
go test ./... -v
go build ./cmd/mcp-task-manager
```

Expected: All PASS, build succeeds

**Step 3: Final commit**

```bash
git add README.md
git commit -m "docs: add CLI usage to README"
```

---

## Summary

| Task | Description |
|------|-------------|
| 1 | Add flaggy dependency |
| 2 | Create CLI output helpers |
| 3 | Create CLI structure with version command |
| 4 | Add list command |
| 5 | Add get and next commands |
| 6 | Add create command |
| 7 | Add update command |
| 8 | Add delete command |
| 9 | Add start and complete commands |
| 10 | Integrate CLI into main.go |
| 11 | Final cleanup and documentation |
