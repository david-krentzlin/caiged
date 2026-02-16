#!/usr/bin/env bash
set -euo pipefail

echo "Starting onboarding..."

if command -v gh >/dev/null 2>&1; then
	if ! gh auth status >/dev/null 2>&1; then
		gh auth login
	fi
fi

if command -v op >/dev/null 2>&1; then
	if ! op account list >/dev/null 2>&1; then
		op signin
	fi
fi

if command -v opencode >/dev/null 2>&1; then
	if [ -n "${OPENCODE_AUTH_CMD:-}" ]; then
		eval "$OPENCODE_AUTH_CMD"
	else
		echo "Set OPENCODE_AUTH_CMD to run the auth flow for opencode."
	fi
fi

echo "Onboarding complete."
