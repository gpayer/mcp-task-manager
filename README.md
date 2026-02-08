# MCP Task Manager

A Go-based MCP (Model Context Protocol) server for task management, designed for Claude and coding agents.

## Overview

MCP Task Manager provides a simple but powerful task management system that integrates with AI coding assistants via the Model Context Protocol. Tasks are stored as human-readable Markdown files with YAML frontmatter, making them easy to version control and inspect.

### Features

- **Markdown-based storage** - Tasks stored as `.md` files with YAML frontmatter
- **Priority-based workflow** - Critical > High > Medium > Low, with oldest-first tiebreaker
- **Agent-friendly tools** - `get_next_task`, `start_task`, `complete_task` for automated workflows
- **Self-healing index** - JSON index cache rebuilds automatically from source files
- **Configurable task types** - Default: `feature`, `bug`; extensible via config

## Installation

### Prerequisites

- Go 1.21 or later

### Install with Go

```bash
go install github.com/gpayer/mcp-task-manager/cmd/mcp-task-manager@latest
```

### Download Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/gpayer/mcp-task-manager/releases):

- `mcp-task-manager-linux-amd64` - Linux (x86_64)
- `mcp-task-manager-linux-arm64` - Linux (ARM64)
- `mcp-task-manager-windows-amd64.exe` - Windows (x86_64)

After downloading, rename and make executable (Linux/macOS):

```bash
mv mcp-task-manager-linux-amd64 mcp-task-manager
chmod +x mcp-task-manager
# Optionally move to a directory in your PATH:
sudo mv mcp-task-manager /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/gpayer/mcp-task-manager.git
cd mcp-task-manager
go build -o mcp-task-manager ./cmd/mcp-task-manager
```

## Usage

### Running the Server

The MCP server communicates via stdio:

```bash
mcp-task-manager
```

### CLI Usage

The same binary also works as a standalone CLI tool when called with arguments:

```bash
# List all tasks
mcp-task-manager list
mcp-task-manager list --status=todo --priority=high
mcp-task-manager list --json

# Get task details
mcp-task-manager get 1
mcp-task-manager get 1 --json

# Create a task
mcp-task-manager create "Fix login bug" -p high -t bug -d "Users can't log in"

# Update a task
mcp-task-manager update 1 --title "New title" -s in_progress

# Delete a task
mcp-task-manager delete 1

# Workflow commands
mcp-task-manager next              # Get highest priority todo task
mcp-task-manager start 1           # Start a task (todo -> in_progress)
mcp-task-manager complete 1        # Complete a task (in_progress -> done)

# Other
mcp-task-manager version
mcp-task-manager --help
```

#### CLI Commands

| Command | Description |
|---------|-------------|
| `list` | List tasks with optional filters (`-s status`, `-p priority`, `-t type`) |
| `get <id>` | Get task details by ID |
| `create <title>` | Create task (defaults: priority=medium, type=feature); use `--parent` for subtasks |
| `update <id>` | Update task fields |
| `delete <id>` | Delete a task |
| `next` | Get highest priority todo task |
| `start <id>` | Move task to in_progress |
| `complete <id>` | Move task to done |
| `version` | Show version |

All commands support `--json` / `-j` for JSON output.

### Claude Desktop Integration

Add to your Claude Desktop configuration (`~/.config/claude/claude_desktop_config.json` on Linux, `~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "task-manager": {
      "command": "/path/to/mcp-task-manager"
    }
  }
}
```

### Claude Code Integration

MCP Task Manager includes a plugin for Claude Code that enables automated task execution with the superpowers plugin.

**Setup:**

1. Add the MCP server:

```bash
claude mcp add --transport stdio task-manager -- mcp-task-manager
```

2. Install the plugin (includes the superpowers-workflow skill):

```bash
# Add the marketplace
/plugin marketplace add gpayer/mcp-task-manager

# Install the plugin
/plugin install mcp-task-manager@mcp-task-manager
```

**Usage:**

Use the `/mcp-task-manager:superpowers-workflow` skill to automatically execute pending tasks with planning, implementation, and code review phases.

### Codex Integration

**Add the MCP server:**

```bash
codex mcp add task-manager -- mcp-task-manager
```

**Paste into Codex to install the skill and command:**

```text
Fetch https://raw.githubusercontent.com/gpayer/mcp-task-manager/main/.codex/INSTALL.md and execute the installation steps exactly.
```

## MCP Tools

### Task Management

| Tool | Description |
|------|-------------|
| `create_task` | Create a new task with title, description, priority, type, and optional `parent_id` for subtasks |
| `update_task` | Modify task fields (title, description, status, priority, type) |
| `list_tasks` | List tasks with optional filters (status, priority, type); use `parent_id` filter for subtasks |
| `get_task` | Get full details of a task by ID (includes subtasks for parent tasks) |
| `delete_task` | Remove a task; use `delete_subtasks` to cascade |

### Agent Workflow

| Tool | Description |
|------|-------------|
| `get_next_task` | Returns highest priority `todo` task |
| `start_task` | Move task from `todo` to `in_progress` |
| `complete_task` | Move task from `in_progress` to `done` |

## Configuration

### Config File

Create `mcp-tasks.yaml` in the working directory:

```yaml
task_types:
  - feature
  - bug
  - chore
  - docs
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MCP_TASKS_DIR` | Directory for task storage | `./tasks` |

## Task Format

Tasks are stored as Markdown files with YAML frontmatter:

```yaml
---
id: 1
title: "Implement user authentication"
status: todo
priority: high
type: feature
created_at: 2025-01-15T10:30:00Z
updated_at: 2025-01-15T10:30:00Z
---

Detailed description in Markdown format.

- Acceptance criteria
- Implementation notes
- Links and references
```

### Status Values

- `todo` - Task is pending
- `in_progress` - Task is actively being worked on
- `done` - Task is completed

### Priority Levels

- `critical` - Highest priority
- `high` - Important tasks
- `medium` - Normal priority (default)
- `low` - Can wait

### Subtasks

Tasks support single-level nesting via the `parent_id` field.

**Creating subtasks:**
```bash
# CLI
mcp-task-manager create "Implement login form" -p high --parent 1

# MCP tool
create_task with parent_id parameter
```

**Automatic behaviors:**
- Starting a subtask auto-starts its parent task
- Completing the last subtask auto-completes the parent
- Parent tasks cannot be completed while subtasks remain incomplete
- `get_next_task` returns subtasks instead of parents with incomplete subtasks

## Project Structure

```
mcp-task-manager/
├── cmd/mcp-task-manager/    # Entry point (MCP server + CLI)
├── internal/
│   ├── cli/                 # CLI command handlers
│   ├── config/              # Configuration loading
│   ├── storage/             # Markdown + index storage
│   ├── task/                # Task model and service
│   └── tools/               # MCP tool handlers
├── tasks/                   # Task storage (created at runtime)
├── mcp-tasks.yaml           # Configuration file
└── CLAUDE.md                # AI assistant instructions
```

## License

MIT License - see LICENSE file for details.
