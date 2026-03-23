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

generate_secret_key() {
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -hex 32
    else
        head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n'
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

is_port_8080_in_use() {
    sudo ss -ltn | grep -q ':8080 '
}

ensure_port_8080_available() {
    if ! is_port_8080_in_use; then
        return
    fi

    echo "Port 8080 is in use. Attempting to stop Docker containers publishing 8080..."
    CONTAINERS="$(sudo docker ps --filter publish=8080 --format '{{.ID}}')"
    if [ -n "$CONTAINERS" ]; then
        sudo docker stop $CONTAINERS
        sudo docker rm $CONTAINERS || true
    fi

    if is_port_8080_in_use; then
        echo "Port 8080 still busy. Attempting to stop non-Docker listener processes..."
        PIDS="$(sudo ss -ltnp | grep ':8080 ' | grep -o 'pid=[0-9]\+' | cut -d= -f2 | sort -u)"
        if [ -n "$PIDS" ]; then
            sudo kill $PIDS || true
            sleep 2
            if is_port_8080_in_use; then
                sudo kill -9 $PIDS || true
                sleep 1
            fi
        fi
    fi

    if is_port_8080_in_use; then
        echo "Port 8080 is still in use after attempts to free it. Aborting migration."
        exit 1
    fi
}

ensure_root_env() {
    sudo mkdir -p "$APP_DIR"

    if [ ! -f "$ENV_FILE" ]; then
        echo "Creating $ENV_FILE with generated SECRET_KEY"
        SECRET_VALUE="$(generate_secret_key)"
        printf "SECRET_KEY=%s\n" "$SECRET_VALUE" | sudo tee "$ENV_FILE" >/dev/null
        sudo chmod 600 "$ENV_FILE"
        return
    fi

    if ! sudo grep -q '^SECRET_KEY=' "$ENV_FILE"; then
        echo "SECRET_KEY missing in $ENV_FILE. Appending generated SECRET_KEY"
        SECRET_VALUE="$(generate_secret_key)"
        printf "SECRET_KEY=%s\n" "$SECRET_VALUE" | sudo tee -a "$ENV_FILE" >/dev/null
        sudo chmod 600 "$ENV_FILE"
    fi
}

require_cmd sudo
require_cmd date
require_cmd cp
ensure_docker
ensure_docker_compose
login_ghcr_if_configured
ensure_root_env

if [ -z "$IMAGE_NAME" ]; then
    echo "Usage: bash implementations/go/scripts/migration.sh <IMAGE_NAME> [IMAGE_TAG]"
    echo "Example: bash implementations/go/scripts/migration.sh ghcr.io/syntaxdevopssquad-sds/whoknows-go 0123abcd"
    exit 1
fi

if [ ! -d "$APP_DIR" ]; then
    echo "App directory not found: $APP_DIR"
    exit 1
fi

if ! sudo grep -q '^SECRET_KEY=' "$ENV_FILE"; then
    echo "Missing SECRET_KEY in $ENV_FILE"
    exit 1
fi

if sudo test -f "$PROD_COMPOSE_FILE"; then
    sudo cp -f "$PROD_COMPOSE_FILE" "$COMPOSE_FILE"
fi

if ! sudo test -f "$COMPOSE_FILE"; then
    echo "Missing compose file: $COMPOSE_FILE"
    exit 1
fi

echo "=== Preparing migration to Docker Compose ==="
cd "$APP_DIR"

if sudo test -f "$DB_FILE"; then
    BACKUP_FILE="$BACKEND_DIR/whoknows.db.bak.$(date +%F-%H%M%S)"
    echo "Creating DB backup: $BACKUP_FILE"
    sudo cp "$DB_FILE" "$BACKUP_FILE"
else
    echo "No existing DB file found at $DB_FILE (will initialize if needed)"
fi

echo "Pulling image before cutover (service stays online)..."
sudo IMAGE_NAME="$IMAGE_NAME" IMAGE_TAG="$IMAGE_TAG" docker compose pull

echo "Stopping old systemd service..."
if sudo systemctl list-unit-files | grep -q "^${SERVICE_NAME}\.service"; then
    sudo systemctl stop "$SERVICE_NAME"
    SYSTEMD_STOPPED=1
else
    echo "Systemd service ${SERVICE_NAME}.service not found; skipping stop step"
fi

ensure_port_8080_available

echo "Starting Docker Compose service..."
sudo IMAGE_NAME="$IMAGE_NAME" IMAGE_TAG="$IMAGE_TAG" docker compose up -d --remove-orphans

sleep 3

if ! sudo docker compose ps --status running | grep -q "whoknows"; then
    echo "Compose service is not running as expected"
    exit 1
fi

echo "=== Migration completed successfully ==="
sudo docker compose ps

trap - ERR
