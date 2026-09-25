#!/usr/bin/env bash
# Measures the image label against the version the binary will report, over every pair of
# build args the pipelines produce from a release tag or from no tag.
#
# The binary side is what the api-builder stage stamps: its ARG lines and the ldflags
# expressions are taken out of the real Dockerfile and evaluated on alpine, and the stamped
# pair is looked up in backend/pkg/version/testdata/version_strings.tsv, the table
# GetBinaryVersion's own test reads. It costs seconds and no Go toolchain.
#
# Usage: ./scripts/check-version-rule.sh
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

# shellcheck source=./version.sh
source ./scripts/version.sh > /dev/null

TABLE="backend/pkg/version/testdata/version_strings.tsv"

SCRATCH="$(mktemp -d)"
trap 'rm -rf "$SCRATCH"' EXIT

failures=0

table_lookup() {
    awk -F'\t' -v v="$1" -v r="$2" '$1 == v && $2 == r { print $3; found = 1; exit } END { exit !found }' "$TABLE"
}

while IFS= read -r row; do
    ver="${row%%$'\t'*}"
    rest="${row#*$'\t'}"
    rev="${rest%%$'\t'*}"
    want="${rest#*$'\t'}"
    got="$(pentagi_version_string "$ver" "$rev")"
    if [ "$got" != "$want" ]; then
        printf 'FAIL  version.sh PACKAGE_VER=%s PACKAGE_REV=%s -> %s, the table says %s\n' \
            "'$ver'" "'$rev'" "$got" "$want"
        failures=$((failures + 1))
    fi
done < "$TABLE"

builder_start="$(grep -n '^FROM .* AS api-builder$' Dockerfile | cut -d: -f1)"
builder_end="$(awk -v s="$builder_start" 'NR > s && /^FROM / { print NR - 1; exit }' Dockerfile)"
builder="$(sed -n "${builder_start},${builder_end}p" Dockerfile)"

stamp_expression() {
    printf '%s\n' "$builder" | sed -n "s/.*-X pentagi\/pkg\/version\.$1=\([^ \"]*\).*/\1/p" | sort -u
}
ver_expr="$(stamp_expression PackageVer)"
rev_expr="$(stamp_expression PackageRev)"
if [ "$(printf '%s\n' "$ver_expr" | grep -c .)" -ne 1 ] || [ "$(printf '%s\n' "$rev_expr" | grep -c .)" -ne 1 ]; then
    echo "check-version-rule: the api-builder builds do not stamp PackageVer and PackageRev one way" >&2
    exit 1
fi

# The final stage alone: a re-declared ARG keeps the default an earlier declaration in the
# same stage gave it, so the two stages stay apart here as they are in the Dockerfile.
last_from="$(grep -n '^FROM ' Dockerfile | tail -1 | cut -d: -f1)"

{
    echo "FROM alpine:3.20 AS stamp"
    printf '%s\n' "$builder" | grep -E '^ARG (PACKAGE_VER|PACKAGE_REV|PACKAGE_EDITION)(=|$)'
    cat <<STAMP
RUN printf '%s\\t%s' "$ver_expr" "$rev_expr" > /stamped
STAMP
    echo "FROM alpine:3.20"
    tail -n "+$last_from" Dockerfile |
        grep -E '^ARG (PACKAGE_VER|PACKAGE_REV|PACKAGE_EDITION|VER_PART|PACKAGE_VERSION_FULL)|^LABEL com\.pentagi\.version='
    echo "COPY --from=stamp /stamped /stamped"
} > "$SCRATCH/Dockerfile"

if [ "$(grep -c '^LABEL' "$SCRATCH/Dockerfile")" -ne 1 ]; then
    echo "check-version-rule: expected exactly one com.pentagi.version LABEL in the Dockerfile" >&2
    exit 1
fi

# How the args are passed is part of the input. A bare `docker build .` leaves both ARGs at
# their declared default; passing an empty string overrides that default; passing only the
# revision leaves PACKAGE_VER at whatever the stage declares.
check() {
    ver="$1"
    rev="$2"
    how="${3:-pass-args}"

    case "$how" in
        no-args) set -- --tag pentagi-version-rule:check ;;
        rev-only) set -- --build-arg PACKAGE_REV="$rev" --tag pentagi-version-rule:check ;;
        *)
            set -- --build-arg PACKAGE_VER="$ver" --build-arg PACKAGE_REV="$rev" \
                --tag pentagi-version-rule:check
            ;;
    esac

    docker build --quiet "$@" "$SCRATCH" > /dev/null

    label="$(docker inspect pentagi-version-rule:check --format '{{index .Config.Labels "com.pentagi.version"}}')"
    stamped="$(docker run --rm pentagi-version-rule:check cat /stamped)"
    stamped_ver="${stamped%%$'\t'*}"
    stamped_rev="${stamped#*$'\t'}"
    shell="$(pentagi_version_string "$ver" "$rev")"

    # An empty stamped PackageVer used to be a failure on its own, because the binary
    # then said "develop" while the label said whatever the expression produced. Both
    # sides now read it as "no release" and report the edition, so the table decides
    # this case like any other.
    verdict=""
    if ! binary="$(table_lookup "$stamped_ver" "$stamped_rev")"; then
        verdict="the stamped pair '$stamped_ver' '$stamped_rev' is not in $TABLE"
    elif [ "$binary" != "$label" ] || [ "$shell" != "$label" ]; then
        verdict="label $label, binary $binary, version.sh $shell"
    fi

    if [ -z "$verdict" ]; then
        printf 'ok    %-9s PACKAGE_VER=%-8s PACKAGE_REV=%-8s -> %s\n' "$how" "'$ver'" "'$rev'" "$label"
    else
        printf 'FAIL  %-9s PACKAGE_VER=%-8s PACKAGE_REV=%-8s -> %s\n' "$how" "'$ver'" "'$rev'" "$verdict"
        failures=$((failures + 1))
    fi
}

# The domain: a release number or nothing, a revision or nothing. A tag named "develop" is
# not a release and not in it — see the note above the LABEL in the Dockerfile. version.sh
# and version.ps1 clear the word; the CI jobs pass it through, so this check does not vouch
# for a pipeline build from such a tag.
check "" "" no-args
check "" "abc1234" rev-only
check "" ""
check "" "abc1234"
check "2.1.0" ""
check "2.1.0" "abc1234"

docker image rm pentagi-version-rule:check > /dev/null 2>&1 || true

if [ "$failures" -ne 0 ]; then
    echo "check-version-rule: $failures check(s) disagree — the label and the binary would not match" >&2
    exit 1
fi

echo "check-version-rule: label, binary and version.sh agree on all six"
