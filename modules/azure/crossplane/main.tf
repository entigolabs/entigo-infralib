data "azurerm_client_config" "this" {}

data "azurerm_resource_group" "this" {
  name = var.resource_group_name
}

locals {
  # Owner, User Access Administrator, Role Based Access Control Administrator
  privileged_roles = "8e3af657-a8ff-443c-a75c-2fe8c4bcb635, 18d7d88d-d35e-4fb5-a5c3-7773c20a72d9, f58310d9-a9f6-439a-9e8d-f62e7b41a168"
  # Monitoring Reader
  node_resource_group_roles = "43d0d8ad-25c7-4714-9337-8ba259a9fe05"

  tags = merge(var.tags, {
    Terraform  = "true"
    Prefix     = var.prefix
    created-by = "entigo-infralib"
  })
}

# Crossplane Azure providers
resource "azurerm_user_assigned_identity" "crossplane" {
  name                = var.prefix
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = local.tags
}

resource "azurerm_federated_identity_credential" "crossplane" {
  name                      = "crossplane"
  user_assigned_identity_id = azurerm_user_assigned_identity.crossplane.id
  audience                  = ["api://AzureADTokenExchange"]
  issuer                    = var.oidc_issuer_url
  subject                   = "system:serviceaccount:${var.kubernetes_namespace}:${var.kubernetes_service_account}"
}

resource "azurerm_role_assignment" "contributor" {
  scope                            = data.azurerm_resource_group.this.id
  role_definition_name             = "Contributor"
  principal_id                     = azurerm_user_assigned_identity.crossplane.principal_id
  skip_service_principal_aad_check = true
}

# Any role except the privileged ones
resource "azurerm_role_assignment" "rbac_administrator" {
  scope                            = data.azurerm_resource_group.this.id
  role_definition_name             = "Role Based Access Control Administrator"
  principal_id                     = azurerm_user_assigned_identity.crossplane.principal_id
  skip_service_principal_aad_check = true
  condition_version                = "2.0"
  condition                        = "((!(ActionMatches{'Microsoft.Authorization/roleAssignments/write'})) OR (@Request[Microsoft.Authorization/roleAssignments:RoleDefinitionId] ForAnyOfAllValues:GuidNotEquals {${local.privileged_roles}})) AND ((!(ActionMatches{'Microsoft.Authorization/roleAssignments/delete'})) OR (@Resource[Microsoft.Authorization/roleAssignments:RoleDefinitionId] ForAnyOfAllValues:GuidNotEquals {${local.privileged_roles}}))"
}

# AKS node resource group (AGC, load balancers, disks, scale sets): only read access for monitoring
resource "azurerm_role_assignment" "node_rbac_administrator" {
  scope                            = var.node_resource_group_id
  role_definition_name             = "Role Based Access Control Administrator"
  principal_id                     = azurerm_user_assigned_identity.crossplane.principal_id
  skip_service_principal_aad_check = true
  condition_version                = "2.0"
  condition                        = "((!(ActionMatches{'Microsoft.Authorization/roleAssignments/write'})) OR (@Request[Microsoft.Authorization/roleAssignments:RoleDefinitionId] ForAnyOfAnyValues:GuidEquals {${local.node_resource_group_roles}})) AND ((!(ActionMatches{'Microsoft.Authorization/roleAssignments/delete'})) OR (@Resource[Microsoft.Authorization/roleAssignments:RoleDefinitionId] ForAnyOfAnyValues:GuidEquals {${local.node_resource_group_roles}}))"
}

# Crossplane core, pulls packages from acr-proxy
resource "azurerm_user_assigned_identity" "core" {
  name                = "${var.prefix}-core"
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = local.tags
}

resource "azurerm_federated_identity_credential" "core" {
  name                      = "crossplane-core"
  user_assigned_identity_id = azurerm_user_assigned_identity.core.id
  audience                  = ["api://AzureADTokenExchange"]
  issuer                    = var.oidc_issuer_url
  subject                   = "system:serviceaccount:${var.kubernetes_namespace}:${var.kubernetes_core_service_account}"
}

resource "azurerm_role_assignment" "core_acr_pull" {
  count                            = var.acr_id != "" ? 1 : 0
  scope                            = var.acr_id
  role_definition_name             = "AcrPull"
  principal_id                     = azurerm_user_assigned_identity.core.principal_id
  skip_service_principal_aad_check = true
}

# Storage account of the infralib agent (state + loki/mimir containers), no agent tag for its name
data "azurerm_resources" "agent_storage" {
  resource_group_name = var.resource_group_name
  type                = "Microsoft.Storage/storageAccounts"
  required_tags = {
    created-by = "entigo-infralib-agent"
  }
}

# The storage account's own identity (agent v1.15+: <prefix>-storage) reads the CMK keys, also for loki/mimir encryption scopes
data "azurerm_storage_account" "agent" {
  count               = var.telemetry_key_resource_id != "" ? 1 : 0
  name                = data.azurerm_resources.agent_storage.resources[0].name
  resource_group_name = var.resource_group_name
}

data "azurerm_user_assigned_identity" "agent" {
  count               = var.telemetry_key_resource_id != "" ? 1 : 0
  name                = provider::azurerm::parse_resource_id(data.azurerm_storage_account.agent[0].identity[0].identity_ids[0]).resource_name
  resource_group_name = provider::azurerm::parse_resource_id(data.azurerm_storage_account.agent[0].identity[0].identity_ids[0]).resource_group_name
}

resource "azurerm_role_assignment" "agent_telemetry_key" {
  count                            = var.telemetry_key_resource_id != "" ? 1 : 0
  scope                            = var.telemetry_key_resource_id
  role_definition_name             = "Key Vault Crypto Service Encryption User"
  principal_id                     = data.azurerm_user_assigned_identity.agent[0].principal_id
  skip_service_principal_aad_check = true
}
