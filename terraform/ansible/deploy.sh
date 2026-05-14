#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR/.."

# Indlæs lokale hemmeligheder (GRAFANA_PASSWORD m.fl.)
ENV_FILE="$SCRIPT_DIR/../../.env"
if [[ -f "$ENV_FILE" ]]; then
  set -a; source "$ENV_FILE"; set +a
fi

echo "🚀 Opretter infrastruktur med Terraform..."
terraform init
terraform apply -auto-approve

APP_IP=$(terraform output -raw public_ip_address)
MONITORING_IP=$(terraform output -raw monitoring_ip)

echo ""
echo "⚙️  Konfigurerer app VM med Ansible..."
cd ansible
perl -pi -e 's/\r//' inventory.ini
ansible-playbook -i inventory.ini playbook.yml

echo ""
echo "📊 Konfigurerer monitoring VM med Ansible..."
perl -pi -e 's/\r//' monitoring-inventory.ini
ansible-playbook -i monitoring-inventory.ini monitoring-playbook.yml \
  -e "app_ip=$APP_IP" \
  -e "grafana_password=${GRAFANA_PASSWORD:-admin}" \
  -e "discord_webhook_url=${DISCORD_WEBHOOK_URL:-}"

echo ""
echo "✅ Done!"
echo "App IP:    $APP_IP"
echo "App SSH:   $(terraform -chdir=.. output -raw ssh_command)"
echo "App URL:   http://$APP_IP"
echo ""
echo "Monitoring IP:  $MONITORING_IP"
echo "Grafana URL:    $(terraform -chdir=.. output -raw grafana_url)"
echo "Monitoring SSH: $(terraform -chdir=.. output -raw monitoring_ssh)"

echo "Monitoring SSH: $(terraform -chdir=.. output -raw monitoring_ssh)"