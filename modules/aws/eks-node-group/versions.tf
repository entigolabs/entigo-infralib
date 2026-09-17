terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.64.0"
    }
    null = {
      source  = "hashicorp/null"
      version = "3.3.2"
    }
    cloudinit = {
      source  = "hashicorp/cloudinit"
      version = "2.4.1"
    }
  }
}
