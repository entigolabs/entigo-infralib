locals {
  subdomain_name = var.subdomain_name != "" ? var.subdomain_name : var.prefix
  domain         = "${local.subdomain_name}.${var.parent_domain}"
}

resource "oci_dns_zone" "pub" {
  compartment_id = var.compartment_id
  name           = local.domain
  zone_type      = "PRIMARY"
  scope          = "GLOBAL"
}
