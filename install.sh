#!/bin/sh
set -e

# install.sh — install mzworker-go (macOS + Linux)
# Usage: curl -fsSL https://raw.githubusercontent.com/mnsrulz/mzworker-go/main/install.sh | bash
#        curl -fsSL https://raw.githubusercontent.com/mnsrulz/mzworker-go/main/install.sh | bash -s -- --version v0.3.0 --prefix /usr/local/bin
#        ./install.sh --version latest --prefix ~/.local/bin

REPO="mnsrulz/mzworker-go"
BIN="mzworker-go"
VERSION=""
PREFIX=""
# default prefix: /usr/local/bin if writable/root else ~/.local/bin

usage() {
  cat <<EOF
Usage: $0 [--version <tag>|latest] [--prefix <dir>]

Options:
  --version   GitHub tag (e.g., v0.3.0) or latest (default: latest)
  --prefix    Install directory (default: ~/.local/bin without sudo, /usr/local/bin if writable/root)
  -h, --help  Show this help

Examples:
  $0 --version v0.3.0
  $0 --version latest --prefix /usr/local/bin
  curl -fsSL https://raw.githubusercontent.com/$REPO/main/install.sh | bash
  curl -fsSL https://raw.githubusercontent.com/$REPO/main/install.sh | bash -s -- --version v0.3.0 --prefix /usr/local/bin

Supported: linux-amd64, linux-arm64, macos-arm64 (Apple Silicon). No darwin-amd64.
EOF
}

while [ $# -gt 0 ]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    --prefix) PREFIX="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage; exit 1 ;;
  esac
done

detect_os_arch() {
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  ARCH=$(uname -m)
  case "$OS" in
    linux) OS="linux" ;;
    darwin) OS="macos" ;;
    *) echo "Unsupported OS: $OS (supported: linux, darwin)" >&2; exit 1 ;;
  esac
  case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported arch: $ARCH (supported: amd64, arm64)" >&2; exit 1 ;;
  esac
  ABI="${OS}-${ARCH}"
  # reject darwin-amd64 (no release asset)
  if [ "$ABI" = "macos-amd64" ]; then
    echo "No release for macos-amd64 (Intel Mac). No darwin-amd64 asset; use Apple Silicon or build from source: go install github.com/$REPO@latest" >&2
    exit 1
  fi
}

resolve_version() {
  if [ -z "$VERSION" ] || [ "$VERSION" = "latest" ]; then
    # try latest including prereleases via API (follow redirect only gives stable)
    if command -v curl >/dev/null 2>&1; then
      # prefer most recent release including prerelease
      VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases?per_page=1" 2>/dev/null | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -1)
      if [ -z "$VERSION" ]; then
        REDIR=$(curl -fsSL -o /dev/null -w "%{url_effective}" "https://github.com/$REPO/releases/latest" 2>/dev/null || true)
        TAG=$(basename "$REDIR")
        if [ -n "$TAG" ] && [ "$TAG" != "latest" ]; then
          VERSION="$TAG"
        fi
      fi
      if [ -z "$VERSION" ] && command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -1)
      fi
    fi
    if [ -z "$VERSION" ] || [ "$VERSION" = "latest" ]; then
      echo "Failed to resolve latest version" >&2
      exit 1
    fi
  fi
  # normalize: ensure v prefix
  case "$VERSION" in
    v*) ;;
    *) VERSION="v$VERSION" ;;
  esac
}

