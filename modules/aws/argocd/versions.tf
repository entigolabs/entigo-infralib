terraform {
  required_version = ">= 1.5"
  required_providers {
    tls = {
      source = "hashicorp/tls"
      version = "4.0.4"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "5.8"
    }
    helm = {
      source = "hashicorp/helm"
      version = "2.10.1"
    }
    null = {
      source = "hashicorp/null"
      version = "3.2.1"
    }
  }
}
