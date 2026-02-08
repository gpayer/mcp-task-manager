# Installing MCP Task Manager Plugin for Codex

Install the skill and command from this repository so Codex can discover and use them. The install is just a clone plus symlinks (or Windows junctions).

## Prerequisites
- Git

## Installation
1. **Clone the repository:**
   ```bash
   git clone https://github.com/gpayer/mcp-task-manager.git ~/.codex/mcp-task-manager
   ```

2. **Link the skill for Codex discovery:**
   ```bash
   mkdir -p ~/.agents/skills
   ln -s ~/.codex/mcp-task-manager/skills/superpowers-workflow ~/.agents/skills/mcp-task-manager
   ```

   **Windows (PowerShell):**
   ```powershell
   New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.agents\skills"
   cmd /c mklink /J "$env:USERPROFILE\.agents\skills\mcp-task-manager" "$env:USERPROFILE\.codex\mcp-task-manager\skills\superpowers-workflow"
   ```

3. **Link the command:**
   ```bash
   mkdir -p ~/.codex/commands
   ln -s ~/.codex/mcp-task-manager/commands/execute-all.md ~/.codex/commands/execute-all.md
   ```

   **Windows (PowerShell):**
   ```powershell
   New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.codex\commands"
   cmd /c mklink /J "$env:USERPROFILE\.codex\commands\execute-all.md" "$env:USERPROFILE\.codex\mcp-task-manager\commands\execute-all.md"
   ```

4. **Restart Codex** to pick up the new skill and command.

## Verify
```bash
ls -la ~/.agents/skills/mcp-task-manager
ls -la ~/.codex/commands/execute-all.md
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
rm ~/.codex/commands/execute-all.md
```
Optionally delete the clone:
```bash
rm -rf ~/.codex/mcp-task-manager
```
