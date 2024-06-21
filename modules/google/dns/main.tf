
data "google_client_config" "this" {}

locals {
  member = "principal://iam.googleapis.com/projects/${var.project_number}/locations/global/workloadIdentityPools/${data.google_client_config.this.project}.svc.id.goog/subject/ns/${var.kns_name}/sa/${var.ksa_name}"
}

resource "google_project_iam_member" "external_dns" {
  member  = local.member
  role    = "roles/dns.reader"
  project = data.google_client_config.this.project
}

resource "google_dns_managed_zone_iam_member" "member" {
  project = data.google_client_config.this.project
  managed_zone = module.dns.name
  role         = "roles/dns.admin"
  member       = local.member
}

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

resource "helm_release" "external-dns" {
  name  = "external-dns"
  chart = "https://github.com/kubernetes-sigs/external-dns/releases/download/external-dns-helm-chart-1.14.2/external-dns-1.14.2.tgz"
  namespace        = "external-dns"
  create_namespace = true
  values = [file("${path.module}/values.yaml")]
}


module "dns_zone" {
  source                             = "./secret"
  prefix = var.prefix
  key = "dns_zone"
  value = trimsuffix(module.dns.domain, ".")
}

