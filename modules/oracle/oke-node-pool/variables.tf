variable "prefix" {
  type = string
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
  type    = number
  default = 2
}

variable "memory_in_gbs" {
  type    = number
  default = 16
}

variable "node_count" {
  type    = number
  default = 3
}

variable "boot_volume_size_in_gbs" {
  type    = string
  default = "50"
}

variable "node_pool_os_type" {
  description = "Operating system family used to pick the node image, e.g. Oracle Linux."
  type        = string
  default     = "ORACLE_LINUX"
}
