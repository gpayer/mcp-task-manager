package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gpayer/mcp-task-manager/internal/task"
)

func TestMarkdownStorage_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)

	now := time.Now().UTC().Truncate(time.Second)
	original := &task.Task{
		ID:          1,
		Title:       "Test Task",
		Description: "This is a test description.\n\nWith multiple paragraphs.",
		Status:      task.StatusTodo,
		Priority:    task.PriorityHigh,
		Type:        "feature",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Save
	if err := storage.Save(original); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join(dir, "001.md")); os.IsNotExist(err) {
		t.Fatal("expected 001.md to exist")
	}

	// Load
	loaded, err := storage.Load(1)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Compare fields
	if loaded.ID != original.ID {
		t.Errorf("ID = %d, want %d", loaded.ID, original.ID)
	}
	if loaded.Title != original.Title {
		t.Errorf("Title = %q, want %q", loaded.Title, original.Title)
	}
	if loaded.Description != original.Description {
		t.Errorf("Description = %q, want %q", loaded.Description, original.Description)
	}
	if loaded.Status != original.Status {
		t.Errorf("Status = %q, want %q", loaded.Status, original.Status)
	}
	if loaded.Priority != original.Priority {
		t.Errorf("Priority = %q, want %q", loaded.Priority, original.Priority)
	}
	if loaded.Type != original.Type {
		t.Errorf("Type = %q, want %q", loaded.Type, original.Type)
	}
}

func TestMarkdownStorage_Delete(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)

	task := &task.Task{
		ID:        1,
		Title:     "To Delete",
		Status:    task.StatusTodo,
		Priority:  task.PriorityMedium,
		Type:      "bug",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := storage.Save(task); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := storage.Delete(1); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if _, err := storage.Load(1); !os.IsNotExist(err) {
		t.Errorf("expected file to be deleted, got error: %v", err)
	}
}

func TestMarkdownStorage_LoadAll(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)

	now := time.Now().UTC()
	tasks := []*task.Task{
		{ID: 1, Title: "Task 1", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature", CreatedAt: now, UpdatedAt: now},
		{ID: 2, Title: "Task 2", Status: task.StatusDone, Priority: task.PriorityLow, Type: "bug", CreatedAt: now, UpdatedAt: now},
		{ID: 3, Title: "Task 3", Status: task.StatusInProgress, Priority: task.PriorityMedium, Type: "feature", CreatedAt: now, UpdatedAt: now},
	}

	for _, tk := range tasks {
		if err := storage.Save(tk); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}

	loaded, err := storage.LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	if len(loaded) != 3 {
		t.Errorf("LoadAll() returned %d tasks, want 3", len(loaded))
	}
}

func TestMarkdownStorage_NextID(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)

	// Empty directory
	id, err := storage.NextID()
	if err != nil {
		t.Fatalf("NextID() error = %v", err)
	}
	if id != 1 {
		t.Errorf("NextID() on empty dir = %d, want 1", id)
	}

	// Add some tasks
	now := time.Now().UTC()
	for _, i := range []int{1, 5, 3} {
		tk := &task.Task{ID: i, Title: "Task", Status: task.StatusTodo, Priority: task.PriorityMedium, Type: "feature", CreatedAt: now, UpdatedAt: now}
		if err := storage.Save(tk); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}

	id, err = storage.NextID()
	if err != nil {
		t.Fatalf("NextID() error = %v", err)
	}
	if id != 6 {
		t.Errorf("NextID() = %d, want 6", id)
	}
}
