# Subtasks Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add subtask support allowing tasks to have child tasks with parent reference.

**Architecture:** Flat file storage with `parent_id` field in YAML frontmatter. Single-level nesting only. Subtasks are regular tasks that happen to have a parent.

**Tech Stack:** Go, YAML frontmatter, existing mock-based testing pattern.

---

## Task 1: Add ParentID to Task Model

**Files:**
- Modify: `internal/task/task.go:41-50`
- Modify: `internal/task/task_test.go` (if needed)

**Step 1: Write the test**

```go
// In internal/task/task_test.go (add new test)
func TestTask_ParentID(t *testing.T) {
	task := Task{
		ID:       1,
		ParentID: nil,
		Title:    "Parent task",
	}
	if task.ParentID != nil {
		t.Error("ParentID should be nil for top-level task")
	}

	parentID := 1
	subtask := Task{
		ID:       2,
		ParentID: &parentID,
		Title:    "Subtask",
	}
	if subtask.ParentID == nil || *subtask.ParentID != 1 {
		t.Error("ParentID should be 1 for subtask")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/task/... -run TestTask_ParentID -v`
Expected: FAIL - ParentID field not found

**Step 3: Add ParentID field to Task struct**

```go
// In internal/task/task.go, update Task struct
type Task struct {
	ID          int       `yaml:"id" json:"id"`
	ParentID    *int      `yaml:"parent_id,omitempty" json:"parent_id,omitempty"`
	Title       string    `yaml:"title" json:"title"`
	Description string    `yaml:"-" json:"description"`
	Status      Status    `yaml:"status" json:"status"`
	Priority    Priority  `yaml:"priority" json:"priority"`
	Type        string    `yaml:"type" json:"type"`
	CreatedAt   time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time `yaml:"updated_at" json:"updated_at"`
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/task/... -run TestTask_ParentID -v`
Expected: PASS

**Step 5: Run all tests to ensure no regression**

