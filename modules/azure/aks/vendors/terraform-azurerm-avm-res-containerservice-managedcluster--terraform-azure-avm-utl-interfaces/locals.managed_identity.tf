locals {
  managed_identities = local.managed_identities_system_assigned != null || local.managed_identities_user_assigned != null ? {
    type         = local.managed_identities_type
    identity_ids = var.managed_identities.user_assigned_resource_ids
  } : null
  managed_identities_system_assigned = var.managed_identities.system_assigned ? "SystemAssigned" : null
  managed_identities_type            = length(local.managed_identities_type_list) > 0 ? join(", ", local.managed_identities_type_list) : null
  managed_identities_type_list       = compact([local.managed_identities_system_assigned, local.managed_identities_user_assigned])
  managed_identities_user_assigned   = length(var.managed_identities.user_assigned_resource_ids) > 0 ? "UserAssigned" : null
}
