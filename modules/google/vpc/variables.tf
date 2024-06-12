variable "prefix" {
  type = string
}

variable "subnet_cidr" {
  type     = string
  default  = "10.0.0.0/16"
}

variable "secondary_cidr_pods" {
  type    = string
  default = "10.4.0.0/14"
}

variable "secondary_cidr_services" {
  type    = string
  default = "10.8.0.0/20"
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}

