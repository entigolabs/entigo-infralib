locals {
  node_pool_name = substr(var.prefix, 0, 40)

  latest_stable_version = data.google_container_engine_versions.this.release_channel_latest_version["STABLE"]
  valid_node_versions   = data.google_container_engine_versions.this.valid_node_versions

  # Get node pools only if cluster data was read
  cluster_node_pools                    = var.preserve_kubernetes_version ? try(data.google_container_cluster.this[0].node_pool, []) : []
  existing_node_pool                    = try([for pool in local.cluster_node_pools : pool if pool.name == local.node_pool_name][0], null)
  existing_node_pool_kubernetes_version = try(local.existing_node_pool.version, "")

  # Determine version: preserve existing if valid and flag is true, otherwise use stable
  kubernetes_version = (
    var.preserve_kubernetes_version && local.existing_node_pool_kubernetes_version != "" && contains(local.valid_node_versions, local.existing_node_pool_kubernetes_version)
    ? local.existing_node_pool_kubernetes_version
    : local.latest_stable_version
  )
}

# https://github.com/terraform-google-modules/terraform-google-kubernetes-engine/tree/main/modules/gke-node-pool
module "gke_node_pool" {
  source  = "terraform-google-modules/kubernetes-engine/google//modules/gke-node-pool"
  version = "43.0.0"

  name               = local.node_pool_name
  cluster            = var.cluster_name
  project_id         = data.google_client_config.this.project
  kubernetes_version = local.kubernetes_version

  location           = var.cluster_region
  node_locations     = length(var.node_locations) > 0 ? var.node_locations : data.google_compute_zones.this.names
  max_pods_per_node  = var.max_pods_per_node
  initial_node_count = ceil(var.min_size / (length(var.node_locations) > 0 ? length(var.node_locations) : length(data.google_compute_zones.this.names)))
  node_count         = var.node_count

  management = {
    auto_repair  = var.auto_repair
    auto_upgrade = var.auto_upgrade
  }

  autoscaling = var.autoscaling != null ? var.autoscaling : {
    total_min_node_count = var.min_size
    total_max_node_count = var.max_size
    location_policy      = var.location_policy
  }

  node_config = var.node_config != null ? var.node_config : {
    disk_size_gb      = var.volume_size
    disk_type         = var.volume_type
    image_type        = "COS_CONTAINERD"
    machine_type      = var.instance_type
    spot              = var.spot_nodes
    service_account   = var.service_account
    boot_disk_kms_key = var.boot_disk_kms_key
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    taint  = var.taints
    tags   = var.tags
    labels = merge(var.labels, { created-by = "entigo-infralib" })
  }

  upgrade_settings = var.upgrade_settings != null ? var.upgrade_settings : {
    max_surge       = var.max_surge
    max_unavailable = var.max_unavailable
    strategy        = "SURGE"
  }

  network_config = var.network_config

  placement_policy = var.placement_policy

  queued_provisioning = var.queued_provisioning

  timeouts = var.timeouts
}

resource "google_kms_crypto_key_iam_member" "boot_disk_kms_key_encrypter_decrypter" {
  count         = var.grant_boot_disk_kms_key_access_to_service_account ? 1 : 0
  crypto_key_id = var.boot_disk_kms_key
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:${var.service_account}"
}
