terraform {
  required_version = ">= 1.5"
  required_providers {
    helm = {
      source = "hashicorp/helm"
      version = "2.11.0"
    }
    external = {
      source = "hashicorp/external"
      version = "2.3.1"
    }
  }
}
