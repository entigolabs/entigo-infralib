
locals {
   yqcomm = "yq -y -i '.version = \"${var.branch == "main" ? "0.1.0" : var.branch}\"' helm/modules/k8s/argocd/Chart.yaml"
}

data "external" "argocd" {
  program = ["bash", "-c", "rm -rf helm && git clone --depth 1 -b ${var.branch} ${var.repository} helm && ${local.yqcomm} && echo '{}'"]
}

locals {
  # This hash forces Terraform to redeploy if a new template file is added or changed, or values are updated
  namespace = var.namespace == "" ? "${local.hname}-aws" : var.namespace
  # chart_hash = sha1(join("", [for f in fileset("helm/modules/k8s/argocd", "**/*.yaml"): filesha1("helm/modules/k8s/argocd/${f}")]))
  values_template = templatefile("${path.module}/values.yaml", {
      hostname = var.hostname
      install_crd = var.install_crd
      workspace = terraform.workspace
      prefix = var.prefix
      namespace = local.namespace
      ingress_group_name = var.ingress_group_name
      ingress_scheme = var.ingress_scheme
  })
  values_hash = sha1(local.values_template)
  
}

resource "helm_release" "argocd" {
  name = var.name == "" ? "${local.hname}-argocd-aws" : var.name
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

resource "aws_ssm_parameter" "argocd_namespace" {
  name  = "/entigo-infralib/${local.hname}/namespace"
  type  = "String"
  value = local.namespace
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}

resource "aws_ssm_parameter" "argocd_hostname" {
  name  = "/entigo-infralib/${local.hname}/hostname"
  type  = "String"
  value = var.hostname
  tags = {
    Terraform = "true"
    Prefix    = var.prefix
    Workspace = terraform.workspace
  }
}
