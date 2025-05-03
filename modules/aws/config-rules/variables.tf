variable "prefix" {
  type = string
}

# https://docs.aws.amazon.com/config/latest/APIReference/API_ResourceIdentifier.html#config-Type-ResourceIdentifier-resourceType
variable "resource_types_to_exclude" {
  type    = list(string)
  default = []
}

variable "logs_bucket" {
  type    = string
  default = ""
}

variable "operational_best_practices_without_s3_conformance_pack_enabled" {
  type    = bool
  default = false
}