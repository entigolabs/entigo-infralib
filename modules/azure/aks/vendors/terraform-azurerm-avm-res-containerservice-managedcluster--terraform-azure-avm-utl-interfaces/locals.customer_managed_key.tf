locals {
  customer_managed_key_identity_client_id          = try(data.azapi_resource.customer_managed_key_identity[0].output.properties.clientId, null)
  customer_managed_key_identity_principal_id       = try(data.azapi_resource.customer_managed_key_identity[0].output.properties.principalId, null)
  customer_managed_key_identity_tenant_id          = try(data.azapi_resource.customer_managed_key_identity[0].output.properties.tenantId, null)
  customer_managed_key_key_name                    = var.customer_managed_key != null ? var.customer_managed_key.key_name : null
  customer_managed_key_key_resource_id             = var.customer_managed_key != null ? var.customer_managed_key.key_version != null ? "${var.customer_managed_key.key_vault_resource_id}/keys/${var.customer_managed_key.key_name}/versions/${var.customer_managed_key.key_version}" : null : null
  customer_managed_key_key_uri                     = var.customer_managed_key != null ? var.customer_managed_key.key_version != null ? "${local.customer_managed_key_versionless_key_uri}/${lookup(var.customer_managed_key, "key_version", "")}" : null : null
  customer_managed_key_key_vault_uri               = var.customer_managed_key != null ? "https://${basename(var.customer_managed_key.key_vault_resource_id)}.${var.customer_managed_key_key_vault_domain}" : null
  customer_managed_key_key_version                 = var.customer_managed_key != null ? lookup(var.customer_managed_key, "key_version", null) : null
  customer_managed_key_versionless_key_resource_id = var.customer_managed_key != null ? "${var.customer_managed_key.key_vault_resource_id}/keys/${var.customer_managed_key.key_name}" : null
  customer_managed_key_versionless_key_uri         = var.customer_managed_key != null ? "${local.customer_managed_key_key_vault_uri}/keys/${var.customer_managed_key.key_name}" : null
}
