resource "null_resource" "clone" {
  provisioner "local-exec" {
    command = "git clone --depth 1 -b ${var.branch} ${var.repository} helm"
  }
  triggers = {
    always_run = timestamp()
  }
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
  depends_on = [null_resource.clone]
}
