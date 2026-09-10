locals {
  availability_domains = data.oci_identity_availability_domains.this.availability_domains[*].name
  node_labels          = merge(var.labels, { created-by = "entigo-infralib" })

  # node_pool_os_type still returns aarch64 and GPU variants alongside the plain x86_64
  # image for the same OS/k8s version - picking the wrong one fails at apply with
  # "Invalid nodeShape: Node shape and image are not compatible." Excluding both by name
  # keeps this matched to VM.Standard.E*.Flex (x86_64, non-GPU) shapes.
  image_id = [
    for s in data.oci_containerengine_node_pool_option.this.sources : s.image_id
    if s.source_type == "IMAGE"
    && !strcontains(lower(s.source_name), "aarch64")
    && !strcontains(lower(s.source_name), "gpu")
  ][0]
}

resource "oci_containerengine_node_pool" "this" {
  cluster_id         = var.cluster_id
  compartment_id     = var.compartment_id
  name               = "${var.prefix}-${var.pool_name}"
  kubernetes_version = var.kubernetes_version
  node_shape         = var.node_shape

  # cluster-autoscaler resizes pools through the OCI API, so after any scale event every
  # terraform plan would show a size diff and every apply would snap the pool back to
  # node_count - and this repo's agent auto-applies, so scaling would be undone on each
  # run. Deliberate deviation from aws/eks-node-group (which doesn't ignore desired_size):
  # node_count therefore only applies at pool creation; resize non-autoscaled pools via
  # the OCI console/CLI (or by recreating the pool).
  lifecycle {
    ignore_changes = [node_config_details[0].size]
  }

  node_shape_config {
    ocpus         = var.ocpus
    memory_in_gbs = var.memory_in_gbs
  }

  node_source_details {
    image_id                = local.image_id
    source_type             = "IMAGE"
    boot_volume_size_in_gbs = var.boot_volume_size_in_gbs
  }

  node_config_details {
    size    = var.node_count
    nsg_ids = var.nsg_ids

    # Encrypts the nodes' boot volumes with a customer-managed key instead of Oracle's. Null
    # rather than "" when unset: the API reads an empty string as an explicit, invalid value.
    # Requires blockstorage to be allowed to use the key - modules/oracle/kms grants that.
    kms_key_id = var.kms_key_id != "" ? var.kms_key_id : null

    # Under VCN-native networking each pod holds a real VCN IP off the pod subnet, taken
    # from secondary VNICs on the node. The other three arguments only apply in that mode,
    # so they are nulled out for Flannel rather than left as empty lists (which the API
    # reads as an explicit, invalid value).
    node_pool_pod_network_option_details {
      cni_type          = var.cni_type
      pod_subnet_ids    = var.cni_type == "OCI_VCN_IP_NATIVE" ? var.pod_subnet_ids : null
      pod_nsg_ids       = var.cni_type == "OCI_VCN_IP_NATIVE" ? var.pod_nsg_ids : null
      max_pods_per_node = var.cni_type == "OCI_VCN_IP_NATIVE" ? var.max_pods_per_node : null
    }

    dynamic "placement_configs" {
      for_each = local.availability_domains
      content {
        availability_domain = placement_configs.value
        subnet_id           = element(var.subnet_ids, placement_configs.key)
      }
    }
  }

  dynamic "initial_node_labels" {
    for_each = local.node_labels
    content {
      key   = initial_node_labels.key
      value = initial_node_labels.value
    }
  }
}
