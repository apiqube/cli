#!/bin/bash
set -e

REPO="apiqube/cli"
BINARY="qube"
INSTALL_DIR="/usr/local/bin"

echo "Resolving OC and architecture..."

OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [[ "$ARCH" == "x86_64" ]]; then
  ARCH="amd64"
elif [[ "$ARCH" == "aarch64" ]] || [[ "$ARCH" == "arm64" ]]; then
  ARCH="arm64"
else
  exit 1
fi

VERSION=$1
if [[ -z "$VERSION" ]]; then
  echo "Fetching latest GitHub version..."
  VERSION=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep -Po '"tag_name": "\K.*?(?=")')
fi

FILENAME="${BINARY}_${VERSION}_${OS}_${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

echo "Downloading $URL..."
curl -L -o $BINARY $URL
chmod +x $BINARY

echo "Installing in $INSTALL_DIR..."
sudo mv $BINARY $INSTALL_DIR/

echo "ApiQube CLI installed successfully!"
$BINARY version
