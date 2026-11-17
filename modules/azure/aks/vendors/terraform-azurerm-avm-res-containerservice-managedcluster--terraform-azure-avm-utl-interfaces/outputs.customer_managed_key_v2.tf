output "customer_managed_key_azapi_v2" {
  description = <<DESCRIPTION
An object containing the following attributes:

- `identity_client_id` - The client ID of the user-assigned identity. Will be null if no user-assigned identity client ID is provided.
- `identity_principal_id` - Always null because Variant 2 does not carry the identity resource ID required to resolve it.
- `identity_tenant_id` - Always null because Variant 2 does not carry the identity resource ID required to resolve it.
- `key_name` - The name of the key. Will be null if no key is provided.
- `key_resource_id` - Always null because Variant 2 does not carry the vault resource ID required to construct it.
- `key_uri` - The URI of the key, including the version. If the key version is not provided, this will be null.
- `key_vault_uri` - The URI of the Key Vault or Managed HSM.
- `key_version` - The version of the key. Will be null if the supplied URI is versionless.
- `versionless_key_resource_id` - Always null because Variant 2 does not carry the vault resource ID required to construct it.
- `versionless_key_uri` - The URI of the key, without the version.
DESCRIPTION
  value = {
    identity_client_id          = local.customer_managed_key_v2_identity_client_id
    identity_principal_id       = null
    identity_tenant_id          = null
    key_name                    = local.customer_managed_key_v2_key_name
    key_resource_id             = null
    key_uri                     = local.customer_managed_key_v2_key_uri
    key_vault_uri               = local.customer_managed_key_v2_key_vault_uri
    key_version                 = local.customer_managed_key_v2_key_version
    versionless_key_resource_id = null
    versionless_key_uri         = local.customer_managed_key_v2_versionless_key_uri
  }
}
