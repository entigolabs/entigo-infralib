output "agc_id" {
  value = azurerm_application_load_balancer.this.id
}

output "agc_name" {
  value = azurerm_application_load_balancer.this.name
}

output "frontend_name" {
  value = azapi_resource.frontend.name
}

output "frontend_fqdn" {
  value = azapi_resource.frontend.output.fqdn
}

output "private_endpoint_ip" {
  value = azurerm_private_endpoint.this.private_service_connection[0].private_ip_address
}

# Source of the AGC traffic to the pods (NetworkPolicies)
output "subnet_cidr" {
  value = data.azurerm_subnet.this.address_prefixes[0]
}
