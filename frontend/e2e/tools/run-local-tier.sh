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
# outside the compose project — `down -v` never removes them. The 9xxxx flow-id
# range (seeded below) makes them unambiguously ours to delete.
remove_e2e_sandboxes() {
    docker ps -aq --filter "name=pentagi-terminal-9" | xargs -r docker rm -f
}

cleanup() {
    remove_e2e_sandboxes
    if [[ "${E2E_KEEP_STACK:-}" != "1" ]]; then
        compose down -v --remove-orphans
    fi
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
