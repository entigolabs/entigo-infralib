variable "prefix" {
  type = string
}

variable "vpc_cidr" {
  type = string
  nullable = false
  default = "10.156.0.0/16"
}

variable "azs" {
  type = number
  nullable = false
  default = 2
}

variable "one_nat_gateway_per_az" {
  type = bool
  nullable = false
  default = true
}

variable "private_subnets" {
  type = list(string)
  nullable = true
  default = null
}

variable "public_subnets" {
  type = list(string)
  nullable = true
  default = null
}

variable "database_subnets" {
  type = list(string)
  nullable = true
  default = null
}

variable "elasticache_subnets" {
  type = list(string)
  nullable = true
  default = null
}

variable "intra_subnets" {
  type = list(string)
  nullable = true
  default = null
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
