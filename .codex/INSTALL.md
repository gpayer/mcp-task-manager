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

4. **Restart Codex** to pick up the new skills and agents.

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
