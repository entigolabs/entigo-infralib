variable "prefix" {
  type = string
}

variable "efs_volumes" {
  description = "The name of the file system"
  type        = list(string)
  default     = ["shared"]
}


variable "efs_csi_service_account_role_arn" {
  description = "AWS EKS EFS CSI Service Account Role ARN"
  type        = string
  default     = ""
}

variable "tags" {
  description = "A map of tags to add to all resources"
  type        = map(string)
  default     = {}
}

variable "performance_mode" {
  description = "The file system performance mode. Can be either `generalPurpose` or `maxIO`. Default is `generalPurpose`"
  type        = string
  default     = null
}

variable "encrypted" {
  description = "If `true`, the disk will be encrypted"
  type        = bool
  default     = true
}

variable "kms_key_arn" {
  description = "The ARN for the KMS encryption key. When specifying `kms_key_arn`, encrypted needs to be set to `true`"
  type        = string
  default     = null
}

variable "provisioned_throughput_in_mibps" {
  description = "The throughput, measured in MiB/s, that you want to provision for the file system. Only applicable with `throughput_mode` set to `provisioned`"
  type        = number
  default     = null
}

variable "throughput_mode" {
  description = "Throughput mode for the file system. Defaults to `bursting`. Valid values: `bursting`, `elastic`, and `provisioned`. When using `provisioned`, also set `provisioned_throughput_in_mibps`"
  type        = string
  default     = null
}

variable "lifecycle_policy" {
  description = "A file system [lifecycle policy](https://docs.aws.amazon.com/efs/latest/ug/API_LifecyclePolicy.html) object"
  type = object({
    transition_to_ia                    = optional(string)
    transition_to_archive               = optional(string)
    transition_to_primary_storage_class = optional(string)
  })
  default  = {
    transition_to_ia = "AFTER_90_DAYS"
    transition_to_primary_storage_class = "AFTER_1_ACCESS"
  }
  nullable = false
}

variable "protection" {
  description = "A map of file protection configurations"
  type = object({
    replication_overwrite = optional(string)
  })
  default = {
    replication_overwrite = "ENABLED"
  }
}

################################################################################
# File System Policy
################################################################################

variable "attach_policy" {
  description = "Determines whether a policy is attached to the file system"
  type        = bool
  default     = true
}

variable "bypass_policy_lockout_safety_check" {
  description = "A flag to indicate whether to bypass the `aws_efs_file_system_policy` lockout safety check. Defaults to `false`"
  type        = bool
  default     = false
}

variable "policy_statements" {
  description = "A list of IAM policy [statements](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document#statement) for custom permission usage"
  type = map(object({
    sid           = optional(string)
    actions       = optional(list(string))
    not_actions   = optional(list(string))
    effect        = optional(string)
    resources     = optional(list(string))
    not_resources = optional(list(string))
    principals = optional(list(object({
      type        = string
      identifiers = list(string)
    })))
    not_principals = optional(list(object({
      type        = string
      identifiers = list(string)
    })))
    conditions = optional(list(object({
      test     = string
      values   = list(string)
      variable = string
    })))
    condition = optional(list(object({
      test     = string
      values   = list(string)
      variable = string
    })))
  }))
  default = null
}

variable "deny_nonsecure_transport" {
  description = "Determines whether `aws:SecureTransport` is required when connecting to elastic file system"
  type        = bool
  default     = true
}

variable "deny_nonsecure_transport_via_mount_target" {
  description = "Determines whether to use the common policy option for denying nonsecure transport which allows all AWS principals when accessed via EFS mounted target"
  type        = bool
  default     = true
}

################################################################################
# Mount Target(s)
################################################################################

variable "mount_targets_subnet_id" {
  description = "A list of subnet ids to mount targets to."
  type = list(string)
  default  = []
  nullable = false
}

################################################################################
# Security Group
################################################################################

variable "create_security_group" {
  description = "Determines whether a security group is created"
  type        = bool
  default     = true
}


variable "security_group_use_name_prefix" {
  description = "Determines whether to use a name prefix for the security group. If `true`, the `security_group_name` value will be used as a prefix"
  type        = bool
  default     = false
}

variable "security_group_vpc_id" {
  description = "The VPC ID where the security group will be created"
  type        = string
  default     = null
}

variable "security_group_cidrs" {
  description = "List of security group cidrs that can use the shares."
  type = list(string)
  default  = []
  nullable = false
}

variable "security_group_ingress_rules" {
  description = "Map of security group ingress rules to add to the security group created"
  type = map(object({
    name = optional(string)

    cidr_ipv4                    = optional(string)
    cidr_ipv6                    = optional(string)
    description                  = optional(string)
    from_port                    = optional(number, 2049)
    ip_protocol                  = optional(string, "tcp")
    prefix_list_id               = optional(string)
    referenced_security_group_id = optional(string)
    region                       = optional(string)
    tags                         = optional(map(string), {})
    to_port                      = optional(number, 2049)
  }))
  default  = {}
  nullable = false
}

variable "security_group_egress_rules" {
  description = "Map of security group egress rules to add to the security group created"
  type = map(object({
    name = optional(string)

    cidr_ipv4                    = optional(string)
    cidr_ipv6                    = optional(string)
    description                  = optional(string)
    from_port                    = optional(number, 2049)
    ip_protocol                  = optional(string, "tcp")
    prefix_list_id               = optional(string)
    referenced_security_group_id = optional(string)
    region                       = optional(string)
    tags                         = optional(map(string), {})
    to_port                      = optional(number, 2049)
  }))
  default  = {}
  nullable = false
}

################################################################################
# Backup Policy
################################################################################

variable "create_backup_policy" {
  description = "Determines whether a backup policy is created"
  type        = bool
  default     = true
}

variable "enable_backup_policy" {
  description = "Determines whether a backup policy is `ENABLED` or `DISABLED`"
  type        = bool
  default     = true
}

################################################################################
# Replication Configuration
################################################################################

variable "create_replication_configuration" {
  description = "Determines whether a replication configuration is created"
  type        = bool
  default     = false
}

variable "replication_configuration_destination" {
  description = "A destination configuration block"
  type = object({
    availability_zone_name = optional(string)
    file_system_id         = optional(string)
    kms_key_id             = optional(string)
    region                 = optional(string)
  })
  default = null
}
