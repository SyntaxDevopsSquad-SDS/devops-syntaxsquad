#!/bin/bash
# =============================================================================
# bootstrap-disk.sh
# Opretter persistent Azure Managed Disk til PostgreSQL data.
# Køres KUN én gang — disken lever udenfor Terraform og overlever destroy.
# =============================================================================
set -e

RESOURCE_GROUP="whoknows-data-rg"
DISK_NAME="whoknows-postgres-data"
LOCATION="norwayeast"
DISK_SIZE_GB=32
SKU="Standard_LRS"

# ── Subscription ID ───────────────────────────────────────────────
SUBSCRIPTION_ID="${1:-$SUBSCRIPTION_ID}"
if [[ -z "$SUBSCRIPTION_ID" ]]; then
  echo "Usage: $0 <subscription-id>"
  echo "       OR: export SUBSCRIPTION_ID=... && bash $0"
  exit 1
fi

echo "========================================"
echo "  Bootstrap: Azure Managed Disk"
echo "========================================"
echo "  Subscription : $SUBSCRIPTION_ID"
echo "  Resource Group: $RESOURCE_GROUP"
echo "  Disk Name    : $DISK_NAME"
echo "  Location     : $LOCATION"
echo "  Size         : ${DISK_SIZE_GB} GB  ($SKU)"
echo ""

# ── Opret resource group (idempotent) ─────────────────────────────
echo "Creating resource group: $RESOURCE_GROUP..."
az group create \
  --name "$RESOURCE_GROUP" \
  --location "$LOCATION" \
  --subscription "$SUBSCRIPTION_ID" \
  --output table

# ── Opret Managed Disk (fejler hvis den allerede eksisterer) ──────
echo ""
echo "Creating Managed Disk: $DISK_NAME..."
az disk create \
  --resource-group "$RESOURCE_GROUP" \
  --name "$DISK_NAME" \
  --location "$LOCATION" \
  --size-gb "$DISK_SIZE_GB" \
  --sku "$SKU" \
  --subscription "$SUBSCRIPTION_ID" \
  --output table

# ── Udskriv disk ID til brug i terraform ─────────────────────────
DISK_ID=$(az disk show \
  --resource-group "$RESOURCE_GROUP" \
  --name "$DISK_NAME" \
  --query id \
  --output tsv \
  --subscription "$SUBSCRIPTION_ID")

echo ""
echo "========================================"
echo "  Done! Disk ID:"
echo "  $DISK_ID"
echo ""
echo "  Denne disk refereres i main.tf:"
echo "  managed_disk_id = \"$DISK_ID\""
echo "========================================"
