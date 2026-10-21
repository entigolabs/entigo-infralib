output "vpc_id" {
  value = azurerm_virtual_network.this.id
}

output "vpc_name" {
  value = azurerm_virtual_network.this.name
}

output "vpc_cidr" {
  value = var.vpc_cidr
}

output "vpc_cidr_block" {
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
  value = var.enable_nat_gateway ? azurerm_nat_gateway.this[0].id : ""
}

output "egress_route_table_id" {
  value = var.egress_next_hop_ip != null ? azurerm_route_table.egress[0].id : ""
}

output "nat_static_ips" {
  value = azurerm_public_ip.nat[*].ip_address
}

output "aks_subnets" {
  value = azurerm_subnet.aks[*].id
}

output "aks_subnet_cidrs" {
  value = local.aks_subnets
}

# Agent default subnet for vpc attached steps
output "container_apps_subnets" {
  value = azurerm_subnet.container_apps[*].id
}

output "container_apps_subnet_cidrs" {
  value = local.container_apps_subnets
}

output "container_apps_environment_ids" {
  value = azurerm_container_app_environment.container_apps[*].id
}

output "custom_subnets" {
  value = { for key, subnet in azurerm_subnet.custom : key => subnet.id }
}

output "custom_subnet_cidrs" {
  value = { for key, subnet in var.custom_subnets : key => subnet.cidr }
}

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

# Spoke mode (or explicit lists): the control and service subnets, otherwise the private subnets like aws/vpc default
output "control_subnets" {
  value = length(local.control_subnets) > 0 ? azurerm_subnet.control[*].id : azurerm_subnet.private[*].id
}

output "service_subnets" {
  value = length(local.service_subnets) > 0 ? azurerm_subnet.service[*].id : azurerm_subnet.private[*].id
}

output "compute_subnets" {
  value = azurerm_subnet.private[*].id
}

output "control_subnets_cidr_blocks" {
  value = length(local.control_subnets) > 0 ? local.control_subnets : local.private_subnets
}

output "service_subnets_cidr_blocks" {
  value = length(local.service_subnets) > 0 ? local.service_subnets : local.private_subnets
}

output "control_subnet_names" {
  value = length(local.control_subnets) > 0 ? azurerm_subnet.control[*].name : azurerm_subnet.private[*].name
}

output "service_subnet_names" {
  value = length(local.service_subnets) > 0 ? azurerm_subnet.service[*].name : azurerm_subnet.private[*].name
}

# Load balancer subnets outside the node subnet (aks lb_subnet_ids) and the platform internal gateway's subnet (istio-gateway), empty without them
output "lb_subnets" {
  value = concat(azurerm_subnet.control[*].id, azurerm_subnet.service[*].id)
}

output "control_subnet_name" {
  value = length(local.control_subnets) > 0 ? azurerm_subnet.control[0].name : ""
}

output "compute_subnets_cidr_blocks" {
  value = local.private_subnets
}

output "vpc_flow_log_id" {
  value = var.enable_flow_log ? azurerm_network_watcher_flow_log.this[0].id : ""
}

output "vpc_flow_log_destination_id" {
  value = var.enable_flow_log ? local.flow_log_storage_account_id : ""
}

output "vpc_flow_log_storage_account_name" {
  value = var.enable_flow_log ? coalesce(one(azurerm_storage_account.flow_log[*].name), element(split("/", local.flow_log_storage_account_id), -1)) : ""
}
