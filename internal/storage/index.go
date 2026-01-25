package storage

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gpayer/mcp-task-manager/internal/task"
)

// getGitCommit walks up from dir to find .git directory and returns current commit hash
// Returns empty string if not a git repo
func getGitCommit(dir string) (string, error) {
	// Walk up to find .git directory
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	currentDir := absDir
	for {
		gitDir := filepath.Join(currentDir, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			// Found .git directory, run git rev-parse HEAD
			cmd := exec.Command("git", "rev-parse", "HEAD")
			cmd.Dir = currentDir
			output, err := cmd.Output()
			if err != nil {
				return "", nil // Not a valid git repo or no commits yet
			}
			return strings.TrimSpace(string(output)), nil
		}

		// Move up one directory
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			// Reached root without finding .git
			return "", nil
		}
		currentDir = parent
	}
}

// IndexEntry contains task metadata without description (stored in index)
type IndexEntry struct {
	ID        int           `json:"id"`
	ParentID  *int          `json:"parent_id,omitempty"`
	Title     string        `json:"title"`
	Status    task.Status   `json:"status"`
	Priority  task.Priority `json:"priority"`
	Type      string        `json:"type"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// IndexFile is the on-disk format for the index
type IndexFile struct {
	GitCommit string        `json:"git_commit"`
	Tasks     []*IndexEntry `json:"tasks"`
}

// taskToEntry converts a Task to an IndexEntry
func taskToEntry(t *task.Task) *IndexEntry {
	return &IndexEntry{
		ID:        t.ID,
		ParentID:  t.ParentID,
		Title:     t.Title,
		Status:    t.Status,
		Priority:  t.Priority,
		Type:      t.Type,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

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

// NextTodo returns the highest priority todo task
// Priority order:
// 1. Subtasks of in_progress parents (sorted by priority, then creation date)
// 2. Top-level todo tasks and subtasks of todo parents (sorted by priority, then creation date)
// Parents with subtasks are skipped - the real work is in the subtasks
func (idx *Index) NextTodo() *task.Task {
	var inProgressSubtasks []*task.Task
	var otherCandidates []*task.Task

	for _, t := range idx.tasks {
		if t.Status != task.StatusTodo {
			continue
		}
		// Skip parents that have subtasks
		if idx.HasSubtasks(t.ID) {
			continue
		}

		// Check if this is a subtask of an in_progress parent
		if t.ParentID != nil {
			if parent, ok := idx.Get(*t.ParentID); ok && parent.Status == task.StatusInProgress {
				inProgressSubtasks = append(inProgressSubtasks, t)
				continue
			}
		}

		otherCandidates = append(otherCandidates, t)
	}

	// Sort function: by priority (lower order = higher priority), then by creation date
	sortByPriorityAndDate := func(tasks []*task.Task) {
		sort.Slice(tasks, func(i, j int) bool {
			if tasks[i].Priority.Order() != tasks[j].Priority.Order() {
				return tasks[i].Priority.Order() < tasks[j].Priority.Order()
			}
			return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
		})
	}

	// Prioritize subtasks of in_progress parents
	if len(inProgressSubtasks) > 0 {
		sortByPriorityAndDate(inProgressSubtasks)
		return inProgressSubtasks[0]
	}

	if len(otherCandidates) == 0 {
		return nil
	}

	sortByPriorityAndDate(otherCandidates)
	return otherCandidates[0]
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
