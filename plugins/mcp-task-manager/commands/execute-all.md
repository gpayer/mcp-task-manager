---
description: Execute all pending tasks using superpowers workflow
---

Use the packaged `superpowers-workflow` skill from this plugin to execute all pending tasks from the task manager MCP server.

The workflow will:
1. Get the next todo task from the task manager
2. If it is a parent task without subtasks, dispatch the configured `planner` agent to create implementation subtasks
3. Execute each subtask with the configured `coder` agent and review coding work with the configured `reviewer` agent when the workflow requires review
4. Stop and ask the user before allowing any fallback if a preferred role agent is unavailable
5. Continue until all tasks are complete

Prerequisites:

- The plugin's `task-manager` MCP server entry is installed and available
- Codex custom agents named `planner`, `coder`, and `reviewer` are installed in Codex agent discovery locations
- Those agents follow the role boundaries described by the packaged `superpowers-workflow` skill
