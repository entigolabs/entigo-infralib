variable "prefix" {
  type = string
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
  
  eks_min_size_map = {
    main = var.eks_main_min_size
    mainarm = var.eks_mainarm_min_size
    tools = var.eks_tools_min_size
    mon = var.eks_mon_min_size
    spot = var.eks_spot_min_size
    db = var.eks_db_min_size
  }
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

