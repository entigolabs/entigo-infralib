output "client_id" {
  value = azurerm_user_assigned_identity.crossplane.client_id
}

output "principal_id" {
  value = azurerm_user_assigned_identity.crossplane.principal_id
}

output "core_client_id" {
  value = azurerm_user_assigned_identity.core.client_id
}

output "core_principal_id" {
  value = azurerm_user_assigned_identity.core.principal_id
}

output "tenant_id" {
  value = data.azurerm_client_config.this.tenant_id
}

output "subscription_id" {
  value = data.azurerm_client_config.this.subscription_id
}

output "resource_group_name" {
  value = var.resource_group_name
}

output "kubernetes_namespace" {
  value = var.kubernetes_namespace
}

output "kubernetes_service_account" {
  value = var.kubernetes_service_account
}

output "kubernetes_core_service_account" {
  value = var.kubernetes_core_service_account
}

output "agent_storage_account_name" {
  value = try(data.azurerm_resources.agent_storage.resources[0].name, "")
}

output "agent_storage_account_id" {
  value = try(data.azurerm_resources.agent_storage.resources[0].id, "")
}
