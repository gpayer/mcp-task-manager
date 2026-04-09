# Installing MCP Task Manager Plugin for Codex

Install the skills and custom agents from this repository so Codex can discover and use them. The install is a clone plus symlinks (or Windows junctions).

## Prerequisites
- Git

## Installation
1. **Clone the repository:**
   ```bash
   git clone https://github.com/gpayer/mcp-task-manager.git ~/.codex/mcp-task-manager
   ```

2. **Link the skills for Codex discovery:**
   ```bash
   mkdir -p ~/.agents/skills
   ln -s ~/.codex/mcp-task-manager/skills ~/.agents/skills/mcp-task-manager
   ```

   **Windows (PowerShell):**
   ```powershell
   New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.agents\skills"
   cmd /c mklink /J "$env:USERPROFILE\.agents\skills\mcp-task-manager" "$env:USERPROFILE\.codex\mcp-task-manager\skills"
   ```

3. **Link the custom agent files into Codex's documented agent directory:**
   ```bash
   mkdir -p ~/.codex/agents
   ln -s ~/.codex/mcp-task-manager/.codex/agents/planner.toml ~/.codex/agents/planner.toml
   ln -s ~/.codex/mcp-task-manager/.codex/agents/coder.toml ~/.codex/agents/coder.toml
   ln -s ~/.codex/mcp-task-manager/.codex/agents/reviewer.toml ~/.codex/agents/reviewer.toml
   ```

   **Windows (PowerShell):**
   ```powershell
   New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.codex\agents"
   New-Item -ItemType SymbolicLink -Path "$env:USERPROFILE\.codex\agents\planner.toml" -Target "$env:USERPROFILE\.codex\mcp-task-manager\.codex\agents\planner.toml"
   New-Item -ItemType SymbolicLink -Path "$env:USERPROFILE\.codex\agents\coder.toml" -Target "$env:USERPROFILE\.codex\mcp-task-manager\.codex\agents\coder.toml"
   New-Item -ItemType SymbolicLink -Path "$env:USERPROFILE\.codex\agents\reviewer.toml" -Target "$env:USERPROFILE\.codex\mcp-task-manager\.codex\agents\reviewer.toml"
   ```

   Codex's subagent docs expect standalone TOML files under `~/.codex/agents/` for personal agents or `.codex/agents/` for project-scoped agents.

4. **Register the new agents in Codex's config**

    Add this snippet to `~/.codex/config.toml`:

    ```TOML
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

    Replace `/home/<you>` with your actual home directory path.

    **Windows config snippet:**

    ```TOML
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

    The symlinks in Step 3 provide stable files for Codex to read, and this config registration is still required for personal agents to be found.

5. **Restart Codex** to pick up the new skills and agents.

## Verify
```bash
ls -la ~/.agents/skills/mcp-task-manager
ls -la ~/.codex/agents/planner.toml
ls -la ~/.codex/agents/coder.toml
ls -la ~/.codex/agents/reviewer.toml
```
You should see symlinks (or links on Windows) pointing into `~/.codex/mcp-task-manager/.codex/agents/`.

## Updating
```bash
cd ~/.codex/mcp-task-manager && git pull
```
Updates are immediate through the symlinks.

## Uninstalling
```bash
rm ~/.agents/skills/mcp-task-manager
rm ~/.codex/agents/planner.toml
rm ~/.codex/agents/coder.toml
rm ~/.codex/agents/reviewer.toml
```
Optionally delete the clone:
```bash
rm -rf ~/.codex/mcp-task-manager
```
