# This resource allows us to look up role definitions by role name.
# The AzureRM provider does this already so we are replicating this functionality here to benefit AzAPI users.
data "azapi_resource_list" "role_definitions" {
  count = local.role_assignment_definition_lookup_enabled ? 1 : 0

  parent_id = var.role_assignment_definition_scope
  type      = "Microsoft.Authorization/roleDefinitions@2022-04-01"
  response_export_values = {
    results = "value[].{id: id, role_name: properties.roleName}"
  }
}

# Generate a random UUID for the role assignment name unless one is supplied via the
# `name` attribute on the role assignment. Using a random UUID avoids issues with
# duplicate UUIDs and unknown values causing replacement (e.g. principal ID).
resource "random_uuid" "role_assignment_name" {
  for_each = var.role_assignments
}

# Generate a random UUID for the private endpoint role assignment name unless one is
# supplied via the `name` attribute on the role assignment.
resource "random_uuid" "role_assignment_name_private_endpoint" {
  for_each = local.role_assignments_private_endpoint_azapi_keys_only
}
