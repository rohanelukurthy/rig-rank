#!/usr/bin/env bash
set -e

REPO="rohanelukurthy/rig-rank"
BIN_NAME="rigrank"

echo "Retrieving latest version of RigRank..."

# Get OS and Architecture
OS=$(uname | tr '[:upper:]' '[:lower:]')
OS_TITLE=$(uname)
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    ARCH="x86_64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
elif [ "$ARCH" = "i386" ]; then
    ARCH="i386"
else
    echo "Unsupported architecture: $ARCH"
    exit 1
fi

# Fetch the latest release tag from GitHub
LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep -o '"tag_name": ".*"' | cut -d'"' -f4)

if [ -z "$LATEST_TAG" ]; then
    echo "Error: Could not determine the latest release version."
    exit 1
fi

echo "Downloading $BIN_NAME $LATEST_TAG ($OS-$ARCH)..."

# Construct the download URL based on goreleaser name_template
TAR_NAME="rig-rank_${OS_TITLE}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/$TAR_NAME"

# Download to a temporary location
TMP_DIR=$(mktemp -d)
curl -sL "$DOWNLOAD_URL" -o "$TMP_DIR/rigrank.tar.gz"

if [ $? -ne 0 ] || [ ! -s "$TMP_DIR/rigrank.tar.gz" ]; then
    echo "Error: Download failed. Does the release artifact exist for this OS/Arch?"
    rm -rf "$TMP_DIR"
    exit 1
fi

echo "Extracting binary..."
tar -xzf "$TMP_DIR/rigrank.tar.gz" -C "$TMP_DIR"

if [ ! -f "$TMP_DIR/$BIN_NAME" ]; then
    echo "Error: Extraction failed, rigrank binary not found."
    rm -rf "$TMP_DIR"
    exit 1
fi

echo "Installing $BIN_NAME to current directory..."

# Move to current directory
mv "$TMP_DIR/$BIN_NAME" "./$BIN_NAME"
chmod +x "./$BIN_NAME"

# Cleanup
rm -rf "$TMP_DIR"

echo "RigRank installation complete!"
echo "The binary has been placed in your current directory."
echo "Run './$BIN_NAME --help' to get started."
