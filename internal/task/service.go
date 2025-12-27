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
	Filter(status *Status, priority *Priority, taskType *string) []*Task
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

// Create creates a new task
func (s *Service) Create(title, description string, priority Priority, taskType string) (*Task, error) {
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if !IsValidPriority(string(priority)) {
		return nil, fmt.Errorf("invalid priority: %s", priority)
	}
	if !s.isValidType(taskType) {
		return nil, fmt.Errorf("invalid task type: %s", taskType)
	}

	now := time.Now().UTC()
	t := &Task{
		ID:          s.index.NextID(),
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

// Get returns a task by ID
func (s *Service) Get(id int) (*Task, error) {
	t, ok := s.index.Get(id)
	if !ok {
		return nil, fmt.Errorf("task not found: %d", id)
	}
	return t, nil
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
func (s *Service) Delete(id int) error {
	if _, err := s.Get(id); err != nil {
		return err
	}

	if err := s.storage.Delete(id); err != nil {
		return err
	}

	s.index.Delete(id)
	return s.index.Save()
}

// List returns all tasks, optionally filtered
func (s *Service) List(status *Status, priority *Priority, taskType *string) []*Task {
	return s.index.Filter(status, priority, taskType)
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

	status := StatusDone
	return s.Update(id, nil, nil, &status, nil, nil)
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
