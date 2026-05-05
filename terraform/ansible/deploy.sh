#!/bin/bash

echo "🚀 Opretter infrastruktur med Terraform..."
cd "$(dirname "$0")/.."
terraform init
terraform apply -auto-approve

echo "⚙️ Konfigurerer VM med Ansible..."
cd ansible
perl -pi -e 's/\r//' inventory.ini
sleep 30
ansible-playbook -i inventory.ini playbook.yml

echo "✅ Done!"
echo "IP: $(terraform -chdir=.. output -raw public_ip_address)"
echo "SSH: $(terraform -chdir=.. output -raw ssh_command)"
echo "🌐 Nginx: http://$(terraform -chdir=.. output -raw public_ip_address)"