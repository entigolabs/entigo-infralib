output "vpc_id" {
  value = azurerm_virtual_network.this.id
}

output "vpc_name" {
  value = azurerm_virtual_network.this.name
}

output "vpc_cidr" {
  value = var.vpc_cidr
}

# Waits for the NAT gateway (association and its public IPs) or the egress route, AKS needs egress at cluster creation
output "private_subnets" {
  value = azurerm_subnet.private[*].id
  depends_on = [
    azurerm_subnet_nat_gateway_association.private,
    azurerm_nat_gateway_public_ip_association.this,
    azurerm_subnet_route_table_association.egress_private,
  ]
}

output "public_subnets" {
  value = azurerm_subnet.public[*].id
}

output "intra_subnets" {
  value = azurerm_subnet.intra[*].id
}

output "database_subnets" {
  value = azurerm_subnet.database[*].id
}

output "agc_subnets" {
  value = azurerm_subnet.agc[*].id
}

output "private_subnet_cidrs" {
  value = local.private_subnets
}

output "public_subnet_cidrs" {
  value = local.public_subnets
}

output "intra_subnet_cidrs" {
  value = local.intra_subnets
}

output "database_subnet_cidrs" {
  value = local.database_subnets
}

output "agc_subnet_cidrs" {
  value = local.agc_subnets
}

output "nat_gateway_id" {
  value = var.enable_nat_gateway ? azurerm_nat_gateway.this[0].id : null
}

output "nat_static_ips" {
  value = azurerm_public_ip.nat[*].ip_address
}

output "apiserver_subnets" {
  value = azurerm_subnet.apiserver[*].id
}

output "apiserver_subnet_cidrs" {
  value = local.apiserver_subnets
}

# Agent default subnet for vpc attached steps
output "pipeline_subnets" {
  value = [for n in local.pipeline_names : azurerm_subnet.pipeline[n].id]
}

output "pipeline_subnet_cidrs" {
  value = local.pipeline_subnets
}

output "pipeline_environment_ids" {
  value = [for n in local.pipeline_names : azurerm_container_app_environment.pipeline[n].id]
}

output "mssql_subnets" {
  value = azurerm_subnet.mssql[*].id
}

output "mssql_subnet_cidrs" {
  value = local.mssql_subnets
}

# Same names as aws/vpc for shared consumers (e.g. platform-apis). AKS nodes use the private subnets,
# Redis/Postgres/MySQL private endpoints the database subnets.
output "name" {
  value = azurerm_virtual_network.this.name
}

output "nat_public_ips" {
  value = azurerm_public_ip.nat[*].ip_address
}

output "private_subnet_names" {
  value = azurerm_subnet.private[*].name
}

output "public_subnet_names" {
  value = azurerm_subnet.public[*].name
}

output "intra_subnet_names" {
  value = azurerm_subnet.intra[*].name
}

output "database_subnet_names" {
  value = azurerm_subnet.database[*].name
}

output "private_subnets_cidr_blocks" {
  value = local.private_subnets
}

output "public_subnets_cidr_blocks" {
  value = local.public_subnets
}

output "intra_subnets_cidr_blocks" {
  value = local.intra_subnets
}

output "database_subnets_cidr_blocks" {
  value = local.database_subnets
}

output "elasticache_subnets_cidr_blocks" {
  value = local.database_subnets
}

output "control_subnets" {
  value = azurerm_subnet.private[*].id
}

output "service_subnets" {
  value = azurerm_subnet.private[*].id
}

output "compute_subnets" {
  value = azurerm_subnet.private[*].id
}

output "control_subnets_cidr_blocks" {
  value = local.private_subnets
}

output "service_subnets_cidr_blocks" {
  value = local.private_subnets
}

output "compute_subnets_cidr_blocks" {
  value = local.private_subnets
}

output "vpc_flow_log_id" {
  value = var.enable_flow_log ? azurerm_network_watcher_flow_log.this[0].id : null
}

output "vpc_flow_log_destination_id" {
  value = var.enable_flow_log ? azurerm_storage_account.flow_log[0].id : null
}
