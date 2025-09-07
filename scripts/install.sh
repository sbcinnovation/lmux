#!/usr/bin/env sh

set -eu

# Determine repository root relative to this script
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)

# Allow overriding target OS/ARCH via environment; default to current toolchain values
TARGET_GOOS=${GOOS:-$(go env GOOS)}
TARGET_GOARCH=${GOARCH:-$(go env GOARCH)}

# Determine install directory
# Priority: INSTALL_DIR > PREFIX/bin > OS default
if [ "${INSTALL_DIR:-}" != "" ]; then
  DEST_DIR="$INSTALL_DIR"
elif [ "${PREFIX:-}" != "" ]; then
  DEST_DIR="$PREFIX/bin"
else
  case "$TARGET_GOOS" in
    darwin|linux)
      DEST_DIR="/usr/local/bin"
      ;;
    *)
      # Fallback; user may adjust with INSTALL_DIR or PREFIX
      DEST_DIR="/usr/local/bin"
      ;;
  esac
fi

EXT=""
if [ "$TARGET_GOOS" = "windows" ]; then
  EXT=".exe"
fi

# Build into a temporary directory to avoid partial installs
TMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'lmux')
OUTPUT="$TMP_DIR/lmux$EXT"

echo "Building lmux for $TARGET_GOOS/$TARGET_GOARCH"

cd "$ROOT_DIR"
GOOS="$TARGET_GOOS" GOARCH="$TARGET_GOARCH" go build -o "$OUTPUT" ./cmd/lmux

# Ensure destination directory exists (use sudo if necessary)
if ! mkdir -p "$DEST_DIR" 2>/dev/null; then
  echo "Creating $DEST_DIR with sudo..."
  sudo mkdir -p "$DEST_DIR"
fi

DEST="$DEST_DIR/lmux$EXT"

# Prefer install(1) if available for proper perms; fall back to cp && chmod
if command -v install >/dev/null 2>&1; then
  if ! install -m 0755 "$OUTPUT" "$DEST" 2>/dev/null; then
    echo "Installing to $DEST with sudo..."
    sudo install -m 0755 "$OUTPUT" "$DEST"
  fi
else
  if ! cp "$OUTPUT" "$DEST" 2>/dev/null; then
    echo "Copying to $DEST with sudo..."
    sudo cp "$OUTPUT" "$DEST"
  fi
  if ! chmod 0755 "$DEST" 2>/dev/null; then
    sudo chmod 0755 "$DEST"
  fi
fi

echo "Installed: $DEST"


