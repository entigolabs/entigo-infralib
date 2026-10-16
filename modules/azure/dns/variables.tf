variable "prefix" {
  type = string
}

variable "resource_group_name" {
  type    = string
  default = ""
}

variable "vpc_ids" {
  type    = list(string)
  default = []
}

variable "domains" {
  description = "Map of domain configurations"
  type = map(object({
    domain_name       = string
    parent_zone_id    = optional(string, "")
    create_zone       = optional(bool, true)
    create_validation = optional(bool, true)
    private           = optional(bool, false)
    vpc_ids           = optional(list(string), [])
    default_public    = optional(bool, null)
    default_private   = optional(bool, null)
  }))
  default = {}
}

variable "tags" {
  type    = map(string)
  default = {}
}
