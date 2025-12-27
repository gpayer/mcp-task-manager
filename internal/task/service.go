package task

import (
	"fmt"
	"time"
)

// Storage interface for task persistence
type Storage interface {
	Save(t *Task) error
	Load(id int) (*Task, error)
	Delete(id int) error
	EnsureDir() error
}

// Index interface for task indexing
type Index interface {
	Load() error
	Save() error
	Get(id int) (*Task, bool)
	Set(t *Task)
	Delete(id int)
	All() []*Task
	Filter(status *Status, priority *Priority, taskType *string, parentID *int) []*Task
	NextTodo() *Task
	NextID() int
	// Subtask methods
	GetSubtasks(parentID int) []*Task
	HasSubtasks(taskID int) bool
	SubtaskCounts(parentID int) (total int, done int)
}

// Service provides task management operations
type Service struct {
	storage    Storage
	index      Index
	validTypes []string
}

// NewService creates a new task service
func NewService(storage Storage, index Index, validTypes []string) *Service {
	return &Service{
		storage:    storage,
		index:      index,
		validTypes: validTypes,
	}
}

// Initialize loads the index and ensures storage is ready
func (s *Service) Initialize() error {
	if err := s.storage.EnsureDir(); err != nil {
		return err
	}
	return s.index.Load()
}

// Create creates a new task (optionally as a subtask)
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

// Get returns a task by ID
func (s *Service) Get(id int) (*Task, error) {
	t, ok := s.index.Get(id)
	if !ok {
		return nil, fmt.Errorf("task not found: %d", id)
	}
	return t, nil
}

// GetWithSubtasks returns a task and its subtasks in one call
func (s *Service) GetWithSubtasks(id int) (*Task, []*Task, error) {
	t, err := s.Get(id)
	if err != nil {
		return nil, nil, err
	}
	subtasks := s.index.GetSubtasks(id)
	return t, subtasks, nil
}

// GetSubtaskCounts returns the count of subtasks for a task
func (s *Service) GetSubtaskCounts(taskID int) (total, done int) {
	return s.index.SubtaskCounts(taskID)
}

// Update modifies a task
func (s *Service) Update(id int, title, description *string, status *Status, priority *Priority, taskType *string) (*Task, error) {
	t, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	if title != nil {
		if *title == "" {
			return nil, fmt.Errorf("title cannot be empty")
		}
		t.Title = *title
	}
	if description != nil {
		t.Description = *description
	}
	if status != nil {
		if !IsValidStatus(string(*status)) {
			return nil, fmt.Errorf("invalid status: %s", *status)
		}
		t.Status = *status
	}
	if priority != nil {
		if !IsValidPriority(string(*priority)) {
			return nil, fmt.Errorf("invalid priority: %s", *priority)
		}
		t.Priority = *priority
	}
	if taskType != nil {
		if !s.isValidType(*taskType) {
			return nil, fmt.Errorf("invalid task type: %s", *taskType)
		}
		t.Type = *taskType
	}

	t.UpdatedAt = time.Now().UTC()

	if err := s.storage.Save(t); err != nil {
		return nil, err
	}

	s.index.Set(t)
	if err := s.index.Save(); err != nil {
		return nil, err
	}

	return t, nil
}

// Delete removes a task
func (s *Service) Delete(id int, deleteSubtasks bool) error {
	t, err := s.Get(id)
	if err != nil {
		return err
	}

	// Check for subtasks
	if s.index.HasSubtasks(id) {
		if !deleteSubtasks {
			total, _ := s.index.SubtaskCounts(id)
			return fmt.Errorf("cannot delete task %d: has %d subtask(s). Use --force to delete this tasks and its subtasks", id, total)
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

// List returns all tasks, optionally filtered
func (s *Service) List(status *Status, priority *Priority, taskType *string, parentID *int) []*Task {
	return s.index.Filter(status, priority, taskType, parentID)
}

// GetNextTask returns the highest priority todo task
func (s *Service) GetNextTask() *Task {
	return s.index.NextTodo()
}

// StartTask moves a task from todo to in_progress
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

// CompleteTask moves a task from in_progress to done
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

// isValidType checks if task type is valid
func (s *Service) isValidType(t string) bool {
	for _, valid := range s.validTypes {
		if t == valid {
			return true
		}
	}
	return false
}
