terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.8.0"
    }
    helm = {
      source = "hashicorp/helm"
      version = "2.4.1"
    }
  }
}
