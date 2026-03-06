# Source this file to set version environment variables
# Usage: . .\scripts\version.ps1

# Get the latest git tag as version
$latestTag = git describe --tags --abbrev=0 2>$null
if (-not $latestTag) {
    $latestTag = "v0.0.0"
}
$env:PACKAGE_VER = $latestTag.TrimStart('v')

# Get current commit hash
$currentCommit = git rev-parse HEAD 2>$null

# Get commit hash of the latest tag
$tagCommit = git rev-list -n 1 $latestTag 2>$null

# Set revision only if current commit differs from tag commit
if ($currentCommit -and ($currentCommit -ne $tagCommit)) {
    $env:PACKAGE_REV = git rev-parse --short HEAD
    $buildType = "development"
    $fullVersion = "$env:PACKAGE_VER-$env:PACKAGE_REV"
} else {
    $env:PACKAGE_REV = ""
    $buildType = "release"
    $fullVersion = $env:PACKAGE_VER
}

# Print version information
Write-Host "======================================"
Write-Host "PentAGI Build Version"
Write-Host "======================================"
Write-Host "PACKAGE_VER: $env:PACKAGE_VER"
if ($env:PACKAGE_REV) {
    Write-Host "PACKAGE_REV: $env:PACKAGE_REV ($buildType)"
} else {
    Write-Host "PACKAGE_REV: ($buildType)"
}
Write-Host "Full version: $fullVersion"
Write-Host "======================================"
Write-Host ""
Write-Host "Environment variables exported:"
Write-Host "  `$env:PACKAGE_VER = $env:PACKAGE_VER"
Write-Host "  `$env:PACKAGE_REV = $env:PACKAGE_REV"
Write-Host ""
