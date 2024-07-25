master_authorized_networks = [
    {
      display_name = "Allow all"
      cidr_block   = "0.0.0.0/0"
    }
    ]

gke_tools_min_size            = 2
gke_tools_max_size            = 4
gke_mon_min_size            = 1
gke_mon_max_size            = 4
gke_main_min_size            = 2
gke_main_max_size            = 6
