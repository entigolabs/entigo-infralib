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
  default  = "1.30"
}

variable "subnets" {
  type = list(string)
}

variable "cluster_primary_security_group_id" {
  type     = string
  nullable = false
  default  = ""
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

variable "labels" {
  type        = map(string)
  default     = null
}

variable "taints" {
  type        = any
  default     = {}
}
