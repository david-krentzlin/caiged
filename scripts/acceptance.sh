#!/usr/bin/env bash
set -euo pipefail

WORKDIR="${WORKDIR:-/workspace}"
MOUNT_DIR="${MOUNT_DIR:-$(pwd)}"
HELLO_IMAGE="${HELLO_IMAGE:-hello-world}"
SPIN="${SPIN:-qa}"
FORCE_BUILD="${FORCE_BUILD:-0}"
CLI_BIN="${CLI_BIN:-./caiged/caiged}"

echo "Running nested container acceptance test"

if [ ! -x "$CLI_BIN" ]; then
	echo "Building caiged CLI"
	make -f caiged/Makefile build
fi

FORCE_FLAG=""
if [ "$FORCE_BUILD" -eq 1 ]; then
	FORCE_FLAG="--force-build"
fi

"$CLI_BIN" "$MOUNT_DIR" \
	--spin "$SPIN" \
	$FORCE_FLAG \
	-- bash -lc "ls \"${WORKDIR}\" >/dev/null && docker run --rm ${HELLO_IMAGE}"

echo "Acceptance test completed"