Run: `go test ./... -v`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/task/task.go internal/task/task_test.go
git commit -m "feat: add ParentID field to Task model"
```

---

## Task 2: Update Markdown Storage for ParentID

**Files:**
- Modify: `internal/storage/markdown.go:44-60` (frontmatter struct)
- Modify: `internal/storage/markdown.go:150-158` (parse struct)
- Modify: `internal/storage/storage_test.go`

**Step 1: Write the test**

```go
// In internal/storage/storage_test.go (add new test)
func TestMarkdownStorage_SaveLoad_WithParentID(t *testing.T) {
	dir := t.TempDir()
	storage := NewMarkdownStorage(dir)

	parentID := 1
	task := &task.Task{
		ID:        2,
		ParentID:  &parentID,
		Title:     "Subtask",
		Status:    task.StatusTodo,
		Priority:  task.PriorityHigh,
		Type:      "feature",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := storage.Save(task); err != nil {
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

	task := &task.Task{
		ID:        1,
		ParentID:  nil,
		Title:     "Top-level task",
		Status:    task.StatusTodo,
		Priority:  task.PriorityHigh,
		Type:      "feature",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := storage.Save(task); err != nil {
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
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/storage/... -run TestMarkdownStorage_SaveLoad_With -v`
Expected: FAIL - ParentID not serialized/parsed

**Step 3: Update Save frontmatter struct**

```go
// In internal/storage/markdown.go, update Save() frontmatter struct (around line 44)
frontmatter := struct {
	ID        int           `yaml:"id"`
	ParentID  *int          `yaml:"parent_id,omitempty"`
	Title     string        `yaml:"title"`
	Status    task.Status   `yaml:"status"`
	Priority  task.Priority `yaml:"priority"`
	Type      string        `yaml:"type"`
	CreatedAt string        `yaml:"created_at"`
	UpdatedAt string        `yaml:"updated_at"`
}{
	ID:        t.ID,
	ParentID:  t.ParentID,
	Title:     t.Title,
	Status:    t.Status,
	Priority:  t.Priority,
	Type:      t.Type,
	CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
}
```

**Step 4: Update parse frontmatter struct**

```go
// In internal/storage/markdown.go, update parse() fm struct (around line 150)
var fm struct {
	ID        int    `yaml:"id"`
	ParentID  *int   `yaml:"parent_id"`
	Title     string `yaml:"title"`
	Status    string `yaml:"status"`
	Priority  string `yaml:"priority"`
	Type      string `yaml:"type"`
	CreatedAt string `yaml:"created_at"`
	UpdatedAt string `yaml:"updated_at"`
}
```

**Step 5: Update parse() Task construction**

```go
// In internal/storage/markdown.go, update Task construction (around line 176)
return &task.Task{
	ID:          fm.ID,
	ParentID:    fm.ParentID,
	Title:       fm.Title,
	Description: strings.TrimSpace(bodyBuf.String()),
	Status:      task.Status(fm.Status),
	Priority:    task.Priority(fm.Priority),
	Type:        fm.Type,
	CreatedAt:   createdAt,
	UpdatedAt:   updatedAt,
}, nil
```

**Step 6: Run tests to verify they pass**

Run: `go test ./internal/storage/... -run TestMarkdownStorage_SaveLoad_With -v`
Expected: PASS

**Step 7: Run all tests**

Run: `go test ./... -v`
Expected: All PASS

**Step 8: Commit**

```bash
git add internal/storage/markdown.go internal/storage/storage_test.go
git commit -m "feat: add ParentID support to markdown storage"
```

---

## Task 3: Add Subtask Helper Methods to Index

**Files:**
- Modify: `internal/storage/index.go`
- Add tests to: `internal/storage/storage_test.go`

**Step 1: Write the test for GetSubtasks**

```go
// In internal/storage/storage_test.go
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
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/storage/... -run "TestIndex_(GetSubtasks|HasSubtasks|SubtaskCounts)" -v`
Expected: FAIL - methods not found

**Step 3: Implement helper methods**

```go
// Add to internal/storage/index.go

// GetSubtasks returns all subtasks of a parent task
func (idx *Index) GetSubtasks(parentID int) []*task.Task {
	var result []*task.Task
	for _, t := range idx.tasks {
		if t.ParentID != nil && *t.ParentID == parentID {
			result = append(result, t)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// HasSubtasks returns true if the task has any subtasks
func (idx *Index) HasSubtasks(taskID int) bool {
	for _, t := range idx.tasks {
		if t.ParentID != nil && *t.ParentID == taskID {
			return true
		}
	}
	return false
}

// SubtaskCounts returns (total, done) counts for a parent task
func (idx *Index) SubtaskCounts(parentID int) (total int, done int) {
	for _, t := range idx.tasks {
		if t.ParentID != nil && *t.ParentID == parentID {
			total++
			if t.Status == task.StatusDone {
				done++
			}
		}
	}
	return
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/storage/... -run "TestIndex_(GetSubtasks|HasSubtasks|SubtaskCounts)" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/storage/index.go internal/storage/storage_test.go
git commit -m "feat: add subtask helper methods to Index"
```

---

## Task 4: Update Index Interface in Service

**Files:**
- Modify: `internal/task/service.go:16-27` (Index interface)
- Modify: `internal/task/service_test.go` (mockIndex)

**Step 1: Update Index interface**

```go
// In internal/task/service.go, update Index interface
type Index interface {
	Load() error
	Save() error
	Get(id int) (*Task, bool)
	Set(t *Task)
	Delete(id int)
	All() []*Task
	Filter(status *Status, priority *Priority, taskType *string) []*Task
	NextTodo() *Task
	NextID() int
	// Subtask methods
	GetSubtasks(parentID int) []*Task
	HasSubtasks(taskID int) bool
	SubtaskCounts(parentID int) (total int, done int)
}
```

**Step 2: Update mockIndex in tests**

```go
// In internal/task/service_test.go, add to mockIndex

func (m *mockIndex) GetSubtasks(parentID int) []*Task {
	var result []*Task
	for _, t := range m.tasks {
		if t.ParentID != nil && *t.ParentID == parentID {
			result = append(result, t)
		}
	}
	return result
}

func (m *mockIndex) HasSubtasks(taskID int) bool {
	for _, t := range m.tasks {
		if t.ParentID != nil && *t.ParentID == taskID {
			return true
		}
	}
	return false
}

func (m *mockIndex) SubtaskCounts(parentID int) (total int, done int) {
	for _, t := range m.tasks {
		if t.ParentID != nil && *t.ParentID == parentID {
			total++
			if t.Status == StatusDone {
				done++
			}
		}
	}
	return
}
```

**Step 3: Run all tests**

Run: `go test ./... -v`
Expected: All PASS

**Step 4: Commit**

```bash
git add internal/task/service.go internal/task/service_test.go
git commit -m "feat: add subtask methods to Index interface"
```

---

## Task 5: Update Create to Support Subtasks

**Files:**
- Modify: `internal/task/service.go:54-87` (Create method)
- Modify: `internal/task/service_test.go`

**Step 1: Write the test**

```go
// In internal/task/service_test.go
func TestService_CreateSubtask(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	// Create parent
	parent, _ := svc.Create("Parent", "Desc", PriorityHigh, "feature", nil)

	// Create subtask
	subtask, err := svc.CreateSubtask("Subtask", "Desc", PriorityMedium, "feature", parent.ID)
	if err != nil {
		t.Fatalf("CreateSubtask() error = %v", err)
	}

	if subtask.ParentID == nil || *subtask.ParentID != parent.ID {
		t.Errorf("ParentID = %v, want %d", subtask.ParentID, parent.ID)
	}
}

func TestService_CreateSubtask_InvalidParent(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	// Try to create subtask with non-existent parent
	_, err := svc.CreateSubtask("Subtask", "Desc", PriorityMedium, "feature", 999)
	if err == nil {
		t.Error("CreateSubtask() with invalid parent should fail")
	}
}

func TestService_CreateSubtask_NestedSubtask(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	parent, _ := svc.Create("Parent", "Desc", PriorityHigh, "feature", nil)
	subtask, _ := svc.CreateSubtask("Subtask", "Desc", PriorityMedium, "feature", parent.ID)

	// Try to create subtask under subtask (should fail - single level only)
	_, err := svc.CreateSubtask("Nested", "Desc", PriorityLow, "feature", subtask.ID)
	if err == nil {
		t.Error("CreateSubtask() under subtask should fail")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/task/... -run TestService_CreateSubtask -v`
Expected: FAIL - CreateSubtask method not found

**Step 3: Update Create method signature and add CreateSubtask**

```go
// In internal/task/service.go

// Create creates a new top-level task
func (s *Service) Create(title, description string, priority Priority, taskType string, parentID *int) (*Task, error) {
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if !IsValidPriority(string(priority)) {
		return nil, fmt.Errorf("invalid priority: %s", priority)
	}
	if !s.isValidType(taskType) {
		return nil, fmt.Errorf("invalid task type: %s", taskType)
	}

	// Validate parent if provided
	if parentID != nil {
		parent, ok := s.index.Get(*parentID)
		if !ok {
			return nil, fmt.Errorf("parent task not found: %d", *parentID)
		}
		if parent.ParentID != nil {
			return nil, fmt.Errorf("cannot create subtask under a subtask (single level only)")
		}
	}

	now := time.Now().UTC()
	t := &Task{
		ID:          s.index.NextID(),
		ParentID:    parentID,
		Title:       title,
		Description: description,
		Status:      StatusTodo,
		Priority:    priority,
		Type:        taskType,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.storage.Save(t); err != nil {
		return nil, err
	}

	s.index.Set(t)
	if err := s.index.Save(); err != nil {
		return nil, err
	}

	return t, nil
}

// CreateSubtask creates a subtask under a parent
func (s *Service) CreateSubtask(title, description string, priority Priority, taskType string, parentID int) (*Task, error) {
	return s.Create(title, description, priority, taskType, &parentID)
}
```

**Step 4: Update existing Create calls in tests**

Update all existing `svc.Create(...)` calls to add `nil` as the last argument.

**Step 5: Run tests to verify they pass**

Run: `go test ./internal/task/... -v`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/task/service.go internal/task/service_test.go
git commit -m "feat: add subtask creation support to Service"
```

---

## Task 6: Update StartTask for Auto-Start Parent

**Files:**
- Modify: `internal/task/service.go:172-184` (StartTask method)
- Modify: `internal/task/service_test.go`

**Step 1: Write the test**

```go
// In internal/task/service_test.go
func TestService_StartTask_AutoStartsParent(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	parent, _ := svc.Create("Parent", "Desc", PriorityHigh, "feature", nil)
	subtask, _ := svc.CreateSubtask("Subtask", "Desc", PriorityMedium, "feature", parent.ID)

	// Start subtask
	_, err := svc.StartTask(subtask.ID)
	if err != nil {
		t.Fatalf("StartTask() error = %v", err)
	}

	// Parent should now be in_progress
	parent, _ = svc.Get(parent.ID)
	if parent.Status != StatusInProgress {
		t.Errorf("parent status = %q, want in_progress", parent.Status)
	}
}

func TestService_StartTask_ParentAlreadyStarted(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	parent, _ := svc.Create("Parent", "Desc", PriorityHigh, "feature", nil)
	svc.StartTask(parent.ID) // Start parent first
	subtask, _ := svc.CreateSubtask("Subtask", "Desc", PriorityMedium, "feature", parent.ID)

	// Start subtask - should not error even though parent is started
	_, err := svc.StartTask(subtask.ID)
	if err != nil {
		t.Fatalf("StartTask() error = %v", err)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/task/... -run TestService_StartTask_Auto -v`
Expected: FAIL - parent not auto-started

**Step 3: Update StartTask method**

```go
// In internal/task/service.go
func (s *Service) StartTask(id int) (*Task, error) {
	t, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	if t.Status != StatusTodo {
		return nil, fmt.Errorf("task %d is not in todo status (current: %s)", id, t.Status)
	}

	// Auto-start parent if this is a subtask
	if t.ParentID != nil {
		parent, err := s.Get(*t.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent task not found: %d", *t.ParentID)
		}
		if parent.Status == StatusTodo {
			status := StatusInProgress
			if _, err := s.Update(*t.ParentID, nil, nil, &status, nil, nil); err != nil {
				return nil, fmt.Errorf("failed to start parent task: %w", err)
			}
		}
	}

	status := StatusInProgress
	return s.Update(id, nil, nil, &status, nil, nil)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/task/... -run TestService_StartTask -v`
Expected: All PASS

**Step 5: Commit**

```bash
git add internal/task/service.go internal/task/service_test.go
git commit -m "feat: auto-start parent when starting subtask"
```

---

## Task 7: Update CompleteTask with Blocking and Auto-Complete

**Files:**
- Modify: `internal/task/service.go:186-199` (CompleteTask method)
- Modify: `internal/task/service_test.go`

**Step 1: Write the test for blocking**

```go
// In internal/task/service_test.go
func TestService_CompleteTask_BlocksIfSubtasksIncomplete(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	parent, _ := svc.Create("Parent", "Desc", PriorityHigh, "feature", nil)
	svc.CreateSubtask("Subtask", "Desc", PriorityMedium, "feature", parent.ID)

	// Start parent
	svc.StartTask(parent.ID)

	// Try to complete parent - should fail
	_, err := svc.CompleteTask(parent.ID)
	if err == nil {
		t.Error("CompleteTask() should fail when subtasks are incomplete")
	}
}

func TestService_CompleteTask_AutoCompletesParent(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	parent, _ := svc.Create("Parent", "Desc", PriorityHigh, "feature", nil)
	sub1, _ := svc.CreateSubtask("Sub 1", "Desc", PriorityMedium, "feature", parent.ID)
	sub2, _ := svc.CreateSubtask("Sub 2", "Desc", PriorityMedium, "feature", parent.ID)

	// Start and complete sub1
	svc.StartTask(sub1.ID)
	svc.CompleteTask(sub1.ID)

	// Parent should still be in_progress
	parent, _ = svc.Get(parent.ID)
	if parent.Status != StatusInProgress {
		t.Errorf("parent status = %q, want in_progress", parent.Status)
	}

	// Start and complete sub2
	svc.StartTask(sub2.ID)
	svc.CompleteTask(sub2.ID)

	// Parent should now be done
	parent, _ = svc.Get(parent.ID)
	if parent.Status != StatusDone {
		t.Errorf("parent status = %q, want done", parent.Status)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/task/... -run "TestService_CompleteTask_(Blocks|Auto)" -v`
Expected: FAIL

**Step 3: Update CompleteTask method**

```go
// In internal/task/service.go
func (s *Service) CompleteTask(id int) (*Task, error) {
	t, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	if t.Status != StatusInProgress {
		return nil, fmt.Errorf("task %d is not in progress (current: %s)", id, t.Status)
	}

	// Check if this task has incomplete subtasks
	subtasks := s.index.GetSubtasks(id)
	incompleteCount := 0
	for _, sub := range subtasks {
		if sub.Status != StatusDone {
			incompleteCount++
		}
	}
	if incompleteCount > 0 {
		return nil, fmt.Errorf("cannot complete task %d: has %d incomplete subtask(s)", id, incompleteCount)
	}

	// Complete this task
	status := StatusDone
	completed, err := s.Update(id, nil, nil, &status, nil, nil)
	if err != nil {
		return nil, err
	}

	// If this is a subtask, check if all siblings are done -> auto-complete parent
	if t.ParentID != nil {
		siblings := s.index.GetSubtasks(*t.ParentID)
		allDone := true
		for _, sib := range siblings {
			if sib.Status != StatusDone {
				allDone = false
				break
			}
		}
		if allDone {
			if _, err := s.Update(*t.ParentID, nil, nil, &status, nil, nil); err != nil {
				return nil, fmt.Errorf("failed to auto-complete parent: %w", err)
			}
		}
	}

	return completed, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/task/... -run TestService_CompleteTask -v`
Expected: All PASS

**Step 5: Commit**

```bash
git add internal/task/service.go internal/task/service_test.go
git commit -m "feat: block completion with incomplete subtasks, auto-complete parent"
```

---

## Task 8: Update Delete with Subtask Protection

**Files:**
- Modify: `internal/task/service.go:148-159` (Delete method)
- Modify: `internal/task/service_test.go`

**Step 1: Write the test**

```go
// In internal/task/service_test.go
func TestService_Delete_BlocksIfHasSubtasks(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	parent, _ := svc.Create("Parent", "Desc", PriorityHigh, "feature", nil)
	svc.CreateSubtask("Subtask", "Desc", PriorityMedium, "feature", parent.ID)

	// Try to delete parent - should fail
	err := svc.Delete(parent.ID, false)
	if err == nil {
		t.Error("Delete() should fail when task has subtasks")
	}
}

func TestService_Delete_ForceDeletesSubtasks(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	parent, _ := svc.Create("Parent", "Desc", PriorityHigh, "feature", nil)
	sub, _ := svc.CreateSubtask("Subtask", "Desc", PriorityMedium, "feature", parent.ID)

	// Force delete parent
	err := svc.Delete(parent.ID, true)
	if err != nil {
		t.Fatalf("Delete(force=true) error = %v", err)
	}

	// Both should be gone
	if _, err := svc.Get(parent.ID); err == nil {
		t.Error("parent should be deleted")
	}
	if _, err := svc.Get(sub.ID); err == nil {
		t.Error("subtask should be deleted")
	}
}

func TestService_Delete_SubtaskAllowed(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	parent, _ := svc.Create("Parent", "Desc", PriorityHigh, "feature", nil)
	sub, _ := svc.CreateSubtask("Subtask", "Desc", PriorityMedium, "feature", parent.ID)

	// Delete subtask - should work
	err := svc.Delete(sub.ID, false)
	if err != nil {
		t.Fatalf("Delete(subtask) error = %v", err)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/task/... -run "TestService_Delete_(Blocks|Force|Subtask)" -v`
Expected: FAIL - Delete signature mismatch

**Step 3: Update Delete method**

```go
// In internal/task/service.go
func (s *Service) Delete(id int, deleteSubtasks bool) error {
	t, err := s.Get(id)
	if err != nil {
		return err
	}

	// Check for subtasks
	if s.index.HasSubtasks(id) {
		if !deleteSubtasks {
			total, _ := s.index.SubtaskCounts(id)
			return fmt.Errorf("cannot delete task %d: has %d subtask(s). Use delete_subtasks to force", id, total)
		}

		// Delete all subtasks first
		subtasks := s.index.GetSubtasks(id)
		for _, sub := range subtasks {
			if err := s.storage.Delete(sub.ID); err != nil {
				return fmt.Errorf("failed to delete subtask %d: %w", sub.ID, err)
			}
			s.index.Delete(sub.ID)
		}
	}

	if err := s.storage.Delete(t.ID); err != nil {
		return err
	}

	s.index.Delete(id)
	return s.index.Save()
}
```

**Step 4: Update existing Delete calls**

Update all existing `svc.Delete(id)` calls to `svc.Delete(id, false)`.

**Step 5: Run tests to verify they pass**

Run: `go test ./internal/task/... -v`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/task/service.go internal/task/service_test.go
git commit -m "feat: add subtask protection to Delete with force option"
```

---

## Task 9: Update NextTodo to Skip Parents with Subtasks

**Files:**
- Modify: `internal/storage/index.go:138-160` (NextTodo method)
- Modify: `internal/storage/storage_test.go`

**Step 1: Write the test**

```go
// In internal/storage/storage_test.go
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
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/storage/... -run TestIndex_NextTodo -v`
Expected: FAIL - returns parent instead of subtask

**Step 3: Update NextTodo method**

```go
// In internal/storage/index.go
func (idx *Index) NextTodo() *task.Task {
	var candidates []*task.Task
	for _, t := range idx.tasks {
		if t.Status != task.StatusTodo {
			continue
		}
		// Skip parents that have subtasks
		if idx.HasSubtasks(t.ID) {
			continue
		}
		candidates = append(candidates, t)
	}

	if len(candidates) == 0 {
		return nil
	}

	// Sort by priority (lower order = higher priority), then by creation date
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Priority.Order() != candidates[j].Priority.Order() {
			return candidates[i].Priority.Order() < candidates[j].Priority.Order()
		}
		return candidates[i].CreatedAt.Before(candidates[j].CreatedAt)
	})

	return candidates[0]
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/storage/... -run TestIndex_NextTodo -v`
Expected: All PASS

**Step 5: Commit**

```bash
git add internal/storage/index.go internal/storage/storage_test.go
git commit -m "feat: NextTodo skips parents with subtasks"
```

---

## Task 10: Update Filter to Support parent_id Filter

**Files:**
- Modify: `internal/storage/index.go:118-136` (Filter method)
- Modify: `internal/task/service.go` (Index interface + List method)
- Modify: `internal/task/service_test.go`

**Step 1: Write the test**

```go
// In internal/storage/storage_test.go
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
		t.Errorf("Filter(parent_id=0) = %d, want 2", len(result))
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/storage/... -run TestIndex_Filter_ByParentID -v`
Expected: FAIL - Filter signature mismatch

**Step 3: Update Filter method signature and implementation**

```go
// In internal/storage/index.go
// Filter returns tasks matching the given criteria
// parentID: nil = all tasks, 0 = top-level only, >0 = subtasks of that parent
func (idx *Index) Filter(status *task.Status, priority *task.Priority, taskType *string, parentID *int) []*task.Task {
	var result []*task.Task
	for _, t := range idx.tasks {
		if status != nil && t.Status != *status {
			continue
		}
		if priority != nil && t.Priority != *priority {
			continue
		}
		if taskType != nil && t.Type != *taskType {
			continue
		}
		if parentID != nil {
			if *parentID == 0 {
				// Top-level only
				if t.ParentID != nil {
					continue
				}
			} else {
				// Subtasks of specific parent
				if t.ParentID == nil || *t.ParentID != *parentID {
					continue
				}
			}
		}
		result = append(result, t)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}
```

**Step 4: Update Index interface**

```go
// In internal/task/service.go
Filter(status *Status, priority *Priority, taskType *string, parentID *int) []*Task
```

**Step 5: Update mockIndex.Filter**

```go
// In internal/task/service_test.go
func (m *mockIndex) Filter(status *Status, priority *Priority, taskType *string, parentID *int) []*Task {
	var result []*Task
	for _, t := range m.tasks {
		if status != nil && t.Status != *status {
			continue
		}
		if priority != nil && t.Priority != *priority {
			continue
		}
		if taskType != nil && t.Type != *taskType {
			continue
		}
		if parentID != nil {
			if *parentID == 0 {
				if t.ParentID != nil {
					continue
				}
			} else {
				if t.ParentID == nil || *t.ParentID != *parentID {
					continue
				}
			}
		}
		result = append(result, t)
	}
	return result
}
```

**Step 6: Update Service.List**

```go
// In internal/task/service.go
func (s *Service) List(status *Status, priority *Priority, taskType *string, parentID *int) []*Task {
	return s.index.Filter(status, priority, taskType, parentID)
}
```

**Step 7: Update existing List calls**

Update all existing `svc.List(status, priority, type)` calls to add `nil` as the fourth argument.

**Step 8: Run tests**

Run: `go test ./... -v`
Expected: All PASS

**Step 9: Commit**

```bash
git add internal/storage/index.go internal/task/service.go internal/task/service_test.go
git commit -m "feat: add parent_id filter to List"
```

---

## Task 11: Add GetWithSubtasks Method

**Files:**
- Modify: `internal/task/service.go`
- Modify: `internal/task/service_test.go`

**Step 1: Write the test**

```go
// In internal/task/service_test.go
func TestService_GetWithSubtasks(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	parent, _ := svc.Create("Parent", "Desc", PriorityHigh, "feature", nil)
	sub1, _ := svc.CreateSubtask("Sub 1", "Desc", PriorityMedium, "feature", parent.ID)
	sub2, _ := svc.CreateSubtask("Sub 2", "Desc", PriorityLow, "feature", parent.ID)

	task, subtasks, err := svc.GetWithSubtasks(parent.ID)
	if err != nil {
		t.Fatalf("GetWithSubtasks() error = %v", err)
	}

	if task.ID != parent.ID {
		t.Errorf("task ID = %d, want %d", task.ID, parent.ID)
	}
	if len(subtasks) != 2 {
		t.Errorf("subtasks count = %d, want 2", len(subtasks))
	}

	// Verify subtask IDs
	ids := map[int]bool{sub1.ID: false, sub2.ID: false}
	for _, s := range subtasks {
		ids[s.ID] = true
	}
	for id, found := range ids {
		if !found {
			t.Errorf("subtask %d not found", id)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/task/... -run TestService_GetWithSubtasks -v`
Expected: FAIL - method not found

**Step 3: Implement GetWithSubtasks**

```go
// In internal/task/service.go

// GetWithSubtasks returns a task and its subtasks
func (s *Service) GetWithSubtasks(id int) (*Task, []*Task, error) {
	t, err := s.Get(id)
	if err != nil {
		return nil, nil, err
	}
	subtasks := s.index.GetSubtasks(id)
	return t, subtasks, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/task/... -run TestService_GetWithSubtasks -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/task/service.go internal/task/service_test.go
git commit -m "feat: add GetWithSubtasks method"
```

---

## Task 12: Update MCP Tools - create_task

**Files:**
- Modify: `internal/tools/management.go:14-35` (create_task tool)
- Modify: `internal/tools/management.go:104-118` (handler)

**Step 1: Update create_task tool registration**

```go
// In internal/tools/management.go
createTool := mcp.NewTool("create_task",
	mcp.WithDescription("Create a new task"),
	mcp.WithString("title",
		mcp.Required(),
		mcp.Description("Task title"),
	),
	mcp.WithString("description",
		mcp.Description("Task description (markdown supported)"),
	),
	mcp.WithString("priority",
		mcp.Required(),
		mcp.Description("Task priority"),
		mcp.Enum("critical", "high", "medium", "low"),
	),
	mcp.WithString("type",
		mcp.Required(),
		mcp.Description("Task type"),
		mcp.Enum(validTypes...),
	),
	mcp.WithNumber("parent_id",
		mcp.Description("Parent task ID (for creating subtasks)"),
	),
)
```

**Step 2: Update createTaskHandler**

```go
// In internal/tools/management.go
func createTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		title := req.GetString("title", "")
		description := req.GetString("description", "")
		priority := task.Priority(req.GetString("priority", ""))
		taskType := req.GetString("type", "")

		var parentID *int
		args := req.GetArguments()
		if _, ok := args["parent_id"]; ok {
			id := req.GetInt("parent_id", 0)
			if id > 0 {
				parentID = &id
			}
		}

		t, err := svc.Create(title, description, priority, taskType, parentID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return taskResult(t)
	}
}
```

**Step 3: Build and test manually**

Run: `go build ./...`
Expected: Success

**Step 4: Commit**

```bash
git add internal/tools/management.go
git commit -m "feat: add parent_id parameter to create_task MCP tool"
```

---

## Task 13: Update MCP Tools - delete_task

**Files:**
- Modify: `internal/tools/management.go:76-84` (delete_task tool)
- Modify: `internal/tools/management.go:172-181` (handler)

**Step 1: Update delete_task tool registration**

```go
// In internal/tools/management.go
deleteTool := mcp.NewTool("delete_task",
	mcp.WithDescription("Delete a task"),
	mcp.WithNumber("id",
		mcp.Required(),
		mcp.Description("Task ID"),
	),
	mcp.WithBoolean("delete_subtasks",
		mcp.Description("Force delete all subtasks (required if task has subtasks)"),
	),
)
```

**Step 2: Update deleteTaskHandler**

```go
// In internal/tools/management.go
func deleteTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := req.GetInt("id", 0)
		deleteSubtasks := req.GetBool("delete_subtasks", false)

		if err := svc.Delete(id, deleteSubtasks); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Task %d deleted", id)), nil
	}
}
```

**Step 3: Build and test**

Run: `go build ./...`
Expected: Success

**Step 4: Commit**

```bash
git add internal/tools/management.go
git commit -m "feat: add delete_subtasks parameter to delete_task MCP tool"
```

---

## Task 14: Update MCP Tools - list_tasks

**Files:**
- Modify: `internal/tools/management.go:85-101` (list_tasks tool)
- Modify: `internal/tools/management.go:184-216` (handler)

**Step 1: Update list_tasks tool registration**

```go
// In internal/tools/management.go
listTool := mcp.NewTool("list_tasks",
	mcp.WithDescription("List tasks with optional filters"),
	mcp.WithString("status",
		mcp.Description("Filter by status"),
		mcp.Enum("todo", "in_progress", "done"),
	),
	mcp.WithString("priority",
		mcp.Description("Filter by priority"),
		mcp.Enum("critical", "high", "medium", "low"),
	),
	mcp.WithString("type",
		mcp.Description("Filter by task type"),
		mcp.Enum(validTypes...),
	),
	mcp.WithNumber("parent_id",
		mcp.Description("Filter by parent (0 for top-level only, >0 for subtasks of that parent)"),
	),
)
```

**Step 2: Update listTasksHandler**

```go
// In internal/tools/management.go
func listTasksHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var status *task.Status
		var priority *task.Priority
		var taskType *string
		var parentID *int

		args := req.GetArguments()
		if _, ok := args["status"]; ok {
			s := task.Status(req.GetString("status", ""))
			status = &s
		}
		if _, ok := args["priority"]; ok {
			p := task.Priority(req.GetString("priority", ""))
			priority = &p
		}
		if _, ok := args["type"]; ok {
			v := req.GetString("type", "")
			taskType = &v
		}
		if _, ok := args["parent_id"]; ok {
			id := req.GetInt("parent_id", -1)
			if id >= 0 {
				parentID = &id
			}
		}

		tasks := svc.List(status, priority, taskType, parentID)

		if len(tasks) == 0 {
			return mcp.NewToolResultText("No tasks found"), nil
		}

		data, err := json.MarshalIndent(tasks, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}
```

**Step 3: Build and test**

Run: `go build ./...`
Expected: Success

**Step 4: Commit**

```bash
git add internal/tools/management.go
git commit -m "feat: add parent_id filter to list_tasks MCP tool"
```

---

## Task 15: Update MCP Tools - get_task to Include Subtasks

**Files:**
- Modify: `internal/tools/management.go:120-131` (handler)

**Step 1: Create response struct and update handler**

```go
// In internal/tools/management.go

// TaskWithSubtasks is the response for get_task
type TaskWithSubtasks struct {
	*task.Task
	Subtasks []*task.Task `json:"subtasks,omitempty"`
}

func getTaskHandler(svc *task.Service) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := req.GetInt("id", 0)

		t, subtasks, err := svc.GetWithSubtasks(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		response := TaskWithSubtasks{
			Task:     t,
			Subtasks: subtasks,
		}

		data, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(data)), nil
	}
}
```

**Step 2: Build and test**

Run: `go build ./...`
Expected: Success

**Step 3: Commit**

```bash
git add internal/tools/management.go
git commit -m "feat: get_task includes subtasks in response"
```

---

## Task 16: Update CLI - create command

**Files:**
- Modify: `internal/cli/cli.go` (add --parent flag)
- Modify: `internal/cli/commands.go:129-153` (cmdCreate)

**Step 1: Check cli.go for create command setup**

Read `internal/cli/cli.go` to understand how flags are set up.

**Step 2: Add parent flag to create command**

```go
// In the create subcommand setup, add:
var parentID int
createCmd.Int(&parentID, "p", "parent", "Parent task ID (for creating subtasks)")
```

**Step 3: Update cmdCreate signature and implementation**

```go
// In internal/cli/commands.go
func cmdCreate(stdout, stderr io.Writer, jsonOutput bool, title, priority, taskType, description string, parentID int) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	var parentPtr *int
	if parentID > 0 {
		parentPtr = &parentID
	}

	t, err := svc.Create(title, description, task.Priority(priority), taskType, parentPtr)
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

**Step 4: Update cli.go to pass parentID**

**Step 5: Build and test**

Run: `go build ./... && ./mcp-task-manager create --help`
Expected: Shows --parent flag

**Step 6: Commit**

```bash
git add internal/cli/cli.go internal/cli/commands.go
git commit -m "feat: add --parent flag to CLI create command"
```

---

## Task 17: Update CLI - delete command

**Files:**
- Modify: `internal/cli/cli.go` (add --delete-subtasks flag)
- Modify: `internal/cli/commands.go:203-227` (cmdDelete)

**Step 1: Add delete-subtasks flag**

```go
// In the delete subcommand setup, add:
var deleteSubtasks bool
deleteCmd.Bool(&deleteSubtasks, "s", "delete-subtasks", "Force delete all subtasks")
```

**Step 2: Update cmdDelete**

```go
// In internal/cli/commands.go
func cmdDelete(stdout, stderr io.Writer, jsonOutput bool, id int, deleteSubtasks bool) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if err := svc.Delete(id, deleteSubtasks); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	msg := fmt.Sprintf("Task #%d deleted.", id)
	if jsonOutput {
		if err := FormatJSONMessage(stdout, msg, id); err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
	} else {
		fmt.Fprintln(stdout, msg)
	}

	return 0
}
```

**Step 3: Build and test**

Run: `go build ./... && ./mcp-task-manager delete --help`
Expected: Shows --delete-subtasks flag

**Step 4: Commit**

```bash
git add internal/cli/cli.go internal/cli/commands.go
git commit -m "feat: add --delete-subtasks flag to CLI delete command"
```

---

## Task 18: Update CLI - list command

**Files:**
- Modify: `internal/cli/cli.go` (add --parent flag)
- Modify: `internal/cli/commands.go:31-71` (cmdList)

**Step 1: Add parent flag to list command**

```go
// In the list subcommand setup, add:
var parentID int = -1  // -1 means no filter
listCmd.Int(&parentID, "p", "parent", "Filter by parent (0 for top-level only)")
```

**Step 2: Update cmdList**

```go
// In internal/cli/commands.go
func cmdList(stdout, stderr io.Writer, jsonOutput bool, status, priority, taskType string, parentID int) int {
	svc, _, err := initService()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	var statusPtr *task.Status
	var priorityPtr *task.Priority
	var typePtr *string
	var parentPtr *int

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
	if parentID >= 0 {
		parentPtr = &parentID
	}

	tasks := svc.List(statusPtr, priorityPtr, typePtr, parentPtr)

	if jsonOutput {
		if tasks == nil {
			tasks = []*task.Task{}
		}
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

**Step 3: Build and test**

Run: `go build ./... && ./mcp-task-manager list --help`
Expected: Shows --parent flag

**Step 4: Commit**

```bash
git add internal/cli/cli.go internal/cli/commands.go
git commit -m "feat: add --parent filter to CLI list command"
```

---

## Task 19: Update CLI Output - Show Subtask Counts

**Files:**
- Modify: `internal/cli/output.go` (FormatTaskTable)
- Modify: `internal/cli/output_test.go`

**Step 1: This requires access to index for subtask counts**

The output formatters don't have access to the service/index. Two options:
1. Pass subtask counts as part of Task struct (add computed field)
2. Change FormatTaskTable to accept additional data

For simplicity, we'll add a SubtaskInfo field to Task that gets populated by List.

**Alternative approach:** Create a ListResult type that includes subtask counts.

```go
// In internal/task/task.go, add:
type TaskListItem struct {
	*Task
	SubtaskTotal int `json:"subtask_total,omitempty"`
	SubtaskDone  int `json:"subtask_done,omitempty"`
}
```

**Step 2: Update Service.List to return TaskListItem**

```go
// In internal/task/service.go
func (s *Service) List(status *Status, priority *Priority, taskType *string, parentID *int) []*TaskListItem {
	tasks := s.index.Filter(status, priority, taskType, parentID)
	result := make([]*TaskListItem, len(tasks))
	for i, t := range tasks {
		total, done := s.index.SubtaskCounts(t.ID)
		result[i] = &TaskListItem{
			Task:         t,
			SubtaskTotal: total,
			SubtaskDone:  done,
		}
	}
	return result
}
```

**Step 3: Update FormatTaskTable**

```go
// In internal/cli/output.go
func FormatTaskTable(tasks []*task.TaskListItem) string {
	if len(tasks) == 0 {
		return "No tasks found.\n"
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTitle\tStatus\tPriority\tType\tSubtasks")
	fmt.Fprintln(w, "--\t-----\t------\t--------\t----\t--------")

	for _, t := range tasks {
		subtaskStr := "-"
		if t.SubtaskTotal > 0 {
			subtaskStr = fmt.Sprintf("%d/%d done", t.SubtaskDone, t.SubtaskTotal)
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			t.ID, truncate(t.Title, 40), t.Status, t.Priority, t.Type, subtaskStr)
	}
	w.Flush()
	return buf.String()
}
```

**Step 4: Update all callers and tests**

**Step 5: Build and test**

Run: `go build ./... && ./mcp-task-manager list`
Expected: Shows Subtasks column

**Step 6: Commit**

```bash
git add internal/task/task.go internal/task/service.go internal/cli/output.go internal/cli/commands.go
git commit -m "feat: show subtask counts in CLI list output"
```

---

## Task 20: Final Integration Test

**Files:**
- Create: `internal/task/subtask_integration_test.go`

**Step 1: Write integration test**

```go
// In internal/task/subtask_integration_test.go
package task

import (
	"testing"
)

func TestSubtaskIntegration(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	// Create parent with subtasks
	parent, _ := svc.Create("Implement feature", "Big feature", PriorityHigh, "feature", nil)
	sub1, _ := svc.CreateSubtask("Write tests", "TDD", PriorityHigh, "feature", parent.ID)
	sub2, _ := svc.CreateSubtask("Implement code", "The code", PriorityHigh, "feature", parent.ID)
	sub3, _ := svc.CreateSubtask("Update docs", "Docs", PriorityLow, "feature", parent.ID)

	// get_next_task should return subtask, not parent
	next := svc.GetNextTask()
	if next.ID == parent.ID {
		t.Error("GetNextTask should not return parent with subtasks")
	}

	// Start sub1 - should auto-start parent
	svc.StartTask(sub1.ID)
	parent, _ = svc.Get(parent.ID)
	if parent.Status != StatusInProgress {
		t.Error("Parent should be auto-started")
	}

	// Complete sub1
	svc.CompleteTask(sub1.ID)

	// Try to complete parent - should fail
	_, err := svc.CompleteTask(parent.ID)
	if err == nil {
		t.Error("Should not complete parent with incomplete subtasks")
	}

	// Complete remaining subtasks
	svc.StartTask(sub2.ID)
	svc.CompleteTask(sub2.ID)
	svc.StartTask(sub3.ID)
	svc.CompleteTask(sub3.ID)

	// Parent should be auto-completed
	parent, _ = svc.Get(parent.ID)
	if parent.Status != StatusDone {
		t.Error("Parent should be auto-completed when all subtasks done")
	}

	t.Log("Integration test passed!")
}
```

**Step 2: Run integration test**

Run: `go test ./internal/task/... -run TestSubtaskIntegration -v`
Expected: PASS

**Step 3: Run all tests**

Run: `go test ./... -v`
Expected: All PASS

**Step 4: Commit**

```bash
git add internal/task/subtask_integration_test.go
git commit -m "test: add subtask integration test"
```

---

## Task 21: Update CLAUDE.md Documentation

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Update task schema section**

Add `parent_id` to the schema example and document subtask behavior.

**Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: document subtask support in CLAUDE.md"
```

---

## Summary

21 tasks total covering:
- Data model changes (Tasks 1-4)
- Service logic (Tasks 5-11)
- MCP tool updates (Tasks 12-15)
- CLI updates (Tasks 16-19)
- Testing and documentation (Tasks 20-21)

Each task is self-contained with TDD approach: write test → verify fail → implement → verify pass → commit.
