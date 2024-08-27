variable "prefix" {
  type = string
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
  eks_min_size_map = merge(
    {
      eks_main_min_size     = var.eks_main_min_size
      eks_mainarm_min_size  = var.eks_mainarm_min_size
      eks_tools_min_size    = var.eks_tools_min_size
      eks_mon_min_size      = var.eks_mon_min_size
      eks_spot_min_size     = var.eks_spot_min_size
      eks_db_min_size       = var.eks_db_min_size
    },
    var.eks_extra_min_sizes
  )
}

variable "cluster_name" {
  type = string
}

variable "eks_main_min_size" {
  type    = number
  nullable = false
  default = 2
}

variable "eks_mainarm_min_size" {
  type    = number
  nullable = false
  default = 0
}

variable "eks_tools_min_size" {
  type    = number
  nullable = false
  default = 2
}

variable "eks_mon_min_size" {
  type    = number
  nullable = false
  default = 1
}

variable "eks_spot_min_size" {
  type    = number
  nullable = false
  default = 0
}

variable "eks_db_min_size" {  
  type    = number
  nullable = false
  default = 0
}

variable "eks_extra_min_sizes" {
  type    = map(number)
  nullable = false
  default = {}
}
