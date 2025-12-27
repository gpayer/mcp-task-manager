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

	output := FormatTaskDetail(tk, nil)

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

func TestFormatTaskDetailWithSubtasks(t *testing.T) {
	created := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	updated := time.Date(2025, 1, 15, 11, 0, 0, 0, time.UTC)
	parentID := 1
	tk := &task.Task{
		ID:        1,
		Title:     "Parent task",
		Status:    task.StatusInProgress,
		Priority:  task.PriorityHigh,
		Type:      "feature",
		CreatedAt: created,
		UpdatedAt: updated,
	}
	subtasks := []*task.Task{
		{ID: 2, Title: "Subtask 1", Status: task.StatusDone, ParentID: &parentID},
		{ID: 3, Title: "Subtask 2", Status: task.StatusTodo, ParentID: &parentID},
	}

	output := FormatTaskDetail(tk, subtasks)

	if !strings.Contains(output, "Subtasks (2)") {
		t.Error("expected 'Subtasks (2)' in output")
	}
	if !strings.Contains(output, "#2 [done] Subtask 1") {
		t.Error("expected subtask 1 in output")
	}
	if !strings.Contains(output, "#3 [todo] Subtask 2") {
		t.Error("expected subtask 2 in output")
	}
}

func TestFormatTaskTable(t *testing.T) {
	tasks := []*task.Task{
		{ID: 1, Title: "First task", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature"},
		{ID: 2, Title: "Second task", Status: task.StatusDone, Priority: task.PriorityLow, Type: "bug"},
	}

	output := FormatTaskTable(tasks, nil)

	if !strings.Contains(output, "ID") {
		t.Error("expected header row")
	}
	if !strings.Contains(output, "First task") {
		t.Error("expected first task")
	}
	if !strings.Contains(output, "Second task") {
		t.Error("expected second task")
	}
	if !strings.Contains(output, "Subtasks") {
		t.Error("expected Subtasks column header")
	}
}

func TestFormatTaskTableWithSubtasks(t *testing.T) {
	tasks := []*task.Task{
		{ID: 1, Title: "Parent task", Status: task.StatusInProgress, Priority: task.PriorityHigh, Type: "feature"},
	}

	// Subtask counts passed externally (as would be computed by cmdList)
	subtaskCounts := map[int]SubtaskCounts{
		1: {Total: 3, Done: 2},
	}

	output := FormatTaskTable(tasks, subtaskCounts)

	// Parent task should show [2/3] (2 done out of 3 subtasks)
	if !strings.Contains(output, "[2/3]") {
		t.Errorf("expected subtask count [2/3] for parent task, got:\n%s", output)
	}
}

func TestFormatTaskTableEmpty(t *testing.T) {
	output := FormatTaskTable([]*task.Task{}, nil)
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
