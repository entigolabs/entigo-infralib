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

# variable "resource_tagging_rule_enabled" {
#  description = "Enable or disable the resource tagging rule"
#  type        = bool
#  default     = false
# }

# variable "required_tag_keys" {
#   type        = list(string)
#   description = "List of required tag keys, max 9 tags."
#   default     = ["Owner", "Environment", "Project"]
# }