# The identifier every consumer needs. modules/oracle/dns takes it as certificate_authority_id
# and issues its wildcard from it, the way modules/aws-v2/route53 takes certificate_authority_arn.
#
# Empty rather than null when no CA is created: the agent renders this into a string input.
output "certificate_authority_id" {
  value = var.create_ca ? oci_certificates_management_certificate_authority.this[0].id : ""
}

output "certificate_authority_name" {
  value = var.create_ca ? oci_certificates_management_certificate_authority.this[0].name : ""
}

# Clients trust certificates from this CA only after importing the CA's own certificate, and
# terraform cannot output the PEM: it comes from the Certificates *data plane* API and the OCI
# provider only wraps certificates_management. Fetch it with the CLI:
#
#   oci certificates certificate-authority-bundle get \
#     --certificate-authority-id "$(terraform output -raw certificate_authority_id)" \
#     --query 'data."certificate-pem"' --raw-output > ca.pem
output "certificate_authority_bundle_command" {
  description = "Ready-made CLI command that exports this CA's PEM, which terraform cannot read itself."
  value       = var.create_ca ? "oci certificates certificate-authority-bundle get --certificate-authority-id ${oci_certificates_management_certificate_authority.this[0].id} --query 'data.\"certificate-pem\"' --raw-output" : ""
}
