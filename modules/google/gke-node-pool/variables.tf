variable "prefix" {
  type = string
}

variable "kubernetes_version" {
  type    = string
  default = "1.33."
}

variable "preserve_kubernetes_version" {
  type        = bool
  default     = false
  description = "Preserve existing node pool Kubernetes version if valid, otherwise use latest stable version. Must be false if cluster does not exist yet."
}

variable "cluster_name" {
  type = string
}

variable "cluster_region" {
  type = string
}

variable "node_count" {
  type    = number
  default = null
}

variable "auto_repair" {
  type    = bool
  default = true
}

variable "auto_upgrade" {
  type    = bool
  default = false
}

variable "min_size" {
  type    = number
  default = 1
}

variable "max_size" {
  type    = number
  default = 3
}

variable "location_policy" {
  type    = string
  default = "BALANCED"
}

variable "autoscaling" {
  description = "If set then min_size, max_size and location_policy are ignored."
  type        = any
  default     = null
}

variable "instance_type" {
  type    = string
  default = "e2-standard-2"
}

variable "node_locations" {
  type    = list(string)
  default = []
}

variable "spot_nodes" {
  type    = bool
  default = false
}

variable "volume_size" {
  type    = number
  default = 50
}

variable "max_pods_per_node" {
  type    = number
  default = null
}

variable "volume_type" {
  type    = string
  default = "pd-standard"
}

variable "service_account" {
  type    = string
  default = ""
}

variable "boot_disk_kms_key" {
  type    = string
  default = ""
}

variable "taints" {
  type = list(object({
    key    = string
    value  = string
    effect = string
  }))
  default = []
}

variable "labels" {
  type    = map(string)
  default = {}
}

variable "tags" {
  type    = list(string)
  default = []
}

variable "node_config" {
  type    = any
  default = null
}

variable "max_surge" {
  type    = number
  default = 1
}

variable "max_unavailable" {
  type    = number
  default = 0
}

variable "upgrade_settings" {
  description = "If set then max_surge and max_unavailable are ignored."
  type        = any
  default     = null
}

variable "network_config" {
  type    = any
  default = null
}

variable "placement_policy" {
  type    = any
  default = null
}

variable "queued_provisioning" {
  type    = any
  default = null
}

variable "timeouts" {
  type    = any
  default = {}
}

variable "grant_boot_disk_kms_key_access_to_service_account" {
  type    = bool
  default = false
}
