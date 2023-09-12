provider "helm" {
  kubernetes {
    config_context="arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz"
    config_path = "~/.kube/config"
  }
}
