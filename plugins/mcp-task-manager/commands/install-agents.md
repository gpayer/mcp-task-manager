---
description: Install mcp-task-manager role agents globally for Codex
---

Install the `planner`, `coder`, and `reviewer` role agents from the installed `mcp-task-manager` plugin into Codex's global agent discovery directory.

Run the installer from the newest installed plugin version:

```bash
plugin_root=$(ls -dt ~/.codex/plugins/cache/mcp-task-manager/mcp-task-manager/* | head -n 1)
bash "$plugin_root/scripts/install-codex-agents.sh"
```

After the installer finishes, restart Codex so the global agents from `~/.codex/agents/` are discovered in every session.
