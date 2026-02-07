#!/bin/bash
#
# Edictflow Agent Installer
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/kamilrybacki/edictflow/main/scripts/install-agent.sh | bash
#
# Or with a specific version:
#   curl -fsSL https://raw.githubusercontent.com/kamilrybacki/edictflow/main/scripts/install-agent.sh | bash -s -- v1.0.0
#

set -e

REPO="kamilrybacki/edictflow"
BINARY_NAME="edictflow"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            ;;
        *)
            error "Unsupported operating system: $OS"
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    if [ "$OS" = "windows" ]; then
        PLATFORM="${PLATFORM}.exe"
    fi

    info "Detected platform: $PLATFORM"
}

# Get the latest version from GitHub
get_latest_version() {
    info "Fetching latest version..."
    LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$LATEST" ]; then
        error "Failed to fetch latest version"
    fi
    echo "$LATEST"
}

# Download and install the binary
install() {
    VERSION="${1:-$(get_latest_version)}"
    info "Installing Edictflow Agent ${VERSION}..."

    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}-${PLATFORM}"
    TMP_FILE=$(mktemp)

    info "Downloading from: $DOWNLOAD_URL"
    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"; then
        error "Failed to download binary"
    fi

    chmod +x "$TMP_FILE"

    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        info "Requesting sudo to install to ${INSTALL_DIR}..."
        sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    success "Edictflow Agent ${VERSION} installed to ${INSTALL_DIR}/${BINARY_NAME}"
}

# Verify installation
verify() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        VERSION=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown")
        success "Verification passed: $VERSION"
    else
        warn "Binary not found in PATH. You may need to add ${INSTALL_DIR} to your PATH."
    fi
}

# Print usage instructions
print_usage() {
    echo ""
    echo -e "${GREEN}Installation complete!${NC}"
    echo ""
    echo "Quick Start:"
    echo "  1. Login to your Edictflow server:"
    echo "     ${BINARY_NAME} login --server https://your-server.com"
    echo ""
    echo "  2. Start the daemon:"
    echo "     ${BINARY_NAME} start"
    echo ""
    echo "  3. Check status:"
    echo "     ${BINARY_NAME} status"
    echo ""
    echo "For more information:"
    echo "  ${BINARY_NAME} --help"
    echo "  https://github.com/${REPO}"
    echo ""
}

# Main
main() {
    echo ""
    echo -e "${BLUE}Edictflow Agent Installer${NC}"
    echo "=========================="
    echo ""

    detect_platform
    install "$1"
    verify
    print_usage
}

main "$@"
