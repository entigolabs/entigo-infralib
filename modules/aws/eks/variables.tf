variable "prefix" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "private_subnets" {
  type = list(string)
}

variable "public_subnets" {
  type    = list(string)
  default = []
}

variable "eks_api_access_cidrs" {
  type    = list(string)
  default = []
}

variable "eks_nodeport_access_cidrs" {
  type    = list(string)
  default = ["127.0.0.1/32"] #Empty list will cause EKS to fail on first run. So this is a dummy value.
}

variable "eks_cluster_version" {
  type     = string
  nullable = false
  default  = "1.35"
}

variable "use_latest_ami_release_version" {
  description = "By default the node AMI versions are infralib release specific"
  type        = bool
  default     = false
  nullable    = false
}

variable "cluster_iam_role_arn" {
  type     = string
  nullable = true
  default  = null
}

variable "cluster_identity_providers" {
  description = "Map of cluster identity provider configurations to enable for the cluster. Note - this is different/separate from IRSA"
  type        = any
  default     = {}
}

variable "authentication_mode" {
  type     = string
  nullable = false
  default  = "API" # Possible values are API, CONFIG_MAP or API_AND_CONFIG_MAP
}

variable "enable_cluster_creator_admin_permissions" {
  type     = bool
  nullable = false
  default  = true
}

variable "iam_admin_role" {
  type     = string
  nullable = false
  default  = "" #Usually "AWSReservedSSO_AWSAdministratorAccess_.*" or "AWSReservedSSO_AdministratorAccess_.*"
}

variable "aws_auth_user" {
  type     = string
  nullable = false
  default  = ""
}

variable "iam_role_additional_policies" {
  type = list(string)
  default = []
}

variable "additional_access_entries" {
  type = map(object({
    principal_arn       = string
    user_name = optional(string)
    kubernetes_groups   = optional(list(string))
    policy_associations = map(object({
      policy_arn = string
      access_scope = object({
        type       = string
        namespaces = optional(list(string))
      })
    }))
  }))
  default = {}
}

variable "eks_cluster_public" {
  type     = bool
  nullable = false
  default  = false
}

variable "eks_main_min_size" {
  type     = number
  nullable = false
  default  = 1
}

variable "eks_main_desired_size" {
  type     = number
  nullable = false
  default  = 0
}

variable "eks_main_max_size" {
  type     = number
  nullable = false
  default  = 3
}

variable "eks_main_instance_types" {
  type    = list(string)
  default = ["t3.large"]
}

variable "eks_main_capacity_type" {
  type    = string
  default = "ON_DEMAND"
}

variable "eks_main_ami_type" {
  type    = string
  default = "AL2023_x86_64_STANDARD"
}

variable "eks_main_ami_release_version" {
  description = "The AMI version. Defaults to current AMI release version for the given infralib release"
  type        = string
  default     = null
}

variable "eks_main_volume_size" {
  type    = number
  default = 100
}

variable "eks_main_volume_iops" {
  type    = number
  default = 3000
}

variable "eks_main_volume_type" {
  type    = string
  default = "gp3"
}

variable "eks_main_subnets" {
  type    = list(string)
  default = []
}

variable "eks_mon_min_size" {
  type     = number
  nullable = false
  default  = 1
}

variable "eks_mon_desired_size" {
  type     = number
  nullable = false
  default  = 0
}

variable "eks_mon_max_size" {
  type     = number
  nullable = false
  default  = 3
}

variable "eks_mon_instance_types" {
  type    = list(string)
  default = ["t3.large"]
}

variable "eks_mon_capacity_type" {
  type    = string
  default = "ON_DEMAND"
}

variable "eks_mon_ami_type" {
  type    = string
  default = "AL2023_x86_64_STANDARD"
}

variable "eks_mon_ami_release_version" {
  description = "The AMI version. Defaults to current AMI release version for the given infralib release"
  type        = string
  default     = null
}

variable "eks_mon_volume_size" {
  type    = number
  default = 50
}

variable "eks_mon_volume_iops" {
  type    = number
  default = 3000
}

variable "eks_mon_volume_type" {
  type    = string
  default = "gp3"
}

variable "eks_mon_subnets" {
  type    = list(string)
  default = []
}

variable "eks_tools_min_size" {
  type     = number
  nullable = false
  default  = 1
}

variable "eks_tools_desired_size" {
  type     = number
  nullable = false
  default  = 0
}

variable "eks_tools_max_size" {
  type     = number
  nullable = false
  default  = 3
}

variable "eks_tools_instance_types" {
  type    = list(string)
  default = ["t3.large"]
}

variable "eks_tools_capacity_type" {
  type    = string
  default = "ON_DEMAND"
}

