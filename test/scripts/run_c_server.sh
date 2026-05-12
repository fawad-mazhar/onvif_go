#!/usr/bin/env bash
# run_c_server.sh - Build and run the onvif_simple_server C reference
# server (P0.1) for parity-fixture capture.
#
# Usage:
#   test/scripts/run_c_server.sh [build|run|stop|logs]
#
# Environment overrides:
#   C_SRC      Path to the playground C checkout
#              (default: /Users/fawadmazhar/github/codes/oma/playground/onvif_simple_server)
#   IMAGE      Docker image tag (default: onvif-c-reference:latest)
#   CONTAINER  Container name (default: onvif-c-reference)
#   HOST_PORT  Host port to expose (default: 18080; container listens on 8080)
#   CONF       Canonical fixture config
#              (default: <repo>/test/fixtures/config/server.conf)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

C_SRC="${C_SRC:-/Users/fawadmazhar/github/codes/oma/playground/onvif_simple_server}"
IMAGE="${IMAGE:-onvif-c-reference:latest}"
CONTAINER="${CONTAINER:-onvif-c-reference}"
HOST_PORT="${HOST_PORT:-18080}"
CONF="${CONF:-${REPO_ROOT}/test/fixtures/config/server.conf}"
DOCKERFILE="${REPO_ROOT}/test/fixtures/c-reference/Dockerfile"

cmd="${1:-run}"

require_docker() {
    if ! command -v docker >/dev/null 2>&1; then
        echo "ERROR: docker not found in PATH. Install Docker Desktop or equivalent." >&2
        exit 2
    fi
}

case "$cmd" in
    build)
        require_docker
        if [[ ! -d "$C_SRC" ]]; then
            echo "ERROR: C source not found at $C_SRC" >&2
            exit 2
        fi
        echo "Building $IMAGE from $C_SRC"
        docker build -f "$DOCKERFILE" -t "$IMAGE" "$C_SRC"
        ;;

    run)
        require_docker
        if [[ ! -f "$CONF" ]]; then
            echo "ERROR: fixture config not found at $CONF" >&2
            exit 2
        fi
        if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
            echo "Image $IMAGE not built yet; running build first."
            "$0" build
        fi
        # Stop any prior container so the script is idempotent.
        docker rm -f "$CONTAINER" >/dev/null 2>&1 || true
        echo "Starting $CONTAINER on host port $HOST_PORT (config: $CONF)"
        docker run -d --rm \
            --name "$CONTAINER" \
            -p "${HOST_PORT}:8080" \
            -v "${CONF}:/etc/onvif_simple_server.conf:ro" \
            "$IMAGE" >/dev/null
        echo "C reference server: http://localhost:${HOST_PORT}/onvif/device_service"
        ;;

    stop)
        require_docker
        docker rm -f "$CONTAINER" >/dev/null 2>&1 || true
        echo "Stopped $CONTAINER"
        ;;

    logs)
        require_docker
        docker logs -f "$CONTAINER"
        ;;

    *)
        echo "Usage: $0 [build|run|stop|logs]" >&2
        exit 1
        ;;
esac
