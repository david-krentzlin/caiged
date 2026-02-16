#!/usr/bin/env bash
set -euo pipefail

echo "Authenticating tools..."

GH_CONFIG_DIR="${GH_CONFIG_DIR:-/root/.config/gh}"

if command -v gh >/dev/null 2>&1; then
	if ! gh auth status >/dev/null 2>&1; then
		if [ -w "$GH_CONFIG_DIR" ]; then
			gh auth login
		else
			echo "gh config is read-only at ${GH_CONFIG_DIR}."
			echo "Re-run caiged with --mount-gh-rw or --no-mount-gh to login here."
		fi
	else
		echo "gh already authenticated"
	fi
else
	echo "gh is not installed"
fi

if command -v op >/dev/null 2>&1; then
	if ! op account list >/dev/null 2>&1; then
		op signin
	else
		echo "op already authenticated"
	fi
else
	echo "op is not installed"
fi

echo "Authentication complete."
