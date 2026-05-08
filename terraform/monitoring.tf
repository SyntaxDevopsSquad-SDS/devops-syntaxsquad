provider "digitalocean" {
  token = var.do_token
}

# DigitalOcean Droplet til Prometheus + Grafana + Watchdog
# Forudsætning: SSH-nøgle skal være registreret i DigitalOcean
#   1. Gå til digitalocean.com → Settings → Security → SSH Keys
#   2. Tilføj indholdet af ~/.ssh/id_rsa.pub
#   3. Kopiér fingerprint og sæt do_ssh_key_fingerprint i terraform.tfvars
resource "digitalocean_droplet" "monitoring" {
  name     = "whoknows-monitoring"
  region   = "ams3"
  size     = "s-1vcpu-2gb"
  image    = "ubuntu-22-04-x64"
  ssh_keys = [var.do_ssh_key_fingerprint]
}

# Persistent volume til Prometheus data — oprettet via terraform/scripts/bootstrap-do-volume.sh
# Lever UDENFOR Terraform, så det overlever 'terraform destroy'.
data "digitalocean_volume" "prometheus_data" {
  name   = "whoknows-prometheus-data"
  region = "ams3"
}

resource "digitalocean_volume_attachment" "prometheus" {
  droplet_id = digitalocean_droplet.monitoring.id
  volume_id  = data.digitalocean_volume.prometheus_data.id
}

resource "local_file" "monitoring_inventory" {
  content  = "[monitoring]\n${digitalocean_droplet.monitoring.ipv4_address} ansible_user=root ansible_ssh_private_key_file=~/.ssh/id_rsa ansible_ssh_common_args='-o StrictHostKeyChecking=no'\n"
  filename = "${path.module}/ansible/monitoring-inventory.ini"
}

resource "cloudflare_record" "monitoring" {
  zone_id         = var.cloudflare_zone_id
  name            = "monitor"
  content         = digitalocean_droplet.monitoring.ipv4_address
  type            = "A"
  ttl             = 1
  proxied         = true
  allow_overwrite = true
}
