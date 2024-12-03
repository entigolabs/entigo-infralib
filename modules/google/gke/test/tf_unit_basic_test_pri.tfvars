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
gke_mon_node_locations = "europe-north1-c"
