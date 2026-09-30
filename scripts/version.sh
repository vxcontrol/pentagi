#!/bin/bash
# Source this file to set version environment variables
# Usage: source ./scripts/version.sh

# Get the latest git tag as version
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null) || LATEST_TAG="v0.0.0"
PACKAGE_VER="${LATEST_TAG#v}"

# The edition is the prerelease identifier every version string carries. Override it for
# an Enterprise build: PENTAGI_EDITION=ee source ./scripts/version.sh
PENTAGI_EDITION="${PENTAGI_EDITION:-ce}"

# A tag named after the edition, or the legacy "develop", is not a release. GetBinaryVersion
# reports the same string whether PackageVer is empty or one of those words, but the
# Dockerfile's label expression cannot tell them apart — it would read the word as a release
# number. Clearing it here keeps the three implementations in agreement.
case "$PACKAGE_VER" in
    develop|ce|ee) PACKAGE_VER="" ;;
esac

# Get current commit hash
CURRENT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "")

# Get commit hash of the latest tag
TAG_COMMIT=$(git rev-list -n 1 "$LATEST_TAG" 2>/dev/null || echo "")

# Set revision only if current commit differs from tag commit
if [ -n "$CURRENT_COMMIT" ] && [ "$CURRENT_COMMIT" != "$TAG_COMMIT" ]; then
    # A fixed width, not --short: git widens the abbreviation with the size of the clone,
    # so CI and a developer checkout would stamp the same commit with different revisions.
    PACKAGE_REV=$(printf '%s' "$CURRENT_COMMIT" | cut -c1-7)
else
    PACKAGE_REV=""
fi

# The string the binary will report, by the same rule as GetBinaryVersion in
# backend/pkg/version. The Dockerfile composes the image label from the same two args
# with an independent expression; build-image.sh compares the two, and
# check-version-rule.sh compares them over a matrix, so the rule cannot drift in one
# place unnoticed.
pentagi_version_string() {
    _ver="${1:-}"
    _rev="${2:-}"
    _ed="${PENTAGI_EDITION:-ce}"
    # The "h" keeps the revision an alphanumeric semver identifier: an all-digit
    # hex revision with a leading zero is not a legal numeric one.
    _rp="h"

    case "$_ver" in
        develop|ce|ee) _ver="" ;;
    esac

    if [ -z "$_ver" ]; then
        if [ -z "$_rev" ]; then
            printf '%s' "$_ed"
        else
            printf '%s' "$_ed.$_rp$_rev"
        fi
        return
    fi

    if [ -z "$_rev" ]; then
        printf '%s' "$_ver-$_ed"
    else
        printf '%s' "$_ver-$_ed.$_rp$_rev"
    fi
}

PACKAGE_VERSION_FULL="$(pentagi_version_string "$PACKAGE_VER" "$PACKAGE_REV")"

# Export variables for use in docker build
export PACKAGE_VER
export PENTAGI_EDITION
export PACKAGE_REV
export PACKAGE_VERSION_FULL

# Print version information
echo "======================================"
echo "PentAGI Build Version"
echo "======================================"
echo "PACKAGE_VER: $PACKAGE_VER"
if [ -n "$PACKAGE_REV" ]; then
    echo "PACKAGE_REV: $PACKAGE_REV (development)"
    echo "Full version: $PACKAGE_VERSION_FULL"
else
    echo "PACKAGE_REV: (release)"
    echo "Full version: $PACKAGE_VERSION_FULL"
fi
echo "======================================"
echo ""
echo "Environment variables exported:"
echo "  \$PACKAGE_VER = $PACKAGE_VER"
echo "  \$PACKAGE_REV = $PACKAGE_REV"
echo ""
