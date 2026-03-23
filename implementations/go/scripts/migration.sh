#!/bin/bash
set -euo pipefail

APP_DIR="/opt/whoknows"
BACKEND_DIR="$APP_DIR/implementations/go/backend"
COMPOSE_FILE="$APP_DIR/docker-compose.yml"
PROD_COMPOSE_FILE="$APP_DIR/docker-compose.prod.yml"
ENV_FILE="$APP_DIR/.env"
DB_FILE="$BACKEND_DIR/whoknows.db"
SERVICE_NAME="whoknows"
SYSTEMD_STOPPED=0

IMAGE_NAME="${1:-${IMAGE_NAME:-}}"
IMAGE_TAG="${2:-${IMAGE_TAG:-latest}}"

rollback() {
    echo ""
    echo "Migration failed. Attempting rollback..."

    if [ "$SYSTEMD_STOPPED" -eq 1 ]; then
        sudo systemctl start "$SERVICE_NAME" || true
        sudo systemctl status "$SERVICE_NAME" --no-pager || true
    fi

    echo "Rollback done (best effort)."
}

trap rollback ERR

require_cmd() {
    if ! command -v "$1" >/dev/null 2>&1; then
        echo "Missing required command: $1"
        exit 1
    fi
}

ensure_docker_compose() {
    if sudo docker compose version >/dev/null 2>&1; then
        return
    fi

    echo "Docker Compose plugin not found. Attempting installation..."
    if command -v apt-get >/dev/null 2>&1; then
        sudo apt-get update
        if ! sudo apt-get install -y docker-compose-plugin; then
            sudo apt-get install -y docker-compose-v2
        fi
    else
        echo "Could not auto-install docker compose (apt-get not available)."
        echo "Please install Docker Compose plugin manually and re-run."
        exit 1
    fi

    if ! sudo docker compose version >/dev/null 2>&1; then
        echo "Docker Compose installation failed."
        exit 1
    fi
}

ensure_docker() {
    if command -v docker >/dev/null 2>&1; then
        sudo systemctl enable docker >/dev/null 2>&1 || true
        sudo systemctl start docker >/dev/null 2>&1 || true
        return
    fi

    echo "Docker not found. Attempting installation..."
    if command -v apt-get >/dev/null 2>&1; then
        sudo apt-get update
        if ! sudo apt-get install -y docker.io docker-compose-plugin; then
            sudo apt-get install -y docker.io docker-compose-v2
        fi
        sudo systemctl enable docker
        sudo systemctl start docker
    else
        echo "Could not auto-install Docker (apt-get not available)."
        echo "Please install Docker Engine manually and re-run."
        exit 1
    fi

    if ! command -v docker >/dev/null 2>&1; then
        echo "Docker installation failed."
        exit 1
    fi
}

login_ghcr_if_configured() {
    if [ -n "${GHCR_USER:-}" ] && [ -n "${GHCR_PAT:-}" ]; then
        echo "Logging in to GHCR..."
        echo "$GHCR_PAT" | sudo docker login ghcr.io -u "$GHCR_USER" --password-stdin
    else
        echo "GHCR credentials not provided; proceeding without docker registry login"
    fi
}

require_cmd sudo
require_cmd date
require_cmd cp
ensure_docker
ensure_docker_compose
login_ghcr_if_configured

if [ -z "$IMAGE_NAME" ]; then
    echo "Usage: bash implementations/go/scripts/migration.sh <IMAGE_NAME> [IMAGE_TAG]"
    echo "Example: bash implementations/go/scripts/migration.sh ghcr.io/syntaxdevopssquad-sds/whoknows-go 0123abcd"
    exit 1
fi

if [ ! -d "$APP_DIR" ]; then
    echo "App directory not found: $APP_DIR"
    exit 1
fi

if [ ! -f "$ENV_FILE" ]; then
    echo "Missing $ENV_FILE (must contain SECRET_KEY=...)"
    exit 1
fi

if ! grep -q '^SECRET_KEY=' "$ENV_FILE"; then
    echo "Missing SECRET_KEY in $ENV_FILE"
    exit 1
fi

if [ -f "$PROD_COMPOSE_FILE" ]; then
    cp -f "$PROD_COMPOSE_FILE" "$COMPOSE_FILE"
fi

if [ ! -f "$COMPOSE_FILE" ]; then
    echo "Missing compose file: $COMPOSE_FILE"
    exit 1
fi

echo "=== Preparing migration to Docker Compose ==="
cd "$APP_DIR"

if [ -f "$DB_FILE" ]; then
    BACKUP_FILE="$BACKEND_DIR/whoknows.db.bak.$(date +%F-%H%M%S)"
    echo "Creating DB backup: $BACKUP_FILE"
    cp "$DB_FILE" "$BACKUP_FILE"
else
    echo "No existing DB file found at $DB_FILE (will initialize if needed)"
fi

echo "Pulling image before cutover (service stays online)..."
IMAGE_NAME="$IMAGE_NAME" IMAGE_TAG="$IMAGE_TAG" sudo docker compose pull

echo "Stopping old systemd service..."
if sudo systemctl list-unit-files | grep -q "^${SERVICE_NAME}\.service"; then
    sudo systemctl stop "$SERVICE_NAME"
    SYSTEMD_STOPPED=1
else
    echo "Systemd service ${SERVICE_NAME}.service not found; skipping stop step"
fi

echo "Starting Docker Compose service..."
IMAGE_NAME="$IMAGE_NAME" IMAGE_TAG="$IMAGE_TAG" sudo docker compose up -d --remove-orphans

sleep 3

if ! sudo docker compose ps --status running | grep -q "whoknows"; then
    echo "Compose service is not running as expected"
    exit 1
fi

echo "=== Migration completed successfully ==="
sudo docker compose ps

trap - ERR
