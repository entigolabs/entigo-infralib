repository  = "https://github.com/entigolabs/entigo-infralib-release.git"
branch = "main"
path = "modules/k8s/argocd"

values = <<EOT
argocd:
  crds:
    install: false
  server:
    config:
      url: https://argocd.runner-main-biz-int.infralib.entigo.io
    ingress:
      hosts:
      - argocd.runner-main-biz-int.infralib.entigo.io
EOT
