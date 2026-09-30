data "azurerm_key_vault" "credentials" {
  count               = var.key_vault_id != "" ? 1 : 0
  name                = provider::azurerm::parse_resource_id(var.key_vault_id).resource_name
  resource_group_name = provider::azurerm::parse_resource_id(var.key_vault_id).resource_group_name
}

resource "random_string" "suffix" {
  length  = 4
  upper   = false
  special = false
}

locals {
  # Globally unique, alphanumerics, max 50
  name = "${substr(replace(var.prefix, "/[^a-zA-Z0-9]/", ""), 0, 46)}${random_string.suffix.result}"

  registries = {
    hub  = "docker.io"
    ghcr = "ghcr.io"
    gcr  = "gcr.io"
    ecr  = "public.ecr.aws"
    quay = "quay.io"
    mcr  = "mcr.microsoft.com"
    k8s  = "registry.k8s.io"
  }

  credentials = {
    hub  = { username = var.hub_username_secret, token = var.hub_access_token_secret }
    ghcr = { username = var.ghcr_username_secret, token = var.ghcr_access_token_secret }
    gcr  = { username = var.gcr_username_secret, token = var.gcr_access_token_secret }
    quay = { username = var.quay_username_secret, token = var.quay_access_token_secret }
  }
  registries_with_credentials = toset([for k, c in local.credentials : k if c.username != "" && c.token != ""])
  # Azure rejects Docker Hub cache rules without credentials
  cache_rules = { for k, v in local.registries : k => v if k != "hub" || contains(local.registries_with_credentials, "hub") }

  private_endpoint = var.private_endpoint_enabled

  tags = merge(var.tags, {
    Terraform  = "true"
    Prefix     = var.prefix
    created-by = "entigo-infralib"
  })
}

resource "azurerm_container_registry" "this" {
  name                     = local.name
  location                 = var.location
  resource_group_name      = var.resource_group_name
  sku                      = var.sku
  retention_policy_in_days = var.sku == "Premium" ? var.retention_policy_in_days : null
  # Azure enables data endpoints with a private endpoint
  data_endpoint_enabled         = local.private_endpoint
  public_network_access_enabled = var.public_network_access_enabled
  tags                          = local.tags

  identity {
    type         = var.encryption == null ? "SystemAssigned" : "SystemAssigned, UserAssigned"
    identity_ids = var.encryption == null ? null : [var.encryption.identity_id]
  }

  dynamic "encryption" {
    for_each = var.encryption == null ? [] : [var.encryption]
    content {
      key_vault_key_id   = encryption.value.key_vault_key_id
      identity_client_id = encryption.value.identity_client_id
    }
  }

  lifecycle {
    precondition {
      condition     = var.encryption == null || var.sku == "Premium"
      error_message = "ACR customer managed keys (encryption) require sku = \"Premium\"."
    }
    precondition {
      condition     = (!local.private_endpoint && var.public_network_access_enabled) || var.sku == "Premium"
      error_message = "ACR private endpoint and disabled public network access require sku = \"Premium\"."
    }
    precondition {
      condition     = !local.private_endpoint || (var.private_endpoint_subnet_id != null && var.private_endpoint_vnet_id != null)
      error_message = "private_endpoint_enabled needs private_endpoint_subnet_id and private_endpoint_vnet_id."
    }
  }
}

resource "azurerm_container_registry_credential_set" "this" {
  for_each              = local.registries_with_credentials
  name                  = "acr-proxy-${each.value}"
  container_registry_id = azurerm_container_registry.this.id
  login_server          = local.registries[each.value]

  identity {
    type = "SystemAssigned"
  }

  authentication_credentials {
    username_secret_id = "${data.azurerm_key_vault.credentials[0].vault_uri}secrets/${local.credentials[each.value].username}"
    password_secret_id = "${data.azurerm_key_vault.credentials[0].vault_uri}secrets/${local.credentials[each.value].token}"
  }

  lifecycle {
    precondition {
      condition     = var.key_vault_id != ""
      error_message = "Registry credential secrets need key_vault_id."
    }
  }
}

resource "azurerm_role_assignment" "credential_secrets" {
  for_each = { for p in flatten([
    for k in local.registries_with_credentials : [
      for f in ["username", "token"] : { key = "${k}-${f}", registry = k, secret = local.credentials[k][f] }
    ]
  ]) : p.key => p }
  scope                            = "${var.key_vault_id}/secrets/${each.value.secret}"
  role_definition_name             = "Key Vault Secrets User"
  principal_id                     = azurerm_container_registry_credential_set.this[each.value.registry].identity[0].principal_id
  skip_service_principal_aad_check = true
}

# AKS bootstrap artifact cache: name and repos are fixed by AKS
resource "azurerm_container_registry_cache_rule" "aks_bootstrap" {
  count                 = var.aks_bootstrap_cache_rule ? 1 : 0
  name                  = "aks-managed-mcr"
  container_registry_id = azurerm_container_registry.this.id
  source_repo           = "mcr.microsoft.com/*"
  target_repo           = "aks-managed-repository/*"
}

# <acr>.azurecr.io/<source>/<repo>
resource "azurerm_container_registry_cache_rule" "this" {
  for_each              = local.cache_rules
  name                  = "acr-proxy-${each.key}"
  container_registry_id = azurerm_container_registry.this.id
  source_repo           = "${each.value}/*"
  target_repo           = "${each.value}/*"
  credential_set_id     = contains(local.registries_with_credentials, each.key) ? azurerm_container_registry_credential_set.this[each.key].id : null
}

resource "azurerm_private_dns_zone" "acr" {
  count               = local.private_endpoint ? 1 : 0
  name                = "privatelink.azurecr.io"
  resource_group_name = var.resource_group_name
  tags                = local.tags
}

resource "azurerm_private_dns_zone_virtual_network_link" "acr" {
  count                = local.private_endpoint ? 1 : 0
  name                 = "${var.prefix}-acr"
  private_dns_zone_id  = azurerm_private_dns_zone.acr[0].id
  virtual_network_id   = var.private_endpoint_vnet_id
  registration_enabled = false
  tags                 = local.tags
}

resource "azurerm_private_endpoint" "acr" {
  count               = local.private_endpoint ? 1 : 0
  name                = "${var.prefix}-acr"
  location            = var.location
  resource_group_name = var.resource_group_name
  subnet_id           = var.private_endpoint_subnet_id
  tags                = local.tags

  private_service_connection {
    name                           = "${var.prefix}-acr"
    private_connection_resource_id = azurerm_container_registry.this.id
    subresource_names              = ["registry"]
    is_manual_connection           = false
  }

  private_dns_zone_group {
    name                 = "acr"
    private_dns_zone_ids = [azurerm_private_dns_zone.acr[0].id]
  }
}

resource "azurerm_container_registry_task" "purge" {
  count                 = var.purge_enabled ? 1 : 0
  name                  = "acr-proxy-purge"
  container_registry_id = azurerm_container_registry.this.id
  tags                  = local.tags

  platform {
    os = "Linux"
  }

  encoded_step {
    task_content = base64encode(<<-EOT
      version: v1.1.0
      steps:
        - cmd: acr purge --untagged-only --ago ${var.purge_untagged_days}d
          disableWorkingDirectoryOverride: true
          timeout: 3600
        - cmd: acr purge --filter '.*:.*' --ago ${var.purge_tagged_days}d --untagged
          disableWorkingDirectoryOverride: true
          timeout: 3600
    EOT
    )
  }

  timer_trigger {
    name     = "daily"
    schedule = var.purge_schedule
    enabled  = true
  }
}
