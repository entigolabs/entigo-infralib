# The service gateway targets the regional "All Services" object so traffic to
# OCI services (Object Storage, etc.) stays on the Oracle backbone.
data "oci_core_services" "all" {
  filter {
    name   = "name"
    values = ["All .* Services In Oracle Services Network"]
    regex  = true
  }
}
