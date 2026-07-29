#!/bin/bash
# Docker API client wrapper for the DinD Unix socket.
# Usage: docker-dind-sock [docker-commands]

DIND_ENV="${DIND_ENV:-/etc/docker/dind/dind.env}"
if [ -r "${DIND_ENV}" ]; then
    . "${DIND_ENV}"
fi

RUN_PATH="${RUN_PATH:-/var/run/docker-dind}"

export DOCKER_HOST="unix://${RUN_PATH}/docker.sock"
export DOCKER_TLS_VERIFY=
export DOCKER_CERT_PATH=

# Show connection info if no arguments provided
if [ $# -eq 0 ]; then
    echo "Docker dind socket connection configured:"
    echo "  Host: ${DOCKER_HOST}"
    echo ""
    echo "Usage: docker-dind-sock [docker-commands]"
    echo "Examples:"
    echo "  docker-dind-sock version"
    echo "  docker-dind-sock ps"
    echo "  docker-dind-sock images"
    exit 0
fi

exec docker "$@"
