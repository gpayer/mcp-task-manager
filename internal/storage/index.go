package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/gpayer/mcp-task-manager/internal/task"
)

// Index is an in-memory cache of all tasks
type Index struct {
	tasks   map[int]*task.Task
	dir     string
	storage *MarkdownStorage
}

// NewIndex creates a new index for the given directory
func NewIndex(dir string, storage *MarkdownStorage) *Index {
	return &Index{
		tasks:   make(map[int]*task.Task),
		dir:     dir,
		storage: storage,
	}
}

// indexPath returns path to the index file
func (idx *Index) indexPath() string {
	return filepath.Join(idx.dir, ".index.json")
}

// Rebuild scans all markdown files and rebuilds the index
func (idx *Index) Rebuild() error {
	tasks, err := idx.storage.LoadAll()
	if err != nil {
		return err
	}

	idx.tasks = make(map[int]*task.Task)
	for _, t := range tasks {
		idx.tasks[t.ID] = t
	}

	return idx.Save()
}

// Save persists the index to disk
func (idx *Index) Save() error {
	if err := os.MkdirAll(idx.dir, 0755); err != nil {
		return err
	}

	tasks := idx.All()
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := idx.indexPath() + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, idx.indexPath())
}

// Load reads the index from disk (or rebuilds if missing/corrupt)
func (idx *Index) Load() error {
	data, err := os.ReadFile(idx.indexPath())
	if err != nil {
		// Index missing, rebuild from files
		return idx.Rebuild()
	}

	var tasks []*task.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		// Index corrupt, rebuild from files
		return idx.Rebuild()
	}

	idx.tasks = make(map[int]*task.Task)
	for _, t := range tasks {
		idx.tasks[t.ID] = t
	}

	return nil
}

// Get returns a task by ID
func (idx *Index) Get(id int) (*task.Task, bool) {
	t, ok := idx.tasks[id]
	return t, ok
}

// Set adds or updates a task in the index
func (idx *Index) Set(t *task.Task) {
	idx.tasks[t.ID] = t
}

// Delete removes a task from the index
func (idx *Index) Delete(id int) {
	delete(idx.tasks, id)
}

// All returns all tasks sorted by ID
func (idx *Index) All() []*task.Task {
	tasks := make([]*task.Task, 0, len(idx.tasks))
	for _, t := range idx.tasks {
		tasks = append(tasks, t)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
	return tasks
}

// Filter returns tasks matching the given criteria
func (idx *Index) Filter(status *task.Status, priority *task.Priority, taskType *string) []*task.Task {
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
		result = append(result, t)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// NextTodo returns the highest priority todo task
func (idx *Index) NextTodo() *task.Task {
	var candidates []*task.Task
	for _, t := range idx.tasks {
		if t.Status == task.StatusTodo {
			candidates = append(candidates, t)
		}
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

// NextID returns the next available task ID
func (idx *Index) NextID() int {
	maxID := 0
	for id := range idx.tasks {
		if id > maxID {
			maxID = id
		}
	}
	return maxID + 1
}

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
