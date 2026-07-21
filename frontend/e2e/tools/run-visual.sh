#!/usr/bin/env bash
# Visual snapshots run ONLY inside the pinned Playwright container so baselines
# are pixel-stable across every machine and CI. The image tag is derived from
# the installed @playwright/test version — the pin cannot drift by construction.
# Usage: run-visual.sh [--update]   (E2E_SKIP_BUILD=1 reuses the existing dist)
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
FRONTEND="$REPO_ROOT/frontend"

PLAYWRIGHT_VERSION="$(node -p "require('$FRONTEND/node_modules/@playwright/test/package.json').version")"
IMAGE="mcr.microsoft.com/playwright:v${PLAYWRIGHT_VERSION}-noble"

if [[ "${E2E_SKIP_BUILD:-}" != "1" ]]; then
    (cd "$FRONTEND" && corepack pnpm run build)
fi

UPDATE_ARGS=()
if [[ "${1:-}" == "--update" ]]; then
    # `=all` binds the mode so a trailing spec-file filter isn't swallowed as the
    # option's optional [mode] argument (`--update-snapshots <file>` fails to parse).
    UPDATE_ARGS=(--update-snapshots=all)
    shift
fi

# --user: the image defaults to root, and on Linux a root-owned --update would
# write baselines/test-results into the repo that git/rm can't touch without
# sudo (masked on macOS Docker Desktop by its uid mapping).
docker run --rm \
    -v "$REPO_ROOT:/work" \
    -w /work/frontend \
    -e E2E_VISUAL=1 \
    -e E2E_VISUAL_CONTAINER=1 \
    -e CI="${CI:-}" \
    -e HOME=/tmp \
    --user "$(id -u):$(id -g)" \
    --ipc=host \
    "$IMAGE" \
    npx playwright test -c e2e/playwright.config.ts ${UPDATE_ARGS[@]+"${UPDATE_ARGS[@]}"} "$@"
