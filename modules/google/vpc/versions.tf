terraform {
  required_version = ">= 1.5"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "7.35.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.9.0"
    }

  }
}
