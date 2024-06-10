variable "prefix" {
  type = string
}

variable "region" {
  type = string
  default = "europe-north1"
}

variable "subnet_cidr" {
  type     = string
  default  = "244.178.0.0/16"
}

variable "secondary_cidr_pods" {
  type    = string
  default = "172.20.72.0/21"
}

variable "secondary_cidr_services" {
  type    = string
  default = "10.127.8.0/24"
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
