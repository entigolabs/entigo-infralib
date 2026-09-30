output "acr_id" {
  value = azurerm_container_registry.this.id
}

output "acr_login_server" {
  value = azurerm_container_registry.this.login_server
}

output "hub_registry" {
  value = contains(keys(local.cache_rules), "hub") ? "${azurerm_container_registry.this.login_server}/docker.io" : null
}

output "ghcr_registry" {
  value = "${azurerm_container_registry.this.login_server}/ghcr.io"
}

output "gcr_registry" {
  value = "${azurerm_container_registry.this.login_server}/gcr.io"
}

output "ecr_registry" {
  value = "${azurerm_container_registry.this.login_server}/public.ecr.aws"
}

output "quay_registry" {
  value = "${azurerm_container_registry.this.login_server}/quay.io"
}

output "mcr_registry" {
  value = "${azurerm_container_registry.this.login_server}/mcr.microsoft.com"
}

output "k8s_registry" {
  value = "${azurerm_container_registry.this.login_server}/registry.k8s.io"
}

# AKS bootstrap from the cache must wait for the private endpoint
output "private_endpoint_id" {
  value = local.private_endpoint ? azurerm_private_endpoint.acr[0].id : null
}

output "aks_bootstrap_cache_rule_id" {
  value = var.aks_bootstrap_cache_rule ? azurerm_container_registry_cache_rule.aks_bootstrap[0].id : null
}
