data "azurerm_client_config" "this" {}

data "azapi_resource_list" "vm_skus" {
  count                  = var.availability_zones == null ? 1 : 0
  type                   = "Microsoft.Compute/skus@2021-07-01"
  parent_id              = "/subscriptions/${data.azurerm_client_config.this.subscription_id}"
  query_parameters       = { "$filter" = ["location eq '${var.location}'"] }
  response_export_values = { vms = "value[?resourceType=='virtualMachines'].{name: name, zones: locationInfo[0].zones, restricted_zones: restrictions[?type=='Zone'].restrictionInfo.zones[], location_restrictions: restrictions[?type=='Location'].reasonCode}" }
}

locals {
  # Zones restricted for the subscription (NotAvailableForSubscription) are left out, like azure/aks
  sku = one([
    for v in try(data.azapi_resource_list.vm_skus[0].output.vms, []) : v
    if v.name == var.instance_type && length(coalesce(try(v.location_restrictions, null), [])) == 0
  ])
  availability_zones = var.availability_zones != null ? var.availability_zones : (local.sku == null ? [] : sort(setsubtract(
    coalesce(try(local.sku.zones, null), []), coalesce(try(local.sku.restricted_zones, null), [])
  )))

  node_pool_name = substr(replace(lower(var.name != "" ? var.name : var.prefix), "/[^a-z0-9]/", ""), 0, 12)

  # AKS adds the taint and label to spot pools, declared to avoid drift
  spot_taint = "kubernetes.azure.com/scalesetpriority=spot:NoSchedule"

  tags = merge(var.tags, {
    Terraform  = "true"
    Prefix     = var.prefix
    created-by = "entigo-infralib"
  })
}

# Changes to zones, VM size, max pods, disks or subnet rotate the pool through a temporary pool (AVM's agentpool only
# replaces it on a VM size change), priority (spot) changes replace it
resource "azurerm_kubernetes_cluster_node_pool" "this" {
  name                        = local.node_pool_name
  temporary_name_for_rotation = "${substr(local.node_pool_name, 0, 9)}tmp"
  kubernetes_cluster_id       = var.cluster_id
  mode                        = var.mode
  orchestrator_version        = var.kubernetes_version
  vm_size                     = var.instance_type
  auto_scaling_enabled        = true
  min_count                   = var.min_size
  max_count                   = var.max_size
  max_pods                    = var.max_pods
  os_disk_size_gb             = var.volume_size
  os_disk_type                = var.volume_type
  os_sku                      = var.os_sku
  zones                       = local.availability_zones
  vnet_subnet_id              = var.vnet_subnet_id
  node_labels                 = merge(var.labels, var.spot_nodes ? { "kubernetes.azure.com/scalesetpriority" = "spot" } : {}, { created-by = "entigo-infralib" })
  node_taints                 = concat(var.taints, var.spot_nodes ? [local.spot_taint] : [])
  priority                    = var.spot_nodes ? "Spot" : "Regular"
  eviction_policy             = var.spot_nodes ? "Delete" : null
  spot_max_price              = var.spot_nodes ? var.spot_max_price : null
  tags                        = local.tags

  # AKS rejects surge settings on spot pools
  dynamic "upgrade_settings" {
    for_each = var.spot_nodes ? [] : [var.upgrade_settings != null ? var.upgrade_settings : { max_surge = var.max_surge }]
    content {
      max_surge                     = upgrade_settings.value.max_surge
      drain_timeout_in_minutes      = try(upgrade_settings.value.drain_timeout_in_minutes, null)
      node_soak_duration_in_minutes = try(upgrade_settings.value.node_soak_duration_in_minutes, null)
    }
  }

  lifecycle {
    ignore_changes = [node_count]
    precondition {
      condition     = var.availability_zones != null || local.sku != null
      error_message = "instance_type ${var.instance_type} is not available in location '${var.location}' for this subscription."
    }
  }
}
