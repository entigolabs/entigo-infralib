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

# https://learn.microsoft.com/azure/application-gateway/for-containers/container-networking#limitations
variable "subnet_id" {
  type        = string
  description = "Association subnet, a /24 delegated to Microsoft.ServiceNetworking/trafficControllers used by no other AGC: default vpc private_agc_subnets[0], vpc control_agc_subnets or service_agc_subnets in spoke mode"
}

variable "private_endpoint_subnet_id" {
  type        = string
  description = "Subnet of the private endpoint, the clients' IP of the private frontend (vpc control_subnets / service_subnets in spoke mode)"
}

variable "private_dns_zone_id" {
  type        = string
  description = "privatelink.alb.azure.com zone (dns agc_private_dns_zone_id) for the frontend FQDN"
}

variable "alb_controller_principal_id" {
  type        = string
  description = "ALB controller identity (aks alb_controller_principal_id)"
  validation {
    condition     = var.alb_controller_principal_id != ""
    error_message = "alb_controller_principal_id is empty: the aks module creates the ALB controller identity only with agc_subnet_ids."
  }
}

variable "tags" {
  type    = map(string)
  default = {}
}
