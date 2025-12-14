package task

import "time"

// Status represents the current state of a task
type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

// Priority represents task priority level
type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"
)

// PriorityOrder returns numeric order for sorting (lower = higher priority)
func (p Priority) Order() int {
	switch p {
	case PriorityCritical:
		return 0
	case PriorityHigh:
		return 1
	case PriorityMedium:
		return 2
	case PriorityLow:
		return 3
	default:
		return 99
	}
}

// Task represents a single task
type Task struct {
	ID          int       `yaml:"id" json:"id"`
	Title       string    `yaml:"title" json:"title"`
	Description string    `yaml:"-" json:"description"` // Stored in markdown body
	Status      Status    `yaml:"status" json:"status"`
	Priority    Priority  `yaml:"priority" json:"priority"`
	Type        string    `yaml:"type" json:"type"`
	CreatedAt   time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time `yaml:"updated_at" json:"updated_at"`
}

// IsValidStatus checks if status is valid
func IsValidStatus(s string) bool {
	switch Status(s) {
	case StatusTodo, StatusInProgress, StatusDone:
		return true
	}
	return false
}

// IsValidPriority checks if priority is valid
func IsValidPriority(p string) bool {
	switch Priority(p) {
	case PriorityCritical, PriorityHigh, PriorityMedium, PriorityLow:
		return true
	}
	return false
}
