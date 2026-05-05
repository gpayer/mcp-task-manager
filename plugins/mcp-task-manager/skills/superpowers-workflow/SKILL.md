---
name: superpowers-workflow
description: Execute task manager tasks using installed planner, coder, and reviewer agents with explicit fallback approval
---

# Superpowers Workflow

## Overview

Execute tasks from the task manager MCP server using Codex custom role agents.

This packaged skill is self-contained for plugin installation:

- It expects a `task-manager` MCP server to be available from this plugin's `.mcp.json`
- It packages Codex custom agents named `planner`, `coder`, and `reviewer` under `agents/`
- Those agents should follow these role boundaries:
  - `planner`: planning-only, creates executable subtasks and does not implement code
  - `coder`: implementation-only, executes one assigned task and reports status clearly
  - `reviewer`: review-only, validates spec compliance and code quality independently

Do not assume repo-root-only helper files are available after installation.

- Parent tasks without subtasks dispatch `planner`
- Executable subtasks dispatch `coder`
- Every coding task dispatches `reviewer` after `coder` finishes
- Pure documentation or other low-complexity non-coding tasks dispatch `reviewer` only when the task explicitly calls for review work or the user asks for an independent review

This skill is the workflow controller. It owns task selection, task state changes, fallback decisions, phase transitions, and communication between role agents and the user when needed. Execute the workflow sequentially: after each `spawn_agent` call, wait for that subagent to finish before doing any other workflow step. Do not rely on background notifications to resume the workflow.

The workflow controller must not spawn subagents with a forked or cloned context, so `fork_context` must be `false`.

**Announce at start:** "Using superpowers-workflow to execute pending tasks."

## Agent Resolution Policy

For each role, resolve agents in this order:

1. The matching role-specific Codex agent, using the packaged plugin agent when available or a configured override from Codex's agent discovery locations
2. Another available role-appropriate agent
3. Default agent

Never silently downgrade.

If the preferred packaged custom agent cannot be used, stop and ask the user which fallback to allow before continuing. Make the downgrade explicit so the user understands the workflow is leaving the intended task-manager guardrails.

Apply this rule independently for:

- planning fallback for `planner`
- implementation fallback for `coder`
- review fallback for `reviewer`

## Model Selection Policy

Use role-level model guidance rather than hardcoded vendor-specific mappings.

- `planner`: prefer the most capable available reasoning model, usually with high reasoning effort
- `reviewer`: prefer the most capable available reasoning model, usually with high reasoning effort
- `coder`: prefer a quicker model for bounded mechanical work, but escalate to a stronger model for multi-file integration, ambiguous changes, or debugging-heavy tasks

If `coder` becomes blocked because the current model is too weak for the task, re-dispatch with a stronger model before escalating to the user, unless the blocker is missing context rather than model capability.

## Process

### Phase 1: Get Task

1. Call `get_next_task`
2. If no tasks are available:
   - Verify the worktree is clean
   - If only task-state changes remain, commit them with `git add tasks/ && git commit -m "chore: update task states"`
   - Announce "All tasks completed." and stop
3. If the result is a subtask: go to Phase 3
4. If the result is a parent task:
   - Call `get_task` to check for existing subtasks
   - If it has subtasks: go to Phase 3
   - If it has no subtasks: go to Phase 2

### Phase 2: Planning

Use this phase for parent tasks without subtasks.

1. Call `start_task` with the parent task ID before dispatch
2. Resolve `planner` using the policy above
3. If the preferred `planner` agent cannot be used:
   - do not continue automatically
   - ask the user which fallback to allow
   - only proceed after the user confirms the fallback choice
4. Dispatch `planner` with:

```text
Plan subtasks for Task #{id}: {title}
Priority: {priority}
Type: {type}

Description:
{parent task description}

Context:
- You are the planning phase only.
- Use `writing-plans` when converting approved intent into executable subtasks or plan structure.
- Keep scope aligned with the parent task.
- Do not implement code.
- If requirements are unclear, report `NEEDS_CONTEXT` with the missing information.

For each implementation task in your plan, call `create_task`:
- `title`: "Task N: [Component/Action]"
- `description`: Full task spec
- `priority`: {priority}
- `type`: {type}
- `parent_id`: {id}

Each subtask description must be self-contained and include:
- Files
- Steps
- Code guidance
- Verification command
- Commit message

Report the subtasks you created.
```

5. Immediately call `wait_agent` for that planner and do not take any other workflow step until it finishes
6. After `planner` returns:
   - If it reports `NEEDS_CONTEXT`, get clarification before proceeding
   - Verify subtasks were created
   - Proceed to Phase 3

### Phase 3: Execution

For each executable subtask, dispatch the role agent the task actually requires. Most implementation subtasks go to `coder`. Every coding task must then be reviewed by `reviewer`. Pure documentation or other low-complexity non-coding tasks use `reviewer` only when the task explicitly asks for review work or the user requests an independent review.

1. Call `start_task` with the subtask ID
2. Call `get_task` to get the full subtask details
3. Resolve `coder` using the policy above
4. If the preferred `coder` agent cannot be used:
   - stop before dispatch
   - ask the user which fallback to allow
   - continue only after the user confirms
5. Dispatch `coder` with:

```text
Task #{id}: {title}
Priority: {priority}
Parent: Task #{parent_id}: {parent_title}

{subtask description from task manager}

Instructions:
- You are the execution phase only.
- Use relevant execution skills such as `test-driven-development`, `systematic-debugging`, and `verification-before-completion` when applicable.
- Follow the task steps exactly as written.
- Report one of: `DONE`, `DONE_WITH_CONCERNS`, `NEEDS_CONTEXT`, `BLOCKED`.
- Include the commit message from the subtask spec, or propose a precise replacement if the implementation changed scope.
- Do not complete the task-manager task; the workflow controller handles task state.
- Do not create git commits; the workflow controller commits after the subtask is complete.
```

6. Immediately call `wait_agent` for that coder and do not take any other workflow step until it finishes
7. Handle `coder` results:
   - `DONE`: continue to review if required
   - `DONE_WITH_CONCERNS`: read the concerns, address any scope or correctness questions, then continue to review if required
   - `NEEDS_CONTEXT`: provide the missing context and re-dispatch
   - `BLOCKED`: stop and escalate with the blocker
8. If review is required, resolve `reviewer` using the same fallback policy and dispatch:

```text
Review the implementation for Task #{id}: {title}

Original spec:
{subtask description}

Review mode:
- follow the review scope requested by the task or user
- identify anything missing or extra
- return concrete findings with severity and file references
```

9. Immediately call `wait_agent` for that reviewer and do not take any other workflow step until it finishes
10. If review findings exist, send them back to `coder` and repeat the execution-review loop until the task is acceptable or blocked
11. When the subtask is complete, call `complete_task`
12. Parent tasks auto-complete when their last subtask is done
13. Review `git status` and stage all files changed for this completed subtask, including `tasks/`; do not stage unrelated pre-existing or user changes
14. Commit immediately using the commit message reported by `coder`
15. If there are no staged changes, do not create an empty commit; escalate because a completed subtask should normally leave task-state changes at minimum
16. Return to Phase 1

## Packaged Command

This plugin also exposes a packaged command description at `commands/execute-all.md`. Use it as the plugin-local entry for running this workflow after installation.
