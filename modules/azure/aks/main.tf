data "azurerm_client_config" "this" {}

data "azurerm_resource_group" "this" {
  name = var.resource_group_name
}

data "azapi_resource_list" "vm_skus" {
  type                   = "Microsoft.Compute/skus@2021-07-01"
  parent_id              = "/subscriptions/${data.azurerm_client_config.this.subscription_id}"
  query_parameters       = { "$filter" = ["location eq '${var.location}'"] }
  response_export_values = { vms = "value[?resourceType=='virtualMachines'].{name: name, zones: locationInfo[0].zones}" }
}

locals {
  # AKS adds it to spot pools, declared to avoid drift
  spot_taint = "kubernetes.azure.com/scalesetpriority=spot:NoSchedule"

  disk_encryption = var.disk_encryption_key_id != ""

  sku_zones = { for v in data.azapi_resource_list.vm_skus.output.vms : v.name => sort(coalesce(v.zones, [])) }
  pool_zones = {
    for pool, size in { main = var.aks_main_instance_type, mon = var.aks_mon_instance_type, tools = var.aks_tools_instance_type } :
    pool => var.availability_zones != null ? var.availability_zones : lookup(local.sku_zones, size, [])
  }

  maintenance_window = {
    duration_hours = var.maintenance_window.duration_hours
    start_time     = var.maintenance_window.start_time
    utc_offset     = var.maintenance_window.utc_offset
    schedule       = { weekly = { day_of_week = var.maintenance_window.day_of_week, interval_weeks = 1 } }
  }
  # AKS reads the schedules by these names
  maintenance_configurations = merge(
    var.upgrade_channel != "none" ? { upgrade = { name = "aksManagedAutoUpgradeSchedule", maintenance_window = local.maintenance_window } } : {},
    contains(["NodeImage", "SecurityPatch"], var.node_os_upgrade_channel) ? { node_os = { name = "aksManagedNodeOSUpgradeSchedule", maintenance_window = local.maintenance_window } } : {},
  )

  tags = merge(var.tags, {
    Terraform  = "true"
    Prefix     = var.prefix
    created-by = "entigo-infralib"
  })

  # tools is the AKS system pool: CriticalAddonsOnly taint instead of tools=true, AKS system pods only tolerate that
  pools = {
    main = {
      name                      = "main"
      vm_size                   = var.aks_main_instance_type
      enable_auto_scaling       = true
      min_count                 = var.aks_main_min_size
      max_count                 = var.aks_main_max_size
      max_pods                  = var.aks_main_max_pods
      os_disk_size_gb           = var.aks_main_volume_size
      os_disk_type              = var.aks_main_volume_type
      os_sku                    = var.os_sku
      availability_zones        = local.pool_zones["main"]
      node_labels               = { "main" = "true", created-by = "entigo-infralib" }
      node_taints               = var.aks_main_spot_nodes ? [local.spot_taint] : null
      scale_set_priority        = var.aks_main_spot_nodes ? "Spot" : null
      scale_set_eviction_policy = var.aks_main_spot_nodes ? "Delete" : null
      spot_max_price            = var.aks_main_spot_nodes ? -1 : null
      upgrade_settings          = var.aks_main_spot_nodes ? null : { max_surge = var.max_surge }
      orchestrator_version      = var.kubernetes_version
      vnet_subnet_id            = var.vnet_subnet_id
      tags                      = local.tags
    }
    mon = {
      name                      = "mon"
      vm_size                   = var.aks_mon_instance_type
      enable_auto_scaling       = true
      min_count                 = var.aks_mon_min_size
      max_count                 = var.aks_mon_max_size
      max_pods                  = var.aks_mon_max_pods
      os_disk_size_gb           = var.aks_mon_volume_size
      os_disk_type              = var.aks_mon_volume_type
      os_sku                    = var.os_sku
      availability_zones        = local.pool_zones["mon"]
      node_labels               = { "mon" = "true", created-by = "entigo-infralib" }
      node_taints               = concat(["mon=true:NoSchedule"], var.aks_mon_spot_nodes ? [local.spot_taint] : [])
      scale_set_priority        = var.aks_mon_spot_nodes ? "Spot" : "Regular"
      scale_set_eviction_policy = var.aks_mon_spot_nodes ? "Delete" : null
      spot_max_price            = var.aks_mon_spot_nodes ? -1 : null
      upgrade_settings          = var.aks_mon_spot_nodes ? null : { max_surge = var.max_surge }
      orchestrator_version      = var.kubernetes_version
      vnet_subnet_id            = var.vnet_subnet_id
      tags                      = local.tags
    }
    tools = {
      name                 = "tools"
      vm_size              = var.aks_tools_instance_type
      enable_auto_scaling  = true
      min_count            = var.aks_tools_min_size
      max_count            = var.aks_tools_max_size
      max_pods             = var.aks_tools_max_pods
      os_disk_size_gb      = var.aks_tools_volume_size
      os_disk_type         = var.aks_tools_volume_type
      os_sku               = var.os_sku
      availability_zones   = local.pool_zones["tools"]
      node_labels          = { "tools" = "true", created-by = "entigo-infralib" }
      node_taints          = ["CriticalAddonsOnly=true:NoSchedule"]
      upgrade_settings     = { max_surge = var.max_surge }
      orchestrator_version = var.kubernetes_version
      vnet_subnet_id       = var.vnet_subnet_id
      tags                 = local.tags
    }
  }

  default_agent_pool = merge(local.pools["tools"], { mode = "System", type = "VirtualMachineScaleSets" })
  agent_pools        = { for k, v in local.pools : k => merge(v, { mode = "User" }) if k != "tools" && v.max_count > 0 }
}

