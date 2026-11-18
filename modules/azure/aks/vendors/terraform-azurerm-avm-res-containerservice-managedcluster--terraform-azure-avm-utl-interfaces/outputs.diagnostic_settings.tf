output "diagnostic_settings_azapi" {
  description = <<DESCRIPTION
A map of diagnostic settings for use in azapi_resource, the value is an object containing the following attributes:

- `type` - The type of the resource.
- `name` - The name of the resource.
- `body` - The body of the resource.
DESCRIPTION
  value       = local.diagnostic_settings_azapi
}
