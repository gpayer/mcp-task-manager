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
	dirty             bool
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

func (idx *Index) reset() {
	idx.entries = make(map[int]*IndexEntry)
	idx.relationsBySource = make(map[int][]task.RelationEdge)
	idx.relationsByTarget = make(map[int][]task.RelationEdge)
	idx.dirty = false
}

func (idx *Index) rebuildFromTasks(tasks []*task.Task) error {
	idx.reset()

	for _, t := range tasks {
		idx.entries[t.ID] = taskToEntry(t)
	}

	// Build relation edges from task frontmatter.
	for _, t := range tasks {
		for _, rel := range t.Relations {
			edge := task.RelationEdge{
				Type:   rel.Type,
				Source: t.ID,
				Target: rel.Task,
			}
			idx.addEdge(edge)
			// Symmetric types generate a reverse edge.
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

	if len(tasks) == 0 {
		idx.dirty = false
		return nil
	}

	return idx.Save()
}

// Rebuild scans all markdown files and rebuilds the index
func (idx *Index) Rebuild() error {
	tasks, err := idx.storage.LoadAll()
	if err != nil {
		return err
	}

	return idx.rebuildFromTasks(tasks)
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
	if err := os.Rename(tmpPath, idx.indexPath()); err != nil {
		return err
	}
	idx.dirty = false
	return nil
}

// Load reads the index from disk (or rebuilds if missing/corrupt)
func (idx *Index) Load() error {
	data, err := os.ReadFile(idx.indexPath())
	if err != nil {
		if os.IsNotExist(err) {
			return idx.Rebuild()
		}
		return err
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
	idx.dirty = false
	return nil
}

func (idx *Index) syncIfStale() {
	if idx.dirty {
		return
	}
	stale, err := idx.isStaleOnDisk()
	if err != nil || !stale {
		return
	}
	_ = idx.Rebuild()
}

func (idx *Index) isStaleOnDisk() (bool, error) {
	entries, err := os.ReadDir(idx.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	indexInfo, err := os.Stat(idx.indexPath())
	indexMissing := os.IsNotExist(err)
	if err != nil && !indexMissing {
		return false, err
	}

	taskCount := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		taskCount++

		if indexMissing {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return false, err
		}
		if info.ModTime().After(indexInfo.ModTime()) {
			return true, nil
		}
	}

	if taskCount == 0 {
		return false, nil
	}

	if indexMissing {
		return len(idx.entries) == 0, nil
	}

	if taskCount != len(idx.entries) {
		return true, nil
	}

	return false, nil
}

// GetEntry returns an entry by ID (metadata only, no description)
func (idx *Index) GetEntry(id int) (*IndexEntry, bool) {
	idx.syncIfStale()
	e, ok := idx.entries[id]
	return e, ok
}

// Get returns a full task by ID (loads description from disk)
func (idx *Index) Get(id int) (*task.Task, bool) {
	idx.syncIfStale()
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
	idx.dirty = true
}

// Delete removes a task from the index
func (idx *Index) Delete(id int) {
	delete(idx.entries, id)
	idx.dirty = true
}

// All returns all tasks sorted by ID
func (idx *Index) All() []*task.Task {
	idx.syncIfStale()
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
	idx.syncIfStale()
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

type nextTodoGroupKey struct {
	priorityOrder    int
	createdAt        time.Time
	id               int
	inProgressParent bool
}

type nextTodoGroup struct {
	key   nextTodoGroupKey
	tasks []*task.Task
}

func (idx *Index) isActionableForNextTodo(e *IndexEntry) bool {
	if e.Status != task.StatusTodo && e.Status != task.StatusInProgress {
		return false
	}
	if idx.HasSubtasks(e.ID) {
		return false
	}
	return !idx.isBlocked(e.ID)
}

func (idx *Index) nextTodoGroupForEntry(e *IndexEntry) (int, nextTodoGroupKey) {
	groupID := e.ID
	key := nextTodoGroupKey{
		priorityOrder:    e.Priority.Order(),
		createdAt:        e.CreatedAt,
		id:               e.ID,
		inProgressParent: e.Status == task.StatusInProgress,
	}

	if e.ParentID != nil {
		if parent, ok := idx.GetEntry(*e.ParentID); ok {
			groupID = parent.ID
			key = nextTodoGroupKey{
				priorityOrder:    parent.Priority.Order(),
				createdAt:        parent.CreatedAt,
				id:               parent.ID,
				inProgressParent: parent.Status == task.StatusInProgress,
			}
		}
	}

	return groupID, key
}

// NextTodo returns the highest priority actionable todo task.
// Parent tasks with subtasks are skipped. Subtasks inherit their parent's
// priority, creation date, and ID for group selection, then compete within the
// winning group by their own priority, creation date, and ID.
func (idx *Index) NextTodo() *task.Task {
	idx.syncIfStale()
	groups := make(map[int]*nextTodoGroup)

	for _, e := range idx.entries {
		if !idx.isActionableForNextTodo(e) {
			continue
		}

		candidate := entryToTask(e)
		groupID, key := idx.nextTodoGroupForEntry(e)

		group, ok := groups[groupID]
		if !ok {
			group = &nextTodoGroup{key: key}
			groups[groupID] = group
		}
		group.tasks = append(group.tasks, candidate)
	}

	if len(groups) == 0 {
		return nil
	}

	groupList := make([]*nextTodoGroup, 0, len(groups))
	for _, group := range groups {
		groupList = append(groupList, group)
	}

	sort.Slice(groupList, func(i, j int) bool {
		left := groupList[i].key
		right := groupList[j].key
		if left.inProgressParent != right.inProgressParent {
			return left.inProgressParent
		}
		if left.priorityOrder != right.priorityOrder {
			return left.priorityOrder < right.priorityOrder
		}
		if !left.createdAt.Equal(right.createdAt) {
			return left.createdAt.Before(right.createdAt)
		}
		return left.id < right.id
	})

	winningGroup := groupList[0].tasks
	sort.Slice(winningGroup, func(i, j int) bool {
		if winningGroup[i].Status != winningGroup[j].Status {
			return winningGroup[i].Status == task.StatusInProgress
		}
		if winningGroup[i].Priority.Order() != winningGroup[j].Priority.Order() {
			return winningGroup[i].Priority.Order() < winningGroup[j].Priority.Order()
		}
		if !winningGroup[i].CreatedAt.Equal(winningGroup[j].CreatedAt) {
			return winningGroup[i].CreatedAt.Before(winningGroup[j].CreatedAt)
		}
		return winningGroup[i].ID < winningGroup[j].ID
	})

	return winningGroup[0]
}

// NextID returns the next available task ID
func (idx *Index) NextID() int {
	idx.syncIfStale()
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
	idx.syncIfStale()
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
	idx.syncIfStale()
	for _, e := range idx.entries {
		if e.ParentID != nil && *e.ParentID == taskID {
			return true
		}
	}
	return false
}

// SubtaskCounts returns (total, done) counts for a parent task
func (idx *Index) SubtaskCounts(parentID int) (total int, done int) {
	idx.syncIfStale()
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
	idx.syncIfStale()
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
	idx.dirty = true
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
	idx.dirty = true
}

// GetRelationsForTask returns all edges where task is source OR target
func (idx *Index) GetRelationsForTask(taskID int) []task.RelationEdge {
	idx.syncIfStale()
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
	idx.syncIfStale()
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

	idx.dirty = true
	return removed
}
