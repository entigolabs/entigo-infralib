variable "prefix" {
  type = string
}

variable "project_id" {
  type = string
}

variable "region" {
  type = string
}

variable "node_locations" {
  type = string
  default = ""
}

variable "master_ipv4_cidr_block" {
  type = string
  default = "10.1.0.0/28"
}

variable "kubernetes_version" {
  type = string
  default = "1.27.11-gke.1062004"
}

variable "machine_type" {
  type = string
  default = "e2-medium"
}

variable "master_authorized_networks" {
  type = list(object({
    cidr_block = string
    display_name = string
  }))
  default = [
    {
      display_name = "Whitelist 1 - Entigo VPN"
      cidr_block   = "13.51.186.14/32"
    },
    {
      display_name = "Whitelist 2 - Entigo VPN"
      cidr_block   = "13.53.208.166/32"
    }
  ]
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}

