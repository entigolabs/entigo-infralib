variable "prefix" {
  type = string
}

variable "kubernetes_service_account_crossplane" {
  type = string
  description = "Kubernetes service account name for crossplane"
  default = "crossplane"
}

variable "kubernetes_service_account_crossplane_google" {
  type = string
  description = "Kubernetes service account name for crossplane"
  default = "crossplane-google"
}

variable "kubernetes_namespace" {
  type = string
  description = "Kubernetes namespace name for crossplane"
  default = "crossplane-system"
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
