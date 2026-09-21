# Read for its management endpoint: every key operation goes to the vault's own hostname
# rather than the regional one, and only the vault knows it.
data "oci_kms_vault" "this" {
  count    = var.create_vault ? 0 : 1
  vault_id = var.vault_id

  lifecycle {
    # Caught here rather than left to the provider, which reports an empty OCID as
    # "can not marshal a nil pointer".
    precondition {
      condition     = var.vault_id != ""
      error_message = "vault_id is empty: set it to the OCID of the vault to hold these keys, or set create_vault = true."
    }
  }
}

# Same adopt-instead-of-create pattern as the vault above, one lookup per key: create_keys =
# false reuses whatever already carries these names in the vault rather than creating a fresh
# (randomly suffixed) set on every apply and orphaning the last one. ENABLED is a key's "in
# use" state, the equivalent of a vault's ACTIVE.
data "oci_kms_keys" "data" {
  count               = var.create_keys ? 0 : 1
  compartment_id      = var.compartment_id
  management_endpoint = local.management_endpoint

  filter {
    name   = "display_name"
    values = [var.data_key_name]
  }

  filter {
    name   = "state"
    values = ["ENABLED"]
  }
}

data "oci_kms_keys" "config" {
  count               = var.create_keys ? 0 : 1
  compartment_id      = var.compartment_id
  management_endpoint = local.management_endpoint

  filter {
    name   = "display_name"
    values = [var.config_key_name]
  }

  filter {
    name   = "state"
    values = ["ENABLED"]
  }
}

data "oci_kms_keys" "telemetry" {
  count               = var.create_keys ? 0 : 1
  compartment_id      = var.compartment_id
  management_endpoint = local.management_endpoint

  filter {
    name   = "display_name"
    values = [var.telemetry_key_name]
  }

  filter {
    name   = "state"
    values = ["ENABLED"]
  }
}

# Only looked up when the CA key is wanted at all - create_ca_key = false means no CA key
# either way, same as the create path.
data "oci_kms_keys" "ca" {
  count               = var.create_ca_key && !var.create_keys ? 1 : 0
  compartment_id      = var.compartment_id
  management_endpoint = local.management_endpoint

  filter {
    name   = "display_name"
    values = [var.ca_key_name]
  }

  filter {
    name   = "state"
    values = ["ENABLED"]
  }
}
