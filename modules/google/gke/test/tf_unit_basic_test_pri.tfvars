master_authorized_networks = [
    {
      display_name = "Allow all"
      cidr_block   = "0.0.0.0/0"
    }
    ]



gke_main_min_size            = 1
gke_main_max_size            = 2
gke_main_instance_type      = "e2-standard-4"
gke_main_node_locations     = "europe-north1-a"

gke_tools_min_size            = 1
gke_tools_max_size            = 3
gke_tools_instance_type      = "e2-standard-4"
gke_tools_node_locations     = "europe-north1-a"

gke_mon_min_size            = 1
gke_mon_max_size            = 2
gke_mon_node_locations       = "europe-north1-a"

gke_mainarm_node_locations  = "europe-north1-a"

gke_db_node_locations       = "europe-north1-a"

gke_spot_node_locations     = "europe-north1-a"