resource "azurerm_user_assigned_identity" "aks" {
  name                = var.prefix
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = local.tags

  lifecycle {
    precondition {
      condition     = !var.api_server_vnet_integration_enabled || var.api_server_subnet_id != ""
      error_message = "api_server_vnet_integration_enabled needs api_server_subnet_id (vpc apiserver_subnets)."
    }
    precondition {
      condition     = !local.disk_encryption || var.kms_key_vault_id != ""
      error_message = "disk_encryption_key_id needs kms_key_vault_id."
    }
  }
}

resource "azurerm_role_assignment" "aks_network" {
  scope                            = var.vnet_subnet_id
  role_definition_name             = "Network Contributor"
  principal_id                     = azurerm_user_assigned_identity.aks.principal_id
  skip_service_principal_aad_check = true
}

resource "azurerm_role_assignment" "aks_apiserver_network" {
  count                            = var.api_server_vnet_integration_enabled ? 1 : 0
  scope                            = var.api_server_subnet_id
  role_definition_name             = "Network Contributor"
  principal_id                     = azurerm_user_assigned_identity.aks.principal_id
  skip_service_principal_aad_check = true
}

# Bootstrap cache: the kubelet identity needs AcrPull before the nodes bootstrap, so it can't be the AKS-created one
resource "azurerm_user_assigned_identity" "kubelet" {
  count               = var.bootstrap_cache_enabled ? 1 : 0
  name                = "${var.prefix}-kubelet"
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = local.tags

  lifecycle {
    precondition {
      condition     = var.acr_id != ""
      error_message = "bootstrap_cache_enabled needs acr_id (Premium acr-proxy with private endpoint and aks_bootstrap_cache_rule)."
    }
  }
}

resource "azurerm_role_assignment" "kubelet_identity_operator" {
  count                            = var.bootstrap_cache_enabled ? 1 : 0
  scope                            = azurerm_user_assigned_identity.kubelet[0].id
  role_definition_name             = "Managed Identity Operator"
  principal_id                     = azurerm_user_assigned_identity.aks.principal_id
  skip_service_principal_aad_check = true
}

resource "azurerm_role_assignment" "kubelet_bootstrap_acr_pull" {
  count                            = var.bootstrap_cache_enabled ? 1 : 0
  scope                            = var.acr_id
  role_definition_name             = "AcrPull"
  principal_id                     = azurerm_user_assigned_identity.kubelet[0].principal_id
  skip_service_principal_aad_check = true
}

