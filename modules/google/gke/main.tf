resource "google_service_account" "service_account" {
  account_id   = local.hname
  display_name = local.hname
}

locals {

  gke_managed_node_groups_all = [
        {
            name               = "main"
            machine_type       = var.gke_main_instance_type
            node_locations     = data.google_client_config.this.zone
            initial_node_count = var.gke_main_min_size
            min_count          = var.gke_main_min_size
            max_count          = var.gke_main_max_size
            max_pods_per_node  = var.gke_main_max_pods
            disk_size_gb       = var.gke_main_volume_size
            disk_type          = var.gke_main_volume_type
            image_type         = "COS_CONTAINERD"
            auto_repair        = true
            auto_upgrade       = false
            spot               = false
        },
        {
            name               = "mainarm"
            machine_type       = var.gke_mainarm_instance_type
            node_locations     = data.google_client_config.this.zone
            initial_node_count = var.gke_mainarm_min_size
            min_count          = var.gke_mainarm_min_size
            max_count          = var.gke_mainarm_max_size
            max_pods_per_node  = var.gke_mainarm_max_pods
            disk_size_gb       = var.gke_mainarm_volume_size
            disk_type          = var.gke_mainarm_volume_type
            image_type         = "COS_CONTAINERD"
            auto_repair        = true
            auto_upgrade       = false
            spot               = false
        },
        {
            name               = "spot"
            machine_type       = var.gke_spot_instance_type
            node_locations     = data.google_client_config.this.zone
            initial_node_count = var.gke_spot_min_size
            min_count          = var.gke_spot_min_size
            max_count          = var.gke_spot_max_size
            max_pods_per_node  = var.gke_spot_max_pods
            disk_size_gb       = var.gke_spot_volume_size
            disk_type          = var.gke_spot_volume_type
            image_type         = "COS_CONTAINERD"
            auto_repair        = true
            auto_upgrade       = false
            spot               = true
        },
        {
            name               = "mon"
            machine_type       = var.gke_mon_instance_type
            node_locations     = data.google_client_config.this.zone
            initial_node_count = var.gke_mon_min_size
            min_count          = var.gke_mon_min_size
            max_count          = var.gke_mon_max_size
            max_pods_per_node  = var.gke_mon_max_pods
            disk_size_gb       = var.gke_mon_volume_size
            disk_type          = var.gke_mon_volume_type
            image_type         = "COS_CONTAINERD"
            auto_repair        = true
            auto_upgrade       = false
            spot               = false
        },
        {
            name               = "tools"
            machine_type       = var.gke_tools_instance_type
            node_locations     = data.google_client_config.this.zone
            initial_node_count = var.gke_tools_min_size
            min_count          = var.gke_tools_min_size
            max_count          = var.gke_tools_max_size
            max_pods_per_node  = var.gke_tools_max_pods
            disk_size_gb       = var.gke_tools_volume_size
            disk_type          = var.gke_tools_volume_type
            image_type         = "COS_CONTAINERD"
            auto_repair        = true
            auto_upgrade       = false
            spot               = false
        },
        {
            name               = "db"
            machine_type       = var.gke_db_instance_type
            node_locations     = data.google_client_config.this.zone
            initial_node_count = var.gke_db_min_size
            min_count          = var.gke_db_min_size
            max_count          = var.gke_db_max_size
            max_pods_per_node  = var.gke_db_max_pods
            disk_size_gb       = var.gke_db_volume_size
            disk_type          = var.gke_db_volume_type
            image_type         = "COS_CONTAINERD"
            auto_repair        = true
            auto_upgrade       = false
            spot               = false
        }
      ]

  gke_managed_node_groups = concat(local.gke_managed_node_groups_all, var.gke_managed_node_groups_extra)
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
  master_global_access_enabled    = var.master_global_access_enabled
  #istio                           = false //only in beta module
  issue_client_certificate        = false
  enable_private_endpoint         = var.enable_private_endpoint
  enable_private_nodes            = true
  remove_default_node_pool        = true
  enable_shielded_nodes           = false
  identity_namespace              = "enabled"
  node_metadata                   = "GKE_METADATA"
  horizontal_pod_autoscaling      = true
  enable_vertical_pod_autoscaling = false
  deletion_protection             = false

  node_pools                      = local.gke_managed_node_groups
  node_pools_labels = {
    all = {}
    main = { "main" = "true" }
    mainarm = { "mainarm" = "true" }
    spot = { "spot" = "true" }
    mon = { "mon" = "true" }
    tools = { "tools" = "true" }
    db = { "db" = "true" }
  }
  
  node_pools_taints = {
    all = []
    mon = [{
                key    = "mon"
                value  = "true"
                effect = "NO_SCHEDULE"
          }]
    tools = [{
                key    = "tools"
                value  = "true"
                effect = "NO_SCHEDULE"
          }]
    db = [{
                key    = "db"
                value  = "true"
                effect = "NO_SCHEDULE"
          }]
  }

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

  node_pools_tags = {
    all = ["k8s-nodes"]
  }

  node_pools_metadata = {
    all = {
      disable-legacy-endpoints = "true"
    }
  }

  master_authorized_networks = var.master_authorized_networks
}


module "cluster_id" {
  source                             = "./secret"
  prefix = var.prefix
  key = "cluster_id"
  value = module.gke.cluster_id
}

module "cluster_endpoint" {
  source                             = "./secret"
  prefix = var.prefix
  key = "cluster_endpoint"
  value = nonsensitive(module.gke.endpoint)
}

module "cluster_name" {
  source                             = "./secret"
  prefix = var.prefix
  key = "cluster_name"
  value = module.gke.name
}

module "region" {
  source                             = "./secret"
  prefix = var.prefix
  key = "region"
  value = module.gke.region
}
