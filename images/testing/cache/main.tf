terraform {
  required_version = ">= 1.4"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.83.1"
    }
    google = {
      source = "hashicorp/google"
      version = "6.20.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.35.1"
    }
    external = {
      source = "hashicorp/external"
      version = "2.3.3"
    }
    tls = {
      source = "hashicorp/tls"
      version = "4.0.6"
    }
    time = {
      source = "hashicorp/time"
      version = "0.12.1"
    }
    cloudinit = {
      source = "hashicorp/cloudinit"
      version = "2.3.5"
    }
    null = {
      source  = "hashicorp/null"
      version = "3.2.3"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.6.3"
    }
  }
}

