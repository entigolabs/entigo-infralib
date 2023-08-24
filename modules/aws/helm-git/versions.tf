terraform {
  required_version = ">= 1.4"
  required_providers {
    helm = {
      source = "hashicorp/helm"
      version = "2.10.1"
    }
    null = {
      source = "hashicorp/null"
      version = "3.2.1"
    }
  }
}
