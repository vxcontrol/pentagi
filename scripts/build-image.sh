#!/usr/bin/env bash
# Builds the image the way the pipeline does — the version and the revision have to come
# from outside, because the api-builder stage copies backend/ only and there is no .git in
# the build to ask. Passing the same two args by hand gets the same result; what this adds
# is the check that the label the Dockerfile composed matches version.sh's rule.
# scripts/check-version-rule.sh ties both to what the binary reports.
#
# Usage: ./scripts/build-image.sh [TAG]      (default tag: local/pentagi:<revision-or-version>)
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

# shellcheck source=./version.sh
source ./scripts/version.sh > /dev/null

TAG="${1:-local/pentagi:${PACKAGE_REV:-$PACKAGE_VERSION_FULL}}"

echo "building $TAG with PACKAGE_VER=$PACKAGE_VER PACKAGE_REV=${PACKAGE_REV:-<none, this is a tag commit>}"

# Only the pair, the way the CI pipelines pass it, so the check below exercises the path CI
# takes from a release tag or from no tag.
docker build \
    --build-arg PACKAGE_VER="$PACKAGE_VER" \
    --build-arg PACKAGE_REV="$PACKAGE_REV" \
    --tag "$TAG" \
    .

# The Dockerfile composes the label from the pair; version.sh composes the same string in
# shell. A difference here means one of the two rules moved.
stamped="$(docker inspect "$TAG" --format '{{index .Config.Labels "com.pentagi.version"}}')"

if [ "$stamped" != "$PACKAGE_VERSION_FULL" ]; then
    echo "build-image: $TAG is stamped '$stamped', expected '$PACKAGE_VERSION_FULL'" >&2
    exit 1
fi

if [ -n "$PACKAGE_REV" ]; then
    echo "built $TAG — the label reads $PACKAGE_VERSION_FULL; the panel puts $PACKAGE_REV on its build line"
else
    echo "built $TAG — a tag commit: the label reads $PACKAGE_VERSION_FULL, and the panel's build line shows the binary hash"
fi
