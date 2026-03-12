# Task Archiving Design

## Overview

Move completed tasks from the active `tasks/` directory to `tasks/archive/`, removing them from the index to keep list views clean and the index small. Supports both manual archiving and optional time-based auto-archiving.

## Storage

- Archived `.md` files move to `tasks/archive/` (same format, unchanged)
- No archive index — queries against archived tasks do a linear scan of `archive/*.md`
- Active index shrinks as tasks are archived

## Archive Rules

- Only `done` tasks can be archived
- Archiving a parent requires all subtasks to be `done`; archives the entire tree
- All relations to/from archived tasks are cleaned up (same as delete cascade)
- Archived tasks are read-only: can be listed and viewed, but not updated/started/completed

## CLI & MCP Interface

| Command / Tool | Description |
|---|---|
| `archive <id>` / `archive_task` | Manually archive a done task (+ subtasks if parent) |
| `list --archived` / `list_tasks(archived: true)` | List archived tasks (slow, no index) |
| `get <id>` / `get_task` | Transparently checks archive if not found in active tasks |

## Configuration

```yaml
# mcp-tasks.yaml
auto_archive:
  enabled: false
  after_days: 30
```

When enabled, done tasks older than `after_days` (since completion / `updated_at`) are automatically archived.

## Auto-Archive Trigger Points

- **On startup**: scan all done tasks, archive those completed more than `after_days` ago
- **On `complete_task`**: check if the completed task's parent tree is fully done and old enough

## Changes by Area

### `config/config.go`
- Add `AutoArchive` struct (`Enabled bool`, `AfterDays int`) to config
- Parse from `mcp-tasks.yaml`
- Defaults: `enabled: false`, `after_days: 30`

### `storage/markdown.go`
- `Archive(id int) error` — move file from `tasks/` to `tasks/archive/`
- `LoadArchived(id int) (*task.Task, error)` — load from archive dir
- `LoadAllArchived() ([]*task.Task, error)` — scan all archived `.md` files
- `archivePath(id int) string` — helper for archive file path

### `storage/index.go`
- On archive: remove entry + clean up relations (reuse existing `Delete` + `RemoveAllRelationsForTask`)
- No archive index needed

### `task/service.go`
- `ArchiveTask(id int) error` — validate done status, handle subtree, clean relations, move files
- `GetAutoArchiveCandidates() []*task.Task` — find done tasks older than threshold
- `RunAutoArchive() error` — archive all candidates
- Update `Get()` to fall back to archive if task not found in active index
- Update `Initialize()` to run auto-archive on startup when enabled

### `tools/management.go`
- Add `archive_task` MCP tool (params: `id`)
- Add `archived` bool filter to `list_tasks` tool
- Update `get_task` to transparently check archive

### `tools/workflow.go`
- After `complete_task`, trigger auto-archive check if enabled

### `cli/cli.go`
- Add `archive` subcommand (params: task ID)
- Add `--archived` flag to `list` subcommand

### `cli/commands.go`
- `cmdArchive(stdout, stderr, jsonOutput, id)` — call `svc.ArchiveTask()`
- Update `cmdList` to pass archived flag, call archive-aware list
- Update `cmdGet` to handle archive fallback transparently

## Not Included (Future)

- `unarchive` / restore task from archive back to active
- Archive index (only needed if archive query performance becomes a problem)
- Cycle detection for blocked_by chains involving archived tasks
