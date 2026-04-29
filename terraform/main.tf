terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "=4.69.0"
    }
  }
}

provider "azurerm" {
  features {}
  subscription_id = var.subscription_id
}

resource "azurerm_resource_group" "whoknows" {
  name     = "whoknows-rg"
  location = "Switzerland North"
}

resource "azurerm_virtual_network" "whoknows" {
  name                = "whoknows-vnet"
  resource_group_name = azurerm_resource_group.whoknows.name
  location            = azurerm_resource_group.whoknows.location
  address_space       = ["10.0.0.0/16"]
}

resource "azurerm_subnet" "whoknows" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.whoknows.name
  virtual_network_name = azurerm_virtual_network.whoknows.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_public_ip" "whoknows" {
  name                = "whoknows-publicip"
  location            = azurerm_resource_group.whoknows.location
  resource_group_name = azurerm_resource_group.whoknows.name
  allocation_method   = "Static"
  sku                 = "Standard"
}

resource "azurerm_network_interface" "whoknows" {
  name                = "whoknows-nic"
  location            = azurerm_resource_group.whoknows.location
  resource_group_name = azurerm_resource_group.whoknows.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.whoknows.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.whoknows.id
  }
}

resource "azurerm_linux_virtual_machine" "whoknows" {
  name                = var.vm_name
  resource_group_name = azurerm_resource_group.whoknows.name
  location            = azurerm_resource_group.whoknows.location
  size                = "Standard_B2ats_v2"
  admin_username      = "adminuser"
  network_interface_ids = [
    azurerm_network_interface.whoknows.id,
  ]
  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }
  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts-gen2"
    version   = "latest"
  }

  disable_password_authentication = true
  admin_ssh_key {
    username   = "adminuser"
    public_key = file("~/.ssh/id_rsa.pub")
  }
  provisioner "remote-exec" {
    inline = split("\n", templatefile("${path.module}/inline_commands.sh", {}))

    connection {
      type        = "ssh"
      user        = "adminuser"
      private_key = file("~/.ssh/id_rsa")
      host        = self.public_ip_address
      timeout     = "2m"
    }
  }

}

resource "azurerm_network_security_group" "whoknows_nsg" {
  name                = "whoknows-nsg"
  location            = azurerm_resource_group.whoknows.location
  resource_group_name = azurerm_resource_group.whoknows.name

  security_rule {
    name                       = "allow-8080"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "8080"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}

resource "azurerm_network_security_rule" "whoknows_ssh_rule" {
  name                        = "SSH"
  priority                    = 1000
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "22"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  network_security_group_name = azurerm_network_security_group.whoknows_nsg.name
  resource_group_name         = azurerm_resource_group.whoknows.name
}

resource "azurerm_subnet_network_security_group_association" "whoknows_assoc" {
  subnet_id                 = azurerm_subnet.whoknows.id
  network_security_group_id = azurerm_network_security_group.whoknows_nsg.id
}

resource "azurerm_network_interface_security_group_association" "whoknows_nic_assoc" {
  network_interface_id      = azurerm_network_interface.whoknows.id
  network_security_group_id = azurerm_network_security_group.whoknows_nsg.id
}
