# variables.tf
variable "subscription_id" {
  description = "The azure subscription ID"
  type        = string
}
variable "vm_name" {
  description = "The name of the virtual machine"
  type        = string
  default     = "main-vm"
}
variable "cloudflare_api_token" {
  description = "Cloudflare API token med DNS edit rettigheder"
  type        = string
  sensitive   = true
}
variable "cloudflare_zone_id" {
  description = "Cloudflare Zone ID for domænet"
  type        = string
}

variable "do_token" {
  description = "DigitalOcean API token"
  type        = string
  sensitive   = true
}

variable "do_ssh_key_fingerprint" {
  description = "Fingerprint på SSH-nøgle registreret i DigitalOcean (Settings → Security → SSH Keys)"
  type        = string
}