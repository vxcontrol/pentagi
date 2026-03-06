#!/bin/bash
# Source this file to set version environment variables
# Usage: source ./scripts/version.sh

# Get the latest git tag as version
PACKAGE_VER=$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.0")

# Get current commit hash
CURRENT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "")

# Get commit hash of the latest tag
TAG_COMMIT=$(git rev-list -n 1 "$(git describe --tags --abbrev=0 2>/dev/null || echo HEAD)" 2>/dev/null || echo "")

# Set revision only if current commit differs from tag commit
if [ -n "$CURRENT_COMMIT" ] && [ "$CURRENT_COMMIT" != "$TAG_COMMIT" ]; then
    PACKAGE_REV=$(git rev-parse --short HEAD)
else
    PACKAGE_REV=""
fi

# Export variables for use in docker build
export PACKAGE_VER
export PACKAGE_REV

# Print version information
echo "======================================"
echo "PentAGI Build Version"
echo "======================================"
echo "PACKAGE_VER: $PACKAGE_VER"
if [ -n "$PACKAGE_REV" ]; then
    echo "PACKAGE_REV: $PACKAGE_REV (development)"
    echo "Full version: $PACKAGE_VER-$PACKAGE_REV"
else
    echo "PACKAGE_REV: (release)"
    echo "Full version: $PACKAGE_VER"
fi
echo "======================================"
echo ""
echo "Environment variables exported:"
echo "  \$PACKAGE_VER = $PACKAGE_VER"
echo "  \$PACKAGE_REV = $PACKAGE_REV"
echo ""
