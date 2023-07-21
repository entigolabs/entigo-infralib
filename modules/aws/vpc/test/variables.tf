variable "prefix" {
  type = string
}

variable "vpc_cidr" {
  type = string
}

variable "one_nat_gateway_per_az" {
  type = bool
  default = null
}

variable "private_subnets" {
  type = list(string)
  default = null
}

variable "public_subnets" {
  type = list(string)
  default = null
}

variable "database_subnets" {
  type = list(string)
  default = null
}

variable "elasticache_subnets" {
  type = list(string)
  default = null
}

variable "intra_subnets" {
  type = list(string)
  default = null
}
