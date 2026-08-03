#!/bin/bash
# Docker-in-Docker container management script.
#
# Runs a single hardened DinD container and enforces the OPA authorization
# plugin as a MANAGED plugin installed INSIDE the DinD daemon (no sidecar).
# The plugin reads its policy from DinD's /etc/docker (mounted as /opa by the
# managed plugin); we mount the host policy directory onto DinD's /etc/docker.
#
# The plugin is installed once on first run and persists in the DinD data
# volume, so subsequent starts come up with authz directly. First run does a
# two-phase bootstrap: start without authz -> install plugin -> restart with
# authz enabled (fail-closed).
#
# Worker containers keep using the same DinD socket / TLS endpoint and are
# unaware of the authz layer.
#
# Configuration: every variable below can be overridden in /etc/docker/dind/dind.env
# (KEY=value lines) or via the environment.
#
# SECURITY NOTES (from pentest findings):
#
#   1. authz.rego blocks ALL bind-mounts of host paths for nested containers.
#      This closes the "chroot bypass": even if a nested container could mount
#      DinD's root and chroot into it, the OPA policy denies the mount.
#
#   2. The PentAGI backend HTTP service MUST run on HTTPS (SERVER_USE_SSL=true)
#      and MUST NOT be accessible from worker containers. Worker containers have
#      --network host; any plaintext HTTP port on the host will be reachable.
#      Exposing TLS certificates or sensitive files via an unauthenticated HTTP
#      endpoint completely defeats all isolation layers.
#      Ensure DATA_DIR and any directory served by the backend NEVER contains
#      TLS client certificates for the outer Docker daemon.
#
#   3. TLS client certificates for the outer Docker daemon MUST be stored
#      outside any web-accessible or world-readable path. CERTS_PATH below is
#      /etc/docker/dind/certs - keep it there and verify that no HTTP service
#      is configured to serve files from that directory or its parents.
#
#   4. POLICY UPDATE ISOLATION: when authz.rego changes, ALL existing nested
#      containers inside DinD must be purged before the new policy takes effect.
#      Any container created under the previous (weaker) policy retains its
#      original HostConfig — --pid=host, --network=host, and permissive seccomp
#      profiles survive a DinD restart if the data volume is not cleared.
#      This script detects authz.rego changes via SHA-256 hash and purges all
#      nested containers automatically when the policy is updated.
#      To force a purge without a policy change: PURGE=yes run-dind

set -e

# Per-run overrides (PURGE=yes / NETWORK=bridge run-dind) must survive the
# config file, so remember them before it is sourced.
_PURGE="${PURGE:-}"
_NETWORK="${NETWORK:-}"

DIND_ENV="${DIND_ENV:-/etc/docker/dind/dind.env}"
if [ -r "${DIND_ENV}" ]; then
    . "${DIND_ENV}"
fi

CONTAINER_NAME="${CONTAINER_NAME:-docker-dind}"
IMAGE="${IMAGE:-docker:dind}"
API_ADDRESS="${API_ADDRESS:-0.0.0.0}"
DOCKER_PORT="${DOCKER_PORT:-3376}"
METRICS_ADDRESS="${METRICS_ADDRESS:-0.0.0.0}"
METRICS_PORT="${METRICS_PORT:-9324}"
SCRIPTS_PATH="${SCRIPTS_PATH:-/etc/docker/dind/scripts}"
CERTS_PATH="${CERTS_PATH:-/etc/docker/dind/certs}"
DATA_PATH="${DATA_PATH:-/var/lib/docker-dind}"
RUN_PATH="${RUN_PATH:-/var/run/docker-dind}"
PIDS_LIMIT="${PIDS_LIMIT:-2048}"
CPU_LIMIT="${CPU_LIMIT:-2}"
MEMORY_LIMIT="${MEMORY_LIMIT:-4G}"
RESTART_POLICY="${RESTART_POLICY:-always}"
LOG_MAX_SIZE="${LOG_MAX_SIZE:-50m}"
LOG_MAX_FILE="${LOG_MAX_FILE:-7}"
# Comma-separated DNS servers for all nested containers; empty = system defaults.
DNS_SERVERS="${DNS_SERVERS:-}"

