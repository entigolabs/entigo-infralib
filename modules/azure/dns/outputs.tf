output "pub_zone_id" {
  description = "Zone ID of the domain marked as default public"
  value       = try(local.zone_ids[local.default_public_keys[0]], null)

  # Also covers create_zone = false only and an empty map, like the aws-v2/route53 index lookup
  precondition {
    condition     = length(local.default_public_keys) == 1 && length(local.default_private_keys) == 1
    error_message = "Exactly one domain must be default_public = true and exactly one default_private = true."
  }
}

output "pub_domain" {
  description = "Domain name of the zone marked as default public"
  value       = try(local.domains[local.default_public_keys[0]].domain_name, null)
}

output "int_zone_id" {
  description = "Zone ID of the domain marked as default private"
  value       = try(local.zone_ids[local.default_private_keys[0]], null)
}

output "int_domain" {
  description = "Domain name of the zone marked as default private"
  value       = try(local.domains[local.default_private_keys[0]].domain_name, null)
}

output "int_private_domain" {
  description = "Default private domain when it is a private zone, empty otherwise (e.g. a single public domain)"
  value       = try(local.domains[local.default_private_keys[0]].private ? local.domains[local.default_private_keys[0]].domain_name : "", "")
}

output "int_cert_zone_id" {
  description = "Validation zone ID of the default private domain (ACME DNS-01), empty when it has none"
  value       = try(azurerm_dns_zone.validation[local.default_private_keys[0]].id, "")
}

output "int_cert_zone_name" {
  description = "Validation zone name of the default private domain, empty when it has none"
  value       = try(azurerm_dns_zone.validation[local.default_private_keys[0]].name, "")
}

output "zone_ids" {
  description = "Map of domain keys to their zone IDs"
  value       = local.zone_ids
}

output "domain_names" {
  description = "Map of domain keys to their domain names"
  value       = { for k, v in var.domains : k => v.domain_name }
}

output "validation_zone_ids" {
  description = "Map of domain keys to their validation zone IDs (private domains)"
  value       = { for k, v in azurerm_dns_zone.validation : k => v.id }
}

output "nameservers" {
  description = "Map of public domain keys to their name servers, for delegation from the parent zone"
  value = {
    for k, v in local.domains : k => (v.create_zone ? azurerm_dns_zone.this[k].name_servers : data.azurerm_dns_zone.existing[k].name_servers)
    if !v.private
  }
}

output "validation_nameservers" {
  description = "Map of domain keys to their validation zone name servers, for delegation from the parent zone"
  value       = { for k, v in azurerm_dns_zone.validation : k => v.name_servers }
}
