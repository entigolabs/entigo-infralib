variable "prefix" {
  type = string
}

variable "vpc_ids" {
  type    = list(string)
  default = []
}

variable "create_public" {
  type    = bool
  default = true
}

variable "create_private" {
  type    = bool
  default = true
}

variable "parent_zone_id" {
  type    = string
  default = ""
}

variable "parent_domain" {
  type    = string
  default = ""
}

variable "public_subdomain_name" {
  type    = string
  default = ""
}

variable "private_subdomain_name" {
  type    = string
  default = ""
}

variable "dns_policy_vpc_ids" {
  type    = list(string)
  default = []
}

variable "enable_inbound_forwarding" {
  type    = bool
  default = false
}

variable "enable_logging" {
  type    = bool
  default = false
}

variable "alternative_name_servers" {
  type = list(object({
    ipv4_address    = string
    forwarding_path = optional(string)
  }))
  default = null
}