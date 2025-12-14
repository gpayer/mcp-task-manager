---
name: task-manager-workflow
description: Use when executing all todo tasks from the task manager MCP - iterates through tasks using get_next_task, creates plans with writing-plans skill, executes with subagents, and marks complete
---

# Task Manager Workflow

## Overview

Execute all `todo` tasks from the task manager MCP server. For each task: create a detailed implementation plan, execute it using subagents, and mark complete. Continue until no tasks remain.

**Core principle:** Automated task queue processing with planning and subagent execution.

**Announce at start:** "I'm using the task-manager-workflow skill to execute all pending tasks."

## Prerequisites

- Task manager MCP server must be running and accessible
- Tasks must exist in `todo` status
- Use `mcp__task-manager__list_tasks` with `status: todo` to verify tasks exist

## The Process

### Step 1: Get Next Task

1. Call `mcp__task-manager__get_next_task`
2. If no tasks available: Announce completion and stop
3. If task found: Proceed to Step 2

### Step 2: Start Task

1. Call `mcp__task-manager__start_task` with the task ID
2. Create TodoWrite entry for tracking

### Step 3: Create Implementation Plan

**REQUIRED SUB-SKILL:** Use superpowers:writing-plans

1. Announce: "I'm using the writing-plans skill to create the implementation plan for Task #[ID]: [Title]"
2. Read relevant codebase files to understand context
3. Create detailed plan in `docs/plans/YYYY-MM-DD-<feature-name>.md`
4. Plan must have bite-sized tasks with exact file paths and code

### Step 4: Execute Plan with Subagents

For each task in the plan:

1. Launch a Task tool subagent with `model: sonnet`:
   ```
   Execute Task N from the plan at [plan-path]

   Your task is to:
   1. [Specific steps from plan]
   2. Run verification command
   3. Commit with message: [exact message]

   Working directory: [project-path]

   Follow the plan exactly. Do not skip steps.
   ```

2. Wait for subagent completion
3. Verify subagent succeeded
4. Continue to next task in plan

### Step 5: Complete Task

1. Call `mcp__task-manager__complete_task` with the task ID
2. Update TodoWrite to mark complete
3. Return to Step 1 (get next task)

## Subagent Configuration

**Always use these settings for execution subagents:**
- `subagent_type: general-purpose`
- `model: sonnet` (cost-effective for execution tasks)
- Provide complete context in the prompt (don't assume context is shared)

**Prompt template:**
```
Execute Task [N] from the plan at [path]

Your task is to:
1. [Step 1 from plan]
2. [Step 2 from plan]
3. [Verification step]
4. [Commit step with exact message]

Working directory: [absolute-path]

Follow the plan exactly. Do not skip steps.
```

## When to Stop

**Stop the workflow when:**
- `get_next_task` returns no tasks
- A subagent fails repeatedly (3 attempts)
- Plan creation encounters blocking questions
- User requests to stop

**On failure:**
1. Do NOT mark the task as complete
2. Report what failed and why
3. Ask user how to proceed

## Error Handling

**Subagent failure:**
1. Read the error output
2. If fixable: adjust prompt and retry (max 2 retries)
3. If not fixable: stop and report to user

**Plan creation blocked:**
1. Use AskUserQuestion to clarify requirements
2. Do not guess at implementation details
3. Update plan based on answers

## Example Session

```
User: Execute all pending tasks

Claude: I'm using the task-manager-workflow skill to execute all pending tasks.

[Calls get_next_task]
Found Task #1: Add README.md (priority: medium)

[Calls start_task with id: 1]
Task #1 is now in_progress.

I'm using the writing-plans skill to create the implementation plan for Task #1: Add README.md.

[Creates plan at docs/plans/2025-12-14-readme.md]

Executing plan with subagent...
[Launches Task tool with sonnet model]

Subagent completed successfully. README.md created and committed.

[Calls complete_task with id: 1]
Task #1 completed.

[Calls get_next_task]
Found Task #2: Add testing (priority: medium)

[... continues until no tasks remain ...]

All tasks completed. Summary:
- Task #1: Add README.md - DONE
- Task #2: Add testing - DONE
```

## Remember

- Always start tasks before working on them
- Always complete tasks after finishing
- Use writing-plans skill for every task (no shortcuts)
- Use sonnet model for subagents (cost-effective)
- Provide full context to subagents (they have no shared memory)
- Stop on failure, don't mark incomplete work as done
