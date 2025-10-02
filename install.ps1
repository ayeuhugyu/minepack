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
  [Win32.NativeMethods]::SendMessageTimeout($HWND_BROADCAST,
    $WM_SETTINGCHANGE,
    [UIntPtr]::Zero,
    "Environment",
    2,
    5000,
    [ref] $result
  ) | Out-Null
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
$Arch = "amd64"
if ([System.Environment]::Is64BitOperatingSystem -and [System.Environment]::Is64BitProcess) {
  $CpuArch = (Get-CimInstance Win32_Processor).Architecture
  # 9 = ARM64, 12 = ARM64 on x64 emulation
  if ($CpuArch -eq 9 -or $CpuArch -eq 12) {
    $Arch = "arm64"
  }
}

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
  return 1
}

$TagName = $Release.tag_name
if (-not $TagName) {
  Write-Output "Failed to fetch latest release information"
  return 1
}

# Find the appropriate asset
$AssetName = "minepack-windows-${Arch}.exe"
$Asset = $Release.assets | Where-Object { $_.name -eq $AssetName } | Select-Object -First 1

if (-not $Asset) {
  Write-Output "No binary found for Windows ${Arch}"
  return 1
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
  $VersionOutput = & $InstallPath --version 2>&1
  if ($LASTEXITCODE -ne 0) {
    throw "Binary execution failed"
  }
} catch {
  Write-Output "Installation verification failed. The binary may be corrupted."
  return 1
}

Write-Output ""
Write-Host "Successfully installed minepack $TagName!" -ForegroundColor Green
Write-Host "Binary location: $InstallPath" -ForegroundColor Green

# Update PATH if needed
if (-not $NoPathUpdate) {
  $CurrentPath = Get-Env -Key "Path"
  
  if ($CurrentPath -notlike "*$InstallDir*") {
    Write-Output ""
    Write-Output "Adding $InstallDir to PATH..."
    
    $NewPath = if ($CurrentPath) {
      "${CurrentPath};${InstallDir}"
    } else {
      $InstallDir
    }
    
    Write-Env -Key "Path" -Value $NewPath
    
    # Update current session
    $env:Path = "${env:Path};${InstallDir}"
    
    Write-Host "Added to PATH successfully!" -ForegroundColor Green
    Write-Host "You may need to restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
  } else {
    Write-Output "$InstallDir is already in PATH"
  }
}

Write-Output ""
Write-Output "Run 'minepack --help' to get started"
Write-Output "(You may need to restart your terminal first)"
