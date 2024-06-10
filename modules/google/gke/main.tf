data "google_compute_zones" "available" {
  region = "europe-north1"
  status = "UP"
}

resource "google_service_account" "service_account" {
  account_id   = "${local.hname}-sa"
  display_name = "${local.hname}-sa"
}


module "gke" {
  source = "terraform-google-modules/kubernetes-engine/google//modules/beta-private-cluster"
  version = "31.0.0"

  name       = "${local.hname}-gke"
  kubernetes_version     = "1.27.11-gke.1062004"
  release_channel        = "UNSPECIFIED" # in order to disable auto upgrade
  region                 = "europe-north1"
  network                = google_compute_network.vpc.name
  subnetwork             = google_compute_subnetwork.subnet.name
  master_ipv4_cidr_block = "10.1.0.0/28"
  ip_range_pods          = google_compute_subnetwork.subnet.secondary_ip_range[0].range_name
  ip_range_services      = google_compute_subnetwork.subnet.secondary_ip_range[1].range_name

  service_account                 = google_service_account.service_account.email
  master_authorized_networks      = var.master_authorized_networks
  master_global_access_enabled    = false
  istio                           = false
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
            name               = "node-pool"
            machine_type       = "e2-small"
            node_locations     = join(",", slice(data.google_compute_zones.available.names, 0, 3))
            initial_node_count = 1
            min_count          = 1
            max_count          = 2
            max_pods_per_node  = 64
            disk_size_gb       = 30
            disk_type          = "pd-standard"
            image_type         = "COS_CONTAINERD"
            auto_repair        = true
            auto_upgrade       = false
            preemptible        = false
            max_surge          = 1
            max_unavailable    = 0
        },
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


}