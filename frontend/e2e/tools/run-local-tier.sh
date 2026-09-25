#!/usr/bin/env bash
# Tier-2 runner: branch image + isolated compose stack (coexists with a dev
# stack — own project/ports) + specs/real/** against the live backend.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
export PENTAGI_IMAGE="${PENTAGI_IMAGE:-local/pentagi:e2e}"
port_is_taken() {
    (exec 3<>"/dev/tcp/127.0.0.1/$1") >/dev/null 2>&1
}

port_for() {
    local variable="$1" port="$2"
    local chosen="${!variable:-}"

    if [ -n "$chosen" ]; then
        if port_is_taken "$chosen"; then
            echo "e2e: port $chosen is already serving something — set $variable to a free port" >&2
            exit 1
        fi
        printf '%s' "$chosen"
        return
    fi

    while port_is_taken "$port"; do
        port=$((port + 1))
    done
    printf '%s' "$port"
}

PENTAGI_LISTEN_PORT="$(port_for PENTAGI_LISTEN_PORT 8444)"
PGVECTOR_LISTEN_PORT="$(port_for PGVECTOR_LISTEN_PORT 5433)"
export PENTAGI_LISTEN_PORT PGVECTOR_LISTEN_PORT

# --env-file /dev/null: without it compose auto-loads the developer's .env and
# the stack inherits their live provider keys, DOCKER_HOST, and config paths.
compose() {
    docker compose -p pentagi-e2e --env-file /dev/null \
        -f "$REPO_ROOT/docker-compose.yml" \
        -f "$REPO_ROOT/docker-compose.e2e.yml" "$@"
}

# Sandbox containers are spawned by the backend straight on the docker socket,
# outside the compose project — `down -v` never removes them, and they carry no
# label saying whose they are. The seeded 9xxxx range is not ours alone either:
# a developer stack that ever ran a tier-2 seed keeps flows with those ids, and
# removing them by name pattern destroys that stack's sandboxes and their data
# volumes. So we take the range as it is before the run, remove only what this run adds — and,
# separately, the sandboxes a previous run of this script recorded as its own and never got to
# remove, because the reseeded ids land straight on their names.
sandbox_names() {
    docker ps -a --format '{{.Names}}' --filter 'name=^/?pentagi-terminal-9[0-9]{4,}$' | sort
}

sandbox_volumes() {
    docker volume ls -q --filter 'name=^pentagi-terminal-9[0-9]{4,}-data$' | sort
}

# What a previous run created and never got to remove: an aborted run, a kept stack, a killed shell.
# Not under test-results/: playwright deletes that directory whole at the start of every run.
OWNED_STATE="${OWNED_STATE:-$REPO_ROOT/frontend/e2e/.tier2-sandboxes}"
OWNED_VOLUME_STATE="${OWNED_VOLUME_STATE:-$REPO_ROOT/frontend/e2e/.tier2-sandbox-volumes}"

remove_owned_leftovers() {
    OWNED_SURVIVORS=""
    OWNED_VOLUME_SURVIVORS=""

    if [[ -f "$OWNED_STATE" ]]; then
        comm -12 <(sort -u "$OWNED_STATE") <(sandbox_names) | xargs -r -n 1 docker rm -f || true
        OWNED_SURVIVORS="$(comm -12 <(sort -u "$OWNED_STATE") <(sandbox_names))"
        rm -f "$OWNED_STATE"
    fi

    if [[ -f "$OWNED_VOLUME_STATE" ]]; then
        # A successor container may legitimately hold the volume by now; leave it to the next run.
        comm -12 <(sort -u "$OWNED_VOLUME_STATE") <(sandbox_volumes) | xargs -r -n 1 docker volume rm || true
        OWNED_VOLUME_SURVIVORS="$(comm -12 <(sort -u "$OWNED_VOLUME_STATE") <(sandbox_volumes))"
        rm -f "$OWNED_VOLUME_STATE"
    fi
}

