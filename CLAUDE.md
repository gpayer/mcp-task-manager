# MCP Task Manager

A Go-based MCP server for task management, designed for Claude and coding agents.

**Module path:** `github.com/gpayer/mcp-task-manager`

## Architecture

```
┌─────────────────────────────────────────────┐
│              MCP Server (stdio)             │
├─────────────────────────────────────────────┤
│                 Tool Handlers               │
│  (create, update, list, get_next_task...)   │
├─────────────────────────────────────────────┤
│               Task Service                  │
│    (business logic, validation, sorting)    │
├─────────────────────────────────────────────┤
│                  Storage                    │
│  (markdown files + JSON index cache)        │
└─────────────────────────────────────────────┘
         │                        │
         ▼                        ▼
    ./tasks/*.md           ./tasks/.index.json
```

## Design Decisions

### Storage
- **Source of truth:** Markdown files with YAML frontmatter (`./tasks/*.md`)
- **Index cache:** `./tasks/.index.json` - rebuildable from .md files on startup
- **Location:** Project-local by default, configurable via `MCP_TASKS_DIR` env var
- **Config file:** `./mcp-tasks.yaml`

### Task Schema

```yaml
---
id: 42
title: "Task title"
status: todo          # todo | in_progress | done
priority: high        # critical | high | medium | low
type: feature         # configurable, defaults: feature, bug
created_at: 2025-01-15T10:30:00Z
updated_at: 2025-01-15T10:30:00Z
---

Markdown description here.
```

### Task Identification
- Numeric auto-incrementing IDs
- Filenames: `001.md`, `002.md`, etc.

### Task Lifecycle
- Simple 3-state workflow: `todo` → `in_progress` → `done`
- Blocked state can be derived (future: via `blocked_by` field)

### Priority Ordering
- Named levels: `critical` > `high` > `medium` > `low`
- Tiebreaker: creation date (older first)

## MCP Tools (MVP)

### Task Management
| Tool | Description |
|------|-------------|
| `create_task` | Create a new task with title, description, priority, type |
| `update_task` | Modify task fields |
| `list_tasks` | List tasks with optional filters (status, priority, type) |
| `get_task` | Get full details of a task by ID |
| `delete_task` | Remove a task |

### Agent Workflow
| Tool | Description |
|------|-------------|
| `get_next_task` | Returns highest priority `todo` task |
| `start_task` | Move task from `todo` to `in_progress` |
| `complete_task` | Move task from `in_progress` to `done` |

## Configuration

`mcp-tasks.yaml`:
```yaml
task_types:
  - feature
  - bug
```

Override data directory: `MCP_TASKS_DIR=/path/to/tasks`

## Dependencies

- **MCP SDK:** `github.com/mark3labs/mcp-go` - Third-party Go MCP implementation
- **YAML parsing:** `gopkg.in/yaml.v3` - For frontmatter and config

## Project Structure

```
mcp-task-manager/
├── cmd/
│   └── mcp-task-manager/
│       └── main.go              # Entry point, MCP server setup
├── internal/
│   ├── config/
│   │   └── config.go            # Config loading (file + env)
│   ├── storage/
│   │   ├── storage.go           # Storage interface
│   │   ├── markdown.go          # Markdown file operations
│   │   └── index.go             # JSON index cache
│   ├── task/
│   │   ├── task.go              # Task model/types
│   │   └── service.go           # Business logic
│   └── tools/
│       ├── tools.go             # Tool registration
│       ├── management.go        # create, update, list, get, delete
│       └── workflow.go          # get_next_task, start, complete
├── mcp-tasks.yaml               # Default config (for reference)
├── go.mod
├── go.sum
├── CLAUDE.md
└── README.md
```

## Behavior & Error Handling

### get_next_task
- Returns highest priority `todo` task (priority order, then oldest first)
- If no `todo` tasks exist, returns "no tasks available" message (not an error)

### Index Cache
- Rebuilt on server startup by scanning all .md files
- Updated in-memory and persisted after each write operation
- Self-healing: if index is missing/corrupt, rebuild from .md files

### Concurrency
- MVP assumes single-server, no locking
- File writes are atomic (write to temp file, then rename)

### Validation
- Task IDs: positive integers
- Status: `todo` | `in_progress` | `done`
- Priority: `critical` | `high` | `medium` | `low`
- Type: must be in configured list (default: `feature`, `bug`)

## Future Considerations (Post-MVP)
- Subtasks
- Comments/history
- Task dependencies (`blocked_by`)
- Additional task types (chore, docs, refactor)
