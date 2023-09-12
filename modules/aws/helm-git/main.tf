data "external" "helm" {
  program = ["bash", "-c", "rm -rf helm && git clone --depth 1 -b ${var.branch} ${var.repository} helm && echo '{}'"]
}

resource "helm_release" "helm" {
  name = var.name == "" ? "${local.hname}-helm-git" : var.name
  chart            = "helm/${var.path}" 
  namespace        = var.namespace == "" ? "${local.hname}-helm-git" : var.namespace
  create_namespace = var.create_namespace
  values = [
    templatefile("${path.module}/values.yaml", {
      values               = var.values
    })
  ]
  depends_on = [data.external.helm]
}
