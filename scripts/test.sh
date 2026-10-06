#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

echo "==> Checking documentation and scaffold..."
python "$ROOT_DIR/scripts/validate_scaffold.py"

echo "==> Running frontend typecheck, units, e2e & build..."
cd "$ROOT_DIR/web"
npm run typecheck
npm run test
npm run test:e2e
npm run build

echo "==> Running Go unit tests and vet..."
cd "$ROOT_DIR"
go test -v ./internal/... ./cmd/...
go vet ./internal/... ./cmd/...

echo "==> All verification checks PASSED!"