variable "eks_tools_ami_type" {
  type    = string
  default = "AL2023_x86_64_STANDARD"
}

variable "eks_tools_ami_release_version" {
  description = "The AMI version. Defaults to current AMI release version for the given infralib release"
  type        = string
  default     = null
}

variable "eks_tools_volume_size" {
  type    = number
  default = 50
}

variable "eks_tools_volume_iops" {
  type    = number
  default = 3000
}

variable "eks_tools_volume_type" {
  type    = string
  default = "gp3"
}

variable "eks_tools_subnets" {
  type    = list(string)
  default = []
}

variable "cluster_enabled_log_types" {
  type     = list(string)
  nullable = false
  default  = ["api", "authenticator"]
}


variable "eks_managed_node_groups_extra" {
  type     = map(any)
  nullable = false
  default  = {}
}

variable "cloudwatch_log_group_kms_key_id" {
  type = string
  default = ""
}

variable "cloudwatch_log_group_retention_in_days" {
  description = "Number of days to retain EKS control plane log events in the CloudWatch log group"
  type        = number
  nullable    = false
  default     = 90
}

variable "cloudwatch_log_group_class" {
  description = "Log class of the CloudWatch log group. Possible values are STANDARD or INFREQUENT_ACCESS"
  type        = string
  nullable    = true
  default     = null
}

variable "cloudwatch_log_group_tags" {
  description = "Additional tags to add to the CloudWatch log group"
  type        = map(string)
  nullable    = false
  default     = {}
}

variable "node_encryption_kms_key_arn" {
  type = string
  default = ""
}

variable "node_ssh_key_pair_name" {
  type = string
  nullable = true
  default = null
}

variable "node_launch_template_tags" {
  type        = map(string)
  default     = {}
  nullable = false
}

variable "node_repair_config" {
  description = "Disabled when null, enable with true"
  type = object({
    enabled                                 = bool
    max_parallel_nodes_repaired_count       = optional(number)
    max_parallel_nodes_repaired_percentage  = optional(number)
    max_unhealthy_node_threshold_count      = optional(number)
    max_unhealthy_node_threshold_percentage = optional(number)
    node_repair_config_overrides = optional(list(object({
      min_repair_wait_time_mins = number
      node_monitoring_condition = string
      node_unhealthy_reason     = string
      repair_action             = string
    })))
  })
  default = null
}

variable "node_repair_wait_time_mins" {
  description = "Wait time in minutes before repairing a node, applied to all repair triggering node conditions when node_repair_config does not set node_repair_config_overrides. 10-120 in steps of 10"
  type        = number
  nullable    = false
  default     = 20
  validation {
    condition     = var.node_repair_wait_time_mins >= 10 && var.node_repair_wait_time_mins <= 120 && var.node_repair_wait_time_mins % 10 == 0
    error_message = "node_repair_wait_time_mins must be between 10 and 120 and a multiple of 10."
  }
}

variable "coredns_addon_version" {
  type = string
  default = "v1.13.2-eksbuild.4"
}

variable "kube_proxy_addon_version" {
  type = string
  default = "v1.35.3-eksbuild.5"
}

variable "vpc_cni_addon_version" {
  type = string
  default = "v1.21.2-eksbuild.2"
}

variable "ebs_csi_addon_version" {
  type = string
  default = "v1.59.0-eksbuild.1"
}

variable "efs_csi_addon_version" {
  type = string
  default = "v3.2.0-eksbuild.1"
}

variable "node_monitoring_agent_addon_version" {
  type = string
  default = "v1.7.2-eksbuild.1"
}

variable "enable_vpc_cni_prefix_delegation" {
  type    = string
  default = "true"
}

variable "enable_vpc_cni_network_policy" {
  type    = string
  default = "true"
}

variable "enable_efs_csi" {
  type    = bool
  default = false
}

variable "enable_node_monitoring_agent" {
  type        = bool
  nullable    = false
  default     = false
}

variable "coredns_config" {
  type    = any
  default = {}
}

variable "kube_proxy_config" {
  type    = any
  default = {}
}

variable "vpc_cni_env_config" {
  type    = any
  default = {}
}

variable "ebs_csi_controller_config" {
  type    = any
  default = {}
}

variable "efs_csi_controller_config" {
  type    = any
  default = {}
}

variable "node_monitoring_agent_config" {
  description = "eks-node-monitoring-agent EKS addon configuration_values"
  type     = any
  nullable = false
  default  = {}
}


variable "tags" {
  type        = map(string)
  description = "Additional tags for the EKS cluster resource only (does not affect node groups)"
  default     = {}
}
