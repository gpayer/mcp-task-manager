---
description: Execute all pending tasks using superpowers workflow
---

Use the superpowers-workflow skill to execute all pending tasks from the task manager MCP server.

The workflow will:
1. Get the next todo task from the task manager
2. If it's a parent task without subtasks: create an implementation plan as subtasks
3. Execute each subtask using subagent-driven-development (implementer + spec review + code review)
4. Continue until all tasks are complete
