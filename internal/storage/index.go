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
	GitCommit string              `json:"git_commit"`
	Tasks     []*IndexEntry       `json:"tasks"`
	Relations []task.RelationEdge `json:"relations,omitempty"`
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

// SymmetricRelationType is a relation type where reverse edges are auto-generated
const SymmetricRelationType = "relates_to"

// BlockingRelationType is the relation type that affects task execution order
const BlockingRelationType = "blocked_by"

// Index is an in-memory cache of all tasks
type Index struct {
	entries           map[int]*IndexEntry
	relationsBySource map[int][]task.RelationEdge
	relationsByTarget map[int][]task.RelationEdge
	dir               string
	storage           *MarkdownStorage
}

// NewIndex creates a new index for the given directory
func NewIndex(dir string, storage *MarkdownStorage) *Index {
	return &Index{
		entries:           make(map[int]*IndexEntry),
		relationsBySource: make(map[int][]task.RelationEdge),
		relationsByTarget: make(map[int][]task.RelationEdge),
		dir:               dir,
		storage:           storage,
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
	idx.relationsBySource = make(map[int][]task.RelationEdge)
	idx.relationsByTarget = make(map[int][]task.RelationEdge)

	for _, t := range tasks {
		idx.entries[t.ID] = taskToEntry(t)
	}

	// Build relation edges from task frontmatter
	for _, t := range tasks {
		for _, rel := range t.Relations {
			edge := task.RelationEdge{
				Type:   rel.Type,
				Source: t.ID,
				Target: rel.Task,
			}
			idx.addEdge(edge)
			// Symmetric types generate a reverse edge
			if rel.Type == SymmetricRelationType {
				reverse := task.RelationEdge{
					Type:   rel.Type,
					Source: rel.Task,
					Target: t.ID,
				}
				idx.addEdge(reverse)
			}
		}
	}

	return idx.Save()
}

// Save persists the index to disk
func (idx *Index) Save() error {
	if err := os.MkdirAll(idx.dir, 0755); err != nil {
		return err
	}

	// Get current git commit
	gitCommit, _ := getGitCommit(idx.dir)

	// Build entries slice sorted by ID
	entries := make([]*IndexEntry, 0, len(idx.entries))
	for _, e := range idx.entries {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ID < entries[j].ID
	})

	// Build relations slice
	var relations []task.RelationEdge
	for _, edges := range idx.relationsBySource {
		relations = append(relations, edges...)
	}
	sort.Slice(relations, func(i, j int) bool {
		if relations[i].Source != relations[j].Source {
			return relations[i].Source < relations[j].Source
		}
		if relations[i].Target != relations[j].Target {
			return relations[i].Target < relations[j].Target
		}
		return relations[i].Type < relations[j].Type
	})

	indexFile := IndexFile{
		GitCommit: gitCommit,
		Tasks:     entries,
		Relations: relations,
	}

	data, err := json.MarshalIndent(indexFile, "", "  ")
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
		return idx.Rebuild() // Missing index
	}

	var indexFile IndexFile
	if err := json.Unmarshal(data, &indexFile); err != nil {
		return idx.Rebuild() // Corrupt or old format index
	}

	// Check git commit - rebuild if stale
	currentCommit, _ := getGitCommit(idx.dir)
	if currentCommit != "" && indexFile.GitCommit != currentCommit {
		return idx.Rebuild() // Stale index (git changed)
	}

	// Empty commit in file with non-empty current = stale (migration case)
	if currentCommit != "" && indexFile.GitCommit == "" {
		return idx.Rebuild()
	}

	// Load entries into memory
	idx.entries = make(map[int]*IndexEntry)
	for _, e := range indexFile.Tasks {
		idx.entries[e.ID] = e
	}

	// Load relations into memory
	idx.relationsBySource = make(map[int][]task.RelationEdge)
	idx.relationsByTarget = make(map[int][]task.RelationEdge)
	for _, edge := range indexFile.Relations {
		idx.addEdge(edge)
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
		// Skip blocked tasks (have unresolved blocked_by relations)
		if idx.isBlocked(e.ID) {
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

// isBlocked checks if a task has any unresolved blocked_by relations
func (idx *Index) isBlocked(taskID int) bool {
	for _, e := range idx.relationsBySource[taskID] {
		if e.Type == BlockingRelationType {
			if target, ok := idx.entries[e.Target]; ok && target.Status != task.StatusDone {
				return true
			}
		}
	}
	return false
}

// addEdge adds an edge to both lookup maps (internal helper, no persistence)
func (idx *Index) addEdge(edge task.RelationEdge) {
	idx.relationsBySource[edge.Source] = append(idx.relationsBySource[edge.Source], edge)
	idx.relationsByTarget[edge.Target] = append(idx.relationsByTarget[edge.Target], edge)
}

// removeEdge removes an edge from both lookup maps (internal helper, no persistence)
func (idx *Index) removeEdge(edge task.RelationEdge) {
	// Remove from source map
	src := idx.relationsBySource[edge.Source]
	for i, e := range src {
		if e.Type == edge.Type && e.Source == edge.Source && e.Target == edge.Target {
			idx.relationsBySource[edge.Source] = append(src[:i], src[i+1:]...)
			break
		}
	}
	// Remove from target map
	tgt := idx.relationsByTarget[edge.Target]
	for i, e := range tgt {
		if e.Type == edge.Type && e.Source == edge.Source && e.Target == edge.Target {
			idx.relationsByTarget[edge.Target] = append(tgt[:i], tgt[i+1:]...)
			break
		}
	}
}

// AddRelation adds a relation edge to the index
func (idx *Index) AddRelation(edge task.RelationEdge) {
	idx.addEdge(edge)
	// Symmetric types generate a reverse edge
	if edge.Type == SymmetricRelationType {
		reverse := task.RelationEdge{
			Type:   edge.Type,
			Source: edge.Target,
			Target: edge.Source,
		}
		idx.addEdge(reverse)
	}
}

// RemoveRelation removes a relation edge from the index
func (idx *Index) RemoveRelation(edge task.RelationEdge) {
	idx.removeEdge(edge)
	// Symmetric types also remove the reverse edge
	if edge.Type == SymmetricRelationType {
		reverse := task.RelationEdge{
			Type:   edge.Type,
			Source: edge.Target,
			Target: edge.Source,
		}
		idx.removeEdge(reverse)
	}
}

// GetRelationsForTask returns all edges where task is source OR target
func (idx *Index) GetRelationsForTask(taskID int) []task.RelationEdge {
	seen := make(map[task.RelationEdge]bool)
	var result []task.RelationEdge

	for _, e := range idx.relationsBySource[taskID] {
		if !seen[e] {
			seen[e] = true
			result = append(result, e)
		}
	}
	for _, e := range idx.relationsByTarget[taskID] {
		if !seen[e] {
			seen[e] = true
			result = append(result, e)
		}
	}
	return result
}

// GetBlockers returns target IDs from blocked_by edges where source == taskID
func (idx *Index) GetBlockers(taskID int) []int {
	var blockers []int
	for _, e := range idx.relationsBySource[taskID] {
		if e.Type == BlockingRelationType {
			blockers = append(blockers, e.Target)
		}
	}
	return blockers
}

// RemoveAllRelationsForTask removes all relations where task appears as source or target
// Returns the removed edges so the service knows which other task files to update
func (idx *Index) RemoveAllRelationsForTask(taskID int) []task.RelationEdge {
	var removed []task.RelationEdge

	// Remove edges where task is source
	for _, e := range idx.relationsBySource[taskID] {
		removed = append(removed, e)
		// Remove from target's map
		tgt := idx.relationsByTarget[e.Target]
		for i, te := range tgt {
			if te.Type == e.Type && te.Source == e.Source && te.Target == e.Target {
				idx.relationsByTarget[e.Target] = append(tgt[:i], tgt[i+1:]...)
				break
			}
		}
	}
	delete(idx.relationsBySource, taskID)

	// Remove edges where task is target
	for _, e := range idx.relationsByTarget[taskID] {
		removed = append(removed, e)
		// Remove from source's map
		src := idx.relationsBySource[e.Source]
		for i, se := range src {
			if se.Type == e.Type && se.Source == e.Source && se.Target == e.Target {
				idx.relationsBySource[e.Source] = append(src[:i], src[i+1:]...)
				break
			}
		}
	}
	delete(idx.relationsByTarget, taskID)

	return removed
}
