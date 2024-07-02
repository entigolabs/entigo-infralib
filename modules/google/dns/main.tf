
locals {
  domain = var.parent_zone_id != "" ? data.google_dns_managed_zone.parent_zone[0].dns_name : var.parent_domain
  
  public_subdomain_name = var.public_subdomain_name != "" ? var.public_subdomain_name : local.hname
  private_subdomain_name =  var.private_subdomain_name != "" ? var.private_subdomain_name : "${local.hname}-int"
  
  pub_domain = var.create_public ? "${local.public_subdomain_name}.${local.domain}" : local.domain
  int_domain = var.create_private ? "${local.private_subdomain_name}.${local.domain}" : local.pub_domain

}


resource "google_dns_managed_zone" "pub" {
  count = var.create_public ? 1 : 0
  name          = trimsuffix(replace(local.pub_domain, ".", "-"), "-")
  dns_name      = local.pub_domain
  description   = local.pub_domain
  labels        = {}
  visibility    = "public"
  force_destroy = true

  cloud_logging_config {
    enable_logging = false
  }
}

resource "google_dns_managed_zone" "int" {
  count = var.create_private ? 1 : 0
  name          = trimsuffix(replace(local.int_domain, ".", "-"), "-")
  dns_name      = local.int_domain
  description   = local.int_domain
  labels        = {}
  visibility    = "private"
  force_destroy = true

  private_visibility_config {
    networks {
      network_url = var.vpc_id
    }
  }
}

resource "google_dns_record_set" "pub-ns" {
  count = var.parent_zone_id != "" && var.create_public ? 1 : 0
  name         = local.pub_domain
  type         = "NS"
  ttl          = 300
  managed_zone = var.parent_zone_id
  rrdatas      = google_dns_managed_zone.pub[0].name_servers
}

resource "google_dns_record_set" "int-ns" {
  count = var.parent_zone_id != "" && var.create_private ? 1 : 0
  name         = local.int_domain
  type         = "NS"
  ttl          = 300
  managed_zone = var.parent_zone_id
  rrdatas      = google_dns_managed_zone.int[0].name_servers
}

locals {
  zone = var.parent_zone_id != "" ? data.google_dns_managed_zone.parent_zone[0].name : ""
  pub_zone = var.create_public ? google_dns_managed_zone.pub[0].name : local.zone
  int_zone = var.create_private ? google_dns_managed_zone.int[0].name : local.pub_zone
}


module "pub_zone_id" {
  source                             = "./secret"
  prefix = var.prefix
  key = "pub_zone_id"
  value = local.pub_zone
}

module "pub_domain" {
  source                             = "./secret"
  prefix = var.prefix
  key = "pub_domain"
  value = trimsuffix(local.pub_domain, ".")
}

module "int_zone_id" {
  source                             = "./secret"
  prefix = var.prefix
  key = "int_zone_id"
  value = local.int_zone
}

module "int_domain" {
  source                             = "./secret"
  prefix = var.prefix
  key = "int_domain"
  value = trimsuffix(local.int_domain, ".")
}
