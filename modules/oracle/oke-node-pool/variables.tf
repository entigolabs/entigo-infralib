variable "prefix" {
  type = string
}

variable "pool_name" {
  description = "Suffix appended to prefix to name this node pool, e.g. main/mon/tools for the cluster's default pools, or a custom name for additional pools."
  type        = string
  default     = "pool"
}

variable "compartment_id" {
  description = "OCID of the compartment that will contain the node pool."
  type        = string
}

variable "cluster_id" {
  type = string
}

variable "kubernetes_version" {
  type = string
}

variable "subnet_ids" {
  description = "Subnets nodes are placed in - one per availability domain, in order. Reused across ADs if fewer are given than ADs available."
  type        = list(string)
}

variable "node_shape" {
  type    = string
  default = "VM.Standard.E4.Flex"
}

variable "ocpus" {
  # 1 OCPU = 2 vCPU-equivalent threads, so this plus 8GB memory below matches an AWS
  # t3.large (2 vCPU / 8GB) - not 2 OCPUs, which would be ~4 vCPU-equivalent.
  type    = number
  default = 1
}

variable "memory_in_gbs" {
  type    = number
  default = 8
}

variable "node_count" {
  type    = number
  default = 3
}

variable "boot_volume_size_in_gbs" {
  type    = string
  default = "50"
}

variable "labels" {
  description = "Kubernetes node labels applied to every node in the pool, in addition to the created-by label this module always sets."
  type        = map(string)
  default     = {}
}

variable "nsg_ids" {
  description = "Network security groups applied to every node's VNIC."
  type        = list(string)
  default     = []
}

variable "node_pool_os_type" {
  description = "Operating system family used to pick the node image. Valid values: OL7, OL8, UBUNTU."
  type        = string
  default     = "OL8"
}
