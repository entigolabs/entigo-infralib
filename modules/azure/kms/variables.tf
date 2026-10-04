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
}

variable "delete_lock_enabled" {
  type        = bool
  default     = true
  description = "CanNotDelete lock on the vault against deletion outside Terraform"
}

variable "sku_name" {
  type        = string
  default     = "premium"
  description = "premium costs the same for software keys and allows RSA-HSM"
}

variable "key_type" {
  type        = string
  default     = "RSA"
  description = "RSA (software) or RSA-HSM. Changing it replaces the keys"
}

variable "key_size" {
  type    = number
  default = 3072
}

variable "key_rotation_period" {
  type        = string
  default     = null
  description = "ISO 8601, e.g. P90D, min 7 days, null = no rotation"
  validation {
    condition     = var.key_rotation_period == null || (can(regex("^P(\\d+Y)?(\\d+M)?(\\d+D)?$", var.key_rotation_period)) && (can(regex("[1-9]\\d*[YM]", var.key_rotation_period)) || tonumber(try(regex("(\\d+)D$", var.key_rotation_period)[0], "0")) >= 7))
    error_message = "key_rotation_period must be an ISO 8601 duration in years, months and days (e.g. P90D, P1Y) of at least 7 days."
  }
}

variable "admin_object_ids" {
  type    = list(string)
  default = []
}

variable "data_key_users" {
  type    = map(string)
  default = {}
}

variable "config_key_users" {
  type    = map(string)
  default = {}
}

variable "telemetry_key_users" {
  type    = map(string)
  default = {}
}
