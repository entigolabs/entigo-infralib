output "key_vault_id" {
  value = azurerm_key_vault.this.id
}

output "key_vault_name" {
  value = azurerm_key_vault.this.name
}

output "key_vault_uri" {
  value = azurerm_key_vault.this.vault_uri
}

# Versionless, follows key rotation (like the AWS alias)
output "data_key_id" {
  value = azurerm_key_vault_key.this["data"].versionless_id
}

output "data_key_version_id" {
  value = azurerm_key_vault_key.this["data"].id
}

output "data_key_resource_id" {
  value = azurerm_key_vault_key.this["data"].resource_versionless_id
}

# Versionless, follows key rotation (like the AWS alias)
output "config_key_id" {
  value = azurerm_key_vault_key.this["config"].versionless_id
}

output "config_key_version_id" {
  value = azurerm_key_vault_key.this["config"].id
}

output "config_key_resource_id" {
  value = azurerm_key_vault_key.this["config"].resource_versionless_id
}

# Versionless, follows key rotation (like the AWS alias)
output "telemetry_key_id" {
  value = azurerm_key_vault_key.this["telemetry"].versionless_id
}

output "telemetry_key_version_id" {
  value = azurerm_key_vault_key.this["telemetry"].id
}

output "telemetry_key_resource_id" {
  value = azurerm_key_vault_key.this["telemetry"].resource_versionless_id
}
