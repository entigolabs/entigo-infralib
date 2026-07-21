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
