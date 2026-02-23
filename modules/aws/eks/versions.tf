terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.33.0"
    }
    null = {
      source  = "hashicorp/null"
      version = "3.2.4"
    }
    cloudinit = {
      source  = "hashicorp/cloudinit"
      version = "2.3.7"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "4.2.1"
    }
    time = {
      source  = "hashicorp/time"
      version = "0.13.1"
    }
  }
}
