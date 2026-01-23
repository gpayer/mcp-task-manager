# Index Improvements Design

## Problem Statement

Two issues with current index management:

1. **Stale index after `git pull`** - The index doesn't detect when markdown files have changed via git operations, requiring manual deletion of `tasks/.index.json`

2. **Index bloat** - Full task descriptions are stored in the index when only metadata is needed for most operations

## Solution Overview

1. Store git commit hash in index, rebuild automatically when HEAD changes
2. Remove descriptions from index, load from `.md` files on demand

## Design

### Index File Format

**New format** (`.index.json`):

```json
{
  "git_commit": "abc123def456...",
  "tasks": [
    {
      "id": 1,
      "parent_id": null,
      "title": "Add README.md",
      "status": "done",
      "priority": "medium",
      "type": "feature",
      "created_at": "2025-12-14T16:51:00Z",
      "updated_at": "2025-12-14T20:38:18Z"
    }
  ]
}
```

**Changes from current format:**
- Top-level object with `git_commit` field instead of raw array
- Task entries no longer include `description`

### New Types

In `internal/storage/index.go`:

```go
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

type IndexFile struct {
    GitCommit string        `json:"git_commit"`
    Tasks     []*IndexEntry `json:"tasks"`
}
```

### Index Struct Changes

- Store `map[int]*IndexEntry` instead of `map[int]*task.Task`
- Add method `GetFullTask(id int) (*task.Task, error)` - reads `.md` file, returns complete task with description
- Existing `Get(id int)` returns `*IndexEntry` (metadata only)

### Git Commit Detection

New helper function:

```go
func getGitCommit(dir string) (string, error) {
    // Walk up from dir to find .git directory
    // Run: git rev-parse HEAD
    // Return commit hash or empty string if not a git repo
}
```

**Load behavior:**

```go
func (idx *Index) Load() error {
    data, err := os.ReadFile(idx.indexPath())
    if err != nil {
        return idx.Rebuild()  // Missing index
    }

    var indexFile IndexFile
    if err := json.Unmarshal(data, &indexFile); err != nil {
        return idx.Rebuild()  // Corrupt index
    }

    currentCommit, _ := getGitCommit(idx.dir)
    if currentCommit != "" && indexFile.GitCommit != currentCommit {
        return idx.Rebuild()  // Stale index (git changed)
    }

    // Load entries into memory
    idx.entries = make(map[int]*IndexEntry)
    for _, e := range indexFile.Tasks {
        idx.entries[e.ID] = e
    }
    return nil
}
```

**Edge cases:**
- Not a git repo: skip commit check, use index as-is
- Git command fails: skip commit check, use index as-is
- Empty commit in index file: always rebuild (handles migration from old format)

### Data Access Pattern

Two-tier access:
- **Index**: fast metadata for listing/filtering
- **Disk**: read `.md` file when description is needed

**When to load from disk:**
- `Get()`, `GetWithSubtasks()` - read from disk for full task details
- `List()` - returns tasks without descriptions (metadata only)

### Affected Code

| File | Changes |
|------|---------|
| `internal/storage/index.go` | New `IndexEntry`, `IndexFile` types; git commit check on load; `GetFullTask()` method; change internal map to use `IndexEntry` |
| `internal/task/service.go` | `Get()` and `GetWithSubtasks()` use `GetFullTask()`; `List()` converts entries to tasks without description |
| `internal/cli/cli.go` | Add `rebuild-index` subcommand (optional, later) |

### Backward Compatibility

When loading an old-format index (raw array without `git_commit`), JSON unmarshal into `IndexFile` will fail or produce empty `GitCommit`. This triggers a rebuild automatically - no explicit migration code needed.

### Future: CLI Command

```
mcp-task-manager rebuild-index
```

- Deletes `.index.json` and calls `Rebuild()`
- Escape hatch for corrupted index
- Implementation deferred to later

## Out of Scope

- Git hooks (not needed with commit-based detection)
- In-memory description caching
