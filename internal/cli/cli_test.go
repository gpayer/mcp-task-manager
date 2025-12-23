package cli

import (
	"bytes"
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
