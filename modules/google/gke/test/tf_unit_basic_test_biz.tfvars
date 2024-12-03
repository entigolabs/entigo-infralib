master_authorized_networks = [
  {
    display_name = "Allow all"
    cidr_block   = "0.0.0.0/0"
  }
]

gke_main_min_size        = 1
gke_main_max_size        = 3
gke_main_spot_nodes      = true
gke_main_location_policy = "ANY"

gke_tools_min_size        = 1
gke_tools_max_size        = 2
gke_tools_spot_nodes      = true
gke_tools_location_policy = "ANY"

gke_mon_min_size       = 1
gke_mon_max_size       = 3
gke_mon_spot_nodes     = true
gke_mon_node_locations = "europe-north1-b"

gke_managed_node_groups_extra = [
  {
    name               = "custom"
    machine_type       = "e2-standard-2"
    node_locations     = "europe-north1-a"
    initial_node_count = 0
    min_count          = 0
    max_count          = 0
    max_pods_per_node  = 64
    disk_size_gb       = 10
    disk_type          = "pd-standard"
    image_type         = "COS_CONTAINERD"
    auto_repair        = true
    auto_upgrade       = false
    spot               = false
  }
]
