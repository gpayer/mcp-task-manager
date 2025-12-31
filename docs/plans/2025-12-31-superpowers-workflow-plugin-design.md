# Superpowers Workflow Plugin Design

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a Claude Code plugin that integrates the task manager MCP with superpowers skills, using subtasks to store implementation plans.

**Architecture:** Parent tasks get decomposed into subtasks (one per implementation task). Execution uses superpowers:subagent-driven-development with task manager as the storage backend instead of plan files + TodoWrite.

**Tech Stack:** Claude Code plugin (markdown skills), task manager MCP tools

---

## Plugin Structure

```
mcp-task-manager/
├── .claude-plugin/
│   └── plugin.json              # Plugin manifest
├── skills/
│   └── superpowers-workflow/
│       └── SKILL.md             # Main workflow skill
├── commands/
│   └── execute-all.md           # Slash command
├── cmd/                         # (existing MCP server code)
├── internal/                    # (existing MCP server code)
└── ...
```

## Workflow Process

```
┌─────────────────────────────────────────────────────────────────┐
│                    superpowers-workflow                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. get_next_task ──► Parent task (no subtasks)                │
│          │                                                      │
│          ▼                                                      │
│  2. start_task(parent)                                         │
│          │                                                      │
│          ▼                                                      │
│  3. PLANNING PHASE (modified writing-plans behavior)           │
│     - Analyze parent task requirements                          │
│     - Create subtasks via create_task(parent_id=N)             │
│     - Each subtask = one implementation task with full spec    │
│          │                                                      │
│          ▼                                                      │
│  4. EXECUTION PHASE (per subtask)                              │
│     ┌─────────────────────────────────────────────────────────┐│
│     │ get_next_task ──► Returns subtask                       ││
│     │ start_task(subtask)                                     ││
│     │ Dispatch implementer subagent                           ││
│     │ Dispatch spec-reviewer subagent                         ││
│     │ Dispatch code-quality-reviewer subagent                 ││
│     │ complete_task(subtask)                                  ││
│     │         │                                               ││
│     │         ▼                                               ││
│     │ Loop until no subtasks remain                           ││
│     └─────────────────────────────────────────────────────────┘│
│          │                                                      │
│          ▼                                                      │
│  5. Parent auto-completes when last subtask done               │
│          │                                                      │
│          ▼                                                      │
│  6. Back to step 1 (next parent task)                          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Planning Phase

When the workflow encounters a parent task without subtasks, it enters planning mode.

Instead of writing to `docs/plans/`, create subtasks directly:

For each implementation task in the plan:
1. Call `create_task` with:
   - title: "Task N: [Component/Action]"
   - description: Full task spec in markdown (files, steps, code, verification)
   - priority: same as parent
   - type: same as parent
   - parent_id: the parent task ID

### Subtask Description Format

Each subtask description should contain:

```markdown
**Files:**
- Create: `exact/path/to/file.go`
- Modify: `exact/path/to/existing.go`
- Test: `tests/exact/path/to/test.go`

**Steps:**
1. Write failing test
2. Run test to verify failure
3. Implement minimal code
4. Run test to verify pass
5. Commit

**Code:** (inline if reasonable, reference file if large)

**Verification:** `go test ./...`

**Commit message:** `feat: add specific feature`
```

### Guidance for Plan Writer

- Design tasks to fit in subtask descriptions
- External files only for large examples/data (create in `docs/examples/` if needed)
- Each task should be completable by one subagent session
- Include everything the implementer needs - they have no other context

## Execution Phase (Superpowers Integration)

The execution phase invokes `superpowers:subagent-driven-development` with modified task sourcing.

### Key Differences from Standard subagent-driven-development

| Standard | Our Workflow |
|----------|--------------|
| Read plan file, extract tasks | Call `get_next_task` to get subtask |
| Track progress with TodoWrite | Track progress with task manager MCP |
| Manual task completion tracking | Call `complete_task` when done |

### Subagent Dispatch Pattern

For each subtask:

1. **Implementer subagent**
   - Receives: subtask description from `get_task`
   - Does: implement, test, commit, self-review
   - May ask clarifying questions

2. **Spec-reviewer subagent**
   - Receives: same subtask description + git diff
   - Verifies: implementation matches spec exactly
   - Loop until approved

3. **Code-quality-reviewer subagent**
   - Receives: same subtask description + git diff
   - Reviews: code quality, patterns, issues
   - Loop until approved

4. `complete_task(subtask_id)`

### What We Pass to Subagents

```
Task #{id}: {title}
Priority: {priority}
Parent: Task #{parent_id}: {parent_title}

{subtask description - full markdown spec}

---
Use superpowers:test-driven-development for implementation.
```

## Error Handling

### Subagent Failure

1. **Implementer fails or gets stuck**
   - Do NOT complete subtask
   - Report error to user
   - User can: retry, skip, or intervene manually

2. **Spec reviewer finds issues (loop)**
   - Implementer fixes
   - Re-review until approved
   - Max 3 iterations, then escalate to user

3. **Code reviewer finds issues (loop)**
   - Implementer fixes
   - Re-review until approved
   - Max 3 iterations, then escalate to user

### Planning Phase Failures

1. **Requirements unclear**
   - Use AskUserQuestion to clarify
   - Do NOT create subtasks until requirements are clear

2. **Task too large to decompose**
   - Ask user to break down the parent task first
   - Or suggest breakdown and get approval

### Task State Edge Cases

1. **get_next_task returns nothing**
   - All done, announce completion

2. **Parent has mix of done/todo subtasks (resuming)**
   - `get_next_task` returns next todo subtask
   - Continues where it left off

3. **User manually completes/modifies tasks**
   - Workflow respects current state
   - Picks up from `get_next_task` result

### Workflow Interruption

1. **User stops mid-execution**
   - Current subtask stays in_progress
   - Can resume later with `/execute-all`

2. **Session ends unexpectedly**
   - Same as above - task state persists in markdown files

## Implementation Tasks

### Task 1: Create plugin manifest

**Files:**
- Create: `.claude-plugin/plugin.json`

**Content:**
```json
{
  "name": "mcp-task-manager",
  "description": "Task manager MCP server with superpowers workflow integration",
  "version": "1.0.0",
  "repository": "https://github.com/gpayer/mcp-task-manager",
  "license": "MIT",
  "keywords": ["task-manager", "mcp", "workflow", "superpowers"]
}
```

### Task 2: Create superpowers-workflow skill

**Files:**
- Create: `skills/superpowers-workflow/SKILL.md`

**Content:** Full skill with planning and execution phases as designed above.

### Task 3: Create execute-all command

**Files:**
- Create: `commands/execute-all.md`

**Content:**
```markdown
---
description: Execute all pending tasks using superpowers workflow
---

Use the superpowers-workflow skill to execute all pending tasks
from the task manager MCP server.
```

### Task 4: Clean up old local skills

**Files:**
- Remove: `.claude/skills/task-manager-workflow/SKILL.md`
- Remove: `.claude/commands/task-manager:execute-all.md`

## Dependencies

- superpowers plugin must be installed
- Task manager MCP server must be configured and running

## Future Considerations

- `workflow` skill (non-superpowers version) - simpler execution without subagent reviews
- Additional commands like `/plan-task` (just planning) or `/execute-next` (single task)
