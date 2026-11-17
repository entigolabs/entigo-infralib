output "lock_azapi" {
  description = <<DESCRIPTION
An object for use in azapi_resource with the following attributes:

- `type` - The type of the resource.
- `name` - The name of the resource.
- `body` - The body of the resource.
DESCRIPTION
  value       = local.lock_azapi
}

output "lock_private_endpoint_azapi" {
  description = <<DESCRIPTION
A flattened map of objects containing for use in azapi_resource with the following attributes:

- `pe_key` - The key of the private endpoint, used to look up the parent id.
- `type` - The type of the resource.
- `name` - The name of the resource.
- `body` - The body of the resource.

These locks should be used for private endpoints defined in var.private_endpoints.
DESCRIPTION
  value       = local.lock_private_endpoint_azapi
}
