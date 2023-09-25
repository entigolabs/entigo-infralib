variable "prefix" {
  type = string
}

variable "repository" {
  type = string
  default = "https://github.com/entigolabs/entigo-infralib-release.git"
}

variable "branch" {
  type = string
  default = "main"
}

variable "name" {
  type = string
  default = ""
}

variable "namespace" {
  type = string
  default = ""
}

variable "hostname" {
  type = string
}

variable "create_namespace" {
  type = bool
  default = true
}

variable "argocd_apps_name" {
  type = string
  default = "ep-applications"
}

variable "install_crd" {
  type = bool
  default = true
}

variable "ingress_group_name" {
  type = string
  default = "internal"
}

variable "ingress_scheme" {
  type = string
  default = "internal"
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
