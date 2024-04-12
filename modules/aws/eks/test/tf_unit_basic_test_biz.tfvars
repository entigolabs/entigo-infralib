eks_cluster_public           = true
eks_mon_single_subnet = false
cluster_enabled_log_types    = [] #Temporarily disabled, see https://entigo.atlassian.net/browse/RD-8
eks_spot_min_size            = 0
eks_spot_max_size            = 0
eks_db_min_size              = 0
eks_db_max_size              = 0
eks_main_min_size            = 3
eks_main_max_size            = 6
eks_main_instance_types   = ["t3.micro"]
eks_mainarm_min_size         = 1
eks_mainarm_max_size         = 2
eks_mainarm_instance_types   = ["t4g.micro"]
eks_managed_node_groups_extra = {
  altarm = {
        min_size        = 1
        desired_size    = 1
        max_size        = 1
        instance_types  = ["t4g.micro"]
        capacity_type   = "ON_DEMAND"
        release_version = ""
        ami_type        = "AL2_ARM_64"
        launch_template_tags = {
          Terraform = "true"
        }
        labels = {
          altarm = "true"
        }
        block_device_mappings = {
          xvda = {
            device_name = "/dev/xvda"
            ebs = {
              volume_size           = 30
              volume_iops           = 1000
              volume_type           = "gp3"
              delete_on_termination = true
            }
          }
        }
      }
}
