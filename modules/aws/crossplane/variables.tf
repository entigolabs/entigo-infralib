variable "prefix" {
  type = string
}

variable "eks_prefix" {
  type = string
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
