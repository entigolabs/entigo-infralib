master_authorized_networks = [
    {
      display_name = "Allow all"
      cidr_block   = "0.0.0.0/0"
    }
    ]

gke_spot_min_size            = 0
gke_spot_max_size            = 0
gke_db_min_size              = 0
gke_db_max_size              = 0
gke_tools_min_size            = 3
gke_tools_max_size            = 8
gke_mon_min_size            = 1
gke_mon_max_size            = 4
gke_main_min_size            = 3
gke_main_max_size            = 8
gke_main_instance_type       = "e2-medium"
gke_mainarm_min_size         = 0
gke_mainarm_max_size         = 0
gke_mainarm_instance_types   = "e2-medium"
gke_managed_node_groups_extra = [
        {
            name               = "custom"
            machine_type       = "e2-medium"
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
