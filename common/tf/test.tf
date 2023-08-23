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

provider "aws" {
  region = "us-east-1"
  alias  = "us-east-1"
}

provider "kubernetes" {
  host                   = module.test.cluster_endpoint
  cluster_ca_certificate = base64decode(module.test.cluster_certificate_authority_data)

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = ["eks", "get-token", "--cluster-name", module.test.cluster_name]
  }
}
