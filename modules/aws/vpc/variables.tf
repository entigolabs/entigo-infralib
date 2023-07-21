variable "prefix" {
  type = string
}

variable "vpc_cidr" {
  type = string
}

variable "one_nat_gateway_per_az" {
  type = bool
  nullable = false
  default = false
}

variable "private_subnets" {
  type = list(string)
}

variable "public_subnets" {
  type = list(string)
}

variable "database_subnets" {
  type = list(string)
  nullable = false
  default = []
}

variable "elasticache_subnets" {
  type = list(string)
  nullable = false
  default = []
}

variable "intra_subnets" {
  type = list(string)
  nullable = false
  default = []
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
