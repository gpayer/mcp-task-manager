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
	if !strings.Contains(buf.String(), `"id": 1`) {
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
	if !strings.Contains(output, `"id": 5`) {
		t.Error("expected id field")
	}
}
