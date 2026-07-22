output "zone_id" {
  value = oci_dns_zone.pub.id
}

output "domain" {
  value = oci_dns_zone.pub.name
}

# No private-zone split yet (unlike aws/route53 and google/dns's create_private option) -
# this module only ever creates the one public zone, so int_domain is just an alias of
# domain for now. Kept as a separate output since modules/k8s/argocd and others reference
# .toutput.<dns-module>.int_domain by convention across all clouds.
output "int_domain" {
  value = oci_dns_zone.pub.name
}

# The NS delegation record must be added manually in the parent zone (which typically
# lives in a different cloud/provider, e.g. Route53) - these are the values to add.
output "name_servers" {
  value = oci_dns_zone.pub.nameservers[*].hostname
}
