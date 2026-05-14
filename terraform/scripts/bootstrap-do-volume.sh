#!/bin/bash
# =============================================================================
# bootstrap-do-volume.sh
# Opretter persistent DigitalOcean Volume til Prometheus data.
# Køres KUN én gang — volumet lever udenfor Terraform og overlever destroy.
#
# Forudsætning: doctl er installeret og autentificeret (doctl auth init)
# =============================================================================
set -e

VOLUME_NAME="whoknows-prometheus-data"
REGION="ams3"
SIZE_GIB=20

echo "========================================"
echo "  Bootstrap: DigitalOcean Volume"
echo "========================================"
echo "  Volume Name : $VOLUME_NAME"
echo "  Region      : $REGION"
echo "  Size        : ${SIZE_GIB} GiB"
echo ""

# ── Tjek om doctl er tilgængelig ─────────────────────────────────
if ! command -v doctl &> /dev/null; then
  echo "ERROR: doctl ikke fundet."
  echo "  Installer: https://docs.digitalocean.com/reference/doctl/how-to/install/"
  echo "  Autentificer: doctl auth init"
  exit 1
fi

# ── Tjek om volume allerede eksisterer ───────────────────────────
EXISTING=$(doctl compute volume list --format Name --no-header 2>/dev/null | grep -w "$VOLUME_NAME" || true)
if [[ -n "$EXISTING" ]]; then
  echo "Volume '$VOLUME_NAME' eksisterer allerede — springer over."
  doctl compute volume list --format ID,Name,SizeGigaBytes,Region --no-header | grep "$VOLUME_NAME"
  exit 0
fi

# ── Opret volume ──────────────────────────────────────────────────
echo "Creating DigitalOcean Volume: $VOLUME_NAME..."
doctl compute volume create "$VOLUME_NAME" \
  --region "$REGION" \
  --size "$SIZE_GIB"

echo ""
echo "========================================"
echo "  Done! Volume detaljer:"
doctl compute volume list \
  --format ID,Name,SizeGigaBytes,Region \
  --no-header | grep "$VOLUME_NAME"
echo ""
echo "  Volumet refereres automatisk i monitoring.tf via:"
echo "  data \"digitalocean_volume\" \"prometheus_data\" {"
echo "    name   = \"$VOLUME_NAME\""
echo "    region = \"$REGION\""
echo "  }"
echo "========================================"
