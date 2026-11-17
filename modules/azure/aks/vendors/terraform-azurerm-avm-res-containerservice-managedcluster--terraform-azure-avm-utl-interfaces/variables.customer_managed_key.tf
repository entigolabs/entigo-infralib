variable "customer_managed_key" {
  type = object({
    key_vault_resource_id = string
    key_name              = string
    key_version           = optional(string, null)
    user_assigned_identity = optional(object({
      resource_id = string
    }), null)
  })
  default     = null
  description = <<DESCRIPTION
An object containing the following attributes:

- `key_vault_resource_id` - The resource ID of the key vault.
- `key_name` - The name of the key.
- `key_version` - (Optional) The version of the key. If not provided, the latest version will be used.
- `user_assigned_identity` - (Optional) An object containing the resource ID of the user assigned identity.
  - `resource_id` - The resource ID of the user assigned identity.
DESCRIPTION
}

variable "customer_managed_key_key_vault_domain" {
  type        = string
  default     = "vault.azure.net"
  description = "The domain name for the key vault. Default is `vault.azure.net`."
  nullable    = false
}
