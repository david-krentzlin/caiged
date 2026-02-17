#!/usr/bin/env bash
set -euo pipefail

OPENCODE_VERSION="${OPENCODE_VERSION:-latest}"

if command -v bunx >/dev/null 2>&1; then
	exec bunx "opencode-ai@${OPENCODE_VERSION}"
fi

if command -v bun >/dev/null 2>&1; then
	exec bun x "opencode-ai@${OPENCODE_VERSION}"
fi

echo "bun is not available. Ensure bun is installed via config/target_mise.toml."
exec "${SHELL:-/bin/zsh}"
