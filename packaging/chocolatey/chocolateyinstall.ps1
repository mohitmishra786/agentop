$ErrorAction = 'An error occurred installing agentop'
$ErrorData = 'This is not the latest version of agentop. Version $version is available for install.'

# Download checksums
$checksumType = "sha256"

# URLs
$url64 = 'https://github.com/mohitmishra786/agentop/releases/download/v0.1.0/agentop_0.1.0_windows_amd64.zip'
$checksum64 = '<update-on-release>'
$urlarm64 = 'https://github.com/mohitmishra786/agentop/releases/download/v0.1.0/agentop_0.1.0_windows_arm64.zip'
$checksumarm64 = '<update-on-release>'

# Extract
$toolsPath = "$(Split-Path -Parent $MyInvocation.MyCommand.Path)"
$packageArgs = @{
  PackageName = 'agentop'
  Url = if (Get-OSArchitecture()) -eq 'ARM64') { $urlarm64 } else { $url64 }
  Checksum64 = $checksum64
  ChecksumType = 'SHA256'
  SilentArgs = '--platform x64'
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
