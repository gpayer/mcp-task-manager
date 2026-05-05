#!/usr/bin/env bash
set -euo pipefail

plugin_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
agents_src="$plugin_root/agents"
codex_home="${CODEX_HOME:-$HOME/.codex}"
agents_dst="$codex_home/agents"

if [[ ! -d "$agents_src" ]]; then
  echo "Agent source directory not found: $agents_src" >&2
  exit 1
fi

mkdir -p "$agents_dst"

installed=0
for name in planner coder reviewer; do
  src="$agents_src/$name.toml"
  dst="$agents_dst/$name.toml"

  if [[ ! -f "$src" ]]; then
    echo "Missing agent definition: $src" >&2
    exit 1
  fi

  if [[ -e "$dst" && ! -L "$dst" ]]; then
    echo "Refusing to replace existing non-symlink: $dst" >&2
    echo "Move it aside or remove it, then rerun this installer." >&2
    exit 1
  fi

  ln -sfn "$src" "$dst"
  echo "Installed $name -> $src"
  installed=$((installed + 1))
done

echo "Installed $installed Codex role agents in $agents_dst"
echo "Restart Codex so new global agents are discovered."
