# Source this file to set version environment variables
# Usage: . .\scripts\version.ps1

# Get the latest git tag as version
$latestTag = git describe --tags --abbrev=0 2>$null
if (-not $latestTag) {
    $latestTag = "v0.0.0"
}
$env:PACKAGE_VER = $latestTag.TrimStart('v')

# The edition is the prerelease identifier every version string carries. Override it for
# an Enterprise build: $env:PENTAGI_EDITION = "ee"
if (-not $env:PENTAGI_EDITION) {
    $env:PENTAGI_EDITION = "ce"
}

# A tag named after the edition, or the legacy "develop", is not a release — see
# scripts/version.sh.
if (@("develop", "ce", "ee") -ccontains $env:PACKAGE_VER) {
    $env:PACKAGE_VER = ""
}

# Get current commit hash
$currentCommit = git rev-parse HEAD 2>$null

# Get commit hash of the latest tag
$tagCommit = git rev-list -n 1 $latestTag 2>$null

# Set revision only if current commit differs from tag commit
if ($currentCommit -and ($currentCommit -ne $tagCommit)) {
    # A fixed width, not --short: git widens the abbreviation with the size of the clone.
    $env:PACKAGE_REV = $currentCommit.Substring(0, 7)
    $buildType = "development"
} else {
    $env:PACKAGE_REV = ""
    $buildType = "release"
}

# The string the binary will report, by the same rule as GetBinaryVersion in
# backend/pkg/version and scripts/version.sh: the edition is a prerelease identifier and
# the revision is a second identifier after a dot.
$edition = $env:PENTAGI_EDITION
if ($env:PACKAGE_VER) {
    $versionPart = "$env:PACKAGE_VER-$edition"
} else {
    $versionPart = $edition
}
if ($env:PACKAGE_REV) {
    # "h" keeps the revision an alphanumeric semver identifier — see scripts/version.sh.
    $env:PACKAGE_VERSION_FULL = "$versionPart.h$env:PACKAGE_REV"
} else {
    $env:PACKAGE_VERSION_FULL = $versionPart
}
$fullVersion = $env:PACKAGE_VERSION_FULL

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
Write-Host "  `$env:PACKAGE_VERSION_FULL = $env:PACKAGE_VERSION_FULL"
Write-Host ""
