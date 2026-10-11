#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_PATH="$SCRIPT_DIR/quant-practice"

if [ ! -f "$BIN_PATH" ]; then
    echo "Error: quant-practice binary not found at $BIN_PATH"
    exit 1
fi

chmod +x "$BIN_PATH" 2>/dev/null || true
echo "Launching Quant Methods Practice..."
exec "$BIN_PATH" "$@"
