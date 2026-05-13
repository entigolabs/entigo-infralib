terraform {
  required_version = ">= 1.4"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.44.0"
    }
    google = {
      source = "hashicorp/google"
      version = "7.31.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "3.0.1"
    }
    external = {
      source = "hashicorp/external"
      version = "2.3.3"
    }
    tls = {
      source = "hashicorp/tls"
      version = "4.3.0"
    }
    time = {
      source = "hashicorp/time"
      version = "0.14.0"
    }
    cloudinit = {
      source = "hashicorp/cloudinit"
      version = "2.4.0"
    }
    null = {
      source  = "hashicorp/null"
      version = "3.3.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.8.1"
    }
  }
}

