data "google_client_config" "this" {}

data "google_project" "this" {}

data "google_kms_key_ring" "kms_key_ring" {
  count    = var.create_kms_key_ring ? 0 : 1
  name     = var.kms_key_ring_name != "" ? var.kms_key_ring_name : var.prefix
  location = data.google_client_config.this.region
}