locals {
  subdomain_name = var.subdomain_name != "" ? var.subdomain_name : var.prefix
  domain         = "${local.subdomain_name}.${var.parent_domain}"

  # An explicitly supplied certificate_ocid wins: that is how a deployment brings a
  # publicly trusted certificate instead of using the CA below.
  create_cert = var.create_cert && var.certificate_ocid == ""
}

resource "oci_dns_zone" "pub" {
  compartment_id = var.compartment_id
  name           = local.domain
  zone_type      = "PRIMARY"
  scope          = "GLOBAL"
}

# Certificate authority and certificate names must be unique in the tenancy *including
# objects that are only scheduled for deletion*, and neither can be deleted immediately
# (OCI enforces a minimum waiting period). A teardown followed by a rebuild the same day
# would collide with its own leftovers without this.
resource "random_string" "suffix" {
  length  = 8
  lower   = true
  upper   = false
  numeric = true
  special = false
}

# OCI has no equivalent of ACM's free publicly trusted issuance, so where modules/aws/route53
# asks ACM for "*.${domain}" this asks an internal CA of our own for the same thing. The
# certificates are therefore trusted only by clients that have imported the CA - see the
# README for how to fetch its PEM.
#
# The lifetime is set explicitly because OCI's default validity is measured in months, which
# is fine for a leaf certificate and wrong for a root: replacing the CA means redistributing
# it to every client.
resource "time_offset" "ca_validity" {
  count        = local.create_cert ? 1 : 0
  offset_years = var.ca_validity_years
}

resource "oci_certificates_management_certificate_authority" "this" {
  count          = local.create_cert ? 1 : 0
  compartment_id = var.compartment_id
  name           = "${var.prefix}-root-ca-${random_string.suffix.result}"
  description    = "Root CA issuing the wildcard certificate for ${local.domain}"

  # Must be an HSM-protected asymmetric key: OCI Certificates rejects software-protected
  # keys outright. modules/oracle/kms creates one and this is wired to it by agent_input.yaml.
  kms_key_id = var.ca_key_id

  certificate_authority_config {
    config_type = "ROOT_CA_GENERATED_INTERNALLY"

    subject {
      common_name = "${local.domain} root CA"
    }

    validity {
      time_of_validity_not_after = time_offset.ca_validity[0].rfc3339
    }
  }

  # Stated rather than left to OCI's defaults, which are undocumented and would put the
  # ceiling on what this CA may issue uncomfortably close to the validity of the
  # certificate below.
  certificate_authority_rules {
    rule_type                                   = "CERTIFICATE_AUTHORITY_ISSUANCE_EXPIRY_RULE"
    leaf_certificate_max_validity_duration      = var.leaf_certificate_max_validity
    certificate_authority_max_validity_duration = var.subordinate_ca_max_validity
  }

  lifecycle {
    precondition {
      condition     = var.ca_key_id != ""
      error_message = "ca_key_id is empty: add an oracle/kms module to the same step (it supplies ca_key_id through .toptout), or set create_cert = false / certificate_ocid to bring your own certificate."
    }
  }
}

# One wildcard certificate for the whole zone, because modules/k8s/oci-native-ingress-controller
# terminates TLS on an OCI load balancer listener and a listener holds exactly one key pair -
# there is no SNI, so every app on the cluster is served from this one certificate.
#
# The wildcard is repeated in the SANs rather than left in the common name alone: modern
# clients ignore CN entirely and match against subjectAltName only.
resource "oci_certificates_management_certificate" "wildcard" {
  count          = local.create_cert ? 1 : 0
  compartment_id = var.compartment_id
  name           = "${var.prefix}-wildcard-${random_string.suffix.result}"
  description    = "Wildcard certificate for ${local.domain}, served by every app behind the native ingress controller"

  certificate_config {
    config_type                     = "ISSUED_BY_INTERNAL_CA"
    issuer_certificate_authority_id = oci_certificates_management_certificate_authority.this[0].id
    certificate_profile_type        = "TLS_SERVER"

    subject {
      common_name = "*.${local.domain}"
    }

    subject_alternative_names {
      type  = "DNS"
      value = "*.${local.domain}"
    }

    subject_alternative_names {
      type  = "DNS"
      value = local.domain
    }
  }

  # OCI renews the certificate itself, which is the whole reason this module can own it -
  # an imported certificate would have to be re-issued and re-imported by hand every time.
  # Validity is left at OCI's default (three months), so renewing every 60 days with 15
  # days of headroom leaves a wide margin.
  #
  # KNOWN GAP: a renewal produces a new *version* of the same certificate, and an OCI load
  # balancer listener has been observed to keep serving the previous version until the
  # ingress controller re-pushes it. Until that is confirmed fixed for CA-issued
  # certificates, treat renewal as needing a `kubectl rollout restart` of
  # oci-native-ingress-controller.
  certificate_rules {
    rule_type              = "CERTIFICATE_RENEWAL_RULE"
    renewal_interval       = var.certificate_renewal_interval
    advance_renewal_period = var.certificate_advance_renewal_period
  }
}
