terraform {
  required_version = ">= 1.4"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.18.1"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.21.1"
    }
    helm = {
      source = "hashicorp/helm"
      version = "2.4.1"
    }
    external = {
      source = "hashicorp/external"
      version = "2.3.1"
    }
    tls = {
      source = "hashicorp/tls"
      version = "4.0.4"
    }
    time = {
      source = "hashicorp/time"
      version = "0.9.1"
    }
    cloudinit = {
      source = "hashicorp/cloudinit"
      version = "2.3.2"
    }
    
  }
}

