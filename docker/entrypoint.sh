#!/usr/bin/env bash
set -euo pipefail

WORKDIR="${AGENT_WORKDIR:-/workspace}"
DAEMON_MODE="${AGENT_DAEMON:-0}"
OPENCODE_CONFIG_DIR="${OPENCODE_CONFIG_DIR:-/root/.config/opencode}"

mkdir -p "$WORKDIR"
cd "$WORKDIR"

if [ -d /opt/agent/spin ]; then
	mkdir -p "$OPENCODE_CONFIG_DIR/agents"

	# Copy AGENTS.md to agents directory with spin name
	SPIN_NAME="${AGENT_SPIN:-default}"
	if [ -f /opt/agent/spin/AGENTS.md ]; then
		cp /opt/agent/spin/AGENTS.md "$OPENCODE_CONFIG_DIR/agents/${SPIN_NAME}.md"
	elif [ -f /opt/agent/spin/AGENT.md ]; then
		cp /opt/agent/spin/AGENT.md "$OPENCODE_CONFIG_DIR/agents/${SPIN_NAME}.md"
	fi

	# Copy skills and MCP configs
	if [ -d /opt/agent/spin/skills ]; then
		cp -R /opt/agent/spin/skills "$OPENCODE_CONFIG_DIR/"
	fi
	if [ -d /opt/agent/spin/mcp ]; then
		cp -R /opt/agent/spin/mcp "$OPENCODE_CONFIG_DIR/"
	fi

	# Create opencode.json with spin agent configuration
	cat >"$OPENCODE_CONFIG_DIR/opencode.json" <<EOF
{
  "agent": {
    "${SPIN_NAME}": {
      "description": "Spin-specific agent: ${SPIN_NAME}",
      "mode": "primary",
      "prompt": "{file:$OPENCODE_CONFIG_DIR/agents/${SPIN_NAME}.md}"
    }
  },
  "default_agent": "${SPIN_NAME}"
}
EOF
fi

if [ "$#" -gt 0 ]; then
	exec "$@"
fi

if [ "$DAEMON_MODE" = "1" ]; then
	# Start tmux server and create session with OpenCode server
	SESSION_NAME="opencode-server"

	# Kill existing session if it exists
	tmux kill-session -t "$SESSION_NAME" 2>/dev/null || true

	# Start tmux session with OpenCode server
	# The server will use OPENCODE_SERVER_PASSWORD from environment
	tmux new-session -d -s "$SESSION_NAME" \
		"bunx opencode-ai serve --port 4096 --hostname 0.0.0.0; exec /bin/zsh"

	# Keep container running by monitoring the tmux session
	# If the session dies, the container will exit
	while tmux has-session -t "$SESSION_NAME" 2>/dev/null; do
		sleep 5
	done

	exit 0
fi

exec "${SHELL:-/bin/bash}"
