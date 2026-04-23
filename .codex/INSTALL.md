# Installing MCP Task Manager for Codex

Use the Codex marketplace flow first. The packaged plugin provides the `superpowers-workflow` skill, the `/execute-all` command, and a packaged `.mcp.json` entry for the `task-manager` MCP server. It does not install the `mcp-task-manager` binary for you, and it does not register personal `planner`, `coder`, or `reviewer` agents automatically.

## Prerequisites
- `mcp-task-manager` installed and available on `PATH`

Install the binary with Go:

```bash
go install github.com/gpayer/mcp-task-manager/cmd/mcp-task-manager@latest
```

Or install a release binary and ensure the executable is named `mcp-task-manager`.

## Installation
1. **Add this repository as a Codex marketplace**
   ```bash
   codex plugin marketplace add gpayer/mcp-task-manager
   ```

   If you are testing from a local checkout instead of GitHub, you can add the local marketplace root directory instead:
   ```bash
   codex plugin marketplace add /absolute/path/to/mcp-task-manager
   ```

2. **Install the plugin inside Codex**

   Open Codex and run:

   ```text
   /plugin install mcp-task-manager@mcp-task-manager
   ```

   This installs the plugin package located at `plugins/mcp-task-manager/` in this repository. That package includes `plugins/mcp-task-manager/.mcp.json`, so you do not need a separate `codex mcp add` step when the `mcp-task-manager` executable is already on `PATH`.

3. **Optional but recommended: register the role agents used by `superpowers-workflow`**

   The workflow is designed to use `planner`, `coder`, and `reviewer` custom agents. Those agent definitions are not auto-created by the plugin install. Create them in `~/.codex/agents/` for a personal setup or `.codex/agents/` for a project-scoped setup.

   If you have this repository checked out locally, you can copy the reference TOML files from `.codex/agents/` in the repo. Otherwise, create equivalent files manually.

4. **Register those agents in `~/.codex/config.toml` if your Codex setup requires explicit agent entries**

   ```toml
   [agents.planner]
   description = "Planning-only agent that decomposes approved work into executable subtasks without implementing code."
   config_file = "/home/<you>/.codex/agents/planner.toml"

   [agents.coder]
   description = "Implementation-only agent that executes one assigned task inside scope and reports status clearly."
   config_file = "/home/<you>/.codex/agents/coder.toml"

   [agents.reviewer]
   description = "Review-only agent that validates assigned work for spec compliance and code quality."
   config_file = "/home/<you>/.codex/agents/reviewer.toml"
   ```

   **Windows paths:**

   ```toml
   [agents.planner]
   description = "Planning-only agent that decomposes approved work into executable subtasks without implementing code."
   config_file = "C:\\Users\\<you>\\.codex\\agents\\planner.toml"

   [agents.coder]
   description = "Implementation-only agent that executes one assigned task inside scope and reports status clearly."
   config_file = "C:\\Users\\<you>\\.codex\\agents\\coder.toml"

   [agents.reviewer]
   description = "Review-only agent that validates assigned work for spec compliance and code quality."
   config_file = "C:\\Users\\<you>\\.codex\\agents\\reviewer.toml"
   ```

5. **Restart Codex**

## Verify
```bash
command -v mcp-task-manager
```

Inside Codex, confirm the plugin is installed and then run either:

```text
/execute-all
```

or:

```text
$superpowers-workflow
```

If you skipped the custom-agent setup, the workflow can still run, but it will stop and ask before falling back from the preferred `planner`, `coder`, or `reviewer` roles.

## Updating
1. Upgrade the marketplace checkout:
   ```bash
   codex plugin marketplace upgrade gpayer/mcp-task-manager
   ```
2. Reinstall or upgrade the plugin from inside Codex if your Codex version requires it.

## Uninstalling
1. Remove the plugin from Codex.
2. Remove any custom agent files you added under `~/.codex/agents/` or `.codex/agents/`.
3. Remove the marketplace entry if you no longer want this repository available:
   ```bash
   codex plugin marketplace remove gpayer/mcp-task-manager
   ```
