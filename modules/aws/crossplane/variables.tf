variable "prefix" {
  type = string
}

variable "eks_oidc_provider" {
  type = string
}

variable "eks_oidc_provider_arn" {
  type = string
}

variable "eks_region" {
  type = string
}

variable "eks_account" {
  type = string
}


locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
