data "google_client_config" "this" {}

data "google_container_engine_versions" "this" {
  location       = data.google_client_config.this.zone
  version_prefix = var.kubernetes_version
}
