$ErrorActionPreference = 'Stop'

# URLs and checksums
$url64 = 'https://github.com/mohitmishra786/agentop/releases/download/v0.1.0/agentop_v0.1.0_windows_x86_64.zip'
$checksum64 = 'f770744f131f04e90d199a2fac7c9e4d740e2f5bb06a2a538374ebd286cc9747'
$urlarm64 = 'https://github.com/mohitmishra786/agentop/releases/download/v0.1.0/agentop_v0.1.0_windows_arm64.zip'
$checksumarm64 = '42b4c5a6859c37cb24ecdea8fb48c6c7331c2dd76f7c2d30b9f73b014c6a8f04'

# Determine architecture
if ([Environment]::Is64BitOperatingSystem -and $ENV:PROCESSOR_ARCHITECTURE -eq 'ARM64') {
    $url = $urlarm64
    $checksum = $checksumarm64
} else {
    $url = $url64
    $checksum = $checksum64
}

$packageArgs = @{
    PackageName  = 'agentop'
    Url64bit     = $url
    Checksum64   = $checksum
    ChecksumType = 'sha256'
}

Install-ChocolateyZipPackage @packageArgs

# Move binary to tools directory
$toolsPath = "$(Split-Path -Parent $MyInvocation.MyCommand.Path)"
$exePath = Join-Path $toolsPath "agentop.exe"
if (Test-Path $exePath) {
    # Binary is already in the right location after extraction
    Write-Host "agentop installed successfully!"
} else {
    throw "agentop.exe not found after extraction"
}

function Install-ChocolateyInstall {
  param($InstallPath)
  
  Write-Host "Installing agentop..."
  
  Get-RemoteFiles -Internalize "$($packageArgs['Url'])" -FileName "agentop.zip" -OutPath "$env:TEMP\agentop.zip"
  
  Write-Host "Extracting..."
  Expand-Archive "$env:TEMP\agentop.zip" -DestinationPath "$InstallPath" -Force
  
  Write-Host "Installing..."
  $exePath = Join-Path $InstallPath "agentop.exe"
  if (Test-Path $exePath) {
    Move-Item -Path "$exePath" -Destination "$InstallPath\agentop.exe"
  }
  
  Write-Host "Cleaning up..."
  Remove-Item "$env:TEMP\agentop.zip"
  
  Write-Host "agentop installed!"
}

function Get-ToolsLocation() {
  return Get-ToolPath
}
