# Download and verify an official pkx Windows release.
param(
    [string]$Version = "",
    [string]$InstallDir = (Join-Path $HOME ".local\bin"),
    [switch]$Help
)
$ErrorActionPreference = "Stop"
if ($Help) {
    Write-Host "Usage: .\install.ps1 [-Version vX.Y.Z] [-InstallDir DIRECTORY]"
    Write-Host "Downloads an official pkx release and verifies its SHA256."
    Write-Host "Default: latest published release; installation directory: ~/.local/bin."
    exit 0
}

if (-not $Version) {
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/AmpedWasTaken/pkx/releases/latest" -Headers @{ "User-Agent" = "pkx-installer" }
        $Version = [string]$release.tag_name
    } catch {
        throw "No published pkx release found. Build from source or pass -Version explicitly."
    }
}
if ($Version -notmatch '^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.-]+)?$') {
    throw "Invalid release version '$Version' (expected vX.Y.Z)."
}

switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()) {
    'X64' { $arch = 'amd64' }
    'Arm64' { $arch = 'arm64' }
    default { throw "Unsupported Windows CPU architecture. Build pkx from source." }
}
$asset = "pkx_${Version}_windows_${arch}.zip"
$base = "https://github.com/AmpedWasTaken/pkx/releases/download/$Version"
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $tmp -Force | Out-Null
try {
    $archive = Join-Path $tmp $asset
    $checksums = Join-Path $tmp 'SHA256SUMS'
    Invoke-WebRequest -Uri "$base/$asset" -OutFile $archive
    Invoke-WebRequest -Uri "$base/SHA256SUMS" -OutFile $checksums

    $pattern = "(?m)^([0-9A-Fa-f]{64})\s+(?:\./)?$([regex]::Escape($asset))\s*$"
    $match = [regex]::Match((Get-Content -Raw $checksums), $pattern)
    if (-not $match.Success) { throw "Release checksums do not contain $asset." }
    $expected = $match.Groups[1].Value
    $actual = (Get-FileHash -Algorithm SHA256 -Path $archive).Hash
    if ($expected -ne $actual) { throw "SHA256 mismatch; refusing to install." }

    $unpack = Join-Path $tmp 'unpack'
    Expand-Archive -Path $archive -DestinationPath $unpack
    $binary = Join-Path $unpack 'pkx.exe'
    if (-not (Test-Path -PathType Leaf $binary)) { throw "Archive does not contain pkx.exe." }
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Copy-Item -Path $binary -Destination (Join-Path $InstallDir 'pkx.exe') -Force
    Write-Host "Installed pkx $Version to $(Join-Path $InstallDir 'pkx.exe')"
    if (($env:PATH -split ';') -notcontains $InstallDir) {
        Write-Host "Add $InstallDir to your PATH if pkx is not recognized."
    }
} finally {
    Remove-Item -Path $tmp -Recurse -Force -ErrorAction SilentlyContinue
}
