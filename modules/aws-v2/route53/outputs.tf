
output "pub_zone_id" {
  description = "Zone ID of the domain marked as default public"
  value = local.effective_zone_ids[local.default_public_key]
}

output "pub_domain" {
  description = "Domain name of the zone marked as default public"
  value = local.domains_with_defaults[local.default_public_key].domain_name
}

output "pub_cert_arn" {
  description = "Certificate ARN of the domain marked as default public"
  value = length(local.default_public_key) > 0 && local.domains_with_defaults[local.default_public_key].create_certificate ? aws_acm_certificate.this[local.default_public_key].arn : null
}

output "int_zone_id" {
  description = "Zone ID of the domain marked as default private"
  value = local.effective_zone_ids[local.default_private_key]
}

output "int_domain" {
  description = "Domain name of the zone marked as default private"
  value = local.domains_with_defaults[local.default_private_key].domain_name
}

output "int_cert_arn" {
  description = "Certificate ARN of the domain marked as default private"
  value = length(local.default_private_key) > 0 && local.domains_with_defaults[local.default_private_key].create_certificate ? aws_acm_certificate.this[local.default_private_key].arn : null
}

output "zone_ids" {
  description = "Map of domain keys to their zone IDs"
  value = local.effective_zone_ids
}

output "domain_names" {
  description = "Map of domain keys to their domain names"
  value = { for k, v in var.domains : k => v.domain_name }
}

output "certificate_arns" {
  description = "Map of domain keys to their certificate ARNs (if created)"
  value = { for k, v in aws_acm_certificate.this : k => v.arn }
}

output "validation_zone_ids" {
  description = "Map of domain keys to their validation zone IDs (for private domains with certificates)"
  value = { for k, v in aws_route53_zone.validation : k => v.zone_id }
}

output "nameservers" {
  description = "Map of domain keys to their nameservers"
  value = {
    for k, v in var.domains : k => (
      v.create_zone == false ? 
      data.aws_route53_zone.existing[k].name_servers : 
      aws_route53_zone.this[k].name_servers
    )
  }
}

output "validation_nameservers" {
  description = "Map of domain keys to their validation zone nameservers (for private domains with certificates)"
  value = { for k, v in aws_route53_zone.validation : k => v.name_servers }
}

