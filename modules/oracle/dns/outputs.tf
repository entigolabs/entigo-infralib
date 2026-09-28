# The pub_*/int_* pairs name one domain each, chosen by default_public/default_private, so
# that a consumer can ask for "the public domain" without knowing how many exist. Same
# contract as modules/aws-v2/route53, with OCIDs where AWS has ARNs.
#
# The precondition lives here rather than on a resource because outputs are always evaluated:
# a deployment whose zones all have create_zone = false creates no resource that could carry it.
output "pub_zone_id" {
  description = "Zone OCID of the domain marked as default public"
  value       = local.default_public_key != "" ? local.zone_ids[local.default_public_key] : ""

  precondition {
    condition     = length(local.default_public_keys) <= 1 && length(local.default_private_keys) <= 1
    error_message = "More than one domain is marked as a default: default_public=${jsonencode(local.default_public_keys)}, default_private=${jsonencode(local.default_private_keys)}. Exactly one domain of each kind may be the default."
  }

  # Without this, an ambiguous set - two public domains and no flags, say - leaves every
  # pub_*/int_* output an empty string instead of failing. Nothing downstream notices: the apps
  # would be deployed with an empty hostname and an empty certificate OCID.
  precondition {
    condition     = length(var.domains) == 0 || local.default_public_key != ""
    error_message = "No domain could be chosen as the default. A default is only inferred for a single domain, or for the only public/private one of its kind - with ${local.public_count} public and ${local.private_count} private domains, set default_public = true (and default_private = true, if any are private) on the ones the pub_*/int_* outputs should describe."
  }
}

output "pub_domain" {
  description = "Domain name of the zone marked as default public"
  value       = local.default_public_key != "" ? local.domain_names[local.default_public_key] : ""
}

output "pub_cert_ocid" {
  description = "Certificate OCID of the domain marked as default public"
  value       = local.default_public_key != "" ? local.certificate_ocids[local.default_public_key] : ""
}

output "int_zone_id" {
  description = "Zone OCID of the domain marked as default private, falling back to the public default"
  value       = local.default_private_key != "" ? local.zone_ids[local.default_private_key] : ""
}

output "int_domain" {
  description = "Domain name of the zone marked as default private, falling back to the public default"
  value       = local.default_private_key != "" ? local.domain_names[local.default_private_key] : ""
}

# Consumed by the Oracle apps as the oci-native-ingress.oraclecloud.com/certificate-ocid
# annotation - every app on a domain uses this same value, because they share one load
# balancer listener and a listener holds one key pair.
output "int_cert_ocid" {
  description = "Certificate OCID of the domain marked as default private, falling back to the public default"
  value       = local.default_private_key != "" ? local.certificate_ocids[local.default_private_key] : ""
}

output "zone_ids" {
  description = "Map of domain keys to their zone OCIDs"
  value       = local.zone_ids
}

output "domain_names" {
  description = "Map of domain keys to their domain names"
  value       = local.domain_names
}

output "certificate_ocids" {
  description = "Map of domain keys to their certificate OCIDs, empty for domains with no certificate"
  value       = local.certificate_ocids
}

# The NS delegation has to be added by hand in the parent zone whenever that zone is not in
# OCI DNS - which is the usual case here, the parent living in Route53. These are the values
# to add. Domains with a parent_zone_id had it created for them.
output "nameservers" {
  description = "Map of domain keys to their nameservers, for zones this module created"
  value       = local.nameservers
}

# The CA that signed the certificates, passed straight through from the input. Kept as an
# output so a consumer needing the issuer does not also have to know which module created it -
# modules/oracle/pca has the command that exports its PEM.
output "certificate_authority_id" {
  description = "OCID of the certificate authority used for domains that named none themselves"
  value       = var.certificate_authority_id
}
