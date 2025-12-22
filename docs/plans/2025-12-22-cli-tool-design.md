# CLI Tool Design

## Overview

Add CLI subcommands to mcp-task-manager so it can be used as a standalone command-line tool while maintaining backwards compatibility with MCP server mode.

## Command Structure

```
mcp-task-manager                     # No args → MCP server mode
mcp-task-manager --help              # Show CLI help
mcp-task-manager <subcommand>        # CLI mode
```

### Subcommands

| Command | Description |
|---------|-------------|
| `list` | List tasks with optional filters |
| `get` | Get task details by ID |
| `create` | Create a new task |
| `update` | Update an existing task |
| `delete` | Delete a task |
| `start` | Start a task (todo → in_progress) |
| `complete` | Complete a task (in_progress → done) |
| `next` | Get highest priority todo task |
| `version` | Show version information |

### Global Flag

- `--json` / `-j` - Output in JSON format (available on all subcommands)

### Exit Codes

- 0 = success
- 1 = error (task not found, validation error, etc.)

Errors go to stderr, results go to stdout.

## Subcommand Details

### `list` - List tasks with filters

```
mcp-task-manager list [flags]
  --status, -s     Filter by status (todo|in_progress|done)
  --priority, -p   Filter by priority (critical|high|medium|low)
  --type, -t       Filter by type (feature|bug)
  --json, -j       Output as JSON
```

### `get <id>` - Get single task

```
mcp-task-manager get 3
mcp-task-manager get 3 --json
```

### `create <title>` - Create task

```
mcp-task-manager create "Fix login bug" [flags]
  --priority, -p    Priority (default: medium)
  --type, -t        Type (default: feature)
  --description, -d Description text
  --json, -j        Output created task as JSON
```

### `update <id>` - Update task fields

```
mcp-task-manager update 3 [flags]
  --title           New title
  --status, -s      New status
  --priority, -p    New priority
  --type, -t        New type
  --description, -d New description
  --json, -j        Output updated task as JSON
```

### `delete <id>` - Delete task

```
mcp-task-manager delete 3
```

### `start <id>` / `complete <id>` - Workflow shortcuts

```
mcp-task-manager start 3      # Sets status to in_progress
mcp-task-manager complete 3   # Sets status to done
```

### `next` - Get highest priority todo task

```
mcp-task-manager next         # Shows next task
mcp-task-manager next --json  # JSON output
```

### `version` - Show version

```
mcp-task-manager version      # Prints: mcp-task-manager v0.1.0
```

## Output Formats

### Human-readable (default)

**`list` - Table format:**
```
ID   Title                              Status        Priority   Type
3    Make mcp-task-manager a CLI tool   todo          high       feature
4    Add subtasks support               todo          high       feature
```

**`get` / `next` / `create` / `update` - Detail format:**
```
Task #3
Title:       Make mcp-task-manager a CLI tool
Status:      todo
Priority:    high
Type:        feature
Created:     2025-12-22 21:47:07
Updated:     2025-12-22 21:47:07

Description:
Add CLI subcommands to mcp-task-manager so it can be used as a
standalone command-line tool...
```

**`delete` / `start` / `complete` - Confirmation message:**
```
Task #3 deleted.
Task #3 started.
Task #3 completed.
```

**`next` when no tasks available:**
```
No tasks available.
```

### JSON output (`--json`)

- Single task → JSON object
- List → JSON array
- Messages (delete/start/complete) → `{"message": "Task #3 deleted.", "id": 3}`

## Code Structure

### New package: `internal/cli/`

```
internal/cli/
├── cli.go          # Main CLI setup, flaggy config, Run() entry point
├── commands.go     # Subcommand definitions and handlers
└── output.go       # Human-readable and JSON formatting helpers
```

### Changes to `cmd/mcp-task-manager/main.go`

```go
func main() {
    // Check if CLI mode (any arguments provided)
    if len(os.Args) > 1 {
        cli.Run()  // Handles all CLI commands, exits when done
        return
    }

    // MCP server mode (existing code)
    cfg, err := config.Load()
    // ... rest of current main.go
}
```

### CLI initialization in `cli.Run()`

1. Load config and initialize service (same as MCP mode)
2. Set up flaggy with subcommands
3. Parse arguments
4. Execute matching subcommand handler
5. Exit with appropriate code

### Code Reuse

- `config.Load()` - same config loading
- `task.Service` - all business logic already exists
- Only new code is CLI parsing and output formatting

## Dependencies

- `github.com/integrii/flaggy` - CLI argument parsing

## Decision Log

- **CLI library**: flaggy - lightweight, flexible, no framework overhead
- **Detection logic**: `len(os.Args) > 1` triggers CLI mode
- **Task IDs**: Positional arguments (not flags)
- **Defaults for create**: priority=medium, type=feature
- **Version**: Subcommand (`version`) not flag (`--version`)
