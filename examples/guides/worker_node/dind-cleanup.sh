#!/bin/bash
# =============================================================================
# dind-cleanup — remove nested DinD containers older than MAX_AGE_HOURS.
#
# Called hourly by the dind-cleanup.timer systemd unit.
# Uses /usr/local/bin/docker-dind-sock to connect to the DinD unix socket.
#
# Both running and stopped containers are removed. Worker containers are
# expected to be short-lived; anything alive for more than the threshold is
# considered stale and is force-removed.
# =============================================================================

set -uo pipefail

# A per-run override (MAX_AGE_HOURS=1 dind-cleanup) wins over the config file.
_MAX_AGE_HOURS="${MAX_AGE_HOURS:-}"

DIND_ENV="${DIND_ENV:-/etc/docker/dind/dind.env}"
if [ -r "${DIND_ENV}" ]; then
    . "${DIND_ENV}"
fi

MAX_AGE_HOURS="${_MAX_AGE_HOURS:-${MAX_AGE_HOURS:-24}}"
DOCKER="/usr/local/bin/docker-dind-sock"

log() { echo "[$(date '+%Y-%m-%dT%H:%M:%S%z')] dind-cleanup: $*"; }

# If DinD is not running, skip silently — Persistent=true ensures the timer
# retries at the next scheduled tick.
if ! "${DOCKER}" info >/dev/null 2>&1; then
    log "DinD daemon unreachable — skipping"
    exit 0
fi

# --filter until=Xh lists containers created more than X hours ago.
mapfile -t STALE < <("${DOCKER}" ps -a -q --filter "until=${MAX_AGE_HOURS}h" 2>/dev/null)

STALE_COUNT=${#STALE[@]}

if [ "$STALE_COUNT" -eq 0 ]; then
    log "No containers older than ${MAX_AGE_HOURS}h — nothing to do"
    exit 0
fi

log "Found $STALE_COUNT container(s) older than ${MAX_AGE_HOURS}h:"
"${DOCKER}" ps -a --filter "until=${MAX_AGE_HOURS}h" \
    --format 'table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.CreatedAt}}' 2>/dev/null || true

"${DOCKER}" rm -f "${STALE[@]}"
log "Removed $STALE_COUNT container(s)"
