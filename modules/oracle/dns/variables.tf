variable "prefix" {
  type = string
}

variable "compartment_id" {
  description = "OCID of the compartment that will contain the DNS zones and certificates."
  type        = string
}

# Module-level default for the private zones' VCN, mirroring vpc_id in
# modules/aws-v2/route53. Wired from modules/oracle/vpc by agent_input.yaml; a domain may
# override it.
variable "vcn_id" {
  description = "OCID of the VCN whose resolver holds the private zones. Only needed when a domain has private = true and names no vcn_id of its own."
  type        = string
  default     = ""
}

# Same shape as modules/aws-v2/route53's `domains`, so a deployment reads the same on both
# clouds. Differences are only where OCI has no equivalent:
#
#   - no create_validation. ACM proves domain ownership with a DNS record; OCI Certificates
#     has no validation subsystem at all - a private CA simply signs what it is asked to.
#   - vcn_id rather than vpc_id, and a view_id escape hatch. A Route53 private zone is
#     attached to the VPC directly; an OCI private zone lives in a DNS *view* held by the
#     VCN's resolver, so the view is looked up from the VCN for you.
#   - certificate_ocid, which AWS does not need: it is the only way to bring a publicly
#     trusted certificate on OCI, since terraform cannot own an imported one.
variable "domains" {
  description = "Map of domain configurations. Each key names a domain; see the README for the fields and for how default_public/default_private select which one the pub_*/int_* outputs describe."
  type = map(object({
    domain_name = string

    create_zone = optional(bool, true)
    # OCID of the parent zone, when that zone is also in OCI DNS - the NS delegation is then
    # created here. Leave empty when the parent lives elsewhere (Route53, a registrar), which
    # is the usual case: the delegation is then added by hand from the nameservers output.
    parent_zone_id = optional(string, "")

    # A private zone is resolvable only inside a VCN. Give it vcn_id (or leave it to the
    # module-level vcn_id) and the VCN's resolver's default view is looked up; view_id
    # overrides that lookup when a zone belongs in a view of its own.
    private = optional(bool, false)
    vcn_id  = optional(string, "")
    view_id = optional(string, "")

    create_certificate = optional(bool, true)
    # Defaults to the module-level certificate_authority_id, which is wired from
    # modules/oracle/pca. Set per domain to issue from a different authority.
    certificate_authority_id = optional(string, "")
    # An existing certificate to use instead of issuing one - in practice a publicly trusted
    # one. Suppresses issuance for this domain and is passed through to the outputs unchanged.
    certificate_ocid = optional(string, "")

    default_public  = optional(bool, null)
    default_private = optional(bool, null)
  }))
  default = {}
}

# The OCI analogue of certificate_authority_arn in modules/aws-v2/route53, but module-level
# rather than per-domain: it is wired from modules/oracle/pca by agent_input.yaml, and a
# per-domain default would have to be templated inside the `domains` value itself. Domains
# that want a different authority override it there.
variable "certificate_authority_id" {
  description = "OCID of the certificate authority that issues certificates for domains which do not name one themselves. Wired from modules/oracle/pca, or set to a CA held elsewhere."
  type        = string
  default     = ""
}

variable "name_salt" {
  description = "Bump this to rotate the random suffix on certificate names. Needed when a create fails after OCI accepted it: the orphan keeps the name, and no retry can reuse it."
  type        = string
  default     = "1"
}

# Left at OCI's default certificate validity (three months) rather than pinned here: an
# explicit validity is an absolute timestamp, which would have to be regenerated on every
# renewal. These two rules are what keep certificates fresh inside that window, and they
# apply to every domain - a per-domain renewal schedule has no use case yet.
variable "certificate_renewal_interval" {
  description = "How often OCI renews the certificates, as an ISO 8601 duration."
  type        = string
  default     = "P60D"
}

variable "certificate_advance_renewal_period" {
  description = "How far ahead of the renewal target OCI starts the renewal, as an ISO 8601 duration. Must be shorter than certificate_renewal_interval."
  type        = string
  default     = "P15D"
}
