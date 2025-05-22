variable "prefix" {
  type = string
}

# https://docs.aws.amazon.com/config/latest/APIReference/API_ResourceIdentifier.html#config-Type-ResourceIdentifier-resourceType
variable "resource_types_to_exclude" {
  type    = list(string)
  default = []
}

variable "config_logs_bucket" {
  type    = string
  default = ""
}

variable "operational_best_practices_conformance_pack_enabled" {
  type    = bool
  default = false
}

variable "iam_password_policy_enabled" {
  type    = bool
  default = false
}

variable "multi_region_cloudtrail_enabled" {
  type    = bool
  default = false
}

variable "cloudtrail_logs_bucket" {
  type    = string
  default = ""
}

variable "required_tag_keys" {
  type        = list(string)
  description = "List of required tag keys."
  default     = ["created-by"]

  validation {
    condition     = length(var.required_tag_keys) <= 6
    error_message = "You can specify a maximum of 6 required tag keys."
  }
}

variable "aggregate_authorization_account_id" {
  type    = string
  default = ""
}

variable "aggregate_authorization_authorized_aws_region" {
  type    = string
  default = ""
}