# Task Relations Design

## Motivation

Tasks can be created in a different order than the optimal execution order. Relations — primarily `blocked_by` — let agents know which tasks are actionable and which must wait. Secondary relation types (`relates_to`, `duplicate_of`) add informational metadata.

## Relation Types

| Type | Semantic | Behavioral effect | Symmetric |
|------|----------|-------------------|-----------|
| `blocked_by` | Source can't proceed until target is done | Affects `get_next_task`, `start_task`, `get_task`, `list_tasks` | No |
| `relates_to` | Informational link | None | Yes |
| `duplicate_of` | Source is a duplicate of target | None | No |

Relation types are configurable via `mcp-tasks.yaml`. Behavioral effects are hardcoded to specific type names (`blocked_by`), not to the config list.

## Data Model

### Task Frontmatter

The `relations` field is optional. Omitted when empty.

```yaml
---
id: 5
title: "Implement auth"
status: todo
priority: high
type: feature
relations:
  - type: blocked_by
    task: 3
  - type: relates_to
    task: 7
---
```

**Storage rules:**
- `blocked_by`: stored only on the blocked task (the source). The blocker gets no frontmatter entry.
- `relates_to`: stored on one side only. The index generates the reverse edge.
- `duplicate_of`: stored only on the duplicate (the source).
- Validation: target task must exist, no self-references.

### Index File

`.index.json` gains a top-level `relations` array with the canonical edge format:

```json
{
  "tasks": [...],
  "relations": [
    { "type": "blocked_by", "source": 5, "target": 3 },
    { "type": "relates_to", "source": 5, "target": 7 },
    { "type": "relates_to", "source": 7, "target": 5 }
  ]
}
```

- Built from scanning all task frontmatter on startup
- Symmetric types generate two edges automatically
- `blocked_by` generates one edge (source = blocked task, target = blocker)

### In-Memory Index

```go
type RelationEdge struct {
    Type   string `json:"type"`
    Source int    `json:"source"`
    Target int    `json:"target"`
}
```

Two lookup maps for fast queries:
- `relationsBySource map[int][]RelationEdge` — "what relations does task X have?"
- `relationsByTarget map[int][]RelationEdge` — "what tasks point at task X?"

Both populated on index rebuild and kept in sync on add/remove.

### Index Methods

- `AddRelation(edge RelationEdge)` — adds to both maps and persists
- `RemoveRelation(edge RelationEdge)` — removes from both maps and persists
- `GetRelationsForTask(taskID int) []RelationEdge` — all edges where task is source OR target
- `GetBlockers(taskID int) []int` — target IDs from `blocked_by` edges where source == taskID
- `RemoveAllRelationsForTask(taskID int) []RelationEdge` — for delete cascade; returns removed edges so service knows which other task files to update

### Configuration

`mcp-tasks.yaml`:

```yaml
task_types:
  - feature
  - bug
relation_types:
  - blocked_by
  - relates_to
  - duplicate_of
```

`relation_types` is optional — defaults to the three above if omitted.

## MCP Tools

### New Tools

**`add_relation(source, type, target)`**
- Adds a relation to the source task's frontmatter
- Validates: both tasks exist, type is in configured list, no self-reference, no duplicate relation
- Returns the created edge

**`remove_relation(source, type, target)`**
- Removes a relation from the source task's frontmatter
- Returns error if relation doesn't exist

### Modified Tools

**`get_next_task`** — blocking-awareness:
- Existing subtask prioritization runs first (prefer subtasks of in-progress parents)
- Then filters out tasks where any `blocked_by` target is not `done`

**`start_task`** — guard:
- If task has unresolved `blocked_by` relations, returns error: "Task 5 is blocked by tasks: 3 (todo), 8 (in_progress)"

**`get_task`** — enriched response:
- Includes the task's relations list
- Includes a derived `blocked: true/false` field
- If blocked, includes which tasks are blocking and their status

**`list_tasks`** — enriched response:
- Each task gains a `blocked: true/false` derived field

**`delete_task`** — cleanup:
- When deleting a task, remove all relations where it appears as source or target
- Updates other tasks' frontmatter as needed

## Service Layer

### New Methods

**`AddRelation(source, relationType, target)`:**
1. Validate both tasks exist
2. Validate relation type is configured
3. Validate no self-reference
4. Validate no duplicate relation
5. Load source task, append to relations, save
6. Update index (for symmetric types, add reverse edge too)

**`RemoveRelation(source, relationType, target)`:**
1. Validate relation exists in source task's frontmatter
2. Remove it, save
3. Update index (for symmetric types, remove reverse edge too)

**`IsBlocked(taskID) (bool, []BlockingInfo)`:**
- Queries index for `blocked_by` edges where source == taskID
- Checks each target task's status
- Returns true + list of unresolved blockers if any target is not `done`

### Cascade on Delete

- Scan all tasks for relations referencing the deleted ID
- Remove those entries from their frontmatter and save
- Rebuild affected index entries

### Interaction with Subtask Logic

- Relations and subtasks are orthogonal — a subtask can be blocked by a task outside its parent
- `get_next_task` applies both filters: skip parents with incomplete subtasks AND skip blocked tasks

## CLI

- `list` command shows a `[BLOCKED]` indicator next to blocked tasks
- `get` command shows relations in the detail view

## Out of Scope

- **Cycle detection** — circular `blocked_by` chains make both tasks stuck. User fixes manually.
- **Transitive blocking** — A blocked_by B and B blocked_by C: only direct blockers checked for A. B being blocked doesn't cascade.
- **Auto-close duplicates** — `duplicate_of` is purely informational.
- **Relation metadata** — no comments, timestamps, or other fields on relations.
- **Multi-level subtask interaction** — relations don't change the single-level nesting constraint.
