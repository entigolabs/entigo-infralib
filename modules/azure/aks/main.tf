data "azurerm_client_config" "this" {}

data "azurerm_resource_group" "this" {
  name = var.resource_group_name
}

data "azapi_resource_list" "vm_skus" {
  count                  = var.availability_zones == null ? 1 : 0
  type                   = "Microsoft.Compute/skus@2021-07-01"
  parent_id              = "/subscriptions/${data.azurerm_client_config.this.subscription_id}"
  query_parameters       = { "$filter" = ["location eq '${var.location}'"] }
  response_export_values = { vms = "value[?resourceType=='virtualMachines'].{name: name, zones: locationInfo[0].zones, restricted_zones: restrictions[?type=='Zone'].restrictionInfo.zones[], location_restrictions: restrictions[?type=='Location'].reasonCode}" }
}

locals {
  disk_encryption = var.disk_encryption_key_id != ""

  # Zones restricted for the subscription (NotAvailableForSubscription) are left out
  skus = { for v in try(data.azapi_resource_list.vm_skus[0].output.vms, []) : v.name => v }
  sku_zones = {
    for name, v in local.skus : name => sort(setsubtract(coalesce(try(v.zones, null), []), coalesce(try(v.restricted_zones, null), [])))
    if length(coalesce(try(v.location_restrictions, null), [])) == 0
  }
  zones_for = { for size in distinct([var.aks_main_instance_type, var.aks_mon_instance_type, var.aks_tools_instance_type]) :
    size => var.availability_zones != null ? var.availability_zones : lookup(local.sku_zones, size, [])
  }

  # Sizes checked against the SKU list when the zones aren't set
  checked_sizes = var.availability_zones == null ? merge(
    { tools = var.aks_tools_instance_type },
    var.aks_main_max_size > 0 ? { main = var.aks_main_instance_type } : {},
    var.aks_mon_max_size > 0 ? { mon = var.aks_mon_instance_type } : {},
  ) : {}
  unavailable_sizes = [for pool, size in local.checked_sizes : "${pool} (${size})" if !contains(keys(local.sku_zones), size)]

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

  # tools is the AKS system pool (AVM default pool): CriticalAddonsOnly taint instead of tools=true, AKS system pods
  # only tolerate that
  default_agent_pool = {
    name                 = "tools"
    mode                 = "System"
    type                 = "VirtualMachineScaleSets"
    vm_size              = var.aks_tools_instance_type
    enable_auto_scaling  = true
    min_count            = var.aks_tools_min_size
    max_count            = var.aks_tools_max_size
    max_pods             = var.aks_tools_max_pods
    os_disk_size_gb      = var.aks_tools_volume_size
    os_disk_type         = var.aks_tools_volume_type
    os_sku               = var.os_sku
    availability_zones   = local.zones_for[var.aks_tools_instance_type]
    node_labels          = { "tools" = "true", created-by = "entigo-infralib" }
    node_taints          = ["CriticalAddonsOnly=true:NoSchedule"]
    upgrade_settings     = { max_surge = var.max_surge }
    orchestrator_version = var.kubernetes_version
    vnet_subnet_id       = var.vnet_subnet_id
    tags                 = local.tags
  }

  # Agent jobs reach a public API server from the NAT gateway IPs
  authorized_ip_ranges = length(var.api_server_authorized_ip_ranges) == 0 ? [] : distinct(concat(
    var.api_server_authorized_ip_ranges, [for ip in var.nat_public_ips : "${ip}/32"]
  ))
}

