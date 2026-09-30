variable "prefix" {
  type = string
}

variable "resource_group_name" {
  description = "Resource group for all resources."
  type        = string
}

variable "location" {
  description = "Azure region, defaults to the agent location."
  type        = string
}

variable "sku" {
  type        = string
  default     = "Standard"
  description = "Basic has no cache rules. Premium adds private endpoints, customer managed keys and untagged retention"
  validation {
    condition     = contains(["Standard", "Premium"], var.sku)
    error_message = "sku must be Standard or Premium (Basic has no artifact cache rules)."
  }
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
  type        = string
  default     = ""
  description = "RBAC Key Vault with the registry credentials, the agent vault by default"
}

variable "hub_username_secret" {
  type        = string
  default     = ""
  description = "Key Vault secret name, \"\" = anonymous. Docker Hub has no cache without credentials"
}

variable "hub_access_token_secret" {
  type        = string
  default     = ""
  description = "Key Vault secret name, \"\" = anonymous"
}

variable "ghcr_username_secret" {
  type        = string
  default     = ""
  description = "Key Vault secret name, \"\" = anonymous"
}

variable "ghcr_access_token_secret" {
  type        = string
  default     = ""
  description = "Key Vault secret name, \"\" = anonymous"
}

variable "gcr_username_secret" {
  type        = string
  default     = ""
  description = "Key Vault secret name, \"\" = anonymous"
}

variable "gcr_access_token_secret" {
  type        = string
  default     = ""
  description = "Key Vault secret name, \"\" = anonymous"
}

variable "quay_username_secret" {
  type        = string
  default     = ""
  description = "Key Vault secret name, \"\" = anonymous"
}

variable "quay_access_token_secret" {
  type        = string
  default     = ""
  description = "Key Vault secret name, \"\" = anonymous"
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
  type    = string
  default = null
}

variable "private_endpoint_vnet_id" {
  type    = string
  default = null
}

variable "public_network_access_enabled" {
  type        = bool
  default     = true
  description = "false = private endpoints only"
}

variable "aks_bootstrap_cache_rule" {
  type        = bool
  default     = false
  description = "Cache rule for the AKS bootstrap artifact cache, needs the private endpoint"
}

variable "tags" {
  type    = map(string)
  default = {}
}
