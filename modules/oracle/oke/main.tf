locals {
  # OCI returns available versions oldest-first; take the last (newest) one when the
  # caller doesn't pin a specific version.
  versions           = data.oci_containerengine_cluster_option.this.kubernetes_versions
  kubernetes_version = var.kubernetes_version != "" ? var.kubernetes_version : local.versions[length(local.versions) - 1]

  # OCI rejects a private subnet for the endpoint when is_public_ip_enabled is true
  # ("must be a public subnet if public ip enabled"), so the subnet choice must follow it.
  endpoint_subnet_id = var.is_public_ip_enabled ? var.public_subnet_id : var.private_subnet_id
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