download() {
  detect_os_arch
  resolve_version
  ASSET="${BIN}-${ABI}"
  LEGACY_ASSET="mzworker-${ABI}"
  URL="https://github.com/$REPO/releases/download/${VERSION}/${ASSET}"
  LEGACY_URL="https://github.com/$REPO/releases/download/${VERSION}/${LEGACY_ASSET}"
  TMP=$(mktemp -d 2>/dev/null || mktemp -d -t mzworker)
  DST="$TMP/$ASSET"
  echo "Downloading $URL..."
  downloaded=0
  if command -v curl >/dev/null 2>&1; then
    if curl -fsSL -o "$DST" "$URL" 2>/dev/null; then
      downloaded=1
    elif curl -fsSL -o "$DST" "$LEGACY_URL" 2>/dev/null; then
      echo "Fallback to legacy asset $LEGACY_ASSET"
      downloaded=1
    else
      echo "Download failed: $URL (tried $LEGACY_URL, check version exists: https://github.com/$REPO/releases/tag/$VERSION)" >&2
      exit 1
    fi
  elif command -v wget >/dev/null 2>&1; then
    if wget -qO "$DST" "$URL" 2>/dev/null; then
      downloaded=1
    elif wget -qO "$DST" "$LEGACY_URL" 2>/dev/null; then
      echo "Fallback to legacy asset $LEGACY_ASSET"
      downloaded=1
    else
      echo "Download failed: $URL" >&2; exit 1
    fi
  else
    echo "curl or wget required" >&2; exit 1
  fi
  chmod +x "$DST"
  # choose install dir: default ~/.local/bin without sudo, /usr/local/bin if prefix given or writable
  if [ -z "$PREFIX" ]; then
    if [ -w "/usr/local/bin" ] 2>/dev/null || [ "$(id -u 2>/dev/null || echo 1000)" = "0" ]; then
      INSTALL_DIR="/usr/local/bin"
    else
      INSTALL_DIR="$HOME/.local/bin"
      if [ "$INSTALL_DIR" = "/"".local/bin" ] || [ -z "$HOME" ]; then
        INSTALL_DIR="$HOME/.local/bin"
      fi
      echo "No write permission for /usr/local/bin, installing to $INSTALL_DIR (no sudo)..."
    fi
  else
    INSTALL_DIR="$PREFIX"
  fi
  mkdir -p "$INSTALL_DIR" 2>/dev/null || {
    if command -v sudo >/dev/null 2>&1; then
      sudo mkdir -p "$INSTALL_DIR"
    else
      echo "Cannot create $INSTALL_DIR" >&2; exit 1
    fi
  }
  if [ ! -w "$INSTALL_DIR" ] 2>/dev/null; then
    if command -v sudo >/dev/null 2>&1; then
      echo "Trying sudo for $INSTALL_DIR..."
      sudo mkdir -p "$INSTALL_DIR" 2>/dev/null || true
      if [ ! -w "$INSTALL_DIR" ]; then
        echo "No write permission for $INSTALL_DIR. Try: $0 --prefix ~/.local/bin" >&2
        exit 1
      fi
    else
      echo "No write permission for $INSTALL_DIR. Try: $0 --prefix ~/.local/bin or sudo $0" >&2
      exit 1
    fi
  fi
  echo "Installing to $INSTALL_DIR/$BIN..."
  if [ -w "$INSTALL_DIR" ] 2>/dev/null; then
    mv "$DST" "$INSTALL_DIR/$BIN"
  else
    sudo mv "$DST" "$INSTALL_DIR/$BIN"
  fi
  # compat symlink for old mzworker name
  if [ -w "$INSTALL_DIR" ] 2>/dev/null; then
    ln -sf "$INSTALL_DIR/$BIN" "$INSTALL_DIR/mzworker" 2>/dev/null || true
  else
    sudo ln -sf "$INSTALL_DIR/$BIN" "$INSTALL_DIR/mzworker" 2>/dev/null || true
  fi
  rm -rf "$TMP"
  echo "Installed $BIN $VERSION to $INSTALL_DIR/$BIN"
  if command -v "$INSTALL_DIR/$BIN" >/dev/null 2>&1; then
    "$INSTALL_DIR/$BIN" --version 2>&1 | head -1 || true
  fi
  # PATH hint
  case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *) echo "Note: $INSTALL_DIR not in PATH. Add: export PATH=\"\$PATH:$INSTALL_DIR\"" >&2 ;;
  esac
}

download
