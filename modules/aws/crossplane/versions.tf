terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.18.1"
    }
    helm = {
      source = "hashicorp/helm"
      version = "2.4.1"
    }
  }
}
