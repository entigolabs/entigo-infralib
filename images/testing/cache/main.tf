terraform {
  required_version = ">= 1.4"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.77.0"
    }
    google = {
      source = "hashicorp/google"
      version = "6.11.1"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.29.0"
    }
    external = {
      source = "hashicorp/external"
      version = "2.3.3"
    }
    tls = {
      source = "hashicorp/tls"
      version = "4.0.5"
    }
    time = {
      source = "hashicorp/time"
      version = "0.11.1"
    }
    cloudinit = {
      source = "hashicorp/cloudinit"
      version = "2.3.4"
    }
    null = {
      source  = "hashicorp/null"
      version = "3.2.2"
    }
  }
}

