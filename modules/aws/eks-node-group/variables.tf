variable "prefix" {
  type = string
}

variable "cluster_name" {
  type     = string
  nullable = false
  default  = ""
}

variable "cluster_version" {
  type     = string
  nullable = false
  default  = "1.35"
}

variable "subnets" {
  type = list(string)
}

variable "cluster_service_cidr" {
  type        = string
  default     = ""
}

variable "security_group_ids" {
  type = list(string)
  default = []
}

variable "iam_role_additional_policies" {
  type = list(string)
  default = []
}

variable "min_size" {
  type     = number
  nullable = false
  default  = 1
}

variable "desired_size" {
  type     = number
  nullable = false
  default  = 2
}

variable "max_size" {
  type     = number
  nullable = false
  default  = 4
}

variable "instance_types" {
  type    = list(string)
  default = ["t3.large"]
}

variable "capacity_type" {
  type    = string
  default = "ON_DEMAND"
}

variable "ami_type" {
  description = "Type of Amazon Machine Image (AMI) associated with the EKS Node Group. AL2023_ARM_64_STANDARD for arm AL2023_x86_64_STANDARD for x86"
  type        = string
  default     = "AL2023_x86_64_STANDARD"
}

variable "ami_release_version" {
  description = "The AMI version. Defaults to current AMI release version for the given infralib release"
  type        = string
  default     = null
}

variable "use_latest_ami_release_version" {
  type        = bool
  default     = false
  nullable    = false
}

variable "volume_size" {
  type    = number
  default = 100
}

variable "volume_iops" {
  type    = number
  default = 3000
}

variable "volume_type" {
  type    = string
  default = "gp3"
}

variable "block_device_mappings" {
  description = "If set then volume_size, volume_iops and volume_type are ignored."
  type        = any
  nullable    = true
  default     = null
}

variable "encryption_kms_key_arn" {
  type = string
  default = ""
}

variable "key_name" {
  type     = string
  nullable = false
  default  = ""
}

variable "pre_bootstrap_user_data" {
  type        = string
  default     = ""
}

variable "cloudinit_pre_nodeadm_content" {
  description = "Value of cloud-init document content"
  type = string
  default = ""
}

variable "labels" {
  type        = map(string)
  default     = null
}

variable "taints" {
  type        = any
  default     = {}
}

variable "node_repair_config" {
  description = "Disabled when null, enable with { enabled = true }"
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

variable "tags" {
  type    = map(string)
  default = {}
}
