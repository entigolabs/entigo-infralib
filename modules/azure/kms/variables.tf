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

variable "tags" {
  type    = map(string)
  default = {}
}

variable "purge_protection_enabled" {
  type        = bool
  default     = true
  description = "Required by disk encryption sets and storage CMK"
}

variable "soft_delete_retention_days" {
  type    = number
  default = 30
  validation {
    condition     = var.soft_delete_retention_days >= 7 && var.soft_delete_retention_days <= 90
    error_message = "soft_delete_retention_days must be between 7 and 90."
  }
}

variable "delete_lock_enabled" {
  type        = bool
  default     = true
  description = "CanNotDelete lock on the vault (like google/kms prevent_destroy), needs Owner or User Access Administrator"
}

variable "sku_name" {
  type        = string
  default     = "premium"
  description = "premium costs the same for software keys and allows RSA-HSM"
  validation {
    condition     = contains(["standard", "premium"], var.sku_name)
    error_message = "sku_name must be standard or premium."
  }
}

variable "key_type" {
  type        = string
  default     = "RSA"
  description = "RSA (software) or RSA-HSM. Changing it replaces the keys"
  validation {
    condition     = contains(["RSA", "RSA-HSM"], var.key_type)
    error_message = "key_type must be RSA or RSA-HSM."
  }
}

variable "key_size" {
  type    = number
  default = 3072
}

variable "key_rotation_period" {
  type        = string
  default     = null
  description = "ISO 8601, e.g. P90D, null = no rotation"
}

variable "admin_object_ids" {
  type        = list(string)
  default     = []
  description = "Key Vault Crypto Officer, the caller is always added"
}

variable "data_key_users" {
  type        = map(string)
  default     = {}
  description = "Additional principals with a role on the key, name => object id. Consumer modules grant their own identities"
}

variable "config_key_users" {
  type        = map(string)
  default     = {}
  description = "Additional principals with a role on the key, name => object id. Consumer modules grant their own identities"
}

variable "telemetry_key_users" {
  type        = map(string)
  default     = {}
  description = "Additional principals with a role on the key, name => object id. Consumer modules grant their own identities"
}
