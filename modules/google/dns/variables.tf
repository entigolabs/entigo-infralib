variable "prefix" {
  type = string
}

variable "vpc_id" {
  type    = string
  default = ""
}

variable "create_public" {
  type    = bool
  default = true
}

variable "create_private" {
  type    = bool
  default = true
}


variable "parent_zone_id" {
  type    = string
  default = ""
}

variable "parent_domain" {
  type    = string
  default = ""
}

variable "public_subdomain_name" {
  type    = string
  default = ""
}

variable "private_subdomain_name" {
  type    = string
  default = ""
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
