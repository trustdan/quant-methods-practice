#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
DIST_DIR="$ROOT_DIR/dist/release"
VERSION="0.1.0"

echo "=========================================================="
echo "  Quant Methods Practice — Release Package Verification   "
echo "=========================================================="

# 1. Detect platform
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
fi

PKG_NAME="quant-practice-v${VERSION}-${OS}-${ARCH}"
ARCHIVE_PATH="$DIST_DIR/${PKG_NAME}.tar.gz"

if [ ! -f "$ARCHIVE_PATH" ]; then
    echo "Error: Archive not found at $ARCHIVE_PATH"
    exit 1
fi

echo "Testing native release package: $(basename "$ARCHIVE_PATH")"

# 2. Checksum validation
echo "--> [1/6] Verifying SHA-256 checksum..."
EXPECTED_HASH=$(grep "${PKG_NAME}.tar.gz" "$DIST_DIR/SHA256SUMS.txt" | awk '{print $1}')
if command -v shasum >/dev/null 2>&1; then
    ACTUAL_HASH=$(shasum -a 256 "$ARCHIVE_PATH" | awk '{print $1}')
else
    ACTUAL_HASH=$(sha256sum "$ARCHIVE_PATH" | awk '{print $1}')
fi

if [ "$EXPECTED_HASH" != "$ACTUAL_HASH" ]; then
    echo "ERROR: Checksum mismatch! Expected $EXPECTED_HASH, got $ACTUAL_HASH"
    exit 1
fi
echo "    Checksum OK: $ACTUAL_HASH"

# 3. Isolated extraction
TEMP_TEST_DIR=$(mktemp -d /tmp/quant-verify-XXXXXX)
trap 'rm -rf "$TEMP_TEST_DIR"' EXIT

echo "--> [2/6] Extracting into isolated sandbox ($TEMP_TEST_DIR)..."
tar -xzf "$ARCHIVE_PATH" -C "$TEMP_TEST_DIR"
SANDBOX="$TEMP_TEST_DIR/$PKG_NAME"

# Check required files
for req in "quant-practice" "start.sh" "README.txt" "LICENSE" "NOTICE"; do
    if [ ! -f "$SANDBOX/$req" ]; then
        echo "ERROR: Missing required file in archive: $req"
        exit 1
    fi
done

# Check forbidden files (no sources, test files, sqlite databases, or env files)
if find "$SANDBOX" -type f \( -name "*.go" -o -name "*.ts" -o -name "*.db" -o -name "*.env" \) | grep -q .; then
    echo "ERROR: Archive contains source or private files!"
    exit 1
fi
echo "    Package structure and hygiene verified."

# 4. CLI offline executions in sandbox
echo "--> [3/6] Testing CLI commands in isolated sandbox..."
cd "$SANDBOX"

# Version check
./quant-practice -version | grep "0.1.0"

# Diagnostics check
echo "    Running offline doctor & diagnostics..."
DIAG_OUTPUT=$(./quant-practice -data-dir "$SANDBOX/data" -diagnostics)
echo "$DIAG_OUTPUT" | grep "ALL SYSTEMS HEALTHY"
echo "$DIAG_OUTPUT" | grep "21 approved question templates validated offline"
echo "$DIAG_OUTPUT" | grep "Active Bank Resolution" | grep "embedded assets"

# List bank check
echo "    Verifying 21 curriculum questions from embedded assets..."
BANK_OUTPUT=$(./quant-practice -list-bank)
echo "$BANK_OUTPUT" | grep "Approved Curriculum Bank (embedded assets): 21 questions"
echo "$BANK_OUTPUT" | grep "binomial_fair_coin_exactly_two"
echo "$BANK_OUTPUT" | grep "regression_sales_advertising"

# Math engine derivation check
echo "    Verifying analytical math engine offline..."
MATH_OUTPUT=$(./quant-practice -eval-binomial "n=4,p=0.5,k=2")
echo "$MATH_OUTPUT" | grep "0.375"

# 5. Backup and export tests
echo "--> [4/6] Testing database backup and progress export..."
./quant-practice -data-dir "$SANDBOX/data" -backup "$SANDBOX/backup.db"
if [ ! -s "$SANDBOX/backup.db" ]; then
    echo "ERROR: Backup file missing or empty"
    exit 1
fi

./quant-practice -data-dir "$SANDBOX/data" -export-data "$SANDBOX/export.json"
if [ ! -s "$SANDBOX/export.json" ]; then
    echo "ERROR: Export file missing or empty"
    exit 1
fi
grep "exported_at" "$SANDBOX/export.json" >/dev/null

# 6. Loopback HTTP Server & UI Verification
echo "--> [5/6] Testing loopback server and embedded UI..."
SERVER_LOG="$SANDBOX/server.log"
./quant-practice -data-dir "$SANDBOX/data" -port 0 -no-browser > "$SERVER_LOG" 2>&1 &
SERVER_PID=$!

# Wait for server to bind port
PORT=""
for i in {1..30}; do
    if grep -q "127.0.0.1:" "$SERVER_LOG" 2>/dev/null; then
        PORT=$(grep -oE "127.0.0.1:[0-9]+" "$SERVER_LOG" | head -n 1 | cut -d: -f2)
        break
    fi
    sleep 0.2
done

if [ -z "$PORT" ]; then
    echo "ERROR: Server failed to start on loopback port"
    cat "$SERVER_LOG"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi

echo "    Server listening on 127.0.0.1:$PORT (PID $SERVER_PID)"

# Verify root HTML serves embedded UI
HTML_RESP=$(curl -s "http://127.0.0.1:$PORT/")
echo "$HTML_RESP" | grep "<div id=\"root\"></div>" >/dev/null
echo "$HTML_RESP" | grep "Quant Methods Practice" >/dev/null
echo "$HTML_RESP" | grep "/vendor/mathjax/tex-svg.js" >/dev/null

# Verify public arcade scores API works offline
SCORES_RESP=$(curl -s "http://127.0.0.1:$PORT/api/arcade/scores")
echo "$SCORES_RESP" | grep '"scores"' >/dev/null

# Terminate server
kill -TERM $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

echo "--> [6/6] Clean teardown of temporary sandbox..."
echo ""
echo "=========================================================="
echo "  RELEASE VERIFICATION PASSED! ALL CHECKS SUCCESSFUL.    "
echo "=========================================================="
