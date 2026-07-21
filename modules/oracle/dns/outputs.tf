output "zone_id" {
  value = oci_dns_zone.pub.id
}

output "domain" {
  value = oci_dns_zone.pub.name
}

# The NS delegation record must be added manually in the parent zone (which typically
# lives in a different cloud/provider, e.g. Route53) - these are the values to add.
output "name_servers" {
  value = oci_dns_zone.pub.nameservers[*].hostname
}
