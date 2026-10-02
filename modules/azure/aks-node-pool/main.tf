data "azurerm_client_config" "this" {}

data "azapi_resource_list" "vm_skus" {
  count                  = var.availability_zones == null ? 1 : 0
  type                   = "Microsoft.Compute/skus@2021-07-01"
  parent_id              = "/subscriptions/${data.azurerm_client_config.this.subscription_id}"
  query_parameters       = { "$filter" = ["location eq '${var.location}'"] }
  response_export_values = { vms = "value[?resourceType=='virtualMachines'].{name: name, zones: locationInfo[0].zones}" }
}

locals {
  availability_zones = var.availability_zones != null ? var.availability_zones : sort(coalesce(one([
    for v in data.azapi_resource_list.vm_skus[0].output.vms : v.zones if v.name == var.instance_type
  ]), []))

  node_pool_name = substr(replace(lower(var.name != "" ? var.name : var.prefix), "/[^a-z0-9]/", ""), 0, 12)

  # AKS adds it to spot pools, declared to avoid drift
  spot_taint = "kubernetes.azure.com/scalesetpriority=spot:NoSchedule"

  tags = merge(var.tags, {
    Terraform  = "true"
    Prefix     = var.prefix
    created-by = "entigo-infralib"
  })
}

# Name check at plan time, AVM would only fail at apply
resource "terraform_data" "name" {
  input = local.node_pool_name

  lifecycle {
    precondition {
      condition     = can(regex("^[a-z][a-z0-9]{0,11}$", local.node_pool_name)) && !contains(["main", "mon", "tools"], local.node_pool_name)
      error_message = "Node pool name ${local.node_pool_name} must start with a letter and not be main, mon or tools, set name."
    }
  }
}

# https://github.com/Azure/terraform-azurerm-avm-res-containerservice-managedcluster/tree/main/modules/agentpool
# Disk encryption set, kubelet identity and bootstrap cache come from the cluster
module "aks_node_pool" {
  source  = "Azure/avm-res-containerservice-managedcluster/azurerm//modules/agentpool"
  version = "0.8.3"

  name                 = terraform_data.name.output
  parent_id            = var.cluster_id
  mode                 = var.mode
  type                 = "VirtualMachineScaleSets"
  orchestrator_version = var.kubernetes_version
  vnet_subnet_id       = var.vnet_subnet_id

  vm_size             = var.instance_type
  enable_auto_scaling = true
  min_count           = var.min_size
  max_count           = var.max_size
  max_pods            = var.max_pods
  availability_zones  = local.availability_zones

  os_sku          = var.os_sku
  os_disk_size_gb = var.volume_size
  os_disk_type    = var.volume_type

  scale_set_priority        = var.spot_nodes ? "Spot" : "Regular"
  scale_set_eviction_policy = var.spot_nodes ? "Delete" : null
  spot_max_price            = var.spot_nodes ? var.spot_max_price : null

  node_taints = concat(var.taints, var.spot_nodes ? [local.spot_taint] : [])
  node_labels = merge(var.labels, { created-by = "entigo-infralib" })
  tags        = local.tags

  # AKS rejects surge settings on spot pools
  upgrade_settings = var.spot_nodes ? null : (var.upgrade_settings != null ? var.upgrade_settings : {
    max_surge = var.max_surge
  })
}
