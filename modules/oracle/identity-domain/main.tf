locals {
  name_suffix = var.name_salt ? "-${random_string.suffix[0].result}" : ""
  domain_name = "${var.prefix}${local.name_suffix}"
}

resource "random_string" "suffix" {
  count = var.name_salt ? 1 : 0

  length  = 8
  lower   = true
  upper   = false
  numeric = true
  special = false
}

# A domain's users and groups live in the domain's own home compartment, so IAM grants for
# them are ordinary compartment-scoped statements ("manage users in compartment id X") -
# unlike classic/default-domain users, which are always tenancy-scoped.
#
# Loki uses this for its Object Storage access: the S3-compatibility endpoint accepts only a
# Customer Secret Key (requires a User), and Crossplane creates that user to mint the key. See
# modules/oracle/oke for the grant and modules/k8s/loki for the Crossplane templates.
resource "oci_identity_domain" "this" {
  compartment_id = var.compartment_id
  display_name   = local.domain_name
  home_region    = var.region
  license_type   = "free"
  description    = "Compartment-scoped identity domain for per-app service users (Loki, future apps)"

  # UpdateDomain needs a tenancy-scoped grant this test deliberately doesn't hold, and that
  # call fails no matter which field triggered it - confirmed empirically, not just for
  # license_type. So this resource is create-only for us: any drift Terraform notices on any
  # field must never be applied.
  lifecycle {
    ignore_changes = all
  }
}
