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
    helm = {
      source = "hashicorp/helm"
      version = "2.10.1"
    }
  }
}

provider "aws" {
  ignore_tags {
      key_prefixes = ["kubernetes.io/cluster/"]
  }
}

provider "aws" {
  region = "us-east-1"
  alias  = "us-east-1"
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

provider "helm" {
  kubernetes {
    config_context="arn:aws:eks:eu-north-1:877483565445:cluster/runner-main-biz"
    config_path = "~/.kube/config"
  }
} 
