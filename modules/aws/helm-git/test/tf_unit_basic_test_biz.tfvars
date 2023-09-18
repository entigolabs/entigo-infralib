path = "modules/k8s/argocd"

values = <<EOT
argocd:
  crds:
    install: false
  server:
    config:
      url: https://argocd-helm.runner-main-biz-int.infralib.entigo.io
    ingress:
      hosts:
      - argocd-helm.runner-main-biz-int.infralib.entigo.io
EOT
