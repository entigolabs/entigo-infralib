variable "prefix" {
  type = string
}

variable "vpc_prefix" {
  type = string
}

variable "eks_cluster_version" {
  type = string
  default = "1.26"
}

variable "iam_admin_role" {
  type = string
  default = "AWSReservedSSO_AdministratorAccess_.*" #Sometimes "AWSReservedSSO_AWSAdministratorAccess_.*" ?
}

variable "eks_cluster_public" {
  type = bool
  default = false
}

variable "eks_main_min_size" {
  type = number
  default = 1
}

variable "eks_main_max_size" {
  type = number
  default = 3
}

variable "eks_main_instance_types" {
  type = list(string)
  default = ["t3.large"]
}

variable "eks_spot_min_size" {
  type = number
  default = 1
}

variable "eks_spot_max_size" {
  type = number
  default = 3
}

variable "eks_spot_instance_types" {
  type = list(string)
  default = [
        "t3.medium",
        "t3.large"
  ]
}

variable "eks_db_min_size" {
  type = number
  default = 1
}

variable "eks_db_max_size" {
  type = number
  default = 3
}

variable "eks_db_instance_types" {
  type = list(string)
  default = ["t3.large"]
}


variable "eks_monitoring_min_size" {
  type = number
  default = 1
}

variable "eks_monitoring_max_size" {
  type = number
  default = 3
}

variable "eks_monitoring_instance_types" {
  type = list(string)
  default = ["t3.large"]
}

variable "eks_monitoring_single_subnet" {
  type = bool
  default = true
}

variable "cluster_enabled_log_types" {
  type = list(string)
  default = ["api", "authenticator"]
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
