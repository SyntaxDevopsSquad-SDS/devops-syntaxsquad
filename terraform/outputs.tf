output "public_ip_address" {
  value = azurerm_public_ip.whoknows.ip_address
}

output "ssh_command" {
  value = "ssh ${azurerm_linux_virtual_machine.whoknows.admin_username}@${azurerm_public_ip.whoknows.ip_address}"
}