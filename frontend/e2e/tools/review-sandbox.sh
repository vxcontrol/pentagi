#!/usr/bin/env bash
# A throwaway checkout for agents that need to MUTATE code to test something — inject a
# regression, delete a guard, probe a fixture. Review runs kept doing that in the live tree:
# probe specs left behind, tracked files edited, and once an `rm -rf` of a directory the run
# did not create.
#
# Usage:
#   review-sandbox.sh create [--dirty] [--with-deps]   # prints the sandbox path on stdout
#   review-sandbox.sh clean [<path>]                   # one sandbox, or every stale one
#
#   --dirty      carry the uncommitted working tree across, for a scope that includes it
#   --with-deps  hardlink node_modules in, for agents that run pnpm/playwright/vitest
#
# A git worktree carries no ignored files, so node_modules has to be supplied. It is hardlinked
# rather than symlinked because pnpm treats a symlinked node_modules as foreign and tries to
# purge and reinstall it — through the link, into the real tree. Hardlinks share data, so the
# copy is cheap, but a tool that rewrites a dependency IN PLACE (rather than the usual
# write-temp-then-rename) would still reach the original: don't hand --with-deps to an agent
# whose job is patching libraries.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
SANDBOX_ROOT="${TMPDIR:-/tmp}/pentagi-review-sandboxes"

create() {
    local carry_dirty="" with_deps="" path

    for arg in "$@"; do
        case "$arg" in
            --dirty) carry_dirty=1 ;;
            --with-deps) with_deps=1 ;;
            *)
                echo "unknown option: $arg" >&2
                exit 2
                ;;
        esac
    done

    path="$SANDBOX_ROOT/wt-$$-$(date +%s)"
    mkdir -p "$SANDBOX_ROOT"
    git -C "$REPO_ROOT" worktree add --detach --quiet "$path" HEAD

    if [[ -n "$carry_dirty" ]] && ! git -C "$REPO_ROOT" diff --quiet HEAD; then
        git -C "$REPO_ROOT" diff HEAD | git -C "$path" apply --allow-empty
    fi

    if [[ -n "$with_deps" ]]; then
        for dir in frontend/node_modules node_modules; do
            [[ -d "$REPO_ROOT/$dir" ]] && cp -al "$REPO_ROOT/$dir" "$path/$dir"
        done
    fi

    echo "$path"
}

clean() {
    local target="${1:-}"

    if [[ -n "$target" ]]; then
        git -C "$REPO_ROOT" worktree remove --force "$target" 2>/dev/null || rm -rf "$target"
    else
        rm -rf "$SANDBOX_ROOT"
    fi

    git -C "$REPO_ROOT" worktree prune
}

case "${1:-}" in
    create)
        shift
        create "$@"
        ;;
    clean) clean "${2:-}" ;;
    *)
        echo "usage: $(basename "$0") create [--dirty] [--with-deps] | clean [<path>]" >&2
        exit 2
        ;;
esac
