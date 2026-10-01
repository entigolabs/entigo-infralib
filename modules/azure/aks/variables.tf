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
  type        = string
  default     = "1.35"
  description = "Minor version, AKS picks the latest patch"
}

variable "sku_tier" {
  type        = string
  default     = "Standard"
  description = "Free, Standard (uptime SLA) or Premium (long term support)"
}

variable "upgrade_channel" {
  type        = string
  default     = "none"
  description = "none or patch, stable/rapid would change the minor version under kubernetes_version"
  validation {
    condition     = contains(["none", "patch"], var.upgrade_channel)
    error_message = "upgrade_channel must be none or patch."
  }
}

variable "node_os_upgrade_channel" {
  type        = string
  default     = "None"
  description = "None, NodeImage, SecurityPatch or Unmanaged. None = new node images only with Kubernetes upgrades or az aks nodepool upgrade --node-image-only"
  validation {
    condition     = contains(["None", "Unmanaged", "SecurityPatch", "NodeImage"], var.node_os_upgrade_channel)
    error_message = "node_os_upgrade_channel must be None, Unmanaged, SecurityPatch or NodeImage."
  }
}

variable "maintenance_window" {
  type = object({
    day_of_week    = optional(string, "Sunday")
    start_time     = optional(string, "00:00")
    duration_hours = optional(number, 4)
    utc_offset     = optional(string, "+00:00")
  })
  default     = {}
  description = "Weekly window for upgrade_channel patch and the NodeImage/SecurityPatch channels, 4-24 hours"
}

variable "pod_cidr" {
  type        = string
  default     = "10.10.0.0/16"
  description = "CNI Overlay pod range, must not overlap the VNet"
}

variable "service_cidr" {
  type    = string
  default = "10.11.0.0/16"
}

variable "dns_service_ip" {
  type    = string
  default = "10.11.0.10"
}

variable "outbound_type" {
  type    = string
  default = "userAssignedNATGateway"
}

variable "api_server_vnet_integration_enabled" {
  type        = bool
  default     = true
  description = "API server IP in api_server_subnet_id. One-way on an existing cluster, needs az aks stop/start"
}

variable "api_server_subnet_id" {
  type    = string
  default = ""
}

variable "private_cluster_enabled" {
  type        = bool
  default     = true
  description = "No public API server endpoint, agent steps then need vpc attach"
  validation {
    condition     = !var.private_cluster_enabled || var.api_server_vnet_integration_enabled
    error_message = "private_cluster_enabled requires api_server_vnet_integration_enabled."
  }
}

variable "api_server_authorized_ip_ranges" {
  type = list(string)
  default = [
    "13.51.186.14/32",  # Entigo VPN 1
    "13.53.208.166/32", # Entigo VPN 2
  ]
  description = "Public endpoint allowlist, [] = open"
}

variable "admin_object_ids" {
  type        = list(string)
  default     = []
  description = "Entra ID object ids with Azure Kubernetes Service RBAC Cluster Admin, the caller is always added"
}

variable "disable_local_accounts" {
  type    = bool
  default = true
}

variable "acr_id" {
  type        = string
  default     = ""
  description = "Registry with AcrPull for the kubelet identity"
}

variable "bootstrap_cache_enabled" {
  type        = bool
  default     = false
  description = "Nodes pull AKS system images from acr_id (Premium, private endpoint, aks_bootstrap_cache_rule). Existing cluster: needs a node image upgrade of all pools"
}

variable "agc_subnet_ids" {
  type        = list(string)
  default     = []
  description = "Application Gateway for Containers subnets (vpc agc_subnets), creates the ALB controller identity for the azure-gateway k8s module"
}

variable "kms_key_vault_id" {
  type    = string
  default = ""
}

variable "disk_encryption_key_id" {
  type        = string
  default     = ""
  description = "Versionless key id for the disk encryption set (node OS and PVC disks). Only at cluster creation"
}

variable "control_plane_logs_enabled" {
  type        = bool
  default     = false
  description = "Diagnostic setting to a Log Analytics workspace (Microsoft managed keys)"
}

variable "control_plane_log_categories" {
  type        = list(string)
  default     = ["kube-apiserver", "guard"]
  description = "guard = Entra ID authentication, like aws/eks api + authenticator"
}

variable "control_plane_logs_retention_days" {
  type    = number
  default = 30
}

variable "availability_zones" {
  type        = list(string)
  default     = null
  description = "null = all zones where the pool's VM size is available, [] = no zones. Fixed at pool creation"
}

variable "max_surge" {
  type        = string
  default     = "10%"
  description = "Upgrade surge of non-spot pools"
}

variable "node_resource_group_name" {
  type        = string
  default     = ""
  description = "\"\" = <prefix>-nodes-<location>. Only at cluster creation"
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
  default     = null
  description = "AKS cluster autoscaler settings for all pools, null = AKS defaults"
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
  type    = number
  default = 2
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

variable "aks_managed_node_groups_extra" {
  type        = any
  default     = {}
  description = "Extra pools in the AVM agent_pools format"
}

variable "tags" {
  type    = map(string)
  default = {}
}
