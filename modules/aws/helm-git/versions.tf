terraform {
  required_version = ">= 1.5"
  required_providers {
    helm = {
      source = "hashicorp/helm"
      version = "2.10.1"
    }
    external = {
      source = "hashicorp/external"
      version = "2.3.1"
    }
  }
}
