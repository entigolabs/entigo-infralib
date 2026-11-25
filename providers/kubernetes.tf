provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
  ignore_annotations = ["helm\\.sh\\/resource-policy","meta\\.helm\\.sh\\/release-name","meta\\.helm\\.sh\\/release-namespace","argocd\\.argoproj\\.io\\/sync-wave","argocd\\.argoproj\\.io\\/sync-options"]
  ignore_labels = ["app\\.kubernetes\\.io\\/managed-by","argocd\\.argoproj\\.io\\/instance"]
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    # shackle signs the token itself; the image has no AWS CLI. The region
    # comes from AWS_REGION, which every step has.
    command     = "shackle"
    args        = ["token", "aws", "--cluster", module.eks.cluster_name]
  }
}
