variable "prefix" {
  type = string
}

variable "name" {
  type    = string
  default = ""
}

variable "cluster_id" {
  type = string
}

variable "kubernetes_version" {
  type    = string
  default = null
}

variable "vnet_subnet_id" {
  type = string
}

variable "pod_subnet_id" {
  type        = string
  default     = ""
  description = "Azure CNI Pod Subnet: vpc pod subnet of this pool (default the cluster's), \"\" on an Overlay cluster. Changing it rotates the pool"
}

variable "cluster_pod_subnet_id" {
  type        = string
  default     = ""
  description = "The cluster's pod subnet (aks pod_subnet_id): a pool on another pod subnet gives the cluster identity Network Contributor on it"
}

variable "cluster_identity_principal_id" {
  type        = string
  default     = ""
  description = "Principal of the cluster identity (aks identity_principal_id), for the pod subnet role"
}

variable "mode" {
  type    = string
  default = "User"
}

variable "min_size" {
  type    = number
  default = 1
}

variable "max_size" {
  type    = number
  default = 3
}

variable "instance_type" {
  type    = string
  default = "Standard_D2s_v6"
}

variable "location" {
  type    = string
  default = ""
}

variable "availability_zones" {
  type    = list(string)
  default = null
}

variable "spot_nodes" {
  type    = bool
  default = false
}

variable "spot_max_price" {
  type        = number
  default     = -1
  description = "-1 = up to the on-demand price"
}

variable "os_sku" {
  type    = string
  default = "AzureLinux"
}

variable "volume_size" {
  type    = number
  default = 50
}

variable "volume_type" {
  type        = string
  default     = "Managed"
  description = "Managed or Ephemeral"
}

variable "max_pods" {
  type    = number
  default = 64
}

variable "taints" {
  type        = list(string)
  default     = []
  description = "key=value:Effect"
}

variable "labels" {
  type    = map(string)
  default = {}
}

variable "tags" {
  type    = map(string)
  default = {}
}

variable "max_surge" {
  type    = string
  default = "10%"
}

variable "upgrade_settings" {
  type = object({
    max_surge                     = string
    drain_timeout_in_minutes      = optional(number)
    node_soak_duration_in_minutes = optional(number)
  })
  default     = null
  description = "Overrides max_surge. Ignored for spot pools"
}
