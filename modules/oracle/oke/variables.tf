variable "prefix" {
  type = string
}

variable "compartment_id" {
  description = "OCID of the compartment that will contain the cluster."
  type        = string
}

variable "vcn_id" {
  type = string
}

# Applies to both ingress load balancer NSGs: the public one accepts these ports from
# 0.0.0.0/0, the internal one from the VCN CIDR only. NIC does not manage NSGs - it only
# attaches a load balancer to the ones named in its IngressClass - so a listener on a port
# missing from this list comes up healthy and receives nothing.
variable "lb_ingress_ports" {
  description = "TCP ports the ingress load balancer NSGs accept. 80 and 443 are what the default IngressClass listens on; add a port here before pointing an app's https-listener-port at it."
  type        = list(number)
  default     = [80, 443]
}

variable "private_subnet_id" {
  description = "Subnet for the Kubernetes API endpoint when is_public_ip_enabled is false."
  type        = string
}

variable "public_subnet_id" {
  description = "Subnet for the Kubernetes API endpoint when is_public_ip_enabled is true. OCI requires the endpoint subnet to be public (prohibit_public_ip_on_vnic = false) whenever a public IP is assigned to it."
  type        = string
  default     = ""
}

variable "is_public_ip_enabled" {
  type     = bool
  nullable = false
  default  = false
}

variable "service_lb_subnet_ids" {
  description = "Subnets used for LoadBalancer-type Kubernetes services, typically a public subnet."
  type        = list(string)
  default     = []
}

# The in-cluster Crossplane OCI provider's identity, named in the bootstrap policy so the
# grant can be scoped to that workload rather than to every instance in the compartment.
# Same pair aws/crossplane and google/crossplane take, with the same defaults - the service
# account name is pinned by modules/k8s/crossplane-oracle's DeploymentRuntimeConfig rather
# than generated per provider revision, which is what makes it nameable from here.
variable "kubernetes_service_account" {
  type        = string
  description = "Kubernetes service account name for oracle crossplane provider"
  default     = "crossplane-oracle"
}

variable "kubernetes_namespace" {
  type        = string
  description = "Kubernetes namespace name for crossplane"
  default     = "crossplane-system"
}

variable "kubernetes_version" {
  description = "Kubernetes version for the cluster and its node pools. A minor version takes the newest patch OKE offers for it; an exact one (1.36.1) pins that patch. Pinned so a released infralib is the stack it was tested on - override only to hold back an upgrade."
  type        = string
  default     = "1.36"
}

variable "pods_cidr" {
  type    = string
  default = "10.244.0.0/16"
}

variable "services_cidr" {
  type    = string
  default = "10.96.0.0/16"
}

variable "node_subnet_ids" {
  description = "Default subnets nodes are placed in - one per availability domain, in order. Reused across ADs if fewer are given than ADs available. Used by main/mon/tools unless overridden per-pool below."
  type        = list(string)
  default     = []
}

variable "pod_subnet_ids" {
  description = "Subnets pods draw their VCN IPs from. The cluster uses OCI_VCN_IP_NATIVE pod networking, so this is required; OKE rejects a pod subnet that is public or scoped to a single availability domain. Wired from modules/oracle/vpc's pod_subnets output."
  type        = list(string)
}

variable "max_pods_per_node" {
  description = "Pod capacity per node. Capped by the node shape: MIN((VNICs - 1) * 31, 110), since one VNIC serves the node and each of the rest carries 31 pod IPs. Flexible shapes get one VNIC per OCPU with a floor of two, so the 1-OCPU pool defaults allow exactly 31 - raise the pool's ocpus before raising this."
  type        = number
  default     = 31
}

variable "oke_main_node_count" {
  type     = number
  nullable = false
  default  = 0
}

variable "oke_main_ocpus" {
  type    = number
  default = 1
}

variable "oke_main_memory_in_gbs" {
  type    = number
  default = 8
}

variable "oke_main_node_shape" {
  type    = string
  default = "VM.Standard.E4.Flex"
}

variable "oke_main_node_pool_os_type" {
  type    = string
  default = "OL8"
}

variable "oke_main_boot_volume_size_in_gbs" {
  type    = string
  default = "50"
}

variable "oke_main_subnet_ids" {
  description = "Overrides node_subnet_ids for the main pool. Defaults to node_subnet_ids when empty."
  type        = list(string)
  default     = []
}

variable "oke_mon_node_count" {
  type     = number
  nullable = false
  default  = 0
}

variable "oke_mon_ocpus" {
  type    = number
  default = 1
}

variable "oke_mon_memory_in_gbs" {
  type    = number
  default = 8
}

variable "oke_mon_node_shape" {
  type    = string
  default = "VM.Standard.E4.Flex"
}

