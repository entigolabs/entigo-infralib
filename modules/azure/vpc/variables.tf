variable "prefix" {
  type = string
}

variable "resource_group_name" {
  type    = string
  default = ""
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

variable "nat_static_ip_count" {
  type        = number
  default     = 1
  description = "About 64k SNAT ports per public IP, max 16"
  validation {
    condition     = var.nat_static_ip_count >= 1 && var.nat_static_ip_count <= 16
    error_message = "nat_static_ip_count must be between 1 and 16."
  }
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
  type        = list(string)
  default     = null
  description = "Container Apps environment subnet"
}

variable "pipeline_subnet_names" {
  type    = list(string)
  default = []
}

variable "enable_mssql_subnets" {
  type        = bool
  default     = false
  description = "Create the Azure SQL Managed Instance subnets"
}

variable "mssql_subnets" {
  type        = list(string)
  default     = null
  description = "Azure SQL Managed Instance subnets"
}

variable "mssql_subnet_names" {
  type    = list(string)
  default = []
}
