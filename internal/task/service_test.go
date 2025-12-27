package task

import (
	"testing"
	"time"
)

// mockStorage implements Storage interface for testing
type mockStorage struct {
	tasks map[int]*Task
}

func newMockStorage() *mockStorage {
	return &mockStorage{tasks: make(map[int]*Task)}
}

func (m *mockStorage) Save(t *Task) error {
	m.tasks[t.ID] = t
	return nil
}

func (m *mockStorage) Load(id int) (*Task, error) {
	return m.tasks[id], nil
}

func (m *mockStorage) Delete(id int) error {
	delete(m.tasks, id)
	return nil
}

func (m *mockStorage) EnsureDir() error {
	return nil
}

// mockIndex implements Index interface for testing
type mockIndex struct {
	tasks  map[int]*Task
	nextID int
}

func newMockIndex() *mockIndex {
	return &mockIndex{tasks: make(map[int]*Task), nextID: 1}
}

func (m *mockIndex) Load() error { return nil }
func (m *mockIndex) Save() error { return nil }

func (m *mockIndex) Get(id int) (*Task, bool) {
	t, ok := m.tasks[id]
	return t, ok
}

func (m *mockIndex) Set(t *Task) {
	m.tasks[t.ID] = t
	if t.ID >= m.nextID {
		m.nextID = t.ID + 1
	}
}

func (m *mockIndex) Delete(id int) {
	delete(m.tasks, id)
}

func (m *mockIndex) All() []*Task {
	result := make([]*Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		result = append(result, t)
	}
	return result
}

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

func (m *mockIndex) NextTodo() *Task {
	var best *Task
	for _, t := range m.tasks {
		if t.Status != StatusTodo {
			continue
		}
		if best == nil || t.Priority.Order() < best.Priority.Order() ||
			(t.Priority.Order() == best.Priority.Order() && t.CreatedAt.Before(best.CreatedAt)) {
			best = t
		}
	}
	return best
}

func (m *mockIndex) NextID() int {
	return m.nextID
}

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

func TestService_Create(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	if err := svc.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	task, err := svc.Create("Test Task", "Description", PriorityHigh, "feature", nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if task.ID != 1 {
		t.Errorf("ID = %d, want 1", task.ID)
	}
	if task.Title != "Test Task" {
		t.Errorf("Title = %q, want %q", task.Title, "Test Task")
	}
	if task.Status != StatusTodo {
		t.Errorf("Status = %q, want %q", task.Status, StatusTodo)
	}
}

func TestService_Create_Validation(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	tests := []struct {
		name     string
		title    string
		priority Priority
		taskType string
		wantErr  bool
	}{
		{"valid", "Title", PriorityHigh, "feature", false},
		{"empty title", "", PriorityHigh, "feature", true},
		{"invalid priority", "Title", Priority("invalid"), "feature", true},
		{"invalid type", "Title", PriorityHigh, "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(tt.title, "desc", tt.priority, tt.taskType, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	task, _ := svc.Create("Original", "Desc", PriorityMedium, "feature", nil)

	newTitle := "Updated"
	newStatus := StatusInProgress
	updated, err := svc.Update(task.ID, &newTitle, nil, &newStatus, nil, nil)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if updated.Title != "Updated" {
		t.Errorf("Title = %q, want %q", updated.Title, "Updated")
	}
	if updated.Status != StatusInProgress {
		t.Errorf("Status = %q, want %q", updated.Status, StatusInProgress)
	}
}

func TestService_Delete(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	task, _ := svc.Create("To Delete", "Desc", PriorityMedium, "feature", nil)

	if err := svc.Delete(task.ID, false); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if _, err := svc.Get(task.ID); err == nil {
		t.Error("Get() after Delete() should return error")
	}
}

func TestService_TaskWorkflow(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	// Create
	task, _ := svc.Create("Workflow Test", "Desc", PriorityHigh, "feature", nil)
	if task.Status != StatusTodo {
		t.Fatalf("initial status = %q, want todo", task.Status)
	}

	// Start
	task, err := svc.StartTask(task.ID)
	if err != nil {
		t.Fatalf("StartTask() error = %v", err)
	}
	if task.Status != StatusInProgress {
		t.Errorf("status after start = %q, want in_progress", task.Status)
	}

	// Complete
	task, err = svc.CompleteTask(task.ID)
	if err != nil {
		t.Fatalf("CompleteTask() error = %v", err)
	}
	if task.Status != StatusDone {
		t.Errorf("status after complete = %q, want done", task.Status)
	}
}

func TestService_StartTask_InvalidState(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	task, _ := svc.Create("Test", "Desc", PriorityHigh, "feature", nil)

	// Start the task
	svc.StartTask(task.ID)

	// Try to start again - should fail
	_, err := svc.StartTask(task.ID)
	if err == nil {
		t.Error("StartTask() on in_progress task should fail")
	}
}

func TestService_CompleteTask_InvalidState(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	task, _ := svc.Create("Test", "Desc", PriorityHigh, "feature", nil)

	// Try to complete without starting - should fail
	_, err := svc.CompleteTask(task.ID)
	if err == nil {
		t.Error("CompleteTask() on todo task should fail")
	}
}

func TestService_GetNextTask(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	// Empty
	if got := svc.GetNextTask(); got != nil {
		t.Errorf("GetNextTask() on empty = %v, want nil", got)
	}

	// Add tasks with different priorities
	svc.Create("Low", "Desc", PriorityLow, "feature", nil)
	time.Sleep(time.Millisecond) // Ensure different timestamps
	svc.Create("Critical", "Desc", PriorityCritical, "bug", nil)

	next := svc.GetNextTask()
	if next == nil {
		t.Fatal("GetNextTask() returned nil")
	}
	if next.Priority != PriorityCritical {
		t.Errorf("GetNextTask() priority = %q, want critical", next.Priority)
	}
}

func TestService_List(t *testing.T) {
	svc := NewService(newMockStorage(), newMockIndex(), []string{"feature", "bug"})
	svc.Initialize()

	svc.Create("Task 1", "Desc", PriorityHigh, "feature", nil)
	svc.Create("Task 2", "Desc", PriorityLow, "bug", nil)
	svc.Create("Task 3", "Desc", PriorityMedium, "feature", nil)

	// All
	all := svc.List(nil, nil, nil, nil)
	if len(all) != 3 {
		t.Errorf("List() all = %d, want 3", len(all))
	}

	// By type
	featureType := "feature"
	features := svc.List(nil, nil, &featureType, nil)
	if len(features) != 2 {
		t.Errorf("List() by feature = %d, want 2", len(features))
	}
}

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
