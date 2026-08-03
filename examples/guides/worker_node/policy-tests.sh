#!/usr/bin/env bash
# =============================================================================
# DinD policy test-suite (OPA authz + hardened seccomp).
#
# WHERE TO RUN:
#   Inside a PentAGI worker container (a child of the *host* Docker daemon) whose
#   `docker` CLI is wired to the hardened *DinD* daemon. Every `docker run` below
#   therefore creates a NESTED container subject to:
#     - the OPA authorization plugin (authz.rego)  — gates containers/create + exec
#     - the default seccomp profile (seccomp.json) — kernel-enforced floor
#     - the per-container device cgroup + capability bounding set
#
# WHAT IT CHECKS:
#   POSITIVE tests  — legitimate pentest/app workloads MUST work (no false positives):
#       nmap / masscan / curl+openssl / nginx+apache / java / python+sqlmap /
#       volumes / tmpfs / --network=host / DNS / raw sockets / mkfifo /
#       char-device mknod / docker pull / container lifecycle (start/stop/restart/
#       pause/unpause/kill/logs/inspect/top/stats/rename) / image management
#       (list/tag/rmi) / safe volume create / bridge network create /
#       runtime features (--init/--user/--env/--label/-p/--workdir/--entrypoint) /
#       compose-style multi-container workflow with shared volumes.
#   NEGATIVE tests  — host-escape primitives MUST be blocked:
#       --privileged, dangerous --cap-add, --device, --security-opt, host
#       namespaces, host bind-mounts, privileged exec, mknod block-device, mount /
#       images/load, plugin install, swarm init, sysctl injection,
#       DeviceRequests, custom runtime, volume bind-emulation, macvlan/ipvlan.
#
# USAGE:
#   ./policy-tests.sh              # run everything
#   DOCKER="docker -H tcp://10.100.1.3:2376 --tls" ./policy-tests.sh
#   TARGET_IP=10.100.1.3 ./policy-tests.sh
#
# EXIT CODE: non-zero if any test FAILED (a failed NEGATIVE test = security gap,
#            a failed POSITIVE test = false positive that would break agents).
# =============================================================================

set -uo pipefail

# ---- configuration ---------------------------------------------------------
# DOCKER may be a plain binary name ("docker") or a multi-word invocation
# ("docker --tls --tlscacert … -H tcp://…"). Split into an array so that
# "${DOCKER[@]}" expands each flag as a separate word when used as a command.
# shellcheck disable=SC2206
DOCKER=(${DOCKER:-docker})
TARGET_IP="${TARGET_IP:-10.100.1.3}"     # host reachable from nested containers
HTTP_PORT="${HTTP_PORT:-8080}"           # plaintext HTTP service on the host
DIND_TLS_PORT="${DIND_TLS_PORT:-2376}"   # DinD daemon mTLS API
OUTER_TLS_PORT="${OUTER_TLS_PORT:-3376}" # outer daemon mTLS API
SSH_PORT="${SSH_PORT:-22}"

# Networking for tests that must reach the HOST service IP ($TARGET_IP).
# Nested containers on the DinD *bridge* network are NAT'd and frequently CANNOT
# reach the host's own service IP (interface binding / hairpin), whereas the
# host network namespace can — and the OPA policy intentionally ALLOWS
# --network=host. This is also how PentAGI agents reach host-local targets.
# Override with TARGET_NET="" to force the default bridge.
TARGET_NET="${TARGET_NET:---network host}"

# Images (override via env if your registry differs). Pinned where it matters.
IMG_ALPINE="${IMG_ALPINE:-alpine:3.20}"
IMG_CURL="${IMG_CURL:-curlimages/curl:8.10.1}"
IMG_NGINX="${IMG_NGINX:-nginx:1.27-alpine}"
IMG_HTTPD="${IMG_HTTPD:-httpd:2.4-alpine}"
IMG_NMAP="${IMG_NMAP:-instrumentisto/nmap:7.95}"
IMG_JAVA="${IMG_JAVA:-eclipse-temurin:21-jdk}"
IMG_PY="${IMG_PY:-python:3.12-alpine}"

LABEL="pentagi-dind-test"
NETPREFIX="pgdtnet"
VOLPREFIX="pgdtvol"

# opa-docker-authz / daemon denial signatures (CLI surfaces the policy message).
DENY_RE='authorization denied|opa-docker-authz|administrative policy|not allowed|bind-mount not allowed|capability not allowed|privileged containers|namespace is not allowed|security options are not allowed|device passthrough|device cgroup|received 403|access denied'

# ---- presentation ----------------------------------------------------------
if [ -t 1 ]; then
  GRN=$'\e[32m'; RED=$'\e[31m'; YLW=$'\e[33m'; BLU=$'\e[36m'; BLD=$'\e[1m'; RST=$'\e[0m'
else
  GRN=""; RED=""; YLW=""; BLU=""; BLD=""; RST=""
fi

TOTAL=0; PASS=0; FAIL=0; SKIP=0; INFO=0
declare -a FAILED_NAMES=()

section(){ printf "\n${BLU}${BLD}=== %s ===${RST}\n" "$1"; }

record(){ # status name [detail]
  local st="$1" name="$2" detail="${3:-}"
  case "$st" in
    PASS) TOTAL=$((TOTAL+1)); PASS=$((PASS+1)); printf "  ${GRN}[ PASS ]${RST} %s\n" "$name" ;;
    FAIL) TOTAL=$((TOTAL+1)); FAIL=$((FAIL+1)); FAILED_NAMES+=("$name")
          printf "  ${RED}[ FAIL ]${RST} %s\n" "$name"
          [ -n "$detail" ] && printf "           ${RED}↳ %s${RST}\n" "$detail" ;;
    SKIP) TOTAL=$((TOTAL+1)); SKIP=$((SKIP+1)); printf "  ${YLW}[ SKIP ]${RST} %s\n" "$name"
          [ -n "$detail" ] && printf "           ${YLW}↳ %s${RST}\n" "$detail" ;;
    INFO) INFO=$((INFO+1)); printf "  ${BLU}[ INFO ]${RST} %s\n" "$name"
          [ -n "$detail" ] && printf "           ${BLU}↳ %s${RST}\n" "$detail" ;;
  esac
}

LAST_OUT=""; LAST_RC=0
cap(){ LAST_OUT="$("$@" 2>&1)"; LAST_RC=$?; }
tailout(){ echo "$LAST_OUT" | tail -n 2 | tr '\n' ' ' | cut -c1-300; }

# Positive: arbitrary docker subcommand must succeed (not docker run).
expect_cmd_ok(){ local name="$1"; shift
  cap "${DOCKER[@]}" "$@"
  if [ "$LAST_RC" -eq 0 ]; then record PASS "$name"
  else record FAIL "$name" "rc=$LAST_RC: $(tailout)"; fi
}

# Negative: arbitrary docker subcommand must be authz-denied.
expect_cmd_deny(){ local name="$1"; shift
  cap "${DOCKER[@]}" "$@"
  if [ "$LAST_RC" -ne 0 ] && echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
    record PASS "$name"
  elif [ "$LAST_RC" -ne 0 ]; then
    record FAIL "$name" "failed, but NOT via authz denial: $(tailout)"
  else
    record FAIL "$name" "SECURITY GAP: request was ALLOWED (expected denial)"
  fi
}

