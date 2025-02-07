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
  default  = "1.30"
}

variable "cluster_iam_role_arn" {
  type     = string
  nullable = true
  default  = null
}

variable "authentication_mode" {
  type     = string
  nullable = false
  default  = "API_AND_CONFIG_MAP" # Possible values are API, CONFIG_MAP or API_AND_CONFIG_MAP
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
  default  = 2
}

variable "eks_main_max_size" {
  type     = number
  nullable = false
  default  = 4
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

variable "eks_mon_min_size" {
  type     = number
  nullable = false
  default  = 1
}

variable "eks_mon_desired_size" {
  type     = number
  nullable = false
  default  = 1
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

variable "eks_mon_single_subnet" {
  type     = bool
  nullable = false
  default  = true
}

variable "eks_tools_min_size" {
  type     = number
  nullable = false
  default  = 1
}

variable "eks_tools_desired_size" {
  type     = number
  nullable = false
  default  = 2
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

variable "eks_tools_single_subnet" {
  type     = bool
  nullable = false
  default  = false
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

variable "node_encryption_kms_key_arn" {
  type = string
  default = ""
}

variable "node_ssh_key_pair_name" {
  type = string
  nullable = true
  default = null
}

variable "bootstrap_self_managed_addons" {
  type    = bool
  nullable = true
  default = true
}

variable "coredns_addon_version" {
  type = string
  default = "v1.11.3-eksbuild.2"
}

variable "kube_proxy_addon_version" {
  type = string
  default = "v1.30.6-eksbuild.3"
}

variable "vpc_cni_addon_version" {
  type = string
  default = "v1.19.0-eksbuild.1"
}

variable "ebs_csi_addon_version" {
  type = string
  default = "v1.37.0-eksbuild.1"
}
