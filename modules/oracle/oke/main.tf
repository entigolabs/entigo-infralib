locals {
  # OCI returns available versions oldest-first; take the last (newest) one when the
  # caller doesn't pin a specific version.
  versions           = data.oci_containerengine_cluster_option.this.kubernetes_versions
  kubernetes_version = var.kubernetes_version != "" ? var.kubernetes_version : local.versions[length(local.versions) - 1]

  # OCI rejects a private subnet for the endpoint when is_public_ip_enabled is true
  # ("must be a public subnet if public ip enabled"), so the subnet choice must follow it.
  endpoint_subnet_id = var.is_public_ip_enabled ? var.public_subnet_id : var.private_subnet_id

  main_subnet_ids  = length(var.oke_main_subnet_ids) > 0 ? var.oke_main_subnet_ids : var.node_subnet_ids
  mon_subnet_ids   = length(var.oke_mon_subnet_ids) > 0 ? var.oke_mon_subnet_ids : var.node_subnet_ids
  tools_subnet_ids = length(var.oke_tools_subnet_ids) > 0 ? var.oke_tools_subnet_ids : var.node_subnet_ids
}

resource "oci_containerengine_cluster" "this" {
  compartment_id     = var.compartment_id
  name               = var.prefix
  vcn_id             = var.vcn_id
  kubernetes_version = local.kubernetes_version

  endpoint_config {
    is_public_ip_enabled = var.is_public_ip_enabled
    subnet_id            = local.endpoint_subnet_id
  }

  cluster_pod_network_options {
    cni_type = "FLANNEL_OVERLAY"
  }

  options {
    service_lb_subnet_ids = var.service_lb_subnet_ids

    kubernetes_network_config {
      pods_cidr     = var.pods_cidr
      services_cidr = var.services_cidr
    }
  }
}

module "main" {
  count  = var.oke_main_node_count > 0 ? 1 : 0
  source = "../oke-node-pool"

  prefix                  = var.prefix
  pool_name               = "main"
  compartment_id          = var.compartment_id
  cluster_id              = oci_containerengine_cluster.this.id
  kubernetes_version      = local.kubernetes_version
  subnet_ids              = local.main_subnet_ids
  node_shape              = var.oke_main_node_shape
  ocpus                   = var.oke_main_ocpus
  memory_in_gbs           = var.oke_main_memory_in_gbs
  node_count              = var.oke_main_node_count
  boot_volume_size_in_gbs = var.oke_main_boot_volume_size_in_gbs
  node_pool_os_type       = var.oke_main_node_pool_os_type
  labels                  = { main = "true" }
}

module "mon" {
  count  = var.oke_mon_node_count > 0 ? 1 : 0
  source = "../oke-node-pool"

  prefix                  = var.prefix
  pool_name               = "mon"
  compartment_id          = var.compartment_id
  cluster_id              = oci_containerengine_cluster.this.id
  kubernetes_version      = local.kubernetes_version
  subnet_ids              = local.mon_subnet_ids
  node_shape              = var.oke_mon_node_shape
  ocpus                   = var.oke_mon_ocpus
  memory_in_gbs           = var.oke_mon_memory_in_gbs
  node_count              = var.oke_mon_node_count
  boot_volume_size_in_gbs = var.oke_mon_boot_volume_size_in_gbs
  node_pool_os_type       = var.oke_mon_node_pool_os_type
  # No NO_SCHEDULE taint - oci_containerengine_node_pool has no taint attribute in the
  # provider schema (see NOTES.md "Known, permanent-for-now limitation"). Label-only.
  labels = { mon = "true" }
}

module "tools" {
  count  = var.oke_tools_node_count > 0 ? 1 : 0
  source = "../oke-node-pool"

  prefix                  = var.prefix
  pool_name               = "tools"
  compartment_id          = var.compartment_id
  cluster_id              = oci_containerengine_cluster.this.id
  kubernetes_version      = local.kubernetes_version
  subnet_ids              = local.tools_subnet_ids
  node_shape              = var.oke_tools_node_shape
  ocpus                   = var.oke_tools_ocpus
  memory_in_gbs           = var.oke_tools_memory_in_gbs
  node_count              = var.oke_tools_node_count
  boot_volume_size_in_gbs = var.oke_tools_boot_volume_size_in_gbs
  node_pool_os_type       = var.oke_tools_node_pool_os_type
  labels                  = { tools = "true" }
}
