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
  validation {
    condition     = can(cidrnetmask(var.vpc_cidr)) && try(tonumber(split("/", var.vpc_cidr)[1]) <= 24, false)
    error_message = "vpc_cidr must be an IPv4 CIDR of /24 or larger."
  }
}

variable "subnet_split_mode" {
  type        = string
  nullable    = false
  default     = "default"
  description = "default: one private subnet for nodes and load balancers. spoke: separate control (AGC of the internal gateway) and service (AGC of the service gateway) /24s delegated to Application Gateway for Containers, the node subnet named compute, no intra subnet, always a database subnet; outputs control_subnets and service_subnets point to the new subnets"
  validation {
    condition     = contains(["default", "spoke"], var.subnet_split_mode)
    error_message = "subnet_split_mode must be default or spoke."
  }
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

variable "control_subnets" {
  type    = list(string)
  default = null
}

variable "service_subnets" {
  type    = list(string)
  default = null
}

variable "control_subnet_names" {
  type    = list(string)
  default = []
}

variable "service_subnet_names" {
  type    = list(string)
  default = []
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

variable "aks_subnet_names" {
  type    = list(string)
  default = []
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

variable "flow_log_storage_account_id" {
  type        = string
  default     = null
  description = "Existing storage account for the flow logs (same region, e.g. a central log archive) instead of the module's own"
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
  description = "Private IP of a firewall/NVA (e.g. in a peered hub): 0.0.0.0/0 of the private and Container Apps subnets goes through it. Needs enable_nat_gateway = false and aks outbound_type = userDefinedRouting"
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
variable "aks_subnets" {
  type        = list(string)
  default     = null
  description = "AKS API Server VNet Integration subnet, min /28"
}

# https://learn.microsoft.com/azure/container-apps/custom-virtual-networks?tabs=workload-profiles-env#subnet
variable "container_apps_subnets" {
  type    = list(string)
  default = null
}

variable "container_apps_subnet_names" {
  type    = list(string)
  default = []
}

variable "container_apps_subnet_service_endpoints" {
  type    = list(string)
  default = ["Microsoft.Storage", "Microsoft.KeyVault"]
}

variable "container_apps_zone_redundancy_enabled" {
  type    = bool
  default = true
}

# https://learn.microsoft.com/azure/azure-sql/managed-instance/connectivity-architecture-overview#virtual-network-requirements
variable "custom_subnets" {
  type = map(object({
    cidr                   = string
    name                   = optional(string)
    delegation             = optional(string)
    delegation_actions     = optional(list(string), ["Microsoft.Network/virtualNetworks/subnets/join/action"])
    network_security_group = optional(bool, false)
    route_table            = optional(bool, false)
    nat_gateway            = optional(bool, false)
  }))
  nullable    = false
  default     = {}
  description = "Extra subnets in the spare last quarter of vpc_cidr, e.g. SQL Managed Instance (delegation, NSG and route table) or AzureBastionSubnet. Name defaults to <prefix>-<key>, nat_gateway gives the subnet the egress of the private subnets"
  validation {
    condition     = alltrue([for key, subnet in var.custom_subnets : can(regex("^[a-z0-9]{1,40}$", key))])
    error_message = "custom_subnets keys must be 1-40 lowercase letters or digits, the agent reads the outputs as custom_subnets[key]."
  }
}
