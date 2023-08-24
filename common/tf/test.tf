terraform {
  backend "s3" {}
  required_version = ">= 1.4"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">4"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~>2.0"
    }
  }
}

provider "aws" {
}

provider "kubernetes" {
  host                   = module.test.cluster_endpoint
  cluster_ca_certificate = base64decode(module.test.cluster_certificate_authority_data)
  ignore_annotations = ["helm\\.sh\\/resource-policy","meta\\.helm\\.sh\\/release-name","meta\\.helm\\.sh\\/release-namespace"]
  ignore_labels = ["app\\.kubernetes\\.io\\/managed-by"]
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = ["eks", "get-token", "--cluster-name", module.test.cluster_name]
  }
}
