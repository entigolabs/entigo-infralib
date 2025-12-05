locals {
  kms_data_key_encrypters = var.kms_data_key_additional_encrypters
  kms_data_key_decrypters = var.kms_data_key_additional_decrypters
  kms_data_key_encrypters_decrypters = setunion(
    toset(var.kms_data_key_additional_encrypters_decrypters),
    toset([
      "serviceAccount:service-${data.google_project.this.number}@compute-system.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@container-engine-robot.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@gcp-sa-cloud-sql.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@gs-project-accounts.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@cloud-redis.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@cloud-filer.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@gcp-sa-artifactregistry.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@gcp-sa-pubsub.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@gcp-sa-secretmanager.iam.gserviceaccount.com",
    ])
  )

  kms_config_key_encrypters = var.kms_config_key_additional_encrypters
  kms_config_key_decrypters = var.kms_config_key_additional_decrypters
  kms_config_key_encrypters_decrypters = setunion(
    toset(var.kms_config_key_additional_encrypters_decrypters),
    toset([
      "serviceAccount:service-${data.google_project.this.number}@compute-system.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@container-engine-robot.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@gcp-sa-cloud-sql.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@gs-project-accounts.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@cloud-redis.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@cloud-filer.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@gcp-sa-artifactregistry.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@gcp-sa-pubsub.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@gcp-sa-secretmanager.iam.gserviceaccount.com",
    ])
  )

  kms_telemetry_key_encrypters = var.kms_telemetry_key_additional_encrypters
  kms_telemetry_key_decrypters = var.kms_telemetry_key_additional_decrypters
  kms_telemetry_key_encrypters_decrypters = setunion(
    toset(var.kms_telemetry_key_additional_encrypters_decrypters),
    toset([
      "serviceAccount:service-${data.google_project.this.number}@compute-system.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@container-engine-robot.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@gcp-sa-cloud-sql.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@gs-project-accounts.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@cloud-redis.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@cloud-filer.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@gcp-sa-artifactregistry.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@gcp-sa-pubsub.iam.gserviceaccount.com",
      "serviceAccount:service-${data.google_project.this.number}@gcp-sa-secretmanager.iam.gserviceaccount.com",
    ])
  )

  labels = merge(var.labels, { created-by = "entigo-infralib" })

  kms_key_ring_name      = var.kms_key_ring_name != "" ? var.kms_key_ring_name : "${var.prefix}-${random_string.suffix.result}"
  kms_data_key_name      = "${var.prefix}-data-${random_string.suffix.result}"
  kms_config_key_name    = "${var.prefix}-config-${random_string.suffix.result}"
  kms_telemetry_key_name = "${var.prefix}-telemetry-${random_string.suffix.result}"
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
resource "google_kms_key_ring" "kms_key_ring" {
  count    = var.create_kms_key_ring ? 1 : 0
  name     = local.kms_key_ring_name
  project  = data.google_client_config.this.project
  location = data.google_client_config.this.region
}

# KMS data key
resource "google_kms_crypto_key" "kms_data_key" {
  name                          = local.kms_data_key_name
  key_ring                      = var.create_kms_key_ring ? google_kms_key_ring.kms_key_ring[0].id : data.google_kms_key_ring.kms_key_ring[0].id
  rotation_period               = var.kms_key_rotation_period
  purpose                       = "ENCRYPT_DECRYPT"
  import_only                   = false
  skip_initial_version_creation = false

  lifecycle {
    prevent_destroy = true
  }

  destroy_scheduled_duration = var.kms_destroy_scheduled_duration

  version_template {
    algorithm        = "GOOGLE_SYMMETRIC_ENCRYPTION"
    protection_level = "SOFTWARE"
  }

  labels = local.labels
}

resource "google_kms_crypto_key_iam_binding" "kms_data_encrypters" {
  role          = "roles/cloudkms.cryptoKeyEncrypter"
  crypto_key_id = google_kms_crypto_key.kms_data_key.id
  members       = local.kms_data_key_encrypters
}

resource "google_kms_crypto_key_iam_binding" "kms_data_decrypters" {
  role          = "roles/cloudkms.cryptoKeyDecrypter"
  crypto_key_id = google_kms_crypto_key.kms_data_key.id
  members       = local.kms_data_key_decrypters
}

resource "google_kms_crypto_key_iam_binding" "kms_data_encrypt_decrypt" {
  crypto_key_id = google_kms_crypto_key.kms_data_key.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  members       = local.kms_data_key_encrypters_decrypters
}

# KMS config key
resource "google_kms_crypto_key" "kms_config_key" {
  name                          = local.kms_config_key_name
  key_ring                      = var.create_kms_key_ring ? google_kms_key_ring.kms_key_ring[0].id : data.google_kms_key_ring.kms_key_ring[0].id
  rotation_period               = var.kms_key_rotation_period
  destroy_scheduled_duration    = var.kms_destroy_scheduled_duration
  purpose                       = "ENCRYPT_DECRYPT"
  import_only                   = false
  skip_initial_version_creation = false

  lifecycle {
    prevent_destroy = true
  }

  version_template {
    algorithm        = "GOOGLE_SYMMETRIC_ENCRYPTION"
    protection_level = "SOFTWARE"
  }

  labels = local.labels
}

resource "google_kms_crypto_key_iam_binding" "kms_config_encrypters" {
  role          = "roles/cloudkms.cryptoKeyEncrypter"
  crypto_key_id = google_kms_crypto_key.kms_config_key.id
  members       = local.kms_config_key_encrypters
}

resource "google_kms_crypto_key_iam_binding" "kms_config_decrypters" {
  role          = "roles/cloudkms.cryptoKeyDecrypter"
  crypto_key_id = google_kms_crypto_key.kms_config_key.id
  members       = local.kms_config_key_decrypters
}

resource "google_kms_crypto_key_iam_binding" "kms_config_encrypt_decrypt" {
  crypto_key_id = google_kms_crypto_key.kms_config_key.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  members       = local.kms_config_key_encrypters_decrypters
}

# KMS telemetry key
resource "google_kms_crypto_key" "kms_telemetry_key" {
  name                          = local.kms_telemetry_key_name
  key_ring                      = var.create_kms_key_ring ? google_kms_key_ring.kms_key_ring[0].id : data.google_kms_key_ring.kms_key_ring[0].id
  rotation_period               = var.kms_key_rotation_period
  purpose                       = "ENCRYPT_DECRYPT"
  import_only                   = false
  skip_initial_version_creation = false

  lifecycle {
    prevent_destroy = true
  }

  destroy_scheduled_duration = var.kms_destroy_scheduled_duration

  version_template {
    algorithm        = "GOOGLE_SYMMETRIC_ENCRYPTION"
    protection_level = "SOFTWARE"
  }

  labels = local.labels
}

resource "google_kms_crypto_key_iam_binding" "kms_telemetry_encrypters" {
  role          = "roles/cloudkms.cryptoKeyEncrypter"
  crypto_key_id = google_kms_crypto_key.kms_telemetry_key.id
  members       = local.kms_telemetry_key_encrypters
}

resource "google_kms_crypto_key_iam_binding" "kms_telemetry_decrypters" {
  role          = "roles/cloudkms.cryptoKeyDecrypter"
  crypto_key_id = google_kms_crypto_key.kms_telemetry_key.id
  members       = local.kms_telemetry_key_decrypters
}

resource "google_kms_crypto_key_iam_binding" "kms_telemetry_encrypt_decrypt" {
  crypto_key_id = google_kms_crypto_key.kms_telemetry_key.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  members       = local.kms_telemetry_key_encrypters_decrypters
}
