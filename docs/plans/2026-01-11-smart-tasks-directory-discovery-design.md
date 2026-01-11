# Smart Tasks Directory Discovery

## Overview

The task manager will intelligently locate an existing tasks directory instead of always creating one in the current directory. This prevents accidental creation of orphan task directories when running commands from the wrong location.

## Problem Statement

Currently, the `tasks` directory is always created implicitly in the current working directory, even when running read-only commands. This causes issues when:

1. Running commands from a subdirectory of a project that already has tasks
2. Running commands in the wrong directory accidentally
3. Expecting the tool to find an existing project automatically

## Design

### Behavior by Command Type

| Scenario | Write commands (`create_task`) | Read commands (`list_tasks`, etc.) |
|----------|-------------------------------|-----------------------------------|
| `MCP_TASKS_DIR` set | Use that path directly, create if needed | Use that path directly, error if not found |
| Found `mcp-tasks.yaml` in parent | Use that directory, create `tasks` subdir if needed | Use that directory, error if `tasks` missing |
| Found `tasks` dir in parent | Use that directory | Use that directory |
| Nothing found | Create `./tasks` in current directory | Error: "No tasks directory found" |

### Search Order

When `MCP_TASKS_DIR` is not set:

1. Look for `mcp-tasks.yaml` in current directory, then each parent up to filesystem root (`/`)
2. If config found, use that directory as project root
3. If no config found, look for `tasks` directory using same upward search
4. If `tasks` directory found, use it directly

### Search Boundaries

- Search continues all the way to filesystem root (`/`)
- No special handling for git repositories, home directories, etc. (avoids edge cases with submodules/worktrees)

### Directory Creation

- Only write commands (`create_task`) implicitly create the `tasks` directory
- When creating, use current working directory (not a parent)
- Do not create `mcp-tasks.yaml` automatically - config remains optional

## Implementation

### Config Changes (`internal/config/config.go`)

Add `ProjectFound` field to track whether an existing project was discovered:

```go
type Config struct {
    TaskTypes    []string `yaml:"task_types"`
    DataDir      string   `yaml:"-"`
    ProjectFound bool     `yaml:"-"` // Whether an existing project was discovered
}
```

Add `FindProjectRoot()` function:

```go
// FindProjectRoot searches for an existing project root by looking for
// mcp-tasks.yaml or a tasks directory, starting from cwd and moving up.
// Returns the directory containing the config/tasks, or empty string if not found.
func FindProjectRoot() (string, error) {
    cwd, err := os.Getwd()
    if err != nil {
        return "", err
    }

    dir := cwd
    for {
        // First priority: mcp-tasks.yaml
        if _, err := os.Stat(filepath.Join(dir, "mcp-tasks.yaml")); err == nil {
            return dir, nil
        }

        // Second priority: tasks directory
        if info, err := os.Stat(filepath.Join(dir, "tasks")); err == nil && info.IsDir() {
            return dir, nil
        }

        // Move to parent
        parent := filepath.Dir(dir)
        if parent == dir {
            // Reached filesystem root
            return "", nil
        }
        dir = parent
    }
}
```

Modify `Load()` to use `FindProjectRoot()` when `MCP_TASKS_DIR` is not set.

### Storage Changes (`internal/storage/markdown.go`)

Remove automatic `EnsureDir()` call from `Save()`. Directory creation becomes the caller's responsibility.

### Service Changes (`internal/task/service.go`)

Add directory creation logic for write operations:

```go
func (s *Service) Create(t *Task) error {
    // Only create directory on write operations
    if err := s.storage.EnsureDir(); err != nil {
        return err
    }
    // ... rest of create logic
}
```

Add validation for read operations:

```go
func (s *Service) EnsureProjectExists() error {
    if !s.config.ProjectFound {
        return fmt.Errorf("no tasks directory found. Create a task to initialize one here, or set MCP_TASKS_DIR")
    }
    return nil
}
```

### Tool/CLI Integration

**Read-only operations** (must call `EnsureProjectExists()` first):
- `list_tasks`
- `get_task`
- `get_next_task`

**Write operations** (create directory if needed via service layer):
- `create_task` - creates directory if needed

**Operations requiring existing task** (implicitly require project):
- `update_task`
- `delete_task`
- `start_task`
- `complete_task`

### Config Loading Flow

```
Load() called
    |
    v
MCP_TASKS_DIR set? --yes--> Use that path, ProjectFound = true
    | no
    v
FindProjectRoot()
    |
    v
Found? --yes--> Use found path, ProjectFound = true
    | no
    v
Use "./tasks", ProjectFound = false
```

## Error Messages

**No project found (read command):**
```
No tasks directory found.
Create a task to initialize one here, or set MCP_TASKS_DIR.
```

**Config found but no tasks directory (read command):**
```
No tasks directory found (config file found at /path/to/mcp-tasks.yaml).
Create a task to initialize, or check your configuration.
```

## Edge Cases

| Case | Behavior |
|------|----------|
| `mcp-tasks.yaml` exists but `tasks/` doesn't | Write: create `tasks/` subdir. Read: error with config path hint |
| Symlinked `tasks` directory | Works - `os.Stat` follows symlinks |
| `tasks` is a file, not directory | Ignored during search, continues looking upward |
| No read permission on parent dir | Search stops there, continues to next parent |
| Running from `/` | No upward search possible, uses cwd behavior |

## Testing

- Test upward search finds `mcp-tasks.yaml` in parent
- Test upward search finds `tasks/` directory in parent
- Test `mcp-tasks.yaml` takes precedence over `tasks/` at same level
- Test search stops at filesystem root
- Test `MCP_TASKS_DIR` overrides search completely
- Test read commands error when no project found
- Test write commands create directory in cwd when no project found
- Test error message includes config path when config found but no tasks dir
