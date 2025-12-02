output "prefix" {
  value = var.prefix
}

output "kms_data_key_id" {
  value = google_kms_crypto_key.kms_data_key.id
}

output "kms_config_key_id" {
  value = google_kms_crypto_key.kms_config_key.id
}

output "kms_telemetry_key_id" {
  value = google_kms_crypto_key.kms_telemetry_key.id
}
