#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

echo "==> Building frontend assets..."
cd "$ROOT_DIR/web"
npm run build

echo "==> Compiling Go binary..."
mkdir -p "$ROOT_DIR/bin"
cd "$ROOT_DIR"
go build -o "$ROOT_DIR/bin/quant-practice" ./cmd/quant-practice

echo "==> Build complete: bin/quant-practice"
