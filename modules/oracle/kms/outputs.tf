output "vault_id" {
  value = local.vault_id
}

output "vault_name" {
  value = local.vault_name
}

# Needed by anything that creates further keys in this vault - key operations do not go to
# the regional API endpoint, they go to the vault's own management endpoint.
output "vault_management_endpoint" {
  value = local.management_endpoint
}

output "vault_crypto_endpoint" {
  value = var.create_vault ? oci_kms_vault.this[0].crypto_endpoint : data.oci_kms_vaults.this[0].vaults[0].crypto_endpoint
}

# Named to match modules/aws/kms and modules/google/kms, so a module consuming "the data
# key" reads identically on every cloud. There is no _arn twin here: an OCID is the only
# identifier OCI has.
output "data_key_id" {
  value = oci_kms_key.data.id
}

output "config_key_id" {
  value = oci_kms_key.config.id
}

output "telemetry_key_id" {
  value = oci_kms_key.telemetry.id
}

# Consumed by modules/oracle/dns as ca_key_id, via .toptout so that a deployment without
# this module still plans. Empty rather than null for the same reason - the agent renders
# the value into a string input.
#
# The depends_on is what stops a certificate authority being created before it is allowed to
# use this key: consumers of this output inherit the wait, and a CA that starts too early
# fails permanently rather than retrying.
output "ca_key_id" {
  value      = var.create_ca_key ? oci_kms_key.ca[0].id : ""
  depends_on = [time_sleep.ca_policy]
}
