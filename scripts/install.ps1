#!/usr/bin/env pwsh
# Minepack installation script for Windows
param(
  [String]$Version = "latest",
  # Skips adding minepack to the user's PATH
  [Switch]$NoPathUpdate = $false
)

$ErrorActionPreference = "Stop"

# Check for x64 Windows
if (-not ((Get-CimInstance Win32_ComputerSystem)).SystemType -match "x64-based") {
  Write-Output "Install Failed:"
  Write-Output "Minepack for Windows is currently only available for x86 64-bit Windows.`n"
  return 1
}

# Environment functions for PATH management
# Based on https://github.com/prefix-dev/pixi/pull/692
function Publish-Env {
  try {
    if (-not ("Win32.NativeMethods" -as [Type])) {
      Add-Type -Namespace Win32 -Name NativeMethods -MemberDefinition @"
[DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
public static extern IntPtr SendMessageTimeout(
    IntPtr hWnd, uint Msg, UIntPtr wParam, string lParam,
    uint fuFlags, uint uTimeout, out UIntPtr lpdwResult);
"@
    }
    $HWND_BROADCAST = [IntPtr] 0xffff
    $WM_SETTINGCHANGE = 0x1a
    $result = [UIntPtr]::Zero
    
    # Use shorter timeout and don't wait for response
    $success = [Win32.NativeMethods]::SendMessageTimeout($HWND_BROADCAST,
      $WM_SETTINGCHANGE,
      [UIntPtr]::Zero,
      "Environment",
      2,  # SMTO_ABORTIFHUNG
      2000,  # 2 second timeout instead of 5
      [ref] $result
    )
    
    if (-not $success) {
      Write-Verbose "SendMessageTimeout failed, but continuing anyway"
    }
  } catch {
    Write-Verbose "Failed to broadcast environment change: $_"
    # Don't fail the whole installation for this
  }
}

function Write-Env {
  param([String]$Key, [String]$Value)

  $RegisterKey = Get-Item -Path 'HKCU:'
  $EnvRegisterKey = $RegisterKey.OpenSubKey('Environment', $true)
  
  if ($null -eq $Value) {
    $EnvRegisterKey.DeleteValue($Key)
  } else {
    $RegistryValueKind = if ($Value.Contains('%')) {
      [Microsoft.Win32.RegistryValueKind]::ExpandString
    } elseif ($EnvRegisterKey.GetValue($Key)) {
      $EnvRegisterKey.GetValueKind($Key)
    } else {
      [Microsoft.Win32.RegistryValueKind]::String
    }
    $EnvRegisterKey.SetValue($Key, $Value, $RegistryValueKind)
  }

  Publish-Env
}

function Get-Env {
  param([String] $Key)

  $RegisterKey = Get-Item -Path 'HKCU:'
  $EnvRegisterKey = $RegisterKey.OpenSubKey('Environment')
  $EnvRegisterKey.GetValue($Key, $null, [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames)
}

# Determine architecture
$Arch = "amd64"  # Default to amd64 for maximum compatibility

# Only use ARM64 if we're absolutely sure it's needed and supported
if ([System.Environment]::Is64BitOperatingSystem -and [System.Environment]::Is64BitProcess) {
  $CpuArch = (Get-CimInstance Win32_Processor).Architecture
  # Only use ARM64 if processor is ARM64 AND we're running native ARM64 PowerShell
  if (($CpuArch -eq 9) -and ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq "Arm64")) {
    $Arch = "arm64"
  }
}

Write-Output "Detected architecture: $Arch"

# Set installation directory
$InstallDir = "${env:USERPROFILE}\.minepack\bin"
if (-not (Test-Path $InstallDir)) {
  New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
  Write-Output "Created directory: $InstallDir"
}

$BinaryName = "minepack.exe"
$InstallPath = Join-Path $InstallDir $BinaryName

# Fetch latest release information
Write-Output "Fetching latest release information..."
$ReleaseUrl = "https://api.github.com/repos/ayeuhugyu/minepack/releases/latest"

try {
  $Release = Invoke-RestMethod -Uri $ReleaseUrl -UseBasicParsing
} catch {
  Write-Output "Error fetching release information: $_"
  Write-Output "Please check your internet connection and try again."
  return 1
}

$TagName = $Release.tag_name
if (-not $TagName) {
  Write-Output "Failed to fetch latest release information"
  Write-Output "Please check your internet connection and try again."
  return 1
}

# Find the appropriate asset
$AssetName = "minepack-windows-${Arch}.exe"
$Asset = $Release.assets | Where-Object { $_.name -eq $AssetName } | Select-Object -First 1

if (-not $Asset) {
  Write-Output "No binary found for Windows ${Arch}"
  Write-Output "Available assets:"
  $Release.assets | ForEach-Object { Write-Output "  - $($_.name)" }
  
  # Fallback to amd64 if ARM64 wasn't found
  if ($Arch -eq "arm64") {
    Write-Output "Falling back to amd64 binary..."
    $Arch = "amd64"
    $AssetName = "minepack-windows-${Arch}.exe"
    $Asset = $Release.assets | Where-Object { $_.name -eq $AssetName } | Select-Object -First 1
    
    if (-not $Asset) {
      Write-Output "No amd64 binary found either"
      return 1
    }
  } else {
    return 1
  }
}

$DownloadUrl = $Asset.browser_download_url

Write-Output "Installing minepack $TagName for Windows ${Arch}..."
Write-Output "Download URL: $DownloadUrl"

# Download binary
Write-Output "Downloading..."
try {
  # Create backup if exists
  if (Test-Path $InstallPath) {
    $BackupPath = "${InstallPath}.backup"
    Write-Output "Creating backup of existing installation..."
    Copy-Item $InstallPath $BackupPath -Force
    Write-Output "Backup created at: $BackupPath"
  }

  # Download with progress
  $ProgressPreference = 'SilentlyContinue'
  Invoke-WebRequest -Uri $DownloadUrl -OutFile $InstallPath -UseBasicParsing
  $ProgressPreference = 'Continue'
  
} catch {
  Write-Output "Failed to download minepack: $_"
  return 1
}

# Verify installation
Write-Output "Verifying installation..."
try {
  # Test if the binary exists and is executable
  if (-not (Test-Path $InstallPath)) {
    throw "Binary not found at $InstallPath"
  }
  
  # Try to run the binary with --help (more likely to exist than --version)
  $output = & $InstallPath --help 2>&1
  $exitCode = $LASTEXITCODE
  
  # Check if it's a valid executable (exit code 0 for --help is good, or if it shows help text)
  if ($exitCode -eq 0 -or $output -match "minepack|Usage:|Commands:") {
    Write-Output "Installation verified successfully!"
  } else {
    throw "Binary execution failed with exit code: $exitCode"
  }
} catch {
  Write-Output "Installation verification failed: $_"
  Write-Output "The binary may be corrupted or incompatible with your system."
  
  # Restore backup if it exists
  $BackupPath = "${InstallPath}.backup"
  if (Test-Path $BackupPath) {
    Write-Output "Restoring backup..."
    Copy-Item $BackupPath $InstallPath -Force
    Remove-Item $BackupPath -Force
  }
  return 1
}

Write-Output ""
Write-Host "Successfully installed minepack $TagName!" -ForegroundColor Green
Write-Host "Binary location: $InstallPath" -ForegroundColor Green

# Update PATH if needed
if (-not $NoPathUpdate) {
  try {
    $CurrentPath = Get-Env -Key "Path"
    
    if ($CurrentPath -notlike "*$InstallDir*") {
      Write-Output ""
      Write-Output "Adding $InstallDir to PATH..."
      
      $NewPath = if ($CurrentPath) {
        "${CurrentPath};${InstallDir}"
      } else {
        $InstallDir
      }
      
      # Try to update PATH with timeout protection
      try {
        Write-Env -Key "Path" -Value $NewPath
        Write-Host "Added to PATH successfully!" -ForegroundColor Green
      } catch {
        Write-Output "Warning: Failed to update system PATH: $_"
        Write-Output "You can manually add $InstallDir to your PATH"
      }
      
      # Update current session (this always works)
      $env:Path = "${env:Path};${InstallDir}"
      Write-Host "PATH updated for current session." -ForegroundColor Green
      Write-Host "You may need to restart your terminal for permanent PATH changes to take effect." -ForegroundColor Yellow
    } else {
      Write-Output "$InstallDir is already in PATH"
    }
  } catch {
    Write-Output "Warning: Could not check or update PATH: $_"
    Write-Output "You can manually add $InstallDir to your PATH environment variable"
    
    # Still try to update current session
    try {
      $env:Path = "${env:Path};${InstallDir}"
      Write-Output "PATH updated for current session only."
    } catch {
      Write-Output "Could not update PATH for current session either."
    }
  }
}

Write-Output ""
Write-Output "Run 'minepack --help' to get started"
Write-Output "(You may need to restart your terminal first)"
