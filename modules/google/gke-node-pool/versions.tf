terraform {
  required_version = ">= 1.5"
  required_providers {
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "7.16.0"
    }
  }
}