# Force-purge all nested containers regardless of policy change.
# Set via environment: PURGE=yes run-dind
PURGE="${_PURGE:-${PURGE:-no}}"

# Hash file tracks the authz.rego SHA-256 at the time DinD last started.
# When the hash changes, all nested containers are purged before DinD restarts
# so that no container retains privileges granted by the superseded policy.
AUTHZ_HASH_FILE="${DATA_PATH}/.authz-policy-hash"

# Network mode for the DinD container (and, by extension, for the nested
# containers it runs). Overridable via environment: NETWORK=bridge run-dind
#   host   - DinD shares the host network namespace; the daemon/metrics bind
#            directly on API_ADDRESS/METRICS_ADDRESS; no host port publishing.
#   bridge - DinD gets its own netns; the daemon/metrics listen on 0.0.0.0
#            inside and are published to API_ADDRESS/METRICS_ADDRESS via -p.
NETWORK="${_NETWORK:-${NETWORK:-host}}"

# ---- Authorization plugin (opa-docker-authz, MANAGED plugin in DinD) --------
AUTHZ_PLUGIN_REF="${AUTHZ_PLUGIN_REF:-openpolicyagent/opa-docker-authz-v2:0.9}"
AUTHZ_ALIAS="${AUTHZ_ALIAS:-opa-docker-authz}"
# Host directory containing authz.rego. The WHOLE directory is mounted onto
# DinD's /etc/docker (the managed plugin bind-mounts DinD's /etc/docker as
# /opa). Mounting the directory - not the single file - is required: a
# single-file bind would be a nested mount inside /etc/docker and would NOT
# propagate into the plugin (it would see an empty file -> "empty module").
AUTHZ_POLICY_DIR="${AUTHZ_POLICY_DIR:-/etc/docker/dind/authz}"
AUTHZ_POLICY_HOST="${AUTHZ_POLICY_DIR}/authz.rego"
# Marker (inside the DinD data volume) recording that the plugin is installed.
AUTHZ_MARKER="${DATA_PATH}/.authz-plugin-installed"
DIND_SOCK="unix://${RUN_PATH}/docker.sock"

# (Re)create the DinD container. $1 = "yes" to enable the authz plugin.
start_dind() {
    local use_authz="$1"
    docker rm -f "${CONTAINER_NAME}" >/dev/null 2>&1 || true

    # Network-mode-specific flags: in host mode the daemon/metrics bind directly
    # on the real address and no ports are published; in bridge mode they listen
    # on 0.0.0.0 inside the container and are published via -p.
    local net_args=()
    local api_host metrics_addr
    if [ "${NETWORK}" = "host" ]; then
        net_args=(--network host)
        api_host="tcp://${API_ADDRESS}:${DOCKER_PORT}"
        metrics_addr="${METRICS_ADDRESS}:${METRICS_PORT}"
    else
        net_args=(
            --network bridge
            -p "${API_ADDRESS}:${DOCKER_PORT}:${DOCKER_PORT}"
            -p "${METRICS_ADDRESS}:${METRICS_PORT}:${METRICS_PORT}"
        )
        api_host="tcp://0.0.0.0:${DOCKER_PORT}"
        metrics_addr="0.0.0.0:${METRICS_PORT}"
    fi

    local daemon_args=("--metrics-addr=${metrics_addr}")
    if [ "${use_authz}" = "yes" ]; then
        daemon_args+=("--authorization-plugin=${AUTHZ_ALIAS}")
    fi

    docker run -d \
        "${net_args[@]}" \
        --cap-drop ALL \
        --cap-add SYS_ADMIN \
        --cap-add NET_ADMIN \
        --cap-add NET_RAW \
        --cap-add SETUID \
        --cap-add SETGID \
        --cap-add MKNOD \
        --cap-add FOWNER \
        --cap-add DAC_OVERRIDE \
        --cap-add CHOWN \
        --cap-add AUDIT_WRITE \
        --cap-add KILL \
        --cap-add SYS_CHROOT \
        --cap-add FSETID \
        --cap-add SETFCAP \
        --cap-add SETPCAP \
        --cap-add NET_BIND_SERVICE \
        --device /dev/fuse \
        --device-cgroup-rule 'b 7:* rmw' \
        --security-opt seccomp=unconfined \
        --security-opt apparmor=unconfined \
        --security-opt systempaths=unconfined \
        --cgroupns host \
        -v /sys/fs/cgroup:/sys/fs/cgroup:rw \
        -e DOCKER_HOST=unix:///var/run/dind/docker.sock \
        -e DOCKER_TLS_CERTDIR=/certs \
        -e DOCKER_API_HOST="${api_host}" \
        -e DOCKER_DNS_SERVERS="${DNS_SERVERS}" \
        -v ${SCRIPTS_PATH}/dockerd-entrypoint.sh:/usr/local/bin/dockerd-entrypoint.sh:ro \
        -v ${CERTS_PATH}:/certs \
        -v ${DATA_PATH}:/var/lib/docker \
        -v ${RUN_PATH}:/var/run/dind \
        -v "${AUTHZ_POLICY_DIR}:/etc/docker:ro" \
        --pids-limit ${PIDS_LIMIT} \
        --cpus ${CPU_LIMIT} \
        --memory ${MEMORY_LIMIT} \
        --name ${CONTAINER_NAME} \
        --restart ${RESTART_POLICY} \
        --log-opt max-size=${LOG_MAX_SIZE} \
        --log-opt max-file=${LOG_MAX_FILE} \
        ${IMAGE} \
        "${daemon_args[@]}"
}

