data "google_client_config" "this" {}

data "google_kms_key_ring" "this" {
  count    = var.create_key_ring ? 0 : 1
  name     = var.key_ring_name != "" ? var.key_ring_name : var.prefix
  location = data.google_client_config.this.region
}