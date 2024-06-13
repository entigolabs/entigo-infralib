variable "prefix" {
  type = string
}

variable "vpc_cidr" {
  type     = string
  nullable = false
  default  = "10.156.0.0/16"
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
  type     = list(string)
  default  = []
}

variable "public_subnet_names" {
  type     = list(string)
  default  = []
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

variable "enable_nat_gateway" {
  type     = bool
  nullable = false
  default  = true
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}

