#!/bin/bash
set -euo pipefail

APP_DIR="/opt/whoknows"
COMPOSE_FILE="$APP_DIR/docker-compose.yml"

IMAGE_NAME="${1:-${IMAGE_NAME:-}}"
IMAGE_TAG="${2:-${IMAGE_TAG:-latest}}"

if [[ -z "$IMAGE_NAME" ]]; then
    echo "Usage: bash implementations/go/scripts/deploy_compose.sh <IMAGE_NAME> [IMAGE_TAG]"
    exit 1
fi

if [[ ! -d "$APP_DIR" ]]; then
    echo "Missing app dir: $APP_DIR"
    exit 1
fi

if ! sudo test -f "$COMPOSE_FILE"; then
    echo "Missing compose file: $COMPOSE_FILE"
    exit 1
fi

if [[ -n "${GHCR_USER:-}" ]] && [[ -n "${GHCR_PAT:-}" ]]; then
    echo "Logging in to GHCR..."
    echo "$GHCR_PAT" | sudo docker login ghcr.io -u "$GHCR_USER" --password-stdin
fi

cd "$APP_DIR"

echo "=== Pulling image ==="
sudo IMAGE_NAME="$IMAGE_NAME" IMAGE_TAG="$IMAGE_TAG" docker compose pull

echo "=== Stopping existing containers ==="
sudo docker compose down

echo "=== Updating service ==="
sudo IMAGE_NAME="$IMAGE_NAME" IMAGE_TAG="$IMAGE_TAG" docker compose up -d
sleep 2

if ! sudo docker compose ps --status running | grep -q "whoknows"; then
    echo "Deploy failed: whoknows is not running"
    sudo docker compose ps
    exit 1
fi

echo "=== Deploy completed ==="
sudo docker compose ps
