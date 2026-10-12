# Download and verify an official pkx Windows release.

param(
    [string]$Version = "",
    [string]$InstallDir = (Join-Path $HOME ".local\bin"),
    [switch]$AddToPath,
    [switch]$Help
)
$ErrorActionPreference = "Stop"
if ($Help) {
    Write-Host "Usage: .\install.ps1 [-Version vX.Y.Z] [-InstallDir DIRECTORY] [-AddToPath]"
    Write-Host "Downloads an official pkx release and verifies its SHA256."
    Write-Host "Default: latest published release; installation directory: ~/.local/bin."
    Write-Host "-AddToPath: persist the install directory in your Windows USER PATH (no admin)."
    Write-Host "Open a NEW CMD/PowerShell window afterwards; existing windows keep their old PATH."
    exit 0
}


# Windows PATH is case-insensitive. Keep existing entries and never duplicate the
# canonical install directory in the user's persistent environment variables.
function Add-PkxUserPath {
    param([Parameter(Mandatory=$true)][string]$Directory)
    $fullDir = [System.IO.Path]::GetFullPath($Directory).TrimEnd([char[]]@('\', '/'))
    $original = [Environment]::GetEnvironmentVariable('Path', [EnvironmentVariableTarget]::User)
    $entries = @($original -split ';' | Where-Object { $_ -and $_.Trim() })
    $alreadyPresent = $false
    foreach ($entry in $entries) {
        $candidate = [Environment]::ExpandEnvironmentVariables($entry.Trim().Trim('"'))
        $candidate = $candidate.TrimEnd([char[]]@('\', '/'))
        if ([string]::Equals($candidate, $fullDir, [StringComparison]::OrdinalIgnoreCase)) {
            $alreadyPresent = $true
            break
        }
    }
    if (-not $alreadyPresent) {
        $newPath = if ([string]::IsNullOrWhiteSpace($original)) {
            $fullDir
        } else {
            $original.TrimEnd(';') + ';' + $fullDir
        }
        [Environment]::SetEnvironmentVariable('Path', $newPath, [EnvironmentVariableTarget]::User)
        Write-Host "Added $fullDir to your user PATH (persisted, no administrator required)."
    } else {
        Write-Host "Already on your user PATH: $fullDir"
    }
    Write-Host "Close all existing CMD / PowerShell windows and open a NEW terminal."
    Write-Host "Then verify: where pkx   (CMD) or Get-Command pkx   (PowerShell)"
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
    $InstallDir = [System.IO.Path]::GetFullPath($InstallDir)
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    $destination = Join-Path $InstallDir 'pkx.exe'
    Copy-Item -Path $binary -Destination $destination -Force
    Write-Host "Installed pkx $Version to $destination"
    if ($AddToPath) {
        Add-PkxUserPath -Directory $InstallDir
    } elseif (($env:Path -split ';') -notcontains $InstallDir) {
        Write-Host "For global CMD access, rerun with -AddToPath (opt-in), then open a new terminal."
    }
} finally {
    Remove-Item -Path $tmp -Recurse -Force -ErrorAction SilentlyContinue
}
