output "id" {
  value = oci_identity_domain.this.id
}

output "idcs_endpoint" {
  value = oci_identity_domain.this.url
}

output "display_name" {
  value = oci_identity_domain.this.display_name
}
