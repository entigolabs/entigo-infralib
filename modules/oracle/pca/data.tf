# Same adopt-instead-of-create pattern as modules/oracle/kms's vault/key lookups: reuses
# whatever already carries ca_name rather than creating a fresh (randomly suffixed) CA on
# every apply and orphaning the last one. No ca_key_id needed here - an already-ACTIVE CA's
# signing key is fixed for its lifetime, so there is nothing to grant or wait on.
data "oci_certificates_management_certificate_authorities" "this" {
  count          = var.create_ca && var.adopt_ca ? 1 : 0
  compartment_id = var.compartment_id

  filter {
    name   = "name"
    values = [var.ca_name]
  }

  filter {
    name   = "state"
    values = ["ACTIVE"]
  }
}
