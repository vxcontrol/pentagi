#!/usr/bin/env bash
# A throwaway checkout for agents that need to MUTATE code to test something — inject a
# regression, delete a guard, probe a fixture. Review runs kept doing that in the live tree:
# probe specs left behind, tracked files edited, and once an `rm -rf` of a directory the run
# did not create.
#
# Usage:
#   review-sandbox.sh create [--dirty] [--with-deps]   # prints the sandbox path on stdout
#   review-sandbox.sh clean <path>                     # one sandbox, refuses anything else
#   review-sandbox.sh clean --all                      # sweep sandboxes untouched for 2h
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
SANDBOX_ROOT="${PENTAGI_SANDBOX_ROOT:-${TMPDIR:-/tmp}}/pentagi-review-sandboxes"
# A hardlink cannot cross filesystems, so --with-deps is unusable when $TMPDIR is its own mount
# (tmpfs /tmp, a separate /home, a container volume): `cp -al` fails outright and no sandbox is
# built. This sibling of the repo is on the repo's filesystem by construction.
# Under the escape hatch BOTH roots move: `clean --all` sweeps this one too, so a caller that cannot
# afford a sweep of the repo's parent (the unit test, which otherwise rm -rf's real directories on the
# developer's machine) needs a single variable that contains every path this script may delete.
FALLBACK_ROOT="${PENTAGI_SANDBOX_ROOT:+${PENTAGI_SANDBOX_ROOT}/.pentagi-review-sandboxes}"
FALLBACK_ROOT="${FALLBACK_ROOT:-$(dirname "$REPO_ROOT")/.pentagi-review-sandboxes}"

# Pick a root node_modules can actually be hardlinked into: probe with a real link, fall back to
# the repo's own filesystem when the probe hits EXDEV.
linkable_root() {
    local probe="$SANDBOX_ROOT/.linkprobe"

    mkdir -p "$SANDBOX_ROOT"

    if ln "${BASH_SOURCE[0]}" "$probe" 2>/dev/null; then
        rm -f "$probe"
        echo "$SANDBOX_ROOT"

        return
    fi

    if ! mkdir -p "$FALLBACK_ROOT" 2>/dev/null; then
        echo "review-sandbox: $SANDBOX_ROOT is on another filesystem than the repo and" \
            "$FALLBACK_ROOT is not writable — set PENTAGI_SANDBOX_ROOT to a same-filesystem path" >&2
        exit 2
    fi

    echo "review-sandbox: hardlinks cannot reach $SANDBOX_ROOT — using $FALLBACK_ROOT" >&2
    echo "$FALLBACK_ROOT"
}

create() {
    local carry_dirty="" with_deps="" path root="$SANDBOX_ROOT"

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

    if [[ -n "$with_deps" ]]; then
        root="$(linkable_root)"
    fi

    path="$root/wt-$$-$(date +%s)"
    mkdir -p "$root"
    git -C "$REPO_ROOT" worktree add --detach --quiet "$path" HEAD
    # Anything that fails from here leaves a worktree behind, and the caller reads stdout as the
    # sandbox path — so take the worktree down and say so on stderr rather than printing a stub.
    trap 'clean "$path" >/dev/null 2>&1; echo "review-sandbox: failed to prepare $path" >&2' ERR

    if [[ -n "$carry_dirty" ]]; then
        # --binary or a regenerated PNG baseline aborts the patch ("cannot apply binary patch
        # without full index line"). Untracked files are absent from the diff entirely, so a new
        # spec would never reach the sandbox — copy them separately, minus the ignored ones.
        if ! git -C "$REPO_ROOT" diff --quiet HEAD; then
            git -C "$REPO_ROOT" diff HEAD --binary | git -C "$path" apply --allow-empty
        fi

        while IFS= read -r -d '' file; do
            mkdir -p "$path/$(dirname "$file")"
            cp "$REPO_ROOT/$file" "$path/$file"
        done < <(git -C "$REPO_ROOT" ls-files --others --exclude-standard -z)
    fi

    if [[ -n "$with_deps" ]]; then
        for dir in frontend/node_modules node_modules; do
            [[ -d "$REPO_ROOT/$dir" ]] && cp -al "$REPO_ROOT/$dir" "$path/$dir"
        done
    fi

    echo "$path"
}

# `git worktree remove` exits 128 for the main working tree and for any directory that is not a
# worktree, so the rm fallback is what runs for every wrong argument — and being left of `||` it
# does not trip `set -e`. Resolve the target and refuse anything outside the sandbox root before
# going near it. The sweep is opt-in and age-filtered for the same reason: this root is shared by
# every agent on the host, so a bare `clean` used to take out a concurrent agent's live sandbox.
STALE_MINUTES=120

# A worktree directory's own mtime moves only when a top-level entry is added or removed, so it dates
# the sandbox's creation — an agent editing frontend/src/... for hours never refreshes it. Age the
# sandbox by the newest write anywhere inside it instead.
recently_touched() {
    [[ -n "$(find "$1" -mmin "-$STALE_MINUTES" -print -quit 2>/dev/null)" ]]
}

# A strict child of either sandbox root — never a root itself, so a sibling like <root>-old and the
# root's own path are both rejected.
contained() {
    local candidate="$1" root

    for root in "$SANDBOX_ROOT" "$FALLBACK_ROOT"; do
        mkdir -p "$root"

        case "$candidate" in
            "$(cd "$root" && pwd -P)"/?*) return 0 ;;
        esac
    done

    return 1
}

clean() {
    local target="${1:-}" root resolved

    case "$target" in
        --all)
            for root in "$SANDBOX_ROOT" "$FALLBACK_ROOT"; do
                # The roots are derived, not passed in: naming them is how a caller — or a test —
                # can tell where this is about to delete.
                echo "swept $root"
                for path in "$root"/wt-*; do
                    if [[ -d "$path" ]] && ! recently_touched "$path"; then
                        rm -rf "$path"
                    fi
                done
            done
            ;;
        '')
            echo "review-sandbox: clean needs a sandbox path, or --all to sweep stale ones" >&2
            exit 2
            ;;
        *)
            resolved="$(cd "$(dirname "$target")" 2>/dev/null && pwd -P || true)/$(basename "$target")"

            case "$resolved" in
                */. | */..) resolved="" ;;
                *) contained "$resolved" || resolved="" ;;
            esac

            if [[ -z "$resolved" ]]; then
                echo "review-sandbox: refusing to clean '$target' — not a sandbox under" \
                    "$SANDBOX_ROOT or $FALLBACK_ROOT" >&2
                exit 2
            fi

            git -C "$REPO_ROOT" worktree remove --force "$resolved" || rm -rf "$resolved"
            ;;
    esac

    git -C "$REPO_ROOT" worktree prune
}

case "${1:-}" in
    create)
        shift
        create "$@"
        ;;
    clean) clean "${2:-}" ;;
    *)
        echo "usage: $(basename "$0") create [--dirty] [--with-deps] | clean <path>|--all" >&2
        exit 2
        ;;
esac
