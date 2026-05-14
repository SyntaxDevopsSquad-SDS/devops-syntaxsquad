#!/bin/bash
# =============================================================================
# bootstrap-tfstate.sh
# Opretter Azure Storage Account til Terraform remote state (tfstate).
# Køres KUN én gang — FØR første 'terraform init'.
#
# Terraform backend konfiguration (main.tf):
#   backend "azurerm" {
#     resource_group_name  = "tfstate-rg"
#     storage_account_name = "whoknowstfstate"
#     container_name       = "tfstate"
#     key                  = "whoknows.terraform.tfstate"
#   }
# =============================================================================
set -e

RESOURCE_GROUP="tfstate-rg"
STORAGE_ACCOUNT="whoknowstfstate"
CONTAINER="tfstate"
LOCATION="norwayeast"

# ── Subscription ID ───────────────────────────────────────────────
SUBSCRIPTION_ID="${1:-$SUBSCRIPTION_ID}"
if [[ -z "$SUBSCRIPTION_ID" ]]; then
  echo "Usage: $0 <subscription-id>"
  echo "       OR: export SUBSCRIPTION_ID=... && bash $0"
  exit 1
fi

echo "========================================"
echo "  Bootstrap: Terraform Remote State"
echo "========================================"
echo "  Subscription    : $SUBSCRIPTION_ID"
echo "  Resource Group  : $RESOURCE_GROUP"
echo "  Storage Account : $STORAGE_ACCOUNT"
echo "  Container       : $CONTAINER"
echo "  Location        : $LOCATION"
echo ""

# ── Opret resource group ──────────────────────────────────────────
echo "Creating resource group: $RESOURCE_GROUP..."
az group create \
  --name "$RESOURCE_GROUP" \
  --location "$LOCATION" \
  --subscription "$SUBSCRIPTION_ID" \
  --output table

# ── Opret storage account ─────────────────────────────────────────
echo ""
echo "Creating storage account: $STORAGE_ACCOUNT..."
az storage account create \
  --name "$STORAGE_ACCOUNT" \
  --resource-group "$RESOURCE_GROUP" \
  --location "$LOCATION" \
  --sku Standard_LRS \
  --kind StorageV2 \
  --allow-blob-public-access false \
  --subscription "$SUBSCRIPTION_ID" \
  --output table

# ── Opret blob container ──────────────────────────────────────────
echo ""
echo "Creating blob container: $CONTAINER..."
az storage container create \
  --name "$CONTAINER" \
  --account-name "$STORAGE_ACCOUNT" \
  --auth-mode login \
  --output table

echo ""
echo "========================================"
echo "  Done! Terraform remote state klar."
echo ""
echo "  Kør nu:"
echo "    cd terraform"
echo "    terraform init"
echo "========================================"
