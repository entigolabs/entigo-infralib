module "gke_node_pool" {
  source  = "terraform-google-modules/kubernetes-engine/google//modules/gke-node-pool"
  version = "41.0.2"

  name = substr(var.prefix, 0, 40)
  cluster = var.cluster_name
  project_id = data.google_client_config.this.project
  kubernetes_version = data.google_container_engine_versions.this.release_channel_latest_version["STABLE"]

  location = var.cluster_region
  node_locations = length(var.node_locations) > 0 ? var.node_locations : data.google_compute_zones.this.names
  initial_node_count = var.initial_size
  max_pods_per_node = var.max_pods

  management = {
    auto_repair  = true
    auto_upgrade = false
  }

  autoscaling = {
    min_node_count       = var.min_size
    max_node_count       = var.max_size
    total_min_node_count = var.total_min_size
    total_max_node_count = var.total_max_size
    location_policy      = var.location_policy
  }

  # network_config = {}

  node_config = {
    disk_size_gb    = var.volume_size
    disk_type       = var.volume_type
    image_type      = "COS_CONTAINERD"
    machine_type    = var.instance_type
    spot            = var.spot_nodes
    service_account = var.service_account
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
    taint = [
      {
        key    = "spot"
        value  = "true"
        effect = "NO_SCHEDULE"
      }
    ]
    tags = []
    labels = {
      created-by = "entigo-infralib"
      node-pool  = "spot"
    }
  }

  # placement_policy = {}

  # queued_provisioning	= {}

  timeouts = {
    create = "45m"
    update = "45m"
    delete = "45m"
  }

  # upgrade_settings = {}

}