resource "azurerm_disk_encryption_set" "this" {
  count                     = local.disk_encryption ? 1 : 0
  name                      = var.prefix
  location                  = var.location
  resource_group_name       = var.resource_group_name
  key_vault_key_id          = var.disk_encryption_key_id
  auto_key_rotation_enabled = true
  tags                      = local.tags

  identity {
    type = "SystemAssigned"
  }
}

resource "azurerm_role_assignment" "disk_encryption_key" {
  count                            = local.disk_encryption ? 1 : 0
  scope                            = "${var.kms_key_vault_id}/keys/${element(split("/", var.disk_encryption_key_id), 4)}"
  role_definition_name             = "Key Vault Crypto Service Encryption User"
  principal_id                     = azurerm_disk_encryption_set.this[0].identity[0].principal_id
  skip_service_principal_aad_check = true
}

resource "azurerm_role_assignment" "disk_encryption_reader" {
  count                            = local.disk_encryption ? 1 : 0
  scope                            = azurerm_disk_encryption_set.this[0].id
  role_definition_name             = "Reader"
  principal_id                     = azurerm_user_assigned_identity.aks.principal_id
  skip_service_principal_aad_check = true
}


# AKS checks key access at creation, role assignments take up to a minute
resource "time_sleep" "key_roles" {
  count           = local.disk_encryption ? 1 : 0
  create_duration = "60s"
  depends_on      = [azurerm_role_assignment.disk_encryption_key, azurerm_role_assignment.disk_encryption_reader]
}

# https://github.com/Azure/terraform-azurerm-avm-res-containerservice-managedcluster
module "aks" {
  source  = "Azure/avm-res-containerservice-managedcluster/azurerm"
  version = "0.8.3"

  name                = var.prefix
  location            = var.location
  parent_id           = data.azurerm_resource_group.this.id
  dns_prefix          = var.prefix
  node_resource_group = var.node_resource_group_name != "" ? var.node_resource_group_name : "${var.prefix}-nodes-${var.location}"
  enable_telemetry    = false
  tags                = local.tags

  kubernetes_version = var.kubernetes_version
  sku = {
    name = "Base"
    tier = var.sku_tier
  }

  auto_upgrade_profile = {
    upgrade_channel         = var.upgrade_channel
    node_os_upgrade_channel = var.node_os_upgrade_channel
  }
  maintenanceconfiguration = local.maintenance_configurations
  auto_scaler_profile      = var.auto_scaler_profile

  managed_identities = {
    user_assigned_resource_ids = [azurerm_user_assigned_identity.aks.id]
  }

  bootstrap_profile = var.bootstrap_cache_enabled ? {
    artifact_source       = "Cache"
    container_registry_id = var.acr_id
  } : null
  identity_profile = var.bootstrap_cache_enabled ? {
    kubeletidentity = { resource_id = azurerm_user_assigned_identity.kubelet[0].id }
  } : null

  enable_rbac            = true
  disable_local_accounts = var.disable_local_accounts
  aad_profile = {
    managed           = true
    enable_azure_rbac = true
    tenant_id         = data.azurerm_client_config.this.tenant_id
  }

  api_server_access_profile = {
    enable_private_cluster             = var.private_cluster_enabled
    enable_private_cluster_public_fqdn = var.private_cluster_enabled ? false : null
    # [] not null: AVM drops nulls and AKS would keep the old ranges, which private mode rejects
    authorized_ip_ranges    = var.private_cluster_enabled ? [] : var.api_server_authorized_ip_ranges
    enable_vnet_integration = var.api_server_vnet_integration_enabled
    subnet_id               = var.api_server_vnet_integration_enabled ? var.api_server_subnet_id : null
  }

  network_profile = {
    network_plugin      = "azure"
    network_plugin_mode = "overlay"
    network_dataplane   = "cilium"
    network_policy      = "cilium"
    outbound_type       = var.outbound_type
    load_balancer_sku   = "standard"
    pod_cidr            = var.pod_cidr
    service_cidr        = var.service_cidr
    dns_service_ip      = var.dns_service_ip
  }

  oidc_issuer_profile = {
    enabled = true
  }
  security_profile = {
    workload_identity = {
      enabled = true
    }
  }

