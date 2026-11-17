output "managed_identities_azapi" {
  description = <<DESCRIPTION
The Managed Identity configuration for the azapi_resource.
Value is an object with the following attributes:

- `type` - The type of Managed Identity. Possible values are `SystemAssigned`, `UserAssigned`, or `SystemAssigned, UserAssigned`.
- `identity_ids` - A list of User Assigned Managed Identity resource IDs assigned to this resource.
DESCRIPTION
  value       = local.managed_identities
}
