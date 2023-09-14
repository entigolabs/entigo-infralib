provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
  ignore_annotations = ["helm\\.sh\\/resource-policy","meta\\.helm\\.sh\\/release-name","meta\\.helm\\.sh\\/release-namespace","argocd\\.argoproj\\.io\\/sync-wave","argocd\\.argoproj\\.io\\/sync-options"]
  ignore_labels = ["app\\.kubernetes\\.io\\/managed-by","argocd\\.argoproj\\.io\\/instance"]
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
  }
}
