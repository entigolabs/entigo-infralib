# Looked up by name rather than taken as an OCID input, so that create_vault = false reads
# the same way as create_key_ring = false does in modules/google/kms. oci_kms_vaults (the
# plural, list-and-filter data source) is used because the singular oci_kms_vault requires
# the OCID that we are trying to find.
data "oci_kms_vaults" "this" {
  count          = var.create_vault ? 0 : 1
  compartment_id = var.compartment_id

  filter {
    name   = "display_name"
    values = [var.vault_name]
  }

  filter {
    name   = "state"
    values = ["ACTIVE"]
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
