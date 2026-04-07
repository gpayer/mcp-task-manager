package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/integrii/flaggy"
)

func init() {
	// Enable panic instead of exit for testing
	flaggy.PanicInsteadOfExit = true
}

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

	// flaggy may panic on --help, so catch it
	defer func() {
		if r := recover(); r != nil {
			// This is expected for --help flag
			t.Logf("Recovered from panic (expected): %v", r)
		}
	}()

	code := RunWithArgs([]string{"mcp-task-manager", "--help"}, &stdout, &stderr)

	// flaggy exits with 0 on --help
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

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
		t.Errorf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "[") {
		t.Errorf("expected JSON array output, got: %s", output)
	}
}

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

func TestTypeHelpTextIncludesAllowedValues(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := tmpDir + "/mcp-tasks.yaml"
	if err := os.WriteFile(configPath, []byte("task_types:\n  - bug\n  - chore\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	oldCwd, _ := os.Getwd()
	defer os.Chdir(oldCwd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	tests := []struct {
		name    string
		args    []string
		wantOut string
	}{
		{
			name:    "list",
			args:    []string{"mcp-task-manager", "list", "--help"},
			wantOut: "Filter by type (bug|chore)",
		},
		{
			name:    "create",
			args:    []string{"mcp-task-manager", "create", "--help"},
			wantOut: "Type (bug|chore; default: bug)",
		},
		{
			name:    "update",
			args:    []string{"mcp-task-manager", "update", "--help"},
			wantOut: "New type (bug|chore)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := runHelp(t, tt.args)
			if !strings.Contains(output, tt.wantOut) {
				t.Fatalf("help output = %q, want substring %q", output, tt.wantOut)
			}
		})
	}
}

func TestCreateCommandUsesConfiguredDefaultType(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(tmpDir+"/mcp-tasks.yaml", []byte("task_types:\n  - bug\n  - chore\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("MCP_TASKS_DIR", tmpDir+"/tasks")

	var stdout, stderr bytes.Buffer
	code := RunWithArgs([]string{"mcp-task-manager", "create", "Config default type"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "bug") {
		t.Fatalf("expected created task to use configured default type 'bug', got: %s", stdout.String())
	}
}

func TestListCommandNoProject(t *testing.T) {
	// Run from a temp directory with no project markers
	// Use a deeply nested temp dir to avoid any existing tasks/ or mcp-tasks.yaml in /tmp
	tmpDir := t.TempDir()
	nestedDir := tmpDir + "/a/b/c/d"
	os.MkdirAll(nestedDir, 0755)
	originalWd, _ := os.Getwd()
	os.Chdir(nestedDir)
	defer os.Chdir(originalWd)

	// Ensure no MCP_TASKS_DIR is set
	os.Unsetenv("MCP_TASKS_DIR")

	var stdout, stderr bytes.Buffer
	code := RunWithArgs([]string{"mcp-task-manager", "list"}, &stdout, &stderr)

	if code != 1 {
		t.Errorf("expected exit code 1 when no project found, got %d", code)
	}
	if !strings.Contains(stderr.String(), "no tasks directory found") {
		t.Errorf("expected 'no tasks directory found' error, got: %s", stderr.String())
	}
}

func runHelp(t *testing.T, args []string) string {
	t.Helper()

	var stdout, stderr bytes.Buffer
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe stdout: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe stderr: %v", err)
	}
	os.Stdout = wOut
	os.Stderr = wErr

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic (expected): %v", r)
			}
		}()
		RunWithArgs(args, &stdout, &stderr)
	}()
	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	var helpOutput bytes.Buffer
	if _, err := io.Copy(&helpOutput, rOut); err != nil {
		t.Fatalf("copy stdout: %v", err)
	}
	if _, err := io.Copy(&helpOutput, rErr); err != nil {
		t.Fatalf("copy stderr: %v", err)
	}
	return stdout.String() + stderr.String() + helpOutput.String()
}

func TestGetCommandNoProject(t *testing.T) {
	// Run from a temp directory with no project markers
	// Use a deeply nested temp dir to avoid any existing tasks/ or mcp-tasks.yaml in /tmp
	tmpDir := t.TempDir()
	nestedDir := tmpDir + "/a/b/c/d"
	os.MkdirAll(nestedDir, 0755)
	originalWd, _ := os.Getwd()
	os.Chdir(nestedDir)
	defer os.Chdir(originalWd)

	// Ensure no MCP_TASKS_DIR is set
	os.Unsetenv("MCP_TASKS_DIR")

	var stdout, stderr bytes.Buffer
	code := RunWithArgs([]string{"mcp-task-manager", "get", "1"}, &stdout, &stderr)

	if code != 1 {
		t.Errorf("expected exit code 1 when no project found, got %d", code)
	}
	if !strings.Contains(stderr.String(), "no tasks directory found") {
		t.Errorf("expected 'no tasks directory found' error, got: %s", stderr.String())
	}
}

func TestArchiveCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	// Create, start, and complete a task
	var stdout, stderr bytes.Buffer
	RunWithArgs([]string{"mcp-task-manager", "create", "Task to archive"}, &stdout, &stderr)
	RunWithArgs([]string{"mcp-task-manager", "start", "1"}, &stdout, &stderr)
	RunWithArgs([]string{"mcp-task-manager", "complete", "1"}, &stdout, &stderr)

	// Archive it
	stdout.Reset()
	stderr.Reset()
	code := RunWithArgs([]string{"mcp-task-manager", "archive", "1"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "archived") {
		t.Errorf("expected 'archived' message, got: %s", stdout.String())
	}
}

func TestArchiveCommandJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	// Create, start, and complete a task
	var stdout, stderr bytes.Buffer
	RunWithArgs([]string{"mcp-task-manager", "create", "Task to archive JSON"}, &stdout, &stderr)
	RunWithArgs([]string{"mcp-task-manager", "start", "1"}, &stdout, &stderr)
	RunWithArgs([]string{"mcp-task-manager", "complete", "1"}, &stdout, &stderr)

	// Archive it with JSON output
	stdout.Reset()
	stderr.Reset()
	code := RunWithArgs([]string{"mcp-task-manager", "archive", "1", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "archived") {
		t.Errorf("expected 'archived' in JSON output, got: %s", output)
	}
}

func TestArchiveCommandNotDone(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	// Create a task but don't complete it
	var stdout, stderr bytes.Buffer
	RunWithArgs([]string{"mcp-task-manager", "create", "Task not done"}, &stdout, &stderr)

	// Try to archive it - should fail
	stdout.Reset()
	stderr.Reset()
	code := RunWithArgs([]string{"mcp-task-manager", "archive", "1"}, &stdout, &stderr)

	if code != 1 {
		t.Errorf("expected exit code 1 for non-done task, got %d", code)
	}
}

func TestListArchivedCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	// Create, complete, and archive a task
	var stdout, stderr bytes.Buffer
	RunWithArgs([]string{"mcp-task-manager", "create", "Archived task"}, &stdout, &stderr)
	RunWithArgs([]string{"mcp-task-manager", "start", "1"}, &stdout, &stderr)
	RunWithArgs([]string{"mcp-task-manager", "complete", "1"}, &stdout, &stderr)
	RunWithArgs([]string{"mcp-task-manager", "archive", "1"}, &stdout, &stderr)

	// List archived tasks
	stdout.Reset()
	stderr.Reset()
	code := RunWithArgs([]string{"mcp-task-manager", "list", "--archived"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Archived task") {
		t.Errorf("expected archived task in output, got: %s", stdout.String())
	}
}

func TestListArchivedCommandJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	// Create, complete, and archive a task
	var stdout, stderr bytes.Buffer
	RunWithArgs([]string{"mcp-task-manager", "create", "Archived JSON task"}, &stdout, &stderr)
	RunWithArgs([]string{"mcp-task-manager", "start", "1"}, &stdout, &stderr)
	RunWithArgs([]string{"mcp-task-manager", "complete", "1"}, &stdout, &stderr)
	RunWithArgs([]string{"mcp-task-manager", "archive", "1"}, &stdout, &stderr)

	// List archived tasks as JSON
	stdout.Reset()
	stderr.Reset()
	code := RunWithArgs([]string{"mcp-task-manager", "list", "--archived", "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "[") {
		t.Errorf("expected JSON array output, got: %s", output)
	}
	if !strings.Contains(output, "Archived JSON task") {
		t.Errorf("expected archived task in JSON output, got: %s", output)
	}
}

func TestListArchivedEmptyCommand(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MCP_TASKS_DIR", tmpDir)

	// List archived tasks when none exist
	var stdout, stderr bytes.Buffer
	code := RunWithArgs([]string{"mcp-task-manager", "list", "--archived"}, &stdout, &stderr)

	if code != 0 {
		t.Errorf("expected exit code 0, got %d. stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "No tasks") {
		t.Errorf("expected 'No tasks' message, got: %s", stdout.String())
	}
}

func TestNextCommandNoProject(t *testing.T) {
	// Run from a temp directory with no project markers
	// Use a deeply nested temp dir to avoid any existing tasks/ or mcp-tasks.yaml in /tmp
	tmpDir := t.TempDir()
	nestedDir := tmpDir + "/a/b/c/d"
	os.MkdirAll(nestedDir, 0755)
	originalWd, _ := os.Getwd()
	os.Chdir(nestedDir)
	defer os.Chdir(originalWd)

	// Ensure no MCP_TASKS_DIR is set
	os.Unsetenv("MCP_TASKS_DIR")

	var stdout, stderr bytes.Buffer
	code := RunWithArgs([]string{"mcp-task-manager", "next"}, &stdout, &stderr)

	if code != 1 {
		t.Errorf("expected exit code 1 when no project found, got %d", code)
	}
	if !strings.Contains(stderr.String(), "no tasks directory found") {
		t.Errorf("expected 'no tasks directory found' error, got: %s", stderr.String())
	}
}
