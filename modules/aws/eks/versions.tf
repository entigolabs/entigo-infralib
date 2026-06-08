terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.49.0"
    }
    null = {
      source  = "hashicorp/null"
      version = "3.3.0"
    }
    cloudinit = {
      source  = "hashicorp/cloudinit"
      version = "2.4.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "4.3.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "0.14.0"
    }
  }
}
