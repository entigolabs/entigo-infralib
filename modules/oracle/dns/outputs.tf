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

# Consumed by the Oracle apps as the oci-native-ingress.oraclecloud.com/certificate-ocid
# annotation - every app on the cluster uses this same value, because they all share one
# load balancer listener and a listener holds one key pair. Either the wildcard this module
# issued from its own CA, or whatever was passed in as certificate_ocid.
output "certificate_ocid" {
  value = var.certificate_ocid != "" ? var.certificate_ocid : (local.create_cert ? oci_certificates_management_certificate.wildcard[0].id : "")
}

# The CA that signed it. Clients trust the wildcard only after importing this CA's
# certificate, which terraform cannot output: the PEM comes from the Certificates *data
# plane* API, and the OCI provider only wraps certificates_management. Fetch it with
#   oci certificates certificate-authority-bundle get \
#     --certificate-authority-id <this> --query 'data."certificate-pem"' --raw-output
output "certificate_authority_id" {
  value = local.create_cert ? oci_certificates_management_certificate_authority.this[0].id : ""
}

# The NS delegation record must be added manually in the parent zone (which typically
# lives in a different cloud/provider, e.g. Route53) - these are the values to add.
output "name_servers" {
  value = oci_dns_zone.pub.nameservers[*].hostname
}
