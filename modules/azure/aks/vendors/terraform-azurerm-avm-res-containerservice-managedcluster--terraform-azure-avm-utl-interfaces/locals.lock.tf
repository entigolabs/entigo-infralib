locals {
  lock_azapi = var.lock != null ? {
    type = local.lock_type
    name = lookup(var.lock, "name", null)
    body = {
      properties = {
        level = var.lock.kind
        notes = var.lock.notes
      }
    }
  } : null
  # Create a map of private endpoint locks.
  # There can only be one lock per private endpoint,
  # so we use the private endpoint key as the key in the map.
  lock_private_endpoint_azapi = {
    for pe_key, pe_val in var.private_endpoints : pe_key => {
      pe_key = pe_key
      type   = local.lock_type
      name   = lookup(pe_val.lock, "name", null)
      body = {
        properties = {
          level = pe_val.lock.kind
          notes = pe_val.lock.notes
        }
      }
    } if pe_val.lock != null
  }
  lock_type = "Microsoft.Authorization/locks@2020-05-01"
}
