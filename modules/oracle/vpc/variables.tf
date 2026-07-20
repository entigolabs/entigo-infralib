variable "prefix" {
  type = string
}

variable "compartment_id" {
  description = "OCID of the compartment that will contain the VCN and its resources."
  type        = string
}

variable "vpc_cidr" {
  type     = string
  nullable = false
  default  = "10.156.0.0/16"
}

# Optional VCN/subnet DNS label. Must be alphanumeric and start with a letter;
# left unset by default because hyphenated prefixes are not valid labels.
variable "dns_label" {
  type     = string
  nullable = true
  default  = null
}

variable "private_subnets" {
  type     = list(string)
  nullable = true
  default  = null
}

variable "public_subnets" {
  type     = list(string)
  nullable = true
  default  = null
}

variable "database_subnets" {
  type     = list(string)
  nullable = true
  default  = null
}

variable "intra_subnets" {
  type     = list(string)
  nullable = true
  default  = null
}

variable "private_subnet_names" {
  type    = list(string)
  default = []
}

variable "public_subnet_names" {
  type    = list(string)
  default = []
}

variable "database_subnet_names" {
  type     = list(string)
  nullable = true
  default  = []
}

variable "intra_subnet_names" {
  type     = list(string)
  nullable = true
  default  = []
}

variable "enable_internet_gateway" {
  type     = bool
  nullable = false
  default  = true
}

variable "enable_nat_gateway" {
  type     = bool
  nullable = false
  default  = true
}

variable "enable_service_gateway" {
  type     = bool
  nullable = false
  default  = true
}
