# A private OCI zone is not attached to a VCN the way a Route53 private zone is attached to a
# VPC. It belongs to a DNS *view*, and the view that a VCN resolves against is the default view
# of the resolver associated with that VCN - two hops from the VCN OCID a deployment actually
# has. These two data sources are that walk, so a domain only has to say `private = true`.
#
# Skipped for any domain that names a view_id itself, which is how a zone goes into a view of
# its own rather than the VCN's default.
data "oci_core_vcn_dns_resolver_association" "this" {
  for_each = {
    for k, v in local.domains : k => v
    if v.private && v.view_id == ""
  }

  vcn_id = each.value.effective_vcn_id
}

data "oci_dns_resolver" "this" {
  for_each = {
    for k, v in local.domains : k => v
    if v.private && v.view_id == ""
  }

  resolver_id = data.oci_core_vcn_dns_resolver_association.this[each.key].dns_resolver_id
  scope       = "PRIVATE"
}
