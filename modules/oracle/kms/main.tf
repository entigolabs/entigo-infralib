locals {
  vault_name = var.vault_name != "" ? var.vault_name : "${var.prefix}-${random_string.suffix.result}"

  vault_id            = var.create_vault ? oci_kms_vault.this[0].id : data.oci_kms_vaults.this[0].vaults[0].id
  management_endpoint = var.create_vault ? oci_kms_vault.this[0].management_endpoint : data.oci_kms_vaults.this[0].vaults[0].management_endpoint

  data_key_name      = "${var.prefix}-data-${random_string.suffix.result}"
  config_key_name    = "${var.prefix}-config-${random_string.suffix.result}"
  telemetry_key_name = "${var.prefix}-telemetry-${random_string.suffix.result}"
  ca_key_name        = "${var.prefix}-ca-${random_string.suffix.result}"
}

# Every name here carries a random suffix, for the same reason modules/oracle/dns's
# certificate does: deleting a vault or a key only *schedules* the deletion (7 days
# minimum, 30 maximum), and the object keeps its name for the whole waiting period. A
# teardown followed by a rebuild the same day - which is the normal development cycle
# here - would otherwise collide with its own pending-deletion leftovers.
resource "random_string" "suffix" {
  length  = 8
  lower   = true
  upper   = false
  numeric = true
  special = false
}

resource "oci_kms_vault" "this" {
  count          = var.create_vault ? 1 : 0
  compartment_id = var.compartment_id
  display_name   = local.vault_name
  vault_type     = var.vault_type
}

# CreateVault returns before the vault's management endpoint is resolvable, and every key
# below is created against that endpoint rather than the regional one - so without this
# they all fail together with "no such host". See vault_endpoint_wait.
#
# Keyed on the vault's id so it waits again if the vault is ever replaced, and skipped
# entirely when an existing vault is being reused: that one's DNS resolved long ago.
resource "time_sleep" "vault_endpoint" {
  count           = var.create_vault ? 1 : 0
  depends_on      = [oci_kms_vault.this]
  create_duration = var.vault_endpoint_wait

  triggers = {
    vault_id = oci_kms_vault.this[0].id
  }
}

# The three keys mirror modules/aws/kms and modules/google/kms one for one, so that a
# deployment reads the same on all three clouds:
#   data      - block volumes, boot volumes, object storage buckets, databases
#   config    - secrets and cluster configuration (OKE etcd)
#   telemetry - log and metric storage
# Consumers take the OCID; unlike AWS there is no key policy to attach here, because OCI
# authorises key use through IAM policies on the compartment rather than through a
# document on the key itself.
resource "oci_kms_key" "data" {
  depends_on          = [time_sleep.vault_endpoint]
  compartment_id      = var.compartment_id
  display_name        = local.data_key_name
  management_endpoint = local.management_endpoint
  protection_mode     = var.key_protection_mode

  key_shape {
    algorithm = var.key_algorithm
    length    = var.key_length
  }

  dynamic "auto_key_rotation_details" {
    for_each = var.key_rotation_interval_in_days == null ? [] : [1]
    content {
      rotation_interval_in_days = var.key_rotation_interval_in_days
    }
  }
}

resource "oci_kms_key" "config" {
  depends_on          = [time_sleep.vault_endpoint]
  compartment_id      = var.compartment_id
  display_name        = local.config_key_name
  management_endpoint = local.management_endpoint
  protection_mode     = var.key_protection_mode

  key_shape {
    algorithm = var.key_algorithm
    length    = var.key_length
  }

  dynamic "auto_key_rotation_details" {
    for_each = var.key_rotation_interval_in_days == null ? [] : [1]
    content {
      rotation_interval_in_days = var.key_rotation_interval_in_days
    }
  }
}

resource "oci_kms_key" "telemetry" {
  depends_on          = [time_sleep.vault_endpoint]
  compartment_id      = var.compartment_id
  display_name        = local.telemetry_key_name
  management_endpoint = local.management_endpoint
  protection_mode     = var.key_protection_mode

  key_shape {
    algorithm = var.key_algorithm
    length    = var.key_length
  }

  dynamic "auto_key_rotation_details" {
    for_each = var.key_rotation_interval_in_days == null ? [] : [1]
    content {
      rotation_interval_in_days = var.key_rotation_interval_in_days
    }
  }
}

# The signing key for modules/oracle/dns's certificate authority. It lives here rather
# than in that module because this is the module that owns the vault, and because a CA key
# is worth seeing next to the others when reviewing what a deployment pays for: this is the
# only HSM-protected key created by default.
#
# No auto_key_rotation_details: rotating a CA's signing key mid-life would invalidate the
# CA. Certificate rotation is handled by the renewal rule on the certificate itself.
resource "oci_kms_key" "ca" {
  count               = var.create_ca_key ? 1 : 0
  depends_on          = [time_sleep.vault_endpoint]
  compartment_id      = var.compartment_id
  display_name        = local.ca_key_name
  management_endpoint = local.management_endpoint
  protection_mode     = var.ca_key_protection_mode

  key_shape {
    algorithm = var.ca_key_algorithm
    length    = var.ca_key_length
  }
}

# Creating a certificate authority is not enough to make one work: the CA then reaches for
# its signing key as *itself*, and without this it is refused. The CA is created, goes to
# FAILED a few seconds later, and reports only "Authorization failed or requested resource
# not found: Key Id ocid1.key..." in the Console - `lifecycle-details` on the API is empty,
# and terraform just says the service reported an unexpected state. Every CA created here
# failed this way until the grant existed.
#
# Dynamic group rather than a service principal because that is how OCI models it: CAs are
# resources, matched by type, and a policy grants to the group.
resource "oci_identity_dynamic_group" "certificate_authorities" {
  count          = var.create_ca_key ? 1 : 0
  compartment_id = data.oci_identity_compartment.this[0].compartment_id
  name           = "${var.prefix}-certificate-authorities"
  description    = "Certificate authorities in ${var.prefix}'s compartment, so they can use the vault key they sign with"
  matching_rule  = "ALL {resource.type='certificateauthority', resource.compartment.id='${var.compartment_id}'}"
}

resource "oci_identity_policy" "certificate_authorities" {
  count          = var.create_ca_key ? 1 : 0
  compartment_id = var.compartment_id
  name           = "${var.prefix}-certificate-authorities"
  description    = "Lets certificate authorities in this compartment use the keys in it"

  # Attached at the compartment, not the tenancy: unlike the ingress controller's grants
  # this needs nothing outside the compartment, so it stays inside it. Could be narrowed
  # further with "where target.key.id = '<the ca key>'" if one key per CA is ever worth
  # spelling out.
  statements = [
    "Allow dynamic-group ${oci_identity_dynamic_group.certificate_authorities[0].name} to use keys in compartment id ${var.compartment_id}",
  ]
}

# IAM is eventually consistent, and a CA that starts before the grant lands does not retry -
# it goes to FAILED and stays there, needing a teardown that OCI will not do for 7 days. The
# ca_key_id output depends on this, so modules/oracle/dns cannot begin creating the CA until
# the grant has had time to take effect.
resource "time_sleep" "ca_policy" {
  count           = var.create_ca_key ? 1 : 0
  depends_on      = [oci_identity_policy.certificate_authorities]
  create_duration = var.ca_policy_wait

  triggers = {
    policy_id = oci_identity_policy.certificate_authorities[0].id
  }
}
