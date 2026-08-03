variable "prefix" {
  type = string
}

variable "compartment_id" {
  description = "OCID of the compartment that will contain the cluster."
  type        = string
}

variable "vcn_id" {
  type = string
}

variable "private_subnet_id" {
  description = "Subnet for the Kubernetes API endpoint when is_public_ip_enabled is false."
  type        = string
}

variable "public_subnet_id" {
  description = "Subnet for the Kubernetes API endpoint when is_public_ip_enabled is true. OCI requires the endpoint subnet to be public (prohibit_public_ip_on_vnic = false) whenever a public IP is assigned to it."
  type        = string
  default     = ""
}

variable "is_public_ip_enabled" {
  type     = bool
  nullable = false
  default  = false
}

variable "service_lb_subnet_ids" {
  description = "Subnets used for LoadBalancer-type Kubernetes services, typically a public subnet."
  type        = list(string)
  default     = []
}

variable "kubernetes_version" {
  description = "Defaults to the latest version OKE offers in the compartment's region if unset."
  type        = string
  default     = ""
}

variable "pods_cidr" {
  type    = string
  default = "10.244.0.0/16"
}

variable "services_cidr" {
  type    = string
  default = "10.96.0.0/16"
}

variable "node_subnet_ids" {
  description = "Default subnets nodes are placed in - one per availability domain, in order. Reused across ADs if fewer are given than ADs available. Used by main/mon/tools unless overridden per-pool below."
  type        = list(string)
  default     = []
}

# Mirrors aws/eks and google/gke, which always bundle three node groups/pools
# (main/mon/tools) by default - eks-node-group/gke-node-pool (our oke-node-pool) is only
# for *additional* custom pools beyond these three. Set a pool's node_count to 0 to skip
# creating it entirely.
variable "oke_main_node_count" {
  type     = number
  nullable = false
  default  = 1
}

variable "oke_main_ocpus" {
  type    = number
  default = 1
}

variable "oke_main_memory_in_gbs" {
  type    = number
  default = 8
}

variable "oke_main_node_shape" {
  type    = string
  default = "VM.Standard.E4.Flex"
}

variable "oke_main_node_pool_os_type" {
  type    = string
  default = "OL8"
}

variable "oke_main_boot_volume_size_in_gbs" {
  type    = string
  default = "50"
}

variable "oke_main_subnet_ids" {
  description = "Overrides node_subnet_ids for the main pool. Defaults to node_subnet_ids when empty."
  type        = list(string)
  default     = []
}

variable "oke_mon_node_count" {
  type     = number
  nullable = false
  default  = 1
}

variable "oke_mon_ocpus" {
  type    = number
  default = 1
}

variable "oke_mon_memory_in_gbs" {
  type    = number
  default = 8
}

variable "oke_mon_node_shape" {
  type    = string
  default = "VM.Standard.E4.Flex"
}

variable "oke_mon_node_pool_os_type" {
  type    = string
  default = "OL8"
}

variable "oke_mon_boot_volume_size_in_gbs" {
  type    = string
  default = "50"
}

variable "oke_mon_subnet_ids" {
  description = "Overrides node_subnet_ids for the mon pool. Defaults to node_subnet_ids when empty."
  type        = list(string)
  default     = []
}

variable "oke_tools_node_count" {
  type     = number
  nullable = false
  default  = 1
}

variable "oke_tools_ocpus" {
  type    = number
  default = 1
}

variable "oke_tools_memory_in_gbs" {
  type    = number
  default = 8
}

variable "oke_tools_node_shape" {
  type    = string
  default = "VM.Standard.E4.Flex"
}

variable "oke_tools_node_pool_os_type" {
  type    = string
  default = "OL8"
}

variable "oke_tools_boot_volume_size_in_gbs" {
  type    = string
  default = "50"
}

variable "oke_tools_subnet_ids" {
  description = "Overrides node_subnet_ids for the tools pool. Defaults to node_subnet_ids when empty."
  type        = list(string)
  default     = []
}

# cluster-autoscaler sizing, mirroring aws/eks's eks_<pool>_min/max_size naming. Default 0
# = pool not autoscaled; modules/k8s/cluster-autoscaler's agent_input_oracle.yaml only
# wires pools whose max size is set. These are pass-through metadata for the autoscaler
# (exported as outputs) - the pool resource itself ignores post-creation size changes so
# terraform never fights the autoscaler (see oke-node-pool's lifecycle comment).
variable "oke_main_min_size" {
  type     = number
  nullable = false
  default  = 0
}

variable "oke_main_max_size" {
  type     = number
  nullable = false
  default  = 0
}

variable "oke_mon_min_size" {
  type     = number
  nullable = false
  default  = 0
}

variable "oke_mon_max_size" {
  type     = number
  nullable = false
  default  = 0
}

variable "oke_tools_min_size" {
  type     = number
  nullable = false
  default  = 0
}

variable "oke_tools_max_size" {
  type     = number
  nullable = false
  default  = 0
}
