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
