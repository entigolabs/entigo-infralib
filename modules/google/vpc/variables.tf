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

variable "enable_nat_gateway" {
  type     = bool
  nullable = false
  default  = true
}

variable "nat_static_ip_count" {
  type     = number
  nullable = false
  default  = 0
}

variable "nat_enable_endpoint_independent_mapping" {
  type     = bool
  nullable = false
  default  = false
}

variable "nat_enable_dynamic_port_allocation" {
  type     = bool
  nullable = false
  default  = false
}

variable "nat_min_ports_per_vm" {
  type     = number
  nullable = true
  default  = null
}

variable "nat_max_ports_per_vm" {
  type     = number
  nullable = true
  default  = null
}

variable "nat_log_config_enable" {
  type     = bool
  nullable = false
  default  = false
}

variable "nat_log_config_filter" {
  type     = string
  nullable = false
  default  = "ALL"
}

variable "nat_icmp_idle_timeout_sec" {
  type     = number
  nullable = false
  default  = 30
}

variable "nat_tcp_established_idle_timeout_sec" {
  type     = number
  nullable = false
  default  = 1200
}

variable "nat_tcp_transitory_idle_timeout_sec" {
  type     = number
  nullable = false
  default  = 30
}

variable "nat_tcp_time_wait_timeout_sec" {
  type     = number
  nullable = false
  default  = 120
}

variable "nat_udp_idle_timeout_sec" {
  type     = number
  nullable = false
  default  = 30
}
