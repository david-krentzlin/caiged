#!/usr/bin/env bash
set -euo pipefail

WORKDIR="${AGENT_WORKDIR:-/workspace}"
DAEMON_MODE="${AGENT_DAEMON:-0}"
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

if [ "$DAEMON_MODE" = "1" ]; then
	exec sleep infinity
fi

exec "${SHELL:-/bin/bash}"
