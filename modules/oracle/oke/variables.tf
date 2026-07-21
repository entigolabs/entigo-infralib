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

variable "endpoint_subnet_id" {
  description = "Subnet for the Kubernetes API endpoint. A private subnet is recommended; use is_public_ip_enabled to also assign a public IP on the same endpoint."
  type        = string
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
