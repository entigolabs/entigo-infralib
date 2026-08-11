locals {
  name_suffix = var.name_salt ? "-${random_string.suffix[0].result}" : ""

  ca_name     = var.ca_name != "" ? var.ca_name : "${var.prefix}-root-ca${local.name_suffix}"
  common_name = var.common_name != "" ? var.common_name : "${var.prefix} root CA"

  # A CA signed by another CA is a subordinate; one that signs itself is a root. OCI spells
  # the difference as the config type, and only the subordinate form takes an issuer.
  subordinate = var.issuer_certificate_authority_id != ""
  config_type = local.subordinate ? "SUBORDINATE_CA_ISSUED_BY_INTERNAL_CA" : "ROOT_CA_GENERATED_INTERNALLY"

  description = var.description != "" ? var.description : "${local.subordinate ? "Subordinate" : "Root"} CA for ${var.prefix}"

  create_policy = var.create_ca && var.create_ca_policy
}

resource "random_string" "suffix" {
  count = var.name_salt ? 1 : 0

  length  = 8
  lower   = true
  upper   = false
  numeric = true
  special = false
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
  count          = local.create_policy ? 1 : 0
  compartment_id = data.oci_identity_compartment.this[0].compartment_id
  name           = "${var.prefix}-certificate-authorities"
  description    = "Certificate authorities in ${var.prefix}'s compartment, so they can use the vault key they sign with"
  matching_rule  = "ALL {resource.type='certificateauthority', resource.compartment.id='${var.compartment_id}'}"
}

resource "oci_identity_policy" "certificate_authorities" {
  count          = local.create_policy ? 1 : 0
  compartment_id = var.compartment_id
  name           = "${var.prefix}-certificate-authorities"
  description    = "Lets certificate authorities in this compartment use the keys in it"

  # Attached at the compartment, not the tenancy: unlike the ingress controller's grants this
  # needs nothing outside the compartment, so it stays inside it. Could be narrowed further
  # with "where target.key.id = '<the ca key>'" if one key per CA is ever worth spelling out.
  statements = [
    "Allow dynamic-group ${oci_identity_dynamic_group.certificate_authorities[0].name} to use keys in compartment id ${var.compartment_id}",
  ]
}

# IAM is eventually consistent, and a CA that starts before the grant lands does not retry -
# it goes to FAILED and stays there, needing a teardown that OCI will not do for 7 days.
resource "time_sleep" "ca_policy" {
  count           = local.create_policy ? 1 : 0
  depends_on      = [oci_identity_policy.certificate_authorities]
  create_duration = var.ca_policy_wait

  triggers = {
    policy_id = oci_identity_policy.certificate_authorities[0].id
  }
}

# The lifetime is set explicitly because OCI's default validity is measured in months, which
# is fine for a leaf certificate and wrong for a CA: replacing one means redistributing it to
# every client that trusts it.
resource "time_offset" "ca_validity" {
  count        = var.create_ca ? 1 : 0
  offset_years = var.ca_validity_years
}

resource "oci_certificates_management_certificate_authority" "this" {
  count          = var.create_ca ? 1 : 0
  depends_on     = [time_sleep.ca_policy]
  compartment_id = var.compartment_id
  name           = local.ca_name
  description    = local.description
  kms_key_id     = var.ca_key_id

  certificate_authority_config {
    config_type                     = local.config_type
    issuer_certificate_authority_id = local.subordinate ? var.issuer_certificate_authority_id : null
    signing_algorithm               = var.signing_algorithm

    subject {
      common_name            = local.common_name
      organization           = var.organization
      organizational_unit    = var.organizational_unit
      country                = var.country
      state_or_province_name = var.state_or_province_name
      locality_name          = var.locality_name
    }

    # Sub-second precision is not decoration, and ".000" will not do.
    #
    # OCI rejects an RFC3339 timestamp with no fraction - "400-InvalidParameter, Unable to
    # process JSON input", naming no field - so "2036-08-06T09:02:54Z" fails where
    # "...54.000Z" is accepted by the API. But the provider does not send what you write: it
    # parses the string into an SDKTime and re-serialises with time.RFC3339Nano, which
    # *trims trailing zeros*, turning ".000Z" straight back into "Z". Captured off the wire
    # with OCI_GO_SDK_DEBUG: config ".000Z" -> body "2036-08-06T08:44:05Z" -> rejected.
    #
    # A non-zero fraction survives the round trip, hence .5. Safe to hardcode Z because
    # time_offset always emits UTC.
    validity {
      time_of_validity_not_after = formatdate("YYYY-MM-DD'T'hh:mm:ss'.500Z'", time_offset.ca_validity[0].rfc3339)
    }
  }

  # Stated rather than left to OCI's defaults, which are undocumented and would put the
  # ceiling on what this CA may issue uncomfortably close to the validity of the certificates
  # it signs.
  certificate_authority_rules {
    rule_type                                   = "CERTIFICATE_AUTHORITY_ISSUANCE_EXPIRY_RULE"
    leaf_certificate_max_validity_duration      = var.leaf_certificate_max_validity
    certificate_authority_max_validity_duration = var.subordinate_ca_max_validity
  }

  lifecycle {
    precondition {
      condition     = var.ca_key_id != ""
      error_message = "ca_key_id is empty: add an oracle/kms module to the same step (it supplies ca_key_id through .toptout), or set create_ca = false."
    }
  }
}
