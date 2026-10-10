#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

echo "==> Checking documentation and scaffold..."
python3 "$ROOT_DIR/scripts/validate_scaffold.py"

echo "==> Running frontend typecheck, lint and unit tests..."
cd "$ROOT_DIR/web"
npm run typecheck
npm run lint
npm run test

# Browser tests must drive a freshly built executable and embedded UI.
"$SCRIPT_DIR/build.sh"
npm run test:e2e

echo "==> Running Go formatting check, unit tests and vet..."
cd "$ROOT_DIR"
UNFORMATTED="$(gofmt -s -l cmd internal)"
if [ -n "$UNFORMATTED" ]; then
    echo "Go formatting check failed:"
    echo "$UNFORMATTED"
    exit 1
fi
go test -count=1 ./internal/... ./cmd/...
go vet ./internal/... ./cmd/...

echo "==> All verification checks PASSED!"
