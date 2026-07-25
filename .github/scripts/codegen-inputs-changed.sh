#!/usr/bin/env bash
# Decide whether a run has to re-check that frontend/src/graphql/types.ts is fresh.
# Echoes `changed=true|false`. Split out of ci.yml so the range selection is testable
# (frontend/e2e/ci-codegen-gate.unit.test.ts drives it over real git fixtures).
#
# Usage: codegen-inputs-changed.sh <event_name> <before_sha> <base_sha> <head_sha>
set -uo pipefail

EVENT="${1:-}"
BEFORE="${2:-}"
BASE_SHA="${3:-}"
HEAD_SHA="${4:-}"

# On pull_request, `before` is the PR's PREVIOUS HEAD, so a follow-up push diffs a range
# that no longer contains the commit that edited the schema — the gate would skip itself
# on every push after the first. The PR's base is the range that always spans the change.
if [ "$EVENT" = "pull_request" ]; then
    base="$BASE_SHA"
else
    base="$BEFORE"
fi

# When the range can't be resolved (new branch, force-push, tag), check anyway.
if [ -z "$base" ] || [ "$base" = "0000000000000000000000000000000000000000" ] \
    || ! git cat-file -e "$base^{commit}" 2>/dev/null; then
    echo "changed=true"
    exit 0
fi

# The generated file is in the list too: a push that edits only it would otherwise skip
# the check, and the drift surfaces later on someone else's unrelated codegen push.
if git diff --name-only "$base" "$HEAD_SHA" | grep -qE '^(backend/pkg/graph/schema\.graphqls|frontend/graphql-schema\.graphql|frontend/graphql-codegen\.ts|frontend/pnpm-lock\.yaml|frontend/src/graphql/types\.ts)$'; then
    echo "changed=true"
else
    echo "changed=false"
fi
