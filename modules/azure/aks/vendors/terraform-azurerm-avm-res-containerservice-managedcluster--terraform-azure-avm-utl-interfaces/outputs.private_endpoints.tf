output "private_dns_zone_groups_azapi" {
  description = <<DESCRIPTION
A map of private DNS zone groups for use with azapi_resource, the value is an object containing the following attributes:

- `type` - The type of the resource.
- `name` - The name of the resource.
- `body` - The body of the resource.
DESCRIPTION
  value       = local.private_dns_zone_groups
}

output "private_endpoints_azapi" {
  description = <<DESCRIPTION
A map of private endpoints for use with azapi_resource, the value is an object containing the following attributes:

- `type` - The type of the resource.
- `name` - The name of the resource.
- `body` - The body of the resource.
- `tags` - The tags of the resource.
DESCRIPTION
  value       = local.private_endpoints
}
