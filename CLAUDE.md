# MCP Task Manager

A Go-based MCP server for task management, designed for Claude and coding agents. Also works as a standalone CLI tool.

**Module path:** `github.com/gpayer/mcp-task-manager`

## Architecture

```
┌─────────────────────────────────────────────┐
│         Entry Point (main.go)               │
│   (CLI mode if args, MCP server otherwise)  │
├──────────────────────┬──────────────────────┤
│    CLI Commands      │   MCP Tool Handlers  │
│  (list, get, create) │  (create_task, etc.) │
├──────────────────────┴──────────────────────┤
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

## Development Process

### Planning, Code Review & Testing
Use the **task manager MCP tools** to organize work:
- `create_task` - Add new tasks for planned work
- `update_task` - Update task status, priority, or details
- `list_tasks` - Review current task state and priorities
- `delete_task` - Remove obsolete tasks

### Implementation
Use the **task manager MCP tools** to get work assignments:
- `get_next_task` - Get the highest priority todo task
- `get_task` - Get a specific task by ID (when directed)
- `start_task` - Mark a task as in progress before starting work
- `complete_task` - Mark a task as done after finishing

### Code Research & Refactoring
Use the **cclsp MCP tools** (LSP server access) for code navigation:
- `find_definition` - Find where a symbol is defined
- `find_references` - Find all usages of a symbol across the codebase
- `rename_symbol` - Safely rename symbols with LSP support
- `get_diagnostics` - Get language diagnostics (errors, warnings) for a file

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
parent_id: 0          # optional, 0 or omitted = top-level task
relations:            # optional, omitted when empty
  - type: blocked_by
    task: 3
  - type: relates_to
    task: 7
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
- Blocked state derived from `blocked_by` relations (see Relations section)

### Priority Ordering
- Named levels: `critical` > `high` > `medium` > `low`
- Tiebreaker: creation date (older first)

### Subtasks
Tasks support single-level nesting via the `parent_id` field.

**Constraints:**
- Only one level of nesting allowed (subtasks cannot have subtasks)
- Parent task must exist when creating a subtask

**Automatic Behaviors:**
- **Auto-start parent:** Starting a subtask automatically starts its parent (if parent is `todo`)
- **Block parent completion:** Cannot complete a parent task while it has incomplete subtasks
- **Auto-complete parent:** When the last incomplete subtask is completed, the parent is automatically marked `done`

**Delete Protection:**
- Deleting a parent with subtasks requires explicit action:
  - Use `delete_subtasks: true` to cascade delete all subtasks
  - Or delete subtasks individually first

**Agent Workflow Integration:**
- `get_next_task` prioritizes subtasks of `in_progress` parents over other todo tasks (ensures focused completion of started work)
- `get_next_task` skips parent tasks that have incomplete subtasks (returns subtasks instead)
- `list_tasks` shows top-level tasks by default; use `parent_id` filter to list subtasks of a specific parent
- `get_task` includes subtasks in the response for parent tasks

### Relations
Tasks support typed relations to other tasks via the `relations` frontmatter field.

**Relation Types:**

| Type | Semantic | Behavioral effect | Symmetric |
|------|----------|-------------------|-----------|
| `blocked_by` | Source can't proceed until target is done | Affects `get_next_task`, `start_task`, `get_task`, `list_tasks` | No |
| `relates_to` | Informational link | None | Yes |
| `duplicate_of` | Source is a duplicate of target | None | No |

Relation types are configurable via `mcp-tasks.yaml`. Behavioral effects are hardcoded to specific type names (`blocked_by`).

**Storage rules:**
- `blocked_by`: stored only on the blocked task (the source)
- `relates_to`: stored on one side only; the index generates the reverse edge
- `duplicate_of`: stored only on the duplicate (the source)
- Validation: target task must exist, no self-references, no duplicates

**Blocking behavior:**
- `get_next_task` skips tasks with unresolved `blocked_by` relations (target not `done`)
- `start_task` refuses to start a blocked task
- `get_task` includes a derived `blocked` field and blocker details
- `list_tasks` includes a derived `blocked` field per task

**Delete cascade:**
- When deleting a task, all relations referencing it (as source or target) are removed
- Other tasks' frontmatter is updated to remove stale relations

**Interaction with subtasks:**
- Relations and subtasks are orthogonal — a subtask can be blocked by a task outside its parent
- `get_next_task` applies both filters: skip parents with incomplete subtasks AND skip blocked tasks

## MCP Tools

### Task Management
| Tool | Description |
|------|-------------|
| `create_task` | Create a new task with title, description, priority, type, and optional `parent_id` for subtasks |
| `update_task` | Modify task fields |
| `list_tasks` | List tasks with optional filters (status, priority, type, parent_id); top-level tasks by default |
| `get_task` | Get full details of a task by ID (includes subtasks for parent tasks) |
| `delete_task` | Remove a task; use `delete_subtasks: true` to cascade delete subtasks |

### Relations
| Tool | Description |
|------|-------------|
| `add_relation` | Add a relation (`source`, `type`, `target`) between two tasks |
| `remove_relation` | Remove a relation (`source`, `type`, `target`) between two tasks |

### Agent Workflow
| Tool | Description |
|------|-------------|
| `get_next_task` | Returns highest priority `todo` task (skips parents with incomplete subtasks and blocked tasks) |
| `start_task` | Move task from `todo` to `in_progress` (auto-starts parent if subtask; refuses if blocked) |
| `complete_task` | Move task from `in_progress` to `done` (auto-completes parent if last subtask) |

## Configuration

`mcp-tasks.yaml`:
```yaml
task_types:
  - feature
  - bug
