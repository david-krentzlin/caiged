#!/usr/bin/env bash
set -euo pipefail

if command -v opencode >/dev/null 2>&1; then
	exec opencode
fi

echo "opencode is not installed. Add it to config/target_mise.toml or install it manually."
exec "${SHELL:-/bin/zsh}"
