#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"
VERSION="0.1.0"
DIST_DIR="$ROOT_DIR/dist/release"
STAGING_BASE="$ROOT_DIR/dist/staging"

echo "=========================================================="
echo "  Quant Methods Practice — Multi-Platform Release Packager"
echo "  Version: $VERSION"
echo "=========================================================="

# 1. Build frontend production assets
echo ""
echo "==> [1/4] Building production frontend assets..."
cd "$ROOT_DIR/web"
npm run build

# 2. Prepare release and staging directories
echo ""
echo "==> [2/4] Initializing release workspaces..."
rm -rf "$DIST_DIR" "$STAGING_BASE"
mkdir -p "$DIST_DIR" "$STAGING_BASE"

# Platform target definitions: OS, ARCH, BIN_NAME, LAUNCHER, ARCHIVE_EXT
TARGETS=(
    "darwin:arm64:quant-practice:start.sh:tar.gz"
    "darwin:amd64:quant-practice:start.sh:tar.gz"
    "linux:amd64:quant-practice:start.sh:tar.gz"
    "linux:arm64:quant-practice:start.sh:tar.gz"
    "windows:amd64:quant-practice.exe:start.bat:zip"
)

# 3. Compile and assemble archives
echo ""
echo "==> [3/4] Cross-compiling standalone native binaries and assembling packages..."
cd "$ROOT_DIR"

for TARGET in "${TARGETS[@]}"; do
    IFS=":" read -r GOOS GOARCH BIN_NAME LAUNCHER ARCHIVE_EXT <<< "$TARGET"
    PKG_NAME="quant-practice-v${VERSION}-${GOOS}-${GOARCH}"
    PKG_DIR="$STAGING_BASE/$PKG_NAME"
    mkdir -p "$PKG_DIR"

    echo "  --> Compiling $GOOS/$GOARCH..."
    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build \
        -ldflags="-s -w" \
        -o "$PKG_DIR/$BIN_NAME" \
        ./cmd/quant-practice

    # Copy launcher and documentation
    if [ "$LAUNCHER" = "start.sh" ]; then
        cp "$ROOT_DIR/scripts/release-launchers/start.sh" "$PKG_DIR/start.sh"
        chmod +x "$PKG_DIR/start.sh" "$PKG_DIR/$BIN_NAME"
    elif [ "$LAUNCHER" = "start.bat" ]; then
        cp "$ROOT_DIR/scripts/release-launchers/start.bat" "$PKG_DIR/start.bat"
    fi

    cp "$ROOT_DIR/scripts/release-launchers/README.txt" "$PKG_DIR/README.txt"
    cp "$ROOT_DIR/LICENSE" "$PKG_DIR/LICENSE"
    cp "$ROOT_DIR/NOTICE" "$PKG_DIR/NOTICE"

    # Security check: verify no database, credentials, or private files leaked into staging
    if find "$PKG_DIR" -type f \( -name "*.db" -o -name "*.env" -o -name "*secret*" -o -name "*.key" \) | grep -q .; then
        echo "SECURITY ERROR: Sensitive files detected in package $PKG_NAME!"
        exit 1
    fi

    # Create archive
    ARCHIVE_FILE="$DIST_DIR/${PKG_NAME}.${ARCHIVE_EXT}"
    echo "  --> Packaging $ARCHIVE_FILE..."
    if [ "$ARCHIVE_EXT" = "tar.gz" ]; then
        tar -czf "$ARCHIVE_FILE" -C "$STAGING_BASE" "$PKG_NAME"
    elif [ "$ARCHIVE_EXT" = "zip" ]; then
        (cd "$STAGING_BASE" && zip -q -r "$ARCHIVE_FILE" "$PKG_NAME")
    fi
done

# 4. Generate SHA-256 Checksums
echo ""
echo "==> [4/4] Generating SHA-256 Checksums..."
cd "$DIST_DIR"
if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 quant-practice-* > SHA256SUMS.txt
elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum quant-practice-* > SHA256SUMS.txt
fi

echo ""
echo "=========================================================="
echo "  Release Packaging Complete!"
echo "  Artifacts saved in: $DIST_DIR"
echo "=========================================================="
cat SHA256SUMS.txt
echo ""