# Direct HTTP call to the DinD API for endpoints not exposed by the Docker CLI.
# Transport resolution order:
#   1. DOCKER=(docker -H tcp://… --tlscacert …) CLI flags
#   2. DOCKER_HOST / DOCKER_CERT_PATH env (PentAGI worker / guide layout)
#   3. unix socket /var/run/docker.sock
# Always prefixes the Docker API version (required by Engine 29+).
dind_api_post(){ # path json
    local path="$1" body="$2" d="${DOCKER[*]}"
    local endpoint="" curl_args=() api_ver

    if [[ "$d" =~ -H[[:space:]]+tcp://([^[:space:]]+) ]]; then
        endpoint="https://${BASH_REMATCH[1]}"
        [[ "$d" =~ --tlscacert[[:space:]]+([^[:space:]]+) ]] && curl_args+=(--cacert "${BASH_REMATCH[1]}")
        [[ "$d" =~ --tlscert[[:space:]]+([^[:space:]]+) ]]   && curl_args+=(--cert  "${BASH_REMATCH[1]}")
        [[ "$d" =~ --tlskey[[:space:]]+([^[:space:]]+) ]]    && curl_args+=(--key   "${BASH_REMATCH[1]}")
    elif [[ "${DOCKER_HOST:-}" =~ ^tcp://([^[:space:]]+)$ ]]; then
        endpoint="https://${BASH_REMATCH[1]}"
        local cert_path="${DOCKER_CERT_PATH:-}"
        if [ -n "$cert_path" ]; then
            [ -f "$cert_path/ca.pem" ]   && curl_args+=(--cacert "$cert_path/ca.pem")
            [ -f "$cert_path/cert.pem" ] && curl_args+=(--cert  "$cert_path/cert.pem")
            [ -f "$cert_path/key.pem" ]  && curl_args+=(--key   "$cert_path/key.pem")
        fi
    else
        curl_args+=(--unix-socket /var/run/docker.sock)
        endpoint="http://localhost"
    fi

    api_ver=$("${DOCKER[@]}" version --format '{{.Server.APIVersion}}' 2>/dev/null || true)
    api_ver="${api_ver:-1.44}"
    [[ "$path" != /v* ]] && path="/v${api_ver}${path}"

    LAST_OUT=$(curl -sS "${curl_args[@]}" -X POST "${endpoint}${path}" \
        -H 'Content-Type: application/json' -d "$body" 2>&1)
    LAST_RC=$?
}

# ---- image availability (lazy, cached) -------------------------------------
declare -A IMG_STATE
need_image(){
  local img="$1"
  case "${IMG_STATE[$img]:-}" in
    ok) return 0 ;; no) return 1 ;;
  esac
  if "${DOCKER[@]}" image inspect "$img" >/dev/null 2>&1 || "${DOCKER[@]}" pull "$img" >/dev/null 2>&1; then
    IMG_STATE[$img]=ok; return 0
  fi
  IMG_STATE[$img]=no; return 1
}

# ---- generic assertions ----------------------------------------------------
# Negative: container creation must be rejected by the authz plugin.
expect_deny(){ local name="$1"; shift
  cap "${DOCKER[@]}" run --rm --label "$LABEL" "$@"
  if [ "$LAST_RC" -ne 0 ] && echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
    record PASS "$name"
  elif [ "$LAST_RC" -ne 0 ]; then
    record FAIL "$name" "failed, but NOT via authz denial: $(tailout)"
  else
    record FAIL "$name" "SECURITY GAP: request was ALLOWED (expected denial)"
  fi
}

# Positive: container must run and exit 0.
expect_run_ok(){ local name="$1"; shift
  cap "${DOCKER[@]}" run --rm --label "$LABEL" "$@"
  if [ "$LAST_RC" -eq 0 ]; then record PASS "$name"
  else record FAIL "$name" "exit=$LAST_RC: $(tailout)"; fi
}

# Positive: container runs and stdout matches a regex.
expect_contains(){ local name="$1" pat="$2"; shift 2
  cap "${DOCKER[@]}" run --rm --label "$LABEL" "$@"
  if [ "$LAST_RC" -eq 0 ] && echo "$LAST_OUT" | grep -qiE "$pat"; then
    record PASS "$name"
  else
    record FAIL "$name" "rc=$LAST_RC, want /$pat/, got: $(tailout)"
  fi
}

# Negative (in-container): container STARTS but the inner syscall is refused.
expect_inner_fail(){ local name="$1" pat="$2"; shift 2
  cap "${DOCKER[@]}" run --rm --label "$LABEL" "$@"
  if [ "$LAST_RC" -ne 0 ] && echo "$LAST_OUT" | grep -qiE "$pat"; then
    record PASS "$name"
  else
    record FAIL "$name" "rc=$LAST_RC (expected refusal /$pat/): $(tailout)"
  fi
}

# Positive tool test that installs a dependency at runtime (apk/pip). The inner
# command MUST `exit 200` when the dependency install fails, so we can SKIP
# (no package mirror egress) instead of reporting a false FAIL.
expect_tool(){ local name="$1" pat="$2"; shift 2
  cap "${DOCKER[@]}" run --rm --label "$LABEL" "$@"
  if [ "$LAST_RC" -eq 200 ]; then
    record SKIP "$name" "runtime dependency install failed (no package egress?)"
  elif [ "$LAST_RC" -eq 0 ] && echo "$LAST_OUT" | grep -qiE "$pat"; then
    record PASS "$name"
  else
    record FAIL "$name" "rc=$LAST_RC, want /$pat/, got: $(tailout)"
  fi
}

# ---- cleanup ---------------------------------------------------------------
cleanup(){
  local ids nets vols
  ids="$("${DOCKER[@]}" ps -aq --filter "label=$LABEL" 2>/dev/null)"
  [ -n "$ids" ] && "${DOCKER[@]}" rm -f $ids >/dev/null 2>&1
  nets="$("${DOCKER[@]}" network ls -q --filter "name=$NETPREFIX" 2>/dev/null)"
  [ -n "$nets" ] && "${DOCKER[@]}" network rm $nets >/dev/null 2>&1
  vols="$("${DOCKER[@]}" volume ls -q --filter "name=$VOLPREFIX" 2>/dev/null)"
  [ -n "$vols" ] && "${DOCKER[@]}" volume rm $vols >/dev/null 2>&1
}
trap cleanup EXIT

# =============================================================================
# PREFLIGHT
# =============================================================================
section "Preflight"
cap "${DOCKER[@]}" version --format '{{.Server.Version}}'
if [ "$LAST_RC" -ne 0 ]; then
  printf "  ${RED}Cannot reach the Docker daemon via '%s'. Aborting.${RST}\n" "${DOCKER[*]}"
  exit 2
fi
printf "  Docker server version: %s\n" "$LAST_OUT"
printf "  Target host: %s (http:%s, dind-tls:%s, outer-tls:%s, ssh:%s)\n" \
  "$TARGET_IP" "$HTTP_PORT" "$DIND_TLS_PORT" "$OUTER_TLS_PORT" "$SSH_PORT"

if ! need_image "$IMG_ALPINE"; then
  printf "  ${RED}Base image %s unavailable; cannot run the suite.${RST}\n" "$IMG_ALPINE"
  exit 2
fi
# Warm the common images (best effort).
for i in "$IMG_CURL" "$IMG_NMAP"; do need_image "$i" >/dev/null 2>&1 || true; done

# Best-effort egress probe (informational only — package-install tests below
# self-skip via the `exit 200` sentinel, so a wrong guess here costs nothing).
# Probe several well-known hosts because some environments allow the registry
# but block arbitrary egress (which made a single example.com probe misleading).
HAVE_EGRESS=0
cap "${DOCKER[@]}" run --rm --label "$LABEL" "$IMG_ALPINE" sh -c '
  for h in dl-cdn.alpinelinux.org pypi.org 1.1.1.1; do
    if nc -w 3 "$h" 443 </dev/null >/dev/null 2>&1; then echo EGRESS_OK; break; fi
  done'
echo "$LAST_OUT" | grep -q EGRESS_OK && HAVE_EGRESS=1
printf "  Outbound internet egress (best-effort): %s\n" "$([ $HAVE_EGRESS -eq 1 ] && echo yes || echo "no/filtered")"
printf "  Host-target network mode: %s\n" "${TARGET_NET:-<default bridge>}"

# =============================================================================
# RUNTIME COMPONENT VERSION AUDIT
# Extracts component versions from the DinD daemon and checks them against
# known CVE thresholds. A failed version check records FAIL (security gap)
# because the component is vulnerable even if current defense-in-depth
# (RO /proc/sys, OPA, capability restrictions) mitigates active exploitation.
# =============================================================================
section "Runtime component version audit"

RUNC_VER=$("${DOCKER[@]}" version --format \
  '{{range .Server.Components}}{{if eq .Name "runc"}}{{.Version}}{{end}}{{end}}' 2>/dev/null | tr -d '\n')
CTR_VER=$("${DOCKER[@]}" version --format \
  '{{range .Server.Components}}{{if eq .Name "containerd"}}{{.Version}}{{end}}{{end}}' 2>/dev/null | tr -d '\n')
ENG_VER=$("${DOCKER[@]}" version --format '{{.Server.Version}}' 2>/dev/null | tr -d '\n')

record INFO "Docker Engine: ${ENG_VER:-unknown}"
record INFO "containerd:    ${CTR_VER:-unknown}"
record INFO "runc:          ${RUNC_VER:-unknown}"

# Helper: returns 0 (true) when $1 >= $2 (both semver strings).
# Requires GNU sort -V; busybox sort falls back gracefully to SKIP.
_ver_ge(){
  local hi
  hi=$(printf '%s\n%s\n' "$2" "$1" | sort -V 2>/dev/null | tail -1)
  [ "$hi" = "$1" ]
}

# ---- CVE-2024-21626 "Leaky Vessels" (runc working-dir fd leak).
#      Exploitable WITHOUT --privileged on runc < 1.1.12.
LEAKY_MIN="1.1.12"
if [ -z "${RUNC_VER}" ]; then
  record SKIP "runc CVE-2024-21626 (Leaky Vessels)" "could not determine runc version"
elif ! command -v sort >/dev/null 2>&1 || ! printf '%s\n' a b | sort -V >/dev/null 2>&1; then
  record SKIP "runc CVE-2024-21626 (Leaky Vessels)" "sort -V not available for version comparison"
elif _ver_ge "${RUNC_VER}" "${LEAKY_MIN}"; then
  record PASS "runc ${RUNC_VER} >= ${LEAKY_MIN} — CVE-2024-21626 Leaky Vessels patched"
else
  record FAIL "runc ${RUNC_VER} VULNERABLE — CVE-2024-21626 Leaky Vessels (fd leak, no --privileged needed)" \
    "Upgrade DinD image to bundle runc >= ${LEAKY_MIN}."
fi

# ---- CVE-2025-31133/31134/31135 (masked-paths symlink bypass, /dev/console mount
#      race, procfs write redirect). All fixed in runc >= 1.2.8 / >= 1.3.3.
#      Current mitigation: RO /proc/sys mount blocks core_pattern write.
RUNC_SAFE_MIN="1.2.8"
if [ -z "${RUNC_VER}" ]; then
  record SKIP "runc CVE-2025-31133/31134/31135 check" "could not determine runc version"
elif ! printf '%s\n' a b | sort -V >/dev/null 2>&1; then
  record SKIP "runc CVE-2025-31133/31134/31135 check" "sort -V not available"
elif _ver_ge "${RUNC_VER}" "${RUNC_SAFE_MIN}"; then
  record PASS "runc ${RUNC_VER} >= ${RUNC_SAFE_MIN} — CVE-2025-31133/31134/31135 patched"
else
  record FAIL "runc ${RUNC_VER} VULNERABLE — CVE-2025-31133/31134/31135 (procfs write via symlink)" \
    "Upgrade DinD image to runc >= ${RUNC_SAFE_MIN}. Mitigation: RO /proc/sys (must stay RO)."
fi

# =============================================================================
# POSITIVE — baseline & allowed HostConfig features (no false positives)
# =============================================================================
section "POSITIVE — baseline & allowed features"

expect_contains "basic container runs and prints output" "hello-dind" \
  "$IMG_ALPINE" echo hello-dind

expect_run_ok "--network=host is allowed (intentional, for DNS/raw sockets)" \
  --network host "$IMG_ALPINE" true

expect_run_ok "--cap-add NET_ADMIN allowed (VPN/iface config tooling)" \
  --cap-add NET_ADMIN "$IMG_ALPINE" true

expect_run_ok "--cap-add NET_RAW allowed (raw-socket scanners)" \
  --cap-add NET_RAW "$IMG_ALPINE" true

expect_run_ok "--read-only rootfs + --tmpfs writable" \
  --read-only --tmpfs /tmp:rw "$IMG_ALPINE" sh -c 'echo ok > /tmp/x && cat /tmp/x >/dev/null'

expect_run_ok "resource limits (--memory/--cpus/--pids-limit) accepted" \
  --memory 256m --cpus 1 --pids-limit 256 "$IMG_ALPINE" true

# =============================================================================
# POSITIVE — storage (named volume, anonymous volume, tmpfs)
# =============================================================================
section "POSITIVE — volumes & tmpfs"

NAMED_VOL="${VOLPREFIX}_named_$RANDOM"
TOKEN="tok_$RANDOM"
cap "${DOCKER[@]}" run --rm --label "$LABEL" -v "$NAMED_VOL:/data" "$IMG_ALPINE" sh -c "echo $TOKEN > /data/f"
if [ "$LAST_RC" -eq 0 ]; then
  cap "${DOCKER[@]}" run --rm --label "$LABEL" -v "$NAMED_VOL:/data" "$IMG_ALPINE" cat /data/f
  if [ "$LAST_RC" -eq 0 ] && [ "$LAST_OUT" = "$TOKEN" ]; then
    record PASS "named volume persists & is shared across containers"
  else
    record FAIL "named volume read-back" "got: $(tailout)"
  fi
else
  record FAIL "named volume write" "$(tailout)"
fi
"${DOCKER[@]}" volume rm "$NAMED_VOL" >/dev/null 2>&1

expect_run_ok "anonymous volume mounts and is writable" \
  -v /data "$IMG_ALPINE" sh -c 'echo ok > /data/f && cat /data/f >/dev/null'

expect_run_ok "tmpfs mount is writable and is really a tmpfs" \
  --tmpfs /scratch:rw,size=16m "$IMG_ALPINE" \
  sh -c 'echo ok > /scratch/f && grep -qE "tmpfs[[:space:]]+/scratch" /proc/mounts'

# =============================================================================
# POSITIVE — syscalls that must keep working under the hardened seccomp profile
# =============================================================================
section "POSITIVE — allowed syscalls (mkfifo / char device)"

expect_run_ok "mkfifo / mknod(S_IFIFO) works (named pipes)" \
  "$IMG_ALPINE" sh -c 'mkfifo /tmp/fifo && [ -p /tmp/fifo ]'

expect_run_ok "mknod char device works (e.g. /dev/net/tun-style nodes)" \
  "$IMG_ALPINE" sh -c 'mknod /tmp/zero c 1 5 && [ -c /tmp/zero ]'

expect_run_ok "clone()-based threading works (multi-threaded process)" \
  "$IMG_ALPINE" sh -c 'for i in $(seq 1 8); do sleep 0.1 & done; wait'

# =============================================================================
# POSITIVE — networking tooling (curl, DNS, ssh banner)
#
# NOTE: tests that reach the HOST service IP use $TARGET_NET (--network host by
# default). Nested containers on the DinD bridge are NAT'd and usually cannot
# reach the host's own service IP; host networking is policy-allowed and is the
# path agents use. We still record an INFO line documenting bridge reachability.
# =============================================================================
section "POSITIVE — networking & curl"

# Informational: can the DEFAULT bridge reach the host service IP at all?
cap "${DOCKER[@]}" run --rm --label "$LABEL" "$IMG_ALPINE" sh -c "nc -w 4 $TARGET_IP $HTTP_PORT </dev/null >/dev/null 2>&1"
if [ "$LAST_RC" -eq 0 ]; then
  record INFO "default-bridge container CAN reach $TARGET_IP:$HTTP_PORT (TCP)"
else
  record INFO "default-bridge container canNOT reach $TARGET_IP:$HTTP_PORT — host targets use --network host (expected)"
fi

if need_image "$IMG_CURL"; then
  cap "${DOCKER[@]}" run --rm --label "$LABEL" $TARGET_NET "$IMG_CURL" -sS -m 10 -o /dev/null -w '%{http_code}' "http://$TARGET_IP:$HTTP_PORT/"
  if [ "$LAST_RC" -eq 0 ] && echo "$LAST_OUT" | grep -qE '^[1-5][0-9][0-9]$'; then
    record PASS "curl reaches $TARGET_IP:$HTTP_PORT and gets an HTTP response (code $LAST_OUT)"
  else
    record FAIL "curl http://$TARGET_IP:$HTTP_PORT" "rc=$LAST_RC out=$(tailout)"
  fi
else
  record SKIP "curl HTTP test" "image $IMG_CURL unavailable"
fi

expect_contains "DNS resolution via DinD resolver works" "address|Name:|canonical" \
  "$IMG_ALPINE" nslookup example.com

expect_contains "SSH banner readable from $TARGET_IP:$SSH_PORT" "SSH-" \
  $TARGET_NET "$IMG_ALPINE" sh -c "echo | nc -w 5 $TARGET_IP $SSH_PORT"

# =============================================================================
# POSITIVE — TLS handshake + openssl (mTLS endpoints expose a server cert)
# =============================================================================
section "POSITIVE — TLS / openssl"

# nmap ssl-cert: no extra packages needed, retrieves the server certificate.
if need_image "$IMG_NMAP"; then
  expect_contains "nmap ssl-cert retrieves DinD TLS cert ($TARGET_IP:$DIND_TLS_PORT)" \
    "ssl-cert|Subject:|commonName|Issuer:|Public Key" \
    $TARGET_NET "$IMG_NMAP" -Pn -p"$DIND_TLS_PORT" --script ssl-cert "$TARGET_IP"
else
  record SKIP "nmap ssl-cert" "image $IMG_NMAP unavailable"
fi

# openssl s_client: the server certificate is sent even when client-auth is
# required, so x509 subject extraction proves curl/openssl crypto works.
expect_tool "openssl s_client handshake + x509 parse ($TARGET_IP:$DIND_TLS_PORT)" \
  "subject|CN ?=|commonName" \
  $TARGET_NET "$IMG_ALPINE" \
  sh -c "apk add --no-cache openssl >/dev/null 2>&1 || exit 200; echo | openssl s_client -connect $TARGET_IP:$DIND_TLS_PORT 2>/dev/null | openssl x509 -noout -subject 2>/dev/null"

# =============================================================================
# POSITIVE — scanners (nmap connect/SYN/version, masscan): NET_RAW path
# =============================================================================
section "POSITIVE — port scanners (nmap / masscan)"

if need_image "$IMG_NMAP"; then
  expect_contains "nmap TCP connect scan finds open ports on $TARGET_IP" \
    "$HTTP_PORT/tcp +open|$SSH_PORT/tcp +open|$DIND_TLS_PORT/tcp +open" \
    $TARGET_NET "$IMG_NMAP" -Pn -sT -p"$SSH_PORT,$HTTP_PORT,$DIND_TLS_PORT,$OUTER_TLS_PORT" "$TARGET_IP"

  expect_contains "nmap SYN scan works (raw sockets / CAP_NET_RAW)" \
    "$HTTP_PORT/tcp +open" \
    $TARGET_NET "$IMG_NMAP" -Pn -sS -p"$HTTP_PORT" "$TARGET_IP"

  expect_contains "nmap service/version detection (-sV)" \
    "$HTTP_PORT/tcp +open|service|VERSION|http" \
    $TARGET_NET "$IMG_NMAP" -Pn -sV --version-light -p"$HTTP_PORT" "$TARGET_IP"
else
  record SKIP "nmap scans" "image $IMG_NMAP unavailable"
fi

# masscan ships in Alpine's community repo; uses its own raw-packet stack.
# NOTE: scanning the host's OWN IP can miss the SYN-ACK (raw sniffer doesn't see
# loopback traffic), so a clean run of masscan's engine (raw socket opened, scan
# started) already proves the policy/seccomp allow it. We accept either an open
# port discovery OR masscan's normal startup output; only a raw-socket/seccomp
# denial (no such output) fails the test.
expect_tool "masscan runs with raw sockets against $TARGET_IP" \
  "Discovered open port|open port $HTTP_PORT|Starting masscan|Scanning 1 host|rate:" \
  $TARGET_NET --cap-add NET_RAW "$IMG_ALPINE" \
  sh -c '
    apk add --no-cache masscan libpcap libpcap-dev >/dev/null 2>&1 || exit 200
    # masscan dlopen()s libpcap; some builds look for the unversioned soname,
    # which the runtime package does not ship. Provide it if missing.
    if [ ! -e /usr/lib/libpcap.so ]; then
      so="$(ls /usr/lib/libpcap.so.* 2>/dev/null | head -n1)"
      [ -n "$so" ] && ln -sf "$so" /usr/lib/libpcap.so 2>/dev/null || true
    fi
    masscan -p'"$HTTP_PORT"' '"$TARGET_IP"' --rate 2000 --wait 2 2>&1
    true'

# =============================================================================
# POSITIVE — real services: nginx & apache serve content (container-to-container)
# Uses a user-defined bridge network + busybox wget (no extra image dependency).
# =============================================================================
section "POSITIVE — web servers (nginx / apache) + content check"

web_content_test(){ # name image expected-content-regex
  local name="$1" image="$2" pat="$3"
  if ! need_image "$image"; then record SKIP "$name" "image $image unavailable"; return; fi
  local net="${NETPREFIX}$RANDOM" srv="${LABEL}-srv-$RANDOM" i out ok=1
  "${DOCKER[@]}" network create "$net" >/dev/null 2>&1
  if ! "${DOCKER[@]}" run -d --rm --label "$LABEL" --name "$srv" --network "$net" "$image" >/dev/null 2>&1; then
    record FAIL "$name" "failed to start server container"
    "${DOCKER[@]}" network rm "$net" >/dev/null 2>&1; return
  fi
  for i in $(seq 1 25); do
    out="$("${DOCKER[@]}" run --rm --label "$LABEL" --network "$net" "$IMG_ALPINE" \
            wget -q -T 3 -O - "http://$srv/" 2>/dev/null)"
    if echo "$out" | grep -qiE "$pat"; then ok=0; break; fi
    sleep 1
  done
  if [ "$ok" -eq 0 ]; then record PASS "$name"
  else record FAIL "$name" "expected /$pat/ in served content (server not reachable/ready)"; fi
  "${DOCKER[@]}" rm -f "$srv" >/dev/null 2>&1
  "${DOCKER[@]}" network rm "$net" >/dev/null 2>&1
}

web_content_test "nginx serves content (wget over user bridge network)" "$IMG_NGINX" "Welcome to nginx"
web_content_test "apache/httpd serves content (wget over user bridge network)" "$IMG_HTTPD" "It works"

# =============================================================================
# POSITIVE — JVM workload (heavy syscall surface) + outbound TCP from Java
# =============================================================================
section "POSITIVE — Java application"

if need_image "$IMG_JAVA"; then
  JPROG='public class T { public static void main(String[] a) throws Exception {'
  JPROG+=' System.out.println("JAVA_OK");'
  JPROG+=' try (java.net.Socket s=new java.net.Socket()) {'
  JPROG+=' s.connect(new java.net.InetSocketAddress("__IP__", __PORT__), 5000);'
  JPROG+=' System.out.println("SOCKET_CONNECTED"); } } }'
  JPROG="${JPROG/__IP__/$TARGET_IP}"
  JPROG="${JPROG/__PORT__/$HTTP_PORT}"
  expect_contains "JVM compiles+runs and opens an outbound socket" \
    "JAVA_OK.*SOCKET_CONNECTED|SOCKET_CONNECTED" \
    $TARGET_NET -e "PROG=$JPROG" "$IMG_JAVA" \
    bash -c 'printf "%s" "$PROG" > /tmp/T.java && cd /tmp && javac T.java && java T'
else
  record SKIP "Java application test" "image $IMG_JAVA unavailable"
fi

# =============================================================================
# POSITIVE — Python stack (interpreter / stdlib networking / sqlmap / requests)
# =============================================================================
section "POSITIVE — Python stack (sqlmap & friends)"

if need_image "$IMG_PY"; then
  # Interpreter + stdlib (no network, no pip) — must always work under seccomp.
  expect_contains "python interpreter + stdlib (hashlib/json) works" "PY_OK" \
    "$IMG_PY" python -c 'import hashlib,json,base64,ssl,socket; print("PY_OK", hashlib.sha256(b"x").hexdigest()[:8])'

  # Stdlib HTTP client to the host service (no third-party deps).
  PYHTTP='import http.client as h; c=h.HTTPConnection("__IP__",__PORT__,timeout=8); c.request("GET","/"); r=c.getresponse(); print("PYHTTP",r.status)'
  PYHTTP="${PYHTTP/__IP__/$TARGET_IP}"; PYHTTP="${PYHTTP/__PORT__/$HTTP_PORT}"
  expect_contains "python stdlib HTTP GET to $TARGET_IP:$HTTP_PORT" "PYHTTP [1-5][0-9][0-9]" \
    $TARGET_NET "$IMG_PY" python -c "$PYHTTP"

  # sqlmap (real pentest tool) installs via pip and runs against the target.
  expect_tool "sqlmap (pip) installs and runs against $TARGET_IP" "^[0-9]+\\.[0-9]+|sqlmap/[0-9]" \
    $TARGET_NET "$IMG_PY" sh -c '
      pip install --no-cache-dir --quiet sqlmap >/dev/null 2>&1 || exit 200
      command -v sqlmap >/dev/null 2>&1 || exit 200
      sqlmap --version 2>&1
      sqlmap -u "http://'"$TARGET_IP"':'"$HTTP_PORT"'/?id=1" --batch --flush-session \
             --level=1 --risk=1 --technique=B --timeout=8 --retries=1 -v1 2>&1 | tail -n 8 || true'

  # Third-party pip package (requests) + HTTP — confirms pip wheels work.
  RQ='import requests; r=requests.get("http://__IP__:__PORT__/",timeout=8,allow_redirects=False); print("REQUESTS",r.status_code)'
  RQ="${RQ/__IP__/$TARGET_IP}"; RQ="${RQ/__PORT__/$HTTP_PORT}"
  expect_tool "python requests (pip) HTTP GET to $TARGET_IP:$HTTP_PORT" "REQUESTS [1-5][0-9][0-9]" \
    $TARGET_NET "$IMG_PY" sh -c "pip install --no-cache-dir --quiet requests >/dev/null 2>&1 || exit 200; python -c '$RQ'"
else
  record SKIP "Python stack tests" "image $IMG_PY unavailable"
fi

# =============================================================================
# POSITIVE — exec into a running container (normal, non-privileged)
# =============================================================================
section "POSITIVE / NEGATIVE — docker exec"

EXEC_C="${LABEL}-exec-$RANDOM"
if "${DOCKER[@]}" run -d --rm --label "$LABEL" --name "$EXEC_C" "$IMG_ALPINE" sleep 120 >/dev/null 2>&1; then
  cap "${DOCKER[@]}" exec "$EXEC_C" id
  if [ "$LAST_RC" -eq 0 ] && echo "$LAST_OUT" | grep -q "uid="; then
    record PASS "normal docker exec works"
  else
    record FAIL "normal docker exec" "rc=$LAST_RC out=$(tailout)"
  fi
  # NEGATIVE: privileged exec must be blocked (authz exec_violation).
  cap "${DOCKER[@]}" exec --privileged "$EXEC_C" id
  if [ "$LAST_RC" -ne 0 ] && echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
    record PASS "privileged docker exec is DENIED (OPA)"
  else
    record FAIL "privileged docker exec denial" "rc=$LAST_RC out=$(tailout)"
  fi
  "${DOCKER[@]}" rm -f "$EXEC_C" >/dev/null 2>&1
else
  record SKIP "docker exec tests" "could not start helper container"
fi

# =============================================================================
# POSITIVE — Docker registry operations
# POST /images/create?fromImage (docker pull) is NOT blocked.
# Only POST /images/load (tar-archive import) is blocked by OPA.
# =============================================================================
section "POSITIVE — Docker registry operations"

expect_cmd_ok "docker pull from registry (POST /images/create?fromImage) is allowed" \
  pull "$IMG_ALPINE"
expect_cmd_ok "docker images list is allowed" \
  images
expect_cmd_ok "docker image inspect is allowed" \
  image inspect "$IMG_ALPINE"
TAG_IMG="pgdt-tag-$$"
"${DOCKER[@]}" tag "$IMG_ALPINE" "${TAG_IMG}:latest" >/dev/null 2>&1 && \
  record PASS "docker tag is allowed" || record FAIL "docker tag failed"
"${DOCKER[@]}" rmi "${TAG_IMG}:latest" >/dev/null 2>&1 && \
  record PASS "docker rmi (tag removal) is allowed" || record FAIL "docker rmi failed"

# =============================================================================
# POSITIVE — Container lifecycle management
# run / start / stop / restart / pause / unpause / kill /
# logs / inspect / top / stats / rename — normal agent operations.
# =============================================================================
section "POSITIVE — Container lifecycle management"

LIFE_C="${LABEL}-life-$RANDOM"
if "${DOCKER[@]}" run -d --label "$LABEL" --name "$LIFE_C" \
     "$IMG_ALPINE" sleep 300 >/dev/null 2>&1; then
  expect_cmd_ok "docker stop is allowed"                stop  -t 3   "$LIFE_C"
  expect_cmd_ok "docker start is allowed"               start         "$LIFE_C"
  expect_cmd_ok "docker restart is allowed"             restart -t 3  "$LIFE_C"
  expect_cmd_ok "docker pause is allowed"               pause         "$LIFE_C"
  expect_cmd_ok "docker unpause is allowed"             unpause       "$LIFE_C"
  expect_cmd_ok "docker kill (SIGUSR1, no stop) is allowed" \
    kill --signal SIGUSR1                                             "$LIFE_C"
  expect_cmd_ok "docker logs is allowed"                logs          "$LIFE_C"
  expect_cmd_ok "docker inspect is allowed"             inspect       "$LIFE_C"
  expect_cmd_ok "docker top is allowed"                 top           "$LIFE_C"
  expect_cmd_ok "docker stats (--no-stream) is allowed" \
    stats --no-stream                                                 "$LIFE_C"

  NEW_NAME="${LIFE_C}-ren"
  "${DOCKER[@]}" rename "$LIFE_C" "$NEW_NAME" >/dev/null 2>&1 && \
    record PASS "docker rename is allowed" || record FAIL "docker rename failed"
  "${DOCKER[@]}" rename "$NEW_NAME" "$LIFE_C" >/dev/null 2>&1 || true

  "${DOCKER[@]}" rm -f "$LIFE_C" >/dev/null 2>&1
else
  record SKIP "container lifecycle tests" "could not start lifecycle container"
fi

# =============================================================================
# POSITIVE — Container runtime features
# env / labels / workdir / user / --init / restart policy / port publishing /
# entrypoint override / read-only rootfs / docker exec with env+workdir
# =============================================================================
section "POSITIVE — Container runtime features"

expect_contains "env var (-e) is passed into container" "TESTVAL=hello-dind" \
  -e TESTVAL=hello-dind "$IMG_ALPINE" env

expect_contains "multiple -e vars and --label are accepted" "VAR2=world" \
  -e VAR1=hello -e VAR2=world --label app=pentest "$IMG_ALPINE" env

expect_contains "working directory (-w) override works" "/custom/wd" \
  -w /custom/wd "$IMG_ALPINE" pwd

# --user: verify OPA does not block it. The container may exit non-zero if the
# UID has no /etc/passwd entry in the image (environment-specific), but OPA
# must not deny the containers/create request.
cap "${DOCKER[@]}" run --rm --label "$LABEL" --user 65534 "$IMG_ALPINE" sh -c 'echo ok'
if echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
  record FAIL "--user (non-root UID) blocked by OPA (must be allowed)" "$(tailout)"
elif [ "$LAST_RC" -eq 0 ]; then
  record PASS "--user (non-root UID) is allowed and runs successfully"
else
  record INFO "--user (non-root UID) is allowed by OPA (container exit=$LAST_RC — non-OPA)" \
    "$(tailout)"
fi

expect_run_ok "--entrypoint override is allowed" \
  --entrypoint sh "$IMG_ALPINE" -c 'echo ok'

expect_run_ok "--init (tini pid1) flag is allowed" \
  --init "$IMG_ALPINE" true

# --restart conflicts with --rm; test via detached run then remove manually.
RST_C="${LABEL}-rst-$RANDOM"
"${DOCKER[@]}" run -d --label "$LABEL" --name "$RST_C" \
  --restart on-failure:3 "$IMG_ALPINE" sh -c 'exit 0' >/dev/null 2>&1 && \
  record PASS "--restart=on-failure policy is accepted at create time" || \
  record FAIL "--restart=on-failure failed" "rc=$? $(tailout)"
"${DOCKER[@]}" rm -f "$RST_C" >/dev/null 2>&1 || true

expect_run_ok "-p port publishing is allowed" \
  -p 127.0.0.1::9999 "$IMG_ALPINE" true

expect_run_ok "--read-only rootfs with multiple --tmpfs is allowed" \
  --read-only \
  --tmpfs /tmp:rw,size=8m \
  --tmpfs /run:rw,noexec,nosuid,size=4m \
  "$IMG_ALPINE" sh -c 'echo ok > /tmp/x && cat /tmp/x >/dev/null'

# docker exec with env / workdir (non-privileged)
FEAT_C="${LABEL}-feat-$RANDOM"
if "${DOCKER[@]}" run -d --rm --label "$LABEL" --name "$FEAT_C" \
     "$IMG_ALPINE" sleep 120 >/dev/null 2>&1; then
  cap "${DOCKER[@]}" exec "$FEAT_C" sh -c 'echo from-exec'
  [ "$LAST_RC" -eq 0 ] && echo "$LAST_OUT" | grep -q "from-exec" && \
    record PASS "docker exec with sh command works" || \
    record FAIL "docker exec sh command failed" "rc=$LAST_RC: $(tailout)"
  cap "${DOCKER[@]}" exec -e EXEC_VAR=exec-test "$FEAT_C" env
  echo "$LAST_OUT" | grep -q "EXEC_VAR=exec-test" && \
    record PASS "docker exec with -e env var works" || \
    record FAIL "docker exec -e failed" "$(tailout)"
  cap "${DOCKER[@]}" exec -w /tmp "$FEAT_C" pwd
  echo "$LAST_OUT" | grep -q "/tmp" && \
    record PASS "docker exec with -w workdir works" || \
    record FAIL "docker exec -w failed" "$(tailout)"
  "${DOCKER[@]}" rm -f "$FEAT_C" >/dev/null 2>&1
else
  record SKIP "docker exec runtime feature tests" "could not start helper container"
fi

# =============================================================================
# POSITIVE — Safe volume & network management
# Named volume create/inspect/use, anonymous volumes, bridge network create,
# network connect/disconnect, read-only API queries (ps/ls/inspect).
# =============================================================================
section "POSITIVE — Safe volume & network management"

# docker ps (read-only, always safe)
expect_cmd_ok "docker ps is allowed"   ps
expect_cmd_ok "docker ps -a is allowed" ps -a

# Named volume lifecycle
SAFE_VOL="${VOLPREFIX}_safe_$RANDOM"
"${DOCKER[@]}" volume create "$SAFE_VOL" >/dev/null 2>&1 && \
  record PASS "docker volume create (no driver opts) is allowed" || \
  record FAIL "docker volume create failed"
expect_cmd_ok "docker volume ls is allowed"      volume ls
cap "${DOCKER[@]}" volume inspect "$SAFE_VOL"
[ "$LAST_RC" -eq 0 ] && record PASS "docker volume inspect is allowed" || \
  record FAIL "docker volume inspect failed" "rc=$LAST_RC: $(tailout)"
expect_run_ok "container using named volume (write+read) is allowed" \
  -v "$SAFE_VOL:/data" "$IMG_ALPINE" \
  sh -c 'echo persistent > /data/f && cat /data/f >/dev/null'
"${DOCKER[@]}" volume rm "$SAFE_VOL" >/dev/null 2>&1 || true

# Anonymous volume (simulates Dockerfile VOLUME /data instruction)
expect_run_ok "anonymous volume (Dockerfile VOLUME /data simulation) is allowed" \
  -v /anon "$IMG_ALPINE" \
  sh -c 'echo ok > /anon/f && cat /anon/f >/dev/null'

# Safe housekeeping
expect_cmd_ok "docker volume prune (--force) is allowed" \
  volume prune --force
expect_cmd_ok "docker container prune (--force) is allowed" \
  container prune --force

# Bridge network lifecycle
SAFE_NET="${NETPREFIX}_safe_$RANDOM"
"${DOCKER[@]}" network create --driver bridge "$SAFE_NET" >/dev/null 2>&1 && \
  record PASS "docker network create (bridge driver) is allowed" || \
  record FAIL "docker network create (bridge) failed"
expect_cmd_ok "docker network ls is allowed" network ls
cap "${DOCKER[@]}" network inspect "$SAFE_NET"
[ "$LAST_RC" -eq 0 ] && record PASS "docker network inspect is allowed" || \
  record FAIL "docker network inspect failed" "rc=$LAST_RC: $(tailout)"

# docker network connect / disconnect with a running container
NETCONN_C="${LABEL}-nc-$RANDOM"
if "${DOCKER[@]}" run -d --rm --label "$LABEL" --name "$NETCONN_C" \
     "$IMG_ALPINE" sleep 60 >/dev/null 2>&1; then
  "${DOCKER[@]}" network connect "$SAFE_NET" "$NETCONN_C" >/dev/null 2>&1 && \
    record PASS "docker network connect is allowed" || \
    record FAIL "docker network connect failed"
  "${DOCKER[@]}" network disconnect "$SAFE_NET" "$NETCONN_C" >/dev/null 2>&1 && \
    record PASS "docker network disconnect is allowed" || \
    record FAIL "docker network disconnect failed"
  "${DOCKER[@]}" rm -f "$NETCONN_C" >/dev/null 2>&1
else
  record SKIP "network connect/disconnect test" "could not start helper container"
fi
"${DOCKER[@]}" network rm "$SAFE_NET" >/dev/null 2>&1 || true

# =============================================================================
# POSITIVE — Compose-style multi-container workflow (PentAGI agent simulation)
# 1. Create isolated bridge network + shared volume
# 2. Start a "target" service container on the network
# 3. "Agent" queries the target, writes findings to the shared volume
# 4. "Report" reads findings (read-only volume mount)
# 5. Tear everything down
# All operations use only bridge networking + named volumes (no bind mounts).
# =============================================================================
section "POSITIVE — Compose-style multi-container workflow (PentAGI agent)"

COMP_NET="${NETPREFIX}_comp_$RANDOM"
COMP_VOL="${VOLPREFIX}_comp_$RANDOM"
COMP_TARGET="${LABEL}-target-$RANDOM"
COMP_OK=1

if "${DOCKER[@]}" network create --driver bridge "$COMP_NET" >/dev/null 2>&1 \
   && "${DOCKER[@]}" volume create "$COMP_VOL" >/dev/null 2>&1; then

  # "Target" — simple nc-based HTTP stub on the isolated network
  if "${DOCKER[@]}" run -d --rm --label "$LABEL" --name "$COMP_TARGET" \
       --network "$COMP_NET" "$IMG_ALPINE" \
       sh -c 'while true; do
                printf "HTTP/1.1 200 OK\r\nContent-Length: 12\r\n\r\ntarget-alive" | nc -l -p 8080
              done' >/dev/null 2>&1; then
    sleep 2

    # "Agent" — queries target and stores result in shared volume
    cap "${DOCKER[@]}" run --rm --label "$LABEL" \
      --network "$COMP_NET" \
      -v "$COMP_VOL:/findings" \
      --cap-add NET_ADMIN --cap-add NET_RAW \
      "$IMG_ALPINE" \
      sh -c "wget -q -T 5 -O /findings/result.txt \
               http://$COMP_TARGET:8080/ 2>/dev/null || true; \
             echo agent-done > /findings/status"
    if [ "$LAST_RC" -eq 0 ]; then
      record PASS "agent: network scan + volume write (bridge network + named volume)"
    else
      record FAIL "agent container failed" "rc=$LAST_RC: $(tailout)"
      COMP_OK=0
    fi

    # "Report" — reads findings from shared volume (read-only)
    if [ "$COMP_OK" -eq 1 ]; then
      cap "${DOCKER[@]}" run --rm --label "$LABEL" \
        -v "$COMP_VOL:/findings:ro" \
        "$IMG_ALPINE" \
        sh -c 'cat /findings/status'
      if [ "$LAST_RC" -eq 0 ] && echo "$LAST_OUT" | grep -q "agent-done"; then
        record PASS "report: read-only volume mount reads agent findings"
      else
        record FAIL "report container failed to read volume" "rc=$LAST_RC: $(tailout)"
      fi
    fi

    "${DOCKER[@]}" rm -f "$COMP_TARGET" >/dev/null 2>&1 || true
  else
    record SKIP "compose workflow: target service" "could not start target container"
    COMP_OK=0
  fi

  "${DOCKER[@]}" network rm "$COMP_NET" >/dev/null 2>&1 || true
  "${DOCKER[@]}" volume rm "$COMP_VOL" >/dev/null 2>&1 || true
else
  record SKIP "compose workflow" "could not create network or volume"
fi

# =============================================================================
# NEGATIVE — OPA authz must reject host-escape primitives at create time
# =============================================================================
section "NEGATIVE — OPA authz (must be denied)"

expect_deny "--privileged is denied" \
  --privileged "$IMG_ALPINE" true
expect_deny "--cap-add SYS_ADMIN is denied" \
  --cap-add SYS_ADMIN "$IMG_ALPINE" true
expect_deny "--cap-add SYS_PTRACE is denied" \
  --cap-add SYS_PTRACE "$IMG_ALPINE" true
expect_deny "--cap-add SYS_MODULE is denied" \
  --cap-add SYS_MODULE "$IMG_ALPINE" true
expect_deny "--cap-add DAC_READ_SEARCH is denied (shocker)" \
  --cap-add DAC_READ_SEARCH "$IMG_ALPINE" true
expect_deny "--cap-add SYS_RAWIO is denied" \
  --cap-add SYS_RAWIO "$IMG_ALPINE" true
expect_deny "--cap-add ALL is denied" \
  --cap-add ALL "$IMG_ALPINE" true
expect_deny "--device passthrough is denied" \
  --device /dev/null:/dev/xnull "$IMG_ALPINE" true
expect_deny "--device-cgroup-rule is denied" \
  --device-cgroup-rule "c 1:3 rmw" "$IMG_ALPINE" true
expect_deny "--security-opt seccomp=unconfined is denied" \
  --security-opt seccomp=unconfined "$IMG_ALPINE" true
expect_deny "--security-opt apparmor=unconfined is denied" \
  --security-opt apparmor=unconfined "$IMG_ALPINE" true
expect_deny "--security-opt no-new-privileges=false is denied" \
  --security-opt no-new-privileges=false "$IMG_ALPINE" true
expect_deny "--pid=host is denied" \
  --pid host "$IMG_ALPINE" true
expect_deny "--ipc=host is denied" \
  --ipc host "$IMG_ALPINE" true
expect_deny "--ipc=shareable is denied" \
  --ipc shareable "$IMG_ALPINE" true
expect_deny "--uts=host is denied" \
  --uts host "$IMG_ALPINE" true
expect_deny "--userns=host is denied" \
  --userns host "$IMG_ALPINE" true
expect_deny "--cgroupns=host is denied" \
  --cgroupns host "$IMG_ALPINE" true
expect_deny "host bind-mount -v /:/host is denied" \
  -v /:/host "$IMG_ALPINE" true
expect_deny "host bind-mount -v /etc:/x is denied" \
  -v /etc:/x "$IMG_ALPINE" true
expect_deny "host bind-mount --mount type=bind is denied" \
  --mount "type=bind,src=/etc,dst=/x" "$IMG_ALPINE" true

# --pid=container:<id> (namespace-sharing chain) must be denied.
PIDT="${LABEL}-pidt-$RANDOM"
if "${DOCKER[@]}" run -d --rm --label "$LABEL" --name "$PIDT" "$IMG_ALPINE" sleep 60 >/dev/null 2>&1; then
  cap "${DOCKER[@]}" run --rm --label "$LABEL" --pid "container:$PIDT" "$IMG_ALPINE" true
  if [ "$LAST_RC" -ne 0 ] && echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
    record PASS "--pid=container:<id> is denied"
  else
    record FAIL "--pid=container:<id> denial" "rc=$LAST_RC out=$(tailout)"
  fi
  "${DOCKER[@]}" rm -f "$PIDT" >/dev/null 2>&1
else
  record SKIP "--pid=container:<id> test" "could not start target container"
fi

# Old API version must not bypass the plugin hook.
cap env DOCKER_API_VERSION=1.24 "${DOCKER[@]}" run --rm --label "$LABEL" --privileged "$IMG_ALPINE" true
if [ "$LAST_RC" -ne 0 ] && echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
  record PASS "API version downgrade (1.24) does NOT bypass authz"
else
  record FAIL "API version downgrade bypass" "rc=$LAST_RC out=$(tailout)"
fi

# =============================================================================
# NEGATIVE — OPA endpoint coverage (archive / commit / build)
# These endpoints bypass containers/create and were used in prior phases to
# exfiltrate PKI material (archive) and inject malicious image layers (build).
# OPA denies them via is_container_archive / is_container_commit /
# is_image_build (legacy API) / is_buildkit_session (Docker 23+ default).
# =============================================================================
section "NEGATIVE — OPA endpoint coverage (archive / commit / build)"

# GET /containers/{id}/archive — docker cp uses this endpoint internally.
ARCHC="${LABEL}-arch-$RANDOM"
if "${DOCKER[@]}" run -d --rm --label "$LABEL" --name "$ARCHC" "$IMG_ALPINE" sleep 60 >/dev/null 2>&1; then
  TMP_ARCH="/tmp/opa_archive_$$"
  cap "${DOCKER[@]}" cp "$ARCHC:/etc/hostname" "${TMP_ARCH}"
  rm -f "${TMP_ARCH}" 2>/dev/null
  if [ "$LAST_RC" -ne 0 ] && echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
    record PASS "GET /containers/{id}/archive is DENIED (OPA) — docker cp blocked"
  else
    record FAIL "GET /containers/{id}/archive denial (docker cp)" \
      "rc=$LAST_RC out=$(tailout) — SECURITY GAP: file exfiltration possible"
  fi
  "${DOCKER[@]}" rm -f "$ARCHC" >/dev/null 2>&1
else
  record SKIP "archive endpoint test" "could not start helper container"
fi

# POST /commit — docker commit snapshots the container FS including cached creds.
COMC="${LABEL}-com-$RANDOM"
if "${DOCKER[@]}" run -d --rm --label "$LABEL" --name "$COMC" "$IMG_ALPINE" sleep 60 >/dev/null 2>&1; then
  cap "${DOCKER[@]}" commit "$COMC" "opa-test-commit-$$:latest"
  if [ "$LAST_RC" -ne 0 ] && echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
    record PASS "POST /commit is DENIED (OPA) — docker commit blocked"
  else
    record FAIL "POST /commit denial" "rc=$LAST_RC out=$(tailout) — SECURITY GAP: state snapshot possible"
  fi
  "${DOCKER[@]}" rmi "opa-test-commit-$$:latest" >/dev/null 2>&1
  "${DOCKER[@]}" rm -f "$COMC" >/dev/null 2>&1
else
  record SKIP "commit endpoint test" "could not start helper container"
fi

# POST /build — docker build can embed malicious layers.
LAST_OUT=$(printf 'FROM %s\n' "${IMG_ALPINE}" | "${DOCKER[@]}" build --no-cache -t "opa-test-build-$$:latest" -f - . 2>&1)
LAST_RC=$?
if [ "$LAST_RC" -ne 0 ] && echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
  record PASS "POST /build is DENIED (OPA) — docker build blocked"
else
  record FAIL "POST /build denial" "rc=$LAST_RC out=$(tailout) — SECURITY GAP: malicious image injection possible"
  "${DOCKER[@]}" rmi "opa-test-build-$$:latest" >/dev/null 2>&1
fi

# POST /containers/{id}/update with DeviceCgroupRules — injects block device
# cgroup permissions post-creation. While not immediately exploitable (mknod
# S_IFBLK is seccomp-blocked and /dev/ has no block nodes), it bypasses the
# device-cgroup eBPF filter set at create-time and must be denied.
UPDC="${LABEL}-upd-$RANDOM"
if "${DOCKER[@]}" run -d --rm --label "$LABEL" --name "$UPDC" "$IMG_ALPINE" sleep 60 >/dev/null 2>&1; then
  cap "${DOCKER[@]}" update --blkio-weight 0 "$UPDC"
  if [ "$LAST_RC" -eq 0 ]; then
    record INFO "POST /containers/{id}/update with safe fields is allowed (memory/cpu/blkio OK)"
  fi
  # Inject dangerous DeviceCgroupRules — must be denied by OPA update_violation.
  if command -v curl >/dev/null 2>&1; then
    dind_api_post "/containers/$UPDC/update" \
      '{"DeviceCgroupRules":["b 8:* rwm","b *:* rm"]}'
    if echo "$LAST_OUT" | grep -qiE "$DENY_RE|not allowed|denied"; then
      record PASS "POST /containers/{id}/update with DeviceCgroupRules is DENIED (OPA)"
    elif echo "$LAST_OUT" | grep -qiE '"Warnings":\[\]|"Warnings":null'; then
      record FAIL "POST /containers/{id}/update DeviceCgroupRules injection ALLOWED" \
        "Response: $(tailout) — SECURITY GAP: device cgroup bypass possible"
    else
      record SKIP "POST /containers/{id}/update DeviceCgroupRules test" \
        "unexpected response: $(tailout)"
    fi
  else
    record SKIP "POST /containers/{id}/update DeviceCgroupRules test" "curl not available"
  fi
  "${DOCKER[@]}" rm -f "$UPDC" >/dev/null 2>&1
else
  record SKIP "container update endpoint test" "could not start helper container"
fi

# =============================================================================
# NEGATIVE — OPA new endpoint coverage (images/load / plugins / swarm)
# Second hardening round: endpoints added after initial policy deployment.
# =============================================================================
section "NEGATIVE — OPA new endpoints (images/load / plugins / swarm)"

# POST /images/load — tar archive import bypasses registry and can carry
# device nodes or setuid binaries. Must be denied.
cap "${DOCKER[@]}" image load -i /dev/null
if [ "$LAST_RC" -ne 0 ] && echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
  record PASS "POST /images/load (docker image load) is DENIED (OPA)"
else
  record FAIL "POST /images/load denial" \
    "rc=$LAST_RC out=$(tailout) — SECURITY GAP: arbitrary tar import possible"
fi

# POST /plugins/pull — installing plugins grants daemon-level privileges and
# can implement malicious authz/volume/network logic. Must be denied.
cap "${DOCKER[@]}" plugin install --grant-all-permissions \
  "opa-test-nonexistent-plugin-$$" 2>&1
if [ "$LAST_RC" -ne 0 ] && echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
  record PASS "POST /plugins/pull (docker plugin install) is DENIED (OPA)"
else
  record FAIL "POST /plugins/pull denial" \
    "rc=$LAST_RC out=$(tailout) — SECURITY GAP: plugin management possible"
fi

# POST /swarm/init — Swarm services bypass per-container HostConfig validation.
cap "${DOCKER[@]}" swarm init
if [ "$LAST_RC" -ne 0 ] && echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
  record PASS "POST /swarm/init is DENIED (OPA)"
else
  record FAIL "POST /swarm/init denial" \
    "rc=$LAST_RC out=$(tailout) — SECURITY GAP: Swarm service bypass possible"
  "${DOCKER[@]}" swarm leave --force >/dev/null 2>&1 || true
fi

# =============================================================================
# NEGATIVE — OPA containers/create new violations (sysctl / DeviceRequests / runtime)
# =============================================================================
section "NEGATIVE — OPA containers/create new violations"

# --sysctl: with --network=host + CAP_NET_ADMIN, sysctl writes propagate to
# the host network namespace. Must be denied even for non-dangerous keys.
expect_deny "--sysctl in containers/create is DENIED (OPA)" \
  --sysctl "net.ipv4.ip_forward=1" "$IMG_ALPINE" true

# DeviceRequests (GPU/accelerator): not needed for pentest workloads;
# device plugins can expose DMA or sensitive hardware paths.
if command -v curl >/dev/null 2>&1; then
  dind_api_post "/containers/create" \
    "{\"Image\":\"${IMG_ALPINE}\",\"Cmd\":[\"true\"],\"HostConfig\":{\"DeviceRequests\":[{\"Driver\":\"\",\"Count\":-1,\"Capabilities\":[[\"gpu\"]],\"Options\":{}}]}}"
  if echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
    record PASS "DeviceRequests (GPU passthrough) in containers/create is DENIED (OPA)"
  elif echo "$LAST_OUT" | grep -qiE '"Id":'; then
    _cid=$(echo "$LAST_OUT" | grep -o '"Id":"[^"]*"' | head -1 | cut -d'"' -f4)
    [ -n "$_cid" ] && "${DOCKER[@]}" rm -f "$_cid" >/dev/null 2>&1 || true
    record FAIL "DeviceRequests denial" \
      "Container created — SECURITY GAP: device plugin access possible"
  else
    record SKIP "DeviceRequests test" "unexpected curl response: $(tailout)"
  fi
else
  record SKIP "DeviceRequests test" "curl not available"
fi

# --runtime: non-standard runtimes may have different seccomp/capability
# defaults. OPA blocks any runtime that is not empty or "runc".
expect_deny "--runtime=<unsupported> is DENIED (OPA)" \
  --runtime unsupported-test-runtime "$IMG_ALPINE" true

# =============================================================================
# NEGATIVE — OPA volume & network new violations (bind-opts / macvlan / ipvlan)
# =============================================================================
section "NEGATIVE — OPA volume & network violations (bind-opts / L2 drivers)"

# docker volume create with type=none,o=bind — emulates bind-mount from DinD's
# own filesystem into nested containers (e.g. exposing /etc/docker/dind/certs).
expect_cmd_deny \
  "docker volume create with type=none,o=bind is DENIED (OPA)" \
  volume create \
    --driver local --opt type=none --opt o=bind --opt device=/tmp

# o=bind alone (without type=none) is also blocked.
expect_cmd_deny \
  "docker volume create with o=bind is DENIED (OPA)" \
  volume create \
    --driver local --opt o=bind --opt device=/tmp

# docker network create with macvlan — attaches container directly to the
# host's L2 segment, bypassing iptables rules.
NET_MACVLAN="pgdt-macvlan-$$"
cap "${DOCKER[@]}" network create --driver macvlan "$NET_MACVLAN"
if [ "$LAST_RC" -ne 0 ] && echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
  record PASS "docker network create (macvlan) is DENIED (OPA)"
else
  "${DOCKER[@]}" network rm "$NET_MACVLAN" >/dev/null 2>&1 || true
  record FAIL "docker network create macvlan denial" \
    "rc=$LAST_RC out=$(tailout) — SECURITY GAP: L2 network attachment possible"
fi

# docker network create with ipvlan (same risk as macvlan).
NET_IPVLAN="pgdt-ipvlan-$$"
cap "${DOCKER[@]}" network create --driver ipvlan "$NET_IPVLAN"
if [ "$LAST_RC" -ne 0 ] && echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
  record PASS "docker network create (ipvlan) is DENIED (OPA)"
else
  "${DOCKER[@]}" network rm "$NET_IPVLAN" >/dev/null 2>&1 || true
  record FAIL "docker network create ipvlan denial" \
    "rc=$LAST_RC out=$(tailout) — SECURITY GAP: L2 network attachment possible"
fi

# =============================================================================
# NEGATIVE — kernel-enforced floor (seccomp + capability bounding set)
# =============================================================================
section "NEGATIVE — seccomp / capabilities (must be refused in-container)"

expect_inner_fail "mknod BLOCK device is blocked (seccomp) — debugfs escape" \
  "not permitted|Operation not permitted" \
  "$IMG_ALPINE" sh -c 'mknod /tmp/sda b 8 0'

expect_inner_fail "mount() is refused without CAP_SYS_ADMIN" \
  "permission denied|not permitted|must be superuser|Operation not permitted" \
  "$IMG_ALPINE" sh -c 'mount -t tmpfs none /mnt'

expect_inner_fail "raw block-device read via dd on a pre-existing node is refused" \
  "not permitted|permission denied|No such file|cannot open" \
  "$IMG_ALPINE" sh -c 'dd if=/dev/sda bs=512 count=1 of=/dev/null'

# nsenter into host namespaces (needs setns + CAP_SYS_ADMIN, both unavailable).
expect_inner_fail "nsenter into PID 1 namespaces is refused" \
  "not permitted|permission denied|No such|cannot|reassociate|setns" \
  "$IMG_ALPINE" sh -c 'nsenter --target 1 --mount --pid -- true 2>&1 || (apk add --no-cache util-linux >/dev/null 2>&1; nsenter -t 1 -m -p -- true)'

# =============================================================================
# NEGATIVE — defense-in-depth (read-only mounts, cgroup v2)
# Verifies the kernel-enforced controls that provide guaranteed mitigation for
# several CVE classes INDEPENDENT of OPA/seccomp/capability configuration.
# If any of these fail it means the DinD launch flags were changed unsafely.
# =============================================================================
section "NEGATIVE — defense-in-depth (RO mounts / cgroup v2)"

# /proc/sys must be read-only — the ONLY reliable mitigation for
# CVE-2025-31133 (core_pattern write via symlink bypass) and all other
# kernel-parameter attack classes. If this write succeeds the CVE is exploitable
# even with the current runc version.
expect_inner_fail "/proc/sys is read-only — core_pattern write blocked (CVE-2025-31133 mitigation)" \
  "Read-only file system|read-only file system|not permitted|Operation not permitted" \
  "$IMG_ALPINE" sh -c 'echo "|/tmp/escape" > /proc/sys/kernel/core_pattern'

# /proc/sysrq-trigger must be read-only — prevents kernel-panic-based attacks.
# sysrq 'c' (crash) was used in Phase 4 to kill host processes and crash the host.
expect_inner_fail "/proc/sysrq-trigger is read-only — kernel panic vector blocked" \
  "Read-only file system|read-only file system|not permitted|Operation not permitted" \
  "$IMG_ALPINE" sh -c 'echo c > /proc/sysrq-trigger'

# /sys/fs/cgroup must be mounted read-only — prevents cgroup manipulation attacks.
# A writable cgroupfs + CAP_SYS_ADMIN enables the classic release_agent escape.
cap "${DOCKER[@]}" run --rm --label "$LABEL" "$IMG_ALPINE" \
  sh -c 'grep -E "cgroup" /proc/mounts | head -1'
if [ "$LAST_RC" -eq 0 ] && echo "$LAST_OUT" | grep -qE '\bro\b|,ro,|,ro$| ro '; then
  record PASS "/sys/fs/cgroup is mounted read-only"
else
  record FAIL "/sys/fs/cgroup mount options — expected RO" \
    "mounts: $(tailout) — cgroup manipulation may be possible"
fi

# Cgroup v2 (unified hierarchy) must be in use — the v1 release_agent mechanism
# (CVE-2022-0492 class) does not exist in v2.
cap "${DOCKER[@]}" run --rm --label "$LABEL" "$IMG_ALPINE" \
  sh -c 'grep -qF "cgroup2" /proc/mounts && echo CGROUPV2 || echo CGROUPV1'
if [ "$LAST_RC" -eq 0 ] && echo "$LAST_OUT" | grep -q "CGROUPV2"; then
  record PASS "cgroup v2 (unified) in use — release_agent escape class eliminated"
else
  record FAIL "cgroup v1 detected or check failed" \
    "$(tailout) — CVE-2022-0492 class may be applicable; verify cgroup2 mount"
fi

# /proc/sys/kernel/core_pattern must NOT be pre-staged as a pipe handler by a
# previous session. A value starting with '|' means a payload script would execute
# on any container crash, giving root code execution on the host.
cap "${DOCKER[@]}" run --rm --label "$LABEL" "$IMG_ALPINE" \
  sh -c 'cat /proc/sys/kernel/core_pattern'
if echo "$LAST_OUT" | grep -qE '^\|'; then
  if echo "$LAST_OUT" | grep -qiE '/apport|/systemd-coredump'; then
    record INFO "core_pattern uses system crash reporter (Ubuntu/systemd default — expected)" \
      "$(tailout | cut -c1-80)"
  else
    record FAIL "core_pattern contains a PIPE HANDLER — potential trapdoor" \
      "value: $(tailout) — clear with: sysctl -w kernel.core_pattern=core"
  fi
else
  record PASS "core_pattern is not a pipe handler ($(tailout | cut -c1-40))"
fi

# =============================================================================
# NEGATIVE — worker container host-escape prevention
# These checks run directly in the test container (not via docker CLI) and
# verify that the same kernel-enforced isolation that protects the worker
# container is in place: cgroup device rules, PID namespace, capability limits.
# Threat model: attacker has shell access in the worker/test container and
# tries to read the host disk, see host processes, or load kernel modules.
# =============================================================================
section "NEGATIVE — worker container host-escape prevention"

# Block device: the host disk partition (e.g. /dev/sda1) underlies the /work
# bind mount. MKNOD cap allows creating a device node, but cgroup device rules
# must prevent any I/O on block devices not explicitly allowed.
BLKNODE="/tmp/pgdt-blk-$$"
if mknod "$BLKNODE" b 8 1 2>/dev/null; then
  if dd if="$BLKNODE" bs=512 count=1 of=/dev/null 2>/dev/null; then
    rm -f "$BLKNODE"
    record FAIL "mknod+dd read of /dev/sda1 node SUCCEEDED — cgroup device rules not enforced" \
      "CRITICAL: host disk content readable from worker container"
  else
    rm -f "$BLKNODE"
    record PASS "mknod(S_IFBLK b 8:1) succeeds but dd READ BLOCKED — cgroup device rules work"
  fi
else
  record INFO "mknod b 8 1 blocked at creation (seccomp/cap) — block node not creatable"
fi

# PID namespace: PID 1 must be a container process, not the host's systemd/init.
PID1_CMD=$(tr '\0' ' ' < /proc/1/cmdline 2>/dev/null | cut -c1-80)
record INFO "Worker PID 1 command" "${PID1_CMD:-unreadable}"
if echo "$PID1_CMD" | grep -qiE 'systemd|/sbin/init|/lib/systemd'; then
  record FAIL "PID 1 is host init/systemd — worker shares host PID namespace" \
    "All host processes are visible — PID namespace isolation broken"
else
  record PASS "Worker PID 1 is container process (not host init) — PID namespace isolated"
fi

PID_COUNT=$(ls /proc/ 2>/dev/null | grep -c '^[0-9]\+$' || echo 0)
if [ "$PID_COUNT" -gt 200 ]; then
  record FAIL "Visible PID count = $PID_COUNT — likely seeing host process table"
else
  record PASS "Visible PID count = $PID_COUNT (expected for isolated container)"
fi

# /proc/1/root: must point to container root, not host root.
if [ -d /proc/1/root/opt/pentagi ] 2>/dev/null; then
  record FAIL "/proc/1/root/opt/pentagi accessible — container may see host filesystem" \
    "Symlinks in /work pointing to /proc/1/root/... can exfiltrate host files"
else
  record PASS "/proc/1/root/opt/pentagi not visible — /proc/1/root is container root (not host)"
fi

# Capability limits in the worker container itself.
if modprobe dummy 2>/dev/null; then
  rmmod dummy 2>/dev/null || true
  record FAIL "modprobe succeeded in worker — SYS_MODULE present (unexpected)"
else
  record PASS "modprobe denied in worker — SYS_MODULE not granted"
fi

MNTDIR="/tmp/pgdt-mnt-$$"
mkdir -p "$MNTDIR" 2>/dev/null
if mount -t tmpfs none "$MNTDIR" 2>/dev/null; then
  umount "$MNTDIR" 2>/dev/null; rmdir "$MNTDIR" 2>/dev/null
  record FAIL "mount(tmpfs) succeeded in worker — SYS_ADMIN present (unexpected)"
else
  rmdir "$MNTDIR" 2>/dev/null
  record PASS "mount(tmpfs) denied in worker — SYS_ADMIN not granted"
fi

# /proc/sys must be read-only even in the worker (ReadonlyPaths: docker-default).
if echo "1" > /proc/sys/kernel/dmesg_restrict 2>/dev/null; then
  record FAIL "/proc/sys writable in worker — ReadonlyPaths not enforced"
else
  record PASS "/proc/sys read-only in worker (ReadonlyPaths enforced by Docker)"
fi

# =============================================================================
# NEGATIVE — MaskedPaths / ReadonlyPaths / CgroupParent overrides
#
# CONFIRMED ESCAPE VECTOR (Phase 7): Creating a container with ReadonlyPaths:[]
# removes Docker's default /proc/sys read-only protection. Since
# /proc/sys/kernel is NOT namespaced, writing modprobe_path from inside
# the container changes it on the PHYSICAL HOST. APK package install then
# triggers request_module() which executes the payload as root in the init
# mount namespace.
#
# MaskedPaths:[] exposes /proc/kcore, /proc/keys, /sys/firmware/efi/efivars.
# CgroupParent changes which device eBPF programs are active for the container.
#
# Safe rule: absent field = Docker safe defaults (ALLOWED).
#            any explicit array ([] or [...]) = DENIED.
# =============================================================================
section "NEGATIVE — MaskedPaths / ReadonlyPaths / CgroupParent overrides (Phase 7 escape vector)"

# Helper: POST containers/create and expect OPA denial.
# Uses dind_api_post — handles both unix-socket and TLS/TCP transport
# (same as the rest of the script, driven by DOCKER[@]).
# Cleanup on accidental creation uses docker CLI (also transport-aware).
_api_deny_test(){
  local name="$1" json="$2"
  if ! command -v curl >/dev/null 2>&1; then
    record SKIP "$name" "curl not available"
    return
  fi
  dind_api_post "/containers/create" "$json"
  if echo "$LAST_OUT" | grep -qiE "$DENY_RE"; then
    record PASS "$name"
  elif echo "$LAST_OUT" | grep -qiE '"Id":|"Warnings":'; then
    record FAIL "$name" "SECURITY GAP: container was CREATED — $(tailout)"
    local cid
    cid=$(echo "$LAST_OUT" | python3 -c \
      "import sys,json; print(json.load(sys.stdin).get('Id',''))" 2>/dev/null || true)
    [ -n "$cid" ] && "${DOCKER[@]}" rm -f "$cid" >/dev/null 2>&1 || true
  else
    record FAIL "$name" "unexpected response: $(tailout)"
  fi
}

# NEGATIVE: ReadonlyPaths overrides
_api_deny_test \
  "ReadonlyPaths:[] (removes /proc/sys RO, enables modprobe host escape) is DENIED" \
  '{"Image":"'"$IMG_ALPINE"'","Cmd":["true"],"HostConfig":{"ReadonlyPaths":[]}}'

_api_deny_test \
  "ReadonlyPaths:[\"/proc/bus\"] (partial — /proc/sys unprotected) is DENIED" \
  '{"Image":"'"$IMG_ALPINE"'","Cmd":["true"],"HostConfig":{"ReadonlyPaths":["/proc/bus"]}}'

_api_deny_test \
  "ReadonlyPaths:[\"/proc/sys\",\"/proc/bus\"] (any explicit array) is DENIED" \
  '{"Image":"'"$IMG_ALPINE"'","Cmd":["true"],"HostConfig":{"ReadonlyPaths":["/proc/sys","/proc/bus"]}}'

# NEGATIVE: MaskedPaths overrides
_api_deny_test \
  "MaskedPaths:[] (unmasks /proc/kcore, /proc/keys, /sys/firmware/efi) is DENIED" \
  '{"Image":"'"$IMG_ALPINE"'","Cmd":["true"],"HostConfig":{"MaskedPaths":[]}}'

_api_deny_test \
  "MaskedPaths:[\"/proc/acpi\"] (partial — /proc/kcore unmasked) is DENIED" \
  '{"Image":"'"$IMG_ALPINE"'","Cmd":["true"],"HostConfig":{"MaskedPaths":["/proc/acpi"]}}'

# Exact Phase 7 exploit payload: both fields cleared + host networking.
_api_deny_test \
  "MaskedPaths:[] + ReadonlyPaths:[] combined (exact Phase 7 exploit JSON) is DENIED" \
  '{"Image":"'"$IMG_ALPINE"'","Cmd":["sleep","300"],"HostConfig":{"MaskedPaths":[],"ReadonlyPaths":[],"NetworkMode":"host"}}'

# NEGATIVE: CgroupParent
_api_deny_test \
  "CgroupParent:\"/\" (root cgroup, may bypass device eBPF policy) is DENIED" \
  '{"Image":"'"$IMG_ALPINE"'","Cmd":["true"],"HostConfig":{"CgroupParent":"/"}}'

_api_deny_test \
  "CgroupParent:\"system.slice\" (systemd cgroup placement) is DENIED" \
  '{"Image":"'"$IMG_ALPINE"'","Cmd":["true"],"HostConfig":{"CgroupParent":"system.slice"}}'

# POSITIVE: containers WITHOUT these fields work normally.
# Docker applies its own safe defaults when the fields are absent from the request.
expect_run_ok \
  "container without MaskedPaths/ReadonlyPaths/CgroupParent (Docker defaults) is ALLOWED" \
  "$IMG_ALPINE" true

# Primary Phase 7 safety check: modprobe_path must NOT be writable.
# If this write succeeds, a container can change the host's kernel modprobe_path.
expect_inner_fail \
  "modprobe_path NOT writable in default container — Phase 7 primary safety check" \
  "Read-only file system|read-only file system|not permitted|Operation not permitted" \
  "$IMG_ALPINE" sh -c 'echo /tmp/evil > /proc/sys/kernel/modprobe'

# core_pattern must also not be writable (same /proc/sys protection).
expect_inner_fail \
  "core_pattern NOT writable in default container (same /proc/sys RO protection)" \
  "Read-only file system|read-only file system|not permitted|Operation not permitted" \
  "$IMG_ALPINE" sh -c 'echo "|/tmp/payload" > /proc/sys/kernel/core_pattern'

# =============================================================================
# SUMMARY
# =============================================================================
section "Summary"
printf "  total=%d  ${GRN}pass=%d${RST}  ${RED}fail=%d${RST}  ${YLW}skip=%d${RST}  ${BLU}info=%d${RST}\n" \
  "$TOTAL" "$PASS" "$FAIL" "$SKIP" "$INFO"
if [ "$FAIL" -gt 0 ]; then
  printf "\n  ${RED}${BLD}Failures:${RST}\n"
  for n in "${FAILED_NAMES[@]}"; do printf "    ${RED}- %s${RST}\n" "$n"; done
  printf "\n  ${RED}A failed NEGATIVE test = isolation gap. A failed POSITIVE test = false positive that breaks agents.${RST}\n"
  exit 1
fi
printf "\n  ${GRN}${BLD}All executed tests passed.${RST}\n"
exit 0
