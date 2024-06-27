variable "prefix" {
  type = string
}

variable "ksa_name" {
  type = string
  description = "Kubernetes service account name for crossplane"
  default = "crossplane"
}

variable "kns_name" {
  type = string
  description = "Kubernetes namespace name for crossplane"
  default = "crossplane-system"
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
