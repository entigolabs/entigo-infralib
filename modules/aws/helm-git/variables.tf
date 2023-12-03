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

variable "path" {
  type = string
}

variable "values" {
  type = string
}

variable "name" {
  type = string
  default = ""
}

variable "namespace" {
  type = string
  default = ""
}

variable "create_namespace" {
  type = bool
  default = true
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
