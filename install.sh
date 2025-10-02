#!/usr/bin/env bash
# Minepack installation script for Unix-like systems (Linux, macOS)
set -euo pipefail

# Colors for output
Color_Off=''
Red=''
Green=''
Yellow=''
Dim=''
Bold_White=''
Bold_Green=''

if [[ -t 1 ]]; then
    Color_Off='\033[0m'
    Red='\033[0;31m'
    Green='\033[0;32m'
    Yellow='\033[0;33m'
    Dim='\033[0;2m'
    Bold_Green='\033[1;32m'
    Bold_White='\033[1m'
fi

error() {
    echo -e "${Red}error${Color_Off}:" "$@" >&2
    exit 1
}

info() {
    echo -e "${Dim}$@ ${Color_Off}"
}

info_bold() {
    echo -e "${Bold_White}$@ ${Color_Off}"
}

success() {
    echo -e "${Green}$@ ${Color_Off}"
}

# Check for required tools
command -v curl >/dev/null 2>&1 || error 'curl is required to install minepack'
command -v tar >/dev/null 2>&1 || command -v unzip >/dev/null 2>&1 || error 'tar or unzip is required to install minepack'

# Detect platform
platform=$(uname -ms)

case $platform in
'Darwin x86_64' | 'Darwin x64')
    target='darwin-amd64'
    ;;
'Darwin arm64' | 'Darwin aarch64')
    target='darwin-arm64'
    ;;
'Linux aarch64' | 'Linux arm64')
    target='linux-arm64'
    ;;
'Linux x86_64' | 'Linux x64' | 'Linux AMD64' | *)
    target='linux-amd64'
    ;;
esac

if [[ $target == darwin-amd64 ]]; then
    # Check if running in Rosetta
    if [[ $(sysctl -n sysctl.proc_translated 2>/dev/null) == 1 ]]; then
        target='darwin-arm64'
        info "Running in Rosetta 2. Downloading minepack for $target instead"
    fi
fi

# Determine installation directory
# Prefer ~/.local/bin if it's in PATH, otherwise use /usr/local/bin
install_dir="$HOME/.local/bin"
if [[ ! -d "$install_dir" ]]; then
    mkdir -p "$install_dir"
    info "Created directory: $install_dir"
fi

# Check if install_dir is in PATH
if [[ ":$PATH:" != *":$install_dir:"* ]]; then
    info "Note: $install_dir is not in your PATH"
    info "Add it to your PATH by running:"
    info_bold "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
    info "  (or ~/.zshrc, ~/.profile, etc. depending on your shell)"
fi

# Get latest release info from GitHub
info "Fetching latest release information..."
release_url="https://api.github.com/repos/ayeuhugyu/minepack/releases/latest"
release_data=$(curl -sL "$release_url")

# Extract tag name and download URL
tag_name=$(echo "$release_data" | grep '"tag_name":' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
if [[ -z "$tag_name" ]]; then
    error "Failed to fetch latest release information. Please check your internet connection."
fi

# Construct binary name and download URL
binary_name="minepack-${target}"
download_url=$(echo "$release_data" | grep '"browser_download_url":' | grep "$binary_name" | head -1 | sed -E 's/.*"browser_download_url": "([^"]+)".*/\1/')

if [[ -z "$download_url" ]]; then
    error "No binary found for platform $target"
fi

info "Installing minepack $tag_name for $target..."
info "Download URL: $download_url"

# Create temporary directory
tmp_dir=$(mktemp -d)
trap "rm -rf $tmp_dir" EXIT

# Download binary
info "Downloading..."
tmp_file="$tmp_dir/minepack"
if ! curl -fsSL "$download_url" -o "$tmp_file"; then
    error "Failed to download minepack"
fi

# Make executable
chmod +x "$tmp_file"

# Move to installation directory
install_path="$install_dir/minepack"
if [[ -f "$install_path" ]]; then
    info "Replacing existing installation..."
    # Create backup
    backup_path="${install_path}.backup"
    mv "$install_path" "$backup_path"
    info "Created backup at $backup_path"
fi

mv "$tmp_file" "$install_path"

# Verify installation
if ! "$install_path" --version >/dev/null 2>&1; then
    error "Installation verification failed. The binary may be corrupted."
fi

installed_version=$("$install_path" --version | head -1)

success "Successfully installed $installed_version!"
success "Binary location: $install_path"

# Remind about PATH if needed
if [[ ":$PATH:" != *":$install_dir:"* ]]; then
    echo ""
    info_bold "⚠ Remember to add $install_dir to your PATH to use minepack from anywhere!"
fi

echo ""
info "Run 'minepack --help' to get started"
