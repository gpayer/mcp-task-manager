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

func TestIndex_SetGetDelete(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	now := time.Now().UTC()
	tk := &task.Task{
		ID:        1,
		Title:     "Test",
		Status:    task.StatusTodo,
		Priority:  task.PriorityHigh,
		Type:      "feature",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Set
	idx.Set(tk)

	// Get
	got, ok := idx.Get(1)
	if !ok {
		t.Fatal("Get() returned false for existing task")
	}
	if got.Title != tk.Title {
		t.Errorf("Get() title = %q, want %q", got.Title, tk.Title)
	}

	// Delete
	idx.Delete(1)
	_, ok = idx.Get(1)
	if ok {
		t.Error("Get() returned true for deleted task")
	}
}

func TestIndex_Filter(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	now := time.Now().UTC()
	tasks := []*task.Task{
		{ID: 1, Title: "Task 1", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature", CreatedAt: now, UpdatedAt: now},
		{ID: 2, Title: "Task 2", Status: task.StatusDone, Priority: task.PriorityLow, Type: "bug", CreatedAt: now, UpdatedAt: now},
		{ID: 3, Title: "Task 3", Status: task.StatusTodo, Priority: task.PriorityMedium, Type: "feature", CreatedAt: now, UpdatedAt: now},
	}
	for _, tk := range tasks {
		idx.Set(tk)
	}

	// Filter by status
	todoStatus := task.StatusTodo
	filtered := idx.Filter(&todoStatus, nil, nil)
	if len(filtered) != 2 {
		t.Errorf("Filter by todo status returned %d tasks, want 2", len(filtered))
	}

	// Filter by priority
	highPriority := task.PriorityHigh
	filtered = idx.Filter(nil, &highPriority, nil)
	if len(filtered) != 1 {
		t.Errorf("Filter by high priority returned %d tasks, want 1", len(filtered))
	}

	// Filter by type
	featureType := "feature"
	filtered = idx.Filter(nil, nil, &featureType)
	if len(filtered) != 2 {
		t.Errorf("Filter by feature type returned %d tasks, want 2", len(filtered))
	}

	// Combined filter
	filtered = idx.Filter(&todoStatus, nil, &featureType)
	if len(filtered) != 2 {
		t.Errorf("Combined filter returned %d tasks, want 2", len(filtered))
	}
}

func TestIndex_NextTodo(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	// Empty index
	if got := idx.NextTodo(); got != nil {
		t.Errorf("NextTodo() on empty index = %v, want nil", got)
	}

	now := time.Now().UTC()
	tasks := []*task.Task{
		{ID: 1, Title: "Low", Status: task.StatusTodo, Priority: task.PriorityLow, Type: "feature", CreatedAt: now, UpdatedAt: now},
		{ID: 2, Title: "Critical", Status: task.StatusTodo, Priority: task.PriorityCritical, Type: "bug", CreatedAt: now.Add(time.Hour), UpdatedAt: now},
		{ID: 3, Title: "Critical Older", Status: task.StatusTodo, Priority: task.PriorityCritical, Type: "feature", CreatedAt: now, UpdatedAt: now},
		{ID: 4, Title: "Done", Status: task.StatusDone, Priority: task.PriorityCritical, Type: "bug", CreatedAt: now, UpdatedAt: now},
	}
	for _, tk := range tasks {
		idx.Set(tk)
	}

	// Should return Critical Older (highest priority, oldest)
	next := idx.NextTodo()
	if next == nil {
		t.Fatal("NextTodo() returned nil")
	}
	if next.ID != 3 {
		t.Errorf("NextTodo() ID = %d, want 3 (Critical Older)", next.ID)
	}
}

func TestIndex_NextID(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	// Empty
	if got := idx.NextID(); got != 1 {
		t.Errorf("NextID() on empty = %d, want 1", got)
	}

	// Add tasks
	now := time.Now().UTC()
	idx.Set(&task.Task{ID: 5, Title: "Task", Status: task.StatusTodo, Priority: task.PriorityMedium, Type: "feature", CreatedAt: now, UpdatedAt: now})
	idx.Set(&task.Task{ID: 3, Title: "Task", Status: task.StatusTodo, Priority: task.PriorityMedium, Type: "feature", CreatedAt: now, UpdatedAt: now})

	if got := idx.NextID(); got != 6 {
		t.Errorf("NextID() = %d, want 6", got)
	}
}

func TestIndex_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	now := time.Now().UTC()
	tk := &task.Task{
		ID:        1,
		Title:     "Persisted",
		Status:    task.StatusTodo,
		Priority:  task.PriorityHigh,
		Type:      "feature",
		CreatedAt: now,
		UpdatedAt: now,
	}
	idx.Set(tk)

	if err := idx.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Create new index and load
	idx2 := NewIndex(dir, storage)
	if err := idx2.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	got, ok := idx2.Get(1)
	if !ok {
		t.Fatal("Get() after Load() returned false")
	}
	if got.Title != tk.Title {
		t.Errorf("Title after Load() = %q, want %q", got.Title, tk.Title)
	}
}

func TestMarkdownStorage_SaveLoad_WithParentID(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)

	parentID := 1
	tsk := &task.Task{
		ID:        2,
		ParentID:  &parentID,
		Title:     "Subtask",
		Status:    task.StatusTodo,
		Priority:  task.PriorityHigh,
		Type:      "feature",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := storage.Save(tsk); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := storage.Load(2)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.ParentID == nil {
		t.Fatal("ParentID should not be nil")
	}
	if *loaded.ParentID != 1 {
		t.Errorf("ParentID = %d, want 1", *loaded.ParentID)
	}
}

func TestMarkdownStorage_SaveLoad_WithoutParentID(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)

	tsk := &task.Task{
		ID:        1,
		ParentID:  nil,
		Title:     "Top-level task",
		Status:    task.StatusTodo,
		Priority:  task.PriorityHigh,
		Type:      "feature",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := storage.Save(tsk); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := storage.Load(1)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.ParentID != nil {
		t.Error("ParentID should be nil for top-level task")
	}
}

func TestIndex_RebuildFromFiles(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)

	// Save tasks directly to storage
	now := time.Now().UTC()
	tasks := []*task.Task{
		{ID: 1, Title: "Task 1", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature", CreatedAt: now, UpdatedAt: now},
		{ID: 2, Title: "Task 2", Status: task.StatusDone, Priority: task.PriorityLow, Type: "bug", CreatedAt: now, UpdatedAt: now},
	}
	for _, tk := range tasks {
		if err := storage.Save(tk); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}

	// Create index without existing index file (triggers rebuild)
	idx := NewIndex(dir, storage)
	if err := idx.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify tasks were loaded
	all := idx.All()
	if len(all) != 2 {
		t.Errorf("All() after rebuild returned %d tasks, want 2", len(all))
	}
}
