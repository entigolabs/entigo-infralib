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

# OCID of the zone-wide wildcard certificate in OCI Certificates (certificate.tf), or an
# empty string when create_certificate is off. Consumed by the Oracle apps as the
# oci-native-ingress.oraclecloud.com/certificate-ocid annotation - every app on the cluster
# uses this same value, because they all share one load balancer listener.
output "certificate_ocid" {
  value = var.create_certificate ? oci_certificates_management_certificate.wildcard[0].id : ""
}

# The NS delegation record must be added manually in the parent zone (which typically
# lives in a different cloud/provider, e.g. Route53) - these are the values to add.
output "name_servers" {
  value = oci_dns_zone.pub.nameservers[*].hostname
}
