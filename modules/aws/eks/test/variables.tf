variable "prefix" {
  type = string
}

variable "vpc_prefix" {
  type = string
}

variable "eks_cluster_version" {
  type    = string
  default = null
}

variable "iam_admin_role" {
  type    = string
  default = null
}

variable "eks_cluster_public" {
  type    = bool
  default = null
}

variable "eks_main_min_size" {
  type    = number
  default = null
}

variable "eks_main_max_size" {
  type    = number
  default = null
}

variable "eks_main_instance_types" {
  type    = list(string)
  default = null
}

variable "eks_spot_min_size" {
  type    = number
  default = null
}

variable "eks_spot_max_size" {
  type    = number
  default = null
}

variable "eks_spot_instance_types" {
  type = list(string)
  default = null
}

variable "eks_db_min_size" {
  type    = number
  default = null
}

variable "eks_db_max_size" {
  type    = number
  default = null
}

variable "eks_db_instance_types" {
  type    = list(string)
  default = null
}


variable "eks_monitoring_min_size" {
  type    = number
  default = null
}

variable "eks_monitoring_max_size" {
  type    = number
  default = null
}

variable "eks_monitoring_instance_types" {
  type    = list(string)
  default = null
}

variable "eks_monitoring_single_subnet" {
  type    = bool
  default = null
}

variable "cluster_enabled_log_types" {
  type    = list(string)
  default = null
}

