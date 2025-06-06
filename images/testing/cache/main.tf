terraform {
  required_version = ">= 1.4"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.95.0"
    }
    google = {
      source = "hashicorp/google"
      version = "6.38.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.37.1"
    }
    external = {
      source = "hashicorp/external"
      version = "2.3.3"
    }
    tls = {
      source = "hashicorp/tls"
      version = "4.1.0"
    }
    time = {
      source = "hashicorp/time"
      version = "0.13.1"
    }
    cloudinit = {
      source = "hashicorp/cloudinit"
      version = "2.3.7"
    }
    null = {
      source  = "hashicorp/null"
      version = "3.2.4"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.7.2"
    }
  }
}