relation_types:       # optional, defaults to these three
  - blocked_by
  - relates_to
  - duplicate_of
```

Override data directory: `MCP_TASKS_DIR=/path/to/tasks`

## Dependencies

- **MCP SDK:** `github.com/mark3labs/mcp-go` - Third-party Go MCP implementation
- **YAML parsing:** `gopkg.in/yaml.v3` - For frontmatter and config
- **CLI parsing:** `github.com/integrii/flaggy` - Lightweight CLI argument parser

## Project Structure

```
mcp-task-manager/
├── cmd/
│   └── mcp-task-manager/
│       └── main.go              # Entry point (CLI if args, MCP server otherwise)
├── internal/
│   ├── cli/
│   │   ├── cli.go               # CLI entry point and subcommand setup
│   │   ├── cli_test.go          # CLI tests
│   │   ├── commands.go          # Command handlers (list, get, create, etc.)
│   │   ├── output.go            # Output formatters (table, JSON)
│   │   └── output_test.go       # Output formatter tests
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
│       ├── workflow.go          # get_next_task, start, complete
│       └── relations.go         # add_relation, remove_relation
├── mcp-tasks.yaml               # Default config (for reference)
├── go.mod
├── go.sum
├── CLAUDE.md
└── README.md
```

## Behavior & Error Handling

### get_next_task
- Returns highest priority `todo` task (priority order, then oldest first)
- Skips parent tasks that have incomplete subtasks (returns actionable subtasks instead)
- Skips tasks with unresolved `blocked_by` relations
- If no `todo` tasks exist, returns "no tasks available" message (not an error)

### Index Cache
- Rebuilt on server startup by scanning all .md files
- Updated in-memory and persisted after each write operation
- Self-healing: if index is missing/corrupt, rebuild from .md files
- Includes relation edges (with auto-generated reverse edges for symmetric types)

### Concurrency
- MVP assumes single-server, no locking
- File writes are atomic (write to temp file, then rename)

### Validation
- Task IDs: positive integers
- Status: `todo` | `in_progress` | `done`
- Priority: `critical` | `high` | `medium` | `low`
- Type: must be in configured list (default: `feature`, `bug`)
- Relation type: must be in configured list (default: `blocked_by`, `relates_to`, `duplicate_of`)

## Future Considerations (Post-MVP)
- Comments/history
- Cycle detection for `blocked_by` chains
- Additional task types (chore, docs, refactor)
