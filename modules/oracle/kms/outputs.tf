output "vault_id" {
  value = local.vault_id
}

# The name the vault actually carries, which is only the one this module chose when it
# created it - an adopted vault came with its own.
output "vault_name" {
  value = var.create_vault ? local.vault_name : data.oci_kms_vault.this[0].display_name
}

# Needed by anything that creates further keys in this vault - key operations do not go to
# the regional API endpoint, they go to the vault's own management endpoint.
output "vault_management_endpoint" {
  value = local.management_endpoint
}

output "vault_crypto_endpoint" {
  value = var.create_vault ? oci_kms_vault.this[0].crypto_endpoint : data.oci_kms_vault.this[0].crypto_endpoint
}

# Named to match modules/aws/kms and modules/google/kms, so a module consuming "the data
# key" reads identically on every cloud. There is no _arn twin here: an OCID is the only
# identifier OCI has.
#
# The depends_on carries the service-grant wait to every consumer: a bucket, volume or cluster
# that names one of these keys before the grant propagates is rejected outright rather than
# retried, the same trap modules/oracle/pca has with its certificate authority.
output "data_key_id" {
  value      = local.data_key_id
  depends_on = [time_sleep.key_policy]
}

output "config_key_id" {
  value      = local.config_key_id
  depends_on = [time_sleep.key_policy]
}

output "telemetry_key_id" {
  value      = local.telemetry_key_id
  depends_on = [time_sleep.key_policy]
}

# Consumed by modules/oracle/pca as ca_key_id, via .toptout so that a deployment without
# this module still plans. Empty rather than null for the same reason - the agent renders
# the value into a string input.
#
# The grant that lets a certificate authority actually use this key, and the wait for it to
# propagate, live in modules/oracle/pca: they are about the CA rather than the key, and the
# module that creates the CA is the one that has to hold off until the grant lands.
output "ca_key_id" {
  value = local.ca_key_id
}
