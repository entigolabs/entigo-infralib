variable "prefix" {
  type = string
}

variable "compartment_id" {
  description = "OCID of the compartment that will contain the DNS zone."
  type        = string
}

variable "parent_domain" {
  description = "The domain this zone is a subdomain of, e.g. tarmo.entigo.dev. The parent zone lives elsewhere (e.g. Route53) - this module only creates the OCI-side zone and outputs its nameservers for manual NS delegation."
  type        = string
}

variable "subdomain_name" {
  description = "Subdomain label to create under parent_domain. Defaults to prefix if unset."
  type        = string
  default     = ""
}

variable "create_cert" {
  description = "Create an internal root CA and a wildcard certificate for this zone, mirroring the ACM certificate modules/aws/route53 creates. Requires ca_key_id."
  type        = bool
  default     = true
}

variable "ca_key_id" {
  description = "OCID of an HSM-protected asymmetric key for the certificate authority to sign with. Wired from modules/oracle/kms by agent_input.yaml; OCI Certificates will not accept a software-protected key."
  type        = string
  default     = ""
}

variable "ca_validity_years" {
  description = "Lifetime of the root CA. Replacing it means redistributing it to every client that trusts it, so this is deliberately long."
  type        = number
  default     = 10
}

variable "leaf_certificate_max_validity" {
  description = "Longest validity the CA will issue a certificate for, as an ISO 8601 duration. Must comfortably exceed the certificate's own validity (OCI's default is three months)."
  type        = string
  default     = "P365D"
}

variable "subordinate_ca_max_validity" {
  description = "Longest validity the CA will issue a subordinate CA for, as an ISO 8601 duration. Nothing here creates one; it is set so the rule is complete."
  type        = string
  default     = "P1095D"
}

# Left at OCI's default certificate validity (three months) rather than pinned here: an
# explicit validity is an absolute timestamp, which would have to be regenerated on every
# renewal. These two rules are what keep the certificate fresh inside that window.
variable "certificate_renewal_interval" {
  description = "How often OCI renews the wildcard certificate, as an ISO 8601 duration."
  type        = string
  default     = "P60D"
}

variable "certificate_advance_renewal_period" {
  description = "How far ahead of the renewal target OCI starts the renewal, as an ISO 8601 duration. Must be shorter than certificate_renewal_interval."
  type        = string
  default     = "P15D"
}

# Escape hatch for a certificate this module cannot produce - in practice a publicly
# trusted one, since OCI has no free public issuance and its internal CA is trusted only by
# clients that have imported it.
#
# Terraform cannot create such a certificate either, which is worth knowing before trying:
# oci_certificates_management_certificate validates IMPORTED material in CustomizeDiff with
# certificateConfigDiffStringProvided(), which reads the value via diff.GetChange() and
# rejects anything empty - and an unknown value reads as empty. So a certificate produced in
# the same apply always fails plan with "cert_chain_pem is required when config_type is
# IMPORTED". No version guards this with NewValueKnown; checked 8.15 through 8.25. Renewal
# hits the identical check, so it is not only a bootstrap problem. Such a certificate has to
# be issued and imported outside terraform, by hand, every time it expires.
variable "certificate_ocid" {
  description = "OCID of an existing certificate in OCI Certificates covering *.<this zone> and the zone apex. Setting it suppresses the CA and wildcard this module would otherwise create, and is passed through to the certificate_ocid output unchanged."
  type        = string
  default     = ""
}
