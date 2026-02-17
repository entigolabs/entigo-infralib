variable "prefix" {
  type = string
}

variable "kubernetes_version" {
  type    = string
  default = "1.34."
}

variable "preserve_kubernetes_version" {
  type        = bool
  default     = false
  description = "Preserve existing node pool Kubernetes version if valid, otherwise use latest stable version. Must be false if cluster does not exist yet."
}

variable "master_ipv4_cidr_block" {
  type    = string
  default = ""
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
  type     = bool
  nullable = false
  default  = false
}

variable "datapath_provider" {
  description = "ADVANCED_DATAPATH (Dataplane V2) or DATAPATH_PROVIDER_UNSPECIFIED (legacy)"
  type        = string
  default     = "ADVANCED_DATAPATH"
}

variable "network_policy" {
  description = "Enable Calico network policy (requires datapath_provider = DATAPATH_PROVIDER_UNSPECIFIED)"
  type        = bool
  default     = false
}

variable "enable_cilium_clusterwide_network_policy" {
  description = "Enable Cilium network policy (requires datapath_provider = ADVANCED_DATAPATH)"
  type        = bool
  default     = false
}

variable "monitoring_enable_observability_metrics" {
  description = "Enable Dataplane V2 Metrics (requires datapath_provider = ADVANCED_DATAPATH)"
  type        = bool
  default     = false
}

variable "monitoring_enable_observability_relay" {
  description = "Enable Dataplane V2 Observability (requires datapath_provider = ADVANCED_DATAPATH)"
  type        = bool
  default     = false
}

variable "monitoring_enable_managed_prometheus" {
  type    = bool
  default = false
}

variable "monitoring_enabled_components" {
  type    = list(string)
  default = ["SYSTEM_COMPONENTS"]
}

variable "logging_enabled_components" {
  type    = list(string)
  default = ["SYSTEM_COMPONENTS"]
}

variable "dns_allow_external_traffic" {
  type     = bool
  nullable = false
  default  = false
}

variable "deploy_using_private_endpoint" {
  type    = bool
  default = true
}

variable "enable_private_endpoint" {
  type     = bool
  nullable = false
  default  = true
}

variable "gcp_public_cidrs_access_enabled" {
  type     = bool
  nullable = false
  default  = false
}

variable "enable_l4_ilb_subsetting" {
  type     = bool
  nullable = false
  default  = false
}

variable "grant_registry_access" {
  type    = bool
  default = true
}

variable "registry_project_ids" {
  type    = list(string)
  default = []
}

variable "master_authorized_networks" {
  type = list(object({
    cidr_block   = string
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
  type     = number
  nullable = false
  default  = 2
}

variable "gke_main_max_size" {
  type     = number
  nullable = false
  default  = 4
}

variable "gke_main_instance_type" {
  type    = string
  default = "e2-standard-2"
}

variable "gke_main_node_locations" {
  type    = string
  default = ""
}

variable "gke_main_location_policy" {
  type    = string
  default = "BALANCED"
}

variable "gke_main_spot_nodes" {
  type    = bool
  default = false
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

variable "gke_main_max_surge" {
  type    = number
  default = 1
}

variable "gke_mon_min_size" {
  type     = number
  nullable = false
  default  = 1
}

variable "gke_mon_max_size" {
  type     = number
  nullable = false
  default  = 3
}

variable "gke_mon_instance_type" {
  type    = string
  default = "e2-standard-2"
}

variable "gke_mon_node_locations" {
  type    = string
  default = ""
}

variable "gke_mon_location_policy" {
  type    = string
  default = "BALANCED"
}

variable "gke_mon_spot_nodes" {
  type    = bool
  default = false
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

variable "gke_mon_max_surge" {
  type    = number
  default = 1
}

variable "gke_tools_min_size" {
  type     = number
  nullable = false
  default  = 2
}

variable "gke_tools_max_size" {
  type     = number
  nullable = false
  default  = 3
}

variable "gke_tools_instance_type" {
  type    = string
  default = "e2-standard-2"
}

variable "gke_tools_node_locations" {
  type    = string
  default = ""
}

variable "gke_tools_location_policy" {
  type    = string
  default = "BALANCED"
}

variable "gke_tools_spot_nodes" {
  type    = bool
  default = false
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

variable "gke_tools_max_surge" {
  type    = number
  default = 1
}

variable "gke_managed_node_groups_extra" {
  type     = list(any)
  nullable = false
  default  = []
}

variable "boot_disk_kms_key" {
  type    = string
  default = ""
}

variable "database_encryption_kms_key" {
  type    = string
  default = ""
}

variable "gce_pd_csi_driver" {
  type    = bool
  default = true
}

variable "gcs_fuse_csi_driver" {
  type    = bool
  default = false
}

variable "filestore_csi_driver" {
  type    = bool
  default = false
}
