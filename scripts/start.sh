#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
BIN_PATH="$ROOT_DIR/bin/quant-practice"

if [ ! -f "$BIN_PATH" ]; then
    echo "Binary not found at $BIN_PATH. Building first..."
    "$SCRIPT_DIR/build.sh"
fi

echo "Starting Quant Methods Practice..."
exec "$BIN_PATH" "$@"
