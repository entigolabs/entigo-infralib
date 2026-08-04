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
# security list allows all egress), so these NSGs only need to add the missing ingress -
# see https://docs.oracle.com/en-us/iaas/Content/ContEng/Concepts/contengnetworkconfig.htm.
#
# Split in two rather than one shared "allow all" group: the control plane only ever needs
# 6443/12250 from workers, so it gets exactly that. Workers genuinely need ALL protocols
# from each other (Flannel VXLAN, kube-proxy, etc.) per Oracle's own doc, and TCP/ALL from
# the endpoint (flannel CNI row) - narrowing that further would mean reverse-engineering
# Flannel's exact port usage, so it's scoped only as far as source (VCN CIDR), not protocol.
resource "oci_core_network_security_group" "endpoint" {
  compartment_id = var.compartment_id
  vcn_id         = var.vcn_id
  display_name   = "${var.prefix}-oke-endpoint"
}

resource "oci_core_network_security_group_security_rule" "endpoint_kube_api" {
  network_security_group_id = oci_core_network_security_group.endpoint.id
  direction                 = "INGRESS"
  protocol                  = "6" # TCP
  source                    = data.oci_core_vcn.this.cidr_blocks[0]
  source_type               = "CIDR_BLOCK"
  description               = "Worker node -> Kubernetes API"

  tcp_options {
    destination_port_range {
      min = 6443
      max = 6443
    }
  }
}

resource "oci_core_network_security_group_security_rule" "endpoint_oke_control" {
  network_security_group_id = oci_core_network_security_group.endpoint.id
  direction                 = "INGRESS"
  protocol                  = "6" # TCP
  source                    = data.oci_core_vcn.this.cidr_blocks[0]
  source_type               = "CIDR_BLOCK"
  description               = "Worker node -> OKE control plane secondary channel"

  tcp_options {
    destination_port_range {
      min = 12250
      max = 12250
    }
  }
}

resource "oci_core_network_security_group_security_rule" "endpoint_path_mtu_discovery" {
  network_security_group_id = oci_core_network_security_group.endpoint.id
  direction                 = "INGRESS"
  protocol                  = "1" # ICMP
  source                    = "0.0.0.0/0"
  source_type               = "CIDR_BLOCK"
  description               = "Path MTU discovery"

  icmp_options {
    type = 3
    code = 4
  }
}

resource "oci_core_network_security_group" "node" {
  compartment_id = var.compartment_id
  vcn_id         = var.vcn_id
  display_name   = "${var.prefix}-oke-node"
}

resource "oci_core_network_security_group_security_rule" "node_intra_vcn" {
  network_security_group_id = oci_core_network_security_group.node.id
  direction                 = "INGRESS"
  protocol                  = "all"
  source                    = data.oci_core_vcn.this.cidr_blocks[0]
  source_type               = "CIDR_BLOCK"
  description               = "Control plane -> worker node (10250, flannel) and worker node <-> worker node"
}

resource "oci_core_network_security_group_security_rule" "node_path_mtu_discovery" {
  network_security_group_id = oci_core_network_security_group.node.id
  direction                 = "INGRESS"
  protocol                  = "1" # ICMP
  source                    = "0.0.0.0/0"
  source_type               = "CIDR_BLOCK"
  description               = "Path MTU discovery"

  icmp_options {
    type = 3
    code = 4
  }
}

# NSG for the OCI LBs created by oci-native-ingress-controller. The public subnet's
# default security list only allows SSH-22 + ICMP inbound (confirmed live: HTTP to the
# ingress LB timed out until this existed), and NIC does not manage security lists or NSGs
# itself - it only *attaches* the LB to NSGs listed in the IngressClass's
# oci-native-ingress.oraclecloud.com/network-security-group-ids annotation, which
# modules/k8s/oci-native-ingress-controller wires to this NSG's id via lb_nsg_id output.
# Backend LB->node traffic is already covered by the node NSG's allow-all-from-VCN rule.
resource "oci_core_network_security_group" "lb" {
  compartment_id = var.compartment_id
  vcn_id         = var.vcn_id
  display_name   = "${var.prefix}-oke-lb"
}

