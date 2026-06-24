# protoc-gen-jsonschema one-liner installer (Windows, PowerShell 5.1+).
#
#   irm https://raw.githubusercontent.com/sunerpy/protoc-gen-jsonschema/main/scripts/install.ps1 | iex
#
# Env overrides:
#   $env:PGJ_VERSION      pin a release (e.g. 0.0.8 or v0.0.8); default: latest
#   $env:PGJ_INSTALL_DIR  install destination; default: $HOME\.local\bin
$ErrorActionPreference = "Stop"

$Repo = "sunerpy/protoc-gen-jsonschema"
$Bin = "protoc-gen-jsonschema"

function Die($msg) {
  Write-Error $msg
  exit 1
}

# Detect arch (GoReleaser uses raw GOARCH values).
$arch = $env:PROCESSOR_ARCHITECTURE
switch ($arch) {
  "AMD64" { $archPart = "amd64" }
  "ARM64" { $archPart = "arm64" }
  default { Die "unsupported arch: $arch (supported: AMD64, ARM64)" }
}

# Resolve version: env override or latest-release API.
if ($env:PGJ_VERSION) {
  $version = $env:PGJ_VERSION -replace '^v', ''
} else {
  Write-Host "Resolving latest release..."
  $api = "https://api.github.com/repos/$Repo/releases/latest"
  $rel = Invoke-RestMethod -Uri $api -Headers @{ "User-Agent" = "install-script" }
  $version = $rel.tag_name -replace '^v', ''
}
if (-not $version) { Die "could not resolve release version" }

# Asset name matches .goreleaser.yaml: {ProjectName}_{Version}_windows_{Arch}.zip
$asset = "${Bin}_${version}_windows_${archPart}.zip"
$url = "https://github.com/$Repo/releases/download/v$version/$asset"
$installDir = if ($env:PGJ_INSTALL_DIR) { $env:PGJ_INSTALL_DIR } else { "$HOME\.local\bin" }

Write-Host "Installing $Bin v$version (windows/$archPart)"
Write-Host "  from: $url"
Write-Host "  to:   $installDir\$Bin.exe"

$tmp = New-Item -ItemType Directory -Path (Join-Path $env:TEMP ([System.Guid]::NewGuid()))
try {
  $zip = Join-Path $tmp $asset
  Invoke-WebRequest -Uri $url -OutFile $zip
  Expand-Archive -Path $zip -DestinationPath $tmp -Force

  New-Item -ItemType Directory -Force -Path $installDir | Out-Null
  Move-Item -Force -Path (Join-Path $tmp "$Bin.exe") -Destination (Join-Path $installDir "$Bin.exe")

  Write-Host "Installed $Bin to $installDir\$Bin.exe"

  $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
  if ($userPath -notlike "*$installDir*") {
    Write-Host "NOTE: $installDir is not on your PATH. Add it via:"
    Write-Host "  [Environment]::SetEnvironmentVariable('Path', `"$installDir;`$env:Path`", 'User')"
  }
} finally {
  Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
