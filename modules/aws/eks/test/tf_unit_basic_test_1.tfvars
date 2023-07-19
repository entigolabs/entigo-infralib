vpc_prefix                   = "runner-main"
eks_cluster_public           = true
eks_monitoring_single_subnet = false
cluster_enabled_log_types    = [] #Temporarily disabled, see https://entigo.atlassian.net/browse/RD-8
eks_spot_min_size            = 0
eks_spot_max_size            = 0
eks_db_min_size              = 0
eks_db_max_size              = 0
