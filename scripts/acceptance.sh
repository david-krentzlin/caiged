#!/usr/bin/env bash
set -euo pipefail

WORKDIR="${WORKDIR:-/workspace}"
MOUNT_DIR="${MOUNT_DIR:-$(pwd)}"
HELLO_IMAGE="${HELLO_IMAGE:-hello-world}"
SPIN="${SPIN:-qa}"
FORCE_BUILD="${FORCE_BUILD:-0}"

echo "Running nested container acceptance test"

FORCE_FLAG=""
if [ "$FORCE_BUILD" -eq 1 ]; then
	FORCE_FLAG="--force-build"
fi

./scripts/caiged "$MOUNT_DIR" \
	--task "$SPIN" \
	--enable-docker-sock \
	$FORCE_FLAG \
	-- bash -lc "ls \"${WORKDIR}\" >/dev/null && docker run --rm ${HELLO_IMAGE}"

echo "Acceptance test completed"