# Wait until the DinD daemon answers on its unix socket.
wait_dind() {
    local i
    for i in $(seq 1 30); do
        if docker -H "${DIND_SOCK}" info >/dev/null 2>&1; then
            return 0
        fi
        sleep 1
    done
    return 1
}

# Install the managed authz plugin into the DinD daemon if not already present.
install_plugin_if_missing() {
    if docker -H "${DIND_SOCK}" plugin inspect "${AUTHZ_ALIAS}" >/dev/null 2>&1; then
        echo "authz plugin already installed in dind"
        return 0
    fi
    echo "installing authz plugin into dind (one-time download): ${AUTHZ_PLUGIN_REF}"
    docker -H "${DIND_SOCK}" plugin install \
        --grant-all-permissions \
        --alias "${AUTHZ_ALIAS}" \
        "${AUTHZ_PLUGIN_REF}" \
        opa-args="-policy-file /opa/authz.rego"
}

# Purge ALL containers from the DinD daemon.
# This is called when authz.rego changes to ensure no container created under
# the previous policy survives with superseded privileges (--pid=host,
# seccomp=unconfined, host bind-mounts, etc.).
purge_dind_containers() {
    echo "purging all nested containers from DinD..."
    local ids
    ids=$(docker -H "${DIND_SOCK}" ps -aq 2>/dev/null || true)
    if [ -z "${ids}" ]; then
        echo "no containers to purge"
        return 0
    fi
    # Stop running containers first, then remove all.
    docker -H "${DIND_SOCK}" rm -f ${ids} >/dev/null 2>&1 || true
    # Verify.
    local remaining
    remaining=$(docker -H "${DIND_SOCK}" ps -aq 2>/dev/null | wc -l)
    echo "purge complete: ${remaining} containers remaining"
}

# Compute SHA-256 of the current authz.rego policy.
policy_hash() {
    sha256sum "${AUTHZ_POLICY_HOST}" 2>/dev/null | awk '{print $1}'
}

# Return 0 (true) if the policy has changed since the last recorded hash.
policy_changed() {
    if [ "${PURGE}" = "yes" ]; then
        echo "PURGE=yes: treating policy as changed"
        return 0
    fi
    if [ ! -f "${AUTHZ_HASH_FILE}" ]; then
        return 0
    fi
    local current stored
    current=$(policy_hash)
    stored=$(cat "${AUTHZ_HASH_FILE}" 2>/dev/null || true)
    [ "${current}" != "${stored}" ]
}

