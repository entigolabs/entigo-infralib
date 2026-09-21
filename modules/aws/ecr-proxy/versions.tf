terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.65.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "0.14.2"
    }
  }
}
