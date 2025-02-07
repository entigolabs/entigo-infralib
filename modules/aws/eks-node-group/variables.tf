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

variable "node_security_group_id" {
  type     = string
  nullable = false
  default  = ""
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

variable "encryption_kms_key_arn" {
  type = string
  default = ""
}

variable "remote_access" {
  type        = any
  default     = {}
}

variable "labels" {
  type        = map(string)
  default     = null
}

variable "taints" {
  type        = any
  default     = {}
}
