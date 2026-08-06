# Dynamic groups are tenancy-scoped in OCI, so the group below needs the tenancy OCID and
# not the compartment. Same reasoning as modules/oracle/oke's copy of this lookup: these
# compartments are flat, one level below the root, so the parent of var.compartment_id is
# reliably the tenancy.
data "oci_identity_compartment" "this" {
  count = var.create_ca_key ? 1 : 0
  id    = var.compartment_id
}

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
