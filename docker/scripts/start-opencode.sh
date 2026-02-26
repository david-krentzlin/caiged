#!/usr/bin/env bash
set -euo pipefail

OPENCODE_VERSION="${OPENCODE_VERSION:-latest}"
GLOBAL_OPENCODE_BIN="/root/.bun/install/global/node_modules/opencode-ai/bin/opencode"
GLOBAL_OPENCODE_CACHED_BIN="/root/.bun/install/global/node_modules/opencode-ai/bin/.opencode"

if command -v bun >/dev/null 2>&1 && [ -f "$GLOBAL_OPENCODE_BIN" ]; then
	rm -f "$GLOBAL_OPENCODE_CACHED_BIN"
	exec bun "$GLOBAL_OPENCODE_BIN" "$@"
fi

if command -v bunx >/dev/null 2>&1; then
	exec bunx "opencode-ai@${OPENCODE_VERSION}" "$@"
fi

if command -v bun >/dev/null 2>&1; then
	exec bun x "opencode-ai@${OPENCODE_VERSION}" "$@"
fi

echo "bun is not available. Ensure bun is installed via config/target_mise.toml."
exec "${SHELL:-/bin/zsh}"
