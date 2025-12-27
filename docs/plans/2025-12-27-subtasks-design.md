# Subtasks Feature Design

## Overview

Add support for subtasks, allowing tasks to be broken down into smaller, manageable steps. Subtasks are first-class tasks with a parent reference, enabling hierarchical task organization while keeping the data model simple.

## Design Decisions

### Storage: Flat with Parent Reference

- All tasks stored as flat files: `001.md`, `002.md`, etc.
- Subtasks add `parent_id` field in YAML frontmatter
- Shared ID namespace - subtask IDs are auto-incremented like any task
- No nested directories or special naming conventions

### Nesting: Single Level Only

- Subtasks cannot have their own subtasks
- `parent_id` must reference a task with no parent
- Keeps complexity manageable; epics/milestones can be added later

## Data Model

```go
type Task struct {
    ID          int
    ParentID    *int      // nil for top-level tasks
    Title       string
    Description string
    Status      Status
    Priority    Priority  // Subtasks have own priority
    Type        string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

**Frontmatter example:**
```yaml
---
id: 7
parent_id: 4
title: "Write unit tests"
status: todo
priority: high
type: feature
created_at: 2025-12-27T10:00:00Z
updated_at: 2025-12-27T10:00:00Z
---
```

## Task Lifecycle

### Starting Tasks (`start_task`)

- Starting a subtask auto-starts its parent (if parent is `todo`)
- Starting a parent directly is allowed
- No blocking or special restrictions

### Completing Tasks (`complete_task`)

- Completing a parent with incomplete subtasks → **error**: "Cannot complete: has N incomplete subtasks"
- Completing the last incomplete subtask → **auto-completes parent**
- Completing a subtask with siblings incomplete → only completes that subtask

### Deleting Tasks (`delete_task`)

- Deleting a parent with subtasks → **error**: "Cannot delete: has N subtasks"
- Add `delete_subtasks: true` parameter to force cascade deletion
- Deleting a subtask is always allowed

## Query Behavior

### `list_tasks`

- Default: shows only top-level tasks (no `parent_id`)
- Each task shows subtask count: `3/5 done` or `-` if none
- Add `parent_id` filter to list subtasks of a specific task
- Existing filters (status, priority, type) still apply

### `get_task`

- Returns full task details
- Includes `subtasks` array with all child tasks (full details)
- For subtasks, includes `parent_id` in response

### `get_next_task`

- Finds highest priority `todo` task
- If a parent has subtasks, skip the parent - only its subtasks are candidates
- Parents without subtasks are normal candidates
- Priority ordering, then oldest first (existing logic)

## Creating Subtasks

### `create_task` Changes

- Add optional `parent_id` parameter
- Validates:
  - Parent task exists
  - Parent task has no parent itself (enforces single level)
- Returns error if validation fails

**MCP example:**
```
create_task(
  title: "Write unit tests",
  priority: "high",
  type: "feature",
  parent_id: 4
)
```

**CLI example:**
```bash
mcp-task-manager create --parent 4 --priority high "Write unit tests"
```

## Moving & Updating Subtasks

- `update_task` can change `parent_id` to move a subtask to a different parent
- Setting `parent_id` to null promotes subtask to top-level task
- Cannot set `parent_id` on a task that has subtasks (would create nesting)

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Create subtask with invalid parent | Error: "Parent task not found" |
| Create subtask under another subtask | Error: "Cannot create subtask under a subtask" |
| Complete parent with incomplete subtasks | Error: "Cannot complete: has N incomplete subtasks" |
| Delete parent with subtasks | Error: "Cannot delete: has N subtasks" (unless `delete_subtasks: true`) |
| Complete last subtask | Auto-completes parent, updates `updated_at` |
| Start subtask | Auto-starts parent if `todo`, updates `updated_at` |
| Add `parent_id` to task with subtasks | Error: "Cannot make a parent task into a subtask" |

## Summary

| Aspect | Decision |
|--------|----------|
| Storage | Flat files with `parent_id` in frontmatter |
| IDs | Shared namespace, auto-incremented |
| Nesting | Single level only |
| Create | `create_task` with optional `parent_id` |
| Start | Auto-start parent when subtask started |
| Complete | Block if subtasks incomplete; auto-complete parent when all done |
| Delete | Block if has subtasks; `delete_subtasks` param to force |
| List | Top-level by default, show subtask counts |
| Get | Include full subtasks array |
| Next | Skip parents with subtasks, only return subtasks |
| Priority | Subtasks have own priority |
| Move | Can change `parent_id` via update |
