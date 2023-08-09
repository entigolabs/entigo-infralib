terraform {
  backend "s3" {}
  required_version = ">= 1.4"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">4"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~>2.0"
    }
  }
}

provider "aws" {
}

# Test with default values
module "test_route53" {
  source = "../"

  prefix = var.prefix
  vpc_prefix = var.vpc_prefix
  create_public = var.create_public
  create_private = var.create_private
  parent_zone_id = var.parent_zone_id
  parent_domain = var.parent_zone_id
}





