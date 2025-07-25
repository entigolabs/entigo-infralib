variable "prefix" {
  type = string
}

variable "vpc_cidr" {
  description = "Must be at least size /24 for automatic subnet calculation. Recommended at least /21 or larger."
  type     = string
  nullable = false
  default  = "10.156.0.0/16"
}

variable "subnet_split_mode" {
  description = "Define the way we split the VPC cidr into subnets if they are not specified. Possible: default, spoke"
  type     = string
  nullable = false
  default  = "default"
}

variable "secondary_cidr_blocks" {
  type     = list(string)
  nullable = false
  default  = []
}

variable "azs" {
  type     = number
  nullable = false
  default  = 2
}

variable "one_nat_gateway_per_az" {
  type     = bool
  nullable = false
  default  = true
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

variable "elasticache_subnets" {
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

variable "elasticache_subnet_names" {
  type     = list(string)
  nullable = true
  default  = []
}

variable "intra_subnet_names" {
  type     = list(string)
  nullable = true
  default  = []
}

variable "create_multiple_intra_route_tables" {
  type     = bool
  nullable = false
  default  = false
}

variable "create_multiple_public_route_tables" {
  type     = bool
  nullable = false
  default  = false
}

variable "enable_nat_gateway" {
  type     = bool
  nullable = false
  default  = true
}

variable "manage_default_network_acl" {
  type     = bool
  nullable = false
  default  = true
}


variable "map_public_ip_on_launch" {
  type     = bool
  nullable = false
  default  = false
}

variable "enable_flow_log" {
  type     = bool
  nullable = false
  default  = true
}

variable "flow_log_max_aggregation_interval" {
  type = number
  default = 60
}

variable "flow_log_cloudwatch_log_group_retention_in_days" {
  type = number
  default = 7
}

variable "flow_log_cloudwatch_log_group_kms_key_id" {
  type = string
  default = ""
}

variable "create_gateway_s3" {
  type     = bool
  nullable = false
  default  = true
}

variable "create_endpoint_s3" {
  type     = bool
  nullable = false
  default  = false
}

variable "create_endpoint_ecr" {
  type     = bool
  nullable = false
  default  = false
}

variable "create_endpoint_ec2" {
  type     = bool
  nullable = false
  default  = false
}

variable "create_endpoint_sts" {
  type     = bool
  nullable = false
  default  = false
}

variable "endpoints_sg_extra_rules" {
  type     = list(string)
  nullable = false
  default  = []
}