resource "oci_core_network_security_group_security_rule" "lb_http" {
  network_security_group_id = oci_core_network_security_group.lb.id
  direction                 = "INGRESS"
  protocol                  = "6" # TCP
  source                    = "0.0.0.0/0"
  source_type               = "CIDR_BLOCK"
  description               = "Public HTTP to ingress load balancers"

  tcp_options {
    destination_port_range {
      min = 80
      max = 80
    }
  }
}

resource "oci_core_network_security_group_security_rule" "lb_https" {
  network_security_group_id = oci_core_network_security_group.lb.id
  direction                 = "INGRESS"
  protocol                  = "6" # TCP
  source                    = "0.0.0.0/0"
  source_type               = "CIDR_BLOCK"
  description               = "Public HTTPS to ingress load balancers"

  tcp_options {
    destination_port_range {
      min = 443
      max = 443
    }
  }
}

# Dynamic groups are tenancy-scoped in OCI (unlike policies): the resource's compartment_id
# argument must be the tenancy OCID, not the target compartment. This project's compartments
# are flat - a single level below the tenancy root (confirmed live via `oci iam compartment
# get`, whose compartment_id is exactly the tenancy OCID from ~/.oci/config) - so the parent
# of var.compartment_id is reliably the tenancy here. A deeper compartment hierarchy would
# need to walk further up instead of taking the immediate parent.
data "oci_identity_compartment" "this" {
  id = var.compartment_id
}

# Instance Principal identity for in-cluster controllers (Crossplane OCI provider,
# external-dns, cluster-autoscaler). This cluster is a Basic OKE cluster (no OKE Workload
# Identity available), so instance principal - any instance in the compartment - is the
# only no-static-credential auth option. Shared rather than split per-controller since
# they run on the same node pools within the same compartment boundary (per-pod isolation
# is impossible with instance principal anyway - every pod shares the node's identity).
resource "oci_identity_dynamic_group" "controllers" {
  compartment_id = data.oci_identity_compartment.this.compartment_id
  name           = "${var.prefix}-oke-controllers"
  description    = "Instances in ${var.prefix}'s OKE node pools - used by in-cluster controllers (crossplane, external-dns, cluster-autoscaler) via instance principal auth"
  matching_rule  = "ALL {instance.compartment.id = '${var.compartment_id}'}"
}

resource "oci_identity_policy" "controllers" {
  # Bootstrap grant only: per-app permissions (external-dns "manage dns",
  # cluster-autoscaler's node-pool statements, the ingress controller's cert/LB set)
  # live in each k8s module's templates/oracle/ as Crossplane Policy CRs, applied by
  # the crossplane-oracle provider - which is what this statement authorizes. Attached
  # at the compartment (not the tenancy like before the Crossplane migration): OCI only
  # lets a policy grant within the subtree it is attached to, so compartment-level
  # "manage policies" cannot be escalated beyond this compartment. The flip side: a
  # Crossplane Policy CR needing "in tenancy" statements (oci-native-ingress-controller
  # would - public-ips, floating-ips, tag-namespaces) requires widening this grant to
  # tenancy and attaching that CR's policy at the tenancy.
  compartment_id = var.compartment_id
  name           = "${var.prefix}-oke-controllers"
  description    = "Bootstrap grant letting the in-cluster Crossplane OCI provider manage the per-app IAM policies, via instance principal"

  statements = [
    "Allow dynamic-group ${oci_identity_dynamic_group.controllers.name} to manage policies in compartment id ${var.compartment_id}",
  ]
}

resource "oci_containerengine_cluster" "this" {
  compartment_id     = var.compartment_id
  name               = var.prefix
  vcn_id             = var.vcn_id
  kubernetes_version = local.kubernetes_version

  endpoint_config {
    is_public_ip_enabled = var.is_public_ip_enabled
    subnet_id            = local.endpoint_subnet_id
    nsg_ids              = [oci_core_network_security_group.endpoint.id]
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
  nsg_ids                 = [oci_core_network_security_group.node.id]
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
  nsg_ids = [oci_core_network_security_group.node.id]
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
  nsg_ids                 = [oci_core_network_security_group.node.id]
}
