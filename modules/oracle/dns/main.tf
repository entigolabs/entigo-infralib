locals {
  public_count  = length([for k, v in var.domains : k if !v.private])
  private_count = length([for k, v in var.domains : k if v.private])

  # Everything derived per domain, so no resource has to repeat a fallback. Same intent as
  # domains_with_defaults in modules/aws-v2/route53.
  domains = {
    for key, d in var.domains : key => merge(d, {
      # A domain that names no authority uses the module-level one, which is where
      # modules/oracle/pca is wired in.
      certificate_authority_id = d.certificate_authority_id != "" ? d.certificate_authority_id : var.certificate_authority_id

      # An explicitly supplied certificate_ocid wins: that is how a deployment brings a
      # publicly trusted certificate instead of one issued by the CA.
      create_cert = d.create_certificate && d.certificate_ocid == ""

      effective_vcn_id = d.vcn_id != "" ? d.vcn_id : var.vcn_id

      # Which domain the pub_*/int_* outputs describe. Explicit wins; otherwise a single
      # domain is both, and a single domain of a kind is the default for that kind.
      is_default_public = d.default_public != null ? d.default_public : (
        length(var.domains) == 1 ? true : (!d.private && local.public_count == 1)
      )
      is_default_private = d.default_private != null ? d.default_private : (
        length(var.domains) == 1 ? true : (d.private && local.private_count == 1)
      )
    })
  }

  default_public_keys  = [for k, v in local.domains : k if v.is_default_public]
  default_private_keys = [for k, v in local.domains : k if v.is_default_private]

  chosen_public  = length(local.default_public_keys) > 0 ? local.default_public_keys[0] : ""
  chosen_private = length(local.default_private_keys) > 0 ? local.default_private_keys[0] : ""

  # Each kind falls back to the other, a deliberate deviation from modules/aws-v2/route53 which
  # requires exactly one default of *each* and fails a deployment that has only public domains.
  # This is the behaviour modules/aws/route53 had through `int_domain = create_private ? ... :
  # local.pub_domain`, and every Oracle app reads .toutput.<dns>.int_domain - so the fallback is
  # what lets a single-public-zone deployment work at all.
  default_public_key  = local.chosen_public != "" ? local.chosen_public : local.chosen_private
  default_private_key = local.chosen_private != "" ? local.chosen_private : local.chosen_public

  # A private zone needs the view its VCN's resolver keeps; an explicit view_id skips the
  # lookup. Only computed for private domains - the data sources below are keyed the same way.
  view_ids = {
    for k, v in local.domains : k => v.view_id != "" ? v.view_id : data.oci_dns_resolver.this[k].default_view_id
    if v.private
  }

  zone_ids = {
    for k, v in local.domains : k => v.create_zone ? oci_dns_zone.this[k].id : data.oci_dns_zones.existing[k].zones[0].id
  }

  domain_names = { for k, v in local.domains : k => v.domain_name }

  certificate_ocids = {
    for k, v in local.domains : k => v.certificate_ocid != "" ? v.certificate_ocid : (
      v.create_cert ? oci_certificates_management_certificate.this[k].id : ""
    )
  }

  # Only zones this module created report nameservers; an existing zone was delegated long ago.
  nameservers = {
    for k, v in local.domains : k => oci_dns_zone.this[k].nameservers[*].hostname
    if v.create_zone
  }

  name_suffix = var.name_salt ? "-${random_string.suffix[0].result}" : ""
}

resource "oci_dns_zone" "this" {
  for_each = { for k, v in local.domains : k => v if v.create_zone }

  compartment_id = var.compartment_id
  name           = each.value.domain_name
  zone_type      = "PRIMARY"

  # GLOBAL is the public DNS namespace; a PRIVATE zone is resolvable only through the view it
  # belongs to, which is what confines it to a VCN.
  scope   = each.value.private ? "PRIVATE" : "GLOBAL"
  view_id = each.value.private ? local.view_ids[each.key] : null

  lifecycle {
    precondition {
      condition     = !each.value.private || each.value.effective_vcn_id != "" || each.value.view_id != ""
      error_message = "Domain '${each.key}' is private but has no VCN: set vcn_id on the domain, set the module-level vcn_id (wired from oracle/vpc), or name a view_id directly."
    }
  }
}

