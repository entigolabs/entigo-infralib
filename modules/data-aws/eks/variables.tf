variable "prefix" {
  type = string
}

variable "cluster_name" {
  type = string
}

variable "enable_efs_csi" {
  type    = bool
  default = false
}

variable "node_security_group_id" {
  type    = string
  default = null
}