locals {
  customer_managed_key_v2_uri_segments        = var.customer_managed_key_v2 == null ? [] : split("/", var.customer_managed_key_v2.key_vault_key_uri)
  customer_managed_key_v2_identity_client_id  = var.customer_managed_key_v2 == null ? null : var.customer_managed_key_v2.user_assigned_identity == null ? null : var.customer_managed_key_v2.user_assigned_identity.client_id
  customer_managed_key_v2_key_name            = var.customer_managed_key_v2 == null ? null : local.customer_managed_key_v2_uri_segments[4]
  customer_managed_key_v2_key_uri             = var.customer_managed_key_v2 == null ? null : length(local.customer_managed_key_v2_uri_segments) == 6 ? var.customer_managed_key_v2.key_vault_key_uri : null
  customer_managed_key_v2_key_vault_uri       = var.customer_managed_key_v2 == null ? null : join("/", slice(local.customer_managed_key_v2_uri_segments, 0, 3))
  customer_managed_key_v2_key_version         = var.customer_managed_key_v2 == null ? null : length(local.customer_managed_key_v2_uri_segments) == 6 ? local.customer_managed_key_v2_uri_segments[5] : null
  customer_managed_key_v2_versionless_key_uri = var.customer_managed_key_v2 == null ? null : join("/", slice(local.customer_managed_key_v2_uri_segments, 0, 5))
}
