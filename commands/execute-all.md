---
description: Execute all pending tasks using superpowers workflow
---

Use the superpowers-workflow skill to execute all pending tasks from the task manager MCP server.

The workflow will:
1. Get the next todo task from the task manager
2. If it's a parent task without subtasks: dispatch the configured `planner` custom agent to create implementation subtasks
3. Execute each subtask with the configured `coder` custom agent and review it with the configured `reviewer` custom agent when the workflow requires review
4. If a preferred custom role agent is unavailable: stop and ask the user before allowing any fallback
5. Continue until all tasks are complete
