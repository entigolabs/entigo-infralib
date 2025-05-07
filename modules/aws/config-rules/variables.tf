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

variable "operational_best_practices_without_s3_conformance_pack_enabled" {
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

variable "resource_tagging_compliance_pack_enabled" {
 description = "Enable or disable the resource tagging compliance pack"
 type        = bool
 default     = false
}

variable "required_tag_keys" {
 description = "Comma-separated list of required tag keys (up to 9). If empty, just checks for any tag."
 type        = string
 default     = ""
}