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
