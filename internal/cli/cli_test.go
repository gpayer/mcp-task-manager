package cli

import (
	"bytes"
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
