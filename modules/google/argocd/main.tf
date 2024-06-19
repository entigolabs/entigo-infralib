data "external" "argocd" {
  program = ["bash", "-c", "rm -rf helm && git clone --depth 1 -b ${var.branch} ${var.repository} helm && echo '{}'"]
}


locals {
  # This hash forces Terraform to redeploy if a new template file is added or changed, or values are updated
  namespace = var.namespace == "" ? "${local.hname}-argocd-aws" : var.namespace
  # chart_hash = sha1(join("", [for f in fileset("helm/modules/k8s/argocd", "**/*.yaml"): filesha1("helm/modules/k8s/argocd/${f}")]))
  values_template = templatefile("${path.module}/values.yaml", {
      hostname = var.hostname
      install_crd = var.install_crd
      namespace = local.namespace
  })
  values_hash = sha1("${local.values_template} ${var.branch}")
  
}

resource "helm_release" "argocd" {
  name = var.name == "" ? "${local.hname}-argocd-gke" : var.name
  chart            = "helm/modules/k8s/argocd" 
  namespace        = local.namespace
  create_namespace = var.create_namespace
  values = [
    local.values_template
  ]
  set {
    name = "values-hash"
    value = local.values_hash
  }
  depends_on = [data.external.argocd]
}

module "argocd_namespace" {
  source                             = "./secret"
  prefix = var.prefix
  key = "namespace"
  value = local.namespace
}

module "argocd_hostname" {
  source                             = "./secret"
  prefix = var.prefix
  key = "hostname"
  value = var.hostname
}
