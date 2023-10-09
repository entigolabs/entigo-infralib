variable "prefix" {
  type = string
}

variable "eks_oidc_provider" {
  type = string
}

variable "eks_oidc_provider_arn" {
  type = string
}

variable "region" {
  type = string
}

variable "account" {
  type = string
}


locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