# The baseline has to be taken after the sweep: the reseed hands this run the same ids the sweep
# just freed, and a name standing on both sides of the diff cancels out of it.
# `set -u` does not guard the two readers below: they expand the baseline inside a process
# substitution, and a subshell that dies there still yields an empty list, which `comm -13` reads as
# "this run created everything" — including a developer's sandbox.
SANDBOX_BASELINE_TAKEN=0

prepare_sandbox_baseline() {
    remove_owned_leftovers

    SANDBOXES_BEFORE="$(sandbox_names)"
    VOLUMES_BEFORE="$(sandbox_volumes)"
    SANDBOX_BASELINE_TAKEN=1
}

record_owned_sandboxes() {
    if [[ "${SANDBOX_BASELINE_TAKEN:-0}" != "1" ]]; then
        return 0
    fi

    mkdir -p "$(dirname "$OWNED_STATE")"
    {
        comm -13 <(printf '%s\n' "$SANDBOXES_BEFORE") <(sandbox_names)
        printf '%s\n' "${OWNED_SURVIVORS:-}"
    } | sed '/^$/d' | sort -u > "$OWNED_STATE"
    {
        comm -13 <(printf '%s\n' "$VOLUMES_BEFORE") <(sandbox_volumes)
        printf '%s\n' "${OWNED_VOLUME_SURVIVORS:-}"
    } | sed '/^$/d' | sort -u > "$OWNED_VOLUME_STATE"
}

remove_e2e_sandboxes() {
    if [[ "${SANDBOX_BASELINE_TAKEN:-0}" != "1" ]]; then
        echo "sandbox baseline was never taken; refusing to sweep" >&2

        return 0
    fi

    comm -13 <(printf '%s\n' "$SANDBOXES_BEFORE") <(sandbox_names) | xargs -r docker rm -f
    # Each sandbox also gets a `<container>-data` volume, created on the socket outside the compose
    # project — removing the container leaves it behind, so it needs its own cleanup.
    comm -13 <(printf '%s\n' "$VOLUMES_BEFORE") <(sandbox_volumes) | xargs -r docker volume rm
}

cleanup() {
    # Before the early return: a kept stack is the case whose sandboxes the next run must clear.
    record_owned_sandboxes || true

    if [[ "${E2E_KEEP_STACK:-}" == "1" ]]; then
        return
    fi
    # The backend/mock-LLM logs are the only record of the agent loop — capture
    # them before `down -v` destroys the containers, so a red CI run has more
    # than a trace of a stuck UI.
    mkdir -p "$REPO_ROOT/frontend/e2e/test-results"
    compose logs --no-color pentagi mock-llm \
        > "$REPO_ROOT/frontend/e2e/test-results/compose.log" 2>&1 || true
    # The trap is the script's last statement, so an unguarded teardown failure
    # would become the exit status and turn a passing run red.
    remove_e2e_sandboxes || true
    compose down -v --remove-orphans || true
    # Record again over the teardown: the first pass named what this run created, and the next run
    # treats that file as a kill list. Anything the removal above actually took must leave it, or
    # the next run deletes a developer's sandbox that inherited the name.
    record_owned_sandboxes || true
}
prepare_sandbox_baseline
trap cleanup EXIT

if [[ "${E2E_SKIP_BUILD:-}" != "1" ]]; then
    docker build -t "$PENTAGI_IMAGE" "$REPO_ROOT"
fi

compose down -v --remove-orphans
compose up -d --wait --wait-timeout 240 pentagi mock-llm

# Flow ids drive sandbox container names (pentagi-terminal-<id>); the 9xxxx
# offset keeps them clear of any developer stack sharing this docker daemon.
docker exec pentagi-e2e-pgvector psql -U postgres -d pentagidb \
    -c "ALTER SEQUENCE flows_id_seq RESTART WITH 90001" > /dev/null

cd "$REPO_ROOT/frontend"
rm -f e2e/.auth/user.json
E2E_TIER=local E2E_BASE_URL="https://localhost:${PENTAGI_LISTEN_PORT}" corepack pnpm e2e "$@"
