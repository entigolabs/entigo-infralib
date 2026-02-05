output "prefix" {
  value = var.prefix
}

output "location" {
  value = local.location
}

output "key_ring_name" {
  value = local.key_ring_name
}

output "key_ring_id" {
  value = var.create_key_ring ? google_kms_key_ring.this[0].id : data.google_kms_key_ring.this[0].id
}

output "data_key_id" {
  value = google_kms_crypto_key.data.id
}

output "config_key_id" {
  value = google_kms_crypto_key.config.id
}

output "telemetry_key_id" {
  value = google_kms_crypto_key.telemetry.id
}
