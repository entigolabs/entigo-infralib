locals {
  availability_domains = data.oci_identity_availability_domains.this.availability_domains[*].name

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
  name               = "${var.prefix}-pool"
  kubernetes_version = var.kubernetes_version
  node_shape         = var.node_shape

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
    size = var.node_count

    dynamic "placement_configs" {
      for_each = local.availability_domains
      content {
        availability_domain = placement_configs.value
        subnet_id           = element(var.subnet_ids, placement_configs.key)
      }
    }
  }
}
