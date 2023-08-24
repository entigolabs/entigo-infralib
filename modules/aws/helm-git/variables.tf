variable "prefix" {
  type = string
}

variable "repository" {
  type = string
}

variable "branch" {
  type = string
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
