resource "random_integer" "subnet_third_octet" {
  min = 16
  max = 31
}

resource "random_integer" "subnet_fourth_octet_raw" {
  min = 0
  max = 15  # We'll multiply this by 16 later to get alignment
}

locals {
  aligned_fourth_octet = random_integer.subnet_fourth_octet_raw.result * 16
  subnet_cidr = format("172.%d.%d.%d/28", 
    random_integer.subnet_third_octet.result,
    local.aligned_fourth_octet,
    0
  )


  google_compute_zones = join(",", data.google_compute_zones.this.names)

  gke_main_node_locations    = var.gke_main_node_locations != "" ? var.gke_main_node_locations : local.google_compute_zones
  gke_mainarm_node_locations = var.gke_mainarm_node_locations != "" ? var.gke_mainarm_node_locations : local.google_compute_zones
  gke_spot_node_locations    = var.gke_spot_node_locations != "" ? var.gke_spot_node_locations : local.google_compute_zones
  gke_mon_node_locations     = var.gke_mon_node_locations != "" ? var.gke_mon_node_locations : local.google_compute_zones
  gke_tools_node_locations   = var.gke_tools_node_locations != "" ? var.gke_tools_node_locations : local.google_compute_zones
  gke_db_node_locations      = var.gke_db_node_locations != "" ? var.gke_db_node_locations : local.google_compute_zones

  gke_managed_node_groups_all = [
    {
      name               = "main"
      machine_type       = var.gke_main_instance_type
      node_locations     = local.gke_main_node_locations
      location_policy    = var.gke_main_location_policy
      initial_node_count = var.gke_main_min_size
      min_count          = var.gke_main_min_size
      max_count          = var.gke_main_max_size
      max_pods_per_node  = var.gke_main_max_pods
      disk_size_gb       = var.gke_main_volume_size
      disk_type          = var.gke_main_volume_type
      image_type         = "COS_CONTAINERD"
      auto_repair        = true
      auto_upgrade       = false
      spot               = var.gke_main_spot_nodes
    },
    {
      name               = "mainarm"
      machine_type       = var.gke_mainarm_instance_type
      node_locations     = local.gke_mainarm_node_locations
      location_policy    = var.gke_mainarm_location_policy
      initial_node_count = var.gke_mainarm_min_size
      min_count          = var.gke_mainarm_min_size
      max_count          = var.gke_mainarm_max_size
      max_pods_per_node  = var.gke_mainarm_max_pods
      disk_size_gb       = var.gke_mainarm_volume_size
      disk_type          = var.gke_mainarm_volume_type
      image_type         = "COS_CONTAINERD"
      auto_repair        = true
      auto_upgrade       = false
      spot               = var.gke_mainarm_spot_nodes
    },
    {
      name               = "spot"
      machine_type       = var.gke_spot_instance_type
      node_locations     = local.gke_spot_node_locations
      location_policy    = var.gke_spot_location_policy
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
      node_locations     = local.gke_mon_node_locations
      location_policy    = var.gke_mon_location_policy
      initial_node_count = var.gke_mon_min_size
      min_count          = var.gke_mon_min_size
      max_count          = var.gke_mon_max_size
      max_pods_per_node  = var.gke_mon_max_pods
      disk_size_gb       = var.gke_mon_volume_size
      disk_type          = var.gke_mon_volume_type
      image_type         = "COS_CONTAINERD"
      auto_repair        = true
      auto_upgrade       = false
      spot               = var.gke_mon_spot_nodes
    },
    {
      name               = "tools"
      machine_type       = var.gke_tools_instance_type
      node_locations     = local.gke_tools_node_locations
      location_policy    = var.gke_tools_location_policy
      initial_node_count = var.gke_tools_min_size
      min_count          = var.gke_tools_min_size
      max_count          = var.gke_tools_max_size
      max_pods_per_node  = var.gke_tools_max_pods
      disk_size_gb       = var.gke_tools_volume_size
      disk_type          = var.gke_tools_volume_type
      image_type         = "COS_CONTAINERD"
      auto_repair        = true
      auto_upgrade       = false
      spot               = var.gke_tools_spot_nodes
    },
    {
      name               = "db"
      machine_type       = var.gke_db_instance_type
      node_locations     = local.gke_db_node_locations
      location_policy    = var.gke_db_location_policy
      initial_node_count = var.gke_db_min_size
      min_count          = var.gke_db_min_size
      max_count          = var.gke_db_max_size
      max_pods_per_node  = var.gke_db_max_pods
      disk_size_gb       = var.gke_db_volume_size
      disk_type          = var.gke_db_volume_type
      image_type         = "COS_CONTAINERD"
      auto_repair        = true
      auto_upgrade       = false
      spot               = var.gke_db_spot_nodes
    }
  ]

  gke_managed_node_groups = concat(local.gke_managed_node_groups_all, var.gke_managed_node_groups_extra)
}

module "gke" {
  source  = "terraform-google-modules/kubernetes-engine/google//modules/private-cluster"
  version = "34.0.0"

  project_id             = data.google_client_config.this.project
  name                   = var.prefix
  kubernetes_version     = data.google_container_engine_versions.this.release_channel_latest_version["STABLE"]
  release_channel        = "UNSPECIFIED" # in order to disable auto upgrade
  region                 = data.google_client_config.this.region
  network                = var.network
  subnetwork             = var.subnetwork
  master_ipv4_cidr_block = var.master_ipv4_cidr_block == "" ? local.subnet_cidr : var.master_ipv4_cidr_block
  ip_range_pods          = var.ip_range_pods
  ip_range_services      = var.ip_range_services

  # istio                              = false //only in beta module
  service_account_name                   = var.prefix
  grant_registry_access                  = var.grant_registry_access
  registry_project_ids                   = var.registry_project_ids
  master_global_access_enabled           = var.master_global_access_enabled
  enable_l4_ilb_subsetting               = var.enable_l4_ilb_subsetting
  issue_client_certificate               = false
  deploy_using_private_endpoint          = var.deploy_using_private_endpoint
  enable_private_endpoint                = var.enable_private_endpoint
  enable_private_nodes                   = true
  remove_default_node_pool               = true
  enable_shielded_nodes                  = false
  identity_namespace                     = "enabled"
  node_metadata                          = "GKE_METADATA"
  horizontal_pod_autoscaling             = true
  enable_vertical_pod_autoscaling        = false
  deletion_protection                    = false
  gateway_api_channel                    = "CHANNEL_STANDARD"
  monitoring_enable_managed_prometheus   = var.monitoring_enable_managed_prometheus
  monitoring_enabled_components          = var.monitoring_enabled_components
  logging_enabled_components             = var.logging_enabled_components
  insecure_kubelet_readonly_port_enabled = false

  node_pools = local.gke_managed_node_groups
  node_pools_labels = {
    all     = {}
    main    = { "main" = "true" }
    mainarm = { "mainarm" = "true" }
    spot    = { "spot" = "true" }
    mon     = { "mon" = "true" }
    tools   = { "tools" = "true" }
    db      = { "db" = "true" }
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
