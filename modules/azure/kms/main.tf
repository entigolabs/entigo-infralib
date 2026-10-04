data "azurerm_client_config" "this" {}

resource "random_string" "suffix" {
  length  = 4
  upper   = false
  special = false
}

locals {
  # Globally unique and reserved while soft-deleted
  name = "${replace(substr(var.prefix, 0, 19), "/-+$/", "")}-${random_string.suffix.result}"
  keys = toset(["data", "config", "telemetry"])

  users = {
    data      = var.data_key_users
    config    = var.config_key_users
    telemetry = var.telemetry_key_users
  }

  # data: disks and other data, telemetry: monitoring data (wrap/unwrap by Azure services),
  # config: customer app secrets (apps encrypt/decrypt with it)
  key_user_roles = {
    data      = "Key Vault Crypto Service Encryption User"
    config    = "Key Vault Crypto User"
    telemetry = "Key Vault Crypto Service Encryption User"
  }

  key_users = merge([
    for key, users in local.users : {
      for user, principal_id in users : "${key}/${user}" => { key = key, principal_id = principal_id }
    }
  ]...)

  tags = merge(var.tags, {
    Terraform  = "true"
    Prefix     = var.prefix
    created-by = "entigo-infralib"
  })
}

resource "azurerm_key_vault" "this" {
  name                       = local.name
  location                   = var.location
  resource_group_name        = var.resource_group_name
  tenant_id                  = data.azurerm_client_config.this.tenant_id
  sku_name                   = var.sku_name
  rbac_authorization_enabled = true
  purge_protection_enabled   = var.purge_protection_enabled
  soft_delete_retention_days = var.soft_delete_retention_days
  tags                       = local.tags
}

resource "azurerm_management_lock" "this" {
  count      = var.delete_lock_enabled ? 1 : 0
  name       = "${var.prefix}-delete-lock"
  scope      = azurerm_key_vault.this.id
  lock_level = "CanNotDelete"
  notes      = "Customer managed keys: deleting the vault makes encrypted data unrecoverable"
}

# The installing identity (normally the agent job identity) keeps the role, later callers don't replace it
resource "terraform_data" "installer" {
  input = data.azurerm_client_config.this.object_id

  lifecycle {
    ignore_changes = [input]
  }
}

resource "azurerm_role_assignment" "admin" {
  for_each             = merge({ installer = terraform_data.installer.input }, { for id in toset([for i in var.admin_object_ids : lower(i)]) : id => id })
  scope                = azurerm_key_vault.this.id
  role_definition_name = "Key Vault Crypto Officer"
  principal_id         = each.value

  lifecycle {
    precondition {
      condition     = each.key == "installer" || each.value != lower(terraform_data.installer.input)
      error_message = "admin_object_ids must not list the installing identity ${terraform_data.installer.input}, it has the role already."
    }
  }
}

# Data plane role assignments take up to a minute
resource "time_sleep" "rbac" {
  create_duration = "60s"
  depends_on      = [azurerm_role_assignment.admin]
}

resource "azurerm_key_vault_key" "this" {
  for_each     = local.keys
  name         = "${var.prefix}-${each.key}"
  key_vault_id = azurerm_key_vault.this.id
  key_type     = var.key_type
  key_size     = var.key_size
  key_opts     = ["decrypt", "encrypt", "unwrapKey", "wrapKey"]
  tags         = local.tags

  dynamic "rotation_policy" {
    for_each = var.key_rotation_period == null ? [] : [var.key_rotation_period]
    content {
      automatic {
        time_after_creation = rotation_policy.value
      }
    }
  }

  depends_on = [time_sleep.rbac]
}

resource "azurerm_role_assignment" "key_user" {
  for_each             = local.key_users
  scope                = azurerm_key_vault_key.this[each.value.key].resource_versionless_id
  role_definition_name = local.key_user_roles[each.value.key]
  principal_id         = each.value.principal_id
}
