#!/usr/bin/env bash
# Tier-2 runner: branch image + isolated compose stack (coexists with a dev
# stack — own project/ports) + specs/real/** against the live backend.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
export PENTAGI_IMAGE="${PENTAGI_IMAGE:-local/pentagi:e2e}"
export PENTAGI_LISTEN_PORT="${PENTAGI_LISTEN_PORT:-8444}"
export PGVECTOR_LISTEN_PORT="${PGVECTOR_LISTEN_PORT:-5433}"

# --env-file /dev/null: without it compose auto-loads the developer's .env and
# the stack inherits their live provider keys, DOCKER_HOST, and config paths.
compose() {
    docker compose -p pentagi-e2e --env-file /dev/null \
        -f "$REPO_ROOT/docker-compose.yml" \
        -f "$REPO_ROOT/docker-compose.e2e.yml" "$@"
}

# Sandbox containers are spawned by the backend straight on the docker socket,
# outside the compose project — `down -v` never removes them. The anchored
# regex matches only the seeded 9xxxx (5+ digit) flow-id range; an unanchored
# name filter is a substring match and would also remove a dev stack's
# sandboxes for any flow id starting with 9.
remove_e2e_sandboxes() {
    docker ps -aq --filter 'name=^/?pentagi-terminal-9[0-9]{4,}$' | xargs -r docker rm -f
    # Each sandbox also gets a `<container>-data` volume, created on the socket outside the compose
    # project — removing the container leaves it behind, so it needs its own cleanup.
    docker volume ls -q --filter 'name=^pentagi-terminal-9[0-9]{4,}-data$' | xargs -r docker volume rm
}

cleanup() {
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
}
trap cleanup EXIT

if [[ "${E2E_SKIP_BUILD:-}" != "1" ]]; then
    docker build -t "$PENTAGI_IMAGE" "$REPO_ROOT"
fi

remove_e2e_sandboxes
compose down -v --remove-orphans
compose up -d --wait --wait-timeout 240 pentagi mock-llm

# Flow ids drive sandbox container names (pentagi-terminal-<id>); the 9xxxx
# offset keeps them clear of any developer stack sharing this docker daemon.
docker exec pentagi-e2e-pgvector psql -U postgres -d pentagidb \
    -c "ALTER SEQUENCE flows_id_seq RESTART WITH 90001" > /dev/null

cd "$REPO_ROOT/frontend"
rm -f e2e/.auth/user.json
E2E_TIER=local E2E_BASE_URL="https://localhost:${PENTAGI_LISTEN_PORT}" corepack pnpm e2e "$@"
