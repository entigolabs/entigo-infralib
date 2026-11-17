data "azapi_resource" "customer_managed_key_identity" {
  count = var.customer_managed_key == null ? 0 : var.customer_managed_key.user_assigned_identity == null ? 0 : 1

  resource_id = var.customer_managed_key.user_assigned_identity.resource_id
  type        = "Microsoft.ManagedIdentity/userAssignedIdentities@2023-01-31"
  response_export_values = [
    "properties.clientId",
    "properties.principalId",
    "properties.tenantId",
  ]
}
