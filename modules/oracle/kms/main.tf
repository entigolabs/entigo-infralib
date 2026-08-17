locals {
  # The verb and resource-type differ per consumer, so these are written out rather than
  # generated from a list of service names. Getting one wrong does not read as a permission
  # problem: OKE reports "Invalid COMPUTE_INSTANCE: Authorization failed or requested resource
  # not found while provisioning node(s)" and the node pool simply never appears.
  key_statements = concat(
    # OKE etcd encryption. The service reads the key itself here, so "use keys" is right - and
    # it is *not* what worker node boot volumes need.
    var.grant_oke ? ["Allow service oke to use keys in compartment id ${var.compartment_id}"] : [],

    # Worker node boot volumes, which are three statements rather than one. OKE does not
    # encrypt them itself; it delegates to Block Volume, so it needs key-delegates rather than
    # keys, Block Volume needs keys, and since 2024-08-15 the principal that actually asks is
    # the *node pool* - the any-user statement is what Oracle's own documentation now
    # prescribes, and without it node provisioning fails with no mention of a key.
    var.grant_oke ? [
      "Allow service oke to use key-delegates in compartment id ${var.compartment_id}",
      "Allow any-user to use key-delegates in compartment id ${var.compartment_id} where all { request.principal.type = 'nodepool', request.principal.compartment.id = '${var.compartment_id}' }",
    ] : [],

    var.grant_block_storage ? ["Allow service blockstorage to use keys in compartment id ${var.compartment_id}"] : [],

    # Object Storage's principal is region-scoped (objectstorage-eu-frankfurt-1), unlike the
    # others. Dropped rather than half-formed when the region is unknown, so a deployment that
    # never wired one does not create a policy naming "objectstorage-".
    var.grant_object_storage && var.region != "" ? ["Allow service objectstorage-${var.region} to use keys in compartment id ${var.compartment_id}"] : [],

    var.extra_key_statements,
  )

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

# The signing key for modules/oracle/pca's certificate authority. It lives here rather than
# in that module because this is the module that owns the vault - a second vault would cost
# a 7-day deletion floor of its own - and because a CA key is worth seeing next to the
# others when reviewing what a deployment pays for: this is the only HSM-protected key
# created by default.
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

# Lets the OCI services that encrypt on our behalf actually use these keys. Without it the
# failure is not a permission error where you would expect one: OKE refuses to create a
# cluster whose etcd key it cannot read, and a bucket or volume naming a key it cannot use is
# rejected at create time.
#
# One statement per service rather than one policy each, so the whole grant is visible in a
# single object and oci-nuke sweeps it with the rest of the compartment's policies.
resource "oci_identity_policy" "key_services" {
  count          = length(local.key_statements) > 0 ? 1 : 0
  compartment_id = var.compartment_id
  name           = "${var.prefix}-key-services"
  description    = "Lets OCI services encrypt with the keys in this compartment"

  statements = local.key_statements
}

# Consumers inherit this through the key outputs, so nothing encrypts with a key before the
# grant has had time to propagate.
resource "time_sleep" "key_policy" {
  count           = length(local.key_statements) > 0 ? 1 : 0
  depends_on      = [oci_identity_policy.key_services]
  create_duration = var.key_policy_wait

  triggers = {
    policy_id = oci_identity_policy.key_services[0].id
  }
}
