variable "prefix" {
  type = string
}

variable "master_ipv4_cidr_block" {
  type = string
  default = "10.1.0.0/28"
}

variable "network" {
  type = string
}

variable "subnetwork" {
  type = string
}

variable "ip_range_pods" {
  type = string
}

variable "ip_range_services" {
  type = string
}

variable "master_global_access_enabled" {
  type    = bool
  nullable = false
  default = false
}

variable "enable_private_endpoint" {
  type    = bool
  nullable = false
  default = false
}

variable "kubernetes_version" {
  type = string
  default = "1.28.9-gke.1000000"
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

variable "gke_main_min_size" {
  type    = number
  nullable = false
  default = 2
}

variable "gke_main_max_size" {
  type    = number
  nullable = false
  default = 4
}

variable "gke_main_instance_type" {
  type    = string
  default = "e2-medium"
}

variable "gke_main_volume_size" {
  type    = number
  default = 100
}

variable "gke_main_max_pods" {
  type    = number
  default = 64
}

variable "gke_main_volume_type" {
  type    = string
  default = "pd-standard"
}

variable "gke_mainarm_min_size" {
  type    = number
  nullable = false
  default = 0
}

variable "gke_mainarm_max_size" {
  type    = number
  nullable = false
  default = 0
}

variable "gke_mainarm_instance_type" {
  type    = string
  default = "e2-medium"
}

variable "gke_mainarm_volume_size" {
  type    = number
  default = 100
}

variable "gke_mainarm_max_pods" {
  type    = number
  default = 64
}

variable "gke_mainarm_volume_type" {
  type    = string
  default = "pd-standard"
}

variable "gke_spot_min_size" {
  type    = number
  nullable = false
  default = 0
}

variable "gke_spot_max_size" {
  type    = number
  nullable = false
  default = 0
}

variable "gke_spot_instance_type" {
  type = string
  nullable = false
  default = "e2-medium"
}

variable "gke_spot_volume_size" {
  type    = number
  default = 100
}

variable "gke_spot_max_pods" {
  type    = number
  default = 64
}

variable "gke_spot_volume_type" {
  type    = string
  default = "pd-standard"
}

variable "gke_mon_min_size" {
  type    = number
  nullable = false
  default = 1
}

variable "gke_mon_max_size" {
  type    = number
  nullable = false
  default = 3
}

variable "gke_mon_instance_type" {
  type    = string
  default = "e2-medium"
}

variable "gke_mon_volume_size" {
  type    = number
  default = 50
}

variable "gke_mon_max_pods" {
  type    = number
  default = 64
}

variable "gke_mon_volume_type" {
  type    = string
  default = "pd-standard"
}

variable "gke_tools_min_size" {
  type    = number
  nullable = false
  default = 2
}

variable "gke_tools_max_size" {
  type    = number
  nullable = false
  default = 3
}

variable "gke_tools_instance_type" {
  type    = string
  default = "e2-medium"
}

variable "gke_tools_volume_size" {
  type    = number
  default = 50
}

variable "gke_tools_max_pods" {
  type    = number
  default = 64
}

variable "gke_tools_volume_type" {
  type    = string
  default = "pd-standard"
}

variable "gke_db_min_size" {
  type    = number
  nullable = false
  default = 0
}

variable "gke_db_max_size" {
  type    = number
  nullable = false
  default = 0
}

variable "gke_db_instance_type" {
  type    = string
  nullable = false
  default = "e2-medium"
}

variable "gke_db_volume_size" {
  type    = number
  default = 50
}

variable "gke_db_max_pods" {
  type    = number
  default = 64
}

variable "gke_db_volume_type" {
  type    = string
  default = "pd-standard"
}

variable "gke_managed_node_groups_extra" {
  type    = list
  nullable = false
  default = []
}



locals {
  hname = "${var.prefix}-${terraform.workspace}"
}

