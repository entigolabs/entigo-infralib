variable "prefix" {
  type = string
}

variable "zone" {
    type = string
}

variable "domain" {
    type = string
}

variable "type" {
    type = string
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
