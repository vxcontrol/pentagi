#!/bin/bash
# Docker API client wrapper for the DinD TLS endpoint.
# Usage: docker-dind-tls [docker-commands]

DIND_ENV="${DIND_ENV:-/etc/docker/dind/dind.env}"
if [ -r "${DIND_ENV}" ]; then
    . "${DIND_ENV}"
fi

API_ADDRESS="${API_ADDRESS:-0.0.0.0}"
DOCKER_PORT="${DOCKER_PORT:-3376}"
CERTS_PATH="${CERTS_PATH:-/etc/docker/dind/certs}"

# 0.0.0.0 is a bind address, not a destination: connect over loopback instead.
if [ "${API_ADDRESS}" = "0.0.0.0" ]; then
    API_ADDRESS=127.0.0.1
fi

export DOCKER_HOST="tcp://${API_ADDRESS}:${DOCKER_PORT}"
export DOCKER_TLS_VERIFY=1
export DOCKER_CERT_PATH="${CERTS_PATH}/client"

# Show connection info if no arguments provided
if [ $# -eq 0 ]; then
    echo "Docker dind API connection configured:"
    echo "  Host: ${API_ADDRESS}:${DOCKER_PORT}"
    echo "  TLS: enabled"
    echo "  Certificates: ${CERTS_PATH}/client/"
    echo ""
    echo "Usage: docker-dind-tls [docker-commands]"
    echo "Examples:"
    echo "  docker-dind-tls version"
    echo "  docker-dind-tls ps"
    echo "  docker-dind-tls images"
    exit 0
fi

exec docker "$@"
