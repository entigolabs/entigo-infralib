output "key_vault_id" {
  value = azurerm_key_vault.this.id
}

output "key_vault_name" {
  value = azurerm_key_vault.this.name
}

output "key_vault_uri" {
  value = azurerm_key_vault.this.vault_uri
}

# Versionless ids follow key rotation
output "data_key_id" {
  value = azurerm_key_vault_key.this["data"].id
}

output "data_key_versionless_id" {
  value = azurerm_key_vault_key.this["data"].versionless_id
}

output "config_key_id" {
  value = azurerm_key_vault_key.this["config"].id
}

output "config_key_versionless_id" {
  value = azurerm_key_vault_key.this["config"].versionless_id
}

output "telemetry_key_id" {
  value = azurerm_key_vault_key.this["telemetry"].id
}

output "telemetry_key_versionless_id" {
  value = azurerm_key_vault_key.this["telemetry"].versionless_id
}
