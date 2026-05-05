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
| `list` | List tasks with optional filters (`-s status`, `-p priority`, `-t type`, where allowed task types depend on config and default to `feature`, `bug`) |
| `get <id>` | Get task details by ID |
| `create <title>` | Create task (defaults: priority=`medium`, type=first configured task type; with default config that is `feature`; allowed task types depend on config and default to `feature`, `bug`); use `--parent` for subtasks |
| `update <id>` | Update task fields, including `type` (allowed task types depend on config and default to `feature`, `bug`) |
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

Use this path for Claude Code specifically. The Claude plugin and the Codex plugin are packaged differently, so the Claude flow still uses an explicit `claude mcp add` step.

**Setup:**

1. Install the `mcp-task-manager` binary so Claude can launch it:

```bash
go install github.com/gpayer/mcp-task-manager/cmd/mcp-task-manager@latest
```

2. Add the MCP server in Claude Code:

```bash
claude mcp add --transport stdio task-manager -- mcp-task-manager
```

3. Install the Claude Code plugin:

```bash
# Add the marketplace
/plugin marketplace add gpayer/mcp-task-manager

# Install the plugin
/plugin install mcp-task-manager@mcp-task-manager
```

**Usage:**

Use the Claude Code command `/mcp-task-manager:superpowers-workflow` to automatically execute pending tasks with installed `planner`, `coder`, and `reviewer` agents. If one of those agents cannot be used, the workflow stops and asks before allowing a fallback.

### Codex Integration

Use this path for Codex specifically. This repository now acts as a Codex marketplace root: the marketplace catalog lives in `.agents/plugins/marketplace.json`, and the installable Codex plugin package is `plugins/mcp-task-manager/`.

**Prerequisite: install the MCP server binary first**

The Codex plugin package includes `superpowers-workflow`, `/execute-all`, `/install-agents`, packaged role agents, and a packaged `.mcp.json`, but it still expects the `mcp-task-manager` executable to already be available on your `PATH`:

```bash
go install github.com/gpayer/mcp-task-manager/cmd/mcp-task-manager@latest
```

**Add this marketplace and install the plugin**

```bash
codex plugin marketplace add https://github.com/gpayer/mcp-task-manager
```

Inside Codex, install the packaged plugin from that marketplace:

```text
/plugin install mcp-task-manager@mcp-task-manager
```

The plugin package wires in the MCP server definition from `plugins/mcp-task-manager/.mcp.json`, so you do not need a separate `codex mcp add` step as long as `mcp-task-manager` is already installed and resolvable by name.

**Install the role agents globally**

Codex discovers reusable subagents from `~/.codex/agents/` or project-local `.codex/agents/`. After installing or upgrading the plugin, run the packaged installer command once:

```text
/install-agents
```

This creates `planner`, `coder`, and `reviewer` symlinks in `~/.codex/agents/` pointing at the installed plugin's packaged agent definitions. Restart Codex after running it so the agents are available in every session.

**Usage**

Use the Codex skill `$superpowers-workflow` or the packaged command `/execute-all` to automatically execute pending tasks with installed `planner`, `coder`, and `reviewer` agents. If one of those agents cannot be used, the workflow stops and asks before allowing a fallback.

## MCP Tools

### Task Management

| Tool | Description |
|------|-------------|
| `create_task` | Create a new task with title, description, priority, `type`, and optional `parent_id` for subtasks. Allowed task `type` values come from config and default to `feature`, `bug`. |
| `update_task` | Modify task fields (title, description, status, priority, `type`). Allowed task `type` values come from config and default to `feature`, `bug`. |
| `list_tasks` | List tasks with optional filters (status, priority, `type`); use `parent_id` filter for subtasks. Allowed task `type` values come from config and default to `feature`, `bug`. |
| `get_task` | Get full details of a task by ID (includes subtasks for parent tasks) |
| `delete_task` | Remove a task; use `delete_subtasks` to cascade |

### Agent Workflow

| Tool | Description |
|------|-------------|
| `get_next_task` | Returns highest priority `todo` task |
| `start_task` | Move task from `todo` to `in_progress` |
| `complete_task` | Move task from `in_progress` to `done` |

### Relations

| Tool | Description |
|------|-------------|
| `add_relation` | Add a relation between two tasks. Allowed relation `type` values come from config and default to `blocked_by`, `relates_to`, `duplicate_of`. |
| `remove_relation` | Remove a relation between two tasks. Allowed relation `type` values come from config and default to `blocked_by`, `relates_to`, `duplicate_of`. |

## Configuration

### Config File

Create `mcp-tasks.yaml` in the working directory:

```yaml
task_types:
  - feature
  - bug
  - chore
  - docs
relation_types:
  - blocked_by
  - relates_to
  - duplicate_of
```

The `task_types` list defines the allowed values for every task `type` field in the CLI, MCP tools, and task frontmatter. If omitted, the default allowed values are `feature` and `bug`.
The `relation_types` list defines the allowed values for every relation `type` field in MCP tools and task metadata. If omitted, the default allowed values are `blocked_by`, `relates_to`, and `duplicate_of`.

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

The `type` field must be one of the configured `task_types` values. With the default configuration, allowed values are `feature` and `bug`.

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
