#!/usr/bin/env bash
set -euo pipefail

SPIN="${AGENT_SPIN:-unknown}"
WORKDIR="${AGENT_WORKDIR:-/workspace}"
CONFIG_DIR="${OPENCODE_CONFIG_DIR:-/root/.config/opencode}"
OPENCODE_AUTH_FILE="/root/.local/share/opencode/auth.json"

DOCKER_SOCK_STATUS="disabled"
if [ -S /var/run/docker.sock ]; then
	DOCKER_SOCK_STATUS="enabled"
fi

OPENCODE_AUTH_STATUS="not mounted"
if [ -f "$OPENCODE_AUTH_FILE" ]; then
	OPENCODE_AUTH_STATUS="mounted"
fi

cat <<EOF
Caiged environment

Spin: ${SPIN}
Workdir: ${WORKDIR}
Opencode config: ${CONFIG_DIR}
Docker socket: ${DOCKER_SOCK_STATUS}
OpenCode auth: ${OPENCODE_AUTH_STATUS}

Commands:
  ,help         Show this message

Notes:
  - AGENTS.md and skills are copied into ${CONFIG_DIR}
  - Network uses host mode by default unless disabled at launch
EOF
