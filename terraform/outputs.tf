output "public_ip_address" {
  value = azurerm_public_ip.whoknows.ip_address
}

output "ssh_command" {
  value = "ssh ${azurerm_linux_virtual_machine.whoknows.admin_username}@${azurerm_public_ip.whoknows.ip_address}"
}

output "monitoring_ip" {
  value = digitalocean_droplet.monitoring.ipv4_address
}

output "grafana_url" {
  value = "http://monitor.syntax-reborndev.com"
}

output "monitoring_ssh" {
  value = "ssh root@${digitalocean_droplet.monitoring.ipv4_address}"
}