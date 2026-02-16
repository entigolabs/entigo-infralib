data "google_client_config" "this" {}

data "google_compute_zones" "this" {}

data "google_container_engine_versions" "this" {
  location       = var.cluster_region
  version_prefix = var.kubernetes_version
}

data "google_container_cluster" "this" {
  count    = var.preserve_kubernetes_version ? 1 : 0
  name     = var.cluster_name
  location = var.cluster_region
  project  = data.google_client_config.this.project
}
