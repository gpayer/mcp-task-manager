# Installing MCP Task Manager Plugin for Codex

Install the skills and agents from this repository so Codex can discover and use them. The install is just a clone plus symlinks (or Windows junctions).

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

3. **Link the repo-local agents for Codex discovery:**
   ```bash
   mkdir -p ~/.agents/agents
   ln -s ~/.codex/mcp-task-manager/agents ~/.agents/agents/mcp-task-manager
   ```

   **Windows (PowerShell):**
   ```powershell
   New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.agents\agents"
   cmd /c mklink /J "$env:USERPROFILE\.agents\agents\mcp-task-manager" "$env:USERPROFILE\.codex\mcp-task-manager\agents"
   ```

4. **Restart Codex** to pick up the new skills and agents.

## Verify
```bash
ls -la ~/.agents/skills/mcp-task-manager
ls -la ~/.agents/agents/mcp-task-manager
```
You should see symlinks (or junctions on Windows) pointing into `~/.codex/mcp-task-manager`.

## Updating
```bash
cd ~/.codex/mcp-task-manager && git pull
```
Updates are immediate through the symlinks.

## Uninstalling
```bash
rm ~/.agents/skills/mcp-task-manager
rm ~/.agents/agents/mcp-task-manager
```
Optionally delete the clone:
```bash
rm -rf ~/.codex/mcp-task-manager
```
