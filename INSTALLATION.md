# Installation Scripts Documentation

This directory contains automated installation scripts for Minepack that make it easy for users to install the tool without manual setup.

## Files

- **install.sh** - Installation script for Unix-like systems (Linux, macOS)
- **install.ps1** - Installation script for Windows (PowerShell)

## Usage

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/ayeuhugyu/minepack/main/install.sh | bash
```

Or with wget:

```bash
wget -qO- https://raw.githubusercontent.com/ayeuhugyu/minepack/main/install.sh | bash
```

### Windows

Open PowerShell and run:

```powershell
irm https://raw.githubusercontent.com/ayeuhugyu/minepack/main/install.ps1 | iex
```

## Features

### install.sh (Unix/Linux/macOS)

- **Auto-detection**: Automatically detects the operating system (Linux, macOS) and architecture (amd64, arm64)
- **Rosetta detection**: On macOS, detects if running in Rosetta 2 and downloads the correct binary
- **User installation**: Installs to `~/.local/bin` (user directory, no sudo required)
- **PATH guidance**: Provides instructions if the install directory is not in PATH
- **Backup**: Creates a backup of existing installations
- **Verification**: Verifies the binary works after installation
- **Error handling**: Clear error messages with colored output

**Supported platforms:**
- Linux x86_64 (amd64)
- Linux ARM64 (aarch64)
- macOS x86_64 (Intel)
- macOS ARM64 (Apple Silicon)

### install.ps1 (Windows)

- **Auto-detection**: Automatically detects Windows architecture (amd64, arm64)
- **User installation**: Installs to `%USERPROFILE%\.minepack\bin` (no admin required)
- **PATH management**: Automatically adds the installation directory to user PATH
- **Backup**: Creates a backup of existing installations
- **Verification**: Verifies the binary works after installation
- **Environment broadcasting**: Notifies Windows of PATH changes
- **Error handling**: Clear error messages

**Supported platforms:**
- Windows x86_64 (amd64)
- Windows ARM64

**Optional flags:**
- `-NoPathUpdate`: Skip adding to PATH

## How It Works

Both scripts follow a similar process:

1. **Platform Detection**: Determine the OS and architecture
2. **Fetch Latest Release**: Query GitHub API for the latest release
3. **Download Binary**: Download the appropriate binary for the platform
4. **Install**: Move the binary to the installation directory
5. **Verify**: Run `--version` to ensure the binary works
6. **PATH Setup**: Add to PATH or provide instructions

## Customization

### Changing Installation Directory (Unix)

Edit the `install_dir` variable in `install.sh`:

```bash
install_dir="$HOME/.local/bin"  # Change this
```

### Changing Installation Directory (Windows)

Edit the `InstallDir` variable in `install.ps1`:

```powershell
$InstallDir = "${env:USERPROFILE}\.minepack\bin"  # Change this
```

## Testing

### Local Testing (Unix)

```bash
# Make the script executable
chmod +x install.sh

# Run locally
./install.sh
```

### Local Testing (Windows)

```powershell
# Run locally
.\install.ps1
```

## Troubleshooting

### Unix/Linux/macOS

**Problem**: `curl: command not found`
- **Solution**: Install curl: `sudo apt-get install curl` (Ubuntu/Debian) or `brew install curl` (macOS)

**Problem**: Binary not found after installation
- **Solution**: Add `~/.local/bin` to your PATH:
  ```bash
  echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
  source ~/.bashrc
  ```

**Problem**: Permission denied
- **Solution**: The script should not require sudo. Check that `~/.local/bin` is writable.

### Windows

**Problem**: Script execution blocked by policy
- **Solution**: Run PowerShell as administrator and execute:
  ```powershell
  Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
  ```

**Problem**: Binary not found after installation
- **Solution**: Restart your PowerShell/terminal. If still not working, check PATH:
  ```powershell
  $env:Path
  ```

**Problem**: "Not supported on 32-bit Windows"
- **Solution**: Minepack only supports 64-bit Windows. Upgrade to 64-bit Windows.

## Security

### Script Security

- Both scripts download binaries from official GitHub releases only
- The scripts verify binary execution after installation
- Scripts create backups before replacing existing installations
- No sensitive information is collected or transmitted

### Verifying Scripts

Before running, you can review the scripts:

- Unix: https://raw.githubusercontent.com/ayeuhugyu/minepack/main/install.sh
- Windows: https://raw.githubusercontent.com/ayeuhugyu/minepack/main/install.ps1

### Alternative: Manual Installation

If you prefer not to use the automated scripts:

1. Download the binary from [releases](https://github.com/ayeuhugyu/minepack/releases)
2. Extract/rename the binary to `minepack` (or `minepack.exe` on Windows)
3. Move to a directory in your PATH
4. Make executable (Unix: `chmod +x minepack`)

## Development

### Testing Changes

Before committing changes to installation scripts, test them thoroughly:

1. Test syntax validation
2. Test with mock releases
3. Test installation to a temporary directory
4. Verify PATH handling
5. Test backup functionality
6. Test error conditions

### Release Checklist

When creating a new release:

1. Ensure GitHub Actions builds all platform binaries
2. Verify binary naming follows the pattern: `minepack-{os}-{arch}[.exe]`
3. Test installation scripts with the new release
4. Update README if there are any changes to installation process

## References

These scripts were inspired by:
- [Bun installation scripts](https://bun.sh)
- [Rustup installation script](https://rustup.rs)
- [Node.js installation scripts](https://nodejs.org)