variable "oke_mon_node_pool_os_type" {
  type    = string
  default = "OL8"
}

variable "oke_mon_boot_volume_size_in_gbs" {
  type    = string
  default = "50"
}

variable "oke_mon_subnet_ids" {
  description = "Overrides node_subnet_ids for the mon pool. Defaults to node_subnet_ids when empty."
  type        = list(string)
  default     = []
}

variable "oke_tools_node_count" {
  type     = number
  nullable = false
  default  = 2
}

variable "oke_tools_ocpus" {
  type    = number
  default = 2
}

variable "oke_tools_memory_in_gbs" {
  type    = number
  default = 8
}

variable "oke_tools_node_shape" {
  type    = string
  default = "VM.Standard.E4.Flex"
}

variable "oke_tools_node_pool_os_type" {
  type    = string
  default = "OL8"
}

variable "oke_tools_boot_volume_size_in_gbs" {
  type    = string
  default = "50"
}

variable "oke_tools_subnet_ids" {
  description = "Overrides node_subnet_ids for the tools pool. Defaults to node_subnet_ids when empty."
  type        = list(string)
  default     = []
}

variable "oke_main_min_size" {
  type     = number
  nullable = false
  default  = 0
}

variable "oke_main_max_size" {
  type     = number
  nullable = false
  default  = 0
}

variable "oke_mon_min_size" {
  type     = number
  nullable = false
  default  = 0
}

variable "oke_mon_max_size" {
  type     = number
  nullable = false
  default  = 0
}

variable "oke_tools_min_size" {
  type     = number
  nullable = false
  default  = 2
}

variable "oke_tools_max_size" {
  type     = number
  nullable = false
  default  = 3
}

# Encrypts etcd, and therefore every Kubernetes Secret, with a customer-managed key.
#
# NOT wired from modules/oracle/kms automatically, unlike node_kms_key_id below. OCI will not
# re-key an existing cluster, so this is creation-time only: setting it on a live cluster makes
# terraform plan a REPLACEMENT. The agent applies plans unattended, so auto-wiring it would
# turn "add a kms module" into "silently rebuild the cluster". Set it explicitly in a
# deployment's config, on a cluster that does not exist yet.
variable "etcd_kms_key_id" {
  description = "OCID of a key to encrypt etcd with. Empty leaves etcd on Oracle-managed encryption. Creation-time only - setting this on an existing cluster replaces it."
  type        = string
  default     = ""
}

# Passed to all three node pools. Wired from modules/oracle/kms by agent_input.yaml, matching
# node_encryption_kms_key_arn in modules/aws/eks. Changing it replaces the pools' nodes, which
# roll rather than take the cluster down.
variable "node_kms_key_id" {
  description = "OCID of a key to encrypt the worker nodes' boot volumes with. Empty leaves them on Oracle-managed encryption."
  type        = string
  default     = ""
}

# UDP ports the network load balancer NSG accepts. Defaults to WireGuard's, which is the only
# UDP service in the repo; the NSG is unattached until a Service names it, so an unused port
# here costs nothing.
variable "nlb_ingress_udp_ports" {
  description = "UDP ports the network load balancer NSG accepts from anywhere. 51820 is WireGuard."
  type        = list(number)
  default     = [51820]
}

variable "oke_api_access_cidrs" {
  description = "CIDRs allowed to reach the Kubernetes API endpoint on 6443, in addition to the VCN. Empty keeps the endpoint reachable only from inside the VCN."
  type        = list(string)
  nullable    = false
  default     = []
}

# OIDC identity provider for human access to the cluster's Kubernetes API, so a user can get a
# kubeconfig and authenticate with their own identity instead of a cloud credential.
#
# Named after modules/aws/eks's cluster_identity_providers so the agent can wire both from one
# declaration, but the value is a single configuration rather than a map keyed by provider
# name: OCI accepts exactly one per cluster. Same `any` convention as the AWS variable, so a
# deployment's config passes an HCL object.
#
# The claim settings are what the platform's Zitadel expects - groups arrive as a flat array
# and Kubernetes RBAC subjects are written as `oidc:<group>`:
#
#   cluster_identity_providers: |-
#     {
#       issuer_url      = "https://<tenant>.zitadel.cloud"
#       client_id       = "<the workspace's OIDC application client id>"
#       username_claim  = "sub"
#       username_prefix = "oidc:"
#       groups_claim    = "groups"
#       groups_prefix   = "oidc:"
#     }
#
# Enabling this only makes the cluster ACCEPT such tokens. A user authenticated this way still
# has no permissions until something binds their group - see modules/k8s/rbac-bindings.
variable "cluster_identity_providers" {
  description = "OIDC identity provider for Kubernetes API access. OCI supports exactly one per cluster. Empty disables OIDC authentication."
  type        = any
  nullable    = false
  default     = {}
}
