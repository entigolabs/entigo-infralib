variable "prefix" {
  type = string
}

variable "vpc_cidr" {
  type = string
}

variable "one_nat_gateway_per_az" {
  type = bool
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
}

variable "elasticache_subnets" {
  type = list(string)
}

variable "intra_subnets" {
  type = list(string)
}

locals {
  hname = "${var.prefix}-${terraform.workspace}"
}
