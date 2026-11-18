variable "role_assignment_definition_lookup_enabled" {
  type        = bool
  default     = true
  description = <<DESCRIPTION
A control to disable the lookup of role definitions when creating role assignments.
If you disable this then all role assignments must be supplied with a `role_definition_id_or_name` that is a valid role definition ID.
DESCRIPTION
}

variable "role_assignment_definition_scope" {
  type        = string
  default     = null
  description = <<DESCRIPTION
The scope at which the role assignments should be created. Used to look up role definitions by role name.

Must be specified when `role_assignments` are defined.
DESCRIPTION
}

variable "role_assignments" {
  type = map(object({
    name                                   = optional(string, null)
    role_definition_id_or_name             = string
    principal_id                           = string
    description                            = optional(string, null)
    skip_service_principal_aad_check       = optional(bool, false)
    condition                              = optional(string, null)
    condition_version                      = optional(string, null)
    delegated_managed_identity_resource_id = optional(string, null)
    principal_type                         = optional(string, null)
  }))
  default     = {}
  description = <<DESCRIPTION
  A map of role assignments to create. The map key is deliberately arbitrary to avoid issues where map keys maybe unknown at plan time.
  Do not change principal or role definition values in this map after the initial creation of the role assignments as this will cause errors.
  Instead, add a new entry to the map with a new key and remove the old entry.

  - `name` - (Optional) The name of the role assignment. If not set, a random UUID will be generated. Changing this forces the creation of a new resource.
  - `role_definition_id_or_name` - The ID or name of the role definition to assign to the principal.
  - `principal_id` - The ID of the principal to assign the role to.
  - `description` - (Optional) The description of the role assignment.
  - `skip_service_principal_aad_check` - (Optional) No effect when using AzAPI.
  - `condition` - (Optional) The condition which will be used to scope the role assignment.
  - `condition_version` - (Optional) The version of the condition syntax. Leave as `null` if you are not using a condition, if you are then valid values are '2.0'.
  - `delegated_managed_identity_resource_id` - (Optional) The delegated Azure Resource Id which contains a Managed Identity. Changing this forces a new resource to be created. This field is only used in cross-tenant scenario.
  - `principal_type` - (Optional) The type of the `principal_id`. Possible values are `User`, `Group` and `ServicePrincipal`. It is necessary to explicitly set this attribute when creating role assignments if the principal creating the assignment is constrained by ABAC rules that filters on the PrincipalType attribute.
DESCRIPTION
  nullable    = false

  validation {
    error_message = "If role_assignments are specified and role_assignment_definition_lookup_enabled is true, then role_assignment_definition_scope must be set."
    condition     = length(var.role_assignments) > 0 && var.role_assignment_definition_lookup_enabled ? var.role_assignment_definition_scope != null : true
  }
  validation {
    error_message = "Each role_assignments `name`, when supplied, must be a valid lowercase GUID (e.g. 11111111-1111-1111-1111-111111111111)."
    condition = alltrue([
      for ra in var.role_assignments :
      ra.name == null || can(regex("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", ra.name))
    ])
  }
}
