#!/usr/bin/env bash
set -e

REPO="rohanelukurthy/rig-rank"
BIN_NAME="rigrank"

echo "Retrieving latest version of RigRank..."

# Get OS and Architecture
OS=$(uname | tr '[:upper:]' '[:lower:]')
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
TAR_NAME="${BIN_NAME}_${OS^}_${ARCH}.tar.gz"
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

echo "Installing $BIN_NAME..."

# Try placing in /usr/local/bin, fallback to ~/bin or current dir
INSTALL_DIR="/usr/local/bin"

if [ -w "$INSTALL_DIR" ] || sudo -n true 2>/dev/null; then
    # We have write access (or can sudo)
    if [ ! -w "$INSTALL_DIR" ]; then
        echo "Requires sudo privileges to install to $INSTALL_DIR..."
        sudo mv "$TMP_DIR/$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"
    else
        mv "$TMP_DIR/$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"
    fi
    chmod +x "$INSTALL_DIR/$BIN_NAME"
    echo "Successfully installed to $INSTALL_DIR/$BIN_NAME"
else
    # Fallback to current directory
    mv "$TMP_DIR/$BIN_NAME" "./$BIN_NAME"
    chmod +x "./$BIN_NAME"
    echo "Could not write to $INSTALL_DIR. Binary has been placed in the current directory."
    echo "Ensure it is in your PATH, or run it with ./$BIN_NAME"
fi

# Cleanup
rm -rf "$TMP_DIR"

echo "RigRank installation complete! Run '$BIN_NAME --help' to get started."
