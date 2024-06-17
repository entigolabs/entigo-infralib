terraform {
  required_version = ">= 1.5"
  required_providers {
    tls = {
      source = "hashicorp/tls"
      version = "4.0.5"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "5.46.0"
    }
    helmaws = {
      source = "hashicorp/helm"
      version = "2.13.1"
    }
    external = {
      source = "hashicorp/external"
      version = "2.3.3"
    }
  }
}
