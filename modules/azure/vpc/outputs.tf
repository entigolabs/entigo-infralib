output "vpc_id" {
  value = azurerm_virtual_network.this.id
}

output "vpc_name" {
  value = azurerm_virtual_network.this.name
}

output "vpc_cidr" {
  value = var.vpc_cidr
}

# Waits for the NAT association, AKS userAssignedNATGateway needs it at cluster creation
output "private_subnets" {
  value      = azurerm_subnet.private[*].id
  depends_on = [azurerm_subnet_nat_gateway_association.private]
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
  value = azurerm_subnet.pipeline[*].id
}

output "pipeline_subnet_cidrs" {
  value = local.pipeline_subnets
}
