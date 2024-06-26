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

variable "ksa_name" {
  type = string
  description = "Kubernetes service account name for external-dns"
  default = "external-dns"
}

variable "kns_name" {
  type = string
  description = "Kubernetes namespace name for external-dns"
  default = "external-dns"
}


locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
