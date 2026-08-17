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
# Split by role rather than one shared "allow all" group: the control plane only ever needs
# 6443/12250 from workers, so it gets exactly that. Workers genuinely need ALL protocols
# from each other and from the endpoint per Oracle's own doc, so that one is scoped by
# source (VCN CIDR) rather than by protocol - narrowing it would mean pinning down every
# port kube-proxy and the CNI happen to use. Pods get their own group further down, since
# under VCN-native networking they live on a separate subnet with their own VNICs.
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
  description               = "Control plane -> worker node (kubelet 10250) and worker node <-> worker node"
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

# NSG for pod VNICs under VCN-native networking. Pods sit on their own subnet with real VCN
# IPs, so nothing in the node NSG above covers them - and there is a lot that has to reach
# them: other pods on the same and other nodes, the node's own kubelet, the control plane
# calling admission webhooks (oci-native-ingress-controller runs one), and the ingress load
# balancer addressing pods directly as backends. Scoped by source (the VCN) rather than by
# protocol for the same reason the node NSG is: narrowing it would mean pinning down every
# port the CNI, kube-proxy and each admission webhook happen to use. Egress is already open
# through the VCN's default security list.
resource "oci_core_network_security_group" "pods" {
  compartment_id = var.compartment_id
  vcn_id         = var.vcn_id
  display_name   = "${var.prefix}-oke-pods"
}

resource "oci_core_network_security_group_security_rule" "pods_intra_vcn" {
  network_security_group_id = oci_core_network_security_group.pods.id
  direction                 = "INGRESS"
  protocol                  = "all"
  source                    = data.oci_core_vcn.this.cidr_blocks[0]
  source_type               = "CIDR_BLOCK"
  description               = "Pod <-> pod, kubelet -> pod, control plane -> webhooks, load balancer -> pod backends"
}

