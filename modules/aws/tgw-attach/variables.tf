variable "prefix" {
  type = string
}

variable "vpc_id" {
  type = string
  default = ""
}

variable "attachment_subnets" {
  type = list(string)
  default = []
}

variable "routes" {
  type = map(any)
  default = {}
}

variable "dns_support" {
  type = string
  default = "enable"
}

variable "ipv6_support" {
  type = string
  default = "disable"
}

variable "security_group_referencing_support" {
  type = string
  default = "enable"
}

variable "transit_gateway_default_route_table_association" {
  type = bool
  default = true
}

variable "transit_gateway_default_route_table_propagation" {
  type = bool
  default = true
}
