output "customer_managed_key_azapi" {
  description = <<DESCRIPTION
An object containing the following attributes:

- `identity_client_id` - The client ID of the user-assigned identity. Will be null if no user-assigned identity is provided.
- `identity_principal_id` - The principal ID of the user-assigned identity. Will be null if no user-assigned identity is provided.
- `identity_tenant_id` - The tenant ID of the user-assigned identity. Will be null if no user-assigned identity is provided.
- `key_name` - The name of the key. Will be null if no key is provided.
- `key_resource_id` - The resource ID of the key, including the version. If the key version is not provided, this will be null.
- `key_uri` - The URI of the key, including the version. If the key version is not provided, this will be null.
- `key_vault_uri` - The URI of the key vault.
- `key_version` - The version of the key. Will be null if no key is provided.
- `versionless_key_resource_id` - The resource ID of the key, without the version.
- `versionless_key_uri` - The URI of the key, without the version.
DESCRIPTION
  value = {
    identity_client_id          = local.customer_managed_key_identity_client_id
    identity_principal_id       = local.customer_managed_key_identity_principal_id
    identity_tenant_id          = local.customer_managed_key_identity_tenant_id
    key_name                    = local.customer_managed_key_key_name
    key_resource_id             = local.customer_managed_key_key_resource_id
    key_uri                     = local.customer_managed_key_key_uri
    key_vault_uri               = local.customer_managed_key_key_vault_uri
    key_version                 = local.customer_managed_key_key_version
    versionless_key_resource_id = local.customer_managed_key_versionless_key_resource_id
    versionless_key_uri         = local.customer_managed_key_versionless_key_uri
  }
}
