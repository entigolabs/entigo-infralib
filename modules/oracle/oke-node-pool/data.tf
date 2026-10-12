data "oci_identity_availability_domains" "this" {
  compartment_id = var.compartment_id
}

data "oci_containerengine_node_pool_option" "this" {
  node_pool_option_id   = var.cluster_id
  compartment_id        = var.compartment_id
  node_pool_k8s_version = var.kubernetes_version
  node_pool_os_type     = var.node_pool_os_type
}
