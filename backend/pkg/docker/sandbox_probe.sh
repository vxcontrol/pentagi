#!/bin/sh
# Sandbox policy probe. Emits one JSON object per line on stdout: "env" once,
# then "check" per probe, then "summary". Nothing else is printed, so the caller
# unmarshals line by line instead of parsing prose.
#
# Runs inside a worker container and speaks to the daemon that worker is given,
# through the ordinary DOCKER_HOST / DOCKER_TLS_VERIFY / DOCKER_CERT_PATH.
set -u

emit() { printf '%s\n' "$1"; }
die() { emit "{\"kind\":\"fatal\",\"reason\":\"$1\"}"; exit 1; }

case "${DOCKER_HOST:-}" in
    unix://*) endpoint='http://docker'; set -- --unix-socket "${DOCKER_HOST#unix://}" ;;
    tcp://*|https://*|http://*)
        address=${DOCKER_HOST#*://}
        if [ -n "${DOCKER_TLS_VERIFY:-}" ]; then
            [ -n "${DOCKER_CERT_PATH:-}" ] || die "cert_path_required"
            endpoint="https://$address"
            set -- --cacert "${DOCKER_CERT_PATH}/ca.pem" \
                   --cert "${DOCKER_CERT_PATH}/cert.pem" --key "${DOCKER_CERT_PATH}/key.pem"
        else
            endpoint="http://$address"; set --
        fi
        ;;
    *) die "docker_host_unsupported" ;;
esac

CURL="curl --noproxy * --silent --show-error --connect-timeout 3 --max-time 8"

version=$($CURL "$@" "$endpoint/version" 2>/dev/null) || die "version_unreachable"
api_version=$(printf '%s' "$version" | sed -n 's/.*"ApiVersion"[[:space:]]*:[[:space:]]*"\([0-9][0-9.]*\)".*/\1/p')
[ -n "$api_version" ] || die "version_malformed"
api="$endpoint/v$api_version"

# The daemon document is passed through verbatim rather than picked apart here:
# /info carries several fields named ID, and only the root one identifies the
# daemon. The caller unmarshals it, which cannot pick the wrong one.
info=$($CURL "$@" "$api/info" 2>/dev/null | tr -d '\n') || die "info_unreachable"
case "$info" in
    '{'*) ;;
    *) die "info_malformed" ;;
esac
emit "{\"kind\":\"info\",\"api\":\"$api_version\",\"body\":$info}"

nonce=$(cat /proc/sys/kernel/random/uuid 2>/dev/null || echo "$$-$(date +%s)")
work=$(mktemp -d) || die "no_tmpdir"

# status_of METHOD PATH BODY -> "<code> <lowercased body>"
status_of() {
    _m=$1; _p=$2; _b=$3; shift 3
    if [ -n "$_b" ]; then
        set -- "$@" --header 'Content-Type: application/json' --data-binary "$_b"
    fi
    _r=$($CURL "$@" --write-out '\n%{http_code}' --request "$_m" "$api$_p" 2>/dev/null) || _r="
000"
    printf '%s %s' "${_r##*
}" "$(printf '%s' "${_r%
*}" | tr '[:upper:]' '[:lower:]' | tr -d '\n')"
}

# deny ID GROUP METHOD PATH BODY -- the request must be refused by the plugin.
# Every body names an object that does not exist or cannot be acted on, so an
# allowed request still changes nothing; it is reported, not performed.
deny() {
    _id=$1; _g=$2; _m=$3; _p=$4; _b=$5; shift 5
    _out=$(status_of "$_m" "$_p" "$_b" "$@")
    _code=${_out%% *}; _body=${_out#* }
    case "$_code:$_body" in
        403:*authorization?denied*|403:*opa-docker-authz*) _res=pass ;;
        *) _res=fail ;;
    esac
    emit "{\"kind\":\"check\",\"id\":\"$_id\",\"group\":\"$_g\",\"expect\":\"deny\",\"result\":\"$_res\",\"status\":$_code}"
}

# create ID BODY -- a containers/create probe against an image that cannot exist.
create() {
    _id=$1; _hc=$2; shift 2
    deny "$_id" "create" POST "/containers/create?name=sandbox-probe-$nonce-$_id" \
        "{\"Image\":\"sandbox-probe-$nonce:absent\",\"Cmd\":[\"true\"],\"HostConfig\":$_hc}" "$@"
}

