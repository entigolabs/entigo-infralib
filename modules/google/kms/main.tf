locals {
  labels = merge(var.labels, { created-by = "entigo-infralib" })

  location = var.location != "" ? var.location : data.google_client_config.this.region
  key_ring = var.create_key_ring ? google_kms_key_ring.this[0].id : data.google_kms_key_ring.this[0].id

  key_ring_name      = var.key_ring_name != "" ? var.key_ring_name : "${var.prefix}-${random_string.suffix.result}"
  data_key_name      = "${var.prefix}-data-${random_string.suffix.result}"
  config_key_name    = "${var.prefix}-config-${random_string.suffix.result}"
  telemetry_key_name = "${var.prefix}-telemetry-${random_string.suffix.result}"

  # Services for each key type
  data_key_encrypter_decrypter_services = [
    "compute.googleapis.com",
    "container.googleapis.com",
    "sqladmin.googleapis.com",
    "storage.googleapis.com",
    "redis.googleapis.com",
    "memorystore.googleapis.com",
    "file.googleapis.com",
    "artifactregistry.googleapis.com",
    "pubsub.googleapis.com",
  ]
  config_key_encrypter_decrypter_services = [
    "secretmanager.googleapis.com",
    "container.googleapis.com",
  ]
  telemetry_key_encrypter_decrypter_services = [
    "storage.googleapis.com",
  ]

  # KMS Data key
  data_key_encrypters = { for v in var.data_key_additional_encrypters : v => v }
  data_key_decrypters = { for v in var.data_key_additional_decrypters : v => v }
  data_key_encrypters_decrypters = merge(
    { for v in var.data_key_additional_encrypters_decrypters : v => v },
    {
      for service in local.data_key_encrypter_decrypter_services :
      service => "serviceAccount:${var.service_agent_emails[service]}"
      if contains(keys(var.service_agent_emails), service)
    }
  )

  # KMS Config key
  config_key_encrypters = { for v in var.config_key_additional_encrypters : v => v }
  config_key_decrypters = { for v in var.config_key_additional_decrypters : v => v }
  config_key_encrypters_decrypters = merge(
    { for v in var.config_key_additional_encrypters_decrypters : v => v },
    {
      for service in local.config_key_encrypter_decrypter_services :
      service => "serviceAccount:${var.service_agent_emails[service]}"
      if contains(keys(var.service_agent_emails), service)
    }
  )

  # KMS Telemetry key
  telemetry_key_encrypters = { for v in var.telemetry_key_additional_encrypters : v => v }
  telemetry_key_decrypters = { for v in var.telemetry_key_additional_decrypters : v => v }
  telemetry_key_encrypters_decrypters = merge(
    { for v in var.telemetry_key_additional_encrypters_decrypters : v => v },
    {
      for service in local.telemetry_key_encrypter_decrypter_services :
      service => "serviceAccount:${var.service_agent_emails[service]}"
      if contains(keys(var.service_agent_emails), service)
    }
  )
}

# Generate random suffix for resource names
resource "random_string" "suffix" {
  length  = 8
  lower   = true
  upper   = false
  numeric = true
  special = false
}

# Single key ring for all KMS keys
resource "google_kms_key_ring" "this" {
  count    = var.create_key_ring ? 1 : 0
  name     = local.key_ring_name
  project  = data.google_client_config.this.project
  location = local.location
}

# Data key
resource "google_kms_crypto_key" "data" {
  name                          = local.data_key_name
  key_ring                      = local.key_ring
  rotation_period               = var.key_rotation_period
  purpose                       = var.key_purpose
  import_only                   = false
  skip_initial_version_creation = false

  lifecycle {
    prevent_destroy = true
  }

  destroy_scheduled_duration = var.key_destroy_scheduled_duration

  version_template {
    algorithm        = var.key_algorithm
    protection_level = var.key_protection_level
  }

  labels = local.labels
}

resource "google_kms_crypto_key_iam_member" "data_encrypters" {
  for_each      = local.data_key_encrypters
  role          = "roles/cloudkms.cryptoKeyEncrypter"
  crypto_key_id = google_kms_crypto_key.data.id
  member        = each.value
}

resource "google_kms_crypto_key_iam_member" "data_decrypters" {
  for_each      = local.data_key_decrypters
  role          = "roles/cloudkms.cryptoKeyDecrypter"
  crypto_key_id = google_kms_crypto_key.data.id
  member        = each.value
}

resource "google_kms_crypto_key_iam_member" "data_encrypters_decrypters" {
  for_each      = local.data_key_encrypters_decrypters
  crypto_key_id = google_kms_crypto_key.data.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = each.value
}

# Config key
resource "google_kms_crypto_key" "config" {
  name                          = local.config_key_name
  key_ring                      = local.key_ring
  rotation_period               = var.key_rotation_period
  purpose                       = var.key_purpose
  import_only                   = false
  skip_initial_version_creation = false

  lifecycle {
    prevent_destroy = true
  }

  destroy_scheduled_duration = var.key_destroy_scheduled_duration

  version_template {
    algorithm        = var.key_algorithm
    protection_level = var.key_protection_level
  }

  labels = local.labels
}

resource "google_kms_crypto_key_iam_member" "config_encrypters" {
  for_each      = local.config_key_encrypters
  role          = "roles/cloudkms.cryptoKeyEncrypter"
  crypto_key_id = google_kms_crypto_key.config.id
  member        = each.value
}

resource "google_kms_crypto_key_iam_member" "config_decrypters" {
  for_each      = local.config_key_decrypters
  role          = "roles/cloudkms.cryptoKeyDecrypter"
  crypto_key_id = google_kms_crypto_key.config.id
  member        = each.value
}

resource "google_kms_crypto_key_iam_member" "config_encrypters_decrypters" {
  for_each      = local.config_key_encrypters_decrypters
  crypto_key_id = google_kms_crypto_key.config.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = each.value
}

# Telemetry key
resource "google_kms_crypto_key" "telemetry" {
  name                          = local.telemetry_key_name
  key_ring                      = local.key_ring
  rotation_period               = var.key_rotation_period
  purpose                       = var.key_purpose
  import_only                   = false
  skip_initial_version_creation = false

  lifecycle {
    prevent_destroy = true
  }

  destroy_scheduled_duration = var.key_destroy_scheduled_duration

  version_template {
    algorithm        = var.key_algorithm
    protection_level = var.key_protection_level
  }

  labels = local.labels
}

resource "google_kms_crypto_key_iam_member" "telemetry_encrypters" {
  for_each      = local.telemetry_key_encrypters
  role          = "roles/cloudkms.cryptoKeyEncrypter"
  crypto_key_id = google_kms_crypto_key.telemetry.id
  member        = each.value
}

resource "google_kms_crypto_key_iam_member" "telemetry_decrypters" {
  for_each      = local.telemetry_key_decrypters
  role          = "roles/cloudkms.cryptoKeyDecrypter"
  crypto_key_id = google_kms_crypto_key.telemetry.id
  member        = each.value
}

resource "google_kms_crypto_key_iam_member" "telemetry_encrypters_decrypters" {
  for_each      = local.telemetry_key_encrypters_decrypters
  crypto_key_id = google_kms_crypto_key.telemetry.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = each.value
}
