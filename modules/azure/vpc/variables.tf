variable "prefix" {
  type = string
}

variable "resource_group_name" {
  type    = string
  default = ""
}

variable "unique_suffix" {
  type        = string
  description = "Agent uniqueSuffix (8 chars) for the globally unique flow log storage account name"
  validation {
    condition     = can(regex("^[a-z0-9]{8}$", var.unique_suffix))
    error_message = "unique_suffix must be 8 lowercase letters or digits"
  }
}

variable "location" {
  type    = string
  default = ""
}

variable "vpc_cidr" {
  type     = string
  nullable = false
  default  = "10.156.0.0/16"
}

variable "private_subnets" {
  type    = list(string)
  default = null
}

variable "public_subnets" {
  type    = list(string)
  default = null
}

variable "intra_subnets" {
  type    = list(string)
  default = null
}

variable "database_subnets" {
  type    = list(string)
  default = null
}

# https://learn.microsoft.com/azure/application-gateway/for-containers/container-networking#limitations
variable "agc_subnets" {
  type        = list(string)
  default     = null
  description = "Application Gateway for Containers subnet, exactly a /24 with CNI Overlay"
}

variable "private_subnet_names" {
  type    = list(string)
  default = []
}

variable "public_subnet_names" {
  type    = list(string)
  default = []
}

variable "intra_subnet_names" {
  type    = list(string)
  default = []
}

variable "database_subnet_names" {
  type    = list(string)
  default = []
}

variable "agc_subnet_names" {
  type    = list(string)
  default = []
}

variable "apiserver_subnet_names" {
  type    = list(string)
  default = []
}

variable "database_subnet_delegation" {
  type        = string
  default     = null
  description = "Service the database subnets are delegated to, e.g. Microsoft.DBforPostgreSQL/flexibleServers"
}

variable "private_subnet_service_endpoints" {
  type    = list(string)
  default = ["Microsoft.Storage", "Microsoft.KeyVault"]
}

variable "enable_nat_gateway" {
  type    = bool
  default = true
}

# https://learn.microsoft.com/azure/network-watcher/vnet-flow-logs-overview
variable "enable_flow_log" {
  type     = bool
  nullable = false
  default  = true
}

variable "flow_log_retention_days" {
  type    = number
  default = 7
}

variable "flow_log_traffic_analytics_enabled" {
  type    = bool
  default = false
}

# Azure creates one Network Watcher per region and subscription, the flow log resource has to live in its resource group
variable "network_watcher_name" {
  type        = string
  default     = ""
  description = "\"\" = NetworkWatcher_<location>"
}

variable "network_watcher_resource_group_name" {
  type    = string
  default = "NetworkWatcherRG"
}

variable "nat_gateway_sku" {
  type        = string
  default     = "StandardV2"
  description = "StandardV2 is zone redundant, Standard is single zone. Changing it replaces the NAT gateway and its public IPs (new egress IPs)"
}

variable "nat_idle_timeout_minutes" {
  type    = number
  default = 4
}

# https://learn.microsoft.com/azure/aks/egress-outboundtype#outbound-type-of-userdefinedrouting
variable "egress_next_hop_ip" {
  type        = string
  default     = null
  description = "Private IP of a firewall/NVA (e.g. in a peered hub): 0.0.0.0/0 of the private and pipeline subnets goes through it. Needs enable_nat_gateway = false and aks outbound_type = userDefinedRouting"
}

variable "nat_static_ip_count" {
  type        = number
  default     = 1
  description = "About 64k SNAT ports per public IP, max 16"
}

variable "tags" {
  type    = map(string)
  default = {}
}

# https://learn.microsoft.com/azure/aks/api-server-vnet-integration
variable "apiserver_subnets" {
  type        = list(string)
  default     = null
  description = "AKS API Server VNet Integration subnet, min /28"
}

# https://learn.microsoft.com/azure/container-apps/custom-virtual-networks?tabs=workload-profiles-env#subnet
variable "pipeline_subnets" {
  type    = list(string)
  default = null
}

variable "pipeline_subnet_names" {
  type    = list(string)
  default = []
}

variable "pipeline_subnet_service_endpoints" {
  type    = list(string)
  default = ["Microsoft.Storage", "Microsoft.KeyVault"]
}

variable "pipeline_zone_redundancy_enabled" {
  type    = bool
  default = true
}

variable "enable_mssql_subnets" {
  type    = bool
  default = false
}

variable "mssql_subnets" {
  type    = list(string)
  default = null
}

variable "mssql_subnet_names" {
  type    = list(string)
  default = []
}
