# Lazy Index Creation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Prevent MCP server startup from creating `tasks/.index.json` when a repository has no active tasks; create the index only after the first task is written.

**Architecture:** Keep the fix at the storage/index initialization boundary. `Service.Initialize()` should still prepare an in-memory empty index, but `Index.Load()` and/or `Index.Rebuild()` must stop persisting `.index.json` when there are no active task markdown files to index. This preserves the existing write-on-create behavior in `Service.Create()`.

**Tech Stack:** Go, standard library filesystem APIs, existing storage/index tests in `internal/storage`, service tests in `internal/task`

---

### Task 1: Confirm and lock down the failing startup behavior with tests

**Files:**
- Modify: `internal/storage/storage_test.go`
- Modify: `internal/task/service_test.go`

- [ ] **Step 1: Add a storage-level test for missing index + no tasks**

Add a test that creates a temp tasks directory path with no `.md` files and no `.index.json`, constructs `MarkdownStorage` and `Index`, calls `idx.Load()`, and asserts:
- no error
- no `.index.json` file exists afterward
- `idx.All()` returns an empty slice

- [ ] **Step 2: Add a storage-level test for existing empty tasks directory**

Add a test that creates an empty tasks directory, calls `idx.Load()`, and asserts `.index.json` is still absent after load.

- [ ] **Step 3: Add a service initialization regression test**

Add a service test using real `MarkdownStorage` and `Index` with `cfg.ProjectFound = true` and a temp `tasks` path, call `svc.Initialize()`, and assert initialization does not create the tasks dir or `.index.json`.

- [ ] **Step 4: Run targeted tests and verify they fail first**

Run:
```bash
go test ./internal/storage ./internal/task -run 'TestIndex_Load|TestService_Initialize' -v
```

Expected: the new regression test covering no-task startup fails because `.index.json` is created today.

### Task 2: Change index initialization so empty projects stay fileless

**Files:**
- Modify: `internal/storage/index.go`

- [ ] **Step 1: Introduce an explicit empty-index initialization path**

Add a small helper in `internal/storage/index.go` that resets:
- `idx.entries`
- `idx.relationsBySource`
- `idx.relationsByTarget`

This avoids using `Save()` just to get an empty in-memory index.

- [ ] **Step 2: Make `Load()` distinguish “missing index with no tasks” from “missing index with tasks”**

Update `Load()` so that when `.index.json` is missing:
- if there are active task markdown files, rebuild and save as today
- if there are no active task markdown files, initialize empty in memory and return without writing `.index.json`

Use the current storage layer to determine whether any active tasks exist; do not infer this from directory existence alone.

- [ ] **Step 3: Make rebuild-on-corruption/staleness safe for empty projects**

If `Load()` falls back to `Rebuild()` because the index is corrupt, old-format, or stale, ensure the rebuild path does not write a new `.index.json` when the active task set is empty.

The simplest acceptable implementation is either:
- guard `Rebuild()` so it only calls `Save()` when at least one task was loaded, or
- let `Load()` bypass `Rebuild()` entirely when it already knows there are zero active tasks

Prefer the option with the smallest surface area and clearest semantics.

- [ ] **Step 4: Keep create/update/delete behavior unchanged**

Verify by inspection that:
- `Service.Create()` still calls `storage.EnsureDir()`
- first successful `Create()` still writes the markdown task and then `index.Save()`
- update/delete/archive flows still persist the index when active tasks exist

### Task 3: Verify the fix and guard adjacent regressions

**Files:**
- Modify: `internal/storage/storage_test.go` (if an additional regression test is needed)

- [ ] **Step 1: Run targeted regression tests**

Run:
```bash
go test ./internal/storage ./internal/task -run 'TestIndex_Load|TestService_Initialize' -v
```

Expected: all targeted tests pass.

- [ ] **Step 2: Run the broader package tests for touched areas**

Run:
```bash
go test ./internal/storage ./internal/task ./internal/cli -v
```

Expected: all tests pass; no behavior regressions in CLI/service flows.

- [ ] **Step 3: Manually verify first-write behavior**

Optional manual check:
1. Start in a temp repo containing only `mcp-tasks.yaml`
2. Initialize the service or start the MCP server
3. Confirm `tasks/.index.json` does not exist
4. Create the first task
5. Confirm both `tasks/001.md` and `tasks/.index.json` now exist

### Root Cause Summary

- `cmd/mcp-task-manager/main.go` always calls `svc.Initialize()` in MCP mode.
- `internal/task/service.go` unconditionally calls `s.index.Load()` during initialization.
- `internal/storage/index.go` currently treats a missing `.index.json` as “rebuild now”.
- `Rebuild()` loads zero tasks successfully for a missing or empty tasks directory, then still calls `Save()`.
- `Save()` creates `idx.dir` and writes `.index.json`, which violates the intended lazy-project initialization behavior.
