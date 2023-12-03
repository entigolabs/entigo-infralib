variable "prefix" {
  type = string
}

variable "eks_oidc_provider" {
  type = string
}

variable "eks_oidc_provider_arn" {
  type = string
}


locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