group_create() {
    create privileged            '{"Privileged":true}' "$@"
    create host-bind             '{"Binds":["/:/host"]}' "$@"
    create host-bind-structured  '{"Mounts":[{"Type":"bind","Source":"/etc","Target":"/h"}]}' "$@"
    create cap-sys-admin         '{"CapAdd":["SYS_ADMIN"]}' "$@"
    create cap-sys-module        '{"CapAdd":["SYS_MODULE"]}' "$@"
    create device                '{"Devices":[{"PathOnHost":"/dev/null","PathInContainer":"/dev/x","CgroupPermissions":"rwm"}]}' "$@"
    create device-cgroup-rule    '{"DeviceCgroupRules":["c 1:3 rwm"]}' "$@"
    create pid-host              '{"PidMode":"host"}' "$@"
    create ipc-host              '{"IpcMode":"host"}' "$@"
    create userns-host           '{"UsernsMode":"host"}' "$@"
    create seccomp-unconfined    '{"SecurityOpt":["seccomp=unconfined"]}' "$@"
    create apparmor-unconfined   '{"SecurityOpt":["apparmor=unconfined"]}' "$@"
    create masked-paths-cleared  '{"MaskedPaths":[]}' "$@"
    create readonly-paths-cleared '{"ReadonlyPaths":[]}' "$@"
    create cgroup-parent-root    '{"CgroupParent":"/"}' "$@"
    create sysctl                '{"Sysctls":{"net.ipv4.ip_forward":"1"}}' "$@"
    create proc-escape-combined  '{"MaskedPaths":[],"ReadonlyPaths":[],"NetworkMode":"host"}' "$@"
    create custom-runtime        '{"Runtime":"sysbox-runc"}' "$@"
}

group_endpoint() {
    absent="sandbox-probe-$nonce-absent"
    deny privileged-exec  endpoint POST "/containers/$absent/exec" '{"Privileged":true,"Cmd":["true"]}' "$@"
    deny commit           endpoint POST "/commit?container=$absent" '{}' "$@"
    deny archive-read     endpoint GET  "/containers/$absent/archive?path=%2Fetc" '' "$@"
    deny images-load      endpoint POST "/images/load" '' "$@"
    deny image-build      endpoint POST "/build" '' "$@"
    deny swarm-init       endpoint POST "/swarm/init" '{"ListenAddr":""}' "$@"
    deny bind-volume      endpoint POST "/volumes/create" \
        "{\"Name\":\"sandbox-probe-$nonce-vol\",\"Driver\":\"local\",\"DriverOpts\":{\"type\":\"none\",\"device\":\"/\",\"o\":\"bind\"}}" "$@"
    deny macvlan-network  endpoint POST "/networks/create" \
        "{\"Name\":\"sandbox-probe-$nonce-mac\",\"Driver\":\"macvlan\"}" "$@"
    deny ipvlan-network   endpoint POST "/networks/create" \
        "{\"Name\":\"sandbox-probe-$nonce-ipv\",\"Driver\":\"ipvlan\"}" "$@"
}

# ok ID RESULT STATUS -- record a positive check.
ok() {
    emit "{\"kind\":\"check\",\"id\":\"$1\",\"group\":\"positive\",\"expect\":\"allow\",\"result\":\"$2\",\"status\":$3}"
}

