provider "kubernetes" {
  host                   = module.test.cluster_endpoint
  cluster_ca_certificate = base64decode(module.test.cluster_certificate_authority_data)
  ignore_annotations = ["helm\\.sh\\/resource-policy","meta\\.helm\\.sh\\/release-name","meta\\.helm\\.sh\\/release-namespace","argocd\\.argoproj\\.io\\/sync-wave"]
  ignore_labels = ["app\\.kubernetes\\.io\\/managed-by"]
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = ["eks", "get-token", "--cluster-name", module.test.cluster_name]
  }
}
 
