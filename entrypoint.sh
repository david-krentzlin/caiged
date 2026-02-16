#!/usr/bin/env bash
set -euo pipefail

WORKDIR="${AGENT_WORKDIR:-/workspace}"
USE_TMUX="${AGENT_TMUX:-1}"
SESSION_NAME="${AGENT_TMUX_SESSION:-agent}"
OPENCODE_CONFIG_DIR="${OPENCODE_CONFIG_DIR:-/root/.config/opencode}"

mkdir -p "$WORKDIR"
cd "$WORKDIR"

if [ -d /opt/agent/spin ]; then
	mkdir -p "$OPENCODE_CONFIG_DIR"
	if [ -f /opt/agent/spin/AGENTS.md ]; then
		cp /opt/agent/spin/AGENTS.md "$OPENCODE_CONFIG_DIR/AGENTS.md"
	elif [ -f /opt/agent/spin/AGENT.md ] && [ ! -f "$OPENCODE_CONFIG_DIR/AGENTS.md" ]; then
		cp /opt/agent/spin/AGENT.md "$OPENCODE_CONFIG_DIR/AGENTS.md"
	fi
	if [ -d /opt/agent/spin/skills ] && [ ! -d "$OPENCODE_CONFIG_DIR/skills" ]; then
		cp -R /opt/agent/spin/skills "$OPENCODE_CONFIG_DIR/"
	fi
	if [ -d /opt/agent/spin/mcp ] && [ ! -d "$OPENCODE_CONFIG_DIR/mcp" ]; then
		cp -R /opt/agent/spin/mcp "$OPENCODE_CONFIG_DIR/"
	fi
fi

if [ "$#" -gt 0 ]; then
	exec "$@"
fi

if [ "$USE_TMUX" = "1" ] && command -v tmux >/dev/null 2>&1; then
	if tmux has-session -t "$SESSION_NAME" >/dev/null 2>&1; then
		exec tmux attach-session -t "$SESSION_NAME"
	fi

	tmux new-session -d -s "$SESSION_NAME" -n README
	README_CMD='clear; if [ -f /opt/agent/spin/README.md ]; then cat /opt/agent/spin/README.md; else echo "No README.md for this spin."; fi'
	tmux send-keys -t "$SESSION_NAME:README" "$README_CMD" C-m

	tmux new-window -t "$SESSION_NAME" -n opencode "/usr/local/bin/start-opencode"
	tmux select-window -t "$SESSION_NAME:README"
	exec tmux attach-session -t "$SESSION_NAME"
fi

exec "${SHELL:-/bin/bash}"