# The positive group is what the sandbox exists FOR: an agent must be able to
# fetch an image and run something in it. A policy tight enough to fail this is
# as broken as one that lets an escape through, and nothing in the negative
# group would notice. Unlike those probes these create real objects, so every
# one is removed again, by a name only this run uses.
group_positive() {
    [ -n "${PENTAGI_TEST_IMAGE:-}" ] || return 0
    _img=$PENTAGI_TEST_IMAGE
    case "${_img##*/}" in
        *:*) _repo=${_img%:*}; _tag=${_img##*:} ;;
        *)   _repo=$_img;      _tag=latest ;;
    esac
    _ctr="sandbox-probe-$nonce-run"
    _vol="sandbox-probe-$nonce-okvol"
    _net="sandbox-probe-$nonce-oknet"

    # The pull is the headline: everything else an agent does depends on it, and
    # it is the one check that proves the sandbox can reach a registry at all.
    # Given its own budget because a cold first pull is not a fast operation.
    _pull=$(curl --noproxy '*' --silent --show-error --connect-timeout 5 --max-time 600 \
        "$@" --output /dev/null --write-out '%{http_code}' \
        --request POST "$api/images/create?fromImage=$_repo&tag=$_tag" 2>/dev/null) || _pull=000
    [ "$_pull" = "200" ] && ok image-pull pass "$_pull" || { ok image-pull fail "$_pull"; return 0; }

    _out=$(status_of POST "/containers/create?name=$_ctr" \
        "{\"Image\":\"$_img\",\"Cmd\":[\"/bin/sh\",\"-c\",\"echo $nonce; sleep 20\"]}" "$@")
    _code=${_out%% *}
    [ "$_code" = "201" ] && ok container-create pass "$_code" || { ok container-create fail "$_code"; return 0; }

    _out=$(status_of POST "/containers/$_ctr/start" '' "$@"); _code=${_out%% *}
    [ "$_code" = "204" ] && ok container-start pass "$_code" || ok container-start fail "$_code"

    _logs=$(curl --noproxy '*' --silent --connect-timeout 3 --max-time 15 "$@" \
        "$api/containers/$_ctr/logs?stdout=1&stderr=1&follow=0" 2>/dev/null | tr -d '\0')
    case "$_logs" in
        *"$nonce"*) ok container-runs pass 200 ;;
        *)          ok container-runs fail 200 ;;
    esac

    # An ordinary exec is how every agent tool reaches its sandbox.
    _out=$(status_of POST "/containers/$_ctr/exec" '{"Cmd":["/bin/true"],"AttachStdout":true}' "$@")
    _code=${_out%% *}
    [ "$_code" = "201" ] && ok container-exec pass "$_code" || ok container-exec fail "$_code"

    _out=$(status_of POST "/volumes/create" "{\"Name\":\"$_vol\"}" "$@"); _code=${_out%% *}
    [ "$_code" = "201" ] && ok volume-create pass "$_code" || ok volume-create fail "$_code"

    _out=$(status_of POST "/networks/create" "{\"Name\":\"$_net\",\"Driver\":\"bridge\"}" "$@")
    _code=${_out%% *}
    [ "$_code" = "201" ] && ok network-create pass "$_code" || ok network-create fail "$_code"
}

# remove_positive drops everything group_positive may have made. Runs from the
# exit trap as well, so an interrupted probe leaves the sandbox as it found it.
remove_positive() {
    [ -n "${PENTAGI_TEST_IMAGE:-}" ] || return 0
    for _p in "containers/sandbox-probe-$nonce-run?force=1&v=1" \
              "volumes/sandbox-probe-$nonce-okvol?force=1" \
              "networks/sandbox-probe-$nonce-oknet"; do
        curl --noproxy '*' --silent --connect-timeout 3 --max-time 15 "$@" \
            --output /dev/null --request DELETE "$api/$_p" 2>/dev/null || true
    done
}

# A plugin that denies everything would pass every check above while breaking
# every agent, so one ordinary request must get through authorization and fail
# at the image lookup instead.
group_control() {
    _out=$(status_of POST "/containers/create?name=sandbox-probe-$nonce-control" \
        "{\"Image\":\"sandbox-probe-$nonce:absent\",\"Cmd\":[\"true\"],\"HostConfig\":{}}" "$@")
    _code=${_out%% *}; _body=${_out#* }
    case "$_code:$_body" in
        404:*no?such?image*|404:*not?found*) _res=pass ;;
        *) _res=fail ;;
    esac
    emit "{\"kind\":\"check\",\"id\":\"ordinary-create\",\"group\":\"control\",\"expect\":\"allow\",\"result\":\"$_res\",\"status\":$_code}"
}

# The groups touch nothing shared, so they run at once; each writes its own file
# so a half-written line can never reach the caller.
trap 'remove_positive "$@"; rm -rf "$work"' EXIT

group_create   "$@" > "$work/1-create.jsonl" &
group_endpoint "$@" > "$work/2-endpoint.jsonl" &
group_control  "$@" > "$work/3-control.jsonl" &
group_positive "$@" > "$work/4-positive.jsonl" &
wait

cat "$work"/*.jsonl
total=$(cat "$work"/*.jsonl | grep -c '"kind":"check"')
failed=$(cat "$work"/*.jsonl | grep -c '"result":"fail"')
emit "{\"kind\":\"summary\",\"total\":$total,\"passed\":$((total-failed)),\"failed\":$failed}"
[ "$failed" -eq 0 ]
