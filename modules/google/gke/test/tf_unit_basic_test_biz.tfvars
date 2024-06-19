master_authorized_networks = [
    {
      display_name = "Allow all"
      cidr_block   = "0.0.0.0/0"
    }
    ]

eks_spot_min_size            = 0
eks_spot_max_size            = 0
eks_db_min_size              = 0
eks_db_max_size              = 0
eks_main_min_size            = 3
eks_main_max_size            = 6
eks_main_instance_type       = "e2-small"
eks_mainarm_min_size         = 1
eks_mainarm_max_size         = 2
eks_mainarm_instance_types   = "e2-small"
eks_managed_node_groups_extra = [
        {
            name               = "custom"
            machine_type       = "e2-small"
            initial_node_count = 1
            min_count          = 1
            max_count          = 1
            max_pods_per_node  = 64
            disk_size_gb       = 10
            disk_type          = "pd-standard"
            image_type         = "COS_CONTAINERD"
            auto_repair        = true
            auto_upgrade       = false
            spot               = false
        }
]