  disk_encryption_set_id = local.disk_encryption ? azurerm_disk_encryption_set.this[0].id : null

  default_agent_pool = local.default_agent_pool

  agent_pools = merge(local.agent_pools, var.aks_managed_node_groups_extra)

  depends_on = [
    azurerm_role_assignment.aks_network,
    azurerm_role_assignment.aks_apiserver_network,
    azurerm_role_assignment.kubelet_identity_operator,
    azurerm_role_assignment.kubelet_bootstrap_acr_pull,
    time_sleep.key_roles,
  ]
}

resource "azurerm_role_assignment" "kubelet_acr_pull" {
  count                            = var.acr_id != "" && !var.bootstrap_cache_enabled ? 1 : 0
  scope                            = var.acr_id
  role_definition_name             = "AcrPull"
  principal_id                     = module.aks.kubelet_identity.objectId
  skip_service_principal_aad_check = true
}

resource "azurerm_role_assignment" "kubelet_additional" {
  for_each                         = { for i, a in var.kubelet_additional_role_assignments : tostring(i) => a }
  scope                            = each.value.scope
  role_definition_name             = each.value.role
  principal_id                     = var.bootstrap_cache_enabled ? azurerm_user_assigned_identity.kubelet[0].principal_id : module.aks.kubelet_identity.objectId
  skip_service_principal_aad_check = true
}

resource "azurerm_role_assignment" "cluster_admin" {
  for_each             = toset(concat([data.azurerm_client_config.this.object_id], var.admin_object_ids))
  scope                = module.aks.resource_id
  role_definition_name = "Azure Kubernetes Service RBAC Cluster Admin"
  principal_id         = each.value
}

resource "azurerm_log_analytics_workspace" "control_plane" {
  count               = var.control_plane_logs_enabled ? 1 : 0
  name                = "${var.prefix}-logs"
  location            = var.location
  resource_group_name = var.resource_group_name
  sku                 = "PerGB2018"
  retention_in_days   = var.control_plane_logs_retention_days
  tags                = local.tags
}

resource "azurerm_monitor_diagnostic_setting" "control_plane" {
  count                          = var.control_plane_logs_enabled ? 1 : 0
  name                           = "${var.prefix}-control-plane"
  target_resource_id             = module.aks.resource_id
  log_analytics_workspace_id     = azurerm_log_analytics_workspace.control_plane[0].id
  log_analytics_destination_type = "Dedicated"

  dynamic "enabled_log" {
    for_each = toset(var.control_plane_log_categories)
    content {
      category = enabled_log.value
    }
  }
}

# ALB controller (azure-gateway k8s module): its chart needs the client id at render time and ALB-managed AGC lives in the node resource group
resource "azurerm_user_assigned_identity" "alb_controller" {
  count               = length(var.agc_subnet_ids) > 0 ? 1 : 0
  name                = "${var.prefix}-alb-controller"
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = local.tags
}

resource "azurerm_federated_identity_credential" "alb_controller" {
  count                     = length(var.agc_subnet_ids) > 0 ? 1 : 0
  name                      = "alb-controller"
  user_assigned_identity_id = azurerm_user_assigned_identity.alb_controller[0].id
  audience                  = ["api://AzureADTokenExchange"]
  issuer                    = module.aks.oidc_issuer_profile_issuer_url
  subject                   = "system:serviceaccount:azure-alb-system:alb-controller-sa"
}

resource "azurerm_role_assignment" "alb_controller" {
  for_each = length(var.agc_subnet_ids) > 0 ? merge(
    {
      node-rg-reader = { scope = module.aks.node_resource_group_id, role = "Reader" }
      config-manager = { scope = module.aks.node_resource_group_id, role = "AppGw for Containers Configuration Manager" }
    },
    { for i, id in var.agc_subnet_ids : "agc-subnet-${i}" => { scope = id, role = "Network Contributor" } }
  ) : {}
  scope                            = each.value.scope
  role_definition_name             = each.value.role
  principal_id                     = azurerm_user_assigned_identity.alb_controller[0].principal_id
  skip_service_principal_aad_check = true
}