# Persist the current policy hash so future runs can detect changes.
record_policy_hash() {
    mkdir -p "$(dirname "${AUTHZ_HASH_FILE}")"
    policy_hash > "${AUTHZ_HASH_FILE}"
}

# First-run bootstrap: start without authz, install plugin, restart with authz.
bootstrap() {
    echo "bootstrapping authz plugin (first run)..."
    start_dind no
    if ! wait_dind; then
        echo "ERROR: dind did not become ready during bootstrap (no-authz phase)" >&2
        docker logs "${CONTAINER_NAME}" >&2 || true
        exit 1
    fi
    install_plugin_if_missing
    touch "${AUTHZ_MARKER}"
    echo "enabling authz and restarting dind..."
    start_dind yes
    if ! wait_dind; then
        echo "ERROR: dind did not become ready with authz enabled" >&2
        docker logs "${CONTAINER_NAME}" >&2 || true
        exit 1
    fi
}

echo "=========================================="
echo "Docker-in-Docker container management"
echo "=========================================="

if [ ! -f "${AUTHZ_POLICY_HOST}" ]; then
    echo "ERROR: authz policy not found: ${AUTHZ_POLICY_HOST}" >&2
    echo "Deploy authz.rego there before running." >&2
    exit 1
fi

# If already running, do nothing (recreate manually with: docker rm -f docker-dind).
if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "container ${CONTAINER_NAME} is already running"
    echo "to recreate, first remove: docker rm -f ${CONTAINER_NAME}"
    exit 0
fi

# Drop a stopped leftover container so we start clean.
docker rm "${CONTAINER_NAME}" >/dev/null 2>&1 || true

# Policy-change purge: if authz.rego has changed since the last recorded hash
# (or PURGE=yes is set), start DinD temporarily without authz, destroy ALL
# nested containers, then proceed. This prevents containers created under the
# previous (weaker) policy from surviving with superseded privileges.
if [ -f "${AUTHZ_MARKER}" ] && policy_changed; then
    echo "authz policy has changed — purging all nested containers before restart..."
    start_dind no
    if wait_dind; then
        purge_dind_containers
    else
        echo "WARNING: dind did not start for purge pass; containers may survive" >&2
    fi
    docker rm -f "${CONTAINER_NAME}" >/dev/null 2>&1 || true
fi

if [ -f "${AUTHZ_MARKER}" ]; then
    echo "authz plugin previously installed; starting dind with authz..."
    start_dind yes
    if ! wait_dind; then
        echo "dind failed to start with authz (plugin missing from volume?); re-bootstrapping..." >&2
        rm -f "${AUTHZ_MARKER}"
        bootstrap
    fi
else
    bootstrap
fi

# Record the policy hash so the next run can detect changes.
record_policy_hash

echo ""
echo "=========================================="
echo "DinD container created successfully"
echo "=========================================="
echo "Container name: ${CONTAINER_NAME}"
echo "Network mode:   ${NETWORK}"
echo "Authz plugin:   ${AUTHZ_ALIAS} (${AUTHZ_PLUGIN_REF})"
echo "API address: tcp://${API_ADDRESS}:${DOCKER_PORT}"
echo "Metrics address: http://${METRICS_ADDRESS}:${METRICS_PORT}/metrics"
echo "Unix socket: ${RUN_PATH}/docker.sock"
echo "Client certificates: ${CERTS_PATH}/client/"
echo ""
echo "useful commands:"
echo "  # view logs"
echo "  docker logs ${CONTAINER_NAME}"
echo "  # restart container"
echo "  docker restart ${CONTAINER_NAME}"
echo "  # remove container"
echo "  docker rm -f ${CONTAINER_NAME}"
echo "  # recreate container"
echo "  run-dind"
echo "  # use via unix socket"
echo "  docker-dind-sock ps"
echo "  # use via TLS"
echo "  docker-dind-tls ps"
echo ""
