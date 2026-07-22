data "oci_containerengine_cluster_option" "this" {
  cluster_option_id = "all"
  compartment_id    = var.compartment_id
}

data "oci_core_vcn" "this" {
  vcn_id = var.vcn_id
}
