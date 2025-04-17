variable "prefix" {
  type = string
}

variable "vpc_id" {
  type = string
  default = ""
}

variable "domains" {
  description = "Map of domain configurations"
  type = map(object({
    domain_name                = string
    parent_zone_id             = optional(string, "")
    create_zone                = optional(bool, true)
    create_certificate         = optional(bool, true)
    certificate_key_algorithm  = optional(string, "EC_secp384r1")
    certificate_authority_arn  = optional(string, "")
    private                    = optional(bool, false)
    vpc_id                     = optional(string, "")
    default_public             = optional(bool, null)
    default_private            = optional(bool, null)
  }))
  default = {}
}
