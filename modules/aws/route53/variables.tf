variable "prefix" {
  type = string
}

variable "vpc_prefix" {
  type = string
}

variable "create_public" {
  type = bool
  default = true
}

variable "create_cert" {
  type = bool
  default = true
}

variable "create_private" {
  type = bool
  default = true
}

variable "parent_zone_id" {
  type = string
  default = ""
}

variable "parent_domain" {
  type = string
  default = ""
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
