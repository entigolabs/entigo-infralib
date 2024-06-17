resource "helm_release" "argocd" {
  name = "argocd"

  repository       = "https://argoproj.github.io/argo-helm"
  chart            = "argo-cd"
  namespace        = "argocd"
  version          = "7.1.2"
  create_namespace = true
  values = [
    templatefile("${path.module}/values.yaml", {
      hname = local.hname
    })
    ]
}
