resource "google_service_account" "service_account" {
  account_id   = local.hname
  display_name = local.hname
}

module "gke" {
  source = "terraform-google-modules/kubernetes-engine/google//modules/private-cluster"
  version = "31.0.0"

  project_id             = data.google_client_config.this.project
  name                   = local.hname
  kubernetes_version     = var.kubernetes_version
  release_channel        = "UNSPECIFIED" # in order to disable auto upgrade
  region                 = data.google_client_config.this.region
  network                = var.network
  subnetwork             = var.subnetwork
  master_ipv4_cidr_block = var.master_ipv4_cidr_block
  ip_range_pods          = var.ip_range_pods
  ip_range_services      = var.ip_range_services

  service_account                 = google_service_account.service_account.email
  master_global_access_enabled    = false
  #istio                           = false //only in beta module
  issue_client_certificate        = false
  enable_private_endpoint         = false
  enable_private_nodes            = true
  remove_default_node_pool        = true
  enable_shielded_nodes           = false
  identity_namespace              = "enabled"
  node_metadata                   = "GKE_METADATA"
  horizontal_pod_autoscaling      = true
  enable_vertical_pod_autoscaling = false
  deletion_protection             = false

  node_pools                      = [
        {
            name               = "main"
            machine_type       = var.machine_type
            node_locations     = data.google_client_config.this.zone
            initial_node_count = 1
            min_count          = 1
            max_count          = 2
            max_pods_per_node  = 64
            disk_size_gb       = 10
            disk_type          = "pd-standard"
            image_type         = "COS_CONTAINERD"
            auto_repair        = true
            auto_upgrade       = false
            spot               = false

        }
    ]

  node_pools_oauth_scopes = {
    all = [
        "https://www.googleapis.com/auth/monitoring",
        "https://www.googleapis.com/auth/compute",
        "https://www.googleapis.com/auth/devstorage.full_control",
        "https://www.googleapis.com/auth/logging.write",
        "https://www.googleapis.com/auth/service.management",
        "https://www.googleapis.com/auth/servicecontrol",
    ]
  }

  node_pools_labels = {
    all = {}
  }

  node_pools_tags = {
    all = ["k8s-nodes"]
  }

  node_pools_metadata = {
    all = {
      disable-legacy-endpoints = "true"
    }
  }

  node_pools_taints = {
    all = []
  }

  master_authorized_networks = var.master_authorized_networks
}