resource "azurerm_user_assigned_identity" "aks" {
  name                = var.prefix
  location            = var.location
  resource_group_name = var.resource_group_name
  tags                = local.tags

  lifecycle {
    precondition {
      condition     = length(local.unavailable_sizes) == 0
      error_message = "VM sizes not available in ${var.location} for this subscription: ${join(", ", local.unavailable_sizes)}."
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

# Central private DNS zone: AKS writes the API server record and links the zone to the cluster VNet
resource "azurerm_role_assignment" "aks_private_dns_zone" {
  count                            = var.private_dns_zone_id != "" ? 1 : 0
  scope                            = var.private_dns_zone_id
  role_definition_name             = "Private DNS Zone Contributor"
  principal_id                     = azurerm_user_assigned_identity.aks.principal_id
  skip_service_principal_aad_check = true
}

resource "azurerm_role_assignment" "aks_vnet" {
  count                            = var.private_dns_zone_id != "" ? 1 : 0
  scope                            = regex("^(.*)/subnets/[^/]+$", var.vnet_subnet_id)[0]
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
  scope                            = var.disk_encryption_key_resource_id
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

# AKS checks its role assignments at creation, they take up to a minute to propagate
resource "time_sleep" "roles" {
  create_duration = "60s"
  depends_on = [
    azurerm_role_assignment.aks_network,
    azurerm_role_assignment.aks_apiserver_network,
    azurerm_role_assignment.aks_private_dns_zone,
    azurerm_role_assignment.aks_vnet,
    azurerm_role_assignment.kubelet_identity_operator,
    azurerm_role_assignment.kubelet_bootstrap_acr_pull,
    azurerm_role_assignment.disk_encryption_key,
    azurerm_role_assignment.disk_encryption_reader,
  ]
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
    authorized_ip_ranges    = var.private_cluster_enabled ? [] : local.authorized_ip_ranges
    enable_vnet_integration = var.api_server_vnet_integration_enabled
    subnet_id               = var.api_server_vnet_integration_enabled ? var.api_server_subnet_id : null
    private_dns_zone        = var.private_dns_zone_id != "" ? var.private_dns_zone_id : null
  }

  # https://learn.microsoft.com/azure/aks/concepts-network-cni-overview
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

  depends_on = [time_sleep.roles]
}

module "main" {
  count  = var.aks_main_max_size > 0 ? 1 : 0
  source = "./aks-node-pool"

  prefix             = var.prefix
  name               = "main"
  cluster_id         = module.aks.resource_id
  kubernetes_version = var.kubernetes_version
  vnet_subnet_id     = var.vnet_subnet_id
  location           = var.location
  availability_zones = local.zones_for[var.aks_main_instance_type]
  instance_type      = var.aks_main_instance_type
  min_size           = var.aks_main_min_size
  max_size           = var.aks_main_max_size
  max_pods           = var.aks_main_max_pods
  volume_size        = var.aks_main_volume_size
  volume_type        = var.aks_main_volume_type
  os_sku             = var.os_sku
  spot_nodes         = var.aks_main_spot_nodes
  max_surge          = var.max_surge
  labels             = { "main" = "true" }
  tags               = var.tags

  depends_on = [module.aks]
}

module "mon" {
  count  = var.aks_mon_max_size > 0 ? 1 : 0
  source = "./aks-node-pool"

  prefix             = var.prefix
  name               = "mon"
  cluster_id         = module.aks.resource_id
  kubernetes_version = var.kubernetes_version
  vnet_subnet_id     = var.vnet_subnet_id
  location           = var.location
  availability_zones = local.zones_for[var.aks_mon_instance_type]
  instance_type      = var.aks_mon_instance_type
  min_size           = var.aks_mon_min_size
  max_size           = var.aks_mon_max_size
  max_pods           = var.aks_mon_max_pods
  volume_size        = var.aks_mon_volume_size
  volume_type        = var.aks_mon_volume_type
  os_sku             = var.os_sku
  spot_nodes         = var.aks_mon_spot_nodes
  max_surge          = var.max_surge
  labels             = { "mon" = "true" }
  taints             = ["mon=true:NoSchedule"]
  tags               = var.tags

  depends_on = [module.aks]
}

resource "azurerm_role_assignment" "kubelet_acr_pull" {
  count                            = var.acr_id != "" && !var.bootstrap_cache_enabled ? 1 : 0
  scope                            = var.acr_id
  role_definition_name             = "AcrPull"
  principal_id                     = module.aks.kubelet_identity.objectId
  skip_service_principal_aad_check = true
}

resource "azurerm_role_assignment" "kubelet_additional" {
  for_each                         = { for a in var.kubelet_additional_role_assignments : "${a.role}|${a.scope}" => a }
  scope                            = each.value.scope
  role_definition_name             = each.value.role
  principal_id                     = var.bootstrap_cache_enabled ? azurerm_user_assigned_identity.kubelet[0].principal_id : module.aks.kubelet_identity.objectId
  skip_service_principal_aad_check = true
}

# The installing identity (normally the agent job identity) keeps the role, later callers don't replace it
resource "terraform_data" "installer" {
  input = data.azurerm_client_config.this.object_id

  lifecycle {
    ignore_changes = [input]
  }
}

resource "azurerm_role_assignment" "cluster_admin" {
  for_each             = merge({ installer = terraform_data.installer.input }, { for id in toset([for i in var.admin_object_ids : lower(i)]) : id => id })
  scope                = module.aks.resource_id
  role_definition_name = "Azure Kubernetes Service RBAC Cluster Admin"
  principal_id         = each.value

  lifecycle {
    precondition {
      condition     = each.key == "installer" || each.value != lower(terraform_data.installer.input)
      error_message = "admin_object_ids must not list the installing identity ${terraform_data.installer.input}, it has the role already."
    }
  }
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
