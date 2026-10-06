variable "prefix" {
  type = string
}

variable "resource_group_name" {
  type    = string
  default = ""
}

variable "location" {
  type    = string
  default = ""
}

variable "vnet_subnet_id" {
  type        = string
  description = "Node subnet, needs the vpc NAT gateway (outbound_type userAssignedNATGateway)"
}

variable "kubernetes_version" {
  type    = string
  default = "1.36"
}

variable "sku_tier" {
  type        = string
  default     = "Standard"
  description = "Free, Standard (uptime SLA) or Premium (long term support)"
}

variable "upgrade_channel" {
  type    = string
  default = "none"
}

variable "node_os_upgrade_channel" {
  type    = string
  default = "None"
}

variable "maintenance_window" {
  type = object({
    day_of_week    = optional(string, "Sunday")
    start_time     = optional(string, "00:00")
    duration_hours = optional(number, 4)
    utc_offset     = optional(string, "+00:00")
  })
  default = {}
}

variable "pod_cidr" {
  type        = string
  default     = "10.244.0.0/16"
  description = "CNI Overlay pod range, must not overlap the VNet"
}

variable "service_cidr" {
  type    = string
  default = "10.96.0.0/16"
}

variable "dns_service_ip" {
  type    = string
  default = "10.96.0.10"
}

variable "outbound_type" {
  type    = string
  default = "userAssignedNATGateway"
}

variable "api_server_vnet_integration_enabled" {
  type    = bool
  default = true
}

variable "api_server_subnet_id" {
  type    = string
  default = ""
}

variable "private_cluster_enabled" {
  type    = bool
  default = true
}

# https://learn.microsoft.com/azure/aks/private-clusters#configuration-options-for-private-dns
variable "private_dns_zone_id" {
  type        = string
  default     = ""
  description = "Central private DNS zone for the private API server, private.<location>.azmk8s.io or <subzone>.private.<location>.azmk8s.io, create time only. \"\" = AKS creates a zone per cluster"
}

variable "api_server_authorized_ip_ranges" {
  type = list(string)
  default = [
    "13.51.186.14/32",  # Entigo VPN 1
    "13.53.208.166/32", # Entigo VPN 2
  ]
}

variable "nat_public_ips" {
  type    = list(string)
  default = []
}

variable "admin_object_ids" {
  type    = list(string)
  default = []
}

variable "disable_local_accounts" {
  type    = bool
  default = true
}

variable "acr_id" {
  type    = string
  default = ""
}

variable "bootstrap_cache_enabled" {
  type    = bool
  default = false
}

variable "agc_subnet_ids" {
  type    = list(string)
  default = []
}

variable "disk_encryption_key_id" {
  type    = string
  default = ""
}

variable "disk_encryption_key_resource_id" {
  type    = string
  default = ""
  validation {
    condition     = var.disk_encryption_key_id == "" || var.disk_encryption_key_resource_id != ""
    error_message = "disk_encryption_key_id needs disk_encryption_key_resource_id (kms data_key_resource_id)."
  }
}

variable "control_plane_logs_enabled" {
  type    = bool
  default = false
}

variable "control_plane_log_categories" {
  type    = list(string)
  default = ["kube-apiserver", "guard"]
}

variable "control_plane_logs_retention_days" {
  type    = number
  default = 30
}

variable "availability_zones" {
  type    = list(string)
  default = null
}

variable "max_surge" {
  type    = string
  default = "10%"
}

variable "node_resource_group_name" {
  type    = string
  default = ""
}

variable "kubelet_additional_role_assignments" {
  type = list(object({
    scope = string
    role  = string
  }))
  default = []
}

variable "auto_scaler_profile" {
  type = object({
    balance_similar_node_groups           = optional(string)
    daemonset_eviction_for_empty_nodes    = optional(bool)
    daemonset_eviction_for_occupied_nodes = optional(bool)
    expander                              = optional(string)
    ignore_daemonsets_utilization         = optional(bool)
    max_empty_bulk_delete                 = optional(string)
    max_graceful_termination_sec          = optional(string)
    max_node_provision_time               = optional(string)
    max_total_unready_percentage          = optional(string)
    new_pod_scale_up_delay                = optional(string)
    ok_total_unready_count                = optional(string)
    scale_down_delay_after_add            = optional(string)
    scale_down_delay_after_delete         = optional(string)
    scale_down_delay_after_failure        = optional(string)
    scale_down_unneeded_time              = optional(string)
    scale_down_unready_time               = optional(string)
    scale_down_utilization_threshold      = optional(string)
    scan_interval                         = optional(string)
    skip_nodes_with_local_storage         = optional(string)
    skip_nodes_with_system_pods           = optional(string)
  })
  default = null
}

variable "os_sku" {
  type    = string
  default = "AzureLinux"
}

variable "aks_main_instance_type" {
  type    = string
  default = "Standard_D2s_v6"
}

variable "aks_main_min_size" {
  type    = number
  default = 2
}

variable "aks_main_max_size" {
  type        = number
  default     = 6
  description = "0 = no main pool"
}

variable "aks_main_volume_type" {
  type        = string
  default     = "Managed"
  description = "Managed or Ephemeral"
}

variable "aks_main_volume_size" {
  type    = number
  default = 100
}

variable "aks_main_max_pods" {
  type    = number
  default = 64
}

variable "aks_main_spot_nodes" {
  type    = bool
  default = false
}

variable "aks_mon_instance_type" {
  type    = string
  default = "Standard_D2s_v6"
}

variable "aks_mon_min_size" {
  type    = number
  default = 2
}

variable "aks_mon_max_size" {
  type        = number
  default     = 6
  description = "0 = no mon pool"
}

variable "aks_mon_volume_type" {
  type        = string
  default     = "Managed"
  description = "Managed or Ephemeral"
}

variable "aks_mon_volume_size" {
  type    = number
  default = 50
}

variable "aks_mon_max_pods" {
  type    = number
  default = 64
}

variable "aks_mon_spot_nodes" {
  type    = bool
  default = false
}

variable "aks_tools_instance_type" {
  type    = string
  default = "Standard_D2s_v6"
}

variable "aks_tools_min_size" {
  type        = number
  default     = 2
  description = "System pool, at least 1"
}

variable "aks_tools_max_size" {
  type    = number
  default = 6
}

variable "aks_tools_volume_type" {
  type        = string
  default     = "Managed"
  description = "Managed or Ephemeral"
}

variable "aks_tools_volume_size" {
  type    = number
  default = 50
}

variable "aks_tools_max_pods" {
  type    = number
  default = 64
}

variable "tags" {
  type    = map(string)
  default = {}
}
