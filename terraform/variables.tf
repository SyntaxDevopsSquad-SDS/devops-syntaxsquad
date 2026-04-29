# variables.tf
variable "db_admin" {
  description = "PostgreSQL admin brugernavn"
  type        = string
}

variable "db_password" {
  description = "PostgreSQL admin password"
  type        = string
  sensitive   = true
}
variable "subscription_id" {
  description = "The azure subscription ID"
  type        = string
}
variable "vm_name" {
  description = "The name of the virtual machine"
  type        = string
  default     = "main-vm"
}