# The service gateway targets the regional "All Services" object so traffic to
# OCI services (Object Storage, etc.) stays on the Oracle backbone.
data "oci_core_services" "all" {
  filter {
    name   = "name"
    values = ["All .* Services In Oracle Services Network"]
    regex  = true
  }
}

data "oci_core_vcn" "existing" {
  count  = var.vcn_id == null ? 0 : 1
  vcn_id = var.vcn_id

  lifecycle {
    postcondition {
      condition     = self.cidr_blocks != null
      error_message = "VCN ${var.vcn_id} returned no CIDR blocks: check vcn_id, and that this identity can read VCNs in the VCN's compartment."
    }
  }
}