resource "oci_core_network_security_group_security_rule" "pods_path_mtu_discovery" {
  network_security_group_id = oci_core_network_security_group.pods.id
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
# Backend traffic needs no rule here: under VCN-native networking the LB addresses pods
# directly, which the pods NSG above already allows from anywhere in the VCN.
resource "oci_core_network_security_group" "lb" {
  compartment_id = var.compartment_id
  vcn_id         = var.vcn_id
  display_name   = "${var.prefix}-oke-lb"
}

# One rule per port rather than a resource per port, because the port list is now something a
# deployment changes: an app moved to its own listener (a second certificate needs a second
# listener - see modules/k8s/oci-native-ingress-controller) receives nothing until its port is
# open here, and NIC never touches NSGs itself.
resource "oci_core_network_security_group_security_rule" "lb" {
  for_each = { for port in var.lb_ingress_ports : tostring(port) => port }

  network_security_group_id = oci_core_network_security_group.lb.id
  direction                 = "INGRESS"
  protocol                  = "6" # TCP
  source                    = "0.0.0.0/0"
  source_type               = "CIDR_BLOCK"
  description               = "Public ingress to load balancers on ${each.key}"

  tcp_options {
    destination_port_range {
      min = each.value
      max = each.value
    }
  }
}

# A second NSG for *internal* load balancers - same ports, reachable only from inside the VCN.
# modules/k8s/oci-native-ingress-controller attaches this to the load balancer of an
# IngressClass with isPrivate = true, the way it attaches the one above to the public class.
#
# Two NSGs rather than one with both sources because the whole point of a private load balancer
# is that its NSG does not say 0.0.0.0/0. Sharing an NSG would silently make it public.
resource "oci_core_network_security_group" "lb_int" {
  compartment_id = var.compartment_id
  vcn_id         = var.vcn_id
  display_name   = "${var.prefix}-oke-lb-int"
}

resource "oci_core_network_security_group_security_rule" "lb_int" {
  for_each = { for port in var.lb_ingress_ports : tostring(port) => port }

  network_security_group_id = oci_core_network_security_group.lb_int.id
  direction                 = "INGRESS"
  protocol                  = "6" # TCP
  source                    = data.oci_core_vcn.this.cidr_blocks[0]
  source_type               = "CIDR_BLOCK"
  description               = "In-VCN ingress to internal load balancers on ${each.key}"

  tcp_options {
    destination_port_range {
      min = each.value
      max = each.value
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

# Instance Principal identity for the Crossplane OCI provider - any instance in the
# compartment. This is the bootstrap identity only: terraform provisions the platform up to
# argocd and crossplane, and from there crossplane creates each application's own IAM.
#
# Everything else has moved to OKE Workload Identity and is scoped to its own service account
# (see the policies in modules/k8s/{external-dns,cluster-autoscaler,oci-native-ingress-controller}).
# Crossplane stays on instance principal deliberately, which means the grants below - including
# user and group management, which loki needs because the S3-compatibility endpoint accepts
# nothing but a Customer Secret Key - remain node-wide: any pod scheduled on these nodes can
# use them. That is the residual cost of keeping the bootstrap layer in terraform.
resource "oci_identity_dynamic_group" "controllers" {
  compartment_id = data.oci_identity_compartment.this.compartment_id
  name           = "${var.prefix}-oke-controllers"
  description    = "Instances in ${var.prefix}'s OKE node pools - used by the Crossplane OCI provider via instance principal auth; other controllers use OKE Workload Identity"
  matching_rule  = "ALL {instance.compartment.id = '${var.compartment_id}'}"
}

resource "oci_identity_policy" "controllers" {
  # Bootstrap grant only: per-app permissions (external-dns "manage dns",
  # cluster-autoscaler's node-pool statements, the ingress controller's cert/LB set)
  # live in each k8s module's templates/oracle/ as Crossplane Policy CRs, applied by
  # the crossplane-oracle provider - which is what this statement authorizes.
  #
  # Attached at the tenancy, and granting "in tenancy", because OCI only lets a policy
  # grant within the subtree it is attached to and oci-native-ingress-controller needs
  # genuinely tenancy-scoped statements (read public-ips, manage floating-ips, use
  # tag-namespaces - see modules/k8s/oci-native-ingress-controller/templates/policy.yaml).
  # A compartment-attached grant would be the tighter boundary, but it cannot create that
  # policy at all. This is the one grant worth reading carefully: it lets anything running
  # on these nodes write IAM policy anywhere in the tenancy.
  compartment_id = data.oci_identity_compartment.this.compartment_id
  name           = "${var.prefix}-oke-controllers"
  description    = "Bootstrap grant letting the in-cluster Crossplane OCI provider manage the per-app IAM policies, via instance principal"

  statements = [
    "Allow dynamic-group ${oci_identity_dynamic_group.controllers.name} to manage policies in tenancy",
    # Users and groups are needed because loki reaches Object Storage through the
    # S3-compatibility endpoint, and that endpoint authenticates only with a Customer Secret
    # Key, which belongs to an IAM user - there is no instance-principal or Workload Identity
    # path (see modules/k8s/loki/templates/oracle/credentials.yaml). Crossplane therefore has
    # to be able to create the user, put it in a group the policy can name, and mint the key.
    #
    # Read this as the escalation it is: anything running on these nodes can create tenancy
    # users. It becomes unnecessary the moment Loki gains a native OCI backend -
    # https://github.com/grafana/loki/issues/23687 - at which point drop these two lines.
    "Allow dynamic-group ${oci_identity_dynamic_group.controllers.name} to manage users in tenancy",
    "Allow dynamic-group ${oci_identity_dynamic_group.controllers.name} to manage groups in tenancy",
    # And the bucket those chunks go into. loki's own Policy CR grants its IAM *user* "manage
    # objects" and "inspect buckets" - enough to write into a bucket, not to create one - so
    # without this the Bucket CR sits in ApplyFailure reporting "409-BucketAlreadyExists,
    # Either the bucket 'tarmo-loki' ... already exists or you are not authorized to create
    # it". It is the second half of that sentence: the bucket does not exist anywhere in the
    # tenancy, and bucket names are unique per namespace, so an admin would see it if it did.
    #
    # Scoped to the compartment rather than the tenancy, unlike the three above, because
    # nothing needs to create buckets outside it. Goes away with the same upstream fix.
    "Allow dynamic-group ${oci_identity_dynamic_group.controllers.name} to manage buckets in compartment id ${var.compartment_id}",
    # A bucket that names a customer-managed key needs the *caller* to be allowed to use that
    # key, not only Object Storage itself - modules/oracle/kms grants the service, this grants
    # whoever asks. Without it CreateBucket is refused, and the refusal is indistinguishable
    # from the bucket already existing: the same "409-BucketAlreadyExists ... or you are not
    # authorized to create it" above, on a bucket that a GET reports as 404.
    #
    # The same two-principal split caught out worker node boot volumes, where granting only
    # "service oke" left node provisioning failing with no mention of a key at all.
    "Allow dynamic-group ${oci_identity_dynamic_group.controllers.name} to use keys in compartment id ${var.compartment_id}",
  ]
}

resource "oci_containerengine_cluster" "this" {
  compartment_id = var.compartment_id
  name           = var.prefix
  vcn_id         = var.vcn_id
  # OCI defaults an unset type to BASIC_CLUSTER, which is what we ran on until now. Basic
  # clusters have no OKE Workload Identity and no cluster add-ons, so every in-cluster
  # controller is stuck on instance principal (node-wide identity, no per-pod isolation)
  # and nothing Oracle ships as an add-on can be used. Enhanced is not free, but the
  # capability gap is not worth working around. Not a variable: a cluster's type can be
  # upgraded Basic -> Enhanced in place but never downgraded, so offering the choice would
  # only let a consumer pick the one that can't be undone.
  type               = "ENHANCED_CLUSTER"
  kubernetes_version = local.kubernetes_version

  # Encrypts etcd - and so every Kubernetes Secret - with a customer-managed key rather than
  # Oracle's. Null when unset, because the API rejects an empty string.
  #
  # Creation-time only: OCI will not re-key an existing cluster, so setting this on a live one
  # asks terraform to REPLACE the cluster. That is why it is not wired automatically from
  # modules/oracle/kms the way node_kms_key_id is - the agent applies plans without a human
  # reading them, and a silent cluster replacement is not an acceptable surprise.
  kms_key_id = var.etcd_kms_key_id != "" ? var.etcd_kms_key_id : null

  endpoint_config {
    is_public_ip_enabled = var.is_public_ip_enabled
    subnet_id            = local.endpoint_subnet_id
    nsg_ids              = [oci_core_network_security_group.endpoint.id]
  }

  cluster_pod_network_options {
    # Every pod gets a real VCN IP off var.pod_subnet_ids, the same model as the AWS VPC CNI
    # on EKS and VPC-native on GKE. Not just for parity: with the Flannel overlay an OCI
    # load balancer has no route to pod IPs, so oci-native-ingress-controller falls back to
    # nodeIP:nodePort backends and refuses any ClusterIP service outright ("Node port not
    # found for service", pkg/controllers/nodeBackend). VCN-native lets it address pods
    # directly and keeps our Services ClusterIP like everywhere else. The controller picks
    # its backend strategy from this value by querying the cluster, so the two cannot
    # disagree. Fixed at cluster creation - changing it means replacing the cluster.
    cni_type = "OCI_VCN_IP_NATIVE"
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
  pod_subnet_ids          = var.pod_subnet_ids
  pod_nsg_ids             = [oci_core_network_security_group.pods.id]
  max_pods_per_node       = var.max_pods_per_node
  kms_key_id              = var.node_kms_key_id
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
  labels            = { mon = "true" }
  nsg_ids           = [oci_core_network_security_group.node.id]
  pod_subnet_ids    = var.pod_subnet_ids
  pod_nsg_ids       = [oci_core_network_security_group.pods.id]
  max_pods_per_node = var.max_pods_per_node
  kms_key_id        = var.node_kms_key_id
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
  pod_subnet_ids          = var.pod_subnet_ids
  pod_nsg_ids             = [oci_core_network_security_group.pods.id]
  max_pods_per_node       = var.max_pods_per_node
  kms_key_id              = var.node_kms_key_id
}

# NSG for OCI *Network* Load Balancers, which is what a UDP Service gets: the classic Load
# Balancer behind oci-native-ingress-controller is TCP/HTTP only, so a Service asking for UDP
# has to be an NLB (oci.oraclecloud.com/load-balancer-type: "nlb"). wireguard is the reason
# this exists.
#
# Separate from the lb/lb_int NSGs because the protocol differs and because an NSG costs
# nothing until something attaches itself to it - a deployment without wireguard simply never
# references this one.
resource "oci_core_network_security_group" "nlb" {
  compartment_id = var.compartment_id
  vcn_id         = var.vcn_id
  display_name   = "${var.prefix}-oke-nlb"
}

resource "oci_core_network_security_group_security_rule" "nlb_udp" {
  for_each = { for port in var.nlb_ingress_udp_ports : tostring(port) => port }

  network_security_group_id = oci_core_network_security_group.nlb.id
  direction                 = "INGRESS"
  protocol                  = "17" # UDP
  source                    = "0.0.0.0/0"
  source_type               = "CIDR_BLOCK"
  description               = "Public UDP to network load balancers on ${each.key}"

  udp_options {
    destination_port_range {
      min = each.value
      max = each.value
    }
  }
}