# Zones with create_zone = false are assumed to exist already. Looked up by name rather than
# taken as an OCID so that create_zone reads the same as it does on AWS; oci_dns_zones (the
# plural, list-and-filter data source) is used because the singular requires the OCID that we
# are trying to find.
data "oci_dns_zones" "existing" {
  for_each = { for k, v in local.domains : k => v if !v.create_zone }

  compartment_id = var.compartment_id
  name           = each.value.domain_name
  scope          = each.value.private ? "PRIVATE" : "GLOBAL"
  view_id        = each.value.private ? local.view_ids[each.key] : null
  state          = "ACTIVE"
}

# The NS delegation, when the parent zone is in OCI DNS too. When it is not - the usual case
# here, where the parent is in Route53 - this creates nothing and the records are added by
# hand from the nameservers output.
#
# UNTESTED: every deployment so far has had its parent zone outside OCI.
resource "oci_dns_rrset" "ns_delegation" {
  for_each = {
    for k, v in local.domains : k => v
    if v.parent_zone_id != "" && v.create_zone && !v.private
  }

  zone_name_or_id = each.value.parent_zone_id
  domain          = each.value.domain_name
  rtype           = "NS"

  dynamic "items" {
    for_each = oci_dns_zone.this[each.key].nameservers[*].hostname
    content {
      domain = each.value.domain_name
      rtype  = "NS"
      rdata  = items.value
      ttl    = 300
    }
  }
}

# Only created when name_salt asks for it - see that variable for why a rebuild needs it and
# why it is off by default.
resource "random_string" "suffix" {
  count = var.name_salt ? 1 : 0

  length  = 8
  lower   = true
  upper   = false
  numeric = true
  special = false
}

# One wildcard certificate per domain, because modules/k8s/oci-native-ingress-controller
# terminates TLS on an OCI load balancer listener and a listener holds exactly one key pair -
# there is no SNI, so every app served from a given domain shares its certificate.
#
# The wildcard is repeated in the SANs rather than left in the common name alone: modern
# clients ignore CN entirely and match against subjectAltName only.
resource "oci_certificates_management_certificate" "this" {
  for_each = { for k, v in local.domains : k => v if v.create_cert }

  compartment_id = var.compartment_id
  name           = "${var.prefix}-${each.key}${local.name_suffix}"
  description    = "Wildcard certificate for ${each.value.domain_name}"

  certificate_config {
    config_type                     = "ISSUED_BY_INTERNAL_CA"
    issuer_certificate_authority_id = each.value.certificate_authority_id
    certificate_profile_type        = "TLS_SERVER"

    subject {
      common_name = "*.${each.value.domain_name}"
    }

    subject_alternative_names {
      type  = "DNS"
      value = "*.${each.value.domain_name}"
    }

    subject_alternative_names {
      type  = "DNS"
      value = each.value.domain_name
    }
  }

  # OCI renews the certificate itself, which is the whole reason this module can own it - an
  # imported certificate would have to be re-issued and re-imported by hand every time.
  # Validity is left at OCI's default (three months), so renewing every 60 days with 15 days
  # of headroom leaves a wide margin.
  #
  # KNOWN GAP: a renewal produces a new *version* of the same certificate, and an OCI load
  # balancer listener has been observed to keep serving the previous version until the ingress
  # controller re-pushes it. Until that is confirmed fixed for CA-issued certificates, treat
  # renewal as needing a `kubectl rollout restart` of oci-native-ingress-controller.
  certificate_rules {
    rule_type              = "CERTIFICATE_RENEWAL_RULE"
    renewal_interval       = var.certificate_renewal_interval
    advance_renewal_period = var.certificate_advance_renewal_period
  }

  lifecycle {
    precondition {
      condition     = each.value.certificate_authority_id != ""
      error_message = "Domain '${each.key}' has create_certificate = true but no issuing authority: add an oracle/pca module to the same step (it supplies certificate_authority_id through .toptout), name a CA on the domain, or set certificate_ocid to bring your own certificate."
    }
  }
}
