package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

	// Save to disk first (required for Get to work)
	if err := storage.Save(tk); err != nil {
		t.Fatalf("Save() error = %v", err)
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
	filtered := idx.Filter(&todoStatus, nil, nil, nil)
	if len(filtered) != 2 {
		t.Errorf("Filter by todo status returned %d tasks, want 2", len(filtered))
	}

	// Filter by priority
	highPriority := task.PriorityHigh
	filtered = idx.Filter(nil, &highPriority, nil, nil)
	if len(filtered) != 1 {
		t.Errorf("Filter by high priority returned %d tasks, want 1", len(filtered))
	}

	// Filter by type
	featureType := "feature"
	filtered = idx.Filter(nil, nil, &featureType, nil)
	if len(filtered) != 2 {
		t.Errorf("Filter by feature type returned %d tasks, want 2", len(filtered))
	}

	// Combined filter
	filtered = idx.Filter(&todoStatus, nil, &featureType, nil)
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

	// Save to disk first (required for Get to work)
	if err := storage.Save(tk); err != nil {
		t.Fatalf("Save() error = %v", err)
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

func TestIndex_GetSubtasks(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	// Create parent task
	parent := &task.Task{ID: 1, Title: "Parent", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature"}
	storage.Save(parent)

	// Create subtasks
	parentID := 1
	sub1 := &task.Task{ID: 2, ParentID: &parentID, Title: "Sub 1", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature"}
	sub2 := &task.Task{ID: 3, ParentID: &parentID, Title: "Sub 2", Status: task.StatusDone, Priority: task.PriorityMedium, Type: "feature"}
	storage.Save(sub1)
	storage.Save(sub2)

	idx.Load()

	subtasks := idx.GetSubtasks(1)
	if len(subtasks) != 2 {
		t.Errorf("GetSubtasks(1) = %d, want 2", len(subtasks))
	}

	// Non-existent parent
	subtasks = idx.GetSubtasks(99)
	if len(subtasks) != 0 {
		t.Errorf("GetSubtasks(99) = %d, want 0", len(subtasks))
	}
}

func TestIndex_HasSubtasks(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	parent := &task.Task{ID: 1, Title: "Parent", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature"}
	storage.Save(parent)

	parentID := 1
	sub := &task.Task{ID: 2, ParentID: &parentID, Title: "Sub", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature"}
	storage.Save(sub)

	idx.Load()

	if !idx.HasSubtasks(1) {
		t.Error("HasSubtasks(1) = false, want true")
	}
	if idx.HasSubtasks(2) {
		t.Error("HasSubtasks(2) = true, want false")
	}
}

func TestIndex_SubtaskCounts(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	parent := &task.Task{ID: 1, Title: "Parent", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature"}
	storage.Save(parent)

	parentID := 1
	sub1 := &task.Task{ID: 2, ParentID: &parentID, Title: "Sub 1", Status: task.StatusDone, Priority: task.PriorityHigh, Type: "feature"}
	sub2 := &task.Task{ID: 3, ParentID: &parentID, Title: "Sub 2", Status: task.StatusTodo, Priority: task.PriorityMedium, Type: "feature"}
	sub3 := &task.Task{ID: 4, ParentID: &parentID, Title: "Sub 3", Status: task.StatusDone, Priority: task.PriorityLow, Type: "feature"}
	storage.Save(sub1)
	storage.Save(sub2)
	storage.Save(sub3)

	idx.Load()

	total, done := idx.SubtaskCounts(1)
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
	if done != 2 {
		t.Errorf("done = %d, want 2", done)
	}
}

func TestIndex_NextTodo_SkipsParentsWithSubtasks(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	// Create parent with subtask
	parent := &task.Task{ID: 1, Title: "Parent", Status: task.StatusTodo, Priority: task.PriorityCritical, Type: "feature", CreatedAt: time.Now()}
	storage.Save(parent)

	parentID := 1
	sub := &task.Task{ID: 2, ParentID: &parentID, Title: "Subtask", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature", CreatedAt: time.Now()}
	storage.Save(sub)

	// Create standalone task (lower priority)
	standalone := &task.Task{ID: 3, Title: "Standalone", Status: task.StatusTodo, Priority: task.PriorityLow, Type: "feature", CreatedAt: time.Now()}
	storage.Save(standalone)

	idx.Load()

	// Should return subtask (skipping parent even though it's higher priority)
	next := idx.NextTodo()
	if next == nil {
		t.Fatal("NextTodo() returned nil")
	}
	if next.ID != 2 {
		t.Errorf("NextTodo() ID = %d, want 2 (subtask)", next.ID)
	}
}

func TestIndex_NextTodo_ReturnsParentWithoutSubtasks(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	// Create parent without subtasks
	parent := &task.Task{ID: 1, Title: "Parent", Status: task.StatusTodo, Priority: task.PriorityCritical, Type: "feature", CreatedAt: time.Now()}
	storage.Save(parent)

	idx.Load()

	next := idx.NextTodo()
	if next == nil {
		t.Fatal("NextTodo() returned nil")
	}
	if next.ID != 1 {
		t.Errorf("NextTodo() ID = %d, want 1", next.ID)
	}
}

func TestIndex_NextTodo_PrioritizesInProgressParentSubtasks(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	now := time.Now().UTC()

	// Create in_progress parent with todo subtasks
	inProgressParent := &task.Task{ID: 1, Title: "In Progress Parent", Status: task.StatusInProgress, Priority: task.PriorityMedium, Type: "feature", CreatedAt: now}
	storage.Save(inProgressParent)

	parentID1 := 1
	subLowPriority := &task.Task{ID: 2, ParentID: &parentID1, Title: "Low Priority Subtask", Status: task.StatusTodo, Priority: task.PriorityLow, Type: "feature", CreatedAt: now}
	subHighPriority := &task.Task{ID: 3, ParentID: &parentID1, Title: "High Priority Subtask", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature", CreatedAt: now}
	storage.Save(subLowPriority)
	storage.Save(subHighPriority)

	// Create standalone critical task (highest priority but should be deprioritized)
	criticalStandalone := &task.Task{ID: 4, Title: "Critical Standalone", Status: task.StatusTodo, Priority: task.PriorityCritical, Type: "feature", CreatedAt: now}
	storage.Save(criticalStandalone)

	idx.Load()

	// Should return high priority subtask of in_progress parent, not the critical standalone
	next := idx.NextTodo()
	if next == nil {
		t.Fatal("NextTodo() returned nil")
	}
	if next.ID != 3 {
		t.Errorf("NextTodo() ID = %d, want 3 (high priority subtask of in_progress parent)", next.ID)
	}

	// Complete the high priority subtask
	entry3, _ := idx.GetEntry(3)
	completedTask3 := entryToTask(entry3)
	completedTask3.Status = task.StatusDone
	idx.Set(completedTask3)

	// Now should return low priority subtask of in_progress parent
	next = idx.NextTodo()
	if next == nil {
		t.Fatal("NextTodo() returned nil after completing high priority subtask")
	}
	if next.ID != 2 {
		t.Errorf("NextTodo() ID = %d, want 2 (low priority subtask of in_progress parent)", next.ID)
	}

	// Complete the low priority subtask too
	entry2, _ := idx.GetEntry(2)
	completedTask2 := entryToTask(entry2)
	completedTask2.Status = task.StatusDone
	idx.Set(completedTask2)

	// Now should return the critical standalone
	next = idx.NextTodo()
	if next == nil {
		t.Fatal("NextTodo() returned nil after completing all subtasks")
	}
	if next.ID != 4 {
		t.Errorf("NextTodo() ID = %d, want 4 (critical standalone)", next.ID)
	}
}

func TestIndex_NextTodo_TodoParentSubtasksNotPrioritized(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	now := time.Now().UTC()

	// Create todo parent with subtasks (parent not started yet)
	todoParent := &task.Task{ID: 1, Title: "Todo Parent", Status: task.StatusTodo, Priority: task.PriorityCritical, Type: "feature", CreatedAt: now}
	storage.Save(todoParent)

	parentID1 := 1
	subHighPriority := &task.Task{ID: 2, ParentID: &parentID1, Title: "High Priority Subtask", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature", CreatedAt: now}
	storage.Save(subHighPriority)

	// Create standalone medium task
	mediumStandalone := &task.Task{ID: 3, Title: "Medium Standalone", Status: task.StatusTodo, Priority: task.PriorityMedium, Type: "feature", CreatedAt: now}
	storage.Save(mediumStandalone)

	idx.Load()

	// Should return high priority subtask (normal priority sorting, not boosted)
	// because parent is still in todo state, not in_progress
	next := idx.NextTodo()
	if next == nil {
		t.Fatal("NextTodo() returned nil")
	}
	if next.ID != 2 {
		t.Errorf("NextTodo() ID = %d, want 2 (high priority subtask, normal sort order)", next.ID)
	}
}

func TestIndex_Filter_ByParentID(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	parent := &task.Task{ID: 1, Title: "Parent", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature"}
	storage.Save(parent)

	parentID := 1
	sub1 := &task.Task{ID: 2, ParentID: &parentID, Title: "Sub 1", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature"}
	sub2 := &task.Task{ID: 3, ParentID: &parentID, Title: "Sub 2", Status: task.StatusDone, Priority: task.PriorityMedium, Type: "feature"}
	storage.Save(sub1)
	storage.Save(sub2)

	standalone := &task.Task{ID: 4, Title: "Standalone", Status: task.StatusTodo, Priority: task.PriorityLow, Type: "feature"}
	storage.Save(standalone)

	idx.Load()

	// Filter subtasks of parent
	result := idx.Filter(nil, nil, nil, &parentID)
	if len(result) != 2 {
		t.Errorf("Filter(parent_id=1) = %d, want 2", len(result))
	}

	// Filter top-level only (parent_id = 0 means top-level)
	topLevel := 0
	result = idx.Filter(nil, nil, nil, &topLevel)
	if len(result) != 2 {
		t.Errorf("Filter(parent_id=0) = %d, want 2 (parent + standalone)", len(result))
	}

	// No filter - returns all tasks
	result = idx.Filter(nil, nil, nil, nil)
	if len(result) != 4 {
		t.Errorf("Filter(parent_id=nil) = %d, want 4", len(result))
	}
}

func TestGetGitCommit(t *testing.T) {
	// Test in non-git directory
	dir := t.TempDir()
	commit, err := getGitCommit(dir)
	if err != nil {
		t.Fatalf("getGitCommit() in non-git dir error = %v", err)
	}
	if commit != "" {
		t.Errorf("getGitCommit() in non-git dir = %q, want empty", commit)
	}

	// Test in git directory (use the actual project dir)
	// The test is running inside a git repo, so cwd should work
	commit, err = getGitCommit(".")
	if err != nil {
		t.Fatalf("getGitCommit() error = %v", err)
	}
	if commit == "" {
		t.Error("getGitCommit() in git repo returned empty string")
	}
	if len(commit) != 40 {
		t.Errorf("getGitCommit() returned %q, want 40-char SHA", commit)
	}
}

// Test that GetEntry returns metadata without loading from disk
func TestIndex_GetEntry_ReturnsMetadataOnly(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	now := time.Now().UTC()
	tk := &task.Task{
		ID:          1,
		Title:       "Test Task",
		Description: "This description should not be in the entry",
		Status:      task.StatusTodo,
		Priority:    task.PriorityHigh,
		Type:        "feature",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Set the task in the index
	idx.Set(tk)

	// GetEntry should return entry with metadata but no description
	entry, ok := idx.GetEntry(1)
	if !ok {
		t.Fatal("GetEntry() returned false for existing task")
	}
	if entry.ID != 1 {
		t.Errorf("entry.ID = %d, want 1", entry.ID)
	}
	if entry.Title != "Test Task" {
		t.Errorf("entry.Title = %q, want %q", entry.Title, "Test Task")
	}
	if entry.Status != task.StatusTodo {
		t.Errorf("entry.Status = %v, want %v", entry.Status, task.StatusTodo)
	}

	// GetEntry should return false for non-existent task
	_, ok = idx.GetEntry(999)
	if ok {
		t.Error("GetEntry() returned true for non-existent task")
	}
}

// Test that Get loads full task from disk (with description)
func TestIndex_Get_LoadsFromDisk(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	now := time.Now().UTC()
	tk := &task.Task{
		ID:          1,
		Title:       "Test Task",
		Description: "Full description from disk",
		Status:      task.StatusTodo,
		Priority:    task.PriorityHigh,
		Type:        "feature",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Save to disk
	if err := storage.Save(tk); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Set in index (this should store entry, not full task)
	idx.Set(tk)

	// Get should load from disk and include description
	loaded, ok := idx.Get(1)
	if !ok {
		t.Fatal("Get() returned false for existing task")
	}
	if loaded.Description != "Full description from disk" {
		t.Errorf("Get() description = %q, want %q", loaded.Description, "Full description from disk")
	}
	if loaded.Title != "Test Task" {
		t.Errorf("Get() title = %q, want %q", loaded.Title, "Test Task")
	}
}

// Test that All returns tasks without descriptions
func TestIndex_All_ReturnsTasksWithoutDescriptions(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	now := time.Now().UTC()
	tk1 := &task.Task{
		ID:          1,
		Title:       "Task 1",
		Description: "Description 1",
		Status:      task.StatusTodo,
		Priority:    task.PriorityHigh,
		Type:        "feature",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	tk2 := &task.Task{
		ID:          2,
		Title:       "Task 2",
		Description: "Description 2",
		Status:      task.StatusDone,
		Priority:    task.PriorityLow,
		Type:        "bug",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	idx.Set(tk1)
	idx.Set(tk2)

	all := idx.All()
	if len(all) != 2 {
		t.Errorf("All() returned %d tasks, want 2", len(all))
	}

	// Descriptions should be empty
	for _, tk := range all {
		if tk.Description != "" {
			t.Errorf("All() task %d has description %q, want empty", tk.ID, tk.Description)
		}
	}

	// Other fields should be populated
	if all[0].Title != "Task 1" {
		t.Errorf("All() task 1 title = %q, want %q", all[0].Title, "Task 1")
	}
	if all[1].Title != "Task 2" {
		t.Errorf("All() task 2 title = %q, want %q", all[1].Title, "Task 2")
	}
}

func TestIndex_SaveLoad_WithGitCommit(t *testing.T) {
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
	idx.Set(tk)

	if err := idx.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Read raw index file and verify structure
	data, err := os.ReadFile(filepath.Join(dir, ".index.json"))
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}

	var indexFile IndexFile
	if err := json.Unmarshal(data, &indexFile); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}

	// Git commit might be empty if test dir is not in git
	// but tasks should be present
	if len(indexFile.Tasks) != 1 {
		t.Errorf("Tasks count = %d, want 1", len(indexFile.Tasks))
	}
	if indexFile.Tasks[0].Title != "Test" {
		t.Errorf("Task title = %q, want 'Test'", indexFile.Tasks[0].Title)
	}
}

func TestIndex_Load_ParsesNewFormat(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)

	// Create index file in new format (no need to match git, just parse correctly)
	now := time.Now().UTC()
	indexFile := IndexFile{
		GitCommit: "", // Empty is fine for non-git directories
		Tasks: []*IndexEntry{
			{ID: 1, Title: "Task from Index", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature", CreatedAt: now, UpdatedAt: now},
		},
	}
	data, _ := json.MarshalIndent(indexFile, "", "  ")
	os.WriteFile(filepath.Join(dir, ".index.json"), data, 0644)

	// Load should parse the new format successfully
	idx := NewIndex(dir, storage)
	if err := idx.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should have task from index
	entry, ok := idx.GetEntry(1)
	if !ok {
		t.Fatal("Task 1 should exist from index")
	}
	if entry.Title != "Task from Index" {
		t.Errorf("Title = %q, want 'Task from Index'", entry.Title)
	}
}

func TestIndex_Load_RebuildOnGitChange(t *testing.T) {
	// Create temp dir inside project (so we're in git repo)
	dir := filepath.Join(".", "testdata", "rebuild_test")
	os.RemoveAll(dir) // Clean up from previous runs
	defer os.RemoveAll(dir)
	os.MkdirAll(dir, 0755)

	storage := NewMarkdownStorage(dir)

	// Save a task directly to storage
	now := time.Now().UTC()
	tk := &task.Task{ID: 1, Title: "From File", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature", CreatedAt: now, UpdatedAt: now}
	if err := storage.Save(tk); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filepath.Join(dir, "001.md")); err != nil {
		t.Fatalf("Task file not created: %v", err)
	}

	// Create index file with different git commit
	indexFile := IndexFile{
		GitCommit: "0000000000000000000000000000000000000000", // Fake commit that won't match
		Tasks: []*IndexEntry{
			{ID: 99, Title: "Old Cached", Status: task.StatusDone, Priority: task.PriorityLow, Type: "bug", CreatedAt: now, UpdatedAt: now},
		},
	}
	data, _ := json.MarshalIndent(indexFile, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, ".index.json"), data, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Check what git commit we'll get
	currentCommit, _ := getGitCommit(dir)
	t.Logf("Current git commit: %q", currentCommit)
	t.Logf("Index git commit: %q", indexFile.GitCommit)

	// Load index - should rebuild because git commit doesn't match
	idx := NewIndex(dir, storage)
	if err := idx.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should have task from file, not from stale index
	_, ok := idx.GetEntry(1)
	if !ok {
		t.Error("Task 1 (from file) should exist after rebuild")
	}
	_, ok = idx.GetEntry(99)
	if ok {
		t.Error("Task 99 (stale) should not exist after rebuild")
	}
}

func TestIndex_Load_MigratesOldFormat(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)

	// Save task to storage
	now := time.Now().UTC()
	tk := &task.Task{ID: 1, Title: "Task", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature", CreatedAt: now, UpdatedAt: now}
	storage.Save(tk)

	// Write old format index (raw array)
	oldData := `[{"id":1,"title":"Stale","status":"done","priority":"low","type":"bug"}]`
	os.WriteFile(filepath.Join(dir, ".index.json"), []byte(oldData), 0644)

	// Load should fail to parse as IndexFile and rebuild
	idx := NewIndex(dir, storage)
	if err := idx.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should have fresh data from file
	e, ok := idx.GetEntry(1)
	if !ok {
		t.Fatal("Task 1 should exist")
	}
	if e.Title != "Task" {
		t.Errorf("Title = %q, want 'Task' (from file)", e.Title)
	}
	if e.Status != task.StatusTodo {
		t.Errorf("Status = %q, want 'todo' (from file)", e.Status)
	}
}

// === Relation Tests ===

func TestMarkdownStorage_SaveLoad_WithRelations(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)

	now := time.Now().UTC().Truncate(time.Second)
	tsk := &task.Task{
		ID:       5,
		Title:    "Blocked task",
		Status:   task.StatusTodo,
		Priority: task.PriorityHigh,
		Type:     "feature",
		Relations: []task.Relation{
			{Type: "blocked_by", Task: 3},
			{Type: "relates_to", Task: 7},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := storage.Save(tsk); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := storage.Load(5)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded.Relations) != 2 {
		t.Fatalf("Relations count = %d, want 2", len(loaded.Relations))
	}
	if loaded.Relations[0].Type != "blocked_by" || loaded.Relations[0].Task != 3 {
		t.Errorf("Relations[0] = %v, want {blocked_by, 3}", loaded.Relations[0])
	}
	if loaded.Relations[1].Type != "relates_to" || loaded.Relations[1].Task != 7 {
		t.Errorf("Relations[1] = %v, want {relates_to, 7}", loaded.Relations[1])
	}
}

func TestMarkdownStorage_SaveLoad_WithoutRelations(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)

	now := time.Now().UTC().Truncate(time.Second)
	tsk := &task.Task{
		ID:        1,
		Title:     "Plain task",
		Status:    task.StatusTodo,
		Priority:  task.PriorityHigh,
		Type:      "feature",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := storage.Save(tsk); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify the file doesn't contain "relations" field
	data, err := os.ReadFile(filepath.Join(dir, "001.md"))
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	if strings.Contains(string(data), "relations:") {
		t.Errorf("File should not contain 'relations:' field when empty, got:\n%s", string(data))
	}

	loaded, err := storage.Load(1)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Relations) != 0 {
		t.Errorf("Relations should be empty, got %v", loaded.Relations)
	}
}

func TestIndex_AddRemoveRelation(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	// Add a relation
	edge := task.RelationEdge{Type: "blocked_by", Source: 5, Target: 3}
	idx.AddRelation(edge)

	// Verify it's in the index
	relations := idx.GetRelationsForTask(5)
	if len(relations) != 1 {
		t.Fatalf("GetRelationsForTask(5) = %d, want 1", len(relations))
	}
	if relations[0].Type != "blocked_by" || relations[0].Target != 3 {
		t.Errorf("relation = %v, want {blocked_by, 5, 3}", relations[0])
	}

	// Target task should also see the relation
	relations = idx.GetRelationsForTask(3)
	if len(relations) != 1 {
		t.Fatalf("GetRelationsForTask(3) = %d, want 1", len(relations))
	}

	// Remove the relation
	idx.RemoveRelation(edge)
	relations = idx.GetRelationsForTask(5)
	if len(relations) != 0 {
		t.Errorf("GetRelationsForTask(5) after remove = %d, want 0", len(relations))
	}
}

func TestIndex_SymmetricRelation(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	// Add a relates_to relation (symmetric)
	edge := task.RelationEdge{Type: "relates_to", Source: 5, Target: 7}
	idx.AddRelation(edge)

	// Both tasks should see the relation
	rel5 := idx.GetRelationsForTask(5)
	rel7 := idx.GetRelationsForTask(7)

	if len(rel5) != 2 { // original + reverse
		t.Errorf("GetRelationsForTask(5) = %d, want 2 (original + reverse)", len(rel5))
	}
	if len(rel7) != 2 { // original + reverse
		t.Errorf("GetRelationsForTask(7) = %d, want 2 (original + reverse)", len(rel7))
	}

	// Remove the relation - should remove both edges
	idx.RemoveRelation(edge)
	rel5 = idx.GetRelationsForTask(5)
	rel7 = idx.GetRelationsForTask(7)
	if len(rel5) != 0 {
		t.Errorf("GetRelationsForTask(5) after remove = %d, want 0", len(rel5))
	}
	if len(rel7) != 0 {
		t.Errorf("GetRelationsForTask(7) after remove = %d, want 0", len(rel7))
	}
}

func TestIndex_GetBlockers(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	// Add blocked_by relations
	idx.AddRelation(task.RelationEdge{Type: "blocked_by", Source: 5, Target: 3})
	idx.AddRelation(task.RelationEdge{Type: "blocked_by", Source: 5, Target: 8})
	idx.AddRelation(task.RelationEdge{Type: "relates_to", Source: 5, Target: 10})

	blockers := idx.GetBlockers(5)
	if len(blockers) != 2 {
		t.Fatalf("GetBlockers(5) = %d, want 2", len(blockers))
	}

	// Task 3 should not have any blockers
	blockers = idx.GetBlockers(3)
	if len(blockers) != 0 {
		t.Errorf("GetBlockers(3) = %d, want 0", len(blockers))
	}
}

func TestIndex_RemoveAllRelationsForTask(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	// Add multiple relations involving task 5
	idx.AddRelation(task.RelationEdge{Type: "blocked_by", Source: 5, Target: 3})
	idx.AddRelation(task.RelationEdge{Type: "relates_to", Source: 5, Target: 7})
	idx.AddRelation(task.RelationEdge{Type: "blocked_by", Source: 10, Target: 5})

	removed := idx.RemoveAllRelationsForTask(5)
	if len(removed) < 3 { // at least 3 edges: blocked_by 5->3, relates_to 5->7, relates_to 7->5 (reverse), blocked_by 10->5
		t.Logf("removed %d edges", len(removed))
	}

	// Task 5 should have no relations
	rel5 := idx.GetRelationsForTask(5)
	if len(rel5) != 0 {
		t.Errorf("GetRelationsForTask(5) after remove all = %d, want 0", len(rel5))
	}

	// Task 3 should have no relations pointing at 5
	rel3 := idx.GetRelationsForTask(3)
	if len(rel3) != 0 {
		t.Errorf("GetRelationsForTask(3) after remove all = %d, want 0", len(rel3))
	}
}

func TestIndex_NextTodo_SkipsBlockedTasks(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	now := time.Now().UTC()

	// Create tasks: task 1 (blocker, todo), task 2 (blocked by 1, high priority), task 3 (low priority)
	task1 := &task.Task{ID: 1, Title: "Blocker", Status: task.StatusTodo, Priority: task.PriorityMedium, Type: "feature", CreatedAt: now, UpdatedAt: now}
	task2 := &task.Task{ID: 2, Title: "Blocked", Status: task.StatusTodo, Priority: task.PriorityCritical, Type: "feature", CreatedAt: now, UpdatedAt: now}
	task3 := &task.Task{ID: 3, Title: "Unblocked", Status: task.StatusTodo, Priority: task.PriorityLow, Type: "feature", CreatedAt: now, UpdatedAt: now}

	storage.Save(task1)
	storage.Save(task2)
	storage.Save(task3)

	idx.Set(task1)
	idx.Set(task2)
	idx.Set(task3)

	// Block task 2 by task 1
	idx.AddRelation(task.RelationEdge{Type: "blocked_by", Source: 2, Target: 1})

	// NextTodo should skip task 2 (blocked) and return task 1 (medium priority, higher than task 3)
	next := idx.NextTodo()
	if next == nil {
		t.Fatal("NextTodo() returned nil")
	}
	if next.ID != 1 {
		t.Errorf("NextTodo() ID = %d, want 1 (blocker task, medium priority)", next.ID)
	}

	// Mark task 1 as done - now task 2 should be unblocked
	task1.Status = task.StatusDone
	idx.Set(task1)

	next = idx.NextTodo()
	if next == nil {
		t.Fatal("NextTodo() returned nil after unblocking")
	}
	if next.ID != 2 {
		t.Errorf("NextTodo() ID = %d, want 2 (now unblocked, critical priority)", next.ID)
	}
}

func TestIndex_RebuildWithRelations(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)

	now := time.Now().UTC()

	// Create tasks with relations in frontmatter
	task1 := &task.Task{ID: 1, Title: "Blocker", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature", CreatedAt: now, UpdatedAt: now}
	task2 := &task.Task{
		ID: 2, Title: "Blocked", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature",
		Relations: []task.Relation{{Type: "blocked_by", Task: 1}},
		CreatedAt: now, UpdatedAt: now,
	}
	task3 := &task.Task{
		ID: 3, Title: "Related", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature",
		Relations: []task.Relation{{Type: "relates_to", Task: 1}},
		CreatedAt: now, UpdatedAt: now,
	}

	storage.Save(task1)
	storage.Save(task2)
	storage.Save(task3)

	// Create new index and rebuild from files
	idx := NewIndex(dir, storage)
	if err := idx.Rebuild(); err != nil {
		t.Fatalf("Rebuild() error = %v", err)
	}

	// Verify blocked_by edge
	blockers := idx.GetBlockers(2)
	if len(blockers) != 1 || blockers[0] != 1 {
		t.Errorf("GetBlockers(2) = %v, want [1]", blockers)
	}

	// Verify relates_to generates symmetric edges
	rel1 := idx.GetRelationsForTask(1)
	hasRelatesToFrom1 := false
	hasRelatesToFrom3 := false
	for _, r := range rel1 {
		if r.Type == "relates_to" && r.Source == 1 && r.Target == 3 {
			hasRelatesToFrom1 = true
		}
		if r.Type == "relates_to" && r.Source == 3 && r.Target == 1 {
			hasRelatesToFrom3 = true
		}
	}
	if !hasRelatesToFrom1 {
		t.Error("expected reverse relates_to edge (1->3) in task 1 relations")
	}
	if !hasRelatesToFrom3 {
		t.Error("expected original relates_to edge (3->1) in task 1 relations")
	}
}

func TestIndex_SaveLoadWithRelations(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	now := time.Now().UTC()
	tk := &task.Task{ID: 1, Title: "Test", Status: task.StatusTodo, Priority: task.PriorityHigh, Type: "feature", CreatedAt: now, UpdatedAt: now}
	storage.Save(tk)
	idx.Set(tk)

	// Add some relations
	idx.AddRelation(task.RelationEdge{Type: "blocked_by", Source: 2, Target: 1})
	idx.AddRelation(task.RelationEdge{Type: "relates_to", Source: 1, Target: 3})

	if err := idx.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify the index file contains relations
	data, err := os.ReadFile(filepath.Join(dir, ".index.json"))
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	if !strings.Contains(string(data), "blocked_by") {
		t.Error("Index file should contain blocked_by relation")
	}
	if !strings.Contains(string(data), "relates_to") {
		t.Error("Index file should contain relates_to relation")
	}

	// Load into new index and verify relations are restored
	idx2 := NewIndex(dir, storage)
	if err := idx2.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	blockers := idx2.GetBlockers(2)
	if len(blockers) != 1 || blockers[0] != 1 {
		t.Errorf("GetBlockers(2) after load = %v, want [1]", blockers)
	}
}

func TestIndex_Integration_FullFlow(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)
	idx := NewIndex(dir, storage)

	now := time.Now().UTC()

	// Create tasks with descriptions
	task1 := &task.Task{
		ID:          1,
		Title:       "Task 1",
		Description: "Long description for task 1 that should NOT be in index",
		Status:      task.StatusTodo,
		Priority:    task.PriorityHigh,
		Type:        "feature",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	task2 := &task.Task{
		ID:          2,
		Title:       "Task 2",
		Description: "Another long description",
		Status:      task.StatusDone,
		Priority:    task.PriorityLow,
		Type:        "bug",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Save to both storage and index
	if err := storage.Save(task1); err != nil {
		t.Fatalf("storage.Save() error = %v", err)
	}
	if err := storage.Save(task2); err != nil {
		t.Fatalf("storage.Save() error = %v", err)
	}
	idx.Set(task1)
	idx.Set(task2)
	if err := idx.Save(); err != nil {
		t.Fatalf("idx.Save() error = %v", err)
	}

	// Read index file and verify no descriptions
	data, err := os.ReadFile(filepath.Join(dir, ".index.json"))
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}
	if strings.Contains(string(data), "Long description") {
		t.Error("Index file should NOT contain task descriptions")
	}
	if strings.Contains(string(data), "Another long description") {
		t.Error("Index file should NOT contain task descriptions")
	}

	// Load fresh index
	idx2 := NewIndex(dir, storage)
	if err := idx2.Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// All() should return tasks without descriptions
	all := idx2.All()
	if len(all) != 2 {
		t.Fatalf("All() returned %d tasks, want 2", len(all))
	}
	for _, tk := range all {
		if tk.Description != "" {
			t.Errorf("All() task %d has description %q, want empty", tk.ID, tk.Description)
		}
	}

	// Get() should return full task with description
	full, ok := idx2.Get(1)
	if !ok {
		t.Fatal("Get(1) returned false")
	}
	if full.Description != "Long description for task 1 that should NOT be in index" {
		t.Errorf("Get(1) description = %q, want full description", full.Description)
	}

	// Filter() should return tasks without descriptions
	todoStatus := task.StatusTodo
	filtered := idx2.Filter(&todoStatus, nil, nil, nil)
	if len(filtered) != 1 {
		t.Fatalf("Filter() returned %d tasks, want 1", len(filtered))
	}
	if filtered[0].Description != "" {
		t.Errorf("Filter() task has description %q, want empty", filtered[0].Description)
	}
}
