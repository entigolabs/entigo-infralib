data "google_client_config" "this" {}

module "dns" {
  source  = "terraform-google-modules/cloud-dns/google"
  version = "5.2.0"

  project_id                         = data.google_client_config.this.project
  type                               = var.type
  name                               = "${local.hname}-${var.zone}"
  domain                             = "${local.hname}.${var.domain}"

  recordsets = [
    {
      name    = ""
      type    = "NS"
      ttl     = 300
      records = module.dns.name_servers
    },
  ]
}

resource "google_dns_record_set" "root_ns_record" {
  name         = "${local.hname}.${var.domain}"
  type         = "NS"
  ttl          = 300
  managed_zone = var.zone
  rrdatas      = module.dns.name_servers
}

module "dns_zone" {
  source                             = "./secret"
  prefix = var.prefix
  key = "dns_zone"
  value = trimsuffix(module.dns.domain, ".")
}
