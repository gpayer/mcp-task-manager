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

// entryToTask converts an IndexEntry back to a Task (without description)
func entryToTask(e *IndexEntry) *task.Task {
	return &task.Task{
		ID:        e.ID,
		ParentID:  e.ParentID,
		Title:     e.Title,
		Status:    e.Status,
		Priority:  e.Priority,
		Type:      e.Type,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
		// Description intentionally empty
	}
}

// Index is an in-memory cache of all tasks
type Index struct {
	entries map[int]*IndexEntry
	dir     string
	storage *MarkdownStorage
}

// NewIndex creates a new index for the given directory
func NewIndex(dir string, storage *MarkdownStorage) *Index {
	return &Index{
		entries: make(map[int]*IndexEntry),
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

	idx.entries = make(map[int]*IndexEntry)
	for _, t := range tasks {
		idx.entries[t.ID] = taskToEntry(t)
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

	idx.entries = make(map[int]*IndexEntry)
	for _, t := range tasks {
		idx.entries[t.ID] = taskToEntry(t)
	}

	return nil
}

// GetEntry returns an entry by ID (metadata only, no description)
func (idx *Index) GetEntry(id int) (*IndexEntry, bool) {
	e, ok := idx.entries[id]
	return e, ok
}

// Get returns a full task by ID (loads description from disk)
func (idx *Index) Get(id int) (*task.Task, bool) {
	if _, ok := idx.entries[id]; !ok {
		return nil, false
	}
	t, err := idx.storage.Load(id)
	if err != nil {
		return nil, false
	}
	return t, true
}

// Set adds or updates a task in the index
func (idx *Index) Set(t *task.Task) {
	idx.entries[t.ID] = taskToEntry(t)
}

// Delete removes a task from the index
func (idx *Index) Delete(id int) {
	delete(idx.entries, id)
}

// All returns all tasks sorted by ID
func (idx *Index) All() []*task.Task {
	tasks := make([]*task.Task, 0, len(idx.entries))
	for _, e := range idx.entries {
		tasks = append(tasks, entryToTask(e))
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
	for _, e := range idx.entries {
		if status != nil && e.Status != *status {
			continue
		}
		if priority != nil && e.Priority != *priority {
			continue
		}
		if taskType != nil && e.Type != *taskType {
			continue
		}
		if parentID != nil {
			if *parentID == 0 {
				// Top-level only
				if e.ParentID != nil {
					continue
				}
			} else {
				// Subtasks of specific parent
				if e.ParentID == nil || *e.ParentID != *parentID {
					continue
				}
			}
		}
		result = append(result, entryToTask(e))
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

	for _, e := range idx.entries {
		if e.Status != task.StatusTodo {
			continue
		}
		// Skip parents that have subtasks
		if idx.HasSubtasks(e.ID) {
			continue
		}

		// Check if this is a subtask of an in_progress parent
		if e.ParentID != nil {
			if parent, ok := idx.GetEntry(*e.ParentID); ok && parent.Status == task.StatusInProgress {
				inProgressSubtasks = append(inProgressSubtasks, entryToTask(e))
				continue
			}
		}

		otherCandidates = append(otherCandidates, entryToTask(e))
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
	for id := range idx.entries {
		if id > maxID {
			maxID = id
		}
	}
	return maxID + 1
}

// GetSubtasks returns all subtasks of a parent task
func (idx *Index) GetSubtasks(parentID int) []*task.Task {
	var result []*task.Task
	for _, e := range idx.entries {
		if e.ParentID != nil && *e.ParentID == parentID {
			result = append(result, entryToTask(e))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result
}

// HasSubtasks returns true if the task has any subtasks
func (idx *Index) HasSubtasks(taskID int) bool {
	for _, e := range idx.entries {
		if e.ParentID != nil && *e.ParentID == taskID {
			return true
		}
	}
	return false
}

// SubtaskCounts returns (total, done) counts for a parent task
func (idx *Index) SubtaskCounts(parentID int) (total int, done int) {
	for _, e := range idx.entries {
		if e.ParentID != nil && *e.ParentID == parentID {
			total++
			if e.Status == task.StatusDone {
				done++
			}
		}
	}
	return
}
