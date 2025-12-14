#!/bin/sh
# .claude/hooks/format-go.sh
# Post-edit hook to format Go files with gofmt

file_path=$(jq -r '.tool_input.file_path' 2>/dev/null)

if [ -z "$file_path" ]; then
  exit 0  # No file path, skip
fi

# Only process Go files (not generated templ files)
if [[ "$file_path" != *.go ]]; then
  exit 0
fi

# Skip generated templ files - they are auto-generated
if [[ "$file_path" == *_templ.go ]]; then
  exit 0
fi

# Run gofmt
if command -v gofmt &> /dev/null; then
  gofmt -w "$file_path"
  exit 0
else
  echo "gofmt not found in PATH" >&2
  exit 1
fi
