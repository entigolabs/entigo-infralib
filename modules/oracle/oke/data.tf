data "oci_containerengine_cluster_option" "this" {
  cluster_option_id = "all"
  compartment_id    = var.compartment_id
}

data "oci_core_vcn" "this" {
  vcn_id = var.vcn_id
}

# The tenancy's Object Storage namespace. Needed by modules/k8s/loki, which reaches Object
# Storage through the S3-compatibility endpoint and has to build its hostname
# (https://<namespace>.compat.objectstorage.<region>.oraclecloud.com). Tenancy-wide and stable,
# but there is no way to derive it from the compartment OCID, so it has to be looked up.
data "oci_objectstorage_namespace" "this" {
  compartment_id = var.compartment_id
}
