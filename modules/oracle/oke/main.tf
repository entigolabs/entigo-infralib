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

# oracle/vpc doesn't create any security lists/NSGs of its own, so subnets fall back to
# the VCN's auto-created default security list, which only allows SSH-22 and ICMP path
# discovery inbound - nothing on 6443/10250/12250, and nothing between worker nodes. That
# silently breaks node registration ("1 node(s) register timeout"): the control plane and
# worker nodes can never actually reach each other. Egress is already open (the default
# security list allows all egress), so this NSG only needs to add the missing ingress -
# see https://docs.oracle.com/en-us/iaas/Content/ContEng/Concepts/contengnetworkconfig.htm.
# Shared by both the cluster endpoint and every node pool since OCI evaluates all
# security lists/NSGs on a VNIC as a union, not an intersection.
resource "oci_core_network_security_group" "this" {
  compartment_id = var.compartment_id
  vcn_id         = var.vcn_id
  display_name   = "${var.prefix}-oke"
}

resource "oci_core_network_security_group_security_rule" "intra_vcn" {
  network_security_group_id = oci_core_network_security_group.this.id
  direction                 = "INGRESS"
  protocol                  = "all"
  source                    = data.oci_core_vcn.this.cidr_blocks[0]
  source_type               = "CIDR_BLOCK"
  description               = "Control plane <-> worker node <-> worker node (6443, 10250, 12250, flannel overlay)"
}

resource "oci_core_network_security_group_security_rule" "path_mtu_discovery" {
  network_security_group_id = oci_core_network_security_group.this.id
  direction                 = "INGRESS"
  protocol                  = "1"
  source                    = "0.0.0.0/0"
  source_type               = "CIDR_BLOCK"
  description               = "Path MTU discovery"

  icmp_options {
    type = 3
    code = 4
  }
}

resource "oci_containerengine_cluster" "this" {
  compartment_id     = var.compartment_id
  name               = var.prefix
  vcn_id             = var.vcn_id
  kubernetes_version = local.kubernetes_version

  endpoint_config {
    is_public_ip_enabled = var.is_public_ip_enabled
    subnet_id            = local.endpoint_subnet_id
    nsg_ids              = [oci_core_network_security_group.this.id]
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
  nsg_ids                 = [oci_core_network_security_group.this.id]
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
  labels  = { mon = "true" }
  nsg_ids = [oci_core_network_security_group.this.id]
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
  nsg_ids                 = [oci_core_network_security_group.this.id]
}
