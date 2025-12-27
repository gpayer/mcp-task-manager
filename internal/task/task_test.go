package task

import "testing"

func TestPriorityOrder(t *testing.T) {
	tests := []struct {
		priority Priority
		expected int
	}{
		{PriorityCritical, 0},
		{PriorityHigh, 1},
		{PriorityMedium, 2},
		{PriorityLow, 3},
		{Priority("unknown"), 99},
	}

	for _, tt := range tests {
		t.Run(string(tt.priority), func(t *testing.T) {
			if got := tt.priority.Order(); got != tt.expected {
				t.Errorf("Priority(%q).Order() = %d, want %d", tt.priority, got, tt.expected)
			}
		})
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{"todo", true},
		{"in_progress", true},
		{"done", true},
		{"invalid", false},
		{"", false},
		{"TODO", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := IsValidStatus(tt.status); got != tt.valid {
				t.Errorf("IsValidStatus(%q) = %v, want %v", tt.status, got, tt.valid)
			}
		})
	}
}

func TestIsValidPriority(t *testing.T) {
	tests := []struct {
		priority string
		valid    bool
	}{
		{"critical", true},
		{"high", true},
		{"medium", true},
		{"low", true},
		{"invalid", false},
		{"", false},
		{"HIGH", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			if got := IsValidPriority(tt.priority); got != tt.valid {
				t.Errorf("IsValidPriority(%q) = %v, want %v", tt.priority, got, tt.valid)
			}
		})
	}
}

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
