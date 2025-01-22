eks_cluster_public           = true
cluster_enabled_log_types    = [] #Temporarily disabled, see https://entigo.atlassian.net/browse/RD-8
eks_mon_single_subnet = true
eks_tools_desired_size        = 0
eks_tools_max_size            = 0
eks_tools_capacity_type    = "SPOT"
eks_mon_desired_size        = 0
eks_mon_max_size            = 0
eks_mon_capacity_type    = "SPOT"
eks_main_min_size        = 4
eks_main_desired_size    = 0
eks_main_max_size            = 8
eks_main_capacity_type    = "SPOT"
eks_mainarm_capacity_type    = "SPOT"
eks_db_capacity_type    = "SPOT"
eks_nodeport_access_cidrs   = ["10.10.10.10/32"]
iam_admin_role = "AWSReservedSSO_AdministratorAccess_.*"
