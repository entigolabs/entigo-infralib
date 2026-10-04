variable "prefix" {
  type = string
}

variable "resource_group_name" {
  type = string
}

variable "location" {
  type = string
}

variable "sku" {
  type        = string
  default     = "Standard"
  description = "Premium adds private endpoints, customer managed keys and untagged retention"
}

variable "retention_policy_in_days" {
  type    = number
  default = 7
}

variable "purge_enabled" {
  type    = bool
  default = true
}

variable "purge_untagged_days" {
  type    = number
  default = 7
}

variable "purge_tagged_days" {
  type    = number
  default = 90
}

variable "purge_schedule" {
  type        = string
  default     = "0 1 * * *"
  description = "Cron, UTC"
}

variable "key_vault_id" {
  type    = string
  default = ""
}

variable "hub_username_secret" {
  type    = string
  default = ""
}

variable "hub_access_token_secret" {
  type    = string
  default = ""
}

variable "ghcr_username_secret" {
  type    = string
  default = ""
}

variable "ghcr_access_token_secret" {
  type    = string
  default = ""
}

variable "gcr_username_secret" {
  type    = string
  default = ""
}

variable "gcr_access_token_secret" {
  type    = string
  default = ""
}

variable "quay_username_secret" {
  type    = string
  default = ""
}

variable "quay_access_token_secret" {
  type    = string
  default = ""
}

variable "encryption" {
  type = object({
    key_vault_key_id   = string
    identity_id        = string
    identity_client_id = string
  })
  default     = null
  description = "Customer managed key, Premium only and only at registry creation"
}

variable "private_endpoint_enabled" {
  type        = bool
  default     = false
  description = "Premium only"
}

variable "private_endpoint_subnet_id" {
  type     = string
  nullable = false
  default  = ""
  validation {
    condition     = !var.private_endpoint_enabled || var.private_endpoint_subnet_id != ""
    error_message = "private_endpoint_enabled needs private_endpoint_subnet_id."
  }
}

variable "private_endpoint_vnet_id" {
  type     = string
  nullable = false
  default  = ""
  validation {
    condition     = !var.private_endpoint_enabled || var.private_endpoint_vnet_id != "" || var.private_dns_zone_id != ""
    error_message = "private_endpoint_enabled needs private_endpoint_vnet_id (or private_dns_zone_id)."
  }
}

variable "private_dns_zone_id" {
  type        = string
  nullable    = false
  default     = ""
  description = "Existing privatelink.azurecr.io zone (e.g. central hub zone, already linked to the VNet), \"\" = create one"
}

variable "public_network_access_enabled" {
  type        = bool
  default     = true
  description = "false = private endpoints only"
}

variable "aks_bootstrap_cache_rule" {
  type        = bool
  default     = false
  description = "Cache rule for the AKS bootstrap artifact cache (aks bootstrap_cache_enabled needs a Premium registry with the private endpoint)"
}

variable "tags" {
  type    = map(string)
  default = {}
}
