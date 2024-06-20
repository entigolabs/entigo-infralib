terraform {
  required_version = ">= 1.5"
  required_providers {
    google = {
      source = "hashicorp/google"
      version = "5.32.0"
    }
    helmgoogle = {
      source  = "hashicorp/helm"
        version = "2.13.1"
    }
  }
}
