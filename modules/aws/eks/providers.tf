terraform {
  backend "s3" { }
  required_version = ">= 1.5" 
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.6"
    }
    kubernetes = {
      source = "hashicorp/kubernetes"
      version = "2.21.1"
    }
  }
} 
