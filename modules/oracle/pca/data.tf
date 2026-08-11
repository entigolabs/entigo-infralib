# Dynamic groups are tenancy-scoped in OCI, so the group in main.tf needs the tenancy OCID
# and not the compartment. Same reasoning as modules/oracle/oke's copy of this lookup: these
# compartments are flat, one level below the root, so the parent of var.compartment_id is
# reliably the tenancy.
data "oci_identity_compartment" "this" {
  count = local.create_policy ? 1 : 0
  id    = var.compartment_id
}
